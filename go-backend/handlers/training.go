package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/judge"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dslipak/pdf"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// notesCacheEntry stores a previous Gemini result for identical file content.
type notesCacheEntry struct {
	Summary   string
	Questions []services.GeneratedQuestion
}

// notesCache is an in-memory cache keyed by SHA-256 of extracted text.
var notesCache sync.Map

type notesCandidate struct {
	Question services.GeneratedQuestion
	Source   string
	DBID     uint
}

// UploadNotes handles the multipart file ingestion, extraction, and AI orchestration.
func UploadNotes(c *gin.Context) {
	log.Printf("[NOTES] upload received method=%s path=%s", c.Request.Method, c.Request.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[NOTES][panic] [NOTES_PANIC] %v", r)
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error during notes processing",
					"stage": "panic",
				})
			}
		}
	}()

	fail := func(status int, stage, message, logMessage string) {
		log.Printf("[NOTES][%s] %s", stage, logMessage)
		c.JSON(status, gin.H{
			"error": message,
			"stage": stage,
		})
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fail(http.StatusBadRequest, "extract", "File transfer failed: no file found in request", "missing file in multipart request")
		return
	}
	defer file.Close()

	topic := c.PostForm("topic")
	difficulty := c.PostForm("difficulty")
	countStr := c.PostForm("count")
	count, _ := strconv.Atoi(countStr)
	if count <= 0 {
		count = 5
	}

	log.Printf("[NOTES] request details file=%s type=%s size=%d topic=%s difficulty=%s",
		header.Filename, header.Header.Get("Content-Type"), header.Size, topic, difficulty)

	dotIdx := strings.LastIndex(header.Filename, ".")
	if dotIdx < 0 {
		fail(http.StatusBadRequest, "extract", "Unsupported file format. Use .txt or .pdf", "unsupported file type: no extension")
		return
	}
	ext := strings.ToLower(header.Filename[dotIdx:])
	var extractedText string

	log.Printf("[NOTES] file type detected ext=%s", ext)

	if ext == ".txt" {
		data, err := io.ReadAll(file)
		if err != nil {
			fail(http.StatusInternalServerError, "extract", "Failed to read TXT file", "txt read failed")
			return
		}
		extractedText = string(data)
	} else if ext == ".pdf" {
		content, err := pdf.NewReader(file, header.Size)
		if err != nil {
			fail(http.StatusInternalServerError, "extract", "Failed to initialize PDF reader", "pdf reader init failed")
			return
		}

		var buf bytes.Buffer
		nPages := content.NumPage()
		for i := 1; i <= nPages; i++ {
			p := content.Page(i)
			if p.V.IsNull() {
				continue
			}
			s, _ := p.GetPlainText(nil)
			buf.WriteString(s)
		}
		extractedText = buf.String()
	} else {
		fail(http.StatusBadRequest, "extract", "Unsupported file format. Use .txt or .pdf", "unsupported extension")
		return
	}

	extractedText = strings.TrimSpace(extractedText)
	textLen := len(extractedText)
	log.Printf("[NOTES] extracted chars count=%d", textLen)

	if textLen < 50 {
		fail(http.StatusBadRequest, "extract", "Could not extract enough text from file", "extracted text too short")
		return
	}

	// --------------- Notes cache check ---------------
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(extractedText)))
	var cachedSummary string
	var cachedQuestions []services.GeneratedQuestion
	var usedCache bool
	if cached, ok := notesCache.Load(cacheKey); ok {
		entry := cached.(notesCacheEntry)
		log.Printf("[NOTES_CACHE_HIT] key=%s summary_len=%d questions=%d", cacheKey[:12], len(entry.Summary), len(entry.Questions))
		cachedSummary = entry.Summary
		cachedQuestions = entry.Questions
		usedCache = true
	}

	// Create Upload record
	upload := models.Upload{
		Filename:      header.Filename,
		Topic:         topic,
		Status:        "processing",
		ExtractedText: extractedText,
	}
	if err := database.DB.Create(&upload).Error; err != nil {
		fail(http.StatusInternalServerError, "save", "Failed to save upload record", "failed creating upload record")
		return
	}

	var summary string
	var aiQuestionsRaw []services.GeneratedQuestion

	if usedCache {
		summary = cachedSummary
		aiQuestionsRaw = cachedQuestions
		log.Printf("[NOTES] using cached summary+questions, skipping Gemini calls")
	} else {
		log.Printf("[NOTES] summarization started raw_length=%d", textLen)
		// Keep prompt size safe for model input and avoid giant payload failures.
		const maxSummarizeChars = 40000
		if len(extractedText) > maxSummarizeChars {
			extractedText = extractedText[:maxSummarizeChars]
		}

		// 1. Summarize
		var sumErr error
		summary, sumErr = services.SummarizeNotes(extractedText)
		if sumErr != nil {
			log.Printf("[NOTES][summarize] summarization failed filename=%s extracted_len=%d error=%v", header.Filename, textLen, sumErr)
			database.DB.Model(&upload).Update("status", "failed")
			if errors.Is(sumErr, services.ErrGeminiRateLimit) {
				log.Printf("[GEMINI_RATE_LIMIT] notes summarize blocked")
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Gemini free-tier quota exceeded. Please wait a few seconds and try again.",
					"stage": "gemini_rate_limit",
				})
				return
			}
			msg := strings.TrimSpace(sumErr.Error())
			if msg == "" {
				msg = "Failed to summarize uploaded notes"
			}
			fail(http.StatusInternalServerError, "summarize", msg, "gemini summarization failed")
			return
		}
		log.Printf("[NOTES] summarization success summary_chars=%d", len(strings.TrimSpace(summary)))
		database.DB.Model(&upload).Update("summary", summary)

		log.Printf("[NOTES] question generation started target_count=%d", count)
		// 2. Generate notes-based questions (single pass, no retry)
		var genErr error
		aiQuestionsRaw, genErr = services.GenerateQuestionsFromNotesBatched(summary, count, difficulty)
		if genErr != nil {
			log.Printf("[NOTES][generate] generation failed filename=%s extracted_len=%d summary_len=%d error=%v", header.Filename, textLen, len(strings.TrimSpace(summary)), genErr)
			database.DB.Model(&upload).Update("status", "failed")
			if errors.Is(genErr, services.ErrGeminiRateLimit) {
				log.Printf("[GEMINI_RATE_LIMIT] notes generate blocked")
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Gemini free-tier quota exceeded. Please wait a few seconds and try again.",
					"stage": "gemini_rate_limit",
				})
				return
			}
			errText := strings.ToLower(genErr.Error())
			message := "AI failed to generate questions from notes"
			if strings.Contains(errText, "parse ai-generated") {
				message = "Failed to parse AI-generated notes questions"
			}
			if strings.Contains(errText, "no valid questions generated") {
				message = "No valid questions generated from notes"
			}
			fail(http.StatusInternalServerError, "generate", message, "notes generation service returned error")
			return
		}

		// Store in cache for future identical uploads
		notesCache.Store(cacheKey, notesCacheEntry{Summary: summary, Questions: aiQuestionsRaw})
		log.Printf("[NOTES] cached summary+questions key=%s", cacheKey[:12])
	}

	log.Printf("[NOTES_AI] generated_raw=%d", len(aiQuestionsRaw))

	notesValidated := sanitizeNotesQuestions(aiQuestionsRaw)
	log.Printf("[NOTES_AI] after_validation=%d", len(notesValidated))

	notesDeduped := dedupeGeneratedQuestionsByPrompt(notesValidated, nil)
	log.Printf("[NOTES_AI] after_dedup=%d", len(notesDeduped))

	if len(notesDeduped) == 0 {
		log.Printf("[NOTES][generate] no valid questions after validation filename=%s extracted_len=%d parsed_count=%d valid_count=%d", header.Filename, textLen, len(aiQuestionsRaw), len(notesDeduped))
		database.DB.Model(&upload).Update("status", "failed")
		fail(http.StatusInternalServerError, "generate", "No valid questions generated from notes", "validation rejected all generated questions")
		return
	}

	// 3. Build notes-first final list and append vault only for remaining gap.
	finalCandidates := make([]notesCandidate, 0, count)
	for _, q := range notesDeduped {
		if len(finalCandidates) >= count {
			break
		}
		finalCandidates = append(finalCandidates, notesCandidate{
			Question: q,
			Source:   "notes",
		})
	}

	remaining := count - len(finalCandidates)
	vaultAdded := 0
	if remaining > 0 {
		var vaultRows []models.TrainingQuestion
		vaultQuery := database.DB.Where("topic = ? AND source <> ?", topic, "notes")
		if difficulty != "" {
			vaultQuery = vaultQuery.Where("difficulty = ?", difficulty)
		}
		if err := vaultQuery.Order("RANDOM()").Limit(remaining + 20).Find(&vaultRows).Error; err != nil {
			log.Printf("[NOTES][save] vault query failed topic=%s difficulty=%s error=%v", topic, difficulty, err)
		}

		seen := make(map[string]struct{}, len(finalCandidates))
		for _, fc := range finalCandidates {
			seen[database.NormalizePrompt(fc.Question.Prompt)] = struct{}{}
		}
		for _, row := range vaultRows {
			if remaining <= 0 {
				break
			}
			var optArr []string
			if err := json.Unmarshal([]byte(row.Options), &optArr); err != nil {
				log.Printf("[NOTES][save] vault option parse failed id=%d error=%v", row.ID, err)
				continue
			}
			q := services.GeneratedQuestion{
				Type:        row.Type,
				Prompt:      row.Prompt,
				Options:     optArr,
				Answer:      row.Answer,
				Explanation: row.Explanation,
				Difficulty:  row.Difficulty,
			}
			norm := database.NormalizePrompt(q.Prompt)
			if _, exists := seen[norm]; exists {
				continue
			}
			seen[norm] = struct{}{}
			src := strings.TrimSpace(row.Source)
			if src == "" || src == "ai" || src == "seeded" {
				src = "vault"
			}
			finalCandidates = append(finalCandidates, notesCandidate{
				Question: q,
				Source:   src,
				DBID:     row.ID,
			})
			remaining--
			vaultAdded++
		}
	}
	log.Printf("[NOTES_AI] vault_added=%d", vaultAdded)

	// Final dedupe guard.
	finalCandidates = dedupeFinalCandidatesByPrompt(finalCandidates)
	if len(finalCandidates) > count {
		finalCandidates = finalCandidates[:count]
	}

	notesSourceCount := 0
	vaultSourceCount := 0
	for _, fc := range finalCandidates {
		if fc.Source == "notes" {
			notesSourceCount++
		} else {
			vaultSourceCount++
		}
	}
	log.Printf("[NOTES_AI] final_total=%d", len(finalCandidates))
	log.Printf("[NOTES_AI] notes_source_count=%d", notesSourceCount)
	log.Printf("[NOTES_AI] vault_source_count=%d", vaultSourceCount)
	log.Printf("[NOTES_FINAL] requested=%d notes=%d vault=%d", count, notesSourceCount, vaultSourceCount)

	// 4. Persist and create session
	var questionIDs []uint
	responseQuestions := make([]gin.H, 0, len(finalCandidates))
	for _, fc := range finalCandidates {
		q := fc.Question
		if fc.Source == "notes" {
			optJSON, _ := json.Marshal(q.Options)
			tq := models.TrainingQuestion{
				Topic:       topic,
				Type:        q.Type,
				Difficulty:  q.Difficulty,
				Prompt:      q.Prompt,
				Options:     string(optJSON),
				Answer:      q.Answer,
				Explanation: q.Explanation,
				Source:      "notes",
			}
			if err := database.DB.Create(&tq).Error; err != nil {
				log.Printf("[NOTES][save] notes question save failed prompt=%q error=%v", q.Prompt, err)
				continue
			}
			questionIDs = append(questionIDs, tq.ID)
			responseQuestions = append(responseQuestions, gin.H{
				"id":          tq.ID,
				"type":        tq.Type,
				"prompt":      tq.Prompt,
				"options":     q.Options,
				"answer":      tq.Answer,
				"explanation": tq.Explanation,
				"difficulty":  tq.Difficulty,
				"source":      "notes",
			})
			continue
		}

		// Vault fallback keeps existing DB IDs and explicit vault-style source.
		sourceLabel := fc.Source
		if sourceLabel == "" {
			sourceLabel = "vault"
		}
		if fc.DBID == 0 {
			// Defensive fallback path if no DBID is present.
			optJSON, _ := json.Marshal(q.Options)
			tq := models.TrainingQuestion{
				Topic:       topic,
				Type:        q.Type,
				Difficulty:  q.Difficulty,
				Prompt:      q.Prompt,
				Options:     string(optJSON),
				Answer:      q.Answer,
				Explanation: q.Explanation,
				Source:      sourceLabel,
			}
			if err := database.DB.Create(&tq).Error; err != nil {
				log.Printf("[NOTES][save] vault fallback save failed prompt=%q error=%v", q.Prompt, err)
				continue
			}
			fc.DBID = tq.ID
		}
		questionIDs = append(questionIDs, fc.DBID)
		responseQuestions = append(responseQuestions, gin.H{
			"id":          fc.DBID,
			"type":        q.Type,
			"prompt":      q.Prompt,
			"options":     q.Options,
			"answer":      q.Answer,
			"explanation": q.Explanation,
			"difficulty":  q.Difficulty,
			"source":      sourceLabel,
		})
	}

	if len(questionIDs) == 0 {
		log.Printf("[NOTES][save] no persisted questions filename=%s extracted_len=%d summary_len=%d parsed_count=%d valid_count=%d", header.Filename, textLen, len(summary), len(aiQuestionsRaw), len(notesDeduped))
		database.DB.Model(&upload).Update("status", "failed")
		fail(http.StatusInternalServerError, "save", "Failed to save generated questions", "no questions persisted")
		return
	}

	sessionID := uuid.New().String()
	qIDsJSON, _ := json.Marshal(questionIDs)
	session := models.TrainingSession{
		SessionID:   sessionID,
		Topic:       topic,
		QuestionIDs: string(qIDsJSON),
		Status:      "active",
		CreatedAt:   time.Now(),
	}
	if err := database.DB.Create(&session).Error; err != nil {
		log.Printf("[NOTES][session] session create fail error=%v", err)
		database.DB.Model(&upload).Update("status", "failed")
		fail(http.StatusInternalServerError, "session", "Failed to create training session from generated questions", "session insert failed")
		return
	}
	log.Printf("[NOTES] session created session_id=%s question_count=%d", sessionID, len(questionIDs))

	// Finalize upload record
	database.DB.Model(&upload).Updates(map[string]interface{}{
		"status":              "done",
		"questions_generated": len(questionIDs),
	})

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"sessionId":  sessionID,
		"summary":    summary,
		"questions":  responseQuestions,
	})
}

func sanitizeNotesQuestions(questions []services.GeneratedQuestion) []services.GeneratedQuestion {
	valid := make([]services.GeneratedQuestion, 0, len(questions))
	for _, q := range questions {
		if strings.TrimSpace(q.Prompt) == "" ||
			len(q.Options) < 4 ||
			strings.TrimSpace(q.Answer) == "" ||
			strings.TrimSpace(q.Explanation) == "" ||
			strings.TrimSpace(q.Difficulty) == "" {
			continue
		}
		valid = append(valid, q)
	}
	return valid
}

func dedupeGeneratedQuestionsByPrompt(questions []services.GeneratedQuestion, existing []services.GeneratedQuestion) []services.GeneratedQuestion {
	seen := make(map[string]struct{}, len(questions)+len(existing))
	for _, q := range existing {
		seen[database.NormalizePrompt(q.Prompt)] = struct{}{}
	}
	out := make([]services.GeneratedQuestion, 0, len(questions))
	for _, q := range questions {
		norm := database.NormalizePrompt(q.Prompt)
		if _, exists := seen[norm]; exists {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, q)
	}
	return append(existing, out...)
}

func dedupeFinalCandidatesByPrompt(candidates []notesCandidate) []notesCandidate {
	seen := make(map[string]struct{}, len(candidates))
	out := make([]notesCandidate, 0, len(candidates))
	for _, c := range candidates {
		norm := database.NormalizePrompt(c.Question.Prompt)
		if _, exists := seen[norm]; exists {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, c)
	}
	return out
}

type VerifyRequest struct {
	QuestionID       string `json:"questionId"`
	SelectedOptionID string `json:"selectedOptionId,omitempty"`
	WrittenAnswer    string `json:"writtenAnswer,omitempty"`
}

type VerifyResponse struct {
	IsCorrect     bool   `json:"isCorrect"`
	Feedback      string `json:"feedback"`
	Explanation   string `json:"explanation"`
	CorrectOption string `json:"correctOptionId,omitempty"`
}

func VerifyAnswer(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification data"})
		return
	}

	// First try TrainingQuestion (AI/Seeded for Training Module)
	var tq models.TrainingQuestion
	err := database.DB.Where("id = ?", req.QuestionID).First(&tq).Error
	if err == nil {
		// Found in Training Table
		handleTrainingVerification(c, tq, req)
		return
	}

	// Fallback to Arena Question (Vault)
	var mq models.Question
	if err := database.DB.Preload("Options").Where("id = ?", req.QuestionID).First(&mq).Error; err != nil {
		log.Printf("[Verify] Question sinkhole detected: ID=%s", req.QuestionID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Neural Link Compromised. Question mapping failed."})
		return
	}

	handleArenaVerification(c, mq, req)
}

func NormalizeAnswer(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Collapse repeated whitespace
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	// Remove accidental newline differences
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func handleTrainingVerification(c *gin.Context, q models.TrainingQuestion, req VerifyRequest) {
	resp := VerifyResponse{
		Explanation: q.Explanation,
	}

	userRaw := req.SelectedOptionID
	if userRaw == "" {
		userRaw = req.WrittenAnswer
	}

	log.Printf("[Verify] question_id=%d type=%s", q.ID, q.Type)

	if q.Type == "mcq" {
		var optArr []string
		json.Unmarshal([]byte(q.Options), &optArr)

		// Resolve synthetic ID to text if needed
		userText := userRaw
		if strings.HasPrefix(userRaw, "OPT_") && len(optArr) > 0 {
			parts := strings.Split(userRaw, "_")
			if len(parts) >= 3 {
				idxStr := parts[len(parts)-1]
				if idx, err := strconv.Atoi(idxStr); err == nil && idx >= 0 && idx < len(optArr) {
					userText = optArr[idx]
				}
			}
		}

		normalizedUser := NormalizeAnswer(userText)
		normalizedCorrect := NormalizeAnswer(q.Answer)
		isCorrect := normalizedUser == normalizedCorrect

		resp.IsCorrect = isCorrect
		resp.CorrectOption = q.Answer
		if isCorrect {
			resp.Feedback = "Correct! Well done."
		} else {
			resp.Feedback = "Incorrect. The correct answer is: " + q.Answer
		}
	} else if q.Type == "coding" {
		// Run code against test cases
		var testCases []models.TestCase
		if err := json.Unmarshal([]byte(q.TestCases), &testCases); err != nil {
			log.Printf("[Verify ERROR] Failed to unmarshal test cases: %v", err)
			resp.IsCorrect = false
			resp.Feedback = "System Error: Invalid test cases for this challenge."
		} else if len(testCases) == 0 {
			resp.IsCorrect = true // Auto-pass if no test cases (shouldn't happen)
			resp.Feedback = "No test cases found. Passed by default."
		} else {
			svc := judge.GetService()
			passedCount := 0
			// Use a default language if not specified (backend usually defaults to python)
			// Ideally the frontend should send the language, but we'll fallback to "python"
			lang := "python" 

			for _, tc := range testCases {
				execResult, err := svc.Execute(req.WrittenAnswer, lang, tc.Input, 2000)
				if err == nil && judge.Normalize(execResult.Output) == judge.Normalize(tc.ExpectedOutput) {
					passedCount++
				}
			}

			resp.IsCorrect = passedCount == len(testCases)
			if resp.IsCorrect {
				resp.Feedback = fmt.Sprintf("All %d test cases passed! Mastery confirmed.", len(testCases))
			} else {
				resp.Feedback = fmt.Sprintf("Failed: %d/%d test cases passed. Keep debugging!", passedCount, len(testCases))
			}
		}
	} else {
		// Subjective evaluation
		aiEval, err := services.EvaluateAnswer(q.Prompt, q.Answer, req.WrittenAnswer, 10)
		if err != nil {
			resp.IsCorrect = false
			resp.Feedback = "AI evaluation unavailable."
		} else {
			resp.IsCorrect = aiEval.IsCorrect
			resp.Feedback = aiEval.Feedback
		}
	}
	c.JSON(http.StatusOK, resp)
}

func handleArenaVerification(c *gin.Context, q models.Question, req VerifyRequest) {
	resp := VerifyResponse{
		Explanation: q.Explanation,
	}

	isCorrect := false
	for _, opt := range q.Options {
		if opt.IsCorrect {
			resp.CorrectOption = opt.ID
		}
		if opt.ID == req.SelectedOptionID && opt.IsCorrect {
			isCorrect = true
		}
	}
	resp.IsCorrect = isCorrect
	if isCorrect {
		resp.Feedback = "Vault Access Confirmed."
	} else {
		resp.Feedback = "Access Denied. Incorrect credentials."
	}
	c.JSON(http.StatusOK, resp)
}

type GenerateRequest struct {
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
	Count      int    `json:"count"`
}

func GenerateTrainingSession(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid generation parameters"})
		return
	}

	log.Printf("[GENERATE] received: %+v", req)

	if strings.TrimSpace(req.Topic) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic specification required"})
		return
	}

	log.Printf("[AI_GEN_ROUTE] Hit topic=%q difficulty=%q count=%d", req.Topic, req.Difficulty, req.Count)

	requestedCount := req.Count
	if requestedCount <= 0 {
		requestedCount = 10 // Default
	}
	minAI := (requestedCount * 6) / 10 // 60% floor

	var finalQuestions []services.GeneratedQuestion
	seenPrompts := make(map[string]bool)
	var hitRateLimit bool

	// Telemetry Tracker
	telemetry := struct {
		Requested                int
		AIFirst                  int
		AITotal                  int
		FallbackAdded            int
		RemovedDuplicates        int
		RepeatedPromptRejections int
	}{Requested: requestedCount}

	// 1. STAGE 1: Single-pass AI Generation (no retry to save quota)
	aiRes, err := services.GenerateQuestionsBatched(req.Topic, req.Difficulty, requestedCount, nil)
	if err == nil {
		for _, q := range aiRes {
			norm := database.NormalizePrompt(q.Prompt)
			if seenPrompts[norm] {
				telemetry.RemovedDuplicates++
				continue
			}
			if database.CheckPromptExists(req.Topic, q.Prompt) {
				telemetry.RepeatedPromptRejections++
				continue
			}
			seenPrompts[norm] = true
			finalQuestions = append(finalQuestions, q)
		}
		telemetry.AIFirst = len(finalQuestions)
	} else {
		if errors.Is(err, services.ErrGeminiRateLimit) {
			log.Printf("[GEMINI_RATE_LIMIT] GenerateTrainingSession blocked")
			hitRateLimit = true
		} else {
			log.Printf("[AI] ERROR: Generation call failed: %v", err)
		}
	}

	telemetry.AITotal = len(finalQuestions)
	if telemetry.AITotal < minAI {
		log.Printf("[AI_QUOTA_FAIL] requested=%d min_ai=%d actual_ai=%d", requestedCount, minAI, telemetry.AITotal)
	}

	// 3. STAGE 3: Database Fallback (Top-up only)
	if len(finalQuestions) < requestedCount {
		gap := requestedCount - len(finalQuestions)
		log.Printf("[FALLBACK] Pulling %d questions from DB vault", gap)

		dbQuestions, _ := database.GetQuestions(req.Topic, req.Difficulty, gap+10) // fetch extra for dedup
		fallbackCount := 0
		for _, dbq := range dbQuestions {
			if fallbackCount >= gap {
				break
			}
			norm := database.NormalizePrompt(dbq.Prompt)
			if seenPrompts[norm] {
				continue
			}

			seenPrompts[norm] = true
			var optArr []string
			json.Unmarshal([]byte(dbq.Options), &optArr)

			finalQuestions = append(finalQuestions, services.GeneratedQuestion{
				Type:        dbq.Type,
				Prompt:      dbq.Prompt,
				Options:     optArr,
				Answer:      dbq.Answer,
				Explanation: dbq.Explanation,
				Difficulty:  dbq.Difficulty,
			})
			fallbackCount++
		}
		telemetry.FallbackAdded = fallbackCount
	}

	// 4. Persistence & Session Creation
	questionIDs := []uint{}
	for _, q := range finalQuestions {
		// Check if we need to save (only if from AI and not in DB)
		// Actually, to keep it simple, we check if it exists in DB by prompt
		var existingID uint
		err := database.DB.Model(&models.TrainingQuestion{}).
			Select("id").
			Where("topic = ? AND LOWER(TRIM(prompt)) = ?", req.Topic, database.NormalizePrompt(q.Prompt)).
			Scan(&existingID).Error

		if err == nil && existingID > 0 {
			questionIDs = append(questionIDs, existingID)
		} else {
			// Save new AI question
			optJSON, _ := json.Marshal(q.Options)
			tq := models.TrainingQuestion{
				Topic:       req.Topic,
				Type:        q.Type,
				Difficulty:  q.Difficulty,
				Prompt:      q.Prompt,
				Options:     string(optJSON),
				Answer:      q.Answer,
				Explanation: q.Explanation,
				Source:      "ai",
			}
			if err := database.DB.Create(&tq).Error; err == nil {
				questionIDs = append(questionIDs, tq.ID)
			}
		}
	}

	// Logging Telemetry
	log.Printf("[COUNT] requested=%d ai_first=%d ai_total=%d fallback_added=%d final_returned=%d",
		telemetry.Requested, telemetry.AIFirst, telemetry.AITotal, telemetry.FallbackAdded, len(questionIDs))
	log.Printf("[DEDUP] removed_duplicates=%d", telemetry.RemovedDuplicates)
	log.Printf("[VARIETY] repeated_prompt_rejections=%d", telemetry.RepeatedPromptRejections)

	if len(questionIDs) == 0 {
		if hitRateLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Gemini free-tier quota exceeded. Please wait a few seconds and try again.",
				"stage": "gemini_rate_limit",
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to assemble session. Logic vault is empty."})
		return
	}

	// Create Session
	sessionID := uuid.New().String()
	qIDsJSON, _ := json.Marshal(questionIDs)
	session := models.TrainingSession{
		SessionID:   sessionID,
		Topic:       req.Topic,
		QuestionIDs: string(qIDsJSON),
		Status:      "active",
		CreatedAt:   time.Now(),
	}
	database.DB.Create(&session)

	// Build Response
	responseQuestions := make([]gin.H, len(finalQuestions))
	for i, q := range finalQuestions {
		responseQuestions[i] = gin.H{
			"id":          questionIDs[i], // Matching index order
			"topic":       req.Topic,
			"type":        q.Type,
			"difficulty":  q.Difficulty,
			"prompt":      q.Prompt,
			"options":     q.Options,
			"answer":      q.Answer,
			"explanation": q.Explanation,
			"source":      "ai",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"sessionId":  sessionID,
		"topic":      req.Topic,
		"count":      len(responseQuestions),
		"questions":  responseQuestions,
	})
}

func GetTrainingSession(c *gin.Context) {
	sessionID := c.Param("id")

	session, questions, err := database.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session module missing or corrupted"})
		return
	}

	// Map to response format
	responseQuestions := make([]any, len(questions))
	for i, q := range questions {
		var optArr []string
		json.Unmarshal([]byte(q.Options), &optArr)

		responseQuestions[i] = gin.H{
			"id":          q.ID,
			"prompt":      q.Prompt,
			"type":        q.Type,
			"difficulty":  q.Difficulty,
			"options":     optArr,
			"explanation": q.Explanation,
			"source":      q.Source,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": session.SessionID,
		"topic":     session.Topic,
		"status":    session.Status,
		"questions": responseQuestions,
	})
}

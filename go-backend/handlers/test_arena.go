package handlers

import (
	"backend/database"
	"backend/judge"
	"backend/models"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// broadcastLeaderboard triggers a WebSocket broadcast for a test.
// Safe to call even if LeaderboardHub is nil (e.g. during tests).
func broadcastLeaderboard(testID string) {
	if LeaderboardHub != nil {
		go LeaderboardHub.Broadcast(testID)
	}
}

// isTestExpired checks if a test's time window has closed.
// Returns (remainingSeconds, expired).
func isTestExpired(testID string) (int, bool) {
	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		return 0, true
	}
	elapsed := time.Since(test.StartTime)
	remaining := test.DurationSeconds - int(elapsed.Seconds())
	if remaining <= 0 {
		return 0, true
	}
	return remaining, false
}

// ──────────────────────────────────────────────
// GetActiveTest → GET /api/arena/active
// Returns the single currently-active test with timing metadata.
// ──────────────────────────────────────────────
func GetActiveTest(c *gin.Context) {
	var test models.Test
	if err := database.DB.Where("isActive = ? AND isPublished = ?", true, true).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active test"})
		return
	}

	// Count questions
	var questionCount int64
	database.DB.Model(&models.TestQuestion{}).Where("testId = ?", test.ID).Count(&questionCount)

	// Determine status
	now := time.Now()
	elapsed := now.Sub(test.StartTime)
	elapsedSeconds := int(elapsed.Seconds())
	remainingSeconds := test.DurationSeconds - elapsedSeconds

	status := "upcoming"
	if now.Before(test.StartTime) {
		status = "upcoming"
		elapsedSeconds = 0
		remainingSeconds = test.DurationSeconds
	} else if remainingSeconds > 0 {
		status = "live"
	} else {
		status = "ended"
		remainingSeconds = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"test":             test,
		"questionCount":    questionCount,
		"status":           status,
		"remainingSeconds": remainingSeconds,
		"elapsedSeconds":   elapsedSeconds,
	})
}

// ──────────────────────────────────────────────
// ListPublishedTests → GET /api/arena/tests
// Returns all published tests.
// ──────────────────────────────────────────────
func ListPublishedTests(c *gin.Context) {
	var tests []models.Test
	if err := database.DB.Where("isPublished = ?", true).
		Order("startTime desc").Find(&tests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tests"})
		return
	}
	c.JSON(http.StatusOK, tests)
}

// ──────────────────────────────────────────────
// JoinTest → POST /api/arena/tests/:id/join
// Returns existing attempt if user already joined.
// Returns 403 if test time has expired.
// ──────────────────────────────────────────────
func JoinTest(c *gin.Context) {
	testID := c.Param("id")
	userID, _ := c.Get("userID")

	var test models.Test
	if err := database.DB.Where("id = ? AND isPublished = ?", testID, true).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	// Calculate remaining time
	elapsed := time.Since(test.StartTime)
	remainingSeconds := test.DurationSeconds - int(elapsed.Seconds())

	if remainingSeconds <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test has ended"})
		return
	}

	// Check if user already joined — return existing attempt
	var existing models.TestAttempt
	if err := database.DB.Where("userId = ? AND testId = ?", userID, testID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"attempt":          existing,
			"remainingSeconds": remainingSeconds,
		})
		return
	}

	// Create new attempt
	attempt := models.TestAttempt{
		ID:        uuid.New().String(),
		UserID:    userID.(string),
		TestID:    testID,
		StartedAt: time.Now(),
	}

	if err := database.DB.Create(&attempt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attempt":          attempt,
		"remainingSeconds": remainingSeconds,
	})
}

// ──────────────────────────────────────────────
// GetTestAttempt → GET /api/arena/attempts/:id
// Returns questions + remaining_time + existing submissions.
// SECURITY: hidden testcase data stripped before response.
// SECURITY: MCQ isCorrect stripped before response.
// ──────────────────────────────────────────────
func GetTestAttempt(c *gin.Context) {
	attemptID := c.Param("id")
	userID, _ := c.Get("userID")

	var attempt models.TestAttempt
	if err := database.DB.Preload("Test").Where("id = ? AND userId = ?", attemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Calculate remaining time
	elapsed := time.Since(attempt.Test.StartTime)
	remainingSeconds := attempt.Test.DurationSeconds - int(elapsed.Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	// Fetch questions with associations
	var questions []models.TestQuestion
	database.DB.Preload("MCQOptions").Preload("CodingDetail").Preload("TestCases").
		Where("testId = ?", attempt.TestID).Order("position asc").Find(&questions)

	// SECURITY: hidden testcase data stripped before response
	for i := range questions {
		// Strip isCorrect from MCQ options
		for j := range questions[i].MCQOptions {
			questions[i].MCQOptions[j].IsCorrect = false
		}
		// Keep only sample (non-hidden) testcases, strip hidden ones entirely
		var visibleCases []models.TestCase
		for _, tc := range questions[i].TestCases {
			if !tc.IsHidden {
				visibleCases = append(visibleCases, tc)
			}
		}
		questions[i].TestCases = visibleCases
	}

	// Fetch existing submissions for this attempt
	var submissions []models.TestSubmission
	database.DB.Where("attemptId = ?", attemptID).Find(&submissions)

	c.JSON(http.StatusOK, gin.H{
		"attempt":          attempt,
		"questions":        questions,
		"submissions":      submissions,
		"remainingSeconds": remainingSeconds,
	})
}

// ──────────────────────────────────────────────
// SaveMCQ → POST /api/arena/submissions/mcq
// Upserts the user's MCQ answer for a question.
// ──────────────────────────────────────────────
func SaveMCQ(c *gin.Context) {
	userID, exists := c.Get("userID")
	log.Printf("[SUBMIT-MCQ] JWT userID=%v exists=%v", userID, exists)
	if !exists || userID == nil || userID == "" {
		log.Println("[SUBMIT-MCQ] CRITICAL: userID missing from JWT context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenticated userId missing from token"})
		return
	}

	var req struct {
		AttemptID        string `json:"attemptId" binding:"required"`
		QuestionID       string `json:"questionId" binding:"required"`
		SelectedOptionID string `json:"selectedOptionId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attemptId, questionId, and selectedOptionId are required"})
		return
	}
	log.Printf("[SUBMIT-MCQ] userID=%v attemptId=%s questionId=%s optionId=%s", userID, req.AttemptID, req.QuestionID, req.SelectedOptionID)

	// Verify attempt belongs to user
	var attempt models.TestAttempt
	if err := database.DB.Where("id = ? AND userId = ?", req.AttemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Time guard — reject if test window has closed
	if _, expired := isTestExpired(attempt.TestID); expired {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test time has expired"})
		return
	}

	// Upsert: check if submission already exists
	var existing models.TestSubmission
	if err := database.DB.Where("attemptId = ? AND questionId = ?", req.AttemptID, req.QuestionID).First(&existing).Error; err == nil {
		// FIX: targeted update instead of Save() to avoid overwriting unrelated columns
		database.DB.Model(&existing).Updates(map[string]interface{}{"selectedOptionId": req.SelectedOptionID})
		existing.SelectedOptionID = req.SelectedOptionID
		c.JSON(http.StatusOK, existing)
		return
	}

	// Create new submission
	submission := models.TestSubmission{
		ID:               uuid.New().String(),
		AttemptID:        req.AttemptID,
		QuestionID:       req.QuestionID,
		Type:             "mcq",
		SelectedOptionID: req.SelectedOptionID,
	}

	if err := database.DB.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answer"})
		return
	}

	c.JSON(http.StatusOK, submission)
}

// ──────────────────────────────────────────────
// GetLanguages → GET /api/arena/languages
// Returns supported languages with metadata.
// ──────────────────────────────────────────────
func GetLanguages(c *gin.Context) {
	svc := judge.GetService()
	c.JSON(http.StatusOK, gin.H{"languages": svc.GetLanguages()})
}

// ──────────────────────────────────────────────
// SaveDraft → POST /api/arena/submissions/draft
// Auto-saves code without running it.
// ──────────────────────────────────────────────
func SaveDraft(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		AttemptID  string `json:"attemptId" binding:"required"`
		QuestionID string `json:"questionId" binding:"required"`
		Code       string `json:"code" binding:"required"`
		Language   string `json:"language" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attemptId, questionId, code, and language are required"})
		return
	}

	// Verify attempt belongs to user
	var attempt models.TestAttempt
	if err := database.DB.Where("id = ? AND userId = ?", req.AttemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Time guard
	if _, expired := isTestExpired(attempt.TestID); expired {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test time has expired"})
		return
	}

	// Upsert draft
	var existing models.TestSubmission
	if err := database.DB.Where("attemptId = ? AND questionId = ?", req.AttemptID, req.QuestionID).First(&existing).Error; err == nil {
		// FIX: targeted update instead of Save()
		updates := map[string]interface{}{"code": req.Code, "language": req.Language}
		if existing.Verdict == "" {
			updates["verdict"] = "draft"
		}
		database.DB.Model(&existing).Updates(updates)
		c.JSON(http.StatusOK, gin.H{"status": "saved"})
		return
	}

	submission := models.TestSubmission{
		ID:         uuid.New().String(),
		AttemptID:  req.AttemptID,
		QuestionID: req.QuestionID,
		Type:       "coding",
		Code:       req.Code,
		Language:   req.Language,
		Verdict:    "draft",
	}
	database.DB.Create(&submission)
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

// ──────────────────────────────────────────────
// RunCode → POST /api/arena/submissions/run
// Runs code against SAMPLE testcases only.
// Returns per-case results with rich error info.
// Does NOT save to submissions table.
// ──────────────────────────────────────────────
func RunCode(c *gin.Context) {
	var req struct {
		AttemptID  string `json:"attemptId" binding:"required"`
		QuestionID string `json:"questionId" binding:"required"`
		Code       string `json:"code" binding:"required"`
		Language   string `json:"language" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attemptId, questionId, code, and language are required"})
		return
	}

	// Time guard — reject if test window has closed
	var runAttempt models.TestAttempt
	if err := database.DB.Where("id = ?", req.AttemptID).First(&runAttempt).Error; err == nil {
		if _, expired := isTestExpired(runAttempt.TestID); expired {
			c.JSON(http.StatusForbidden, gin.H{"error": "Test time has expired"})
			return
		}
	}

	// Fetch question with coding detail and testcases
	var question models.TestQuestion
	if err := database.DB.Preload("CodingDetail").Preload("TestCases").Where("id = ?", req.QuestionID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	timeLimitMs := 2000
	if question.CodingDetail != nil && question.CodingDetail.TimeLimitMs > 0 {
		timeLimitMs = question.CodingDetail.TimeLimitMs
	}

	svc := judge.GetService()

	type CaseResult struct {
		Input      string `json:"input"`
		Expected   string `json:"expected"`
		Actual     string `json:"actual"`
		Pass       bool   `json:"pass"`
		ErrorType  string `json:"errorType,omitempty"`
		CompileOut string `json:"compileOut,omitempty"`
		DurationMs int64  `json:"durationMs"`
	}

	var results []CaseResult
	var hasCompileError bool
	var compileOutput string

	for _, tc := range question.TestCases {
		// SECURITY: hidden testcase data stripped before response — skip hidden cases entirely
		if tc.IsHidden {
			continue
		}

		execResult, err := svc.Execute(req.Code, req.Language, tc.Input, timeLimitMs)
		actual := execResult.Output
		if err != nil {
			actual = "Execution error: " + err.Error()
		}

		// Detect compile error — only need to report once
		if execResult.ErrorType == "compilation_error" {
			hasCompileError = true
			compileOutput = execResult.CompileOut
		}

		pass := strings.TrimSpace(actual) == strings.TrimSpace(tc.ExpectedOutput)
		results = append(results, CaseResult{
			Input:      tc.Input,
			Expected:   tc.ExpectedOutput,
			Actual:     actual,
			Pass:       pass,
			ErrorType:  execResult.ErrorType,
			CompileOut: execResult.CompileOut,
			DurationMs: execResult.DurationMs,
		})

		// If compile error, skip remaining cases
		if hasCompileError {
			break
		}
	}

	response := gin.H{"results": results}
	if hasCompileError {
		response["error"] = "compilation_error"
		response["compileOutput"] = compileOutput
	}

	c.JSON(http.StatusOK, response)
}

// ──────────────────────────────────────────────
// SubmitCode → POST /api/arena/submissions/code
// Runs code against ALL testcases.
// Sample cases: full detail (input, expected, actual).
// Hidden cases: pass/fail only, NO input/output.
// Saves result to submissions table (upsert).
// ──────────────────────────────────────────────
func SubmitCode(c *gin.Context) {
	userID, exists := c.Get("userID")
	log.Printf("[SUBMIT-CODE] JWT userID=%v exists=%v", userID, exists)
	if !exists || userID == nil || userID == "" {
		log.Println("[SUBMIT-CODE] CRITICAL: userID missing from JWT context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenticated userId missing from token"})
		return
	}

	var req struct {
		AttemptID  string `json:"attemptId" binding:"required"`
		QuestionID string `json:"questionId" binding:"required"`
		Code       string `json:"code" binding:"required"`
		Language   string `json:"language" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attemptId, questionId, code, and language are required"})
		return
	}
	log.Printf("[SUBMIT-CODE] userID=%v attemptId=%s questionId=%s lang=%s", userID, req.AttemptID, req.QuestionID, req.Language)

	// Verify attempt belongs to user
	var attempt models.TestAttempt
	if err := database.DB.Where("id = ? AND userId = ?", req.AttemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Time guard — reject if test window has closed
	if _, expired := isTestExpired(attempt.TestID); expired {
		c.JSON(http.StatusForbidden, gin.H{"error": "Test time has expired"})
		return
	}

	// Fetch question with coding detail and testcases
	var question models.TestQuestion
	if err := database.DB.Preload("CodingDetail").Preload("TestCases").Where("id = ?", req.QuestionID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	timeLimitMs := 2000
	if question.CodingDetail != nil && question.CodingDetail.TimeLimitMs > 0 {
		timeLimitMs = question.CodingDetail.TimeLimitMs
	}

	svc := judge.GetService()

	type CaseResult struct {
		Input      string `json:"input,omitempty"`
		Expected   string `json:"expected,omitempty"`
		Actual     string `json:"actual,omitempty"`
		Pass       bool   `json:"pass"`
		Hidden     bool   `json:"hidden"`
		ErrorType  string `json:"errorType,omitempty"`
		DurationMs int64  `json:"durationMs,omitempty"`
	}

	var results []CaseResult
	passedCount := 0
	totalCount := len(question.TestCases)
	var hasCompileError bool
	var compileOutput string

	for _, tc := range question.TestCases {
		execResult, err := svc.Execute(req.Code, req.Language, tc.Input, timeLimitMs)
		actual := execResult.Output
		if err != nil {
			actual = "Execution error: " + err.Error()
		}

		// Detect compile error
		if execResult.ErrorType == "compilation_error" {
			hasCompileError = true
			compileOutput = execResult.CompileOut
		}

		pass := strings.TrimSpace(actual) == strings.TrimSpace(tc.ExpectedOutput)
		if pass {
			passedCount++
		}

		cr := CaseResult{Pass: pass, Hidden: tc.IsHidden, ErrorType: execResult.ErrorType, DurationMs: execResult.DurationMs}

		// SECURITY: hidden testcase data stripped before response
		if !tc.IsHidden {
			cr.Input = tc.Input
			cr.Expected = tc.ExpectedOutput
			cr.Actual = actual
		}

		results = append(results, cr)

		// If compile error, skip remaining cases
		if hasCompileError {
			break
		}
	}

	// Determine verdict
	verdict := "wrong_answer"
	if hasCompileError {
		verdict = "compilation_error"
	} else if passedCount == totalCount && totalCount > 0 {
		verdict = "accepted"
	}

	// Calculate score: points * (passed / total)
	score := 0
	if totalCount > 0 {
		score = int(math.Round(float64(question.Points) * float64(passedCount) / float64(totalCount)))
	}

	// Upsert submission
	var existing models.TestSubmission
	if err := database.DB.Where("attemptId = ? AND questionId = ?", req.AttemptID, req.QuestionID).First(&existing).Error; err == nil {
		// FIX: targeted update instead of Save() to avoid overwriting unrelated columns
		database.DB.Model(&existing).Updates(map[string]interface{}{
			"code":        req.Code,
			"language":    req.Language,
			"verdict":     verdict,
			"passedCount": passedCount,
			"totalCount":  totalCount,
			"score":       score,
		})
	} else {
		// Create new submission
		submission := models.TestSubmission{
			ID:               uuid.New().String(),
			AttemptID:        req.AttemptID,
			QuestionID:       req.QuestionID,
			Type:             "coding",
			Code:             req.Code,
			Language:         req.Language,
			Verdict:          verdict,
			PassedCount:      passedCount,
			TotalCount:       totalCount,
			Score:            score,
		}
		database.DB.Create(&submission)
	}

	// Broadcast updated leaderboard to all WebSocket clients for this test
	broadcastLeaderboard(attempt.TestID)

	response := gin.H{
		"verdict":     verdict,
		"passedCount": passedCount,
		"totalCount":  totalCount,
		"results":     results,
	}
	if hasCompileError {
		response["compileOutput"] = compileOutput
	}

	c.JSON(http.StatusOK, response)
}

// ──────────────────────────────────────────────
// SubmitTestAttempt → POST /api/arena/attempts/:id/submit
// Calculates final score and closes the attempt.
// MCQ: full points if correct option selected.
// Coding: points * (passed/total) — already scored in SubmitCode.
// ──────────────────────────────────────────────
func SubmitTestAttempt(c *gin.Context) {
	attemptID := c.Param("id")
	userID, exists := c.Get("userID")
	log.Printf("[SUBMIT-ATTEMPT] JWT userID=%v exists=%v attemptId=%s", userID, exists, attemptID)
	if !exists || userID == nil || userID == "" {
		log.Println("[SUBMIT-ATTEMPT] CRITICAL: userID missing from JWT context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenticated userId missing from token"})
		return
	}

	// ── Begin transaction to prevent race with auto-submit goroutine ──
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("[SUBMIT-ATTEMPT] failed to begin tx: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	// Re-read attempt inside transaction to prevent race
	var attempt models.TestAttempt
	if err := tx.Where("id = ? AND userId = ?", attemptID, userID).First(&attempt).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Prevent double-submit (checked inside transaction)
	if !attempt.SubmittedAt.IsZero() {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attempt already submitted"})
		return
	}

	// Get all questions for the test (need MCQ options for grading)
	var questions []models.TestQuestion
	tx.Preload("MCQOptions").Where("testId = ?", attempt.TestID).Find(&questions)

	// Get all submissions for this attempt
	var submissions []models.TestSubmission
	tx.Where("attemptId = ?", attemptID).Find(&submissions)

	// Build submission lookup by questionID
	subMap := make(map[string]*models.TestSubmission)
	for i := range submissions {
		subMap[submissions[i].QuestionID] = &submissions[i]
	}

	totalScore := 0

	for _, q := range questions {
		sub, exists := subMap[q.ID]
		if !exists {
			continue
		}

		if q.Type == "mcq" {
			// MCQ: full points if the selected option is correct
			for _, opt := range q.MCQOptions {
				if opt.ID == sub.SelectedOptionID && opt.IsCorrect {
					sub.Score = q.Points
					sub.Verdict = "accepted"
					break
				}
			}
			if sub.Verdict != "accepted" {
				sub.Score = 0
				sub.Verdict = "wrong_answer"
			}
			// FIX: targeted update instead of Save()
			tx.Model(sub).Updates(map[string]interface{}{
				"score":   sub.Score,
				"verdict": sub.Verdict,
			})
		}
		// Coding scores are already calculated in SubmitCode — just sum them

		totalScore += sub.Score
	}

	// Check if auto-submitted (time ran out)
	var test models.Test
	tx.Where("id = ?", attempt.TestID).First(&test)
	elapsed := time.Since(test.StartTime)
	isAutoSubmitted := int(elapsed.Seconds()) >= test.DurationSeconds

	// FIX: Use targeted Updates() instead of Save() to prevent zero-value overwrites
	submittedAt := time.Now()
	log.Printf("[DB WRITE] SubmitTestAttempt: userID=%s testID=%s attemptID=%s score=%d", attempt.UserID, attempt.TestID, attempt.ID, totalScore)
	result := tx.Model(&models.TestAttempt{}).Where("id = ?", attempt.ID).Updates(map[string]interface{}{
		"score":           totalScore,
		"totalQuestions":  len(questions),
		"timeTaken":       int(time.Since(attempt.StartedAt).Seconds()),
		"submittedAt":     submittedAt,
		"isAutoSubmitted": isAutoSubmitted,
	})
	if result.Error != nil {
		tx.Rollback()
		log.Printf("[SUBMIT-ATTEMPT] update failed: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save attempt"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[SUBMIT-ATTEMPT] commit failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit"})
		return
	}

	log.Printf("[SUBMIT-ATTEMPT] SUCCESS: attemptID=%s score=%d auto=%v", attempt.ID, totalScore, isAutoSubmitted)

	// Broadcast updated leaderboard (after commit, non-blocking)
	broadcastLeaderboard(attempt.TestID)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Attempt submitted",
		"attemptId":       attempt.ID,
		"score":           totalScore,
		"isAutoSubmitted": isAutoSubmitted,
	})
}

// ──────────────────────────────────────────────
// GetAttemptStatus → GET /api/arena/attempts/:id/status
// Lightweight polling endpoint for real-time status.
// Returns: remaining time + per-question verdict/status.
// ──────────────────────────────────────────────
func GetAttemptStatus(c *gin.Context) {
	attemptID := c.Param("id")
	userID, _ := c.Get("userID")

	var attempt models.TestAttempt
	if err := database.DB.Preload("Test").Where("id = ? AND userId = ?", attemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	// Calculate remaining time
	elapsed := time.Since(attempt.Test.StartTime)
	remainingSeconds := attempt.Test.DurationSeconds - int(elapsed.Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	isSubmitted := !attempt.SubmittedAt.IsZero()

	// Fetch questions (just ID, type, position)
	var questions []models.TestQuestion
	database.DB.Select("id, type, position").Where("testId = ?", attempt.TestID).Order("position asc").Find(&questions)

	// Fetch submissions for this attempt
	var submissions []models.TestSubmission
	database.DB.Where("attemptId = ?", attemptID).Find(&submissions)

	// Build submission lookup
	subMap := make(map[string]*models.TestSubmission)
	for i := range submissions {
		subMap[submissions[i].QuestionID] = &submissions[i]
	}

	type QuestionStatus struct {
		QuestionID  string `json:"questionId"`
		Type        string `json:"type"`
		Position    int    `json:"position"`
		Status      string `json:"status"`      // unanswered, answered, accepted, wrong_answer
		Verdict     string `json:"verdict"`     // pending, accepted, wrong_answer, time_limit, runtime_error
		PassedCount int    `json:"passedCount,omitempty"`
		TotalCount  int    `json:"totalCount,omitempty"`
	}

	var qStatuses []QuestionStatus
	for _, q := range questions {
		qs := QuestionStatus{
			QuestionID: q.ID,
			Type:       q.Type,
			Position:   q.Position,
			Status:     "unanswered",
			Verdict:    "pending",
		}

		if sub, exists := subMap[q.ID]; exists {
			if q.Type == "mcq" {
				if sub.SelectedOptionID != "" {
					qs.Status = "answered"
				}
				qs.Verdict = "pending" // MCQ verdict only known after final submit
			} else {
				// Coding
				if sub.Verdict == "accepted" {
					qs.Status = "accepted"
				} else if sub.Code != "" {
					qs.Status = "wrong_answer"
				}
				qs.Verdict = sub.Verdict
				qs.PassedCount = sub.PassedCount
				qs.TotalCount = sub.TotalCount
			}
		}

		qStatuses = append(qStatuses, qs)
	}

	c.JSON(http.StatusOK, gin.H{
		"attemptId":        attempt.ID,
		"remainingSeconds": remainingSeconds,
		"isSubmitted":      isSubmitted,
		"score":            attempt.Score,
		"questions":        qStatuses,
	})
}

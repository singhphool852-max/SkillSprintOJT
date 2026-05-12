package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TrainSessionRequest struct {
	Topic      string `json:"topic"`
	Count      int    `json:"count"`
	Difficulty string `json:"difficulty"`
}

type TrainSessionQuestionResponse struct {
	ID          uint   `json:"id"`
	Topic       string `json:"topic"`
	Type        string `json:"type"`
	Difficulty  string `json:"difficulty"`
	Prompt      string `json:"prompt"`
	Options     any    `json:"options"`
	Answer      string `json:"answer"`
	Explanation string `json:"explanation"`
	Source      string `json:"source"`
}

// CreateTrainSession handles POST /api/train/session.
// Returns seeded/AI questions from the training_questions table.
func CreateTrainSession(c *gin.Context) {
	var req TrainSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session parameters"})
		return
	}

	// Normalise topic to lowercase for DB lookup
	topic := strings.ToLower(req.Topic)
	difficulty := strings.ToLower(req.Difficulty)
	count := req.Count
	if count <= 0 {
		count = 10
	}

	log.Printf("[TrainSession] INBOUND_AUTH: topic=%s difficulty=%s count=%d", topic, difficulty, count)

	// Phase 1: Fetch randomized questions from DB
	questions, err := database.GetQuestions(topic, difficulty, count)
	if err != nil {
		log.Printf("[DB] ERROR: Fetch failure: %v", err)
		questions = []models.TrainingQuestion{}
	}
	log.Printf("[DB] fetched %d questions for topic=%s difficulty=%s", len(questions), topic, difficulty)

	// Phase 2: If deficit found, call Gemini AI
	if len(questions) < count {
		needed := count - len(questions)
		log.Printf("[AI] generating %d questions for topic=%s", needed, topic)

		aiQuestions, aiErr := services.GenerateQuestionsBatched(topic, difficulty, needed, nil)
		if aiErr != nil {
			if errors.Is(aiErr, services.ErrGeminiRateLimit) {
				log.Printf("[GEMINI_RATE_LIMIT] CreateTrainSession blocked")
				if len(questions) == 0 {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Gemini free-tier quota exceeded. Please wait a few seconds and try again.",
						"stage": "gemini_rate_limit",
					})
					return
				}
				log.Printf("[AI] Rate limited but have %d DB questions, continuing", len(questions))
			} else {
				log.Printf("[AI_FAIL] fallback triggered: error=%v", aiErr)
			}
		} else if len(aiQuestions) > 0 {
			// Convert to DB models
			var newModels []models.TrainingQuestion
			for _, q := range aiQuestions {
				optionsJSON, _ := json.Marshal(q.Options)
				newModels = append(newModels, models.TrainingQuestion{
					Topic:       topic,
					Type:        q.Type,
					Difficulty:  q.Difficulty,
					Prompt:      q.Prompt,
					Options:     string(optionsJSON),
					Answer:      q.Answer,
					Explanation: q.Explanation,
					Source:      "ai",
				})
			}

			// Save to DB (deduplicates by prompt)
			savedCount, saveErr := database.SaveQuestions(newModels)
			if saveErr != nil {
				log.Printf("[DB_SAVE] ERROR: Failed to persist AI questions: %v", saveErr)
			} else {
				log.Printf("[DB_SAVE] saved AI questions: counts=%d", savedCount)
			}

			// Refetch all questions to get new IDs
			questions, _ = database.GetQuestions(topic, difficulty, count)
			log.Printf("[AI] success: session now has %d questions", len(questions))
		}
	}

	// Phase 3: If still short, broaden to ALL difficulties for this topic
	if len(questions) < count {
		log.Printf("[TrainSession] Broadening search: topic=%s -> all difficulties (had %d, need %d)", topic, len(questions), count)
		allQuestions, _ := database.GetQuestions(topic, "", count)

		// Merge and deduplicate by ID
		existingIDs := make(map[uint]bool)
		for _, q := range questions {
			existingIDs[q.ID] = true
		}
		for _, q := range allQuestions {
			if !existingIDs[q.ID] && len(questions) < count {
				questions = append(questions, q)
				existingIDs[q.ID] = true
			}
		}
		log.Printf("[TrainSession] After broadening: %d unique questions", len(questions))
	}

	// Phase 4: Termination if no data found
	if len(questions) == 0 {
		log.Printf("[TrainSession] Zero questions available for topic=%s", topic)
		c.JSON(http.StatusNotFound, gin.H{"error": "No training questions available for this topic."})
		return
	}

	// Deduplicate final list by ID
	seenIDs := make(map[uint]bool)
	uniqueQuestions := make([]models.TrainingQuestion, 0, len(questions))
	for _, q := range questions {
		if !seenIDs[q.ID] {
			seenIDs[q.ID] = true
			uniqueQuestions = append(uniqueQuestions, q)
		}
	}
	questions = uniqueQuestions

	// Build session
	sessionID := uuid.New().String()
	questionIDs := make([]uint, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}
	idsJSON, _ := json.Marshal(questionIDs)

	session := models.TrainingSession{
		SessionID:   sessionID,
		Topic:       topic,
		QuestionIDs: string(idsJSON),
		Status:      "active",
		Score:       0,
	}

	if err := database.CreateSession(session); err != nil {
		log.Printf("[TrainSession] ERROR_SESSION_PERSIST_FAILED: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize neural session in vault."})
		return
	}

	// Build response — parse options JSON string into a real array
	responseQuestions := make([]TrainSessionQuestionResponse, len(questions))
	for i, q := range questions {
		var parsedOptions any
		if q.Options != "" && q.Options != "[]" {
			var optArr []string
			if jsonErr := json.Unmarshal([]byte(q.Options), &optArr); jsonErr == nil {
				parsedOptions = optArr
			} else {
				parsedOptions = q.Options 
			}
		} else {
			parsedOptions = []string{}
		}

		responseQuestions[i] = TrainSessionQuestionResponse{
			ID:          q.ID,
			Topic:       q.Topic,
			Type:        q.Type,
			Difficulty:  q.Difficulty,
			Prompt:      q.Prompt,
			Options:     parsedOptions,
			Answer:      q.Answer,
			Explanation: q.Explanation,
			Source:      q.Source,
		}
	}

	log.Printf("[TrainSession] SESSION_ESTABLISHED: id=%s count=%d topic=%s", sessionID, len(responseQuestions), topic)

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"sessionId":  sessionID, // CamelCase alias for AI compatibility
		"topic":      topic,
		"count":      len(responseQuestions),
		"questions":  responseQuestions,
	})
}

// GetTrainSessionDetail handles GET /api/train/session/:id.
func GetTrainSessionDetail(c *gin.Context) {
	sessionID := c.Param("id")
	log.Println("[SESSION_FETCH] id:", sessionID)

	session, questions, err := database.GetSession(sessionID)
	if err != nil {
		log.Printf("[SESSION_FETCH] not found: id=%s error=%v", sessionID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	log.Println("[SESSION_FETCH] found:", session.SessionID)
	log.Println("[SESSION_FETCH] question count:", len(questions))

	responseQuestions := make([]TrainSessionQuestionResponse, len(questions))
	for i, q := range questions {
		var parsedOptions any
		if q.Options != "" && q.Options != "[]" {
			var optArr []string
			if jsonErr := json.Unmarshal([]byte(q.Options), &optArr); jsonErr == nil {
				parsedOptions = optArr
			} else {
				parsedOptions = q.Options
			}
		} else {
			parsedOptions = []string{}
		}

		responseQuestions[i] = TrainSessionQuestionResponse{
			ID:          q.ID,
			Topic:       q.Topic,
			Type:        q.Type,
			Difficulty:  q.Difficulty,
			Prompt:      q.Prompt,
			Options:     parsedOptions,
			Answer:      q.Answer,
			Explanation: q.Explanation,
			Source:      q.Source,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": session.SessionID,
		"sessionId":  session.SessionID,
		"topic":      session.Topic,
		"status":     session.Status,
		"score":      session.Score,
		"questions":  responseQuestions,
	})
}

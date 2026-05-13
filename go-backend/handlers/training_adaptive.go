package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdaptiveSessionRequest defines parameters for starting a practice session
type AdaptiveSessionRequest struct {
	TopicID    string `json:"topicId"`    // Optional: focus on one topic
	Difficulty string `json:"difficulty"` // Optional: override adaptive difficulty
	Mode       string `json:"mode"`       // "mistakes", "adaptive", "recovery"
	AttemptID  string `json:"attemptId"`  // NEW: specific attempt to recover from
}

// ──────────────────────────────────────────────
// StartAdaptiveTraining → POST /api/training/adaptive/start
// Orchestrates the "Intelligent Adaptive Learning" logic.
// ──────────────────────────────────────────────
func StartAdaptiveTraining(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req AdaptiveSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	log.Printf("[TRAIN] Starting session: user=%s mode=%s attempt=%s topic=%s", userID, req.Mode, req.AttemptID, req.TopicID)

	// 1. Identify Weak Topics (if not specified)
	var weakTopics []models.UserTopicStats
	if req.TopicID == "" {
		database.DB.Where("userId = ? AND accuracyPercent < 70", userID).
			Order("accuracyPercent ASC").Limit(3).Find(&weakTopics)
	} else {
		var stat models.UserTopicStats
		database.DB.Where("userId = ? AND topicId = ?", userID, req.TopicID).First(&stat)
		if stat.ID != "" {
			weakTopics = append(weakTopics, stat)
		}
	}

	var sessionQuestions []models.TrainingQuestion
	targetCount := 10

	// 2. Mode: "mistakes" or "recovery" -> Focus strictly on UserWrongQuestions
	if req.Mode == "mistakes" || req.Mode == "recovery" {
		var mistakes []models.UserWrongQuestion
		query := database.DB.Where("userId = ? AND masteredAt IS NULL", userID)
		
		if req.AttemptID != "" {
			query = query.Where("attemptId = ?", req.AttemptID)
			log.Printf("[RECOVERY] Fetching mistakes specifically for attempt: %s", req.AttemptID)
		} else if req.TopicID != "" {
			query = query.Where("topicId = ?", req.TopicID)
		}
		
		// For global mistakes, prioritize worst failures. For specific recovery, show in order.
		if req.AttemptID == "" {
			query = query.Order("wrongCount DESC, reviewCount ASC, createdAt DESC")
		} else {
			query = query.Order("createdAt ASC")
		}
		
		query.Limit(targetCount).Find(&mistakes)
		log.Printf("[RECOVERY] Found %d mistake(s) for user", len(mistakes))

		for _, m := range mistakes {
			tq := &models.TrainingQuestion{
				Prompt:      m.QuestionTitle,
				Topic:       m.TopicID,
				Difficulty:  m.Difficulty,
				Type:        m.QuestionType,
				Answer:      m.CorrectAnswer,
				Source:      "recovery",
			}
			
			if m.QuestionType == "mcq" {
				var opts []models.TestMCQOption
				database.DB.Where("questionId = ?", m.QuestionID).Find(&opts)
				var optTexts []string
				for _, o := range opts {
					optTexts = append(optTexts, o.OptionText)
				}
				optJSON, _ := json.Marshal(optTexts)
				tq.Options = string(optJSON)
			}

			var existing models.TrainingQuestion
			if err := database.DB.Where("topic = ? AND prompt = ?", tq.Topic, tq.Prompt).First(&existing).Error; err == nil {
				tq.ID = existing.ID
			} else {
				database.DB.Create(tq)
			}

			if tq.ID > 0 {
				sessionQuestions = append(sessionQuestions, *tq)
			}

			// In "adaptive" mode, we generate similar variations. 
			// In strict "recovery", we ONLY show the mistakes unless requested otherwise.
			if req.Mode == "adaptive" && len(sessionQuestions) < targetCount {
				simQuestions, err := services.GenerateSimilarQuestions(m.QuestionTitle, m.TopicID, m.Difficulty, 1)
				if err == nil && len(simQuestions) > 0 {
					sq := simQuestions[0]
					optJSON, _ := json.Marshal(sq.Options)
					newTQ := &models.TrainingQuestion{
						Topic:       m.TopicID,
						Type:        sq.Type,
						Difficulty:  sq.Difficulty,
						Prompt:      sq.Prompt,
						Options:     string(optJSON),
						Answer:      sq.Answer,
						Explanation: sq.Explanation,
						Source:      "ai_similar",
					}
					database.DB.Create(newTQ)
					if newTQ.ID > 0 {
						sessionQuestions = append(sessionQuestions, *newTQ)
					}
				}
			}
		}
	}

	// 3. Mode: "adaptive" or top-up logic
	// CRITICAL: If strictly in "recovery" mode for a specific attempt, we might NOT want to top-up
	// but the user's requirement says "Allow the user to retry only those failed questions."
	shouldTopUp := req.Mode == "adaptive" || (req.Mode == "mistakes" && req.AttemptID == "")
	
	if shouldTopUp && len(sessionQuestions) < targetCount {
		log.Printf("[TRAIN] Stage 3: Filling slots (current: %d/%d)", len(sessionQuestions), targetCount)
		
		if len(sessionQuestions) < targetCount && len(weakTopics) > 0 {
			topicToFocus := weakTopics[0]
			diff := "easy"
			if topicToFocus.AccuracyPercent > 40 { diff = "medium" }
			if topicToFocus.AccuracyPercent > 70 { diff = "hard" }
			
			aiQuestions, err := services.GenerateQuestions(topicToFocus.TopicName, diff, 5, nil)
			if err == nil {
				for _, aq := range aiQuestions {
					optJSON, _ := json.Marshal(aq.Options)
					tq := &models.TrainingQuestion{
						Topic:       topicToFocus.TopicID,
						Type:        aq.Type,
						Difficulty:  aq.Difficulty,
						Prompt:      aq.Prompt,
						Options:     string(optJSON),
						Answer:      aq.Answer,
						Explanation: aq.Explanation,
						Source:      "ai_adaptive",
					}
					database.DB.Create(tq)
					if tq.ID > 0 {
						sessionQuestions = append(sessionQuestions, *tq)
					}
				}
			}
		}

		if len(sessionQuestions) < targetCount {
			var pool []models.TrainingQuestion
			limit := targetCount - len(sessionQuestions)
			searchTopic := strings.ToLower(req.TopicID)
			
			if searchTopic != "" {
				database.DB.Order("RAND()").Where("topic = ?", searchTopic).Limit(limit).Find(&pool)
			}
			if len(pool) < limit {
				var fallbackPool []models.TrainingQuestion
				database.DB.Order("RAND()").Limit(limit - len(pool)).Find(&fallbackPool)
				pool = append(pool, fallbackPool...)
			}
			sessionQuestions = append(sessionQuestions, pool...)
		}
	}

	log.Printf("[TRAIN] Final session size: %d", len(sessionQuestions))
	
	// Final ID extraction and validation
	qIDs := []uint{}
	for _, q := range sessionQuestions {
		if q.ID > 0 {
			qIDs = append(qIDs, q.ID)
		}
	}

	if len(qIDs) == 0 {
		log.Printf("[ERROR] Failed to assemble session: zero valid IDs found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Neural Vault is empty. Please upload notes or complete tests first."})
		return
	}

	// Shuffle for variety
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(sessionQuestions), func(i, j int) {
		sessionQuestions[i], sessionQuestions[j] = sessionQuestions[j], sessionQuestions[i]
	})

	// 4. Store Session
	sessionID := uuid.New().String()
	qIDsJSON, _ := json.Marshal(qIDs)
	session := models.TrainingSession{
		SessionID:   sessionID,
		Topic:       req.TopicID,
		QuestionIDs: string(qIDsJSON),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	database.DB.Create(&session)

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"questions": sessionQuestions,
		"mode":      req.Mode,
		"weakTopics": weakTopics,
	})
}

// ──────────────────────────────────────────────
// SubmitAdaptiveAnswer → POST /api/training/adaptive/submit
// Grades a training answer and updates mastery stats.
// ──────────────────────────────────────────────
func SubmitAdaptiveAnswer(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	type Submission struct {
		QuestionID    uint   `json:"questionId"`
		UserAnswer    string `json:"userAnswer"`
		CorrectAnswer string `json:"correctAnswer"`
		IsCorrect     bool   `json:"isCorrect"`
		TopicID       string `json:"topicId"`
	}
	var sub Submission
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission"})
		return
	}

	// 1. If correct, and it was a "Mistake", we might mark it as reviewed or mastered
	if sub.IsCorrect {
		// Try to find if this question exists in the user's wrong list
		// Note: This check is heuristic since TrainingQuestions don't always map 1:1 to TestQuestions
		// but we can match by title/prompt if needed.
		// For now, we'll just track that the user is improving in the topic.
	}

	// 2. Update Topic Stats (incrementally)
	var stats models.UserTopicStats
	if err := database.DB.Where("userId = ? AND topicId = ?", userID, sub.TopicID).First(&stats).Error; err == nil {
		total := stats.TotalAttempted + 1
		correct := stats.TotalCorrect
		if sub.IsCorrect {
			correct++
		}
		wrong := stats.TotalWrong
		if !sub.IsCorrect {
			wrong++
		}

		accuracy := float64(correct) / float64(total) * 100
		
		level := "strong"
		switch {
		case accuracy < 40: level = "critical"
		case accuracy < 60: level = "weak"
		case accuracy < 80: level = "moderate"
		}

		database.DB.Model(&stats).Updates(map[string]interface{}{
			"totalAttempted":  total,
			"totalCorrect":    correct,
			"totalWrong":      wrong,
			"accuracyPercent": accuracy,
			"weakLevel":       level,
			"lastAttemptedAt": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"correct": sub.IsCorrect,
	})
}

// ──────────────────────────────────────────────
// GetMistakesAnalytics → GET /api/admin/analytics/mistakes
// Admin view of most commonly failed questions across the platform.
// ──────────────────────────────────────────────
func GetMistakesAnalytics(c *gin.Context) {
	type CommonMistake struct {
		QuestionTitle string  `json:"questionTitle"`
		TopicID       string  `json:"topicId"`
		FailureCount  int     `json:"failureCount"`
		FailureRate   float64 `json:"failureRate"`
	}

	var results []CommonMistake
	database.DB.Raw(`
		SELECT 
			questionTitle, 
			topicId, 
			COUNT(*) as failure_count,
			(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM test_submissions WHERE questionId = user_wrong_questions.questionId)) as failure_rate
		FROM user_wrong_questions
		GROUP BY questionTitle, topicId
		ORDER BY failure_count DESC
		LIMIT 20
	`).Scan(&results)

	c.JSON(http.StatusOK, results)
}

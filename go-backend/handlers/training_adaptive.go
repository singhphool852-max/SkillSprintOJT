package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdaptiveSessionRequest defines parameters for starting a practice session
type AdaptiveSessionRequest struct {
	TopicID    string `json:"topicId"`    // Optional: focus on one topic
	Difficulty string `json:"difficulty"` // Optional: override adaptive difficulty
	Mode       string `json:"mode"`       // "mistakes", "adaptive", "mastery"
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

	// 1. Identify Weak Topics (if not specified)
	var weakTopics []models.UserTopicStats
	if req.TopicID == "" {
		database.DB.Where("userId = ? AND accuracyPercent < 70", userID).
			Order("accuracyPercent ASC").Limit(3).Find(&weakTopics)
	} else {
		var stat models.UserTopicStats
		database.DB.Where("userId = ? AND topicId = ?", userID, req.TopicID).First(&stat)
		weakTopics = append(weakTopics, stat)
	}

	var sessionQuestions []models.TrainingQuestion
	targetCount := 10

	// 2. Mode: "mistakes" -> Focus strictly on UserWrongQuestions
	if req.Mode == "mistakes" {
		var mistakes []models.UserWrongQuestion
		query := database.DB.Where("userId = ? AND masteredAt IS NULL", userID)
		if req.TopicID != "" {
			query = query.Where("topicId = ?", req.TopicID)
		}
		// PRIORITIZE: Highest WrongCount first (questions failed many times)
		query.Order("wrongCount DESC, reviewCount ASC, createdAt DESC").Limit(targetCount).Find(&mistakes)

		for _, m := range mistakes {
			// Convert UserWrongQuestion back to TrainingQuestion format for the UI
			var tq models.TrainingQuestion
			tq.Prompt = m.QuestionTitle
			tq.Topic = m.TopicID
			tq.Difficulty = m.Difficulty
			tq.Type = m.QuestionType
			tq.Answer = m.CorrectAnswer
			
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
			sessionQuestions = append(sessionQuestions, tq)

			// NEW: Generate a similar question based on THIS specific mistake
			if req.Mode == "adaptive" && len(sessionQuestions) < targetCount {
				log.Printf("[ADAPTIVE] Generating similar variation for question: %s", m.QuestionTitle)
				simQuestions, err := services.GenerateSimilarQuestions(m.QuestionTitle, m.TopicID, m.Difficulty, 1)
				if err == nil && len(simQuestions) > 0 {
					sq := simQuestions[0]
					optJSON, _ := json.Marshal(sq.Options)
					sessionQuestions = append(sessionQuestions, models.TrainingQuestion{
						Topic:       m.TopicID,
						Type:        sq.Type,
						Difficulty:  sq.Difficulty,
						Prompt:      sq.Prompt,
						Options:     string(optJSON),
						Answer:      sq.Answer,
						Explanation: sq.Explanation,
						Source:      "ai_similar",
					})
				}
			}
		}
	}

	// 3. Mode: "adaptive" -> Mix Mistakes, AI Similar, and Random
	if req.Mode == "adaptive" || len(sessionQuestions) < targetCount {
		// Fill remaining with AI-generated similar questions based on weak areas
		if len(weakTopics) > 0 {
			topicToFocus := weakTopics[0]
			
			// Calculate difficulty based on accuracy
			diff := "easy"
			if topicToFocus.AccuracyPercent > 40 { diff = "medium" }
			if topicToFocus.AccuracyPercent > 70 { diff = "hard" }
			
			log.Printf("[ADAPTIVE] Generating AI questions for user %s on topic %s (diff: %s)", userID, topicToFocus.TopicName, diff)
			
			aiQuestions, err := services.GenerateQuestions(topicToFocus.TopicName, diff, 5, nil)
			if err == nil {
				for _, aq := range aiQuestions {
					optJSON, _ := json.Marshal(aq.Options)
					sessionQuestions = append(sessionQuestions, models.TrainingQuestion{
						Topic:       topicToFocus.TopicID,
						Type:        aq.Type,
						Difficulty:  aq.Difficulty,
						Prompt:      aq.Prompt,
						Options:     string(optJSON),
						Answer:      aq.Answer,
						Explanation: aq.Explanation,
						Source:      "ai_adaptive",
					})
				}
			}
		}

		// Fill remaining with random practice questions from the pool
		if len(sessionQuestions) < targetCount {
			var pool []models.TrainingQuestion
			limit := targetCount - len(sessionQuestions)
			
			q := database.DB.Order("RANDOM()")
			if req.TopicID != "" {
				q = q.Where("topic = ?", req.TopicID)
			}
			q.Limit(limit).Find(&pool)
			sessionQuestions = append(sessionQuestions, pool...)
		}
	}

	// Shuffle for variety
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(sessionQuestions), func(i, j int) {
		sessionQuestions[i], sessionQuestions[j] = sessionQuestions[j], sessionQuestions[i]
	})

	// 4. Store Session
	sessionID := uuid.New().String()
	var qIDs []uint
	for _, q := range sessionQuestions {
		qIDs = append(qIDs, q.ID)
	}
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

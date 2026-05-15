package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sort"
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
		query := database.DB.Where("userId = ? AND (masteredAt IS NULL OR masteredAt < '0001-01-02')", userID)
		
		if req.AttemptID != "" {
			query = query.Where("attemptId = ?", req.AttemptID)
			log.Printf("[RECOVERY] Fetching mistakes specifically for attempt: %s", req.AttemptID)
		} else if req.TopicID != "" {
			query = query.Where("topicId = ?", req.TopicID)
		}
		
		// For global mistakes, prioritize worst failures. For specific recovery, show in order.
		if req.AttemptID == "" {
			var allMistakes []models.UserWrongQuestion
			query.Find(&allMistakes)

			// Calculate priority score for each mistake
			type scoredMistake struct {
				Mistake models.UserWrongQuestion
				Score   int
			}
			var scored []scoredMistake
			now := time.Now()

			for _, m := range allMistakes {
				score := m.WrongCount * 5
				score -= m.CorrectStreak * 2

				// Recent failure weight (last 24 hours)
				if now.Sub(m.CreatedAt) < 24*time.Hour || (m.LastReviewedAt != nil && now.Sub(*m.LastReviewedAt) < 24*time.Hour) {
					score += 5
				}

				// Weak topic weight
				for _, wt := range weakTopics {
					if wt.TopicID == m.TopicID {
						if wt.AccuracyPercent < 40 {
							score += 10
						} else if wt.AccuracyPercent < 60 {
							score += 5
						}
						break
					}
				}

				scored = append(scored, scoredMistake{Mistake: m, Score: score})
			}

			// Sort by priority score DESC
			sort.Slice(scored, func(i, j int) bool {
				return scored[i].Score > scored[j].Score
			})

			for i := 0; i < len(scored) && i < targetCount; i++ {
				mistakes = append(mistakes, scored[i].Mistake)
				log.Printf("[ADAPTIVE] Selected mistake %s with priority score %d", scored[i].Mistake.QuestionID, scored[i].Score)
			}
		} else {
			query = query.Order("createdAt ASC").Limit(targetCount)
			query.Find(&mistakes)
		}
		
		log.Printf("[RECOVERY] Found %d mistake(s) for user", len(mistakes))

		// If no mistakes found in recovery/mistakes mode, return early with a clear response
		// instead of falling through to random vault questions
		if len(mistakes) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"sessionId":  "",
				"questions":  []interface{}{},
				"mode":       req.Mode,
				"weakTopics": weakTopics,
				"message":    "No pending wrong questions found. All mastered!",
			})
			return
		}

		for _, m := range mistakes {
			// Fetch the original question to get full details
			var origQuestion models.TestQuestion
			hasOriginal := database.DB.Preload("MCQOptions").Preload("CodingDetail").Preload("TestCases").
				Where("id = ?", m.QuestionID).First(&origQuestion).Error == nil

			prompt := m.QuestionTitle
			if prompt == "" && hasOriginal {
				prompt = origQuestion.Title
			}
			// For coding questions, include the description in the prompt
			if m.QuestionType == "coding" && hasOriginal && origQuestion.Description != "" {
				prompt = prompt + "\n\n" + origQuestion.Description
			}

			tq := &models.TrainingQuestion{
				Prompt:      prompt,
				Topic:       m.TopicID,
				Difficulty:  m.Difficulty,
				Type:        m.QuestionType,
				Answer:      m.CorrectAnswer,
				Source:      "recovery",
			}
			
			if m.QuestionType == "mcq" {
				if hasOriginal {
					var optTexts []string
					for _, o := range origQuestion.MCQOptions {
						optTexts = append(optTexts, o.OptionText)
					}
					optJSON, _ := json.Marshal(optTexts)
					tq.Options = string(optJSON)
					// Ensure correct answer is set
					if tq.Answer == "" {
						for _, o := range origQuestion.MCQOptions {
							if o.IsCorrect {
								tq.Answer = o.OptionText
								break
							}
						}
					}
				} else {
					var opts []models.TestMCQOption
					database.DB.Where("questionId = ?", m.QuestionID).Find(&opts)
					var optTexts []string
					for _, o := range opts {
						optTexts = append(optTexts, o.OptionText)
					}
					optJSON, _ := json.Marshal(optTexts)
					tq.Options = string(optJSON)
				}
			}

			if m.QuestionType == "coding" && hasOriginal {
				if origQuestion.CodingDetail != nil {
					tq.StarterCode = origQuestion.CodingDetail.StarterCode
					tq.Constraints = origQuestion.CodingDetail.Constraints
				}
				if len(origQuestion.TestCases) > 0 {
					tcJSON, _ := json.Marshal(origQuestion.TestCases)
					tq.TestCases = string(tcJSON)
				}
			}

			log.Printf("[ADAPTIVE] Building recovery question: type=%s prompt=%.60s answer=%.30s", m.QuestionType, prompt, tq.Answer)

			var existing models.TrainingQuestion
			if err := database.DB.Where("topic = ? AND prompt = ?", tq.Topic, tq.Prompt).First(&existing).Error; err == nil {
				tq.ID = existing.ID
				log.Printf("[RECOVERY] Mapping mistake to existing vault question ID: %d", tq.ID)
			} else {
				if err := database.DB.Create(tq).Error; err != nil {
					log.Printf("[ADAPTIVE ERROR] Failed to persist recovery question: %v", err)
				} else {
					log.Printf("[ADAPTIVE] Created recovery question ID: %d", tq.ID)
				}
			}

			if tq.ID > 0 {
				sessionQuestions = append(sessionQuestions, *tq)
			} else {
				log.Printf("[ADAPTIVE ERROR] Recovery question has no valid ID, skipping")
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
			
			// Topic prioritization: requested topic > weakest topic > second weakest
			searchTopic := strings.ToLower(req.TopicID)
			if searchTopic == "" && len(weakTopics) > 0 {
				searchTopic = strings.ToLower(weakTopics[0].TopicID)
				log.Printf("[TRAIN] No topic requested. Targeting weakest area: %s", searchTopic)
			}
			
			if searchTopic != "" {
				database.DB.Order("RAND()").Where("topic = ?", searchTopic).Limit(limit).Find(&pool)
			}
			
			// Global Fallback (Stage 5) - only pull seeded or AI-generated questions, NOT recovery junk
			if len(pool) < limit {
				var fallbackPool []models.TrainingQuestion
				database.DB.Where("source IN ?", []string{"seeded", "ai_adaptive", "ai_similar", "notes"}).
					Order("RAND()").Limit(limit - len(pool)).Find(&fallbackPool)
				pool = append(pool, fallbackPool...)
			}
			sessionQuestions = append(sessionQuestions, pool...)
		}
	}

	// 4. Emergency Fallback (Stage 6) - If still zero questions, pick from curated vault only
	if len(sessionQuestions) == 0 {
		log.Printf("[TRAIN] CRITICAL: Session still empty. Attempting emergency fallback (curated sources only).")
		database.DB.Where("source IN ?", []string{"seeded", "ai_adaptive", "ai_similar", "notes"}).
			Order("RAND()").Limit(targetCount).Find(&sessionQuestions)
	}

	log.Printf("[TRAIN] Final session size: %d", len(sessionQuestions))
	
	// 5. Final ID extraction and validation
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

	// 6. Store Session
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
	// Fetch the TrainingQuestion to map it back to the original mistake
	var tq models.TrainingQuestion
	if err := database.DB.Where("id = ?", sub.QuestionID).First(&tq).Error; err == nil {
		if tq.Source == "recovery" {
			var wq models.UserWrongQuestion
			// Match by userID and prompt (since it's a snapshot)
			if err := database.DB.Where("userId = ? AND questionTitle = ?", userID, tq.Prompt).First(&wq).Error; err == nil {
				if sub.IsCorrect {
					wq.CorrectStreak++
					log.Printf("[ADAPTIVE] Correct streak incremented to %d for question %s", wq.CorrectStreak, wq.QuestionID)
					
					// Mastery threshold check
					if wq.CorrectStreak >= 2 {
						now := time.Now()
						wq.MasteredAt = &now
						log.Printf("[ADAPTIVE] Question %s MASTERED!", wq.QuestionID)
					}
				} else {
					wq.CorrectStreak = 0 // Reset streak
					wq.WrongCount++
					log.Printf("[ADAPTIVE] Streak reset and wrong count incremented for question %s", wq.QuestionID)
				}
				
				database.DB.Save(&wq)
			}
		}
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
// Admin view of anti-cheat violations across the platform.
// ──────────────────────────────────────────────
func GetMistakesAnalytics(c *gin.Context) {
	type ViolationRow struct {
		UserName        string `json:"userName"`
		UserEmail       string `json:"userEmail"`
		TestTitle       string `json:"testTitle"`
		ViolationCount  int    `json:"violationCount"`
		FullscreenExits int    `json:"fullscreenExits"`
		TabSwitches     int    `json:"tabSwitches"`
		LastViolation   string `json:"lastViolation"`
	}

	var results []ViolationRow
	database.DB.Raw(`
		SELECT 
			u.name as user_name,
			u.email as user_email,
			t.title as test_title,
			COUNT(v.id) as violation_count,
			SUM(CASE WHEN v.violationType = 'fullscreen_exit' THEN 1 ELSE 0 END) as fullscreen_exits,
			SUM(CASE WHEN v.violationType = 'tab_switch' THEN 1 ELSE 0 END) as tab_switches,
			MAX(v.timestamp) as last_violation
		FROM test_violations v
		JOIN users u ON u.id = v.userId
		JOIN tests t ON t.id = v.testId
		GROUP BY v.userId, v.testId, u.name, u.email, t.title
		ORDER BY violation_count DESC
		LIMIT 50
	`).Scan(&results)

	c.JSON(http.StatusOK, gin.H{
		"mistakes": results,
		"total":    len(results),
	})
}

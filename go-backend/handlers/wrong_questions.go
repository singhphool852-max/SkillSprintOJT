package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExtractWrongQuestionsManual is an exported wrapper for testing
func ExtractWrongQuestionsManual(attempt models.TestAttempt) {
	extractWrongQuestions(attempt)
}

// UpdateUserTopicStatsManual is an exported wrapper for testing
func UpdateUserTopicStatsManual(userID string, testID string) {
	updateUserTopicStats(userID, testID)
}

// extractWrongQuestions — called after test submission.
// Creates UserWrongQuestion rows for every wrong/skipped answer.
// ──────────────────────────────────────────────
func extractWrongQuestions(attempt models.TestAttempt) {
	// Load the test to get topicId
	var test models.Test
	if err := database.DB.Where("id = ?", attempt.TestID).First(&test).Error; err != nil {
		log.Printf("[WRONG-Q] failed to load test %s: %v", attempt.TestID, err)
		return
	}

	// Load all questions for this test
	var questions []models.TestQuestion
	database.DB.Where("testId = ?", attempt.TestID).Find(&questions)

	// Load all submissions for this attempt
	var submissions []models.TestSubmission
	database.DB.Where("attemptId = ?", attempt.ID).Find(&submissions)

	// Build submission lookup by questionID
	subMap := make(map[string]*models.TestSubmission)
	for i := range submissions {
		subMap[submissions[i].QuestionID] = &submissions[i]
	}

	// Check what wrong questions already exist for this attempt (idempotency)
	var existingCount int64
	database.DB.Model(&models.UserWrongQuestion{}).Where("attemptId = ?", attempt.ID).Count(&existingCount)
	if existingCount > 0 {
		log.Printf("[WRONG-Q] attempt %s already processed (%d entries), skipping", attempt.ID, existingCount)
		return
	}

	for _, q := range questions {
		sub, submitted := subMap[q.ID]
		verdict := "skipped"
		userAnswer := ""
		correctAnswer := ""

		if submitted {
			verdict = sub.Verdict
			userAnswer = sub.Code
			if q.Type == "mcq" {
				// Get selected option text
				var selectedOpt models.TestMCQOption
				if err := database.DB.Where("id = ?", sub.SelectedOptionID).First(&selectedOpt).Error; err == nil {
					userAnswer = selectedOpt.OptionText
				}
				// Get correct option text
				var correctOpt models.TestMCQOption
				if err := database.DB.Where("questionId = ? AND isCorrect = ?", q.ID, true).First(&correctOpt).Error; err == nil {
					correctAnswer = correctOpt.OptionText
				}
			}
		}

		// LOGIC: If wrong or skipped, UPSERT into user_wrong_questions
		if verdict != "accepted" && verdict != "draft" {
			var existing models.UserWrongQuestion
			err := database.DB.Where("userId = ? AND questionId = ?", attempt.UserID, q.ID).First(&existing).Error
			
			if err == nil {
				// Update existing: increment wrong count and refresh snapshot
				database.DB.Model(&existing).Updates(map[string]interface{}{
					"attemptId":      attempt.ID,
					"testId":         attempt.TestID,
					"userAnswer":     userAnswer,
					"correctAnswer":  correctAnswer,
					"verdict":        verdict,
					"wrongCount":     existing.WrongCount + 1,
					"correctStreak":  0, // Reset streak on fresh failure
					"masteredAt":     nil, // Un-master if they fail it again
					"pointsLost":     q.Points - subScore(sub),
					"pointsPossible": q.Points,
				})
			} else {
				// Create new
				database.DB.Create(&models.UserWrongQuestion{
					ID:             uuid.New().String(),
					UserID:         attempt.UserID,
					AttemptID:      attempt.ID,
					QuestionID:     q.ID,
					TestID:         attempt.TestID,
					TopicID:        test.TopicID,
					QuestionType:   q.Type,
					QuestionTitle:  q.Title,
					UserAnswer:     userAnswer,
					CorrectAnswer:  correctAnswer,
					Verdict:        verdict,
					WrongCount:     1,
					PointsLost:     q.Points - subScore(sub),
					PointsPossible: q.Points,
				})
			}
		}
	}

	// Update topic stats
	updateUserTopicStats(attempt.UserID, attempt.TestID)
}

func subScore(s *models.TestSubmission) int {
	if s == nil { return 0 }
	return s.Score
}

// ──────────────────────────────────────────────
// updateUserTopicStats — recalculates accuracy per topic for a user.
// Called after every test submission.
// ──────────────────────────────────────────────
func updateUserTopicStats(userID string, testID string) {
	// Get the test's topic
	var test models.Test
	if err := database.DB.Preload("Topic").Where("id = ?", testID).First(&test).Error; err != nil {
		log.Printf("[TOPIC-STATS] failed to load test %s: %v", testID, err)
		return
	}

	if test.TopicID == "" {
		log.Printf("[TOPIC-STATS] test %s has no topic, skipping", testID)
		return
	}

	topicName := ""
	if test.Topic != nil {
		topicName = test.Topic.Name
	}

	// Count total correct, wrong, skipped for this user + topic across all attempts
	type stats struct {
		Total   int
		Correct int
		Wrong   int
		Skipped int
	}
	var s stats

	// Total questions attempted for this topic
	database.DB.Raw(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN ts.verdict = 'accepted' THEN 1 ELSE 0 END) as correct,
			SUM(CASE WHEN ts.verdict IN ('wrong_answer','time_limit','compile_error','runtime_error') THEN 1 ELSE 0 END) as wrong,
			SUM(CASE WHEN ts.verdict = '' OR ts.verdict IS NULL OR ts.verdict = 'draft' THEN 1 ELSE 0 END) as skipped
		FROM test_submissions ts
		JOIN test_attempts ta ON ta.id = ts.attemptId
		JOIN tests t ON t.id = ta.testId
		WHERE ta.userId = ? AND t.topicId = ?
		  AND ta.submittedAt IS NOT NULL
		  AND ta.submittedAt > '0001-01-02'
	`, userID, test.TopicID).Scan(&s)

	// Also count skipped questions (no submission row at all)
	var skippedFromWrong int64
	database.DB.Model(&models.UserWrongQuestion{}).
		Where("userId = ? AND topicId = ? AND verdict = 'skipped'", userID, test.TopicID).
		Count(&skippedFromWrong)

	s.Total += int(skippedFromWrong)
	s.Skipped += int(skippedFromWrong)

	accuracy := float64(0)
	if s.Total > 0 {
		accuracy = float64(s.Correct) / float64(s.Total) * 100
	}

	weakLevel := "strong"
	switch {
	case accuracy < 40:
		weakLevel = "critical"
	case accuracy < 60:
		weakLevel = "weak"
	case accuracy < 80:
		weakLevel = "moderate"
	}

	// Upsert: check if stats row exists
	var existing models.UserTopicStats
	if err := database.DB.Where("userId = ? AND topicId = ?", userID, test.TopicID).First(&existing).Error; err != nil {
		// Create new
		database.DB.Create(&models.UserTopicStats{
			ID:              uuid.New().String(),
			UserID:          userID,
			TopicID:         test.TopicID,
			TopicName:       topicName,
			TotalAttempted:  s.Total,
			TotalCorrect:    s.Correct,
			TotalWrong:      s.Wrong,
			TotalSkipped:    s.Skipped,
			AccuracyPercent: accuracy,
			WeakLevel:       weakLevel,
			LastAttemptedAt: time.Now(),
		})
	} else {
		// Update existing
		database.DB.Model(&existing).Updates(map[string]interface{}{
			"topicName":       topicName,
			"totalAttempted":  s.Total,
			"totalCorrect":    s.Correct,
			"totalWrong":      s.Wrong,
			"totalSkipped":    s.Skipped,
			"accuracyPercent": accuracy,
			"weakLevel":       weakLevel,
			"lastAttemptedAt": time.Now(),
		})
	}

	log.Printf("[TOPIC-STATS] user=%s topic=%s accuracy=%.1f%% level=%s", userID, test.TopicID, accuracy, weakLevel)
}

// ══════════════════════════════════════════════
// API Handlers
// ══════════════════════════════════════════════

// ──────────────────────────────────────────────
// GetUserWrongQuestions → GET /api/training/wrong-questions
// Returns the user's wrong questions, filterable by topic and difficulty.
// ──────────────────────────────────────────────
func GetUserWrongQuestions(c *gin.Context) {
	userID, _ := c.Get("userID")

	query := database.DB.Where("userId = ? AND masteredAt IS NULL OR masteredAt < '0001-01-02'", userID).
		Preload("Question").Preload("Test").
		Order("createdAt DESC")

	// Optional filters
	if topicID := c.Query("topicId"); topicID != "" {
		query = query.Where("topicId = ?", topicID)
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if verdict := c.Query("verdict"); verdict != "" {
		query = query.Where("verdict = ?", verdict)
	}

	var questions []models.UserWrongQuestion
	if err := query.Limit(100).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wrong questions"})
		return
	}

	c.JSON(http.StatusOK, questions)
}

// ──────────────────────────────────────────────
// GetWrongQuestionSummary → GET /api/training/wrong-questions/summary
// Aggregate counts by topic and difficulty.
// ──────────────────────────────────────────────
func GetWrongQuestionSummary(c *gin.Context) {
	userID, _ := c.Get("userID")

	type TopicSummary struct {
		TopicID  string `json:"topicId"`
		Verdict  string `json:"verdict"`
		Count    int    `json:"count"`
	}

	var summaries []TopicSummary
	database.DB.Model(&models.UserWrongQuestion{}).
		Select("topicId as topic_id, verdict, COUNT(*) as count").
		Where("userId = ? AND (masteredAt IS NULL OR masteredAt < '0001-01-02')", userID).
		Group("topicId, verdict").
		Scan(&summaries)

	c.JSON(http.StatusOK, summaries)
}

// ──────────────────────────────────────────────
// GetUserWeakTopics → GET /api/training/weak-topics
// Returns the user's per-topic accuracy stats, sorted by weakness.
// ──────────────────────────────────────────────
func GetUserWeakTopics(c *gin.Context) {
	userID, _ := c.Get("userID")

	var stats []models.UserTopicStats
	if err := database.DB.Where("userId = ?", userID).
		Order("accuracyPercent ASC").
		Find(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weak topics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ──────────────────────────────────────────────
// MarkQuestionReviewed → POST /api/training/wrong-questions/:id/review
// Increments reviewCount and updates lastReviewedAt.
// ──────────────────────────────────────────────
func MarkQuestionReviewed(c *gin.Context) {
	wrongQID := c.Param("id")
	userID, _ := c.Get("userID")

	var wq models.UserWrongQuestion
	if err := database.DB.Where("id = ? AND userId = ?", wrongQID, userID).First(&wq).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong question not found"})
		return
	}

	database.DB.Model(&wq).Updates(map[string]interface{}{
		"reviewCount":    wq.ReviewCount + 1,
		"lastReviewedAt": time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Review recorded", "reviewCount": wq.ReviewCount + 1})
}

// ──────────────────────────────────────────────
// MarkQuestionMastered → POST /api/training/wrong-questions/:id/master
// Sets masteredAt timestamp — stops showing in wrong-question lists.
// ──────────────────────────────────────────────
func MarkQuestionMastered(c *gin.Context) {
	wrongQID := c.Param("id")
	userID, _ := c.Get("userID")

	var wq models.UserWrongQuestion
	if err := database.DB.Where("id = ? AND userId = ?", wrongQID, userID).First(&wq).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wrong question not found"})
		return
	}

	database.DB.Model(&wq).Updates(map[string]interface{}{
		"masteredAt": time.Now(),
	})

	// Recalculate topic stats since mastering a question may change accuracy
	updateUserTopicStats(wq.UserID, wq.TestID)

	c.JSON(http.StatusOK, gin.H{"message": "Question mastered"})
}

// ──────────────────────────────────────────────
// GetTestAttemptsList → GET /api/admin/tests/:id/attempts
// Returns all attempts for a test (admin view).
// ──────────────────────────────────────────────
func GetTestAttemptsList(c *gin.Context) {
	testID := c.Param("id")

	type AttemptRow struct {
		AttemptID       string `json:"attemptId"`
		UserID          string `json:"userId"`
		Username        string `json:"username"`
		Score           int    `json:"score"`
		TotalQuestions  int    `json:"totalQuestions"`
		TimeTaken       int    `json:"timeTaken"`
		ViolationCount  int    `json:"violationCount"`
		IsAutoSubmitted bool   `json:"isAutoSubmitted"`
		SubmittedAt     string `json:"submittedAt"`
	}

	var attempts []AttemptRow
	database.DB.Table("test_attempts").
		Select("test_attempts.id as attempt_id, test_attempts.userId as user_id, user.username, test_attempts.score, test_attempts.totalQuestions as total_questions, test_attempts.timeTaken as time_taken, test_attempts.violationCount as violation_count, test_attempts.isAutoSubmitted as is_auto_submitted, test_attempts.submittedAt as submitted_at").
		Joins("JOIN user ON user.id = test_attempts.userId").
		Where("test_attempts.testId = ? AND test_attempts.submittedAt IS NOT NULL AND test_attempts.submittedAt > '0001-01-02'", testID).
		Order("test_attempts.score DESC").
		Scan(&attempts)

	c.JSON(http.StatusOK, attempts)
}

// ──────────────────────────────────────────────
// GetTestAnalytics → GET /api/admin/tests/:id/analytics
// Per-question success rates for a test (admin view).
// ──────────────────────────────────────────────
func GetTestAnalytics(c *gin.Context) {
	testID := c.Param("id")

	type QuestionStat struct {
		QuestionID    string  `json:"questionId"`
		Title         string  `json:"title"`
		Type          string  `json:"type"`
		Points        int     `json:"points"`
		TotalAttempts int     `json:"totalAttempts"`
		CorrectCount  int     `json:"correctCount"`
		WrongCount    int     `json:"wrongCount"`
		SkipCount     int     `json:"skipCount"`
		SuccessRate   float64 `json:"successRate"`
	}

	var questions []models.TestQuestion
	database.DB.Where("testId = ?", testID).Order("position ASC").Find(&questions)

	stats := make([]QuestionStat, len(questions))
	for i, q := range questions {
		stat := QuestionStat{
			QuestionID: q.ID,
			Title:      q.Title,
			Type:       q.Type,
			Points:     q.Points,
		}

		// Count submissions per verdict
		var total int64
		database.DB.Model(&models.TestSubmission{}).
			Joins("JOIN test_attempts ON test_attempts.id = test_submissions.attemptId").
			Where("test_submissions.questionId = ? AND test_attempts.submittedAt > '0001-01-02'", q.ID).
			Count(&total)
		stat.TotalAttempts = int(total)

		var correct int64
		database.DB.Model(&models.TestSubmission{}).
			Joins("JOIN test_attempts ON test_attempts.id = test_submissions.attemptId").
			Where("test_submissions.questionId = ? AND test_submissions.verdict = 'accepted' AND test_attempts.submittedAt > '0001-01-02'", q.ID).
			Count(&correct)
		stat.CorrectCount = int(correct)
		stat.WrongCount = int(total) - int(correct)

		if total > 0 {
			stat.SuccessRate = float64(correct) / float64(total) * 100
		}

		stats[i] = stat
	}

	c.JSON(http.StatusOK, stats)
}

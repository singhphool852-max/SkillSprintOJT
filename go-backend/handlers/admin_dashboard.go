package handlers

import (
	"log"
	"time"

	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ══════════════════════════════════════════════
// Admin Dashboard Analytics
// ══════════════════════════════════════════════

// ──────────────────────────────────────────────
// GetAdminDashboardStats → GET /api/admin/dashboard/stats
// Returns aggregate stats: total tests, users, attempts, avg scores.
// ──────────────────────────────────────────────
func GetAdminDashboardStats(c *gin.Context) {
	var totalTests int64
	database.DB.Model(&models.Test{}).Count(&totalTests)

	var publishedTests int64
	database.DB.Model(&models.Test{}).Where("isPublished = ?", true).Count(&publishedTests)

	var totalUsers int64
	database.DB.Model(&models.User{}).Count(&totalUsers)

	var adminUsers int64
	database.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminUsers)

	var totalTopics int64
	database.DB.Model(&models.Topic{}).Count(&totalTopics)

	var totalAttempts int64
	database.DB.Model(&models.TestAttempt{}).Count(&totalAttempts)

	var submittedAttempts int64
	database.DB.Model(&models.TestAttempt{}).Where("submittedAt IS NOT NULL").Count(&submittedAttempts)

	var avgScore struct {
		Avg float64 `json:"avg"`
	}
	database.DB.Model(&models.TestAttempt{}).
		Where("submittedAt IS NOT NULL").
		Select("AVG(score) as avg").
		Scan(&avgScore)

	var totalQuestions int64
	database.DB.Model(&models.TestQuestion{}).Count(&totalQuestions)

	var activeTests int64
	database.DB.Model(&models.Test{}).Where("isActive = ?", true).Count(&activeTests)

	// Count users currently active in an arena test
	// Active = started within last 4 hours and not yet submitted
	var activeArenaUsers int64
	result := database.DB.Model(&models.TestAttempt{}).
		Where("submittedAt IS NULL AND startedAt > ?", 
			time.Now().Add(-4*time.Hour)).
		Distinct("userId").
		Count(&activeArenaUsers)
	
	if result.Error != nil {
		log.Printf("[ANALYTICS] activeArenaUsers query error: %v", result.Error)
		activeArenaUsers = 0
	}
	log.Printf("[ANALYTICS] activeArenaUsers count: %d", activeArenaUsers)

	c.JSON(http.StatusOK, gin.H{
		"totalTests":        totalTests,
		"publishedTests":    publishedTests,
		"activeTests":       activeTests,
		"totalUsers":        totalUsers,
		"activeArenaUsers":  activeArenaUsers,
		"adminUsers":        adminUsers,
		"regularUsers":      totalUsers - adminUsers,
		"totalTopics":       totalTopics,
		"totalAttempts":     totalAttempts,
		"submittedAttempts": submittedAttempts,
		"avgScore":          avgScore.Avg,
		"totalQuestions":    totalQuestions,
	})
}

// ──────────────────────────────────────────────
// GetRecentActivity → GET /api/admin/dashboard/recent
// Returns recent test attempts for admin overview.
// ──────────────────────────────────────────────
func GetRecentActivity(c *gin.Context) {
	type RecentAttempt struct {
		AttemptID       string  `json:"attemptId"`
		Username        string  `json:"username"`
		Email           string  `json:"email"`
		TestTitle       string  `json:"testTitle"`
		Score           int     `json:"score"`
		IsAutoSubmitted bool    `json:"isAutoSubmitted"`
		SubmittedAt     string  `json:"submittedAt"`
	}

	var results []RecentAttempt
	database.DB.Table("test_attempts").
		Select("test_attempts.id as attempt_id, user.username, user.email, tests.title as test_title, test_attempts.score, test_attempts.isAutoSubmitted as is_auto_submitted, test_attempts.submittedAt as submitted_at").
		Joins("JOIN user ON user.id = test_attempts.userId").
		Joins("JOIN tests ON tests.id = test_attempts.testId").
		Where("test_attempts.submittedAt IS NOT NULL").
		Order("test_attempts.submittedAt DESC").
		Limit(20).
		Scan(&results)

	c.JSON(http.StatusOK, results)
}

// ══════════════════════════════════════════════
// User Dashboard Analytics
// ══════════════════════════════════════════════

// ──────────────────────────────────────────────
// GetUserDashboardStats → GET /api/dashboard/stats
// Returns the authenticated user's personal stats across all tests.
// ──────────────────────────────────────────────
func GetUserDashboardStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Test attempt stats
	var testStats struct {
		TotalAttempts    int     `json:"totalAttempts"`
		SubmittedCount   int     `json:"submittedCount"`
		HighScore        int     `json:"highScore"`
		AvgScore         float64 `json:"avgScore"`
		TotalScore       int     `json:"totalScore"`
	}
	database.DB.Table("test_attempts").
		Select("COUNT(*) as total_attempts, "+
			"SUM(CASE WHEN submittedAt IS NOT NULL THEN 1 ELSE 0 END) as submitted_count, "+
			"MAX(score) as high_score, "+
			"AVG(CASE WHEN submittedAt IS NOT NULL THEN score END) as avg_score, "+
			"SUM(CASE WHEN submittedAt IS NOT NULL THEN score ELSE 0 END) as total_score").
		Where("userId = ?", userID).
		Scan(&testStats)

	// Legacy arena stats
	var arenaStats struct {
		TotalAttempts int     `json:"totalAttempts"`
		HighScore     int     `json:"highScore"`
		AvgScore      float64 `json:"avgScore"`
	}
	database.DB.Table("attempts").
		Select("COUNT(*) as total_attempts, MAX(score) as high_score, AVG(score) as avg_score").
		Where("userId = ?", userID).
		Scan(&arenaStats)

	// Recent test attempts
	type RecentTestAttempt struct {
		AttemptID       string `json:"attemptId"`
		TestTitle       string `json:"testTitle"`
		Score           int    `json:"score"`
		IsAutoSubmitted bool   `json:"isAutoSubmitted"`
		SubmittedAt     string `json:"submittedAt"`
	}
	var recentTests []RecentTestAttempt
	database.DB.Table("test_attempts").
		Select("test_attempts.id as attempt_id, tests.title as test_title, test_attempts.score, test_attempts.isAutoSubmitted as is_auto_submitted, test_attempts.submittedAt as submitted_at").
		Joins("JOIN tests ON tests.id = test_attempts.testId").
		Where("test_attempts.userId = ? AND test_attempts.submittedAt IS NOT NULL", userID).
		Order("test_attempts.submittedAt DESC").
		Limit(10).
		Scan(&recentTests)

	// Count of available/active tests
	var activeTestCount int64
	database.DB.Model(&models.Test{}).Where("isActive = ? AND isPublished = ?", true, true).Count(&activeTestCount)

	c.JSON(http.StatusOK, gin.H{
		"testStats":       testStats,
		"arenaStats":      arenaStats,
		"recentTests":     recentTests,
		"activeTestCount": activeTestCount,
	})
}

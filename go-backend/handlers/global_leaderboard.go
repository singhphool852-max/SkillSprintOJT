package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ══════════════════════════════════════════════
// Global Leaderboard — All Users Ranked
// GET /api/leaderboard/global
// ══════════════════════════════════════════════

// GlobalLeaderboardEntry represents one user's ranking.
type GlobalLeaderboardEntry struct {
	Rank            int     `json:"rank"`
	UserID          string  `json:"userId"`
	Username        string  `json:"username"`
	TotalScore      int     `json:"totalScore"`
	TestsCompleted  int     `json:"testsCompleted"`
	AvgPercentage   float64 `json:"avgPercentage"`
	HighScore       int     `json:"highScore"`
	Tier            string  `json:"tier"`
}

// GetGlobalLeaderboard → GET /api/leaderboard/global
// Aggregates all submitted test attempts across all tests to produce a global ranking.
// Ranking logic: total score DESC, then earliest completedAt as tiebreak (first to submit wins).
func GetGlobalLeaderboard(c *gin.Context) {
	type LeaderboardRow struct {
		UserID         string `gorm:"column:user_id"`
		Username       string `gorm:"column:username"`
		TestsCount     int    `gorm:"column:tests_count"`
		BestScore      int    `gorm:"column:best_score"`
		TotalScore     int    `gorm:"column:total_score"`
		EarliestSubmit string `gorm:"column:earliest_submit"`
	}

	var entries []LeaderboardRow

	// Use GORM query builder with correct MySQL column names (camelCase)
	query := database.DB.Table("attempts").
		Select("attempts.userId as user_id, "+
			"user.username as username, "+
			"COUNT(DISTINCT attempts.id) as tests_count, "+
			"MAX(attempts.score) as best_score, "+
			"SUM(attempts.score) as total_score, "+
			"MIN(attempts.completedAt) as earliest_submit").
		Joins("JOIN user ON user.id = attempts.userId").
		Where("attempts.completedAt IS NOT NULL").
		Where("user.role != ?", "admin").
		Group("attempts.userId, user.username").
		Order("total_score DESC, earliest_submit ASC").
		Limit(100)

	log.Printf("[Leaderboard] Executing query...")

	if err := query.Scan(&entries).Error; err != nil {
		log.Printf("[Leaderboard] DB ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database query failed",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[Leaderboard] Found %d entries", len(entries))

	// Transform to response format
	results := make([]GlobalLeaderboardEntry, len(entries))
	totalUsers := len(entries)

	for i, entry := range entries {
		rank := i + 1

		// Calculate average percentage
		avgPct := float64(0)
		if entry.TestsCount > 0 && entry.TotalScore > 0 {
			avgPct = float64(entry.TotalScore) / float64(entry.TestsCount)
		}

		// Assign tier based on percentile
		tier := "ROOKIE"
		if totalUsers > 0 {
			percentile := float64(rank) / float64(totalUsers) * 100
			switch {
			case percentile <= 5:
				tier = "APEX"
			case percentile <= 15:
				tier = "CHAMPION"
			case percentile <= 30:
				tier = "VETERAN"
			case percentile <= 50:
				tier = "ELITE"
			case percentile <= 75:
				tier = "WARRIOR"
			default:
				tier = "ROOKIE"
			}
		}

		results[i] = GlobalLeaderboardEntry{
			Rank:           rank,
			UserID:         entry.UserID,
			Username:       entry.Username,
			TotalScore:     entry.TotalScore,
			TestsCompleted: entry.TestsCount,
			AvgPercentage:  float64(int(avgPct*100)) / 100,
			HighScore:      entry.BestScore,
			Tier:           tier,
		}

		log.Printf("[Leaderboard] #%d: %s (ID=%s) Score=%d Tests=%d Best=%d",
			rank, entry.Username, entry.UserID, entry.TotalScore, entry.TestsCount, entry.BestScore)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":    results,
		"totalUsers": totalUsers,
	})
}

// GetLeaderboardDebug → GET /api/leaderboard/debug
// Debug endpoint to check if attempts exist in the database
func GetLeaderboardDebug(c *gin.Context) {
	// Count total attempts
	var totalCount int64
	database.DB.Table("attempts").Count(&totalCount)

	// Count completed attempts
	var completedCount int64
	database.DB.Table("attempts").Where("completedAt IS NOT NULL").Count(&completedCount)

	// Count users
	var userCount int64
	database.DB.Table("user").Count(&userCount)

	// Count non-admin users
	var nonAdminCount int64
	database.DB.Table("user").Where("role != ?", "admin").Count(&nonAdminCount)

	// Get sample attempts
	type SampleAttempt struct {
		ID          string `gorm:"column:id"`
		UserID      string `gorm:"column:userId"`
		Score       int    `gorm:"column:score"`
		CompletedAt string `gorm:"column:completedAt"`
	}
	var samples []SampleAttempt
	database.DB.Table("attempts").
		Select("id, userId, score, completedAt").
		Where("completedAt IS NOT NULL").
		Order("completedAt DESC").
		Limit(5).
		Scan(&samples)

	log.Printf("[LeaderboardDebug] Total=%d Completed=%d Users=%d NonAdmin=%d Samples=%d",
		totalCount, completedCount, userCount, nonAdminCount, len(samples))

	c.JSON(http.StatusOK, gin.H{
		"tableName":         "attempts",
		"userTableName":     "user",
		"totalAttempts":     totalCount,
		"completedAttempts": completedCount,
		"totalUsers":        userCount,
		"nonAdminUsers":     nonAdminCount,
		"sampleAttempts":    samples,
		"message":           "Debug info retrieved successfully",
	})
}

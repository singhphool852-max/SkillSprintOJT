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
// Ranking logic: total score DESC, then earliest submittedAt as tiebreak (first to submit wins).
func GetGlobalLeaderboard(c *gin.Context) {
	type LeaderboardRow struct {
		UserID         string  `gorm:"column:user_id"`
		Username       string  `gorm:"column:username"`
		TestsCount     int     `gorm:"column:tests_count"`
		BestScore      float64 `gorm:"column:best_score"`
		TotalScore     float64 `gorm:"column:total_score"`
		EarliestSubmit string  `gorm:"column:earliest_submit"`
	}

	var entries []LeaderboardRow

	// FIXED: Use test_attempts table (not attempts) and submittedAt (not completedAt)
	query := database.DB.Table("test_attempts").
		Select("test_attempts.userId as user_id, "+
			"user.username as username, "+
			"COUNT(DISTINCT test_attempts.id) as tests_count, "+
			"MAX(test_attempts.score) as best_score, "+
			"SUM(test_attempts.score) as total_score, "+
			"MIN(test_attempts.submittedAt) as earliest_submit").
		Joins("JOIN user ON user.id = test_attempts.userId").
		Where("test_attempts.submittedAt IS NOT NULL").
		Where("user.role != ?", "admin").
		Group("test_attempts.userId, user.username").
		Order("total_score DESC, earliest_submit ASC").
		Limit(100)

	log.Printf("[Leaderboard] Executing query on test_attempts table...")

	if err := query.Scan(&entries).Error; err != nil {
		log.Printf("[Leaderboard] DB ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database query failed",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[Leaderboard] Found %d entries", len(entries))

	// Diagnostic: If no entries found, check why
	if len(entries) == 0 {
		var debugCount int64
		database.DB.Table("test_attempts").Where("submittedAt IS NOT NULL").Count(&debugCount)
		log.Printf("[Leaderboard] DIAGNOSTIC: test_attempts with submittedAt NOT NULL = %d", debugCount)
		
		var totalCount int64
		database.DB.Table("test_attempts").Count(&totalCount)
		log.Printf("[Leaderboard] DIAGNOSTIC: total test_attempts = %d", totalCount)
	}

	// Transform to response format
	results := make([]GlobalLeaderboardEntry, len(entries))
	totalUsers := len(entries)

	for i, entry := range entries {
		rank := i + 1

		// Calculate average percentage
		avgPct := float64(0)
		if entry.TestsCount > 0 && entry.TotalScore > 0 {
			avgPct = entry.TotalScore / float64(entry.TestsCount)
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
			TotalScore:     int(entry.TotalScore),
			TestsCompleted: entry.TestsCount,
			AvgPercentage:  float64(int(avgPct*100)) / 100,
			HighScore:      int(entry.BestScore),
			Tier:           tier,
		}

		log.Printf("[Leaderboard] #%d: %s (ID=%s) Score=%.0f Tests=%d Best=%.0f",
			rank, entry.Username, entry.UserID, entry.TotalScore, entry.TestsCount, entry.BestScore)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":    results,
		"totalUsers": totalUsers,
	})
}

// GetLeaderboardDebug → GET /api/leaderboard/debug
// Comprehensive debug endpoint to diagnose why leaderboard is empty
func GetLeaderboardDebug(c *gin.Context) {
	type DebugResult struct {
		Query  string      `json:"query"`
		Result interface{} `json:"result"`
	}

	results := []DebugResult{}

	// 1. Count total test_attempts
	var totalCount int64
	database.DB.Table("test_attempts").Count(&totalCount)
	results = append(results, DebugResult{
		Query:  "SELECT COUNT(*) FROM test_attempts",
		Result: totalCount,
	})

	// 2. Count test_attempts with submittedAt NOT NULL
	var submittedCount int64
	database.DB.Table("test_attempts").Where("submittedAt IS NOT NULL").Count(&submittedCount)
	results = append(results, DebugResult{
		Query:  "SELECT COUNT(*) FROM test_attempts WHERE submittedAt IS NOT NULL",
		Result: submittedCount,
	})

	// 3. Count users
	var userCount int64
	database.DB.Table("user").Count(&userCount)
	results = append(results, DebugResult{
		Query:  "SELECT COUNT(*) FROM user",
		Result: userCount,
	})

	// 4. Count non-admin users
	var nonAdminCount int64
	database.DB.Table("user").Where("role != ?", "admin").Count(&nonAdminCount)
	results = append(results, DebugResult{
		Query:  "SELECT COUNT(*) FROM user WHERE role != 'admin'",
		Result: nonAdminCount,
	})

	// 5. Sample test_attempts - raw data
	type SampleAttempt struct {
		ID          string  `gorm:"column:id"`
		UserID      string  `gorm:"column:userId"`
		Score       float64 `gorm:"column:score"`
		SubmittedAt string  `gorm:"column:submittedAt"`
	}
	var samples []SampleAttempt
	database.DB.Table("test_attempts").
		Select("id, userId, score, submittedAt").
		Order("submittedAt DESC").
		Limit(5).
		Scan(&samples)
	results = append(results, DebugResult{
		Query:  "SELECT id, userId, score, submittedAt FROM test_attempts ORDER BY submittedAt DESC LIMIT 5",
		Result: samples,
	})

	// 6. Sample users
	type SampleUser struct {
		ID       string `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Role     string `gorm:"column:role"`
	}
	var userSamples []SampleUser
	database.DB.Table("user").
		Select("id, username, role").
		Limit(5).
		Scan(&userSamples)
	results = append(results, DebugResult{
		Query:  "SELECT id, username, role FROM user LIMIT 5",
		Result: userSamples,
	})

	// 7. Try the actual leaderboard query to see what it returns
	type LeaderboardRow struct {
		UserID         string  `gorm:"column:user_id"`
		Username       string  `gorm:"column:username"`
		TestsCount     int     `gorm:"column:tests_count"`
		BestScore      float64 `gorm:"column:best_score"`
		TotalScore     float64 `gorm:"column:total_score"`
		EarliestSubmit string  `gorm:"column:earliest_submit"`
	}
	var leaderboardTest []LeaderboardRow
	testQuery := database.DB.Table("test_attempts").
		Select("test_attempts.userId as user_id, "+
			"user.username as username, "+
			"COUNT(DISTINCT test_attempts.id) as tests_count, "+
			"MAX(test_attempts.score) as best_score, "+
			"SUM(test_attempts.score) as total_score, "+
			"MIN(test_attempts.submittedAt) as earliest_submit").
		Joins("JOIN user ON user.id = test_attempts.userId").
		Where("test_attempts.submittedAt IS NOT NULL").
		Where("user.role != ?", "admin").
		Group("test_attempts.userId, user.username").
		Order("total_score DESC").
		Limit(5)
	
	testQuery.Scan(&leaderboardTest)
	results = append(results, DebugResult{
		Query:  "Leaderboard query (with all conditions)",
		Result: leaderboardTest,
	})

	log.Printf("[LeaderboardDebug] Total=%d Submitted=%d Users=%d NonAdmin=%d",
		totalCount, submittedCount, userCount, nonAdminCount)

	c.JSON(http.StatusOK, gin.H{
		"message": "Debug queries executed",
		"results": results,
		"summary": gin.H{
			"totalAttempts":     totalCount,
			"submittedAttempts": submittedCount,
			"totalUsers":        userCount,
			"nonAdminUsers":     nonAdminCount,
		},
	})
}

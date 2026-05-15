package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
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
	type rawRow struct {
		UserID         string  `db:"user_id"`
		Username       string  `db:"username"`
		TotalScore     int     `db:"total_score"`
		TestsCompleted int     `db:"tests_completed"`
		BestScore      int     `db:"best_score"`
		EarliestSubmit string  `db:"earliest_submit"`
	}

	// Get the actual table name from the model
	var attempt models.Attempt
	tableName := database.DB.NamingStrategy.TableName(attempt.TableName())
	if tableName == "" {
		tableName = "attempts" // fallback
	}

	var user models.User
	userTableName := database.DB.NamingStrategy.TableName(user.TableName())
	if userTableName == "" {
		userTableName = "user" // fallback
	}

	// Build raw SQL query
	sqlQuery := `
		SELECT 
			a.userId as user_id,
			u.username,
			SUM(a.score) as total_score,
			COUNT(DISTINCT a.id) as tests_completed,
			MAX(a.score) as best_score,
			MIN(a.completedAt) as earliest_submit
		FROM ` + tableName + ` a
		JOIN ` + userTableName + ` u ON u.id = a.userId
		WHERE a.completedAt IS NOT NULL
		AND u.role != 'admin'
		GROUP BY a.userId, u.username
		ORDER BY total_score DESC, earliest_submit ASC
		LIMIT 100
	`

	log.Printf("[Leaderboard] Using table: %s, user table: %s", tableName, userTableName)
	log.Printf("[Leaderboard] SQL Query:\n%s", sqlQuery)

	var rows []rawRow
	if err := database.DB.Raw(sqlQuery).Scan(&rows).Error; err != nil {
		log.Printf("[Leaderboard] Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard", "details": err.Error()})
		return
	}

	log.Printf("[Leaderboard] Found %d users", len(rows))

	entries := make([]GlobalLeaderboardEntry, len(rows))
	totalUsers := len(rows)

	for i, r := range rows {
		rank := i + 1
		// Calculate average percentage based on total questions
		avgPct := float64(0)
		if r.TestsCompleted > 0 && r.TotalScore > 0 {
			avgPct = (float64(r.TotalScore) / float64(r.TestsCompleted))
		}

		// Assign same tier to users with same rank (ties share a tier)
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

		entries[i] = GlobalLeaderboardEntry{
			Rank:           rank,
			UserID:         r.UserID,
			Username:       r.Username,
			TotalScore:     r.TotalScore,
			TestsCompleted: r.TestsCompleted,
			AvgPercentage:  float64(int(avgPct*100)) / 100,
			HighScore:      r.BestScore,
			Tier:           tier,
		}

		log.Printf("[Leaderboard] Entry %d: User=%s (ID=%s) Score=%d Tests=%d Best=%d",
			rank, r.Username, r.UserID, r.TotalScore, r.TestsCompleted, r.BestScore)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":    entries,
		"totalUsers": totalUsers,
	})
}

// GetLeaderboardDebug → GET /api/leaderboard/debug
// Debug endpoint to check if attempts exist in the database
func GetLeaderboardDebug(c *gin.Context) {
	var attempt models.Attempt
	tableName := database.DB.NamingStrategy.TableName(attempt.TableName())
	if tableName == "" {
		tableName = "attempts"
	}

	// Count total attempts
	var totalCount int64
	database.DB.Table(tableName).Count(&totalCount)

	// Count completed attempts
	var completedCount int64
	database.DB.Table(tableName).Where("completedAt IS NOT NULL").Count(&completedCount)

	// Count users
	var user models.User
	userTableName := database.DB.NamingStrategy.TableName(user.TableName())
	if userTableName == "" {
		userTableName = "user"
	}
	var userCount int64
	database.DB.Table(userTableName).Count(&userCount)

	// Get sample attempts
	type sampleAttempt struct {
		ID          string `db:"id"`
		UserID      string `db:"userId"`
		Score       int    `db:"score"`
		CompletedAt string `db:"completedAt"`
	}
	var samples []sampleAttempt
	database.DB.Table(tableName).
		Select("id, userId, score, completedAt").
		Where("completedAt IS NOT NULL").
		Order("completedAt DESC").
		Limit(5).
		Scan(&samples)

	log.Printf("[LeaderboardDebug] Table=%s Total=%d Completed=%d Users=%d Samples=%d",
		tableName, totalCount, completedCount, userCount, len(samples))

	c.JSON(http.StatusOK, gin.H{
		"tableName":       tableName,
		"userTableName":   userTableName,
		"totalAttempts":   totalCount,
		"completedAttempts": completedCount,
		"totalUsers":      userCount,
		"sampleAttempts":  samples,
		"message":         "Debug info retrieved successfully",
	})
}

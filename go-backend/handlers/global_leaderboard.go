package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		UserID          string  `gorm:"column:user_id"`
		Username        string  `gorm:"column:username"`
		TotalScore      int     `gorm:"column:total_score"`
		TestsCompleted  int     `gorm:"column:tests_completed"`
		AvgScore        float64 `gorm:"column:avg_score"`
		HighScore       int     `gorm:"column:high_score"`
		EarliestSubmit  string  `gorm:"column:earliest_submit"`
	}

	var rows []rawRow
	query := database.DB.Table("attempts").
		Select("attempts.userId as user_id, "+
			"user.username, "+
			"SUM(attempts.score) as total_score, "+
			"COUNT(DISTINCT attempts.id) as tests_completed, "+
			"AVG(attempts.score) as avg_score, "+
			"MAX(attempts.score) as high_score, "+
			"MIN(attempts.completedAt) as earliest_submit").
		Joins("JOIN user ON user.id = attempts.userId").
		Where("attempts.completedAt IS NOT NULL").
		Where("user.role != 'admin'").
		Group("attempts.userId, user.username").
		Order("total_score DESC, earliest_submit ASC").
		Limit(100)
	
	// Debug: Log the SQL query
	sqlStr := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Scan(&rows)
	})
	log.Printf("[Leaderboard] SQL: %s", sqlStr)
	
	if err := query.Scan(&rows).Error; err != nil {
		log.Printf("[Leaderboard] Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	
	log.Printf("[Leaderboard] Found %d users", len(rows))

	entries := make([]GlobalLeaderboardEntry, len(rows))
	totalUsers := len(rows)

	for i, r := range rows {
		rank := i + 1
		// Calculate average percentage based on total questions
		avgPct := float64(0)
		if r.TestsCompleted > 0 {
			avgPct = r.AvgScore // This is already the average score per test
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
			HighScore:      r.HighScore,
			Tier:           tier,
		}
		
		log.Printf("[Leaderboard] Entry %d: User=%s Score=%d Tests=%d", rank, r.Username, r.TotalScore, r.TestsCompleted)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":    entries,
		"totalUsers": totalUsers,
	})
}

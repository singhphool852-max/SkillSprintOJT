package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
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
// Ranking logic: total score DESC, then tests completed DESC as tiebreak.
func GetGlobalLeaderboard(c *gin.Context) {
	type rawRow struct {
		UserID         string  `json:"userId"`
		Username       string  `json:"username"`
		TotalScore     int     `json:"totalScore"`
		TestsCompleted int     `json:"testsCompleted"`
		AvgScore       float64 `json:"avgScore"`
		HighScore      int     `json:"highScore"`
		TotalMaxScore  int     `json:"totalMaxScore"`
	}

	var rows []rawRow
	database.DB.Table("test_attempts").
		Select("test_attempts.userId as user_id, "+
			"user.username, "+
			"SUM(test_attempts.score) as total_score, "+
			"COUNT(DISTINCT test_attempts.id) as tests_completed, "+
			"AVG(test_attempts.score) as avg_score, "+
			"MAX(test_attempts.score) as high_score, "+
			"SUM(tests.maxScore) as total_max_score").
		Joins("JOIN user ON user.id = test_attempts.userId").
		Joins("JOIN tests ON tests.id = test_attempts.testId").
		Where("test_attempts.submittedAt IS NOT NULL").
		Where("user.role != 'admin'"). // Exclude admins from leaderboard
		Group("test_attempts.userId, user.username").
		Order("total_score DESC, tests_completed DESC").
		Limit(100).
		Scan(&rows)

	entries := make([]GlobalLeaderboardEntry, len(rows))
	totalUsers := len(rows)

	for i, r := range rows {
		rank := i + 1
		avgPct := float64(0)
		if r.TotalMaxScore > 0 {
			avgPct = float64(r.TotalScore) / float64(r.TotalMaxScore) * 100
		}

		// Calculate tier based on rank position
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
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":    entries,
		"totalUsers": totalUsers,
	})
}

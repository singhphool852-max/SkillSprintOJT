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
// Ranking logic: total score DESC, then earliest completedAt as tiebreak (first to submit wins).
func GetGlobalLeaderboard(c *gin.Context) {
	type rawRow struct {
		UserID          string  `json:"userId"`
		Username        string  `json:"username"`
		TotalScore      int     `json:"totalScore"`
		TestsCompleted  int     `json:"testsCompleted"`
		AvgScore        float64 `json:"avgScore"`
		HighScore       int     `json:"highScore"`
		TotalMaxScore   int     `json:"totalMaxScore"`
		EarliestSubmit  string  `json:"earliestSubmit"`
	}

	var rows []rawRow
	database.DB.Table("attempts").
		Select("attempts.userId as user_id, "+
			"user.username, "+
			"SUM(attempts.score) as total_score, "+
			"COUNT(DISTINCT attempts.id) as tests_completed, "+
			"AVG(attempts.score) as avg_score, "+
			"MAX(attempts.score) as high_score, "+
			"SUM(quizzes.maxScore) as total_max_score, "+
			"MIN(attempts.completedAt) as earliest_submit").
		Joins("JOIN user ON user.id = attempts.userId").
		Joins("LEFT JOIN quizzes ON quizzes.id = attempts.quizId").
		Where("attempts.completedAt IS NOT NULL").
		Where("user.role != 'admin'").
		Group("attempts.userId, user.username").
		Order("total_score DESC, earliest_submit ASC").
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
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":    entries,
		"totalUsers": totalUsers,
	})
}

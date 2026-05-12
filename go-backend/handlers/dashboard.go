package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"math"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

// ══════════════════════════════════════════════
// User Dashboard — Personal Analytics
// GET /api/dashboard/full
// ══════════════════════════════════════════════

// TopicAnalysis holds per-topic performance metrics.
type TopicAnalysis struct {
	TopicID     string  `json:"topicId"`
	TopicName   string  `json:"topicName"`
	TestsTaken  int     `json:"testsTaken"`
	AvgScore    float64 `json:"avgScore"`
	MaxScore    int     `json:"maxScore"`
	TotalScore  int     `json:"totalScore"`
	MaxPossible int     `json:"maxPossible"`
	Percentage  float64 `json:"percentage"`
}

// PerformancePoint represents one data point in the performance trend chart.
type PerformancePoint struct {
	TestTitle  string  `json:"testTitle"`
	Score      int     `json:"score"`
	Percentage float64 `json:"percentage"`
	Date       string  `json:"date"`
}

// CompletedTest is a summary of a finished test for the recent logs section.
type CompletedTest struct {
	AttemptID       string  `json:"attemptId"`
	TestID          string  `json:"testId"`
	TestTitle       string  `json:"testTitle"`
	TopicName       string  `json:"topicName"`
	Score           int     `json:"score"`
	MaxPossible     int     `json:"maxPossible"`
	Percentage      float64 `json:"percentage"`
	IsAutoSubmitted bool    `json:"isAutoSubmitted"`
	SubmittedAt     string  `json:"submittedAt"`
}

// GetUserDashboardFull → GET /api/dashboard/full
// Returns comprehensive personal analytics for the logged-in user.
func GetUserDashboardFull(c *gin.Context) {
	userID, _ := c.Get("userID")

	// ── 1. Basic stats ─────────────────────────────────
	var overallStats struct {
		TotalAttempts  int     `json:"totalAttempts"`
		SubmittedCount int     `json:"submittedCount"`
		HighScore      int     `json:"highScore"`
		AvgScore       float64 `json:"avgScore"`
		TotalScore     int     `json:"totalScore"`
	}
	database.DB.Table("test_attempts").
		Select("COUNT(*) as total_attempts, "+
			"SUM(CASE WHEN submittedAt IS NOT NULL AND submittedAt != '' THEN 1 ELSE 0 END) as submitted_count, "+
			"MAX(score) as high_score, "+
			"AVG(CASE WHEN submittedAt IS NOT NULL AND submittedAt != '' THEN score END) as avg_score, "+
			"SUM(CASE WHEN submittedAt IS NOT NULL AND submittedAt != '' THEN score ELSE 0 END) as total_score").
		Where("userId = ?", userID).
		Scan(&overallStats)

	// ── 2. Topic-wise analysis ─────────────────────────
	type topicRow struct {
		TopicID     string  `json:"topicId"`
		TopicName   string  `json:"topicName"`
		TestsTaken  int     `json:"testsTaken"`
		AvgScore    float64 `json:"avgScore"`
		MaxScore    int     `json:"maxScore"`
		TotalScore  int     `json:"totalScore"`
		MaxPossible int     `json:"maxPossible"`
	}
	var topicRows []topicRow
	database.DB.Table("test_attempts").
		Select("topics.id as topic_id, topics.name as topic_name, "+
			"COUNT(DISTINCT test_attempts.id) as tests_taken, "+
			"AVG(test_attempts.score) as avg_score, "+
			"MAX(test_attempts.score) as max_score, "+
			"SUM(test_attempts.score) as total_score, "+
			"SUM(tests.maxScore) as max_possible").
		Joins("JOIN tests ON tests.id = test_attempts.testId").
		Joins("LEFT JOIN topics ON topics.id = tests.topicId").
		Where("test_attempts.userId = ? AND test_attempts.submittedAt IS NOT NULL AND test_attempts.submittedAt != ''", userID).
		Group("topics.id, topics.name").
		Scan(&topicRows)

	topicAnalysis := make([]TopicAnalysis, 0, len(topicRows))
	for _, tr := range topicRows {
		pct := float64(0)
		if tr.MaxPossible > 0 {
			pct = math.Round(float64(tr.TotalScore)/float64(tr.MaxPossible)*10000) / 100
		}
		name := tr.TopicName
		if name == "" {
			name = "General"
		}
		topicAnalysis = append(topicAnalysis, TopicAnalysis{
			TopicID:     tr.TopicID,
			TopicName:   name,
			TestsTaken:  tr.TestsTaken,
			AvgScore:    math.Round(tr.AvgScore*100) / 100,
			MaxScore:    tr.MaxScore,
			TotalScore:  tr.TotalScore,
			MaxPossible: tr.MaxPossible,
			Percentage:  pct,
		})
	}

	// ── 3. Strong & weak points ────────────────────────
	// Sort topics by percentage to find best/worst
	sorted := make([]TopicAnalysis, len(topicAnalysis))
	copy(sorted, topicAnalysis)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Percentage > sorted[j].Percentage
	})

	strongPoints := []string{}
	weakPoints := []string{}
	for i, ta := range sorted {
		if i < 3 && ta.Percentage >= 50 {
			strongPoints = append(strongPoints, ta.TopicName)
		}
	}
	// Weak = bottom performers with < 60%
	for i := len(sorted) - 1; i >= 0 && len(weakPoints) < 3; i-- {
		if sorted[i].Percentage < 60 && sorted[i].TestsTaken > 0 {
			// Don't add if already in strong
			alreadyStrong := false
			for _, s := range strongPoints {
				if s == sorted[i].TopicName {
					alreadyStrong = true
					break
				}
			}
			if !alreadyStrong {
				weakPoints = append(weakPoints, sorted[i].TopicName)
			}
		}
	}

	// ── 4. Performance trend (last 15 tests) ──────────
	type trendRow struct {
		TestTitle   string `json:"testTitle"`
		Score       int    `json:"score"`
		MaxPossible int    `json:"maxPossible"`
		SubmittedAt string `json:"submittedAt"`
	}
	var trendRows []trendRow
	database.DB.Table("test_attempts").
		Select("tests.title as test_title, test_attempts.score, tests.maxScore as max_possible, test_attempts.submittedAt as submitted_at").
		Joins("JOIN tests ON tests.id = test_attempts.testId").
		Where("test_attempts.userId = ? AND test_attempts.submittedAt IS NOT NULL AND test_attempts.submittedAt != ''", userID).
		Order("test_attempts.submittedAt ASC").
		Limit(15).
		Scan(&trendRows)

	performanceTrend := make([]PerformancePoint, 0, len(trendRows))
	for _, tr := range trendRows {
		pct := float64(0)
		if tr.MaxPossible > 0 {
			pct = math.Round(float64(tr.Score) / float64(tr.MaxPossible) * 10000) / 100
		}
		performanceTrend = append(performanceTrend, PerformancePoint{
			TestTitle:  tr.TestTitle,
			Score:      tr.Score,
			Percentage: pct,
			Date:       tr.SubmittedAt,
		})
	}

	// ── 5. Completed tests (recent 10) ─────────────────
	type completedRow struct {
		AttemptID       string `json:"attemptId"`
		TestID          string `json:"testId"`
		TestTitle       string `json:"testTitle"`
		TopicName       string `json:"topicName"`
		Score           int    `json:"score"`
		MaxPossible     int    `json:"maxPossible"`
		IsAutoSubmitted bool   `json:"isAutoSubmitted"`
		SubmittedAt     string `json:"submittedAt"`
	}
	var completedRows []completedRow
	database.DB.Table("test_attempts").
		Select("test_attempts.id as attempt_id, tests.id as test_id, tests.title as test_title, "+
			"COALESCE(topics.name, 'General') as topic_name, "+
			"test_attempts.score, tests.maxScore as max_possible, "+
			"test_attempts.isAutoSubmitted as is_auto_submitted, "+
			"test_attempts.submittedAt as submitted_at").
		Joins("JOIN tests ON tests.id = test_attempts.testId").
		Joins("LEFT JOIN topics ON topics.id = tests.topicId").
		Where("test_attempts.userId = ? AND test_attempts.submittedAt IS NOT NULL AND test_attempts.submittedAt != ''", userID).
		Order("test_attempts.submittedAt DESC").
		Limit(10).
		Scan(&completedRows)

	completedTests := make([]CompletedTest, 0, len(completedRows))
	for _, cr := range completedRows {
		pct := float64(0)
		if cr.MaxPossible > 0 {
			pct = math.Round(float64(cr.Score) / float64(cr.MaxPossible) * 10000) / 100
		}
		completedTests = append(completedTests, CompletedTest{
			AttemptID:       cr.AttemptID,
			TestID:          cr.TestID,
			TestTitle:       cr.TestTitle,
			TopicName:       cr.TopicName,
			Score:           cr.Score,
			MaxPossible:     cr.MaxPossible,
			Percentage:      pct,
			IsAutoSubmitted: cr.IsAutoSubmitted,
			SubmittedAt:     cr.SubmittedAt,
		})
	}

	// ── 6. Global rank ─────────────────────────────────
	// Rank = number of users with higher total score + 1
	var globalRank int64
	database.DB.Table("test_attempts").
		Select("COUNT(DISTINCT userId)").
		Where("submittedAt IS NOT NULL AND submittedAt != ''").
		Group("userId").
		Having("SUM(score) > ?", overallStats.TotalScore).
		Count(&globalRank)
	// +1 for current user's position
	rank := int(globalRank) + 1

	// Total participants
	var totalParticipants int64
	database.DB.Table("test_attempts").
		Where("submittedAt IS NOT NULL AND submittedAt != ''").
		Distinct("userId").
		Count(&totalParticipants)

	// ── 7. Rank tier ───────────────────────────────────
	tier := "ROOKIE"
	if totalParticipants > 0 {
		percentile := float64(rank) / float64(totalParticipants) * 100
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
	if overallStats.SubmittedCount == 0 {
		tier = "UNRANKED"
		rank = 0
	}

	// Active test count
	var activeTestCount int64
	database.DB.Model(&models.Test{}).Where("isActive = ? AND isPublished = ?", true, true).Count(&activeTestCount)

	// ── 8. Mistake stats (NEW) ────────────────────────
	var unmasteredMistakes int64
	database.DB.Model(&models.UserWrongQuestion{}).
		Where("userId = ? AND (masteredAt IS NULL OR masteredAt < '0001-01-02')", userID).
		Count(&unmasteredMistakes)

	var weakTopicStats []models.UserTopicStats
	database.DB.Where("userId = ?", userID).
		Order("accuracyPercent ASC").
		Limit(5).
		Find(&weakTopicStats)

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"totalAttempts":      overallStats.SubmittedCount,
			"highScore":          overallStats.HighScore,
			"avgScore":           math.Round(overallStats.AvgScore*100) / 100,
			"totalScore":         overallStats.TotalScore,
			"unmasteredMistakes": unmasteredMistakes,
		},
		"globalRank":        rank,
		"totalParticipants": totalParticipants,
		"tier":              tier,
		"topicAnalysis":     topicAnalysis,
		"strongPoints":      strongPoints,
		"weakPoints":        weakPoints,
		"weakTopicStats":    weakTopicStats,
		"performanceTrend":  performanceTrend,
		"completedTests":    completedTests,
		"activeTestCount":   activeTestCount,
	})
}

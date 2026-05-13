package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// GetTestResult → GET /api/results/:attemptId
// Returns a detailed result for a completed test attempt.
// Computes and persists a TestResult if one doesn't exist yet.
// ──────────────────────────────────────────────
func GetTestResult(c *gin.Context) {
	attemptID := c.Param("attemptId")
	userID, _ := c.Get("userID")

	// Verify attempt belongs to user and is submitted
	var attempt models.TestAttempt
	if err := database.DB.Preload("Test").Where("id = ? AND userId = ?", attemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	if attempt.SubmittedAt == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attempt has not been submitted yet"})
		return
	}

	// Check if result already computed
	var result models.TestResult
	if err := database.DB.Where("attemptId = ?", attemptID).First(&result).Error; err == nil {
		// Already computed — return it
		c.JSON(http.StatusOK, result)
		return
	}

	// Compute result from submissions
	var submissions []models.TestSubmission
	database.DB.Where("attemptId = ?", attemptID).Find(&submissions)

	var questions []models.TestQuestion
	database.DB.Where("testId = ?", attempt.TestID).Find(&questions)

	mcqCorrect := 0
	mcqTotal := 0
	codingPassed := 0
	codingTotal := 0
	maxPossible := 0

	for _, q := range questions {
		maxPossible += q.Points
		if q.Type == "mcq" {
			mcqTotal++
		} else {
			codingTotal++
		}
	}

	for _, sub := range submissions {
		if sub.Type == "mcq" {
			if sub.Verdict == "accepted" {
				mcqCorrect++
			}
		} else {
			if sub.Verdict == "accepted" {
				codingPassed++
			}
		}
	}

	percentage := float64(0)
	if maxPossible > 0 {
		percentage = float64(attempt.Score) / float64(maxPossible) * 100
	}

	// Calculate rank among submitted attempts for this test
	var rank int64
	database.DB.Model(&models.TestAttempt{}).
		Where("testId = ? AND submittedAt IS NOT NULL AND submittedAt != '' AND score > ?", attempt.TestID, attempt.Score).
		Count(&rank)

	result = models.TestResult{
		ID:              uuid.New().String(),
		AttemptID:       attemptID,
		UserID:          userID.(string),
		TestID:          attempt.TestID,
		TotalScore:      attempt.Score,
		MaxPossible:     maxPossible,
		Percentage:      percentage,
		Rank:            int(rank) + 1,
		MCQCorrect:      mcqCorrect,
		MCQTotal:        mcqTotal,
		CodingPassed:    codingPassed,
		CodingTotal:     codingTotal,
		IsAutoSubmitted: attempt.IsAutoSubmitted,
	}

	database.DB.Create(&result)

	c.JSON(http.StatusOK, result)
}

// ──────────────────────────────────────────────
// ListUserResults → GET /api/results
// Returns all results for the authenticated user.
// ──────────────────────────────────────────────
func ListUserResults(c *gin.Context) {
	userID, _ := c.Get("userID")

	// First, compute results for any submitted attempts that don't have one yet
	var unprocessedAttempts []models.TestAttempt
	database.DB.Where("userId = ? AND submittedAt IS NOT NULL AND submittedAt != ''", userID).Find(&unprocessedAttempts)

	for _, attempt := range unprocessedAttempts {
		var existingResult models.TestResult
		if err := database.DB.Where("attemptId = ?", attempt.ID).First(&existingResult).Error; err != nil {
			// No result yet — compute one
			computeAndSaveResult(attempt)
		}
	}

	// Now fetch all results
	var results []models.TestResult
	if err := database.DB.Preload("Test").Where("userId = ?", userID).Order("calculatedAt desc").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// computeAndSaveResult creates a TestResult row from a submitted attempt.
func computeAndSaveResult(attempt models.TestAttempt) {
	var submissions []models.TestSubmission
	database.DB.Where("attemptId = ?", attempt.ID).Find(&submissions)

	var questions []models.TestQuestion
	database.DB.Where("testId = ?", attempt.TestID).Find(&questions)

	mcqCorrect := 0
	mcqTotal := 0
	codingPassed := 0
	codingTotal := 0
	maxPossible := 0

	for _, q := range questions {
		maxPossible += q.Points
		if q.Type == "mcq" {
			mcqTotal++
		} else {
			codingTotal++
		}
	}

	for _, sub := range submissions {
		if sub.Type == "mcq" && sub.Verdict == "accepted" {
			mcqCorrect++
		}
		if sub.Type == "coding" && sub.Verdict == "accepted" {
			codingPassed++
		}
	}

	percentage := float64(0)
	if maxPossible > 0 {
		percentage = float64(attempt.Score) / float64(maxPossible) * 100
	}

	var rank int64
	database.DB.Model(&models.TestAttempt{}).
		Where("testId = ? AND submittedAt IS NOT NULL AND submittedAt != '' AND score > ?", attempt.TestID, attempt.Score).
		Count(&rank)

	result := models.TestResult{
		ID:              uuid.New().String(),
		AttemptID:       attempt.ID,
		UserID:          attempt.UserID,
		TestID:          attempt.TestID,
		TotalScore:      attempt.Score,
		MaxPossible:     maxPossible,
		Percentage:      percentage,
		Rank:            int(rank) + 1,
		MCQCorrect:      mcqCorrect,
		MCQTotal:        mcqTotal,
		CodingPassed:    codingPassed,
		CodingTotal:     codingTotal,
		IsAutoSubmitted: attempt.IsAutoSubmitted,
	}

	database.DB.Create(&result)
}

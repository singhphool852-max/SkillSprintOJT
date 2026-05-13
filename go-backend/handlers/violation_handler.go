package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// LogViolation → POST /api/arena/violations
// Logs an anti-cheat violation and increments the attempt's ViolationCount.
// Returns the updated count and whether auto-submit should be triggered.
// ──────────────────────────────────────────────
func LogViolation(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		AttemptID     string `json:"attemptId" binding:"required"`
		TestID        string `json:"testId" binding:"required"`
		ViolationType string `json:"violationType" binding:"required"`
		RemainingTime int    `json:"remainingTime"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attemptId, testId, and violationType are required"})
		return
	}

	// Verify attempt belongs to user and is not already submitted
	var attempt models.TestAttempt
	if err := database.DB.Where("id = ? AND userId = ?", req.AttemptID, userID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	if attempt.SubmittedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attempt already submitted"})
		return
	}

	// Log the violation
	violation := models.TestViolation{
		ID:            uuid.New().String(),
		AttemptID:     req.AttemptID,
		UserID:        userID.(string),
		TestID:        req.TestID,
		ViolationType: req.ViolationType,
		Timestamp:     time.Now(),
		RemainingTime: req.RemainingTime,
	}

	if err := database.DB.Create(&violation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log violation"})
		return
	}

	// Increment violation count on the attempt
	attempt.ViolationCount++
	database.DB.Model(&attempt).Update("violationCount", attempt.ViolationCount)

	autoSubmit := attempt.ViolationCount >= 3

	c.JSON(http.StatusOK, gin.H{
		"violationCount": attempt.ViolationCount,
		"autoSubmit":     autoSubmit,
	})
}

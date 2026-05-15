package handlers

import (
	"log"

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

	log.Printf("[VIOLATION] Received: userID=%s attemptID=%s testID=%s type=%s", 
		userID, req.AttemptID, req.TestID, req.ViolationType)

	// Verify attempt belongs to user and is not already submitted
	var attempt models.TestAttempt
	if err := database.DB.Where("id = ? AND userId = ?", req.AttemptID, userID).First(&attempt).Error; err != nil {
		log.Printf("[VIOLATION] Attempt not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	if attempt.SubmittedAt != nil {
		log.Printf("[VIOLATION] Attempt already submitted")
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
		log.Printf("[VIOLATION] Failed to save: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log violation"})
		return
	}

	log.Printf("[VIOLATION] Saved successfully: id=%s userID=%s testID=%s type=%s", 
		violation.ID, violation.UserID, violation.TestID, violation.ViolationType)

	// Increment violation count on the attempt
	attempt.ViolationCount++
	database.DB.Model(&attempt).Update("violationCount", attempt.ViolationCount)

	autoSubmit := attempt.ViolationCount >= 3

	log.Printf("[VIOLATION] Updated attempt: count=%d autoSubmit=%v", 
		attempt.ViolationCount, autoSubmit)

	c.JSON(http.StatusOK, gin.H{
		"violationCount": attempt.ViolationCount,
		"autoSubmit":     autoSubmit,
	})
}

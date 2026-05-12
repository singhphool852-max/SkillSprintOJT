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
// Request structs
// ──────────────────────────────────────────────

type CreateTestRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	TopicID         string    `json:"topicId"`
	Difficulty      string    `json:"difficulty"`
	StartTime       time.Time `json:"startTime" binding:"required"`
	DurationSeconds int       `json:"durationSeconds" binding:"required"`
}

type AddQuestionRequest struct {
	Type        string `json:"type" binding:"required"` // "mcq" or "coding"
	Position    int    `json:"position"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Points      int    `json:"points" binding:"required"`

	// MCQ-specific: array of options
	Options []AddMCQOptionPayload `json:"options,omitempty"`

	// Coding-specific: inline detail
	Constraints string `json:"constraints,omitempty"`
	StarterCode string `json:"starterCode,omitempty"`
	TimeLimitMs int    `json:"timeLimitMs,omitempty"`
}

type AddMCQOptionPayload struct {
	OptionText string `json:"optionText" binding:"required"`
	IsCorrect  bool   `json:"isCorrect"`
}

type AddTestcaseRequest struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expectedOutput" binding:"required"`
	IsHidden       bool   `json:"isHidden"`
}

// ──────────────────────────────────────────────
// CreateTest → POST /api/admin/tests
// ──────────────────────────────────────────────
func CreateTest(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req CreateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, startTime, and durationSeconds are required"})
		return
	}

	test := models.Test{
		ID:              uuid.New().String(),
		Title:           req.Title,
		Description:     req.Description,
		TopicID:         req.TopicID,
		Difficulty:      req.Difficulty,
		StartTime:       req.StartTime,
		DurationSeconds: req.DurationSeconds,
		IsPublished:     false,
		CreatedBy:       userID.(string),
	}

	if err := database.DB.Create(&test).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test"})
		return
	}

	c.JSON(http.StatusOK, test)
}

// ──────────────────────────────────────────────
// ListTests → GET /api/admin/tests
// ──────────────────────────────────────────────
func ListTests(c *gin.Context) {
	var tests []models.Test
	// DEEP PRELOAD: Ensure questions and ALL their nested associations (options, details, cases) are loaded.
	// This prevents the "questions disappearing after refresh" issue in the UI.
	query := database.DB.
		Preload("Creator").
		Preload("Topic").
		Preload("Questions").
		Preload("Questions.MCQOptions").
		Preload("Questions.CodingDetail").
		Preload("Questions.TestCases").
		Order("createdAt desc")

	// By default, filter out soft-deleted tests
	if c.Query("includeDeleted") != "true" {
		query = query.Where("deletedAt IS NULL")
	}

	if err := query.Find(&tests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tests"})
		return
	}
	c.JSON(http.StatusOK, tests)
}

// ──────────────────────────────────────────────
// PublishTest → PATCH /api/admin/tests/:id/publish
// ──────────────────────────────────────────────
func PublishTest(c *gin.Context) {
	testID := c.Param("id")

	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	var req struct {
		IsPublished bool `json:"isPublished"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default: toggle
		test.IsPublished = !test.IsPublished
	} else {
		test.IsPublished = req.IsPublished
	}

	if err := database.DB.Save(&test).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update publish status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Publish status updated", "testId": test.ID, "isPublished": test.IsPublished})
}

// ──────────────────────────────────────────────
// ActivateTest → PATCH /api/admin/tests/:id/activate
// Sets this test as the ONLY active test (deactivates all others).
// Guard: test must be published.
// ──────────────────────────────────────────────
func ActivateTest(c *gin.Context) {
	testID := c.Param("id")

	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	if !test.IsPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test must be published before activating"})
		return
	}

	// Atomic: deactivate all, then activate this one.
	// Reset startTime to NOW so the test window begins from activation.
	now := time.Now()
	tx := database.DB.Begin()
	tx.Model(&models.Test{}).Where("1 = 1").Update("isActive", false)
	tx.Model(&test).Updates(map[string]interface{}{
		"isActive":  true,
		"startTime": now,
	})
	tx.Commit()

	// Reload for response
	database.DB.Where("id = ?", testID).First(&test)

	c.JSON(http.StatusOK, gin.H{"message": "Test activated", "testId": test.ID, "isActive": true, "startTime": test.StartTime})
}

// ──────────────────────────────────────────────
// DeleteTest → DELETE /api/admin/tests/:id
// Transactional cleanup of a test and all its questions/options/details/cases.
// ──────────────────────────────────────────────
func DeleteTest(c *gin.Context) {
	testID := c.Param("id")

	// Verify test exists
	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	// Begin transaction for safe cleanup
	tx := database.DB.Begin()

	// 1. Find all questions for this test
	var questionIDs []string
	tx.Model(&models.TestQuestion{}).Where("testId = ?", testID).Pluck("id", &questionIDs)

	if len(questionIDs) > 0 {
		// 2. Delete nested data for these questions
		tx.Where("questionId IN ?", questionIDs).Delete(&models.TestMCQOption{})
		tx.Where("questionId IN ?", questionIDs).Delete(&models.TestCodingDetail{})
		tx.Where("questionId IN ?", questionIDs).Delete(&models.TestCase{})
		
		// 3. Delete the questions themselves
		tx.Where("testId = ?", testID).Delete(&models.TestQuestion{})
	}

	// 4. Delete results and attempts (optional, but keeps DB clean)
	tx.Where("testId = ?", testID).Delete(&models.TestResult{})
	tx.Where("testId = ?", testID).Delete(&models.TestAttempt{})

	// 5. Finally, delete the test itself
	if err := tx.Delete(&test).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete test"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Test and all associated data deleted successfully"})
}

// ──────────────────────────────────────────────
// ListQuestions → GET /api/admin/tests/:id/questions
// Returns all questions for a test with full associations
// (admin view — includes isCorrect and isHidden)
// ──────────────────────────────────────────────
func ListQuestions(c *gin.Context) {
	testID := c.Param("id")

	var questions []models.TestQuestion
	if err := database.DB.Where("testId = ?", testID).
		Preload("MCQOptions").
		Preload("CodingDetail").
		Preload("TestCases").
		Order("position asc").
		Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
		return
	}

	c.JSON(http.StatusOK, questions)
}

// ──────────────────────────────────────────────
// AddQuestion → POST /api/admin/tests/:id/questions
// Handles both MCQ (with options array) and Coding
// (with constraints/starterCode/timeLimitMs).
// ──────────────────────────────────────────────
func AddQuestion(c *gin.Context) {
	testID := c.Param("id")

	// Verify the test exists
	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	var req AddQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type, title, and points are required"})
		return
	}

	if req.Type != "mcq" && req.Type != "coding" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'mcq' or 'coding'"})
		return
	}

	questionID := uuid.New().String()

	question := models.TestQuestion{
		ID:          questionID,
		TestID:      testID,
		Type:        req.Type,
		Position:    req.Position,
		Title:       req.Title,
		Description: req.Description,
		Points:      req.Points,
	}

	// Begin transaction
	tx := database.DB.Begin()

	if err := tx.Create(&question).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}

	if req.Type == "mcq" {
		// Create MCQ options
		for _, opt := range req.Options {
			option := models.TestMCQOption{
				ID:         uuid.New().String(),
				QuestionID: questionID,
				OptionText: opt.OptionText,
				IsCorrect:  opt.IsCorrect,
			}
			if err := tx.Create(&option).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create option"})
				return
			}
		}
	} else if req.Type == "coding" {
		// Create coding detail row
		detail := models.TestCodingDetail{
			ID:          uuid.New().String(),
			QuestionID:  questionID,
			Constraints: req.Constraints,
			StarterCode: req.StarterCode,
			TimeLimitMs: req.TimeLimitMs,
		}
		if err := tx.Create(&detail).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coding detail"})
			return
		}
	}

	tx.Commit()

	// Reload with associations for response
	var created models.TestQuestion
	database.DB.Preload("MCQOptions").Preload("CodingDetail").Where("id = ?", questionID).First(&created)

	c.JSON(http.StatusOK, created)
}

// ──────────────────────────────────────────────
// AddTestcase → POST /api/admin/questions/:id/testcases
// ──────────────────────────────────────────────
func AddTestcase(c *gin.Context) {
	questionID := c.Param("id")

	// Verify the question exists
	var question models.TestQuestion
	if err := database.DB.Where("id = ?", questionID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	var req AddTestcaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input and expectedOutput are required"})
		return
	}

	tc := models.TestCase{
		ID:             uuid.New().String(),
		QuestionID:     questionID,
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
		IsHidden:       req.IsHidden,
	}

	if err := database.DB.Create(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create testcase"})
		return
	}

	c.JSON(http.StatusOK, tc)
}

// ──────────────────────────────────────────────
// GetTestDetail → GET /api/admin/tests/:id
// Returns a single test with all questions and associations.
// ──────────────────────────────────────────────
func GetTestDetail(c *gin.Context) {
	testID := c.Param("id")

	var test models.Test
	if err := database.DB.Preload("Creator").Preload("Topic").Preload("Questions").Preload("Questions.MCQOptions").Preload("Questions.CodingDetail").Preload("Questions.TestCases").Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	c.JSON(http.StatusOK, test)
}

// ──────────────────────────────────────────────
// UpdateTest → PUT /api/admin/tests/:id
// Updates test metadata (title, startTime, duration, topicId).
// ──────────────────────────────────────────────
func UpdateTest(c *gin.Context) {
	testID := c.Param("id")

	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	var req struct {
		Title           *string    `json:"title"`
		TopicID         *string    `json:"topicId"`
		StartTime       *time.Time `json:"startTime"`
		DurationSeconds *int       `json:"durationSeconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Title != nil {
		test.Title = *req.Title
	}
	if req.TopicID != nil {
		test.TopicID = *req.TopicID
	}
	if req.StartTime != nil {
		test.StartTime = *req.StartTime
	}
	if req.DurationSeconds != nil {
		test.DurationSeconds = *req.DurationSeconds
	}

	if err := database.DB.Save(&test).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update test"})
		return
	}

	c.JSON(http.StatusOK, test)
}

// ──────────────────────────────────────────────
// SoftDeleteTest → DELETE /api/admin/tests/:id
// Soft-deletes a test (sets deletedAt). Does NOT destroy data.
// Historical attempts, results, and submissions are preserved.
// ──────────────────────────────────────────────
func SoftDeleteTest(c *gin.Context) {
	testID := c.Param("id")
	userID, _ := c.Get("userID")

	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	if test.DeletedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test is already deleted"})
		return
	}

	now := time.Now()
	result := database.DB.Model(&models.Test{}).Where("id = ?", testID).Updates(map[string]interface{}{
		"deletedAt":   now,
		"deletedBy":   userID,
		"isActive":    false,
		"isPublished": false,
	})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test soft-deleted", "testId": testID, "deletedAt": now})
}

// ──────────────────────────────────────────────
// RestoreTest → POST /api/admin/tests/:id/restore
// Undoes a soft delete — clears deletedAt/deletedBy.
// ──────────────────────────────────────────────
func RestoreTest(c *gin.Context) {
	testID := c.Param("id")

	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	if test.DeletedAt == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test is not deleted"})
		return
	}

	// Clear soft delete fields (use raw SQL because GORM won't set NULL via Updates)
	database.DB.Exec("UPDATE tests SET deletedAt = NULL, deletedBy = '' WHERE id = ?", testID)

	c.JSON(http.StatusOK, gin.H{"message": "Test restored", "testId": testID})
}

// ──────────────────────────────────────────────
// PermanentDeleteTest → DELETE /api/admin/tests/:id/permanent
// Hard deletes a test and ALL associated data. IRREVERSIBLE.
// ──────────────────────────────────────────────
func PermanentDeleteTest(c *gin.Context) {
	testID := c.Param("id")

	var test models.Test
	if err := database.DB.Where("id = ?", testID).First(&test).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	tx := database.DB.Begin()

	// Get question IDs for cascade
	var questionIDs []string
	tx.Model(&models.TestQuestion{}).Where("testId = ?", testID).Pluck("id", &questionIDs)

	if len(questionIDs) > 0 {
		tx.Where("questionId IN ?", questionIDs).Delete(&models.TestMCQOption{})
		tx.Where("questionId IN ?", questionIDs).Delete(&models.TestCodingDetail{})
		tx.Where("questionId IN ?", questionIDs).Delete(&models.TestCase{})
		tx.Where("testId = ?", testID).Delete(&models.TestQuestion{})
	}

	// Get attempt IDs for cascade
	var attemptIDs []string
	tx.Model(&models.TestAttempt{}).Where("testId = ?", testID).Pluck("id", &attemptIDs)

	if len(attemptIDs) > 0 {
		tx.Where("attemptId IN ?", attemptIDs).Delete(&models.TestSubmission{})
		tx.Where("attemptId IN ?", attemptIDs).Delete(&models.UserWrongQuestion{})
	}

	tx.Where("testId = ?", testID).Delete(&models.TestAttempt{})
	tx.Where("testId = ?", testID).Delete(&models.TestResult{})
	tx.Where("testId = ?", testID).Delete(&models.TestViolation{})
	tx.Delete(&test)

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to permanently delete test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test permanently deleted", "testId": testID})
}

// ──────────────────────────────────────────────
// UpdateQuestion → PUT /api/admin/questions/:id
// ──────────────────────────────────────────────
func UpdateQuestion(c *gin.Context) {
	qID := c.Param("id")

	var question models.TestQuestion
	if err := database.DB.Where("id = ?", qID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Position    *int    `json:"position"`
		Points      *int    `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Title != nil {
		question.Title = *req.Title
	}
	if req.Description != nil {
		question.Description = *req.Description
	}
	if req.Position != nil {
		question.Position = *req.Position
	}
	if req.Points != nil {
		question.Points = *req.Points
	}

	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return
	}

	var updated models.TestQuestion
	database.DB.Preload("MCQOptions").Preload("CodingDetail").Preload("TestCases").Where("id = ?", qID).First(&updated)
	c.JSON(http.StatusOK, updated)
}

// ──────────────────────────────────────────────
// DeleteQuestion → DELETE /api/admin/questions/:id
// Cascade deletes options, coding detail, and test cases.
// ──────────────────────────────────────────────
func DeleteQuestion(c *gin.Context) {
	qID := c.Param("id")

	var question models.TestQuestion
	if err := database.DB.Where("id = ?", qID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	tx := database.DB.Begin()
	tx.Where("questionId = ?", qID).Delete(&models.TestMCQOption{})
	tx.Where("questionId = ?", qID).Delete(&models.TestCodingDetail{})
	tx.Where("questionId = ?", qID).Delete(&models.TestCase{})
	tx.Delete(&question)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Question deleted", "questionId": qID})
}

// ──────────────────────────────────────────────
// DeleteTestcase → DELETE /api/admin/testcases/:id
// ──────────────────────────────────────────────
func DeleteTestcase(c *gin.Context) {
	tcID := c.Param("id")

	var tc models.TestCase
	if err := database.DB.Where("id = ?", tcID).First(&tc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Testcase not found"})
		return
	}

	if err := database.DB.Delete(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete testcase"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Testcase deleted", "testcaseId": tcID})
}

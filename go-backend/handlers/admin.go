package handlers

import (
	"backend/database"
	"backend/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ─── Dashboard Stats ─────────────────────────────────────────────

func AdminGetStats(c *gin.Context) {
	var studentCount int64
	database.DB.Model(&models.User{}).Where("role = ?", "student").Count(&studentCount)

	var arenaCount int64
	database.DB.Model(&models.Arena{}).Count(&arenaCount)

	var attemptCount int64
	database.DB.Model(&models.Attempt{}).Count(&attemptCount)

	c.JSON(http.StatusOK, gin.H{
		"students": studentCount,
		"arenas":   arenaCount,
		"attempts": attemptCount,
	})
}

// ─── Student Management ──────────────────────────────────────────

func AdminListStudents(c *gin.Context) {
	var students []models.User
	database.DB.Where("role = ?", "student").Order("createdAt desc").Find(&students)
	c.JSON(http.StatusOK, students)
}

type CreateStudentRequest struct {
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func AdminCreateStudent(c *gin.Context) {
	var req CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, username, and password are required"})
		return
	}

	// Check if user already exists
	var existing models.User
	if err := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	student := models.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashed),
		Role:     "student",
	}

	if err := database.DB.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "student": student})
}

func AdminDeleteStudent(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Where("id = ? AND role = ?", id, "student").Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete student"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Arena Management ────────────────────────────────────────────

type CreateArenaRequest struct {
	Title           string `json:"title" binding:"required"`
	Difficulty      string `json:"difficulty" binding:"required"`
	CategoryID      string `json:"categoryId" binding:"required"`
	DurationMinutes int    `json:"durationMinutes" binding:"required"`
	Description     string `json:"description"`
}

func AdminCreateArena(c *gin.Context) {
	var req CreateArenaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, difficulty, categoryId, and durationMinutes are required"})
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	arenaID := uuid.New().String()

	arena := models.Arena{
		ID:              arenaID,
		Title:           req.Title,
		Slug:            slug,
		CategoryID:      req.CategoryID,
		Difficulty:      req.Difficulty,
		Status:          "live",
		DurationSeconds: req.DurationMinutes * 60,
		Description:     req.Description,
	}

	if err := database.DB.Create(&arena).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create arena"})
		return
	}

	// Auto-create an active quiz for this arena
	quiz := models.Quiz{
		ID:         uuid.New().String(),
		Title:      req.Title + " Quiz",
		ArenaID:    arenaID,
		CategoryID: req.CategoryID,
		Difficulty: req.Difficulty,
		IsActive:   true,
	}
	database.DB.Create(&quiz)

	// Reload with category
	database.DB.Preload("Category").First(&arena, "id = ?", arenaID)

	c.JSON(http.StatusOK, gin.H{"success": true, "arena": arena})
}

func AdminListArenas(c *gin.Context) {
	var arenas []models.Arena
	database.DB.Preload("Category").Order("createdAt desc").Find(&arenas)
	c.JSON(http.StatusOK, arenas)
}

func AdminDeleteArena(c *gin.Context) {
	id := c.Param("id")

	// Delete questions & options for all quizzes in this arena
	var quizzes []models.Quiz
	database.DB.Where("arenaId = ?", id).Find(&quizzes)
	for _, q := range quizzes {
		// Delete options for each question
		var questions []models.Question
		database.DB.Where("quizId = ?", q.ID).Find(&questions)
		for _, qn := range questions {
			database.DB.Where("questionId = ?", qn.ID).Delete(&models.Option{})
		}
		database.DB.Where("quizId = ?", q.ID).Delete(&models.Question{})
	}
	database.DB.Where("arenaId = ?", id).Delete(&models.Quiz{})
	database.DB.Where("id = ?", id).Delete(&models.Arena{})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Question Management ─────────────────────────────────────────

type CreateQuestionRequest struct {
	Prompt        string              `json:"prompt" binding:"required"`
	MaxScore      int                 `json:"maxScore"`
	CorrectAnswer string              `json:"correctAnswer"`
	Explanation   string              `json:"explanation"`
	Options       []CreateOptionInput `json:"options" binding:"required"`
}

type CreateOptionInput struct {
	Text      string `json:"text" binding:"required"`
	IsCorrect bool   `json:"isCorrect"`
}

func AdminAddQuestion(c *gin.Context) {
	arenaID := c.Param("id")

	var req CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt and options are required"})
		return
	}

	if req.MaxScore == 0 {
		req.MaxScore = 10
	}

	// Find the active quiz for this arena (or create one)
	var quiz models.Quiz
	if err := database.DB.Where("arenaId = ? AND isActive = ?", arenaID, true).First(&quiz).Error; err != nil {
		// No active quiz, create one
		var arena models.Arena
		if err := database.DB.Where("id = ?", arenaID).First(&arena).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Arena not found"})
			return
		}
		quiz = models.Quiz{
			ID:         uuid.New().String(),
			Title:      arena.Title + " Quiz",
			ArenaID:    arenaID,
			CategoryID: arena.CategoryID,
			Difficulty: arena.Difficulty,
			IsActive:   true,
		}
		database.DB.Create(&quiz)
	}

	// Determine correct answer from options
	correctAnswer := req.CorrectAnswer
	for _, opt := range req.Options {
		if opt.IsCorrect && correctAnswer == "" {
			correctAnswer = opt.Text
		}
	}

	question := models.Question{
		ID:            uuid.New().String(),
		QuizID:        quiz.ID,
		Prompt:        req.Prompt,
		Type:          "mcq",
		CorrectAnswer: correctAnswer,
		Explanation:   req.Explanation,
		MaxScore:      req.MaxScore,
	}

	if err := database.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}

	// Create options
	for _, opt := range req.Options {
		option := models.Option{
			ID:         uuid.New().String(),
			QuestionID: question.ID,
			Text:       opt.Text,
			IsCorrect:  opt.IsCorrect,
		}
		database.DB.Create(&option)
	}

	// Reload with options
	database.DB.Preload("Options").First(&question, "id = ?", question.ID)

	c.JSON(http.StatusOK, gin.H{"success": true, "question": question})
}

func AdminGetArenaQuestions(c *gin.Context) {
	arenaID := c.Param("id")

	var quiz models.Quiz
	if err := database.DB.Where("arenaId = ? AND isActive = ?", arenaID, true).First(&quiz).Error; err != nil {
		c.JSON(http.StatusOK, []interface{}{}) // No quiz yet, empty
		return
	}

	var questions []models.Question
	database.DB.Preload("Options").Where("quizId = ?", quiz.ID).Order("createdAt desc").Find(&questions)

	c.JSON(http.StatusOK, questions)
}

func AdminDeleteQuestion(c *gin.Context) {
	id := c.Param("id")
	// Delete options first
	database.DB.Where("questionId = ?", id).Delete(&models.Option{})
	database.DB.Where("id = ?", id).Delete(&models.Question{})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Categories (for arena creation dropdown) ────────────────────

func AdminListCategories(c *gin.Context) {
	var categories []models.QuizCategory
	database.DB.Find(&categories)
	c.JSON(http.StatusOK, categories)
}


package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetArenas(c *gin.Context) {
	var arenas []models.Arena
	if err := database.DB.Preload("Category").Find(&arenas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch arenas"})
		return
	}
	c.JSON(http.StatusOK, arenas)
}

func GetArenaDetail(c *gin.Context) {
	id := c.Param("id")
	var arena models.Arena
	if err := database.DB.Preload("Category").Where("id = ?", id).First(&arena).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arena not found"})
		return
	}
	c.JSON(http.StatusOK, arena)
}

func GetArenaQuizzes(c *gin.Context) {
	arenaID := c.Param("id")
	var quizzes []models.Quiz
	if err := database.DB.Where("arenaId = ? AND isActive = ?", arenaID, true).Find(&quizzes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quizzes"})
		return
	}
	c.JSON(http.StatusOK, quizzes)
}

func GetQuizQuestions(c *gin.Context) {
	quizID := c.Param("quizId")
	var questions []models.Question
	if err := database.DB.Preload("Options").Where("quizId = ?", quizID).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
		return
	}

	// For security, remove `IsCorrect` from Options before sending to frontend, unless evaluating
	for i := range questions {
		questions[i].CorrectAnswer = ""
		for j := range questions[i].Options {
			questions[i].Options[j].IsCorrect = false
		}
	}

	c.JSON(http.StatusOK, questions)
}

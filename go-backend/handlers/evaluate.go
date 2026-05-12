package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EvaluateRequest struct {
	Question      string `json:"question"`
	CorrectAnswer string `json:"correctAnswer"`
	UserAnswer    string `json:"userAnswer"`
	MaxScore      int    `json:"maxScore"`
}

func EvaluateAnswer(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	eval, err := services.EvaluateAnswer(req.Question, req.CorrectAnswer, req.UserAnswer, req.MaxScore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service unavailable"})
		return
	}

	c.JSON(http.StatusOK, eval)
}

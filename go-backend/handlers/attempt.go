package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubmitAttemptRequest struct {
	QuizID    string         `json:"quizId"`
	StartedAt time.Time      `json:"startedAt"`
	Answers   []SubmitAnswer `json:"answers"`
}

type SubmitAnswer struct {
	QuestionID       string `json:"questionId"`
	SelectedOptionID string `json:"selectedOptionId,omitempty"`
	WrittenAnswer    string `json:"writtenAnswer,omitempty"`
}

func SubmitAttempt(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SubmitAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	attemptID := uuid.New().String()
	totalScore := 0

	attemptAnswers := []models.AttemptAnswer{}

	for _, reqAns := range req.Answers {
		ans := models.AttemptAnswer{
			ID:               uuid.New().String(),
			AttemptID:        attemptID,
			QuestionID:       reqAns.QuestionID,
			SelectedOptionID: reqAns.SelectedOptionID,
			WrittenAnswer:    reqAns.WrittenAnswer,
		}

		// 1. Try Standard Arena Question First
		var question models.Question
		if err := database.DB.Preload("Options").Where("id = ?", reqAns.QuestionID).First(&question).Error; err == nil {
			if question.Type == "mcq" {
				isCorrect := false
				for _, opt := range question.Options {
					if opt.ID == reqAns.SelectedOptionID && opt.IsCorrect {
						isCorrect = true
						break
					}
				}
				ans.IsCorrect = isCorrect
				if isCorrect {
					ans.Score = question.MaxScore
					totalScore += question.MaxScore
				}
				ans.Explanation = question.Explanation
				ans.EvaluatedBy = "system"
			} else {
				aiEval, err := services.EvaluateAnswer(question.Prompt, question.CorrectAnswer, reqAns.WrittenAnswer, question.MaxScore)
				if err != nil {
					ans.Score = 0
					ans.Feedback = "AI Evaluation unavailable."
				} else {
					ans.Score = aiEval.Score
					ans.IsCorrect = aiEval.IsCorrect
					ans.Feedback = aiEval.Feedback
					ans.Explanation = aiEval.Explanation
				}
				totalScore += ans.Score
				ans.EvaluatedBy = "AI"
			}
			log.Printf("[AttemptSave] Standard match: ID=%s correct=%v score=%d", question.ID, ans.IsCorrect, ans.Score)

		} else {
			// 2. Try Training Question Fallback
			var tq models.TrainingQuestion
			if err := database.DB.Where("id = ?", reqAns.QuestionID).First(&tq).Error; err != nil {
				log.Printf("[AttemptSave] Warning: Question sinkhole detected: ID=%s", reqAns.QuestionID)
				continue
			}

			if tq.Type == "mcq" {
				var optArr []string
				json.Unmarshal([]byte(tq.Options), &optArr)

				userText := reqAns.SelectedOptionID
				if strings.HasPrefix(userText, "OPT_") && len(optArr) > 0 {
					parts := strings.Split(userText, "_")
					if len(parts) >= 3 {
						idxStr := parts[len(parts)-1]
						if idx, err := strconv.Atoi(idxStr); err == nil && idx >= 0 && idx < len(optArr) {
							userText = optArr[idx]
						}
					}
				}

				normUser := NormalizeAnswer(userText)
				normCorrect := NormalizeAnswer(tq.Answer)
				ans.IsCorrect = (normUser == normCorrect)
				if ans.IsCorrect {
					ans.Score = 10
					totalScore += 10
				}
				ans.Explanation = tq.Explanation
				ans.EvaluatedBy = "system_opt"
			} else {
				// Subjective Training Question
				aiEval, err := services.EvaluateAnswer(tq.Prompt, tq.Answer, reqAns.WrittenAnswer, 10)
				if err != nil {
					ans.Score = 0
					ans.Feedback = "AI Interface Offline."
				} else {
					ans.Score = aiEval.Score
					ans.IsCorrect = aiEval.IsCorrect
					ans.Feedback = aiEval.Feedback
					ans.Explanation = aiEval.Explanation
				}
				totalScore += ans.Score
				ans.EvaluatedBy = "AI_TRAIN"
			}
			log.Printf("[AttemptSave] Training match: ID=%d correct=%v score=%d", tq.ID, ans.IsCorrect, ans.Score)
		}

		attemptAnswers = append(attemptAnswers, ans)
	}

	attempt := models.Attempt{
		ID:             attemptID,
		UserID:         userID.(string),
		QuizID:         req.QuizID,
		Score:          totalScore,
		TotalQuestions: len(req.Answers),
		StartedAt:      req.StartedAt,
		CompletedAt:    time.Now(),
	}

	// Begin transaction
	tx := database.DB.Begin()
	if err := tx.Create(&attempt).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save attempt"})
		return
	}
	if len(attemptAnswers) > 0 {
		if err := tx.Create(&attemptAnswers).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answers"})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":   "Attempt submitted successfully",
		"attemptId": attemptID,
		"score":     totalScore,
	})
}

// EvaluateSubjectiveAnswerMock acts as a mock AI grader.
func EvaluateSubjectiveAnswerMock(prompt string, correct string, actual string, maxScore int) (int, string) {
	// Simple simulation of AI grading.
	if actual == "" {
		return 0, "No answer provided."
	}
	if len(actual) < 10 {
		return 0, "Answer is too brief to be considered correct."
	}
	// Give partial/full credit just to simulate
	return maxScore, fmt.Sprintf("AI Evaluation: Your answer captures the essence of the topic nicely. Detailed and accurate compared to the reference: %s", correct)
}

func GetLeaderboard(c *gin.Context) {
	type LeaderboardEntry struct {
		UserID   string `json:"userId"`
		Username string `json:"username"`
		Score    int    `json:"score"`
	}

	var results []LeaderboardEntry
	database.DB.Table("attempts").
		Select("attempts.userId as user_id, user.username, SUM(attempts.score) as score").
		Joins("JOIN user on user.id = attempts.userId").
		Group("attempts.userId, user.username").
		Order("score desc").
		Limit(10).
		Scan(&results)

	c.JSON(http.StatusOK, results)
}

func GetAttemptResult(c *gin.Context) {
	attemptID := c.Param("id")

	var attempt models.Attempt
	if err := database.DB.Preload("User").Preload("Quiz").Where("id = ?", attemptID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	var answers []models.AttemptAnswer
	database.DB.Where("attemptId = ?", attemptID).Find(&answers)

	c.JSON(http.StatusOK, gin.H{
		"attempt": attempt,
		"answers": answers,
	})
}

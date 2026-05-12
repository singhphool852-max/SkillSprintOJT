package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ──────────────────────────────────────────────
// ListPublicTopics → GET /api/topics
// Returns all topics (for Arena topic listing).
// ──────────────────────────────────────────────
func ListPublicTopics(c *gin.Context) {
	var topics []models.Topic
	if err := database.DB.Order("name asc").Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	c.JSON(http.StatusOK, topics)
}

// ──────────────────────────────────────────────
// ListPublicTestsByTopic → GET /api/topics/:slug/tests
// Returns published tests for a given topic slug.
// Hidden test cases are stripped from the response.
// ──────────────────────────────────────────────
func ListPublicTestsByTopic(c *gin.Context) {
	slug := c.Param("slug")

	var topic models.Topic
	if err := database.DB.Where("slug = ?", slug).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	var tests []models.Test
	if err := database.DB.
		Where("topicId = ? AND isPublished = ?", topic.ID, true).
		Preload("Topic").
		Order("startTime desc").
		Find(&tests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topic": topic,
		"tests": tests,
	})
}

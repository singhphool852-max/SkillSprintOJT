package handlers

import (
	"backend/database"
	"backend/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ──────────────────────────────────────────────
// Request structs
// ──────────────────────────────────────────────

type CreateTopicRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTopicRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ──────────────────────────────────────────────
// CreateTopic → POST /api/admin/topics
// ──────────────────────────────────────────────
func CreateTopic(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic name is required"})
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.Name), " ", "-"))

	// Check for duplicate slug
	var existing models.Topic
	if err := database.DB.Where("slug = ?", slug).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A topic with this name already exists"})
		return
	}

	topic := models.Topic{
		ID:          uuid.New().String(),
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: req.Description,
		CreatedBy:   userID.(string),
	}

	if err := database.DB.Create(&topic).Error; err != nil {
		log.Printf("[TOPIC ERROR] %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topic)
}

// ──────────────────────────────────────────────
// ListTopics → GET /api/admin/topics
// ──────────────────────────────────────────────
func ListTopics(c *gin.Context) {
	var topics []models.Topic
	if err := database.DB.Preload("Creator").Order("createdAt desc").Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	c.JSON(http.StatusOK, topics)
}

// ──────────────────────────────────────────────
// UpdateTopic → PUT /api/admin/topics/:id
// ──────────────────────────────────────────────
func UpdateTopic(c *gin.Context) {
	topicID := c.Param("id")

	var topic models.Topic
	if err := database.DB.Where("id = ?", topicID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	var req UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Name != "" {
		topic.Name = strings.TrimSpace(req.Name)
		topic.Slug = strings.ToLower(strings.ReplaceAll(topic.Name, " ", "-"))
	}
	if req.Description != "" {
		topic.Description = req.Description
	}

	if err := database.DB.Save(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update topic"})
		return
	}

	c.JSON(http.StatusOK, topic)
}

// ──────────────────────────────────────────────
// DeleteTopic → DELETE /api/admin/topics/:id
// ──────────────────────────────────────────────
func DeleteTopic(c *gin.Context) {
	topicID := c.Param("id")

	var topic models.Topic
	if err := database.DB.Where("id = ?", topicID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	// Check if any tests reference this topic
	var testCount int64
	database.DB.Model(&models.Test{}).Where("topicId = ?", topicID).Count(&testCount)
	if testCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete topic with existing tests. Remove or reassign tests first."})
		return
	}

	if err := database.DB.Delete(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete topic"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Topic deleted", "topicId": topicID})
}

package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ipsitapp8/SkillSprintOJT/go-backend/chat"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/middleware"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ChatHub is set by main.go during startup.
var ChatHub *chat.Hub

var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ChatWebSocket handles WebSocket connections for the chat.
func ChatWebSocket(c *gin.Context) {
	// Try to get userID from JWT middleware first
	userID, exists := c.Get("userID")
	
	// If not found, try to extract from query parameter token
	if !exists {
		token := c.Query("token")
		if token == "" {
			log.Printf("[CHAT] No token provided in query or header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		
		// Validate token and extract userID
		claims, err := middleware.ValidateToken(token)
		if err != nil {
			log.Printf("[CHAT] Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		userID = claims.ID
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("[CHAT] User not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[CHAT] WebSocket upgrade failed: %v", err)
		return
	}

	// Check if ChatHub is initialized
	if ChatHub == nil {
		log.Printf("[CHAT] CRITICAL: ChatHub is nil, cannot serve WebSocket")
		conn.Close()
		return
	}

	log.Printf("[CHAT] WebSocket connection established for user: %s (%s)", user.Username, user.ID)
	ChatHub.ServeWS(conn, user.ID, user.Username, user.AvatarURL)
}

// UploadChatFile handles file uploads for chat (images and PDFs).
func UploadChatFile(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := c.Get("userID")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"application/pdf": true,
	}

	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only images (JPEG, PNG, GIF) and PDFs are allowed"})
		return
	}

	originalName := header.Filename
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, ext)
	baseName = sanitizeFilename(baseName)

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s_%s%s", timestamp, userID, baseName, ext)

	uploadDir := "uploads/chat"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[CHAT] Failed to create upload directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("[CHAT] Failed to create file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("[CHAT] Failed to copy file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	fileURL := fmt.Sprintf("/uploads/chat/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": originalName,
	})
}

// GetChatHistory returns the last 200 chat messages from the database.
func GetChatHistory(c *gin.Context) {
	var messages []models.ChatMessage

	if err := database.DB.Order("createdAt DESC").Limit(200).Find(&messages).Error; err != nil {
		log.Printf("[CHAT] Failed to fetch chat history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	// Reverse so oldest is first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	events := make([]chat.ChatEvent, len(messages))
	for i, msg := range messages {
		events[i] = chat.ChatEvent{
			Type:        "message",
			UserID:      msg.UserID,
			Username:    msg.Username,
			Avatar:      msg.Avatar,
			MessageType: msg.MessageType,
			Content:     msg.Content,
			FileName:    msg.FileName,
			Timestamp:   msg.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, events)
}

// sanitizeFilename removes special characters from filename.
func sanitizeFilename(filename string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := reg.ReplaceAllString(filename, "_")
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	return sanitized
}

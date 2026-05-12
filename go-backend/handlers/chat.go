package handlers

import (
	"backend/chat"
	"backend/database"
	"backend/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var ChatHub *chat.Hub

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" ||
			strings.HasSuffix(origin, ".vercel.app") ||
			strings.HasSuffix(origin, ".amplifyapp.com") ||
			origin == "https://skillsprintojt.onrender.com"
	},
}

// ChatWebSocket handles WebSocket connections for the chat
func ChatWebSocket(c *gin.Context) {
	// Get user info from JWT middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Fetch user details from database
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[CHAT] WebSocket upgrade failed: %v", err)
		return
	}

	// Serve the WebSocket connection
	ChatHub.ServeWS(conn, user.ID, user.Username, user.AvatarURL)
}

// UploadChatFile handles file uploads for chat (images and PDFs)
func UploadChatFile(c *gin.Context) {
	// Get user info from JWT middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Check file size (max 10MB)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
		return
	}

	// Check file type
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

	// Sanitize filename
	originalName := header.Filename
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, ext)
	baseName = sanitizeFilename(baseName)

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s_%s%s", timestamp, userID, baseName, ext)

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/chat"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[CHAT] Failed to create upload directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save file
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

	// Return file URL
	fileURL := fmt.Sprintf("/uploads/chat/%s", filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": originalName,
	})
}

// GetChatHistory returns the last 50 chat messages
func GetChatHistory(c *gin.Context) {
	var messages []models.ChatMessage

	// Fetch last 50 messages ordered by created_at DESC
	if err := database.DB.Order("created_at DESC").Limit(50).Find(&messages).Error; err != nil {
		log.Printf("[CHAT] Failed to fetch chat history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	// Reverse the order so oldest is first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// Convert to ChatEvent format
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

// sanitizeFilename removes special characters from filename
func sanitizeFilename(filename string) string {
	// Remove any character that's not alphanumeric, dash, or underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	sanitized := reg.ReplaceAllString(filename, "_")

	// Limit length to 50 characters
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

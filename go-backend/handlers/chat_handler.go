package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/chat"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ChatHub is set by main.go during startup.
var ChatHub *chat.Hub

var chatUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSocket
	},
}

// ChatWebSocket handles the /ws/chat WebSocket connection.
func ChatWebSocket(c *gin.Context) {
	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[CHAT] WebSocket upgrade failed: %v", err)
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("userName")

	userIDStr, _ := userID.(string)
	usernameStr, _ := username.(string)
	if usernameStr == "" {
		usernameStr = "Anonymous"
	}

	client := &chat.Client{
		Hub:      ChatHub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userIDStr,
		Username: usernameStr,
	}

	ChatHub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// UploadChatFile handles POST /api/chat/upload — uploads a file for chat sharing.
func UploadChatFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Create upload directory
	uploadDir := "./uploads/chat"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		return
	}

	fileURL := "/uploads/chat/" + filename
	log.Printf("[CHAT] File uploaded: %s -> %s", header.Filename, fileURL)

	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": header.Filename,
	})
}

// GetChatHistory returns the recent chat history (in-memory).
func GetChatHistory(c *gin.Context) {
	if ChatHub == nil {
		c.JSON(http.StatusOK, gin.H{
			"messages":  []chat.Message{},
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	messages := ChatHub.GetHistory()
	c.JSON(http.StatusOK, gin.H{
		"messages":  messages,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/arena"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/middleware"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// ArenaSessionHub is the global session hub for arena WS.
var ArenaSessionHub *arena.SessionHub

var arenaUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins during development
	},
}

// ArenaSessionWS handles WebSocket connections for arena test sessions.
// Route: /ws/arena/:attemptId
//
// The client connects with their attemptId. The server:
//  1. Validates the attempt exists and belongs to the user
//  2. Sends initial timer_sync event
//  3. Registers the connection for periodic timer broadcasts
//  4. Reads client pings/messages to keep connection alive
//
// Events sent TO client:
//   - timer_sync: {remainingSeconds, status, serverTime}
//   - auto_submit: {attemptId, score, message}
//   - session_ended: {attemptId, message}
//   - test_state: {testId, state}
//
// Events FROM client:
//   - ping: keep-alive
//   - save_answer: {questionId, type, data} → autosave via WS
func ArenaSessionWS(c *gin.Context) {
	attemptID := c.Param("attemptId")

	tokenString := ""
	cookie, err := c.Cookie("auth_token")
	if err == nil {
		tokenString = cookie
	}
	if tokenString == "" {
		reqToken := c.Query("token")
		if reqToken != "" {
			tokenString = reqToken
		}
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	claims := &middleware.SessionPayload{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JWT_SECRET, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}
	userID := claims.ID

	// Validate the attempt exists
	var attempt models.TestAttempt
	if err := database.DB.Preload("Test").Where("id = ?", attemptID).First(&attempt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attempt not found"})
		return
	}

	if attempt.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Attempt does not belong to user"})
		return
	}

	// Upgrade to WebSocket
	conn, err := arenaUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("arena ws: upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	if ArenaSessionHub == nil {
		log.Println("arena ws: hub not initialized")
		return
	}

	// Register this connection
	ArenaSessionHub.Register(attemptID, conn)
	defer ArenaSessionHub.Unregister(attemptID)

	log.Printf("arena ws: client connected for attempt %s", attemptID)

	// Send initial timer sync immediately
	now := time.Now()
	elapsed := now.Sub(attempt.Test.StartTime)
	remaining := attempt.Test.DurationSeconds - int(elapsed.Seconds())

	status := "live"
	if now.Before(attempt.Test.StartTime) {
		status = "upcoming"
		remaining = attempt.Test.DurationSeconds
	} else if remaining <= 0 {
		status = "ended"
		remaining = 0
	}

	// Check if already submitted
	if attempt.SubmittedAt != nil {
		ArenaSessionHub.SendEvent(attemptID, "session_ended", map[string]interface{}{
			"attemptId": attemptID,
			"message":   "Test already submitted",
		})
	} else {
		ArenaSessionHub.SendEvent(attemptID, "timer_sync", arena.TimerSyncData{
			RemainingSeconds: remaining,
			Status:           status,
			ServerTime:       now.Format(time.RFC3339),
		})
	}

	// Read loop — keep connection alive, handle client messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("arena ws: unexpected close for attempt %s: %v", attemptID, err)
			}
			break
		}

		// Handle ping/pong or client messages
		if string(msg) == "ping" {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		}
		// Future: handle save_answer messages via WS if needed
	}

	log.Printf("arena ws: client disconnected for attempt %s", attemptID)
}

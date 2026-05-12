package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/leaderboard"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// LeaderboardHub is set from main.go at startup.
var LeaderboardHub *leaderboard.Hub

// ──────────────────────────────────────────────
// GetTestLeaderboard → GET /api/leaderboard/:testId
// Returns: [{ rank, username, score, solvedCount }]
// Order: score DESC, then submittedAt ASC for tiebreak.
// ──────────────────────────────────────────────
func GetTestLeaderboard(c *gin.Context) {
	testID := c.Param("testId")
	entries := LeaderboardHub.GetLeaderboard(testID)
	c.JSON(http.StatusOK, entries)
}

// WebSocket upgrader — allow connections from frontend origin.
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS is handled at Gin level
	},
}

// ──────────────────────────────────────────────
// LeaderboardWS → WS /ws/leaderboard/:testId
// On connect: send current leaderboard immediately.
// Stay connected for broadcasts until client disconnects.
// ──────────────────────────────────────────────
func LeaderboardWS(c *gin.Context) {
	testID := c.Param("testId")

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func() {
		LeaderboardHub.Unregister(testID, conn)
		conn.Close()
	}()

	// Register client
	LeaderboardHub.Register(testID, conn)

	// Send current leaderboard immediately on connect
	LeaderboardHub.Broadcast(testID)

	// Keep connection alive — read loop handles disconnect detection.
	// We don't expect client messages, but must read to detect close.
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break // client disconnected
		}
	}
}

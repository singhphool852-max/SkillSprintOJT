package leaderboard

import (
	"backend/database"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// LeaderboardEntry is the payload sent to clients. Only username and score — no emails or sensitive data.
type LeaderboardEntry struct {
	Rank           int    `json:"rank"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	Score          int    `json:"score"`
	TotalQuestions int    `json:"total_questions"`
	TimeTaken      int    `json:"time_taken"`
}

// Hub manages WebSocket clients grouped by test_id.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // testID → set of connections
}

// NewHub creates a new leaderboard hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

// Register adds a client connection for a specific test.
func (h *Hub) Register(testID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[testID] == nil {
		h.clients[testID] = make(map[*websocket.Conn]bool)
	}
	h.clients[testID][conn] = true
}

// Unregister removes a client connection. Safe to call multiple times.
func (h *Hub) Unregister(testID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[testID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, testID)
		}
	}
}

// Broadcast queries the current leaderboard for a test and sends it to all connected clients.
func (h *Hub) Broadcast(testID string) {
	entries := queryLeaderboard(testID)

	data, err := json.Marshal(entries)
	if err != nil {
		log.Println("leaderboard: failed to marshal broadcast:", err)
		return
	}

	h.mu.RLock()
	conns := h.clients[testID]
	// Copy slice under read lock to avoid holding lock during writes
	targets := make([]*websocket.Conn, 0, len(conns))
	for conn := range conns {
		targets = append(targets, conn)
	}
	h.mu.RUnlock()

	for _, conn := range targets {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			// Client disconnected — remove without crashing
			log.Println("leaderboard: removing disconnected client")
			h.Unregister(testID, conn)
			conn.Close()
		}
	}
}

// GetLeaderboard returns the current leaderboard for a test (used by REST endpoint too).
func (h *Hub) GetLeaderboard(testID string) []LeaderboardEntry {
	return queryLeaderboard(testID)
}

// queryLeaderboard fetches the ranked leaderboard from the database.
// Order: score DESC, timeTaken ASC, submittedAt ASC.
// Only includes submitted attempts (submittedAt IS NOT NULL and not zero).
func queryLeaderboard(testID string) []LeaderboardEntry {
	type row struct {
		UserID         string `json:"userId"`
		Username       string `json:"username"`
		Score          int    `json:"score"`
		TotalQuestions int    `json:"totalQuestions"`
		TimeTaken      int    `json:"timeTaken"`
	}

	var rows []row
	database.DB.Table("test_attempts").
		Select("test_attempts.userId as user_id, user.username, test_attempts.score, test_attempts.totalQuestions as total_questions, test_attempts.timeTaken as time_taken").
		Joins("JOIN user ON user.id = test_attempts.userId").
		Where("test_attempts.testId = ? AND test_attempts.submittedAt IS NOT NULL AND test_attempts.submittedAt != ''", testID).
		Order("test_attempts.score DESC, test_attempts.timeTaken ASC, test_attempts.submittedAt ASC").
		Scan(&rows)

	entries := make([]LeaderboardEntry, len(rows))
	for i, r := range rows {
		entries[i] = LeaderboardEntry{
			Rank:           i + 1,
			UserID:         r.UserID,
			Username:       r.Username,
			Score:          r.Score,
			TotalQuestions: r.TotalQuestions,
			TimeTaken:      r.TimeTaken,
		}
	}
	return entries
}

package arena

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ──────────────────────────────────────────────
// SessionHub manages per-user WebSocket connections
// for real-time arena session updates: timer sync,
// auto-submit events, test state changes.
// ──────────────────────────────────────────────

// SessionEvent is the envelope for all WS messages.
type SessionEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// TimerSyncData is sent periodically to keep the
// client's timer in sync with the backend clock.
type TimerSyncData struct {
	RemainingSeconds int    `json:"remainingSeconds"`
	Status           string `json:"status"` // "live", "ended", "upcoming"
	ServerTime       string `json:"serverTime"`
}

// SessionHub manages all active arena WebSocket clients.
type SessionHub struct {
	mu      sync.RWMutex
	clients map[string]*websocket.Conn // attemptID → connection
}

// NewSessionHub creates a new arena session hub.
func NewSessionHub() *SessionHub {
	hub := &SessionHub{
		clients: make(map[string]*websocket.Conn),
	}

	// Start background timer broadcaster
	go hub.timerLoop()

	return hub
}

// Register adds a client connection for an attempt.
func (h *SessionHub) Register(attemptID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close existing connection if any (e.g. reconnect)
	if old, ok := h.clients[attemptID]; ok {
		old.Close()
	}
	h.clients[attemptID] = conn
}

// Unregister removes a client connection.
func (h *SessionHub) Unregister(attemptID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, attemptID)
}

// SendEvent sends a typed event to a specific attempt's WS client.
func (h *SessionHub) SendEvent(attemptID string, eventType string, data interface{}) {
	h.mu.RLock()
	conn, ok := h.clients[attemptID]
	h.mu.RUnlock()

	if !ok || conn == nil {
		return
	}

	event := SessionEvent{Type: eventType, Data: data}
	msg, err := json.Marshal(event)
	if err != nil {
		log.Printf("arena ws: marshal error: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("arena ws: write error for attempt %s, removing", attemptID)
		h.Unregister(attemptID)
		conn.Close()
	}
}

// BroadcastAutoSubmit notifies a user that their attempt was auto-submitted.
func (h *SessionHub) BroadcastAutoSubmit(attemptID string, score int) {
	h.SendEvent(attemptID, "auto_submit", map[string]interface{}{
		"attemptId": attemptID,
		"score":     score,
		"message":   "Time expired. Your test has been auto-submitted.",
	})
}

// BroadcastTestStateChange notifies all connected clients about a test state change
// (e.g. admin deactivated a test mid-session).
func (h *SessionHub) BroadcastTestStateChange(testID string, state string) {
	h.mu.RLock()
	// Collect all attempt IDs (we need to look up which attempts belong to this test)
	h.mu.RUnlock()

	var attempts []models.TestAttempt
	database.DB.Where("testId = ? AND (submittedAt IS NULL OR submittedAt = '')", testID).Find(&attempts)

	for _, a := range attempts {
		h.SendEvent(a.ID, "test_state", map[string]interface{}{
			"testId": testID,
			"state":  state,
		})
	}
}

// timerLoop broadcasts timer_sync events every 5 seconds to all connected clients.
func (h *SessionHub) timerLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		attempts := make([]string, 0, len(h.clients))
		for attemptID := range h.clients {
			attempts = append(attempts, attemptID)
		}
		h.mu.RUnlock()

		if len(attempts) == 0 {
			continue
		}

		// Batch-fetch attempt data
		var attemptRows []models.TestAttempt
		database.DB.Preload("Test").Where("id IN ?", attempts).Find(&attemptRows)

		now := time.Now()
		for _, attempt := range attemptRows {
			// Skip already-submitted attempts
			if attempt.SubmittedAt != nil {
				h.SendEvent(attempt.ID, "session_ended", map[string]interface{}{
					"attemptId": attempt.ID,
					"message":   "Test already submitted",
				})
				continue
			}

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

			h.SendEvent(attempt.ID, "timer_sync", TimerSyncData{
				RemainingSeconds: remaining,
				Status:           status,
				ServerTime:       now.Format(time.RFC3339),
			})
		}
	}
}

package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a chat message exchanged between users.
type Message struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Content   string `json:"content"`
	Type      string `json:"type"` // "text", "file", "system"
	FileURL   string `json:"fileUrl,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Client represents a single WebSocket chat connection.
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   string
	Username string
}

// Hub manages all connected chat clients and message broadcasting.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	history    []Message // In-memory recent message history
}

// NewHub creates a new chat hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		history:    make([]Message, 0),
	}
}

// Run starts the hub's event loop. Should be called as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[CHAT] Client connected: %s (%s)", client.Username, client.UserID)

			// Send recent history to newly connected client
			h.mu.RLock()
			for _, msg := range h.history {
				data, _ := json.Marshal(msg)
				select {
				case client.Send <- data:
				default:
				}
			}
			h.mu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("[CHAT] Client disconnected: %s", client.Username)

		case message := <-h.broadcast:
			// Store in history (keep last 100 messages)
			var msg Message
			if err := json.Unmarshal(message, &msg); err == nil {
				h.mu.Lock()
				h.history = append(h.history, msg)
				if len(h.history) > 100 {
					h.history = h.history[len(h.history)-100:]
				}
				h.mu.Unlock()
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// GetHistory returns the recent chat history.
func (h *Hub) GetHistory() []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]Message, len(h.history))
	copy(result, h.history)
	return result
}

// ReadPump reads messages from the WebSocket connection and broadcasts them.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse incoming, enrich with server timestamp
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		msg.UserID = c.UserID
		msg.Username = c.Username
		if msg.Timestamp == "" {
			msg.Timestamp = time.Now().Format(time.RFC3339)
		}

		enriched, _ := json.Marshal(msg)
		c.Hub.Broadcast(enriched)
	}
}

// WritePump sends messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

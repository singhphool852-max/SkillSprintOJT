package chat

import (
 main

	"backend/database"
	"backend/models"
 main
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

main
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

// Client represents a connected WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	username string
	avatar   string
}

// Hub manages all chat clients and message broadcasting
 main
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
 main
	history    []Message // In-memory recent message history
}

// NewHub creates a new chat hub.

}

// ChatEvent represents a chat message or system event
type ChatEvent struct {
	Type        string `json:"type"`         // "message", "user_joined", "user_left", "online_count"
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	MessageType string `json:"message_type"` // "text", "note", "image", "pdf"
	Content     string `json:"content"`
	FileName    string `json:"file_name,omitempty"`
	Timestamp   string `json:"timestamp"`
	OnlineCount int    `json:"online_count,omitempty"`
}

// NewHub creates a new chat hub
main
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
main
		history:    make([]Message, 0),
	}
}

// Run starts the hub's event loop. Should be called as a goroutine.
	}
}

// Run starts the hub's main event loop
main
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
main
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

			count := len(h.clients)
			h.mu.Unlock()

			// Broadcast user joined event
			joinEvent := ChatEvent{
				Type:        "user_joined",
				UserID:      client.userID,
				Username:    client.username,
				Avatar:      client.avatar,
				Timestamp:   time.Now().Format(time.RFC3339),
				OnlineCount: count,
			}
			h.broadcastEvent(joinEvent)

			log.Printf("[CHAT] User joined: %s (%s), online: %d", client.username, client.userID, count)
main

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
 main
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

				close(client.send)
			}
			count := len(h.clients)
			h.mu.Unlock()

			// Broadcast user left event
			leftEvent := ChatEvent{
				Type:        "user_left",
				UserID:      client.userID,
				Username:    client.username,
				Timestamp:   time.Now().Format(time.RFC3339),
				OnlineCount: count,
			}
			h.broadcastEvent(leftEvent)

			log.Printf("[CHAT] User left: %s (%s), online: %d", client.username, client.userID, count)

		case message := <-h.broadcast:
			var event ChatEvent
			if err := json.Unmarshal(message, &event); err != nil {
				log.Printf("[CHAT] Failed to unmarshal broadcast message: %v", err)
				continue
			}

			// Save message to database if it's a chat message
			if event.Type == "message" {
				chatMsg := models.ChatMessage{
					UserID:      event.UserID,
					Username:    event.Username,
					Avatar:      event.Avatar,
					MessageType: event.MessageType,
					Content:     event.Content,
					FileName:    event.FileName,
					CreatedAt:   time.Now(),
				}
				if err := database.DB.Create(&chatMsg).Error; err != nil {
					log.Printf("[CHAT] Failed to save message to DB: %v", err)
				}
			}

			// Broadcast to all connected clients
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, disconnect them
					close(client.send)
 main
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

 main
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

// broadcastEvent sends an event to all connected clients
func (h *Hub) broadcastEvent(event ChatEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[CHAT] Failed to marshal event: %v", err)
		return
	}
	h.broadcast <- data
}

// GetOnlineCount returns the current number of connected clients
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
 main
		return nil
	})

	for {
main
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

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[CHAT] WebSocket error: %v", err)
			}
			break
		}

		var event ChatEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("[CHAT] Failed to unmarshal client message: %v", err)
			continue
		}

		// Ignore ping messages
		if event.Type == "ping" {
			continue
		}

		// Set user info and timestamp
		event.UserID = c.userID
		event.Username = c.username
		event.Avatar = c.avatar
		event.Timestamp = time.Now().Format(time.RFC3339)

		// Broadcast the message
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("[CHAT] Failed to marshal event: %v", err)
			continue
		}
		c.hub.broadcast <- data
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
main
	}()

	for {
		select {
main
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
main
				return
			}

		case <-ticker.C:
 main
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
 main
				return
			}
		}
	}
}
 main


// ServeWS handles WebSocket requests from clients
func (h *Hub) ServeWS(conn *websocket.Conn, userID, username, avatar string) {
	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
		avatar:   avatar,
	}

	h.register <- client

	// Start read and write pumps in separate goroutines
	go client.writePump()
	go client.readPump()
}
 main

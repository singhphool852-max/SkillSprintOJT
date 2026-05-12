package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ChatEvent represents all types of chat events (messages, joins, leaves)
type ChatEvent struct {
	Type        string `json:"type"`                   // "message", "user_joined", "user_left", "ping"
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar,omitempty"`
	MessageType string `json:"messageType,omitempty"`   // "text", "note", "image", "pdf"
	Content     string `json:"content,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	Timestamp   string `json:"timestamp"`
	OnlineCount int    `json:"online_count,omitempty"`
}

// Message represents a simple chat message (kept for backward compatibility).
type Message struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Content   string `json:"content"`
	Type      string `json:"type"`
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
	Avatar   string
}

// Hub manages all connected chat clients and message broadcasting.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	history    []Message
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
			count := len(h.clients)
			h.mu.Unlock()

			// Broadcast user joined event
			joinEvent := ChatEvent{
				Type:        "user_joined",
				UserID:      client.UserID,
				Username:    client.Username,
				Avatar:      client.Avatar,
				Timestamp:   time.Now().Format(time.RFC3339),
				OnlineCount: count,
			}
			h.broadcastEvent(joinEvent)

			log.Printf("[CHAT] User joined: %s (%s), online: %d", client.Username, client.UserID, count)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			count := len(h.clients)
			h.mu.Unlock()

			// Broadcast user left event
			leftEvent := ChatEvent{
				Type:        "user_left",
				UserID:      client.UserID,
				Username:    client.Username,
				Timestamp:   time.Now().Format(time.RFC3339),
				OnlineCount: count,
			}
			h.broadcastEvent(leftEvent)

			log.Printf("[CHAT] User left: %s (%s), online: %d", client.Username, client.UserID, count)

		case message := <-h.broadcast:
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

// broadcastEvent sends an event to all connected clients.
func (h *Hub) broadcastEvent(event ChatEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[CHAT] Failed to marshal event: %v", err)
		return
	}
	h.broadcast <- data
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

// GetOnlineCount returns the current number of connected clients.
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ServeWS handles WebSocket requests from clients.
func (h *Hub) ServeWS(conn *websocket.Conn, userID, username, avatar string) {
	client := &Client{
		Hub:      h,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
		Avatar:   avatar,
	}

	h.register <- client

	go client.WritePump()
	go client.ReadPump()
}

// ReadPump reads messages from the WebSocket connection and broadcasts them.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
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
		event.UserID = c.UserID
		event.Username = c.Username
		event.Avatar = c.Avatar
		event.Timestamp = time.Now().Format(time.RFC3339)

		// Broadcast the message
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("[CHAT] Failed to marshal event: %v", err)
			continue
		}
		c.Hub.Broadcast(data)
	}
}

// WritePump sends messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
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

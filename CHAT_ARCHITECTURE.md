# SkillSprint Chat - Architecture Overview

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND (Next.js)                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              /app/chat/page.tsx                           │  │
│  │  ┌────────────────────────────────────────────────────┐  │  │
│  │  │  • Chat UI (messages, input, buttons)              │  │  │
│  │  │  • Message rendering (text, note, image, pdf)      │  │  │
│  │  │  • File upload modal                               │  │  │
│  │  │  • Note composition modal                          │  │  │
│  │  └────────────────────────────────────────────────────┘  │  │
│  │                          ↕                                │  │
│  │  ┌────────────────────────────────────────────────────┐  │  │
│  │  │         /hooks/useChat.ts                          │  │  │
│  │  │  • WebSocket connection management                 │  │  │
│  │  │  • Message state management                        │  │  │
│  │  │  • Auto-reconnect logic                            │  │  │
│  │  │  • File upload handler                             │  │  │
│  │  └────────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           /components/nav.tsx                             │  │
│  │  • Chat link with 💬 icon                                │  │
│  │  • Online count badge (future)                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ WebSocket (ws://)
                            │ HTTP REST API
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                      BACKEND (Go/Gin)                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    main.go                                │  │
│  │  • Initialize ChatHub                                     │  │
│  │  • Start hub.Run() goroutine                             │  │
│  │  • Register routes                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          ↕                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              /chat/hub.go                                 │  │
│  │  ┌────────────────────────────────────────────────────┐  │  │
│  │  │  Hub (manages all clients)                         │  │  │
│  │  │  • clients map[*Client]bool                        │  │  │
│  │  │  • broadcast chan []byte                           │  │  │
│  │  │  • register/unregister channels                    │  │  │
│  │  └────────────────────────────────────────────────────┘  │  │
│  │  ┌────────────────────────────────────────────────────┐  │  │
│  │  │  Client (one per WebSocket connection)             │  │  │
│  │  │  • conn *websocket.Conn                            │  │  │
│  │  │  • send chan []byte                                │  │  │
│  │  │  • userID, username, avatar                        │  │  │
│  │  │  • readPump() / writePump() goroutines             │  │  │
│  │  └────────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          ↕                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           /handlers/chat.go                               │  │
│  │  • ChatWebSocket() - Upgrade to WebSocket                │  │
│  │  • UploadChatFile() - Handle file uploads                │  │
│  │  • GetChatHistory() - Return last 50 messages            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          ↕                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            /models/chat.go                                │  │
│  │  • ChatMessage struct                                     │  │
│  │  • GORM model definition                                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          ↕                                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ GORM ORM
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                      DATABASE (MySQL)                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              chat_messages table                          │  │
│  │  • id (PRIMARY KEY)                                       │  │
│  │  • userId (VARCHAR, INDEXED)                              │  │
│  │  • username (VARCHAR)                                     │  │
│  │  • avatar (VARCHAR)                                       │  │
│  │  • messageType (VARCHAR) - text/note/image/pdf           │  │
│  │  • content (TEXT) - message or file URL                  │  │
│  │  • fileName (VARCHAR) - original filename                │  │
│  │  • createdAt (TIMESTAMP, INDEXED)                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   FILE STORAGE (Local Disk)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  go-backend/uploads/chat/                                        │
│  • {timestamp}_{userId}_{filename}.jpg                           │
│  • {timestamp}_{userId}_{filename}.png                           │
│  • {timestamp}_{userId}_{filename}.pdf                           │
│                                                                   │
│  Served via: GET /uploads/chat/:filename                         │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

### 1. User Sends Text Message

```
User types message
      ↓
Frontend: useChat.sendMessage()
      ↓
WebSocket: Send JSON event
      ↓
Backend: Client.readPump() receives
      ↓
Backend: Hub.broadcast channel
      ↓
Backend: Save to database
      ↓
Backend: Broadcast to all clients
      ↓
Frontend: All users receive message
      ↓
UI: Message appears in chat
```

### 2. User Uploads File

```
User selects file
      ↓
Frontend: useChat.sendFile()
      ↓
HTTP POST: /api/chat/upload
      ↓
Backend: Validate file (type, size)
      ↓
Backend: Sanitize filename
      ↓
Backend: Save to uploads/chat/
      ↓
Backend: Return file URL
      ↓
Frontend: Send message via WebSocket
      ↓
Backend: Broadcast file message
      ↓
Frontend: All users see file
```

### 3. User Joins Chat

```
User navigates to /chat
      ↓
Frontend: Load chat history (HTTP GET)
      ↓
Backend: Query last 50 messages
      ↓
Frontend: Display messages
      ↓
Frontend: Connect WebSocket
      ↓
Backend: Upgrade to WebSocket
      ↓
Backend: Register client in Hub
      ↓
Backend: Broadcast "user_joined" event
      ↓
Frontend: All users see updated online count
```

## WebSocket Event Types

### Client → Server

```json
{
  "type": "message",
  "message_type": "text|note|image|pdf",
  "content": "message text or file URL",
  "file_name": "optional filename",
  "timestamp": "ISO 8601 timestamp"
}
```

### Server → Client

#### Message Event
```json
{
  "type": "message",
  "user_id": "uuid",
  "username": "john_doe",
  "avatar": "https://...",
  "message_type": "text|note|image|pdf",
  "content": "message or URL",
  "file_name": "optional",
  "timestamp": "ISO 8601"
}
```

#### User Joined Event
```json
{
  "type": "user_joined",
  "user_id": "uuid",
  "username": "john_doe",
  "avatar": "https://...",
  "timestamp": "ISO 8601",
  "online_count": 5
}
```

#### User Left Event
```json
{
  "type": "user_left",
  "user_id": "uuid",
  "username": "john_doe",
  "timestamp": "ISO 8601",
  "online_count": 4
}
```

## Concurrency Model

### Backend Hub Goroutines

```
Main Hub Goroutine (hub.Run())
├─ Listens on register channel
├─ Listens on unregister channel
└─ Listens on broadcast channel

Per-Client Goroutines (2 per connection)
├─ readPump() - Reads from WebSocket
└─ writePump() - Writes to WebSocket
```

### Thread Safety

- Hub uses `sync.RWMutex` for clients map
- Channels used for communication between goroutines
- No shared mutable state between clients

## Security Layers

```
┌─────────────────────────────────────────┐
│  1. JWT Authentication                  │
│     • Validate token on WebSocket       │
│     • Validate token on file upload     │
│     • Validate token on history fetch   │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  2. File Validation                     │
│     • Check file type (whitelist)       │
│     • Check file size (max 10MB)        │
│     • Sanitize filename                 │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  3. CORS Protection                     │
│     • Validate WebSocket origin         │
│     • Validate HTTP origin              │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  4. Data Sanitization                   │
│     • Remove special chars from files   │
│     • Validate JSON structure           │
└─────────────────────────────────────────┘
```

## Scalability Considerations

### Current Implementation (Single Server)
- All clients connect to one Hub instance
- In-memory client map
- Local file storage
- Single database connection pool

### Future Scaling Options

#### Horizontal Scaling
```
Load Balancer
    ↓
┌───────┬───────┬───────┐
│ App 1 │ App 2 │ App 3 │
└───┬───┴───┬───┴───┬───┘
    └───────┴───────┘
          ↓
    Redis Pub/Sub
    (for message broadcast)
```

#### File Storage Scaling
```
Local Disk → Cloud Storage (S3/GCS)
           → CDN for delivery
```

#### Database Scaling
```
Single MySQL → Read Replicas
            → Sharding by date
            → Archive old messages
```

## Performance Characteristics

### WebSocket Connection
- Ping/Pong every 54 seconds
- Read deadline: 60 seconds
- Write deadline: 10 seconds
- Send buffer: 256 messages

### Database Queries
- History fetch: Last 50 messages (LIMIT 50)
- Message insert: Single row insert
- No joins required (denormalized)

### File Upload
- Synchronous processing
- Max 10MB per file
- No compression
- No thumbnail generation

## Monitoring Points

### Key Metrics to Track
1. **Active WebSocket connections** - Hub.GetOnlineCount()
2. **Messages per second** - Rate of broadcast channel
3. **File upload rate** - POST /api/chat/upload
4. **Database query time** - GORM logging
5. **WebSocket errors** - Connection failures
6. **File storage usage** - Disk space in uploads/

### Health Checks
- WebSocket endpoint availability
- Database connectivity
- File system write permissions
- Memory usage (client map size)

## Error Handling

### WebSocket Errors
- Connection closed → Auto-reconnect after 3s
- Write error → Remove client from hub
- Read error → Close connection gracefully

### File Upload Errors
- Invalid type → Return 400 error
- Too large → Return 400 error
- Save failed → Return 500 error
- All errors shown to user

### Database Errors
- Insert failed → Log error, continue
- Query failed → Return empty array
- Connection lost → Retry with backoff

---

This architecture provides a solid foundation for real-time chat while maintaining simplicity and following Go/React best practices.

# SkillSprint - Complete Architecture Guide for Beginners

## Table of Contents
1. [What is SkillSprint?](#what-is-skillsprint)
2. [High-Level Overview](#high-level-overview)
3. [Technology Stack](#technology-stack)
4. [System Architecture](#system-architecture)
5. [Frontend Architecture](#frontend-architecture)
6. [Backend Architecture](#backend-architecture)
7. [Database Design](#database-design)
8. [Key Features & Flows](#key-features--flows)
9. [Security Architecture](#security-architecture)
10. [Deployment Architecture](#deployment-architecture)
11. [Interview Questions](#interview-questions)

---

## 1. What is SkillSprint?

SkillSprint is a **competitive learning platform** where users can:
- Take coding tests and MCQ quizzes
- Compete in live arenas with other users
- Train with AI-generated questions
- Track their progress and rankings
- Chat with other learners in real-time

Think of it as **LeetCode + Kahoot + Discord** combined into one platform.

---

## 2. High-Level Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER DEVICES                             │
│              (Web Browser - Desktop/Mobile)                      │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTPS
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      FRONTEND (Next.js)                          │
│  • React Components                                              │
│  • Client-side routing                                           │
│  • State management                                              │
│  • WebSocket connections                                         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ REST API + WebSocket
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                    BACKEND (Go + Gin)                            │
│  • REST API endpoints                                            │
│  • WebSocket hubs (Chat, Arena, Leaderboard)                    │
│  • Business logic                                                │
│  • Authentication & Authorization                                │
│  • Code execution (Judge system)                                 │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ SQL Queries
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                     DATABASE (MySQL)                             │
│  • Users, Tests, Questions                                       │
│  • Attempts, Results                                             │
│  • Chat messages                                                 │
│  • Leaderboards                                                  │
└─────────────────────────────────────────────────────────────────┘
```


### Request Flow Example

**User takes a test:**
```
1. User clicks "Start Test" → Frontend sends POST /api/arena/tests/:id/join
2. Backend creates TestAttempt record → Returns attempt_id
3. Frontend displays questions → User answers
4. User submits code → Frontend sends POST /api/submissions/run
5. Backend executes code in sandbox → Returns output
6. User submits test → Frontend sends POST /api/attempts/:id/submit
7. Backend calculates score → Updates database
8. Backend broadcasts to WebSocket → Leaderboard updates in real-time
9. Frontend shows results page
```

---

## 3. Technology Stack

### Frontend Stack

| Technology | Purpose | Why? |
|-----------|---------|------|
| **Next.js 14** | React framework | Server-side rendering, routing, API routes |
| **TypeScript** | Type safety | Catch errors at compile time |
| **Tailwind CSS** | Styling | Utility-first CSS, fast development |
| **Radix UI** | Component library | Accessible, unstyled components |
| **React Hook Form** | Form handling | Performance, validation |
| **Zod** | Schema validation | Type-safe validation |
| **Lucide React** | Icons | Modern icon library |
| **WebSocket API** | Real-time communication | Chat, live updates |

### Backend Stack

| Technology | Purpose | Why? |
|-----------|---------|------|
| **Go (Golang)** | Programming language | Fast, concurrent, compiled |
| **Gin** | Web framework | Fast HTTP router, middleware support |
| **GORM** | ORM | Database abstraction, migrations |
| **MySQL** | Database | Relational data, ACID compliance |
| **JWT** | Authentication | Stateless auth tokens |
| **bcrypt** | Password hashing | Secure password storage |
| **Gorilla WebSocket** | WebSocket library | Real-time bidirectional communication |
| **Docker** | Code execution sandbox | Isolated code execution |

### External Services

| Service | Purpose |
|---------|---------|
| **Google OAuth** | Social login |
| **OpenAI API** | AI test generation |
| **AWS Amplify** | Frontend hosting |
| **Render** | Backend hosting |

---

## 4. System Architecture

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │   Browser    │  │    Mobile    │  │    Tablet    │                  │
│  │   (Chrome)   │  │   (Safari)   │  │   (Firefox)  │                  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                  │
│         │                  │                  │                          │
│         └──────────────────┴──────────────────┘                          │
│                            │                                             │
└────────────────────────────┼─────────────────────────────────────────────┘
                             │
                             │ HTTPS (443)
                             │
┌────────────────────────────▼─────────────────────────────────────────────┐
│                         PRESENTATION LAYER                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Next.js Application                           │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Pages (Routes)                                           │  │   │
│  │  │  • / (Landing)                                            │  │   │
│  │  │  • /login, /register                                      │  │   │
│  │  │  • /dashboard                                             │  │   │
│  │  │  • /arena (Live tests)                                    │  │   │
│  │  │  • /train (Practice)                                      │  │   │
│  │  │  • /chat (Community)                                      │  │   │
│  │  │  • /leaderboard                                           │  │   │
│  │  │  • /admin (Admin panel)                                   │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Components                                               │  │   │
│  │  │  • UI components (buttons, cards, modals)                │  │   │
│  │  │  • Feature components (arena, train, chat)               │  │   │
│  │  │  • Layout components (nav, shell)                        │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Hooks (Custom React Hooks)                              │  │   │
│  │  │  • useChat - WebSocket chat                              │  │   │
│  │  │  • useAntiCheat - Proctoring                             │  │   │
│  │  │  • useAuth - Authentication                              │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
└────────────────────────────┬──────────────────────────────────────────────┘
                             │
                             │ REST API (HTTP) + WebSocket (WS)
                             │
┌────────────────────────────▼─────────────────────────────────────────────┐
│                         APPLICATION LAYER                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Go Backend (Gin Framework)                    │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Middleware Layer                                         │  │   │
│  │  │  • CORS (Cross-Origin Resource Sharing)                  │  │   │
│  │  │  • JWT Authentication                                     │  │   │
│  │  │  • Admin Authorization                                    │  │   │
│  │  │  • Rate Limiting                                          │  │   │
│  │  │  • Logging                                                │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Handlers (Controllers)                                   │  │   │
│  │  │  • auth.go - Login, signup, Google OAuth                 │  │   │
│  │  │  • arena.go - Test management                            │  │   │
│  │  │  • attempt.go - Submission handling                      │  │   │
│  │  │  • chat.go - Chat WebSocket                              │  │   │
│  │  │  • admin.go - Admin operations                           │  │   │
│  │  │  • training.go - Practice sessions                       │  │   │
│  │  │  • dashboard.go - User stats                             │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Services (Business Logic)                                │  │   │
│  │  │  • ai_service.go - OpenAI integration                    │  │   │
│  │  │  • judge/executor.go - Code execution                    │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  WebSocket Hubs (Real-time)                              │  │   │
│  │  │  • chat/hub.go - Global chat                             │  │   │
│  │  │  • leaderboard/hub.go - Live rankings                    │  │   │
│  │  │  • arena/session_hub.go - Test sessions                  │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
└────────────────────────────┬──────────────────────────────────────────────┘
                             │
                             │ GORM (ORM)
                             │
┌────────────────────────────▼─────────────────────────────────────────────┐
│                           DATA LAYER                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    MySQL Database                                │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Core Tables                                              │  │   │
│  │  │  • users - User accounts                                  │  │   │
│  │  │  • topics - Subject categories                            │  │   │
│  │  │  • tests - Test definitions                               │  │   │
│  │  │  • test_questions - Questions                             │  │   │
│  │  │  • test_mcq_options - MCQ choices                         │  │   │
│  │  │  • test_cases - Coding test cases                         │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Activity Tables                                          │  │   │
│  │  │  • test_attempts - User test sessions                     │  │   │
│  │  │  • test_submissions - Code/MCQ submissions                │  │   │
│  │  │  • test_results - Final scores                            │  │   │
│  │  │  • test_violations - Anti-cheat logs                      │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Analytics Tables                                         │  │   │
│  │  │  • user_wrong_questions - Mistake tracking                │  │   │
│  │  │  • user_topic_stats - Performance by topic               │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │  Communication Tables                                     │  │   │
│  │  │  • chat_messages - Global chat history                    │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  │                                                                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```


---

## 5. Frontend Architecture

### Directory Structure

```
frontend/
├── app/                          # Next.js 14 App Router
│   ├── layout.tsx               # Root layout (theme, fonts)
│   ├── page.tsx                 # Landing page
│   ├── globals.css              # Global styles
│   │
│   ├── login/                   # Authentication pages
│   │   └── page.tsx
│   ├── register/
│   │   └── page.tsx
│   │
│   ├── dashboard/               # User dashboard
│   │   └── page.tsx
│   │
│   ├── arena/                   # Live competitive tests
│   │   ├── page.tsx            # Arena lobby
│   │   ├── live/
│   │   │   └── page.tsx        # Active tests list
│   │   └── [id]/
│   │       └── play/
│   │           └── page.tsx    # Test taking interface
│   │
│   ├── train/                   # Practice mode
│   │   ├── page.tsx            # Training modes
│   │   ├── session/
│   │   │   └── page.tsx        # Session setup
│   │   └── play/
│   │       └── [id]/
│   │           └── page.tsx    # Practice interface
│   │
│   ├── chat/                    # Community chat
│   │   └── page.tsx
│   │
│   ├── leaderboard/             # Rankings
│   │   └── page.tsx
│   │
│   ├── profile/                 # User profile
│   │   └── page.tsx
│   │
│   ├── results/                 # Test results
│   │   ├── page.tsx            # Results list
│   │   └── [id]/
│   │       └── page.tsx        # Detailed result
│   │
│   ├── admin/                   # Admin panel
│   │   ├── layout.tsx          # Admin layout
│   │   ├── page.tsx            # Admin dashboard
│   │   ├── topics/
│   │   │   └── page.tsx        # Topic management
│   │   ├── tests/
│   │   │   └── [id]/
│   │   │       └── page.tsx    # Test editor
│   │   └── analytics/
│   │       └── page.tsx        # Analytics dashboard
│   │
│   └── api/                     # API routes (Next.js)
│       └── ai/
│           └── generate/
│               └── route.ts    # AI generation endpoint
│
├── components/                  # React components
│   ├── nav.tsx                 # Navigation bar
│   ├── AppShell.tsx            # Layout wrapper
│   ├── ProtectedRoute.tsx      # Auth guard
│   │
│   ├── ui/                     # Reusable UI components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── input.tsx
│   │   ├── select.tsx
│   │   └── ... (57 components)
│   │
│   ├── landing/                # Landing page sections
│   │   ├── hero-section.tsx
│   │   ├── features-section.tsx
│   │   ├── arena-preview-section.tsx
│   │   └── footer.tsx
│   │
│   ├── arena/                  # Arena components
│   │   ├── arena-lobby.tsx
│   │   ├── test-arena.tsx
│   │   └── live-arena.tsx
│   │
│   ├── train/                  # Training components
│   │   ├── training-solver.tsx
│   │   ├── QuestionRenderer.tsx
│   │   ├── SessionSetupPanel.tsx
│   │   └── ai-debrief.tsx
│   │
│   ├── dashboard/              # Dashboard components
│   │   ├── dashboard-content.tsx
│   │   └── performance-chart.tsx
│   │
│   └── auth/                   # Auth components
│       └── login-dialog.tsx
│
├── hooks/                       # Custom React hooks
│   ├── useChat.ts              # Chat WebSocket logic
│   └── useAntiCheat.ts         # Proctoring logic
│
├── lib/                         # Utility functions
│   └── utils.ts
│
├── public/                      # Static assets
│   ├── images/
│   └── fonts/
│
├── package.json                 # Dependencies
├── tsconfig.json               # TypeScript config
├── tailwind.config.ts          # Tailwind config
└── next.config.js              # Next.js config
```

### Key Frontend Concepts

#### 1. **Next.js App Router**
- File-based routing: `app/arena/page.tsx` → `/arena`
- Server components by default (faster initial load)
- Client components with `'use client'` directive

#### 2. **State Management**
- **Local state**: `useState` for component-level state
- **Server state**: Direct API calls (no Redux needed)
- **WebSocket state**: Custom hooks (`useChat`, `useAntiCheat`)

#### 3. **Authentication Flow**
```
User enters credentials
      ↓
POST /api/auth/login
      ↓
Backend validates → Returns JWT token
      ↓
Frontend stores token in localStorage
      ↓
All API requests include: Authorization: Bearer <token>
      ↓
ProtectedRoute component checks token
      ↓
If valid → Show page
If invalid → Redirect to /login
```

#### 4. **WebSocket Connections**

**Chat WebSocket:**
```typescript
// hooks/useChat.ts
const ws = new WebSocket('ws://localhost:8080/ws/chat');

ws.onopen = () => {
  console.log('Connected to chat');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  setMessages(prev => [...prev, message]);
};

ws.send(JSON.stringify({
  type: 'message',
  content: 'Hello world'
}));
```

**Leaderboard WebSocket:**
```typescript
const ws = new WebSocket(`ws://localhost:8080/ws/leaderboard/${testId}`);

ws.onmessage = (event) => {
  const leaderboard = JSON.parse(event.data);
  setRankings(leaderboard);
};
```

#### 5. **Component Patterns**

**Server Component (Default):**
```tsx
// app/dashboard/page.tsx
export default async function DashboardPage() {
  // Can fetch data directly on server
  const stats = await fetch('http://localhost:8080/api/dashboard/stats');
  return <DashboardContent stats={stats} />;
}
```

**Client Component (Interactive):**
```tsx
// components/arena/test-arena.tsx
'use client';

export function TestArena() {
  const [answer, setAnswer] = useState('');
  
  const handleSubmit = async () => {
    await fetch('/api/submissions/code', {
      method: 'POST',
      body: JSON.stringify({ answer })
    });
  };
  
  return <form onSubmit={handleSubmit}>...</form>;
}
```

---

## 6. Backend Architecture

### Directory Structure

```
go-backend/
├── main.go                      # Entry point
│
├── models/                      # Database models (GORM)
│   ├── user.go                 # User model
│   ├── test.go                 # Test, Question, TestCase models
│   ├── attempt.go              # Attempt, Submission models
│   ├── topic.go                # Topic model
│   ├── training.go             # Training models
│   ├── chat.go                 # ChatMessage model
│   └── wrong_question.go       # Analytics models
│
├── handlers/                    # HTTP handlers (controllers)
│   ├── auth.go                 # Login, signup, Google OAuth
│   ├── arena.go                # Test listing, joining
│   ├── attempt.go              # Submission, evaluation
│   ├── chat.go                 # Chat WebSocket, file upload
│   ├── admin.go                # Admin CRUD operations
│   ├── admin_topics.go         # Topic management
│   ├── admin_dashboard.go      # Admin analytics
│   ├── training.go             # Training sessions
│   ├── training_adaptive.go    # Adaptive learning
│   ├── dashboard.go            # User dashboard
│   ├── evaluate.go             # Answer evaluation
│   ├── test_arena.go           # Test execution
│   ├── test_leaderboard.go     # Leaderboard logic
│   ├── global_leaderboard.go   # Global rankings
│   ├── user_results.go         # Results retrieval
│   ├── wrong_questions.go      # Mistake tracking
│   ├── violation_handler.go    # Anti-cheat
│   ├── auto_submit.go          # Auto-submit expired tests
│   ├── ai_test_builder.go      # AI test generation
│   ├── arena_session_ws.go     # Arena WebSocket
│   └── public_topics.go        # Public API
│
├── middleware/                  # HTTP middleware
│   ├── jwt.go                  # JWT authentication
│   └── admin.go                # Admin authorization
│
├── database/                    # Database layer
│   ├── db.go                   # Connection, migrations
│   ├── seed.go                 # Seed data
│   ├── training_seed.go        # Training questions seed
│   └── training_repo.go        # Training repository
│
├── services/                    # Business logic services
│   └── ai_service.go           # OpenAI integration
│
├── judge/                       # Code execution engine
│   ├── service.go              # Judge service
│   └── executor.go             # Docker executor
│
├── chat/                        # Chat WebSocket hub
│   └── hub.go                  # Chat hub logic
│
├── leaderboard/                 # Leaderboard WebSocket hub
│   └── hub.go                  # Leaderboard hub logic
│
├── arena/                       # Arena WebSocket hub
│   └── session_hub.go          # Arena session hub
│
├── uploads/                     # File storage
│   └── chat/                   # Chat file uploads
│
├── .env                         # Environment variables
├── go.mod                       # Go dependencies
└── go.sum                       # Dependency checksums
```


### Backend Architecture Patterns

#### 1. **MVC Pattern (Modified)**

```
Request → Middleware → Handler → Service → Database → Response
```

**Example: User Login**
```go
// handlers/auth.go (Controller)
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    c.BindJSON(&req)
    
    // Service layer
    user, token, err := AuthService.Login(req.Email, req.Password)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }
    
    c.JSON(200, gin.H{
        "token": token,
        "user": user,
    })
}

// services/auth_service.go (Service)
func (s *AuthService) Login(email, password string) (*User, string, error) {
    // Database query
    var user User
    database.DB.Where("email = ?", email).First(&user)
    
    // Verify password
    if !bcrypt.CompareHashAndPassword(user.Password, password) {
        return nil, "", errors.New("invalid password")
    }
    
    // Generate JWT
    token := jwt.GenerateToken(user.ID)
    
    return &user, token, nil
}
```

#### 2. **Middleware Chain**

```go
// main.go
r := gin.Default()

// Global middleware
r.Use(cors.New(corsConfig))
r.Use(logger.Middleware())

// Protected routes
protected := r.Group("/api")
protected.Use(middleware.JWTMiddleware())  // Auth required
{
    protected.GET("/dashboard", handlers.GetDashboard)
}

// Admin routes
admin := r.Group("/api/admin")
admin.Use(middleware.JWTMiddleware())      // Auth required
admin.Use(middleware.AdminOnly())          // Admin role required
{
    admin.POST("/tests", handlers.CreateTest)
}
```

**JWT Middleware:**
```go
// middleware/jwt.go
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "No token"})
            return
        }
        
        // Validate token
        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := jwt.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        // Store user info in context
        c.Set("userID", claims.UserID)
        c.Set("role", claims.Role)
        
        c.Next()
    }
}
```

#### 3. **WebSocket Hub Pattern**

**Hub manages all WebSocket connections:**

```go
// chat/hub.go
type Hub struct {
    clients    map[*Client]bool  // Connected clients
    broadcast  chan []byte        // Broadcast channel
    register   chan *Client       // Register new client
    unregister chan *Client       // Unregister client
    mu         sync.RWMutex       // Thread safety
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            
        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}
```

**Client handles individual connection:**

```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    userID   string
    username string
}

// Read from WebSocket
func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Save to database
        SaveChatMessage(message)
        
        // Broadcast to all clients
        c.hub.broadcast <- message
    }
}

// Write to WebSocket
func (c *Client) writePump() {
    defer c.conn.Close()
    
    for message := range c.send {
        c.conn.WriteMessage(websocket.TextMessage, message)
    }
}
```

#### 4. **Code Execution (Judge System)**

**Sandbox execution using Docker:**

```go
// judge/executor.go
func (e *Executor) Execute(code, language string, input string, timeout int) (*Result, error) {
    // Create temporary file
    filename := fmt.Sprintf("/tmp/code_%s.%s", uuid.New(), getExtension(language))
    ioutil.WriteFile(filename, []byte(code), 0644)
    
    // Build Docker command
    cmd := exec.Command("docker", "run", "--rm",
        "--memory=256m",           // Memory limit
        "--cpus=0.5",              // CPU limit
        "--network=none",          // No network access
        fmt.Sprintf("--timeout=%ds", timeout),
        getDockerImage(language),
        filename,
    )
    
    // Set input
    cmd.Stdin = strings.NewReader(input)
    
    // Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
    defer cancel()
    
    output, err := cmd.CombinedOutput()
    
    return &Result{
        Output:   string(output),
        ExitCode: cmd.ProcessState.ExitCode(),
        Error:    err,
    }, nil
}
```

#### 5. **Database Migrations**

**GORM Auto-Migration:**

```go
// database/db.go
func MigrateModels() {
    // Phase 1: Core tables
    DB.AutoMigrate(
        &models.User{},
        &models.Topic{},
    )
    
    // Phase 2: Content tables
    DB.AutoMigrate(
        &models.Test{},
        &models.TestQuestion{},
        &models.TestCase{},
    )
    
    // Phase 3: Activity tables
    DB.AutoMigrate(
        &models.TestAttempt{},
        &models.TestSubmission{},
    )
}
```

**Model Definition:**

```go
// models/user.go
type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Username  string    `gorm:"unique;not null" json:"username"`
    Password  string    `gorm:"not null" json:"-"`  // Hidden in JSON
    Role      string    `gorm:"default:user" json:"role"`
    Avatar    string    `json:"avatar"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}
```

---

## 7. Database Design

### Entity-Relationship Diagram

```
┌─────────────────┐
│     users       │
├─────────────────┤
│ id (PK)         │
│ email           │
│ username        │
│ password        │
│ role            │
│ avatar          │
│ created_at      │
└────────┬────────┘
         │
         │ 1:N
         │
┌────────▼────────────────┐
│    test_attempts        │
├─────────────────────────┤
│ id (PK)                 │
│ user_id (FK)            │
│ test_id (FK)            │
│ status                  │
│ score                   │
│ started_at              │
│ submitted_at            │
└────────┬────────────────┘
         │
         │ 1:N
         │
┌────────▼────────────────┐
│   test_submissions      │
├─────────────────────────┤
│ id (PK)                 │
│ attempt_id (FK)         │
│ question_id (FK)        │
│ answer                  │
│ is_correct              │
│ points_earned           │
└─────────────────────────┘


┌─────────────────┐
│     tests       │
├─────────────────┤
│ id (PK)         │
│ title           │
│ topic_id (FK)   │
│ difficulty      │
│ duration        │
│ total_points    │
│ is_published    │
│ created_at      │
└────────┬────────┘
         │
         │ 1:N
         │
┌────────▼────────────────┐
│   test_questions        │
├─────────────────────────┤
│ id (PK)                 │
│ test_id (FK)            │
│ question_type           │ (mcq/coding)
│ question_text           │
│ points                  │
│ order_index             │
└────────┬────────────────┘
         │
         ├─────────────────┐
         │ 1:N             │ 1:N
         │                 │
┌────────▼────────┐  ┌────▼──────────────┐
│ test_mcq_options│  │   test_cases      │
├─────────────────┤  ├───────────────────┤
│ id (PK)         │  │ id (PK)           │
│ question_id (FK)│  │ question_id (FK)  │
│ option_text     │  │ input             │
│ is_correct      │  │ expected_output   │
└─────────────────┘  │ is_hidden         │
                     └───────────────────┘


┌─────────────────┐
│     topics      │
├─────────────────┤
│ id (PK)         │
│ name            │
│ slug            │
│ description     │
└────────┬────────┘
         │
         │ 1:N
         │
┌────────▼────────────────┐
│  user_topic_stats       │
├─────────────────────────┤
│ id (PK)                 │
│ user_id (FK)            │
│ topic_id (FK)           │
│ total_attempts          │
│ correct_answers         │
│ accuracy_rate           │
│ last_practiced          │
└─────────────────────────┘


┌─────────────────────────┐
│   chat_messages         │
├─────────────────────────┤
│ id (PK)                 │
│ user_id (FK)            │
│ username                │
│ avatar                  │
│ message_type            │ (text/note/image/pdf)
│ content                 │
│ file_name               │
│ created_at              │
└─────────────────────────┘
```


### Database Schema Details

#### Core Tables

**1. users**
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    avatar TEXT,
    google_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_username (username)
);
```

**2. topics**
```sql
CREATE TABLE topics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_slug (slug)
);
```

**3. tests**
```sql
CREATE TABLE tests (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    topic_id BIGINT,
    difficulty VARCHAR(20),
    duration INT,  -- in minutes
    total_points INT DEFAULT 0,
    is_published BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT FALSE,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (topic_id) REFERENCES topics(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_topic (topic_id),
    INDEX idx_published (is_published, is_active)
);
```

**4. test_questions**
```sql
CREATE TABLE test_questions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    test_id BIGINT NOT NULL,
    question_type VARCHAR(20) NOT NULL,  -- 'mcq' or 'coding'
    question_text TEXT NOT NULL,
    points INT DEFAULT 1,
    order_index INT DEFAULT 0,
    time_limit INT,  -- seconds for coding questions
    memory_limit INT,  -- MB for coding questions
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE,
    INDEX idx_test (test_id),
    INDEX idx_order (test_id, order_index)
);
```

**5. test_mcq_options**
```sql
CREATE TABLE test_mcq_options (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    question_id BIGINT NOT NULL,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT FALSE,
    order_index INT DEFAULT 0,
    FOREIGN KEY (question_id) REFERENCES test_questions(id) ON DELETE CASCADE,
    INDEX idx_question (question_id)
);
```

**6. test_cases**
```sql
CREATE TABLE test_cases (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    question_id BIGINT NOT NULL,
    input TEXT,
    expected_output TEXT NOT NULL,
    is_hidden BOOLEAN DEFAULT FALSE,  -- Hidden test cases
    points INT DEFAULT 1,
    FOREIGN KEY (question_id) REFERENCES test_questions(id) ON DELETE CASCADE,
    INDEX idx_question (question_id)
);
```

#### Activity Tables

**7. test_attempts**
```sql
CREATE TABLE test_attempts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    test_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'in_progress',  -- in_progress, submitted, expired
    score DECIMAL(5,2) DEFAULT 0,
    max_score INT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMP NULL,
    expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (test_id) REFERENCES tests(id),
    INDEX idx_user (user_id),
    INDEX idx_test (test_id),
    INDEX idx_status (status),
    INDEX idx_expires (expires_at)
);
```

**8. test_submissions**
```sql
CREATE TABLE test_submissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    attempt_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    submission_type VARCHAR(20),  -- 'mcq' or 'code'
    answer TEXT,  -- Selected option ID or code
    language VARCHAR(50),  -- For coding questions
    is_correct BOOLEAN DEFAULT FALSE,
    points_earned DECIMAL(5,2) DEFAULT 0,
    execution_time INT,  -- milliseconds
    memory_used INT,  -- KB
    test_cases_passed INT DEFAULT 0,
    test_cases_total INT DEFAULT 0,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (attempt_id) REFERENCES test_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES test_questions(id),
    INDEX idx_attempt (attempt_id),
    INDEX idx_question (question_id)
);
```

**9. test_results**
```sql
CREATE TABLE test_results (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    attempt_id BIGINT UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    test_id BIGINT NOT NULL,
    total_score DECIMAL(5,2),
    max_score INT,
    percentage DECIMAL(5,2),
    rank INT,
    time_taken INT,  -- seconds
    questions_attempted INT,
    questions_correct INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (attempt_id) REFERENCES test_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (test_id) REFERENCES tests(id),
    INDEX idx_user (user_id),
    INDEX idx_test (test_id),
    INDEX idx_score (test_id, total_score DESC)
);
```

**10. test_violations**
```sql
CREATE TABLE test_violations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    attempt_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    violation_type VARCHAR(50) NOT NULL,  -- tab_switch, copy_paste, etc.
    severity VARCHAR(20) DEFAULT 'low',  -- low, medium, high
    details TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (attempt_id) REFERENCES test_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_attempt (attempt_id),
    INDEX idx_user (user_id)
);
```

#### Analytics Tables

**11. user_wrong_questions**
```sql
CREATE TABLE user_wrong_questions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    topic_id BIGINT,
    attempt_id BIGINT,
    user_answer TEXT,
    correct_answer TEXT,
    times_wrong INT DEFAULT 1,
    last_attempted TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_reviewed BOOLEAN DEFAULT FALSE,
    is_mastered BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (question_id) REFERENCES test_questions(id),
    FOREIGN KEY (topic_id) REFERENCES topics(id),
    FOREIGN KEY (attempt_id) REFERENCES test_attempts(id),
    INDEX idx_user (user_id),
    INDEX idx_topic (user_id, topic_id),
    INDEX idx_mastered (user_id, is_mastered)
);
```

**12. user_topic_stats**
```sql
CREATE TABLE user_topic_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    topic_id BIGINT NOT NULL,
    total_attempts INT DEFAULT 0,
    correct_answers INT DEFAULT 0,
    wrong_answers INT DEFAULT 0,
    accuracy_rate DECIMAL(5,2) DEFAULT 0,
    last_practiced TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (topic_id) REFERENCES topics(id),
    UNIQUE KEY unique_user_topic (user_id, topic_id),
    INDEX idx_user (user_id),
    INDEX idx_accuracy (user_id, accuracy_rate)
);
```

#### Communication Tables

**13. chat_messages**
```sql
CREATE TABLE chat_messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    avatar TEXT,
    message_type VARCHAR(20) DEFAULT 'text',  -- text, note, image, pdf
    content TEXT NOT NULL,
    file_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_created (created_at DESC)
);
```

### Database Indexes Strategy

**Why Indexes?**
- Speed up queries (SELECT, WHERE, JOIN)
- Trade-off: Slower writes, more storage

**Indexing Rules:**
1. **Primary Keys**: Auto-indexed
2. **Foreign Keys**: Always index
3. **WHERE clauses**: Index frequently filtered columns
4. **ORDER BY**: Index sorted columns
5. **Composite indexes**: For multi-column queries

**Example Query Optimization:**

```sql
-- Slow query (no index)
SELECT * FROM test_attempts WHERE user_id = 123 AND status = 'submitted';
-- Full table scan: O(n)

-- Fast query (with composite index)
CREATE INDEX idx_user_status ON test_attempts(user_id, status);
-- Index lookup: O(log n)
```

---

## 8. Key Features & Flows

### Feature 1: User Authentication

#### Registration Flow

```
1. User fills registration form
   ↓
2. Frontend validates (email format, password strength)
   ↓
3. POST /api/auth/signup
   {
     "email": "user@example.com",
     "username": "john_doe",
     "password": "SecurePass123"
   }
   ↓
4. Backend validates:
   - Email not already registered
   - Username not taken
   - Password meets requirements
   ↓
5. Hash password with bcrypt
   hashedPassword = bcrypt.GenerateFromPassword(password, 10)
   ↓
6. Insert into database
   INSERT INTO users (email, username, password, role)
   VALUES ('user@example.com', 'john_doe', hashedPassword, 'user')
   ↓
7. Generate JWT token
   token = jwt.Sign({
     user_id: user.ID,
     role: user.Role,
     exp: time.Now().Add(24 * time.Hour)
   })
   ↓
8. Return response
   {
     "token": "eyJhbGciOiJIUzI1NiIs...",
     "user": {
       "id": 1,
       "email": "user@example.com",
       "username": "john_doe",
       "role": "user"
     }
   }
   ↓
9. Frontend stores token in localStorage
   ↓
10. Redirect to /dashboard
```

#### Login Flow

```
1. User enters credentials
   ↓
2. POST /api/auth/login
   {
     "email": "user@example.com",
     "password": "SecurePass123"
   }
   ↓
3. Backend queries database
   SELECT * FROM users WHERE email = 'user@example.com'
   ↓
4. Compare password
   bcrypt.CompareHashAndPassword(user.Password, inputPassword)
   ↓
5. If match:
   - Generate JWT token
   - Return token + user data
   ↓
6. If no match:
   - Return 401 Unauthorized
```

#### Google OAuth Flow

```
1. User clicks "Sign in with Google"
   ↓
2. Redirect to Google OAuth consent screen
   ↓
3. User approves
   ↓
4. Google redirects back with authorization code
   ↓
5. Frontend sends code to backend
   POST /api/auth/google
   { "code": "4/0AY0e-g7..." }
   ↓
6. Backend exchanges code for user info
   - Call Google API
   - Get email, name, picture
   ↓
7. Check if user exists
   SELECT * FROM users WHERE google_id = 'google_user_id'
   ↓
8. If not exists:
   - Create new user
   - Set google_id
   ↓
9. Generate JWT token
   ↓
10. Return token + user data
```


### Feature 2: Taking a Test (Arena Mode)

#### Complete Test Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 1: Test Discovery & Join                                  │
└─────────────────────────────────────────────────────────────────┘

1. User navigates to /arena
   ↓
2. Frontend: GET /api/arena/active
   ↓
3. Backend queries:
   SELECT * FROM tests 
   WHERE is_published = TRUE 
   AND is_active = TRUE
   ↓
4. Display test cards with:
   - Title, difficulty, duration
   - Topic, total points
   - "Join Test" button
   ↓
5. User clicks "Join Test"
   ↓
6. Frontend: POST /api/arena/tests/:id/join
   Headers: { Authorization: Bearer <token> }
   ↓
7. Backend:
   a. Validate user is authenticated
   b. Check if user already has active attempt
   c. Create new test_attempt:
      INSERT INTO test_attempts (
        user_id, test_id, status, started_at, expires_at
      ) VALUES (
        user.ID, test.ID, 'in_progress', NOW(), NOW() + test.Duration
      )
   d. Return attempt_id
   ↓
8. Frontend redirects to /arena/:testId/play?attemptId=<attempt_id>

┌─────────────────────────────────────────────────────────────────┐
│ PHASE 2: Test Taking Interface                                  │
└─────────────────────────────────────────────────────────────────┘

9. Frontend: GET /api/attempts/:attemptId
   ↓
10. Backend returns:
    {
      "attempt": { id, status, started_at, expires_at },
      "test": { title, duration, total_points },
      "questions": [
        {
          "id": 1,
          "type": "mcq",
          "text": "What is 2+2?",
          "points": 1,
          "options": [
            { "id": 1, "text": "3" },
            { "id": 2, "text": "4" },
            { "id": 3, "text": "5" }
          ]
        },
        {
          "id": 2,
          "type": "coding",
          "text": "Write a function to reverse a string",
          "points": 10,
          "test_cases": [
            { "input": "hello", "output": "olleh" }
          ]
        }
      ]
    }
   ↓
11. Frontend displays:
    - Timer (countdown from expires_at)
    - Question navigator
    - Current question
    - Answer input (MCQ options or code editor)
    - "Save Draft" and "Submit" buttons
   ↓
12. Anti-cheat monitoring starts:
    - Track tab switches
    - Detect copy/paste
    - Monitor full-screen exit
    - Log violations to backend

┌─────────────────────────────────────────────────────────────────┐
│ PHASE 3: Answering Questions                                    │
└─────────────────────────────────────────────────────────────────┘

13a. MCQ Question:
    User selects option
    ↓
    Frontend: POST /api/submissions/mcq
    {
      "attempt_id": 123,
      "question_id": 1,
      "selected_option_id": 2
    }
    ↓
    Backend:
    - Check if option is correct
    - Calculate points
    - Save submission:
      INSERT INTO test_submissions (
        attempt_id, question_id, submission_type,
        answer, is_correct, points_earned
      ) VALUES (
        123, 1, 'mcq', '2', TRUE, 1
      )
    ↓
    Return: { "saved": true }

13b. Coding Question:
    User writes code
    ↓
    User clicks "Run Code"
    ↓
    Frontend: POST /api/submissions/run
    {
      "code": "def reverse(s): return s[::-1]",
      "language": "python",
      "input": "hello"
    }
    ↓
    Backend (Judge System):
    a. Create temp file with code
    b. Run in Docker container:
       docker run --rm \
         --memory=256m \
         --cpus=0.5 \
         --network=none \
         --timeout=5s \
         python:3.9 \
         python /tmp/code.py
    c. Capture output
    d. Return:
       {
         "output": "olleh",
         "execution_time": 45,  // ms
         "memory_used": 12,     // MB
         "error": null
       }
    ↓
    User clicks "Submit Code"
    ↓
    Frontend: POST /api/submissions/code
    {
      "attempt_id": 123,
      "question_id": 2,
      "code": "def reverse(s): return s[::-1]",
      "language": "python"
    }
    ↓
    Backend:
    a. Run code against all test cases
    b. Calculate score:
       - 2 test cases passed out of 3
       - Points = (2/3) * 10 = 6.67
    c. Save submission:
       INSERT INTO test_submissions (
         attempt_id, question_id, submission_type,
         answer, language, is_correct, points_earned,
         test_cases_passed, test_cases_total
       ) VALUES (
         123, 2, 'code', 'def reverse...', 'python',
         FALSE, 6.67, 2, 3
       )
    ↓
    Return: {
      "test_cases_passed": 2,
      "test_cases_total": 3,
      "points_earned": 6.67
    }

┌─────────────────────────────────────────────────────────────────┐
│ PHASE 4: Test Submission                                        │
└─────────────────────────────────────────────────────────────────┘

14. User clicks "Submit Test" (or timer expires)
    ↓
15. Frontend: POST /api/attempts/:attemptId/submit
    ↓
16. Backend:
    a. Update attempt status:
       UPDATE test_attempts 
       SET status = 'submitted', submitted_at = NOW()
       WHERE id = attemptId
    
    b. Calculate total score:
       SELECT SUM(points_earned) as total_score
       FROM test_submissions
       WHERE attempt_id = attemptId
    
    c. Get max possible score:
       SELECT SUM(points) as max_score
       FROM test_questions
       WHERE test_id = testId
    
    d. Calculate percentage:
       percentage = (total_score / max_score) * 100
    
    e. Calculate rank:
       SELECT COUNT(*) + 1 as rank
       FROM test_results
       WHERE test_id = testId
       AND total_score > current_user_score
    
    f. Create test result:
       INSERT INTO test_results (
         attempt_id, user_id, test_id,
         total_score, max_score, percentage, rank,
         time_taken, questions_attempted, questions_correct
       ) VALUES (...)
    
    g. Update leaderboard (WebSocket broadcast):
       leaderboardHub.Broadcast({
         "type": "update",
         "test_id": testId,
         "leaderboard": [
           { "rank": 1, "username": "alice", "score": 95 },
           { "rank": 2, "username": "bob", "score": 87 },
           ...
         ]
       })
    
    h. Track wrong answers:
       INSERT INTO user_wrong_questions (
         user_id, question_id, topic_id, attempt_id,
         user_answer, correct_answer
       )
       SELECT ...
       FROM test_submissions
       WHERE attempt_id = attemptId AND is_correct = FALSE
    
    i. Update topic stats:
       INSERT INTO user_topic_stats (...)
       ON DUPLICATE KEY UPDATE
         total_attempts = total_attempts + 1,
         correct_answers = correct_answers + new_correct,
         accuracy_rate = (correct_answers / total_attempts) * 100
    ↓
17. Return result:
    {
      "result_id": 456,
      "total_score": 76.67,
      "max_score": 100,
      "percentage": 76.67,
      "rank": 5,
      "time_taken": 1800,  // seconds
      "questions_attempted": 10,
      "questions_correct": 7
    }
    ↓
18. Frontend redirects to /results/:resultId

┌─────────────────────────────────────────────────────────────────┐
│ PHASE 5: Results Display                                        │
└─────────────────────────────────────────────────────────────────┘

19. Frontend: GET /api/results/:resultId
    ↓
20. Backend returns:
    {
      "result": { score, percentage, rank, time_taken },
      "test": { title, difficulty, topic },
      "submissions": [
        {
          "question": "What is 2+2?",
          "your_answer": "4",
          "correct_answer": "4",
          "is_correct": true,
          "points_earned": 1
        },
        {
          "question": "Reverse string",
          "your_answer": "def reverse...",
          "test_cases_passed": 2,
          "test_cases_total": 3,
          "points_earned": 6.67
        }
      ],
      "leaderboard": [
        { "rank": 1, "username": "alice", "score": 95 },
        { "rank": 5, "username": "you", "score": 76.67 },
        ...
      ]
    }
    ↓
21. Display:
    - Score card with percentage
    - Rank badge
    - Time taken
    - Question-by-question breakdown
    - Leaderboard position
    - "Review Mistakes" button
```


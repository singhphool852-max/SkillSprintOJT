# Interview Terms Explained - SkillSprint Project Reference

## Table of Contents
1. [Problem Statement](#problem-statement)
2. [List of Services](#list-of-services)
3. [Tech Stack](#tech-stack)
4. [User Flow](#user-flow)
5. [APIs](#apis)
6. [Storage](#storage)
7. [Scaling Strategy](#scaling-strategy)
8. [Optimization](#optimization)
9. [Failure Handling](#failure-handling)
10. [Bottlenecks](#bottlenecks)

---

## 1. Problem Statement

### What is SkillSprint?

**Problem**: Traditional learning platforms lack real-time competitive elements and personalized training. Students need:
- A way to test their skills in live competitive environments
- Adaptive learning based on their performance
- Real-time feedback and community interaction
- Comprehensive progress tracking

**Solution**: SkillSprint is a competitive learning platform that combines:
- **Live Arenas**: Real-time competitive tests where users compete against each other
- **Training Mode**: Personalized practice sessions with AI-generated questions
- **Community Chat**: Real-time communication for peer learning
- **Analytics Dashboard**: Track performance, identify weak areas, and monitor progress
- **Anti-Cheat System**: Ensure fair competition with proctoring features


### Core Requirements

**Functional Requirements**:
- User authentication (local + Google OAuth)
- Create and manage tests (MCQ + coding questions)
- Live arena sessions with real-time leaderboards
- Code execution in sandboxed environment
- Real-time chat with file sharing
- Performance analytics and progress tracking
- Admin panel for content management

**Non-Functional Requirements**:
- **Scalability**: Support 10,000+ concurrent users
- **Low Latency**: <200ms API response, <100ms WebSocket updates
- **Availability**: 99.9% uptime
- **Security**: JWT authentication, input validation, anti-cheat measures
- **Consistency**: Strong consistency for test submissions, eventual consistency for leaderboards

---

## 2. List of Services

### Microservices Architecture (Current: Monolithic, Future: Microservices)


#### Current Services in SkillSprint

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Service                          │
│  • Next.js 14 (React)                                        │
│  • Server-side rendering                                     │
│  • Client-side routing                                       │
│  • WebSocket client connections                              │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/WebSocket
┌────────────────────────▼────────────────────────────────────┐
│                    Backend Service (Go)                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  1. Authentication Service                            │  │
│  │     • JWT token generation/validation                 │  │
│  │     • Google OAuth integration                        │  │
│  │     • Password hashing (bcrypt)                       │  │
│  │     Files: handlers/auth.go                           │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  2. Arena Service                                     │  │
│  │     • Test management (CRUD)                          │  │
│  │     • Test attempt tracking                           │  │
│  │     • Live session management                         │  │
│  │     Files: handlers/arena.go, test_arena.go          │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  3. Judge Service                                     │  │
│  │     • Code execution in Docker containers             │  │
│  │     • Test case validation                            │  │
│  │     • Partial scoring                                 │  │
│  │     Files: judge/executor.go, judge/service.go        │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  4. Chat Service                                      │  │
│  │     • WebSocket hub for real-time messaging           │  │
│  │     • File upload handling                            │  │
│  │     • Message persistence                             │  │
│  │     Files: chat/hub.go, handlers/chat.go              │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  5. Leaderboard Service                               │  │
│  │     • Real-time ranking updates                       │  │
│  │     • WebSocket broadcasting                          │  │
│  │     • Score aggregation                               │  │
│  │     Files: leaderboard/hub.go                         │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  6. Training Service                                  │  │
│  │     • Adaptive question selection                     │  │
│  │     • Performance tracking                            │  │
│  │     • Weak area identification                        │  │
│  │     Files: handlers/training.go                       │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  7. AI Service                                        │  │
│  │     • OpenAI integration for test generation          │  │
│  │     • Question synthesis from notes                   │  │
│  │     • AI-powered debrief                              │  │
│  │     Files: services/ai_service.go                     │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  8. Admin Service                                     │  │
│  │     • Content management                              │  │
│  │     • User management                                 │  │
│  │     • Analytics dashboard                             │  │
│  │     Files: handlers/admin.go                          │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  9. Anti-Cheat Service                                │  │
│  │     • Tab switch detection                            │  │
│  │     • Copy-paste monitoring                           │  │
│  │     • Violation logging                               │  │
│  │     Files: handlers/violation_handler.go              │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Database Layer                            │
│  • MySQL (primary data store)                                │
│  • GORM ORM                                                  │
│  Files: database/db.go, models/*.go                          │
└─────────────────────────────────────────────────────────────┘
```

### Service Communication

**Current**: All services in single Go binary (monolithic)
- **Pros**: Simple deployment, no network overhead, easier debugging
- **Cons**: Tight coupling, harder to scale individual services

**Future Migration Path**:
```
Monolith → Modular Monolith → Microservices
```


### Proper Use of Service Names

In SkillSprint context:
- **Authentication Service** = User login/signup/OAuth (handlers/auth.go)
- **Arena Service** = Test management and execution (handlers/arena.go)
- **Judge Service** = Code execution engine (judge/executor.go)
- **Chat Service** = Real-time messaging (chat/hub.go)
- **Leaderboard Service** = Real-time rankings (leaderboard/hub.go)
- **AI Service** = OpenAI integration (services/ai_service.go)

---

## 3. Tech Stack

### Why This Tech Stack?

#### Frontend: Next.js + TypeScript + Tailwind

**Next.js 14 (React Framework)**:
- **Server-Side Rendering (SSR)**: Faster initial page load, better SEO
- **App Router**: File-based routing, nested layouts
- **API Routes**: Backend-for-frontend pattern
- **Example in SkillSprint**: `app/arena/[id]/play/page.tsx` → `/arena/123/play`

**TypeScript**:
- **Type Safety**: Catch errors at compile time
- **Better IDE Support**: Autocomplete, refactoring
- **Example**:
```typescript
interface TestAttempt {
  id: string;
  testId: string;
  userId: string;
  score: number;
  submittedAt: Date;
}
```

**Tailwind CSS**:
- **Utility-First**: Rapid UI development
- **Responsive**: Mobile-first design
- **Example**: `className="flex items-center justify-between p-4 bg-gray-900 rounded-lg"`


#### Backend: Go + Gin + GORM

**Go (Golang)**:
- **Performance**: Compiled language, fast execution
- **Concurrency**: Goroutines for handling multiple WebSocket connections
- **Example in SkillSprint**: Chat hub runs goroutines for each client
```go
// chat/hub.go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case message := <-h.broadcast:
            for client := range h.clients {
                client.send <- message
            }
        }
    }
}
```

**Gin Framework**:
- **Fast HTTP Router**: Radix tree-based routing
- **Middleware Support**: JWT auth, CORS, logging
- **Example**:
```go
r := gin.Default()
r.POST("/api/auth/login", handlers.LoginHandler)
r.GET("/api/arena/tests", middleware.JWTMiddleware(), handlers.GetTests)
```

**GORM (ORM)**:
- **Database Abstraction**: Write Go structs, not SQL
- **Auto-Migration**: Schema changes automatically applied
- **Example**:
```go
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Email    string `gorm:"unique;not null"`
    Username string `gorm:"unique;not null"`
}
database.DB.AutoMigrate(&User{})
```

#### Database: MySQL

**Why MySQL over PostgreSQL/MongoDB?**:
- **ACID Compliance**: Strong consistency for test submissions
- **Relational Data**: Users, tests, attempts have clear relationships
- **Mature Ecosystem**: Well-documented, widely supported
- **Example Schema**:
```sql
users → test_attempts → tests → test_questions → test_cases
```


#### Additional Technologies

**Docker**:
- **Use Case**: Sandboxed code execution in Judge Service
- **Why**: Isolate user code, prevent system access
- **Example**:
```go
// judge/executor.go
cmd := exec.Command("docker", "run", "--rm",
    "--memory=256m",
    "--cpus=0.5",
    "--network=none",
    "python:3.9",
    "python", "/code/solution.py")
```

**WebSocket (Gorilla)**:
- **Use Case**: Real-time chat, leaderboard updates
- **Why**: Bidirectional communication, low latency
- **Example**: Chat messages broadcast to all connected clients instantly

**JWT (JSON Web Tokens)**:
- **Use Case**: Stateless authentication
- **Why**: No server-side session storage needed
- **Example**:
```
Header: Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Payload: { "user_id": 123, "role": "user", "exp": 1234567890 }
```

**OpenAI API**:
- **Use Case**: AI test generation, question synthesis
- **Why**: Generate contextual questions from user notes
- **Example**: Upload PDF notes → AI generates 10 MCQs

---

## 4. User Flow

### Complete User Journey in SkillSprint


#### Flow 1: User Registration & Login

```
1. User visits landing page (/)
   ↓
2. Clicks "Get Started" → Redirects to /register
   ↓
3. Fills form (email, username, password)
   ↓
4. Frontend: POST /api/auth/signup
   ↓
5. Backend: 
   - Validate input
   - Hash password (bcrypt)
   - Insert into users table
   - Generate JWT token
   ↓
6. Frontend: Store token in localStorage
   ↓
7. Redirect to /dashboard
```

**Alternative: Google OAuth**
```
1. Click "Sign in with Google"
   ↓
2. Redirect to Google OAuth consent screen
   ↓
3. Google returns authorization code
   ↓
4. Backend: Exchange code for user info
   ↓
5. Create/update user in database
   ↓
6. Generate JWT token
   ↓
7. Redirect to /dashboard
```

#### Flow 2: Taking a Live Arena Test

```
1. User navigates to /arena
   ↓
2. Frontend: GET /api/arena/tests (fetch available tests)
   ↓
3. User clicks "Join Test"
   ↓
4. Frontend: POST /api/arena/tests/:id/join
   ↓
5. Backend:
   - Create TestAttempt record
   - Return attempt_id and questions
   ↓
6. Frontend: Display test interface
   ↓
7. User answers questions:
   - MCQ: Select option
   - Coding: Write code → POST /api/submissions/run (test against sample cases)
   ↓
8. User clicks "Submit Test"
   ↓
9. Frontend: POST /api/attempts/:id/submit
   ↓
10. Backend:
    - Evaluate all answers
    - Calculate score (partial scoring for coding)
    - Update test_attempts table
    - Broadcast to leaderboard WebSocket
    ↓
11. Frontend: Redirect to /results/:id
    ↓
12. Display score, correct answers, leaderboard rank
```


#### Flow 3: Real-Time Chat

```
1. User navigates to /chat
   ↓
2. Frontend: GET /api/chat/history (fetch last 50 messages)
   ↓
3. Frontend: Connect WebSocket (ws://localhost:8080/ws/chat)
   ↓
4. Backend: Upgrade HTTP to WebSocket
   ↓
5. Backend: Register client in Hub
   ↓
6. Backend: Broadcast "user_joined" event to all clients
   ↓
7. User types message and clicks send
   ↓
8. Frontend: Send JSON via WebSocket
   {
     "type": "message",
     "message_type": "text",
     "content": "Hello world"
   }
   ↓
9. Backend: Client.readPump() receives message
   ↓
10. Backend: Save to chat_messages table
    ↓
11. Backend: Broadcast to all connected clients
    ↓
12. Frontend: All users see message instantly
```

**File Upload Flow**:
```
1. User clicks "Upload File"
   ↓
2. Frontend: POST /api/chat/upload (multipart/form-data)
   ↓
3. Backend:
   - Validate file type (image/pdf only)
   - Validate size (<10MB)
   - Sanitize filename
   - Save to uploads/chat/
   ↓
4. Backend: Return file URL
   ↓
5. Frontend: Send message via WebSocket with file URL
   ↓
6. Backend: Broadcast file message
   ↓
7. All users see file (image preview or PDF link)
```

#### Flow 4: AI Test Generation

```
1. Admin navigates to /admin/tests/new
   ↓
2. Clicks "Generate with AI"
   ↓
3. Uploads PDF notes or enters topic
   ↓
4. Frontend: POST /api/ai/generate
   ↓
5. Backend:
   - Extract text from PDF
   - Call OpenAI API with prompt:
     "Generate 10 MCQ questions on [topic] with 4 options each"
   ↓
6. OpenAI returns JSON with questions
   ↓
7. Backend: Parse and return to frontend
   ↓
8. Frontend: Display generated questions
   ↓
9. Admin reviews and edits
   ↓
10. Admin clicks "Save Test"
    ↓
11. Frontend: POST /api/admin/tests
    ↓
12. Backend: Insert test, questions, options into database
```

---

## 5. APIs

### RESTful API Design in SkillSprint


#### Stateless APIs

**What is Stateless?**
- Server doesn't store client session data
- Each request contains all necessary information (JWT token)
- Enables horizontal scaling (any server can handle any request)

**Example in SkillSprint**:
```
Request 1: GET /api/dashboard/stats
Header: Authorization: Bearer <token>

Request 2: POST /api/arena/tests/123/join
Header: Authorization: Bearer <token>

Both requests can be handled by different backend servers
```

#### Horizontal Scaling

**What is Horizontal Scaling?**
- Add more servers instead of upgrading one server (vertical scaling)
- Load balancer distributes requests across servers

**SkillSprint Scaling**:
```
Current (Single Server):
Client → Backend Server → Database

Scaled (Multiple Servers):
                    ┌─ Backend Server 1 ─┐
Client → Load Balancer ─ Backend Server 2 ─┤→ Database
                    └─ Backend Server 3 ─┘
```

**Why Stateless Enables Scaling**:
- No sticky sessions needed
- Any server can validate JWT token
- Database is single source of truth


#### When to Scale: CPU Spike or Queue Depth

**CPU Spike**:
- **Indicator**: Server CPU usage >80%
- **Cause**: Too many requests for single server
- **Solution**: Add more backend servers
- **SkillSprint Example**: During live arena with 1000+ concurrent users

**Queue Depth**:
- **Indicator**: Requests waiting in queue
- **Cause**: Server processing slower than request rate
- **Solution**: Add workers or scale horizontally
- **SkillSprint Example**: Code execution queue (judge service)

**Monitoring in SkillSprint**:
```go
// Metrics to track
- Request rate (requests/second)
- Response time (p50, p95, p99)
- CPU usage (%)
- Memory usage (%)
- Active WebSocket connections
- Judge queue length
```

---

## 6. Storage

### Database Design in SkillSprint

#### DB, Object Storage, CDN

**Database (MySQL)**:
- **Use**: Structured data (users, tests, attempts)
- **Why**: ACID transactions, relationships
- **Example Tables**:
  - `users`: User accounts
  - `tests`: Test definitions
  - `test_attempts`: User submissions
  - `chat_messages`: Chat history

**Object Storage (Future: S3)**:
- **Use**: Large files (images, PDFs, videos)
- **Why**: Cheaper than DB, scalable
- **SkillSprint Use Cases**:
  - Chat file uploads (currently local disk)
  - User profile pictures
  - Test attachments

**CDN (Future: CloudFront)**:
- **Use**: Static assets (images, CSS, JS)
- **Why**: Serve from edge locations, reduce latency
- **SkillSprint Use Cases**:
  - Frontend assets
  - Chat images
  - Profile pictures


#### DB Schema

**SkillSprint Database Schema**:

```sql
-- Core Tables
users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(100) UNIQUE NOT NULL,
  password VARCHAR(255),
  role VARCHAR(20) DEFAULT 'user',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

topics (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  description TEXT
)

tests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  topic_id BIGINT,
  difficulty VARCHAR(20),
  duration INT,  -- minutes
  total_marks INT,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP,
  FOREIGN KEY (topic_id) REFERENCES topics(id)
)

test_questions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  test_id BIGINT,
  question_type VARCHAR(20),  -- mcq, coding, subjective
  question_text TEXT,
  marks INT,
  order_index INT,
  FOREIGN KEY (test_id) REFERENCES tests(id)
)

test_mcq_options (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  question_id BIGINT,
  option_text TEXT,
  is_correct BOOLEAN,
  FOREIGN KEY (question_id) REFERENCES test_questions(id)
)

test_cases (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  question_id BIGINT,
  input TEXT,
  expected_output TEXT,
  is_sample BOOLEAN,
  FOREIGN KEY (question_id) REFERENCES test_questions(id)
)

-- Activity Tables
test_attempts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  test_id BIGINT,
  score DECIMAL(5,2),
  max_score INT,
  status VARCHAR(20),  -- in_progress, submitted, evaluated
  started_at TIMESTAMP,
  submitted_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (test_id) REFERENCES tests(id),
  INDEX (user_id, submitted_at)
)

test_submissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  attempt_id BIGINT,
  question_id BIGINT,
  answer TEXT,
  is_correct BOOLEAN,
  marks_obtained DECIMAL(5,2),
  submitted_at TIMESTAMP,
  FOREIGN KEY (attempt_id) REFERENCES test_attempts(id),
  FOREIGN KEY (question_id) REFERENCES test_questions(id)
)

-- Chat Tables
chat_messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id VARCHAR(255),
  username VARCHAR(100),
  avatar VARCHAR(255),
  message_type VARCHAR(20),  -- text, note, image, pdf
  content TEXT,
  file_name VARCHAR(255),
  created_at TIMESTAMP,
  INDEX (created_at)
)

-- Analytics Tables
user_wrong_questions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  question_id BIGINT,
  attempt_count INT DEFAULT 1,
  last_attempted TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (question_id) REFERENCES test_questions(id),
  UNIQUE KEY (user_id, question_id)
)
```


#### Indexing

**What is Indexing?**
- Data structure (B-tree) for fast lookups
- Trade-off: Faster reads, slower writes

**Indexes in SkillSprint**:

```sql
-- Primary Key Index (automatic)
users(id) → Fast lookup by user ID

-- Unique Index (automatic on UNIQUE constraint)
users(email) → Fast login by email
users(username) → Fast username availability check

-- Composite Index
test_attempts(user_id, submitted_at) → Fast query for user's test history
INDEX (user_id, submitted_at)

-- Why Composite?
SELECT * FROM test_attempts 
WHERE user_id = 123 
ORDER BY submitted_at DESC 
LIMIT 10;
-- Uses index for both WHERE and ORDER BY
```

**When to Add Index**:
- Columns in WHERE clause
- Columns in ORDER BY
- Foreign keys
- Columns in JOIN conditions

**When NOT to Index**:
- Small tables (<1000 rows)
- Columns with low cardinality (e.g., boolean)
- Frequently updated columns

---

## 7. Scaling Strategy

### Vertical Scaling vs Horizontal Scaling


#### Vertical Scaling (Scale Up)

**Definition**: Increase resources of single server (more CPU, RAM, disk)

**Pros**:
- Simple (no code changes)
- No distributed system complexity
- Consistent performance

**Cons**:
- Hardware limits (can't scale infinitely)
- Single point of failure
- Expensive at high end
- Downtime during upgrade

**SkillSprint Example**:
```
Current: 2 CPU, 4GB RAM
Scaled:  8 CPU, 32GB RAM
```

#### Horizontal Scaling (Scale Out)

**Definition**: Add more servers

**Pros**:
- No hardware limits
- High availability (one server fails, others continue)
- Cost-effective (use commodity hardware)

**Cons**:
- Complex (load balancing, data consistency)
- Network latency between servers
- Requires stateless design

**SkillSprint Horizontal Scaling Plan**:

```
Phase 1: Single Server (Current)
┌─────────────┐
│   Backend   │
│   Server    │
└─────────────┘

Phase 2: Load Balanced (Target: 10K users)
                    ┌─ Backend 1 ─┐
Client → Load Balancer ─ Backend 2 ─┤→ MySQL
                    └─ Backend 3 ─┘

Phase 3: Database Scaling (Target: 100K users)
                    ┌─ Backend 1 ─┐     ┌─ MySQL Master ─┐
Client → Load Balancer ─ Backend 2 ─┤→ ─┤─ MySQL Replica 1
                    └─ Backend 3 ─┘     └─ MySQL Replica 2

Phase 4: Microservices (Target: 1M users)
                    ┌─ Auth Service ─┐
                    ├─ Arena Service ─┤
Client → API Gateway ─┤─ Judge Service ─┤→ MySQL Cluster
                    ├─ Chat Service ─┤   Redis Cluster
                    └─ AI Service ─┘
```


### When to Scale?

**Metrics to Monitor**:

1. **CPU Usage**:
   - Normal: <60%
   - Warning: 60-80%
   - Critical: >80% → Scale horizontally

2. **Memory Usage**:
   - Normal: <70%
   - Warning: 70-85%
   - Critical: >85% → Scale vertically or optimize

3. **Response Time**:
   - Target: p95 <200ms
   - Warning: p95 200-500ms
   - Critical: p95 >500ms → Scale or optimize

4. **Database Connections**:
   - MySQL max: 151 connections
   - Warning: >100 active connections
   - Solution: Connection pooling or read replicas

5. **WebSocket Connections**:
   - Single server limit: ~10K connections
   - Solution: Multiple WebSocket servers with Redis pub/sub

**SkillSprint Scaling Triggers**:
```
Scenario 1: Live Arena with 5000 users
- CPU spike to 90%
- Action: Add 2 more backend servers

Scenario 2: 1000 concurrent code executions
- Judge queue depth >100
- Action: Add more judge workers

Scenario 3: Database slow queries
- Query time >1s
- Action: Add indexes or read replicas
```

---

## 8. Optimization

### Caching

**What is Caching?**
- Store frequently accessed data in fast storage (RAM)
- Reduce database load and improve response time


**Caching Strategies in SkillSprint**:

#### 1. Cache-Aside (Lazy Loading)
```go
// Get user profile
func GetUserProfile(userID string) (*User, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("user:%s", userID)
    cached, err := redis.Get(cacheKey).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }
    
    // Cache miss - fetch from database
    var user User
    database.DB.First(&user, userID)
    
    // Store in cache (TTL: 1 hour)
    userJSON, _ := json.Marshal(user)
    redis.Set(cacheKey, userJSON, time.Hour)
    
    return &user, nil
}
```

**Use Cases in SkillSprint**:
- User profiles
- Test metadata
- Topic lists
- Leaderboard (short TTL)

#### 2. Write-Through Cache
```go
// Update user profile
func UpdateUserProfile(user *User) error {
    // Update database
    database.DB.Save(user)
    
    // Update cache immediately
    cacheKey := fmt.Sprintf("user:%s", user.ID)
    userJSON, _ := json.Marshal(user)
    redis.Set(cacheKey, userJSON, time.Hour)
    
    return nil
}
```

**Use Cases**:
- User settings
- Test configurations

#### 3. Cache Invalidation
```go
// Delete test (invalidate cache)
func DeleteTest(testID string) error {
    // Delete from database
    database.DB.Delete(&Test{}, testID)
    
    // Invalidate cache
    redis.Del(fmt.Sprintf("test:%s", testID))
    redis.Del("tests:list")  // Invalidate list cache
    
    return nil
}
```

**Cache Invalidation Strategies**:
- **TTL (Time To Live)**: Auto-expire after duration
- **Manual Invalidation**: Delete on update/delete
- **Event-Based**: Invalidate on specific events


### Prompt Caching / Optimizations

**Database Query Optimization**:

1. **N+1 Query Problem**:
```go
// BAD: N+1 queries
func GetTestsWithQuestions() []Test {
    var tests []Test
    database.DB.Find(&tests)
    
    for i := range tests {
        // This runs a query for EACH test (N queries)
        database.DB.Model(&tests[i]).Association("Questions").Find(&tests[i].Questions)
    }
    return tests
}

// GOOD: Single query with JOIN
func GetTestsWithQuestions() []Test {
    var tests []Test
    database.DB.Preload("Questions").Find(&tests)  // 1 query with JOIN
    return tests
}
```

2. **Pagination**:
```go
// Avoid loading all records
func GetTests(page, limit int) []Test {
    var tests []Test
    offset := (page - 1) * limit
    database.DB.Offset(offset).Limit(limit).Find(&tests)
    return tests
}
```

3. **Select Only Needed Columns**:
```go
// BAD: Select all columns
database.DB.Find(&users)

// GOOD: Select specific columns
database.DB.Select("id, username, email").Find(&users)
```

**Connection Pooling**:
```go
// database/db.go
func InitDB() {
    sqlDB, _ := DB.DB()
    
    // Set connection pool settings
    sqlDB.SetMaxIdleConns(10)      // Idle connections
    sqlDB.SetMaxOpenConns(100)     // Max connections
    sqlDB.SetConnMaxLifetime(time.Hour)  // Connection lifetime
}
```


### Offload Async Tasks

**What are Async Tasks?**
- Tasks that don't need immediate response
- Run in background to improve response time

**SkillSprint Async Tasks**:

1. **Email Notifications** (Future):
```go
// Synchronous (BAD - blocks response)
func RegisterUser(user *User) error {
    database.DB.Create(user)
    SendWelcomeEmail(user.Email)  // Blocks for 2-3 seconds
    return nil
}

// Asynchronous (GOOD - immediate response)
func RegisterUser(user *User) error {
    database.DB.Create(user)
    
    // Queue email task
    go func() {
        SendWelcomeEmail(user.Email)
    }()
    
    return nil
}
```

2. **Leaderboard Updates**:
```go
// Current: Immediate update via WebSocket
func SubmitTest(attemptID string) {
    // Calculate score
    score := EvaluateAttempt(attemptID)
    
    // Update database
    database.DB.Model(&TestAttempt{}).Where("id = ?", attemptID).Update("score", score)
    
    // Broadcast to WebSocket (async)
    go func() {
        leaderboard := CalculateLeaderboard(testID)
        LeaderboardHub.Broadcast(leaderboard)
    }()
}
```

3. **Analytics Processing**:
```go
// Track user activity asynchronously
func TrackActivity(userID, action string) {
    go func() {
        database.DB.Create(&ActivityLog{
            UserID: userID,
            Action: action,
            Timestamp: time.Now(),
        })
    }()
}
```

**Message Queue Pattern (Future: Kafka/RabbitMQ)**:
```
User submits test → API returns immediately
                  ↓
            Queue task (Kafka)
                  ↓
        Background worker processes
                  ↓
    Update leaderboard, send notifications
```


### Queue + Worker Model (Idempotency)

**What is Queue + Worker?**
- Queue: Stores tasks to be processed
- Worker: Processes tasks from queue
- Decouples task creation from execution

**SkillSprint Use Case: Code Execution**

```go
// Judge Service with Queue

// 1. Queue Structure
type CodeExecutionTask struct {
    ID           string
    SubmissionID string
    Code         string
    Language     string
    TestCases    []TestCase
    Status       string  // pending, processing, completed
}

// 2. Producer (API Handler)
func RunCode(c *gin.Context) {
    var req CodeRequest
    c.BindJSON(&req)
    
    // Create task
    task := CodeExecutionTask{
        ID:           uuid.New().String(),
        SubmissionID: req.SubmissionID,
        Code:         req.Code,
        Language:     req.Language,
        TestCases:    req.TestCases,
        Status:       "pending",
    }
    
    // Add to queue
    JudgeQueue.Push(task)
    
    // Return immediately
    c.JSON(202, gin.H{
        "task_id": task.ID,
        "status": "queued"
    })
}

// 3. Worker (Background Process)
func JudgeWorker() {
    for {
        task := JudgeQueue.Pop()  // Blocking call
        
        // Process task
        result := ExecuteCode(task.Code, task.Language, task.TestCases)
        
        // Update database
        database.DB.Model(&Submission{}).
            Where("id = ?", task.SubmissionID).
            Update("result", result)
        
        // Notify user via WebSocket
        NotifyUser(task.SubmissionID, result)
    }
}

// 4. Start multiple workers
func main() {
    for i := 0; i < 5; i++ {  // 5 concurrent workers
        go JudgeWorker()
    }
}
```

**Idempotency**:
- **Definition**: Running same operation multiple times produces same result
- **Why Important**: Handle duplicate tasks, retries

```go
// Idempotent operation
func ProcessPayment(paymentID string, amount float64) error {
    // Check if already processed
    var payment Payment
    result := database.DB.Where("id = ?", paymentID).First(&payment)
    
    if result.RowsAffected > 0 && payment.Status == "completed" {
        return nil  // Already processed, skip
    }
    
    // Process payment
    ChargeCard(amount)
    
    // Mark as completed
    database.DB.Model(&payment).Update("status", "completed")
    return nil
}
```


### WebP Images, Indexing

**WebP Images**:
- **What**: Modern image format (Google)
- **Benefits**: 25-35% smaller than JPEG/PNG
- **Use in SkillSprint**: Chat image uploads, profile pictures

```go
// Convert uploaded image to WebP
func ConvertToWebP(inputPath, outputPath string) error {
    cmd := exec.Command("cwebp", 
        "-q", "80",  // Quality
        inputPath, 
        "-o", outputPath)
    return cmd.Run()
}

// Usage in chat upload
func UploadChatFile(c *gin.Context) {
    file, _ := c.FormFile("file")
    
    // Save original
    originalPath := fmt.Sprintf("uploads/chat/%s", file.Filename)
    c.SaveUploadedFile(file, originalPath)
    
    // Convert to WebP if image
    if isImage(file.Filename) {
        webpPath := strings.Replace(originalPath, filepath.Ext(originalPath), ".webp", 1)
        ConvertToWebP(originalPath, webpPath)
        
        // Delete original, use WebP
        os.Remove(originalPath)
        c.JSON(200, gin.H{"url": webpPath})
    }
}
```

---

## 9. Failure Handling

### Retry Mechanism, States

**Exponential Backoff Retry**:

```go
// Retry with exponential backoff
func RetryWithBackoff(operation func() error, maxRetries int) error {
    var err error
    
    for i := 0; i < maxRetries; i++ {
        err = operation()
        if err == nil {
            return nil  // Success
        }
        
        // Wait before retry: 1s, 2s, 4s, 8s...
        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(waitTime)
        
        log.Printf("Retry %d/%d after %v", i+1, maxRetries, waitTime)
    }
    
    return fmt.Errorf("failed after %d retries: %v", maxRetries, err)
}

// Usage: External API call
func CallOpenAI(prompt string) (string, error) {
    return RetryWithBackoff(func() error {
        resp, err := openai.CreateCompletion(prompt)
        if err != nil {
            return err
        }
        return nil
    }, 3)  // Max 3 retries
}
```


**State Management in SkillSprint**:

```go
// Test Attempt States
const (
    AttemptStatusNotStarted  = "not_started"
    AttemptStatusInProgress  = "in_progress"
    AttemptStatusSubmitted   = "submitted"
    AttemptStatusEvaluated   = "evaluated"
    AttemptStatusExpired     = "expired"
)

// State transitions
func (a *TestAttempt) Submit() error {
    if a.Status != AttemptStatusInProgress {
        return errors.New("can only submit in-progress attempts")
    }
    
    a.Status = AttemptStatusSubmitted
    a.SubmittedAt = time.Now()
    return database.DB.Save(a).Error
}

// Auto-expire attempts
func AutoExpireAttempts() {
    ticker := time.NewTicker(1 * time.Minute)
    
    for range ticker.C {
        database.DB.Model(&TestAttempt{}).
            Where("status = ? AND started_at < ?", 
                AttemptStatusInProgress, 
                time.Now().Add(-2*time.Hour)).
            Update("status", AttemptStatusExpired)
    }
}
```

### Logging Mechanisms

**Structured Logging in SkillSprint**:

```go
// Use structured logging (not fmt.Println)
import "log"

// Log levels
func LogInfo(message string, fields map[string]interface{}) {
    log.Printf("[INFO] %s %v", message, fields)
}

func LogError(message string, err error, fields map[string]interface{}) {
    log.Printf("[ERROR] %s: %v %v", message, err, fields)
}

// Usage
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    c.BindJSON(&req)
    
    LogInfo("Login attempt", map[string]interface{}{
        "email": req.Email,
        "ip": c.ClientIP(),
    })
    
    user, err := AuthService.Login(req.Email, req.Password)
    if err != nil {
        LogError("Login failed", err, map[string]interface{}{
            "email": req.Email,
        })
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }
    
    LogInfo("Login successful", map[string]interface{}{
        "user_id": user.ID,
        "email": user.Email,
    })
}
```

**Log Aggregation (Future: ELK Stack)**:
```
Application Logs → Filebeat → Logstash → Elasticsearch → Kibana
```


---

## 10. Bottlenecks

### Downtime, Cold Start

**Downtime Scenarios in SkillSprint**:

1. **Database Downtime**:
```
Problem: MySQL server crashes
Impact: All API requests fail
Solution:
- Database replication (master-slave)
- Automatic failover
- Health checks

// Health check endpoint
func HealthCheck(c *gin.Context) {
    // Check database
    err := database.DB.Exec("SELECT 1").Error
    if err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "database": "down"})
        return
    }
    
    c.JSON(200, gin.H{"status": "healthy"})
}
```

2. **Server Downtime**:
```
Problem: Backend server crashes
Impact: Users can't access platform
Solution:
- Multiple backend servers
- Load balancer health checks
- Auto-restart (systemd, Docker)
```

**Cold Start**:
- **Definition**: Delay when starting application from scratch
- **Causes**: Loading dependencies, connecting to database, warming caches

```go
// Optimize cold start
func main() {
    // 1. Connect to database (parallel)
    dbReady := make(chan bool)
    go func() {
        database.InitDB()
        dbReady <- true
    }()
    
    // 2. Initialize services
    router := gin.Default()
    setupRoutes(router)
    
    // 3. Wait for database
    <-dbReady
    
    // 4. Warm cache (optional)
    go warmCache()
    
    // 5. Start server
    router.Run(":8080")
}

func warmCache() {
    // Pre-load frequently accessed data
    var tests []Test
    database.DB.Find(&tests)
    
    for _, test := range tests {
        cacheKey := fmt.Sprintf("test:%d", test.ID)
        testJSON, _ := json.Marshal(test)
        redis.Set(cacheKey, testJSON, time.Hour)
    }
}
```


### Additional Interview Topics

#### Stateless APIs (Detailed)

**Why Stateless?**
```
Stateful (BAD):
User logs in → Server stores session in memory
User makes request → Must go to SAME server

Stateless (GOOD):
User logs in → Server returns JWT token
User makes request → ANY server can validate token
```

**JWT Token Structure**:
```
Header.Payload.Signature

Header: {"alg": "HS256", "typ": "JWT"}
Payload: {"user_id": 123, "role": "user", "exp": 1234567890}
Signature: HMACSHA256(base64(header) + "." + base64(payload), secret)
```

**Implementation in SkillSprint**:
```go
// Generate token
func GenerateToken(userID uint, role string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "role":    role,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// Validate token
func ValidateToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
}
```

#### API Design Best Practices

**RESTful Endpoints in SkillSprint**:
```
GET    /api/tests              - List all tests
GET    /api/tests/:id          - Get specific test
POST   /api/tests              - Create test (admin)
PUT    /api/tests/:id          - Update test (admin)
DELETE /api/tests/:id          - Delete test (admin)

POST   /api/tests/:id/join     - Join test (create attempt)
POST   /api/attempts/:id/submit - Submit attempt
GET    /api/attempts/:id/result - Get result

GET    /api/leaderboard/:testId - Get leaderboard
```

**HTTP Status Codes**:
```
200 OK - Success
201 Created - Resource created
400 Bad Request - Invalid input
401 Unauthorized - No/invalid token
403 Forbidden - Valid token but no permission
404 Not Found - Resource doesn't exist
500 Internal Server Error - Server error
```


#### Database Optimization Techniques

**1. Query Optimization**:
```sql
-- BAD: Full table scan
SELECT * FROM test_attempts WHERE user_id = 123;

-- GOOD: Use index
CREATE INDEX idx_user_id ON test_attempts(user_id);
SELECT * FROM test_attempts WHERE user_id = 123;

-- EXPLAIN to analyze query
EXPLAIN SELECT * FROM test_attempts WHERE user_id = 123;
```

**2. Denormalization for Performance**:
```sql
-- Normalized (multiple JOINs)
SELECT u.username, COUNT(ta.id) as attempt_count
FROM users u
LEFT JOIN test_attempts ta ON u.id = ta.user_id
GROUP BY u.id;

-- Denormalized (add attempt_count to users table)
ALTER TABLE users ADD COLUMN attempt_count INT DEFAULT 0;

-- Update on each attempt
UPDATE users SET attempt_count = attempt_count + 1 WHERE id = ?;

-- Fast query
SELECT username, attempt_count FROM users WHERE id = 123;
```

**3. Database Sharding**:
```
Single Database:
All users → MySQL Server 1

Sharded (by user_id):
Users 1-1000000   → MySQL Shard 1
Users 1000001-2000000 → MySQL Shard 2
Users 2000001-3000000 → MySQL Shard 3

// Shard selection
func GetShardForUser(userID int) *gorm.DB {
    shardNum := userID % 3
    return shards[shardNum]
}
```

#### WebSocket Architecture

**Hub Pattern in SkillSprint**:
```
                    ┌─ Client 1 (readPump, writePump)
                    ├─ Client 2 (readPump, writePump)
Hub (central) ──────┤─ Client 3 (readPump, writePump)
                    ├─ Client 4 (readPump, writePump)
                    └─ Client 5 (readPump, writePump)

Hub manages:
- Register/unregister clients
- Broadcast messages to all
- Handle disconnections
```

**Goroutines per Client**:
```go
// Each client has 2 goroutines
type Client struct {
    conn *websocket.Conn
    send chan []byte
}

// Goroutine 1: Read from WebSocket
func (c *Client) readPump() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        hub.broadcast <- message
    }
}

// Goroutine 2: Write to WebSocket
func (c *Client) writePump() {
    for message := range c.send {
        c.conn.WriteMessage(websocket.TextMessage, message)
    }
}
```


#### Security Best Practices

**1. Password Security**:
```go
import "golang.org/x/crypto/bcrypt"

// Hash password (signup)
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

// Verify password (login)
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

**2. SQL Injection Prevention**:
```go
// BAD: SQL injection vulnerable
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
database.DB.Raw(query).Scan(&user)

// GOOD: Parameterized query
database.DB.Where("email = ?", email).First(&user)
```

**3. CORS Configuration**:
```go
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()
    
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    }))
}
```

**4. Rate Limiting**:
```go
// Prevent abuse
var limiter = rate.NewLimiter(10, 20)  // 10 req/sec, burst 20

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Too many requests"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### Monitoring & Observability

**Key Metrics to Track**:
```go
// 1. Request metrics
type Metrics struct {
    TotalRequests   int64
    FailedRequests  int64
    AvgResponseTime float64
}

// 2. Database metrics
- Query execution time
- Connection pool usage
- Slow query log

// 3. WebSocket metrics
- Active connections
- Messages per second
- Connection errors

// 4. Business metrics
- Active users
- Tests taken per day
- Average test score
```

**Health Check Endpoint**:
```go
func HealthCheck(c *gin.Context) {
    health := gin.H{
        "status": "healthy",
        "timestamp": time.Now(),
        "checks": gin.H{
            "database": checkDatabase(),
            "redis": checkRedis(),
            "disk_space": checkDiskSpace(),
        },
    }
    
    c.JSON(200, health)
}
```


---

## Summary: SkillSprint Architecture at a Glance

### System Overview
```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                              │
│  Next.js Frontend (React + TypeScript + Tailwind)           │
│  - Server-side rendering                                     │
│  - WebSocket clients (chat, leaderboard)                     │
│  - JWT token storage                                         │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS + WebSocket
┌────────────────────────▼────────────────────────────────────┐
│                   APPLICATION LAYER                          │
│  Go Backend (Gin Framework)                                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Services:                                             │  │
│  │ • Auth (JWT, OAuth)                                   │  │
│  │ • Arena (test management)                             │  │
│  │ • Judge (code execution)                              │  │
│  │ • Chat (WebSocket hub)                                │  │
│  │ • Leaderboard (real-time rankings)                    │  │
│  │ • Training (adaptive learning)                        │  │
│  │ • AI (OpenAI integration)                             │  │
│  │ • Admin (content management)                          │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │ GORM ORM
┌────────────────────────▼────────────────────────────────────┐
│                     DATA LAYER                               │
│  MySQL Database                                              │
│  • Users, Tests, Questions, Attempts                         │
│  • Chat messages                                             │
│  • Analytics data                                            │
│  • Indexes on frequently queried columns                     │
└─────────────────────────────────────────────────────────────┘
```

### Key Technologies & Their Purpose

| Technology | Purpose | Why Chosen |
|-----------|---------|------------|
| **Next.js** | Frontend framework | SSR, routing, performance |
| **Go** | Backend language | Fast, concurrent, compiled |
| **Gin** | Web framework | Fast routing, middleware |
| **MySQL** | Database | ACID, relational data |
| **GORM** | ORM | Type-safe queries |
| **JWT** | Authentication | Stateless, scalable |
| **WebSocket** | Real-time | Chat, leaderboard updates |
| **Docker** | Code execution | Sandboxed, isolated |
| **OpenAI** | AI generation | Smart test creation |

### Data Flow Examples

**User Takes Test**:
```
1. POST /api/arena/tests/:id/join → Create attempt
2. GET questions → Display test
3. POST /api/submissions/run → Test code
4. POST /api/attempts/:id/submit → Submit test
5. Backend evaluates → Calculate score
6. WebSocket broadcast → Update leaderboard
7. GET /api/results/:id → Show results
```

**Real-Time Chat**:
```
1. WebSocket connect → Register in hub
2. User sends message → readPump receives
3. Save to database → Persist history
4. Broadcast to all clients → writePump sends
5. All users see message instantly
```

### Scaling Strategy

**Current**: Monolithic (single server)
**Phase 1**: Horizontal scaling (multiple backend servers)
**Phase 2**: Database replication (read replicas)
**Phase 3**: Microservices (separate services)
**Phase 4**: Caching layer (Redis)

### Performance Optimizations

1. **Database**: Indexes, connection pooling, query optimization
2. **Caching**: Redis for frequently accessed data
3. **Async**: Background workers for heavy tasks
4. **WebSocket**: Efficient hub pattern with goroutines
5. **Code Execution**: Queue-based with multiple workers

---

## Interview Preparation Tips

### When Asked About SkillSprint:

1. **Start with Problem Statement**: "SkillSprint solves the problem of..."
2. **Explain Architecture**: "We use a monolithic architecture with Go backend and Next.js frontend..."
3. **Highlight Key Features**: "Real-time chat, live arenas, code execution, AI generation..."
4. **Discuss Scaling**: "Currently single server, planning to scale horizontally..."
5. **Mention Challenges**: "Handling concurrent code execution, real-time updates, security..."

### Common Interview Questions:

**Q: How do you handle concurrent users?**
A: "We use Go's goroutines for WebSocket connections. Each client has 2 goroutines (read/write). The hub pattern manages all clients efficiently."

**Q: How do you ensure security?**
A: "JWT for authentication, bcrypt for passwords, parameterized queries to prevent SQL injection, Docker for code sandboxing, CORS configuration."

**Q: How would you scale to 1 million users?**
A: "Horizontal scaling with load balancer, database replication, Redis caching, microservices architecture, CDN for static assets, message queue for async tasks."

**Q: Explain your database design**
A: "Relational schema with users, tests, questions, attempts. Indexes on frequently queried columns. Foreign keys for referential integrity."

**Q: How does real-time chat work?**
A: "WebSocket connections managed by a hub. When user sends message, it's saved to database and broadcast to all connected clients via goroutines."

---

**Document Created**: For SkillSprint Interview Preparation
**Last Updated**: 2026
**Author**: SkillSprint Development Team


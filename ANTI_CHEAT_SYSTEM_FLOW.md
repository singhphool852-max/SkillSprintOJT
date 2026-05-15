# Anti-Cheat System Flow - Complete Process

## 🎯 Overview

The anti-cheat system monitors student behavior during tests and automatically logs violations, displays warnings, and triggers auto-submission after 3 violations.

---

## 📊 Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    STUDENT JOINS TEST                            │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 1: HOOK INITIALIZATION (useAntiCheat.ts)                  │
├─────────────────────────────────────────────────────────────────┤
│  • Hook is activated when isTestActive = true                   │
│  • Requests fullscreen mode                                     │
│  • Waits 1.5 seconds grace period                               │
│  • Sets isArmedRef = true (monitoring starts)                   │
│  • Registers 5 event listeners:                                 │
│    1. fullscreenchange → detects fullscreen exit                │
│    2. visibilitychange → detects tab switch                     │
│    3. blur → detects Alt+Tab / window switch                    │
│    4. keydown → blocks F12, Ctrl+Shift+I, etc.                  │
│    5. contextmenu → blocks right-click                          │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 2: EVENT TRIGGERED (Browser Event)                        │
├─────────────────────────────────────────────────────────────────┤
│  Student performs suspicious action:                            │
│  • Presses Esc or F11 (exits fullscreen)                        │
│  • Switches to another tab (Ctrl+Tab)                           │
│  • Alt+Tabs to another window                                   │
│  • Presses F12 or Ctrl+Shift+I (DevTools)                       │
│  • Right-clicks (context menu)                                  │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 3: VIOLATION DETECTION (useAntiCheat.ts)                  │
├─────────────────────────────────────────────────────────────────┤
│  • Event listener fires                                         │
│  • Checks if isArmedRef = true (monitoring active)              │
│  • Debounces: ignores if < 500ms since last violation           │
│  • Increments violationCountRef.current                         │
│  • Calls addViolation(type)                                     │
│  • Logs to console: "[AntiCheat] Violation 1/3: TAB_SWITCH"     │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 4: CALLBACK TO PARENT (arena/page.tsx)                    │
├─────────────────────────────────────────────────────────────────┤
│  • Hook calls onViolation(type, count)                          │
│  • Parent component receives callback                           │
│  • handleViolation() executes:                                  │
│    1. Updates violationCount state                              │
│    2. Shows toast notification on screen                        │
│    3. Sends API request to backend                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 5: TOAST NOTIFICATION (Frontend UI)                       │
├─────────────────────────────────────────────────────────────────┤
│  • Toast appears on screen:                                     │
│    "⚠️ Violation 1/3: TAB SWITCH"                               │
│  • Styled with yellow/red warning colors                        │
│  • Auto-dismisses after 4 seconds                               │
│  • If fullscreen exit: shows "Return to Fullscreen" button      │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 6: API REQUEST TO BACKEND                                 │
├─────────────────────────────────────────────────────────────────┤
│  POST /api/arena/violations                                     │
│  Headers: { Authorization: "Bearer <token>" }                   │
│  Body: {                                                        │
│    "attemptId": "uuid-123",                                     │
│    "testId": "uuid-456",                                        │
│    "violationType": "tab_switch",                               │
│    "remainingTime": 1800                                        │
│  }                                                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 7: BACKEND HANDLER (violation_handler.go)                 │
├─────────────────────────────────────────────────────────────────┤
│  LogViolation() function executes:                              │
│  1. Extracts userID from JWT token                              │
│  2. Validates request body (attemptId, testId, violationType)   │
│  3. Verifies attempt belongs to user                            │
│  4. Checks attempt not already submitted                        │
│  5. Creates violation log entry                                 │
│  6. Increments violationCount on test_attempts table            │
│  7. Checks if count >= 3 (auto-submit threshold)                │
│  8. Returns response with updated count                         │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 8: DATABASE STORAGE                                       │
├─────────────────────────────────────────────────────────────────┤
│  TWO TABLES UPDATED:                                            │
│                                                                 │
│  A) test_violations table (new row inserted):                   │
│     ┌──────────────────────────────────────────────┐           │
│     │ id: uuid-789                                 │           │
│     │ attemptId: uuid-123                          │           │
│     │ userId: uuid-user-1                          │           │
│     │ testId: uuid-456                             │           │
│     │ violationType: "tab_switch"                  │           │
│     │ timestamp: 2026-05-15 14:30:00               │           │
│     │ remainingTime: 1800                          │           │
│     │ createdAt: 2026-05-15 14:30:00               │           │
│     └──────────────────────────────────────────────┘           │
│                                                                 │
│  B) test_attempts table (violationCount incremented):           │
│     ┌──────────────────────────────────────────────┐           │
│     │ id: uuid-123                                 │           │
│     │ userId: uuid-user-1                          │           │
│     │ testId: uuid-456                             │           │
│     │ violationCount: 1 → 2 → 3                    │           │
│     │ isAutoSubmitted: false → true (at 3)         │           │
│     │ submittedAt: NULL → timestamp (at 3)         │           │
│     └──────────────────────────────────────────────┘           │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 9: BACKEND RESPONSE                                       │
├─────────────────────────────────────────────────────────────────┤
│  Response: {                                                    │
│    "violationCount": 2,                                         │
│    "autoSubmit": false                                          │
│  }                                                              │
│                                                                 │
│  OR (if 3rd violation):                                         │
│  Response: {                                                    │
│    "violationCount": 3,                                         │
│    "autoSubmit": true                                           │
│  }                                                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 10: AUTO-SUBMIT TRIGGER (if count >= 3)                   │
├─────────────────────────────────────────────────────────────────┤
│  • Hook checks: violationCountRef.current >= maxViolations      │
│  • Calls onAutoSubmit() callback                                │
│  • Parent component triggers test submission                    │
│  • Test is submitted with isAutoSubmitted = true                │
│  • Student sees: "Test auto-submitted due to violations"        │
│  • Hook calls cleanup() to stop monitoring                      │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 11: ADMIN ANALYTICS VIEW                                  │
├─────────────────────────────────────────────────────────────────┤
│  Admin navigates to: /admin/analytics                           │
│  • Frontend fetches: GET /api/admin/analytics/mistakes          │
│  • Backend queries test_violations table with JOINs             │
│  • Returns aggregated violation data                            │
│  • Frontend displays table with:                                │
│    - Student name and email                                     │
│    - Test title                                                 │
│    - Total violation count (color-coded badge)                  │
│    - Fullscreen exits count                                     │
│    - Tab switches count                                         │
│    - Last violation timestamp                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Technical Components

### 1. Frontend Hook: `useAntiCheat.ts`

**Location:** `frontend/hooks/useAntiCheat.ts`

**Purpose:** Monitors browser events and detects violations

**Key Features:**
- Uses `useRef` to avoid stale closure bugs
- Debounces violations (500ms minimum between)
- Grace period (1.5s) before arming to avoid false triggers
- Auto re-enters fullscreen when possible
- Blocks DevTools keys (F12, Ctrl+Shift+I, etc.)
- Blocks right-click context menu

**Event Listeners:**
```typescript
1. fullscreenchange → Detects Esc/F11 press
2. visibilitychange → Detects Ctrl+Tab (tab switch)
3. blur → Detects Alt+Tab (window switch)
4. keydown → Blocks F12, Ctrl+Shift+I, Ctrl+U
5. contextmenu → Blocks right-click
```

**Violation Types:**
- `fullscreen_exit` - Student exits fullscreen mode
- `tab_switch` - Student switches to another tab
- `window_blur` - Student Alt+Tabs to another window
- `blocked_key` - Student tries to open DevTools

---

### 2. Parent Component: `arena/page.tsx`

**Location:** `frontend/app/arena/page.tsx`

**Purpose:** Integrates anti-cheat hook and handles callbacks

**Key Functions:**

**handleViolation(type, count):**
```typescript
1. Updates violationCount state
2. Shows toast: "⚠️ Violation 1/3: TAB SWITCH"
3. Sends POST request to backend
4. Fire-and-forget (doesn't wait for response)
```

**handleAutoSubmit():**
```typescript
1. Called when count >= 3
2. Triggers test submission
3. Marks as isAutoSubmitted = true
```

**handleShowFullscreenWarning(show):**
```typescript
1. Shows/hides "Return to Fullscreen" button
2. Button appears when auto re-enter fails
3. Clicking button requests fullscreen again
```

---

### 3. Backend Handler: `violation_handler.go`

**Location:** `go-backend/handlers/violation_handler.go`

**Endpoint:** `POST /api/arena/violations`

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "attemptId": "uuid-123",
  "testId": "uuid-456",
  "violationType": "tab_switch",
  "remainingTime": 1800
}
```

**Process:**
1. Extract userID from JWT token
2. Validate request body
3. Verify attempt belongs to user
4. Check attempt not already submitted
5. Create violation log in `test_violations` table
6. Increment `violationCount` in `test_attempts` table
7. Check if count >= 3 for auto-submit
8. Return updated count and autoSubmit flag

**Response:**
```json
{
  "violationCount": 2,
  "autoSubmit": false
}
```

---

### 4. Database Tables

#### A) `test_violations` Table

**Purpose:** Stores individual violation events (audit log)

**Schema:**
```sql
CREATE TABLE test_violations (
  id VARCHAR(191) PRIMARY KEY,
  attemptId VARCHAR(191) NOT NULL,
  userId VARCHAR(191) NOT NULL,
  testId VARCHAR(191) NOT NULL,
  violationType VARCHAR(50) NOT NULL,
  timestamp DATETIME NOT NULL,
  remainingTime INT,
  createdAt DATETIME NOT NULL,
  
  INDEX idx_attempt (attemptId),
  INDEX idx_user (userId),
  INDEX idx_test (testId)
);
```

**Example Row:**
```
id: "550e8400-e29b-41d4-a716-446655440000"
attemptId: "123e4567-e89b-12d3-a456-426614174000"
userId: "user-uuid-1"
testId: "test-uuid-1"
violationType: "tab_switch"
timestamp: "2026-05-15 14:30:00"
remainingTime: 1800
createdAt: "2026-05-15 14:30:00"
```

#### B) `test_attempts` Table (violationCount column)

**Purpose:** Tracks total violation count per attempt

**Relevant Columns:**
```sql
violationCount INT DEFAULT 0,
isAutoSubmitted BOOLEAN DEFAULT false,
submittedAt DATETIME NULL
```

**Update Process:**
```sql
-- On each violation
UPDATE test_attempts 
SET violationCount = violationCount + 1 
WHERE id = 'attempt-uuid';

-- On 3rd violation (auto-submit)
UPDATE test_attempts 
SET 
  violationCount = 3,
  isAutoSubmitted = true,
  submittedAt = NOW()
WHERE id = 'attempt-uuid';
```

---

### 5. Admin Analytics Query

**Location:** `go-backend/handlers/training_adaptive.go`

**Endpoint:** `GET /api/admin/analytics/mistakes`

**SQL Query:**
```sql
SELECT 
    u.name as user_name,
    u.email as user_email,
    t.title as test_title,
    COUNT(v.id) as violation_count,
    SUM(CASE WHEN v.violationType = 'fullscreen_exit' THEN 1 ELSE 0 END) as fullscreen_exits,
    SUM(CASE WHEN v.violationType = 'tab_switch' THEN 1 ELSE 0 END) as tab_switches,
    MAX(v.timestamp) as last_violation
FROM test_violations v
JOIN users u ON u.id = v.userId
JOIN tests t ON t.id = v.testId
GROUP BY v.userId, v.testId, u.name, u.email, t.title
ORDER BY violation_count DESC
LIMIT 50
```

**Response:**
```json
{
  "mistakes": [
    {
      "userName": "John Doe",
      "userEmail": "john@example.com",
      "testTitle": "Data Structures Final",
      "violationCount": 5,
      "fullscreenExits": 3,
      "tabSwitches": 2,
      "lastViolation": "2026-05-15T14:30:00Z"
    }
  ],
  "total": 1
}
```

---

## 🎨 UI Elements

### 1. Violation Toast (During Test)

**Appearance:**
```
┌────────────────────────────────────┐
│ ⚠️ Violation 1/3: TAB SWITCH       │
└────────────────────────────────────┘
```

**Styling:**
- Yellow/orange background
- Bold text
- Auto-dismisses after 4 seconds
- Appears at top of screen

### 2. Fullscreen Warning Button

**Appears when:** Auto re-enter fullscreen fails

**Appearance:**
```
┌──────────────────────────────────────────────┐
│ ⚠️ FULLSCREEN REQUIRED                       │
│                                              │
│ You have exited fullscreen mode.            │
│ Further tab switches will count as          │
│ additional violations.                       │
│                                              │
│ [ Return to Fullscreen ]                    │
└──────────────────────────────────────────────┘
```

**Behavior:**
- Blocks test interface
- Clicking button requests fullscreen
- Violation already counted
- Button disappears when fullscreen restored

### 3. Auto-Submit Message

**Appears when:** 3rd violation triggers auto-submit

**Appearance:**
```
┌──────────────────────────────────────────────┐
│ ⚠️ TEST AUTO-SUBMITTED                       │
│                                              │
│ Your test has been automatically submitted   │
│ due to 3 anti-cheat violations.              │
│                                              │
│ [ View Results ]                             │
└──────────────────────────────────────────────┘
```

### 4. Admin Analytics Table

**Location:** `/admin/analytics`

**Appearance:**
```
┌─ ANTI-CHEAT VIOLATIONS ────────────────────────────────────────┐
│ STUDENT    │ TEST      │ WARNINGS │ EXITS │ SWITCHES │ LAST   │
├────────────┼───────────┼──────────┼───────┼──────────┼────────┤
│ John Doe   │ DS Final  │   [3]    │   2   │    1     │ 2:30pm │
│ john@...   │           │   RED    │   🟠  │    🟡    │        │
├────────────┼───────────┼──────────┼───────┼──────────┼────────┤
│ Jane Smith │ Algo Test │   [2]    │   1   │    1     │ 1:15pm │
│ jane@...   │           │  YELLOW  │   🟠  │    🟡    │        │
└────────────┴───────────┴──────────┴───────┴──────────┴────────┘
```

**Color Coding:**
- 1 violation → Gray badge
- 2 violations → Yellow badge
- 3+ violations → Red badge

---

## 📍 Where Logs Are Stored

### Database Location

**Database File:** `dev.db` (SQLite)

**Tables:**
1. **`test_violations`** - Individual violation events
2. **`test_attempts`** - Aggregated violation count

### File System Location

**Development:**
```
/path/to/project/dev.db
```

**Production:**
```
/var/lib/skillsprint/production.db
```

### Accessing Logs

**Via SQLite CLI:**
```bash
sqlite3 dev.db

# View all violations
SELECT * FROM test_violations ORDER BY timestamp DESC LIMIT 10;

# View violations by user
SELECT u.name, v.violationType, v.timestamp 
FROM test_violations v
JOIN users u ON u.id = v.userId
WHERE u.email = 'student@example.com';

# View violation summary
SELECT 
  u.name,
  COUNT(*) as total_violations,
  SUM(CASE WHEN violationType = 'tab_switch' THEN 1 ELSE 0 END) as tab_switches,
  SUM(CASE WHEN violationType = 'fullscreen_exit' THEN 1 ELSE 0 END) as fullscreen_exits
FROM test_violations v
JOIN users u ON u.id = v.userId
GROUP BY u.id;
```

**Via Admin Panel:**
```
Navigate to: http://localhost:3000/admin/analytics
Scroll to: "ANTI-CHEAT VIOLATIONS" section
```

**Via API:**
```bash
curl -X GET http://localhost:8080/api/admin/analytics/mistakes \
  -H "Cookie: token=YOUR_ADMIN_TOKEN" \
  | jq
```

---

## 🔄 Complete Timeline Example

**Student: John Doe**  
**Test: Data Structures Final**  
**Start Time: 2:00 PM**

| Time  | Event | Action | Database | UI |
|-------|-------|--------|----------|-----|
| 2:00 PM | Test starts | Hook armed | `violationCount = 0` | Fullscreen mode |
| 2:15 PM | Presses Esc | Fullscreen exit detected | `violationCount = 1`<br>New row in `test_violations` | Toast: "⚠️ Violation 1/3"<br>Fullscreen button shown |
| 2:16 PM | Clicks button | Returns to fullscreen | No change | Button hidden |
| 2:30 PM | Alt+Tabs | Window blur detected | `violationCount = 2`<br>New row in `test_violations` | Toast: "⚠️ Violation 2/3" |
| 2:45 PM | Ctrl+Tab | Tab switch detected | `violationCount = 3`<br>`isAutoSubmitted = true`<br>`submittedAt = NOW()`<br>New row in `test_violations` | Toast: "⚠️ Violation 3/3"<br>Test auto-submitted<br>Redirect to results |
| 3:00 PM | Admin checks | Views analytics | Query aggregates data | Table shows:<br>John Doe: 3 violations<br>1 fullscreen exit<br>2 tab switches |

---

## 🛡️ Security Features

1. **JWT Authentication** - All API calls require valid token
2. **User Verification** - Backend verifies attempt belongs to user
3. **Debouncing** - Prevents spam violations (500ms minimum)
4. **Grace Period** - 1.5s delay before arming to avoid false triggers
5. **Audit Trail** - Every violation logged with timestamp
6. **Immutable Logs** - Violations cannot be deleted by students
7. **Admin-Only Access** - Only admins can view violation analytics

---

## 🎓 Summary

The anti-cheat system is a **multi-layered monitoring system** that:

1. **Hooks** into browser events via React custom hook
2. **Detects** suspicious behavior (fullscreen exit, tab switch, etc.)
3. **Logs** violations to backend API in real-time
4. **Stores** data in two database tables (violations + attempts)
5. **Displays** warnings to students via toast notifications
6. **Auto-submits** test after 3 violations
7. **Reports** violation analytics to admins

All logs are stored in the **SQLite database** (`dev.db`) in the **`test_violations`** table and can be viewed via the **Admin Analytics** page at `/admin/analytics`.

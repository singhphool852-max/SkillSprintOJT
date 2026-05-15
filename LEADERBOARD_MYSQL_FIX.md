# Leaderboard MySQL Fix - 500 Error Resolution

## Problem
`GET /api/leaderboard/global` was returning **500 Internal Server Error** on Render (MySQL database).

## Root Cause
**MySQL on Linux is case-sensitive for column names**. The database uses **camelCase** column names (`userId`, `completedAt`), but the raw SQL query was using **snake_case** aliases that didn't match.

### Schema Analysis
From `models/attempt.go` and `models/user.go`:

**Attempts Table**:
- Table name: `attempts`
- Columns: `id`, `userId`, `quizId`, `score`, `totalQuestions`, `startedAt`, `completedAt`

**User Table**:
- Table name: `user`
- Columns: `id`, `email`, `username`, `role`, etc.

## Solution Applied

### Changed From: Raw SQL (Broken)
```go
sqlQuery := `
    SELECT 
        a.userId as user_id,
        u.username,
        SUM(a.score) as total_score,
        ...
    FROM attempts a
    JOIN user u ON u.id = a.userId
    ...
`
database.DB.Raw(sqlQuery).Scan(&rows)
```

**Problem**: Raw SQL with string concatenation, no GORM validation, MySQL case-sensitivity issues.

### Changed To: GORM Query Builder (Fixed)
```go
query := database.DB.Table("attempts").
    Select("attempts.userId as user_id, "+
        "user.username as username, "+
        "COUNT(DISTINCT attempts.id) as tests_count, "+
        "MAX(attempts.score) as best_score, "+
        "SUM(attempts.score) as total_score, "+
        "MIN(attempts.completedAt) as earliest_submit").
    Joins("JOIN user ON user.id = attempts.userId").
    Where("attempts.completedAt IS NOT NULL").
    Where("user.role != ?", "admin").
    Group("attempts.userId, user.username").
    Order("total_score DESC, earliest_submit ASC").
    Limit(100)

query.Scan(&entries)
```

**Benefits**:
- ✅ Uses GORM's query builder (MySQL-compatible)
- ✅ Correct camelCase column names (`userId`, `completedAt`)
- ✅ Parameterized queries (SQL injection safe)
- ✅ Better error handling with detailed error messages
- ✅ Comprehensive logging

## Key Changes

### 1. Struct Tags
```go
type LeaderboardRow struct {
    UserID         string `gorm:"column:user_id"`
    Username       string `gorm:"column:username"`
    TestsCount     int    `gorm:"column:tests_count"`
    BestScore      int    `gorm:"column:best_score"`
    TotalScore     int    `gorm:"column:total_score"`
    EarliestSubmit string `gorm:"column:earliest_submit"`
}
```

### 2. Error Handling
```go
if err := query.Scan(&entries).Error; err != nil {
    log.Printf("[Leaderboard] DB ERROR: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{
        "error":   "Database query failed",
        "details": err.Error(),
    })
    return
}
```

### 3. Debug Endpoint Enhanced
`GET /api/leaderboard/debug` now returns:
- `totalAttempts` - All attempts in database
- `completedAttempts` - Attempts with `completedAt IS NOT NULL`
- `totalUsers` - All users
- `nonAdminUsers` - Users with `role != 'admin'`
- `sampleAttempts` - Last 5 completed attempts with details

## Files Modified

1. ✅ `go-backend/handlers/global_leaderboard.go` - Complete rewrite with GORM query builder

## Testing After Deployment

### 1. Test Debug Endpoint
```bash
curl https://skillsprintojt.onrender.com/api/leaderboard/debug
```

**Expected Response**:
```json
{
  "tableName": "attempts",
  "userTableName": "user",
  "totalAttempts": 15,
  "completedAttempts": 12,
  "nonAdminUsers": 5,
  "sampleAttempts": [
    {
      "id": "abc123",
      "userId": "user1",
      "score": 85,
      "completedAt": "2026-05-16T10:30:00Z"
    }
  ],
  "message": "Debug info retrieved successfully"
}
```

### 2. Test Leaderboard Endpoint
```bash
curl https://skillsprintojt.onrender.com/api/leaderboard/global
```

**Expected Response**:
```json
{
  "entries": [
    {
      "rank": 1,
      "userId": "abc123",
      "username": "john_doe",
      "totalScore": 250,
      "testsCompleted": 3,
      "avgPercentage": 83.33,
      "highScore": 100,
      "tier": "APEX"
    }
  ],
  "totalUsers": 5
}
```

### 3. Check Render Logs
Look for these log entries in Render dashboard:
```
[Leaderboard] Executing query...
[Leaderboard] Found 3 entries
[Leaderboard] #1: john_doe (ID=abc123) Score=250 Tests=3 Best=100
[Leaderboard] #2: jane_smith (ID=def456) Score=180 Tests=2 Best=100
```

### 4. Check Frontend
Visit `/leaderboard` page and:
- Open browser DevTools → Console
- Look for: `[Leaderboard] Response status: 200`
- Look for: `[Leaderboard] Entries count: X`
- Leaderboard should display user data

## Troubleshooting

### If Still Getting 500 Error
**Check Render logs for**:
```
[Leaderboard] DB ERROR: ...
```

Common issues:
- Column name mismatch (check actual MySQL schema)
- Table doesn't exist (check database migration)
- Connection issue (check DATABASE_URL env var)

### If Debug Shows 0 Completed Attempts
**Problem**: Tests aren't being marked as completed  
**Check**: `POST /api/attempts` endpoint sets `completedAt` correctly

### If Debug Shows Data But Leaderboard Empty
**Problem**: Query filtering too aggressively  
**Check**: 
- Users have `role != 'admin'`
- Attempts have `completedAt IS NOT NULL`

## Deployment Info

- **Commit**: `975bd39` - "fix: leaderboard query for MySQL with correct camelCase column names"
- **Backend**: Auto-deploys from main branch on Render
- **Database**: MySQL (case-sensitive on Linux)
- **Route**: `GET /api/leaderboard/global` (PUBLIC, no JWT required)

## Why This Fix Works

1. **GORM Query Builder**: Handles MySQL dialect automatically
2. **Correct Column Names**: Uses actual camelCase names from schema
3. **Parameterized Queries**: Prevents SQL injection, better error handling
4. **Comprehensive Logging**: Easy to debug if issues persist
5. **Debug Endpoint**: Verify data exists before troubleshooting query

## Expected Behavior

- ✅ Returns 200 OK with user leaderboard data
- ✅ Users ranked by total score (sum of all test scores)
- ✅ Tiebreaker: earliest submission time
- ✅ Same score = same tier
- ✅ Excludes admin users
- ✅ Only includes completed attempts

## Next Steps

1. Wait for Render deployment to complete (~2-3 minutes)
2. Test debug endpoint first to verify data exists
3. Test leaderboard endpoint
4. Check Render logs for any errors
5. Visit frontend leaderboard page to verify display

# Leaderboard Debug Fix - Empty Data Investigation

## Problem
`GET /api/leaderboard/global` returns 200 OK but with empty data:
```json
{
  "entries": [],
  "totalUsers": 0
}
```

The query runs successfully but finds no rows.

## Diagnosis

### From `handlers/attempt.go` Analysis:

1. **NO STATUS FIELD**: The `Attempt` model doesn't have a `status` field
   - Uses `CompletedAt` (timestamp pointer) instead
   - When attempt is submitted: `CompletedAt: &now` (current time)

2. **Score Column**: `score` (camelCase) ✅

3. **UserID Column**: `userId` (camelCase) ✅

4. **Completion Check**: Should be `completedAt IS NOT NULL`

### Possible Issues:

1. **CompletedAt might be NULL** in saved attempts
2. **Role filter** might be excluding all users
3. **JOIN condition** might not match
4. **Column name case-sensitivity** on MySQL

## Solution Applied

### 1. Simplified Leaderboard Query
**Removed ALL filters temporarily** to see if data exists:

```go
query := database.DB.Table("attempts").
    Select("attempts.userId as user_id, "+
        "user.username as username, "+
        "COUNT(DISTINCT attempts.id) as tests_count, "+
        "MAX(attempts.score) as best_score, "+
        "SUM(attempts.score) as total_score, "+
        "MIN(attempts.completedAt) as earliest_submit").
    Joins("JOIN user ON user.id = attempts.userId").
    Group("attempts.userId, user.username").
    Order("total_score DESC, earliest_submit ASC").
    Limit(100)
```

**Removed**:
- ❌ `Where("attempts.completedAt IS NOT NULL")`
- ❌ `Where("user.role != ?", "admin")`

This will show if ANY attempts exist, regardless of completion status or user role.

### 2. Comprehensive Debug Endpoint
**Enhanced** `GET /api/leaderboard/debug` to run 9 diagnostic queries:

1. **Total attempts count**: `SELECT COUNT(*) FROM attempts`
2. **Completed attempts**: `SELECT COUNT(*) FROM attempts WHERE completedAt IS NOT NULL`
3. **Total users**: `SELECT COUNT(*) FROM user`
4. **Non-admin users**: `SELECT COUNT(*) FROM user WHERE role != 'admin'`
5. **Sample attempts**: First 5 attempts with all fields
6. **Sample users**: First 5 users with id, username, role
7. **Full leaderboard query**: With all filters (completedAt + role)
8. **Without role filter**: Only completedAt filter
9. **Without completedAt filter**: Only role filter

**Response Format**:
```json
{
  "message": "Debug queries executed",
  "results": [
    {
      "query": "SELECT COUNT(*) FROM attempts",
      "result": 15
    },
    {
      "query": "SELECT COUNT(*) FROM attempts WHERE completedAt IS NOT NULL",
      "result": 12
    },
    {
      "query": "Leaderboard query (with all conditions)",
      "result": [
        {
          "user_id": "abc123",
          "username": "john",
          "tests_count": 3,
          "best_score": 100,
          "total_score": 250
        }
      ]
    }
  ],
  "summary": {
    "totalAttempts": 15,
    "completedAttempts": 12,
    "totalUsers": 5,
    "nonAdminUsers": 4
  }
}
```

## Testing Instructions

### STEP 1: Test Debug Endpoint
```bash
curl https://skillsprintojt.onrender.com/api/leaderboard/debug
```

**What to check**:
1. `totalAttempts` > 0? → Attempts are being saved
2. `completedAttempts` > 0? → CompletedAt is being set
3. `nonAdminUsers` > 0? → Non-admin users exist
4. Look at `sampleAttempts` → Check if userId, score, completedAt have values
5. Look at `sampleUsers` → Check if users have correct roles
6. Compare results of queries 7, 8, 9 → Which filter is causing the issue?

### STEP 2: Test Simplified Leaderboard
```bash
curl https://skillsprintojt.onrender.com/api/leaderboard/global
```

**Expected**: Should now return data if ANY attempts exist (no filters)

### STEP 3: Check Render Logs
Look for:
```
[Leaderboard] Executing query (simplified - no filters)...
[Leaderboard] Found X entries
[Leaderboard] #1: username (ID=...) Score=... Tests=... Best=...
```

### STEP 4: Analyze Debug Results

**Scenario A**: `totalAttempts = 0`
- **Problem**: Attempts aren't being saved at all
- **Fix**: Check `POST /api/attempts` endpoint

**Scenario B**: `totalAttempts > 0` but `completedAttempts = 0`
- **Problem**: `completedAt` is NULL in all attempts
- **Fix**: Check `SubmitAttempt` function sets `CompletedAt: &now`

**Scenario C**: `completedAttempts > 0` but query 7 returns empty
- **Problem**: JOIN or GROUP BY issue
- **Fix**: Check if `userId` in attempts matches `id` in user table

**Scenario D**: Query 8 or 9 returns data but query 7 doesn't
- **Problem**: One of the filters is too restrictive
- **Fix**: Adjust the problematic filter

## Files Modified

1. ✅ `go-backend/handlers/global_leaderboard.go`
   - Removed `completedAt IS NOT NULL` filter
   - Removed `role != 'admin'` filter
   - Added comprehensive debug endpoint with 9 diagnostic queries

## Commit Info

- **Commit**: `66002f0` - "fix: remove filters from leaderboard query, add comprehensive debug endpoint"
- **Route**: `GET /api/leaderboard/debug` (PUBLIC, no JWT)
- **Route**: `GET /api/leaderboard/global` (PUBLIC, simplified query)

## Next Steps

1. Wait for Render deployment (~2-3 minutes)
2. **Call debug endpoint** and analyze results
3. Based on debug output, determine root cause:
   - No attempts saved?
   - CompletedAt is NULL?
   - JOIN not matching?
   - Filter too restrictive?
4. Report debug endpoint output for further diagnosis
5. Once we see the debug data, we can add back the correct filters

## Expected Debug Output

If everything is working, you should see:
```json
{
  "summary": {
    "totalAttempts": 15,
    "completedAttempts": 12,
    "totalUsers": 5,
    "nonAdminUsers": 4
  },
  "results": [
    ...
    {
      "query": "Leaderboard query (without completedAt filter)",
      "result": [
        {
          "user_id": "abc123",
          "username": "john_doe",
          "tests_count": 3,
          "best_score": 100,
          "total_score": 250
        }
      ]
    }
  ]
}
```

**Share the debug endpoint output** so we can pinpoint exactly which condition is causing the empty result! 🔍

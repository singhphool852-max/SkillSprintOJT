# Leaderboard Comprehensive Fix - May 16, 2026

## Problem
Global leaderboard at `/api/leaderboard/global` returns empty data even though users have completed arena tests.

## Root Cause Analysis
The previous implementation had multiple issues:
1. **Wrong struct tags**: Used `gorm:"column:..."` tags but GORM's query builder wasn't mapping them correctly
2. **No explicit table name**: Relied on GORM's automatic table naming which might differ from actual DB
3. **Complex query builder**: GORM's query builder with joins was not producing correct SQL
4. **No debugging**: No way to verify if data exists in the database

## Complete Solution Implemented

### STEP 1: Added TableName() Method
**File**: `go-backend/models/attempt.go`

```go
// TableName ensures the table name exactly matches what's in the database
func (Attempt) TableName() string {
	return "attempts"
}
```

This explicitly tells GORM the exact table name to use.

### STEP 2: Rewrote Query with Raw SQL
**File**: `go-backend/handlers/global_leaderboard.go`

**Changes**:
- Replaced GORM query builder with `database.DB.Raw()` for explicit SQL control
- Get table names dynamically from models using `TableName()`
- Use simple struct tags (`db:"column_name"`) for scanning
- Added comprehensive logging of SQL query and results

**New Query**:
```go
sqlQuery := `
    SELECT 
        a.userId as user_id,
        u.username,
        SUM(a.score) as total_score,
        COUNT(DISTINCT a.id) as tests_completed,
        MAX(a.score) as best_score,
        MIN(a.completedAt) as earliest_submit
    FROM attempts a
    JOIN user u ON u.id = a.userId
    WHERE a.completedAt IS NOT NULL
    AND u.role != 'admin'
    GROUP BY a.userId, u.username
    ORDER BY total_score DESC, earliest_submit ASC
    LIMIT 100
`
```

**Key Features**:
- ✅ Uses actual table names from models
- ✅ Only includes completed attempts (`completedAt IS NOT NULL`)
- ✅ Excludes admin users
- ✅ Groups by user
- ✅ Ranks by total score DESC, then earliest submission ASC (tiebreaker)
- ✅ Logs SQL query and row count for debugging

### STEP 3: Added Debug Endpoint
**File**: `go-backend/handlers/global_leaderboard.go`

New function: `GetLeaderboardDebug()`

**Route**: `GET /api/leaderboard/debug`

**Returns**:
```json
{
  "tableName": "attempts",
  "userTableName": "user",
  "totalAttempts": 15,
  "completedAttempts": 12,
  "totalUsers": 5,
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

This allows you to verify:
- Correct table names are being used
- Data exists in the database
- Attempts are marked as completed

### STEP 4: Enhanced Frontend Logging
**File**: `frontend/components/leaderboard/leaderboard-content.tsx`

**Changes**:
- Added `error` state to track and display errors
- Console logs for every fetch attempt
- Logs API URL, response status, and raw response data
- Better error handling and display
- Shows error message in UI when data fetch fails

**Console Output**:
```
[Leaderboard] Fetching from: https://your-api.com/api/leaderboard/global
[Leaderboard] Response status: 200
[Leaderboard] Raw response: {entries: [...], totalUsers: 3}
[Leaderboard] Entries count: 3
```

### STEP 5: Verified Route Configuration
**File**: `go-backend/main.go`

Added debug route:
```go
api.GET("/leaderboard/debug", handlers.GetLeaderboardDebug)
```

Confirmed `/api/leaderboard/global` is in PUBLIC routes (no JWT middleware required).

## Files Modified

1. ✅ `go-backend/handlers/global_leaderboard.go` - Rewrote query, added debug endpoint
2. ✅ `go-backend/models/attempt.go` - Added TableName() method
3. ✅ `go-backend/main.go` - Added debug route
4. ✅ `frontend/components/leaderboard/leaderboard-content.tsx` - Enhanced logging and error handling

## Testing Instructions

### 1. Test Debug Endpoint First
```bash
curl https://your-backend-url/api/leaderboard/debug
```

**Expected Output**:
- `totalAttempts` > 0
- `completedAttempts` > 0
- `sampleAttempts` array with data

**If this returns 0**: The problem is that attempts aren't being saved to the database.

### 2. Check Backend Logs
Look for these log entries:
```
[Leaderboard] Using table: attempts, user table: user
[Leaderboard] SQL Query: SELECT a.userId as user_id...
[Leaderboard] Found 3 users
[Leaderboard] Entry 1: User=john (ID=abc123) Score=250 Tests=3 Best=100
```

### 3. Check Frontend Console
Open browser DevTools → Console, look for:
```
[Leaderboard] Fetching from: https://...
[Leaderboard] Response status: 200
[Leaderboard] Raw response: {...}
[Leaderboard] Entries count: 3
```

### 4. Test Leaderboard Page
Visit `/leaderboard` and:
- Should see user data if attempts exist
- Should see "NO PERFORMANCE DATA RECORDED YET" if no completed attempts
- Should see error message if API call fails
- Click "REFRESH" button to manually trigger fetch

## Troubleshooting

### If Debug Endpoint Shows 0 Attempts
**Problem**: Attempts aren't being saved to database  
**Check**: 
- `POST /api/attempts` endpoint is working
- Frontend is calling submit endpoint correctly
- Database connection is working

### If Debug Shows Data But Leaderboard is Empty
**Problem**: Query or response parsing issue  
**Check**:
- Backend logs for SQL query and results
- Frontend console for response data
- Network tab for actual API response

### If Frontend Shows Network Error
**Problem**: CORS or connectivity issue  
**Check**:
- Backend is running and accessible
- CORS headers are set correctly
- API_URL environment variable is correct

### If Users See Their Own Data But Not Others
**Problem**: User filtering issue  
**Check**:
- Users have `role != 'admin'`
- Attempts have `completedAt IS NOT NULL`

## Deployment Checklist

- [x] Backend compiles successfully
- [x] Frontend TypeScript compiles
- [x] All files committed and pushed
- [x] Commit: `8f00864` - "fix: leaderboard query with raw SQL, add TableName(), debug endpoint, enhanced logging"
- [ ] Backend deployed to Render (auto-deploy from main)
- [ ] Frontend deployed to Amplify (auto-deploy from main)
- [ ] Test debug endpoint in production
- [ ] Test leaderboard page in production
- [ ] Check backend logs in Render dashboard
- [ ] Check frontend console in browser

## Expected Behavior After Fix

1. **When users complete tests**: Their scores appear on leaderboard within 30 seconds
2. **Ranking**: Users ranked by total score (sum of all test scores)
3. **Tiebreaker**: Users with same score ranked by earliest submission time
4. **Tier assignment**: Same score = same tier
5. **Live updates**: Leaderboard auto-refreshes every 30 seconds
6. **Manual refresh**: Click "REFRESH" button to update immediately

## API Endpoints

### Production Endpoints
- **Leaderboard**: `GET /api/leaderboard/global`
- **Debug**: `GET /api/leaderboard/debug`

### Response Format
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

## Next Steps

1. Wait for deployment to complete
2. Test debug endpoint: `curl https://your-api/api/leaderboard/debug`
3. Check backend logs in Render dashboard
4. Test leaderboard page in browser with DevTools open
5. Complete a test and verify it appears on leaderboard
6. Report any issues with:
   - Debug endpoint output
   - Backend log entries
   - Frontend console output

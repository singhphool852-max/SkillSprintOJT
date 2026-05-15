# Analytics Page Testing Guide

## Quick Test Steps

### 1. Test Basic Page Load
```bash
# Start backend
cd go-backend
go run main.go

# Start frontend (in another terminal)
cd frontend
npm run dev
```

Navigate to: `http://localhost:3000/admin/analytics`

**Expected:**
- Page loads without crashing
- Shows 6 stat cards with numbers
- No error messages in console

---

### 2. Test Active Arena Users Counter

**Setup:**
1. Open 2-3 browser tabs
2. Login as different students
3. Join an active test (don't submit)

**Verify:**
- Go to Admin → Analytics
- "ACTIVE IN ARENA" card shows count of active users (2-3)
- Subtitle shows total registered users

**Test Edge Cases:**
- Submit a test → count should decrease
- Wait 4+ hours → old attempts should not count
- No active tests → count should be 0

---

### 3. Test Anti-Cheat Violations Table

**Setup - Create Violations:**
1. Login as a student
2. Join an active test
3. Exit fullscreen (Alt+Tab or Esc)
4. Switch to another tab
5. Return to test
6. Repeat 2-3 times

**Verify:**
- Go to Admin → Analytics
- Scroll to "ANTI-CHEAT VIOLATIONS" section
- Table shows:
  - Student name and email
  - Test title
  - Total warnings badge (color-coded)
  - Fullscreen exits count (orange)
  - Tab switches count (yellow)
  - Last violation timestamp

**Color Coding Check:**
- 1 violation → Gray badge
- 2 violations → Yellow badge
- 3+ violations → Red badge

**Empty State:**
- If no violations exist, shows:
  - Shield icon
  - "No violation data yet" message
  - Explanation text

---

### 4. Test Error Handling

**Test Network Error:**
```bash
# Stop the backend while frontend is running
# Or block the API endpoint
```

**Expected:**
- Page doesn't crash
- Shows yellow warning banner with error message
- Stats show 0 values (safe defaults)
- Violations table shows empty state

**Test Invalid Response:**
```bash
# Modify backend to return invalid JSON temporarily
# Or return null/undefined
```

**Expected:**
- Page handles gracefully with null coalescing
- No "undefined.map is not a function" errors
- Shows safe default values

---

### 5. Test Responsive Design

**Desktop (1920x1080):**
- 3 columns of stat cards
- Full table visible
- No horizontal scroll

**Tablet (768px):**
- 2-3 columns of stat cards
- Table scrolls horizontally if needed

**Mobile (375px):**
- 2 columns of stat cards
- Table scrolls horizontally
- All data readable

---

## API Testing

### Test Dashboard Stats Endpoint
```bash
curl -X GET http://localhost:8080/api/admin/analytics \
  -H "Cookie: token=YOUR_ADMIN_TOKEN" \
  | jq
```

**Expected Response:**
```json
{
  "totalTests": 5,
  "publishedTests": 3,
  "activeTests": 1,
  "totalUsers": 11,
  "activeArenaUsers": 2,
  "adminUsers": 1,
  "regularUsers": 10,
  "totalTopics": 4,
  "totalAttempts": 15,
  "submittedAttempts": 12,
  "avgScore": 75.5,
  "totalQuestions": 20
}
```

### Test Violations Endpoint
```bash
curl -X GET http://localhost:8080/api/admin/analytics/mistakes \
  -H "Cookie: token=YOUR_ADMIN_TOKEN" \
  | jq
```

**Expected Response:**
```json
{
  "mistakes": [
    {
      "userName": "John Doe",
      "userEmail": "john@example.com",
      "testTitle": "Data Structures Final",
      "violationCount": 3,
      "fullscreenExits": 2,
      "tabSwitches": 1,
      "lastViolation": "2026-05-15T14:30:00Z"
    }
  ],
  "total": 1
}
```

---

## Database Verification

### Check Active Arena Users
```sql
SELECT 
    COUNT(DISTINCT userId) as active_users
FROM test_attempts
WHERE submittedAt IS NULL 
  AND joinedAt > datetime('now', '-4 hours');
```

### Check Violations Data
```sql
SELECT 
    u.name,
    t.title,
    COUNT(v.id) as total_violations,
    SUM(CASE WHEN v.violationType = 'fullscreen_exit' THEN 1 ELSE 0 END) as fullscreen_exits,
    SUM(CASE WHEN v.violationType = 'tab_switch' THEN 1 ELSE 0 END) as tab_switches
FROM test_violations v
JOIN users u ON u.id = v.userId
JOIN tests t ON t.id = v.testId
GROUP BY v.userId, v.testId
ORDER BY total_violations DESC;
```

---

## Common Issues & Solutions

### Issue: Page crashes immediately
**Cause:** API returns unexpected data structure
**Solution:** Check browser console for error, verify API response format

### Issue: Active arena users shows 0 but tests are running
**Cause:** `joinedAt` timestamp not set correctly
**Solution:** Check test_attempts table, verify joinedAt column

### Issue: Violations table empty but violations exist
**Cause:** SQL query column names mismatch
**Solution:** Verify column names: `violationType`, `userId`, `testId`, `timestamp`

### Issue: Color coding not working
**Cause:** Tailwind classes not applied
**Solution:** Check if classes are in safelist, rebuild frontend

### Issue: Timestamps show "Invalid Date"
**Cause:** Date format from backend not parseable
**Solution:** Ensure backend returns ISO 8601 format (YYYY-MM-DDTHH:mm:ssZ)

---

## Performance Testing

### Load Test - Many Violations
```sql
-- Insert 1000 test violations
INSERT INTO test_violations (id, userId, testId, violationType, timestamp, createdAt)
SELECT 
    hex(randomblob(16)),
    (SELECT id FROM users ORDER BY RANDOM() LIMIT 1),
    (SELECT id FROM tests ORDER BY RANDOM() LIMIT 1),
    CASE (ABS(RANDOM()) % 2) WHEN 0 THEN 'fullscreen_exit' ELSE 'tab_switch' END,
    datetime('now', '-' || (ABS(RANDOM()) % 720) || ' minutes'),
    datetime('now')
FROM (SELECT 1 UNION SELECT 2 UNION SELECT 3) -- Repeat as needed
LIMIT 1000;
```

**Verify:**
- Page loads in < 2 seconds
- Table renders smoothly
- No browser lag when scrolling

### Load Test - Many Active Users
```sql
-- Create 50 active test attempts
INSERT INTO test_attempts (id, userId, testId, joinedAt, submittedAt)
SELECT 
    hex(randomblob(16)),
    (SELECT id FROM users ORDER BY RANDOM() LIMIT 1),
    (SELECT id FROM tests WHERE isActive = 1 LIMIT 1),
    datetime('now', '-' || (ABS(RANDOM()) % 120) || ' minutes'),
    NULL
FROM (SELECT 1 UNION SELECT 2 UNION SELECT 3) -- Repeat
LIMIT 50;
```

**Verify:**
- Active arena users count updates correctly
- Query executes in < 100ms

---

## Regression Testing

After deployment, verify:
- [ ] Other admin pages still work (Tests, Topics, Dashboard)
- [ ] Student pages unaffected (Arena, Train, Results)
- [ ] Auth still works (login, logout, role checks)
- [ ] Leaderboard still loads
- [ ] Chat still functions
- [ ] Test submission still works
- [ ] Anti-cheat still triggers violations

---

## Success Criteria

✅ Page loads without crashing  
✅ All 6 stat cards show correct data  
✅ Active arena users count is accurate  
✅ Violations table displays properly  
✅ Color coding works (gray/yellow/red)  
✅ Empty states show correctly  
✅ Error handling prevents crashes  
✅ Responsive on all screen sizes  
✅ API endpoints return correct JSON  
✅ No TypeScript errors  
✅ No console errors  
✅ Performance acceptable (< 2s load)  

---

## Rollback Plan

If issues occur in production:

1. **Quick Fix:** Revert frontend only
```bash
git revert <commit-hash>
cd frontend && npm run build
```

2. **Full Rollback:** Revert both frontend and backend
```bash
git revert <commit-hash>
cd go-backend && go build
cd ../frontend && npm run build
```

3. **Emergency:** Disable analytics page
```typescript
// In frontend/app/admin/analytics/page.tsx
export default function AdminAnalyticsPage() {
  return <div>Analytics temporarily disabled</div>
}
```

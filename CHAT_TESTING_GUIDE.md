# SkillSprint Chat - Testing Guide

## Manual Testing Checklist

### ✅ Basic Functionality

#### 1. Page Access
- [ ] Navigate to `/chat` from navbar
- [ ] Page loads without errors
- [ ] Chat link shows 💬 icon in navbar
- [ ] Page requires authentication (redirects to login if not logged in)

#### 2. WebSocket Connection
- [ ] Green dot appears when connected
- [ ] Online count shows in header
- [ ] Connection status updates in real-time
- [ ] Auto-reconnects after disconnect (test by stopping backend)

#### 3. Text Messages
- [ ] Type message in input box
- [ ] Press Enter to send
- [ ] Click Send button to send
- [ ] Message appears immediately
- [ ] Own messages appear on right (pink background)
- [ ] Other users' messages appear on left (gray background)
- [ ] Username shows on messages
- [ ] Timestamp shows on messages
- [ ] Empty messages cannot be sent

#### 4. Chat History
- [ ] Last 50 messages load on page open
- [ ] Messages appear in chronological order (oldest first)
- [ ] Scroll works correctly
- [ ] Auto-scrolls to bottom on new message

### ✅ File Upload

#### 5. Image Upload
- [ ] Click 📎 button
- [ ] Select JPEG file
- [ ] Image uploads successfully
- [ ] Image displays inline
- [ ] Image max width is 300px
- [ ] Repeat for PNG file
- [ ] Repeat for GIF file
- [ ] File over 10MB is rejected
- [ ] Non-image file is rejected (when selecting image)

#### 6. PDF Upload
- [ ] Click 📎 button
- [ ] Select PDF file
- [ ] PDF uploads successfully
- [ ] PDF shows with 📄 icon
- [ ] Filename displays correctly
- [ ] Download button works
- [ ] Downloaded file opens correctly
- [ ] File over 10MB is rejected

#### 7. Upload Error Handling
- [ ] Try uploading .exe file (should be rejected)
- [ ] Try uploading .zip file (should be rejected)
- [ ] Try uploading 15MB file (should be rejected)
- [ ] Error messages display correctly
- [ ] Upload progress indicator shows during upload

### ✅ Notes Feature

#### 8. Note Composition
- [ ] Click 📝 button
- [ ] Note modal opens
- [ ] Type note content
- [ ] Click "SEND NOTE"
- [ ] Note appears with yellow background
- [ ] Note shows 📝 icon
- [ ] Click "CANCEL" closes modal without sending
- [ ] Empty notes cannot be sent

#### 9. Note Display
- [ ] Notes display with yellow background
- [ ] Notes show "NOTE" label
- [ ] Multi-line notes display correctly
- [ ] Long notes wrap correctly
- [ ] Notes preserve line breaks

### ✅ Multi-User Testing

#### 10. Multiple Users (requires 2+ browser windows)
- [ ] Open chat in 2 different browsers/windows
- [ ] Login as different users
- [ ] Send message from User 1
- [ ] Message appears for User 2 immediately
- [ ] Send message from User 2
- [ ] Message appears for User 1 immediately
- [ ] Online count increases when user joins
- [ ] Online count decreases when user leaves
- [ ] Both users see same chat history

#### 11. Real-Time Updates
- [ ] User 1 sends text message → User 2 sees it instantly
- [ ] User 1 uploads image → User 2 sees it instantly
- [ ] User 1 uploads PDF → User 2 sees it instantly
- [ ] User 1 sends note → User 2 sees it instantly
- [ ] User 1 joins → User 2 sees online count increase
- [ ] User 1 leaves → User 2 sees online count decrease

### ✅ UI/UX Testing

#### 12. Visual Design
- [ ] Colors match SkillSprint theme (neon-cyan, neon-pink, neon-yellow)
- [ ] Font is monospace (font-mono)
- [ ] Borders and panels match rest of app
- [ ] Message bubbles have correct styling
- [ ] Hover effects work on buttons
- [ ] Icons display correctly

#### 13. Responsive Design
- [ ] Test on desktop (1920x1080)
- [ ] Test on laptop (1366x768)
- [ ] Test on tablet (768x1024)
- [ ] Test on mobile (375x667)
- [ ] Chat link appears in mobile menu
- [ ] Messages wrap correctly on small screens
- [ ] Input bar works on mobile
- [ ] File upload works on mobile

#### 14. Accessibility
- [ ] Tab navigation works
- [ ] Enter key sends message
- [ ] Buttons have hover states
- [ ] Icons have titles/tooltips
- [ ] Color contrast is sufficient
- [ ] Text is readable

### ✅ Performance Testing

#### 15. Load Testing
- [ ] Send 100 messages rapidly
- [ ] Upload 10 files in succession
- [ ] Scroll through long chat history
- [ ] Keep connection open for 1 hour
- [ ] Reconnect after network interruption
- [ ] No memory leaks in browser
- [ ] No memory leaks in backend

#### 16. Edge Cases
- [ ] Very long message (1000+ characters)
- [ ] Message with special characters (!@#$%^&*)
- [ ] Message with emojis (😀🎉🚀)
- [ ] Message with URLs
- [ ] Message with code blocks
- [ ] Filename with special characters
- [ ] Filename with spaces
- [ ] Very long filename

### ✅ Security Testing

#### 17. Authentication
- [ ] Cannot access `/chat` without login
- [ ] Cannot connect WebSocket without JWT
- [ ] Cannot upload files without JWT
- [ ] Cannot fetch history without JWT
- [ ] Invalid JWT is rejected
- [ ] Expired JWT is rejected

#### 18. File Upload Security
- [ ] .exe files rejected
- [ ] .sh files rejected
- [ ] .bat files rejected
- [ ] Files over 10MB rejected
- [ ] Invalid MIME types rejected
- [ ] Filename sanitization works (special chars removed)

#### 19. Data Privacy
- [ ] Email addresses not exposed in messages
- [ ] Only username and avatar shown
- [ ] Cannot see other users' JWT tokens
- [ ] Cannot access other users' uploaded files without URL

### ✅ Error Handling

#### 20. Network Errors
- [ ] Stop backend → Frontend shows disconnected
- [ ] Start backend → Frontend reconnects automatically
- [ ] Slow network → Upload shows progress
- [ ] Network timeout → Error message shown

#### 21. Database Errors
- [ ] Stop database → Backend logs error
- [ ] Start database → Backend reconnects
- [ ] History fetch fails gracefully
- [ ] Message save fails gracefully

#### 22. File System Errors
- [ ] Delete uploads folder → Upload fails with error
- [ ] Recreate uploads folder → Upload works again
- [ ] Full disk → Upload fails with error

## Automated Testing (Future)

### Unit Tests (Backend)
```go
// Test Hub registration
func TestHubRegister(t *testing.T) { ... }

// Test message broadcasting
func TestHubBroadcast(t *testing.T) { ... }

// Test file validation
func TestFileValidation(t *testing.T) { ... }

// Test filename sanitization
func TestFilenameSanitization(t *testing.T) { ... }
```

### Unit Tests (Frontend)
```typescript
// Test useChat hook
describe('useChat', () => {
  it('connects to WebSocket', () => { ... })
  it('sends messages', () => { ... })
  it('uploads files', () => { ... })
})

// Test ChatPage component
describe('ChatPage', () => {
  it('renders messages', () => { ... })
  it('handles file upload', () => { ... })
  it('opens note modal', () => { ... })
})
```

### Integration Tests
```typescript
// Test end-to-end message flow
describe('Chat E2E', () => {
  it('sends and receives messages', () => { ... })
  it('uploads and displays files', () => { ... })
  it('shows online count', () => { ... })
})
```

## Performance Benchmarks

### Expected Performance
- **Message latency**: < 100ms
- **File upload**: < 2s for 5MB file
- **History load**: < 500ms for 50 messages
- **WebSocket reconnect**: < 3s
- **Concurrent users**: 100+ (single server)

### Load Testing Commands
```bash
# Test WebSocket connections
wscat -c ws://localhost:8080/ws/chat?token=YOUR_TOKEN

# Test file upload
curl -X POST http://localhost:8080/api/chat/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.jpg"

# Test history fetch
curl http://localhost:8080/api/chat/history \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Browser Compatibility Testing

### Desktop Browsers
- [ ] Chrome 120+ (Windows)
- [ ] Chrome 120+ (macOS)
- [ ] Chrome 120+ (Linux)
- [ ] Firefox 120+ (Windows)
- [ ] Firefox 120+ (macOS)
- [ ] Firefox 120+ (Linux)
- [ ] Safari 17+ (macOS)
- [ ] Edge 120+ (Windows)

### Mobile Browsers
- [ ] Chrome (Android)
- [ ] Safari (iOS)
- [ ] Firefox (Android)
- [ ] Samsung Internet (Android)

## Database Testing

### Schema Validation
```sql
-- Verify table exists
SHOW TABLES LIKE 'chat_messages';

-- Verify columns
DESCRIBE chat_messages;

-- Verify indexes
SHOW INDEXES FROM chat_messages;

-- Test insert
INSERT INTO chat_messages (userId, username, messageType, content, createdAt)
VALUES ('test-id', 'test-user', 'text', 'test message', NOW());

-- Test query
SELECT * FROM chat_messages ORDER BY createdAt DESC LIMIT 50;
```

### Data Integrity
- [ ] Messages persist after backend restart
- [ ] Timestamps are correct
- [ ] User IDs match user table
- [ ] File URLs are valid
- [ ] No duplicate messages

## Regression Testing

After any code changes, verify:
- [ ] Existing features still work
- [ ] No new console errors
- [ ] No new TypeScript errors
- [ ] No new Go compilation errors
- [ ] Build succeeds
- [ ] Tests pass

## Bug Report Template

When reporting bugs, include:

```markdown
**Bug Description:**
[Clear description of the issue]

**Steps to Reproduce:**
1. Go to /chat
2. Click upload button
3. Select file
4. ...

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Environment:**
- Browser: Chrome 120
- OS: Windows 11
- Backend version: [commit hash]
- Frontend version: [commit hash]

**Screenshots:**
[Attach screenshots if applicable]

**Console Errors:**
[Paste any console errors]

**Backend Logs:**
[Paste relevant backend logs]
```

## Test Data

### Sample Messages
```
"Hello, world!"
"This is a test message with special chars: !@#$%^&*()"
"Multi-line\nmessage\ntest"
"Very long message: " + "a".repeat(1000)
"Message with emoji: 😀🎉🚀"
"Message with URL: https://example.com"
```

### Sample Files
- `test-image.jpg` (1MB)
- `test-image-large.jpg` (9MB)
- `test-image-too-large.jpg` (15MB)
- `test-document.pdf` (2MB)
- `test-document-large.pdf` (9MB)
- `test-invalid.exe` (should be rejected)

## Success Criteria

All tests must pass before considering the feature complete:
- ✅ All basic functionality tests pass
- ✅ All file upload tests pass
- ✅ All multi-user tests pass
- ✅ All UI/UX tests pass
- ✅ All security tests pass
- ✅ All error handling tests pass
- ✅ No console errors
- ✅ No memory leaks
- ✅ Performance meets benchmarks

---

**Testing Status:** Ready for QA
**Last Updated:** [Current Date]
**Tested By:** [Your Name]

# SkillSprint Community Chat Feature

## Overview
A real-time global chat room where all logged-in users can communicate, share notes, images, and PDF files.

## Features

### Real-Time Messaging
- WebSocket-based instant messaging
- Live online user count
- Connection status indicator
- Auto-reconnect on disconnect

### Message Types
1. **Text Messages** - Regular chat messages
2. **Notes** - Formatted text blocks with special styling (yellow background)
3. **Images** - JPEG, PNG, GIF (displayed inline, max 300px width)
4. **PDFs** - Downloadable documents with file icon

### User Experience
- Messages show username and avatar
- Timestamps on all messages
- Own messages aligned right (purple/pink theme)
- Other users' messages aligned left (gray theme)
- Auto-scroll to latest message
- Chat history loads on page open (last 50 messages)

## Technical Implementation

### Backend (Go)

#### Files Created
- `go-backend/chat/hub.go` - WebSocket hub managing all chat clients
- `go-backend/handlers/chat.go` - HTTP handlers for chat endpoints
- `go-backend/models/chat.go` - ChatMessage database model

#### API Endpoints
- `GET /ws/chat` - WebSocket connection (JWT required)
- `POST /api/chat/upload` - File upload endpoint (JWT required)
- `GET /api/chat/history` - Fetch last 50 messages (JWT required)
- `GET /uploads/chat/:filename` - Serve uploaded files (public)

#### Database Schema
```sql
CREATE TABLE chat_messages (
  id INT AUTO_INCREMENT PRIMARY KEY,
  userId VARCHAR(36) NOT NULL,
  username VARCHAR(255) NOT NULL,
  avatar VARCHAR(255),
  messageType VARCHAR(20) DEFAULT 'text',
  content TEXT,
  fileName VARCHAR(255),
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_userId (userId),
  INDEX idx_createdAt (createdAt)
);
```

#### WebSocket Events
- `message` - Chat message from user
- `user_joined` - User connected to chat
- `user_left` - User disconnected from chat
- `online_count` - Current online user count

#### File Upload
- Max file size: 10MB
- Allowed types: JPEG, PNG, GIF, PDF
- Files stored in: `go-backend/uploads/chat/`
- Filename format: `{timestamp}_{userId}_{sanitized_name}.{ext}`
- Filenames sanitized (alphanumeric, dash, underscore only)

### Frontend (Next.js/React)

#### Files Created
- `frontend/hooks/useChat.ts` - Custom hook for chat functionality
- `frontend/app/chat/page.tsx` - Chat page component

#### Files Modified
- `frontend/components/nav.tsx` - Added chat link with icon

#### Chat Hook (`useChat`)
Provides:
- `messages` - Array of chat messages
- `onlineCount` - Number of online users
- `isConnected` - WebSocket connection status
- `sendMessage(content, type)` - Send text/note message
- `sendFile(file)` - Upload and send file

#### UI Components
- Header with online count and connection status
- Scrollable messages area
- Message bubbles with different styles per type
- Input bar with file upload, note, text input, and send buttons
- Note modal for composing formatted notes
- Upload progress indicator

## Security

### Authentication
- All WebSocket connections require valid JWT token
- All API endpoints require JWT authentication
- Token passed via cookie or Authorization header

### File Upload Security
- File type validation (whitelist only)
- File size limit (10MB)
- Filename sanitization (remove special characters)
- Files stored outside web root
- No executable file types allowed

### Data Privacy
- Only username and avatar shown (no email addresses)
- User ID used for message ownership
- Messages persist in database

## Usage

### For Users
1. Navigate to `/chat` from the navbar
2. See online user count in header
3. Type message and press Enter or click Send
4. Click 📎 to upload image or PDF
5. Click 📝 to compose a formatted note
6. Scroll to view chat history

### For Developers

#### Starting the Backend
```bash
cd go-backend
go run main.go
```

#### Starting the Frontend
```bash
cd frontend
npm run dev
```

#### Environment Variables
Backend requires:
- `MYSQL_DSN` - MySQL connection string
- `JWT_SECRET` - Secret key for JWT tokens

Frontend requires:
- `NEXT_PUBLIC_API_URL` - Backend API URL

## Database Migration
The `chat_messages` table is automatically created via GORM AutoMigrate when the backend starts.

## File Storage
Uploaded files are stored locally in `go-backend/uploads/chat/`. For production, consider:
- Using cloud storage (S3, GCS, Azure Blob)
- Implementing CDN for file delivery
- Adding file cleanup/retention policies

## Future Enhancements
- Private direct messages
- Message reactions/emojis
- User mentions (@username)
- Message search
- File preview for PDFs
- Image thumbnails
- Message editing/deletion
- Typing indicators
- Read receipts
- User roles/moderation
- Message pinning
- Chat rooms/channels

## Testing

### Manual Testing Checklist
- [ ] User can connect to chat
- [ ] Messages appear in real-time
- [ ] Online count updates correctly
- [ ] File upload works for images
- [ ] File upload works for PDFs
- [ ] Notes display with correct styling
- [ ] Chat history loads on page open
- [ ] Auto-scroll works
- [ ] Reconnect works after disconnect
- [ ] Own messages appear on right
- [ ] Other messages appear on left
- [ ] Timestamps display correctly
- [ ] File download works
- [ ] Images display inline
- [ ] Large files rejected (>10MB)
- [ ] Invalid file types rejected

## Troubleshooting

### WebSocket Connection Fails
- Check JWT token is valid
- Verify CORS settings in backend
- Check WebSocket URL format (ws:// not http://)

### Files Not Uploading
- Check file size (<10MB)
- Verify file type is allowed
- Ensure uploads directory exists and is writable
- Check backend logs for errors

### Messages Not Appearing
- Verify WebSocket connection is active
- Check browser console for errors
- Ensure database is running
- Check backend logs for errors

## Performance Considerations
- Chat history limited to last 50 messages
- WebSocket ping/pong every 54 seconds
- File uploads processed synchronously
- No message pagination (yet)
- No rate limiting (yet)

## Browser Compatibility
- Modern browsers with WebSocket support
- Chrome 16+
- Firefox 11+
- Safari 7+
- Edge 12+

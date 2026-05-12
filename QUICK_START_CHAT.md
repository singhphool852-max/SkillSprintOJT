# Quick Start Guide - SkillSprint Community Chat

## 🚀 Getting Started

### 1. Start the Backend
```bash
cd go-backend
go run main.go
```

The backend will:
- Connect to MySQL database
- Auto-create `chat_messages` table
- Start WebSocket hub on `/ws/chat`
- Serve uploaded files from `/uploads/chat`

### 2. Start the Frontend
```bash
cd frontend
npm run dev
```

### 3. Access the Chat
1. Open browser to `http://localhost:3000`
2. Login with your account
3. Click "CHAT" in the navbar (💬 icon)
4. Start chatting!

## 📱 Using the Chat

### Send a Text Message
1. Type your message in the input box
2. Press Enter or click the Send button

### Share a Note
1. Click the 📝 (FileText) button
2. Write your note in the modal
3. Click "SEND NOTE"
4. Note appears with yellow background

### Upload an Image
1. Click the 📎 (Paperclip) button
2. Select a JPEG, PNG, or GIF file (max 10MB)
3. Image uploads and displays inline

### Upload a PDF
1. Click the 📎 (Paperclip) button
2. Select a PDF file (max 10MB)
3. PDF appears with download button

## 🔍 Features at a Glance

| Feature | Description |
|---------|-------------|
| **Real-time messaging** | Messages appear instantly for all users |
| **Online count** | See how many users are currently online |
| **Connection status** | Green dot = connected, Red dot = disconnected |
| **Chat history** | Last 50 messages load when you open the page |
| **Auto-scroll** | Automatically scrolls to newest message |
| **File sharing** | Share images and PDFs up to 10MB |
| **Notes** | Share formatted text blocks with special styling |
| **Timestamps** | Every message shows when it was sent |
| **User avatars** | See profile pictures next to messages |

## 🎨 Message Types

### Text Message
```
Regular chat message with gray background (others) or pink (yours)
```

### Note
```
📝 NOTE
Formatted text block with yellow background
Perfect for sharing code snippets or important info
```

### Image
```
[Image displays inline, max 300px width]
Supports: JPEG, PNG, GIF
```

### PDF
```
📄 filename.pdf
[Download button]
```

## 🔒 Security

- ✅ JWT authentication required
- ✅ File type validation (whitelist only)
- ✅ File size limit (10MB max)
- ✅ Filename sanitization
- ✅ No executable files allowed

## 🛠️ Troubleshooting

### "WebSocket not connected" error
**Solution:** Check that:
1. Backend is running
2. You're logged in (JWT token valid)
3. CORS settings allow your origin

### File upload fails
**Solution:** Verify:
1. File is under 10MB
2. File type is JPEG, PNG, GIF, or PDF
3. `go-backend/uploads/chat/` directory exists

### Messages not appearing
**Solution:** Check:
1. Green dot shows "connected" status
2. Browser console for errors
3. Backend logs for errors

### Can't see chat history
**Solution:** Ensure:
1. Database is running
2. `chat_messages` table exists
3. Backend can connect to database

## 📊 Technical Details

### WebSocket Endpoint
```
ws://localhost:8080/ws/chat?token=YOUR_JWT_TOKEN
```

### API Endpoints
```
POST /api/chat/upload      - Upload file
GET  /api/chat/history     - Get last 50 messages
GET  /uploads/chat/:file   - Download file
```

### Database Table
```sql
chat_messages (
  id, userId, username, avatar,
  messageType, content, fileName, createdAt
)
```

## 🎯 Quick Tips

1. **Press Enter** to send messages quickly
2. **Scroll up** to see older messages
3. **Watch the online count** to see who's active
4. **Use notes** for longer formatted messages
5. **Share images** for visual communication
6. **Upload PDFs** for document sharing

## 📝 Example Usage

### Sharing Study Notes
1. Click 📝 button
2. Paste your notes
3. Click "SEND NOTE"
4. Everyone sees it with yellow highlight

### Sharing a Diagram
1. Click 📎 button
2. Select your diagram image
3. Image appears inline for everyone

### Sharing a PDF Resource
1. Click 📎 button
2. Select PDF file
3. Others can download it

## 🎉 That's It!

You're ready to use the SkillSprint Community Chat. Start connecting with other learners, share resources, and build your study community!

---

**Need help?** Check `CHAT_FEATURE.md` for detailed documentation.

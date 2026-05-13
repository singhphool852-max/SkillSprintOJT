# AI Test Builder Feature - Implementation Summary

## Overview
Added AI-powered test generation from PDF/CSV files for admin users. Admins can upload documents containing notes, MCQs, or coding questions, and AI automatically generates a complete draft test.

## Features Implemented

### Backend (Go)

**New File: `go-backend/handlers/ai_test_builder.go`**
- `HandleAIBuildTest`: Main handler for file upload and AI processing
- `extractTextFromPDF`: Extracts text content from PDF files
- `extractTextFromCSV`: Extracts text content from CSV files
- `generateTestWithOpenAI`: Calls OpenAI API to generate test structure
- `createDraftTestFromAI`: Converts AI response to database models

**Key Features:**
- ✅ PDF parsing using `github.com/ledongthuc/pdf`
- ✅ CSV parsing using standard library
- ✅ OpenAI GPT-4o-mini integration (reads from env vars)
- ✅ Strict JSON response parsing with markdown cleanup
- ✅ Automatic question generation from notes
- ✅ Automatic testcase generation for coding problems
- ✅ Creates tests as DRAFT only (not published)
- ✅ Transactional database operations
- ✅ Proper error handling and validation

**API Endpoint:**
```
POST /api/admin/ai/build-test
Headers: Authorization: Bearer <token>
Body: multipart/form-data with "file" field
Accepts: .pdf, .csv
```

**Response:**
```json
{
  "message": "Test generated successfully",
  "testId": "uuid",
  "test": { /* AI generated test structure */ }
}
```

### Frontend (React/Next.js)

**Modified File: `frontend/app/admin/page.tsx`**

**New UI Components:**
1. **"AI BUILD TEST" Button** (cyan/teal themed)
   - Positioned next to "CREATE TEST" button
   - Opens modal for file upload

2. **AI Upload Modal**
   - File input for PDF/CSV
   - File size display
   - Loading state with spinner
   - Error handling display
   - Information about draft creation
   - Matches existing dark theme

**User Flow:**
1. Admin clicks "AI BUILD TEST" button
2. Modal opens with file upload
3. Admin selects PDF or CSV file
4. Clicks "GENERATE TEST"
5. Loading state shows "GENERATING..."
6. On success: Redirects to test edit page
7. On error: Shows error message in modal

### Environment Configuration

**Required Environment Variables (already configured on Render):**
```bash
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4o-mini  # Optional, defaults to gpt-4o-mini
```

**No hardcoded API keys** - reads from environment using `os.Getenv()`

## AI Prompt Engineering

The AI is instructed to:
- Detect MCQ vs coding questions automatically
- Generate 3-5 MCQs if content has theory/concepts
- Generate 1-3 coding questions if content has algorithms
- Create questions FROM notes if only notes provided
- Generate testcases if missing
- Return strict JSON (no markdown)
- Include proper constraints, starter code, points

**JSON Structure:**
```json
{
  "title": "Test title",
  "description": "Brief description",
  "difficulty": "easy|medium|hard",
  "durationMinutes": 60,
  "mcqQuestions": [
    {
      "title": "Question",
      "description": "Details",
      "points": 10,
      "options": [
        {"text": "Option A", "isCorrect": false},
        {"text": "Option B", "isCorrect": true}
      ]
    }
  ],
  "codingQuestions": [
    {
      "title": "Problem",
      "description": "Statement",
      "points": 20,
      "constraints": "1 <= n <= 10^5",
      "starterCode": "def solution():\n    pass",
      "timeLimitMs": 2000,
      "testCases": [
        {"input": "5", "expectedOutput": "15", "isHidden": false}
      ]
    }
  ]
}
```

## Safety & Validation

✅ **File Type Validation**: Only .pdf and .csv accepted
✅ **Content Length Check**: Minimum 50 characters
✅ **JSON Parsing**: Removes markdown code blocks
✅ **Error Handling**: Comprehensive error messages
✅ **Draft Only**: Tests never auto-published
✅ **Admin Only**: Route protected by JWT + AdminOnly middleware
✅ **Transactional**: Database operations use transactions

## Integration with Existing System

- ✅ Uses existing `models.Test`, `models.TestQuestion`, etc.
- ✅ Uses existing admin authentication
- ✅ Redirects to existing test edit page
- ✅ Follows existing UI theme and patterns
- ✅ No changes to manual test creation flow

## Files Changed

1. **go-backend/handlers/ai_test_builder.go** (NEW)
   - 500+ lines of AI test generation logic

2. **go-backend/main.go**
   - Added route: `admin.POST("/ai/build-test", handlers.HandleAIBuildTest)`

3. **go-backend/go.mod** & **go-backend/go.sum**
   - Added dependency: `github.com/ledongthuc/pdf`

4. **frontend/app/admin/page.tsx**
   - Added AI modal state management
   - Added "AI BUILD TEST" button
   - Added upload modal UI
   - Added file upload handler

## Testing Checklist

- [ ] Upload PDF with MCQs → Generates test with MCQ questions
- [ ] Upload PDF with coding problems → Generates test with coding questions + testcases
- [ ] Upload PDF with only notes → Generates questions from notes
- [ ] Upload CSV with question bank → Parses and generates test
- [ ] Error handling: Invalid file type
- [ ] Error handling: Empty file
- [ ] Error handling: OpenAI API failure
- [ ] Verify test created as DRAFT
- [ ] Verify redirect to edit page
- [ ] Verify admin can edit before publishing

## Deployment Notes

**Render Environment:**
- ✅ OPENAI_API_KEY already configured
- ✅ OPENAI_MODEL already configured
- ✅ No additional setup required

**Dependencies:**
- Go: `github.com/ledongthuc/pdf` (auto-installed)
- Frontend: No new dependencies

## Future Enhancements (Optional)

- Support for DOCX files
- Batch upload (multiple files)
- AI model selection in UI
- Preview generated test before saving
- Edit AI prompt in admin settings
- Support for images in questions
- Language detection and translation

## Usage Example

1. Admin logs in
2. Goes to Admin Portal → Tests
3. Clicks "AI BUILD TEST"
4. Uploads `data_structures_notes.pdf`
5. AI generates:
   - Title: "Data Structures Assessment"
   - 4 MCQs on arrays, linked lists, trees
   - 2 coding problems with testcases
6. Admin reviews and edits
7. Publishes when ready

---

**Status**: ✅ Complete and deployed
**Commit**: `feat: add AI-powered test generation from PDF/CSV files`

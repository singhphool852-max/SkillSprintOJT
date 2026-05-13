# AI Test Builder - Precise Extraction Fix

## Problem Summary
The AI Build Test feature was:
1. **Inventing extra MCQ questions** not present in the PDF (generating 3-5 MCQs when PDF had only 1)
2. **Extracting only 2-4 testcases** instead of all 50 from the PDF table
3. **Wrong input format** for two-parameter problems like Two Sum

## Root Causes

### Bug 1: AI Invents Extra Questions
**Cause**: The OpenAI prompt told the AI to "generate questions" without explicitly instructing it to ONLY extract what's in the document.

**Fix**: Rewrote the prompt in `generateTestWithOpenAI()` to be a "precise extraction" prompt with explicit rules:
- "DO NOT invent new questions"
- "ONLY extract what is explicitly present"
- "Generate exactly as many questions as exist in the document"

### Bug 2: Only 2-4 Testcases Extracted
**Cause**: 
1. AI response was being truncated due to insufficient token limit
2. Temperature was too high (0.7) causing creative generation instead of precise extraction

**Fix**:
- Increased `max_tokens` from 4000 to 8000 to handle 50+ testcases
- Lowered `temperature` from 0.7 to 0.3 for more precise extraction
- Added explicit instruction: "If there are 50 testcases in the table, extract all 50"

### Bug 3: Testcase Input Format
**Cause**: Prompt didn't specify the exact format for two-parameter problems.

**Fix**: Added explicit format instruction in prompt:
```
For two-parameter problems like Two Sum where input has both an array and a target, format the input as:
"nums = [2, 7, 11, 15]\ntarget = 9"

And expected output as: "[0, 1]"
```

## Changes Made

### File: `go-backend/handlers/ai_test_builder.go`

1. **Added `log` import** for debugging output

2. **Added logging after PDF extraction**:
```go
log.Printf("[AI_BUILD] Extracted text length=%d, preview=%.500s", len(extractedText), extractedText)
```

3. **Rewrote OpenAI prompt** to be "precise extraction" focused:
   - Explicit "DO NOT invent" instructions
   - Clear rules for MCQ and coding question extraction
   - Critical rule for testcases: "extract ALL of them. Every single row."
   - Exact format specification for two-parameter inputs
   - Count validation: "Generate exactly as many questions as exist in the document"

4. **Updated OpenAI API call parameters**:
```go
"temperature": 0.3,  // Lower for precise extraction (was 0.7)
"max_tokens":  8000, // Increased to handle 50+ testcases (was 4000)
```

5. **Added logging after AI response**:
```go
log.Printf("[AI_BUILD] AI response length=%d", len(content))
if len(content) < 500 {
    log.Printf("[AI_BUILD] WARNING: Response may be truncated: %s", content)
}
```

6. **Added extraction summary logging**:
```go
log.Printf("[AI_BUILD] Extracted %d MCQ questions, %d coding questions", 
    len(aiTest.MCQQuestions), len(aiTest.CodingQuestions))
for i, coding := range aiTest.CodingQuestions {
    log.Printf("[AI_BUILD] Coding question %d has %d testcases", i+1, len(coding.TestCases))
}
```

7. **Added testcase save logging**:
```go
log.Printf("[AI_BUILD] Saved %d testcases for coding question '%s'", testcaseCount, coding.Title)
```

## Testcase Saving Logic (Already Correct)

The `createDraftTestFromAI()` function correctly saves ALL testcases:
```go
// Create testcases
testcaseCount := 0
for _, tc := range coding.TestCases {
    testcase := models.TestCase{
        ID:             uuid.New().String(),
        QuestionID:     questionID,
        Input:          tc.Input,
        ExpectedOutput: tc.ExpectedOutput,
        IsHidden:       tc.IsHidden,
    }
    if err := tx.Create(&testcase).Error; err != nil {
        tx.Rollback()
        return "", err
    }
    testcaseCount++
}
```

This iterates over ALL testcases in `coding.TestCases` array, so if the AI returns 50 testcases, all 50 will be saved.

## PDF Extraction (Already Correct)

The `extractTextFromPDF()` function correctly reads ALL pages:
```go
numPages := reader.NumPage()
for i := 1; i <= numPages; i++ {
    page := reader.Page(i)
    // ... extract text from each page
}
```

## Expected Result After Fix

When uploading a PDF with:
- 1 MCQ question about time complexity
- 1 Coding question (Two Sum) with 50 testcases in a table

The AI should extract EXACTLY:
- **1 MCQ**: "What is the time complexity of find_duplicates?"
- **1 Coding**: "Two Sum" with 50 testcases
  - First 2-3 testcases: `isHidden: false` (sample)
  - Remaining 47-48 testcases: `isHidden: true` (hidden)

**Total questions created**: 2  
**Total testcases created**: 50

## Testing Instructions

1. Upload a PDF with the test content
2. Check backend logs for:
   ```
   [AI_BUILD] Extracted text length=... preview=...
   [AI_BUILD] AI response length=...
   [AI_BUILD] Extracted 1 MCQ questions, 1 coding questions
   [AI_BUILD] Coding question 1 has 50 testcases
   [AI_BUILD] Saved 50 testcases for coding question 'Two Sum'
   ```
3. Verify in database:
   - 2 questions created (1 MCQ + 1 coding)
   - 50 testcases created for the coding question
   - Testcase input format: `nums = [2, 7, 11, 15]\ntarget = 9`
   - Expected output format: `[0, 1]` (with spaces)

## Notes

- The AI uses OpenAI GPT-4o-mini (not Gemini) for test generation
- Tests are always created as DRAFT (`isPublished: false`)
- Topic validation ensures valid topicId or falls back to first available topic
- All changes preserve existing functionality for CSV uploads

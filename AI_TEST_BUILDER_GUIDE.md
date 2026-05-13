# AI Test Builder - User Guide

## Quick Start

### For Admins

1. **Navigate to Admin Portal**
   - Go to `/admin` page
   - You'll see all existing tests

2. **Click "AI BUILD TEST" Button**
   - Located in the top-right corner (cyan/teal button)
   - Next to the "CREATE TEST" button

3. **Upload Your File**
   - Click "Choose File" or drag & drop
   - Supported formats: **PDF** or **CSV**
   - File should contain:
     - MCQ questions with options
     - Coding problems with examples
     - Study notes (AI will generate questions)
     - Question banks

4. **Generate Test**
   - Click "GENERATE TEST" button
   - Wait 10-30 seconds (AI processing)
   - You'll be redirected to the test edit page

5. **Review & Edit**
   - Check generated questions
   - Edit titles, descriptions, points
   - Add/remove testcases
   - Adjust difficulty and duration

6. **Publish**
   - Click "PUBLISH" when ready
   - Test becomes available to students

---

## File Format Examples

### PDF Format

**Example 1: MCQ Questions**
```
1. What is the time complexity of binary search?
   A) O(n)
   B) O(log n) ✓
   C) O(n²)
   D) O(1)

2. Which data structure uses LIFO?
   A) Queue
   B) Stack ✓
   C) Array
   D) Tree
```

**Example 2: Coding Problems**
```
Problem: Two Sum
Given an array of integers, return indices of two numbers that add up to target.

Example:
Input: nums = [2,7,11,15], target = 9
Output: [0,1]

Constraints:
- 2 <= nums.length <= 10^4
- -10^9 <= nums[i] <= 10^9
```

**Example 3: Study Notes**
```
Binary Search Trees (BST)

A BST is a tree where:
- Left subtree contains nodes with keys less than parent
- Right subtree contains nodes with keys greater than parent
- Both subtrees are also BSTs

Operations:
- Search: O(log n) average, O(n) worst
- Insert: O(log n) average, O(n) worst
- Delete: O(log n) average, O(n) worst
```

### CSV Format

**Example: Question Bank**
```csv
Type,Question,Option A,Option B,Option C,Option D,Correct Answer
MCQ,What is 2+2?,3,4,5,6,B
MCQ,Capital of France?,London,Paris,Berlin,Rome,B
Coding,Reverse a string,,,,,
```

---

## What AI Generates

### For MCQ Questions
- ✅ Question title and description
- ✅ 4 options (A, B, C, D)
- ✅ Correct answer marked
- ✅ Points (default: 10)

### For Coding Questions
- ✅ Problem title and statement
- ✅ Constraints
- ✅ Starter code template
- ✅ Sample testcases (visible)
- ✅ Hidden testcases (for judging)
- ✅ Time limit (default: 2000ms)
- ✅ Points (20-50 based on difficulty)

### For Study Notes
- ✅ Extracts key concepts
- ✅ Generates MCQs from content
- ✅ Creates coding problems if applicable
- ✅ Adds explanations

---

## Tips for Best Results

### 1. **Clear Formatting**
- Use numbered questions
- Mark correct answers clearly (✓, *, or "Correct:")
- Separate questions with blank lines

### 2. **Include Examples**
- For coding problems, provide input/output examples
- AI uses these to generate testcases

### 3. **Specify Constraints**
- Add constraint sections for coding problems
- Helps AI generate appropriate testcases

### 4. **Mix Question Types**
- Combine MCQs and coding in one file
- AI will detect and separate them

### 5. **File Size**
- Keep files under 5MB for faster processing
- Split large question banks into multiple tests

---

## Troubleshooting

### "Failed to parse PDF"
- **Cause**: PDF is image-based (scanned)
- **Solution**: Use text-based PDFs or convert to text first

### "Extracted text is too short"
- **Cause**: File has less than 50 characters
- **Solution**: Add more content or check file format

### "AI generation failed"
- **Cause**: OpenAI API error or timeout
- **Solution**: Try again or simplify the content

### "Invalid JSON response"
- **Cause**: AI returned malformed JSON
- **Solution**: Retry with clearer formatting

### Test has wrong questions
- **Cause**: AI misinterpreted content
- **Solution**: Edit questions manually or reformat source file

---

## Advanced Usage

### Custom Prompts (Future)
You can customize AI behavior by:
- Specifying difficulty in filename: `easy_test.pdf`
- Adding metadata in first line: `DIFFICULTY: hard`
- Including topic tags: `#DataStructures #Algorithms`

### Batch Generation (Future)
- Upload multiple files at once
- Generate test series automatically
- Merge related topics

### Language Support (Future)
- Upload files in any language
- AI detects and generates in same language
- Automatic translation available

---

## Best Practices

### ✅ DO
- Review all generated questions before publishing
- Test coding problems with sample inputs
- Verify correct answers for MCQs
- Adjust points based on difficulty
- Add custom testcases if needed

### ❌ DON'T
- Publish without reviewing
- Trust AI 100% for correctness
- Upload copyrighted content
- Use for high-stakes exams without verification
- Ignore validation errors

---

## Example Workflow

**Scenario**: Creating a Data Structures test

1. **Prepare Content**
   - Collect lecture notes (PDF)
   - Add practice problems
   - Include sample solutions

2. **Upload to AI Builder**
   - Click "AI BUILD TEST"
   - Upload `ds_notes.pdf`
   - Wait for generation

3. **Review Generated Test**
   - Check 5 MCQs on arrays, trees, graphs
   - Verify 2 coding problems with testcases
   - Edit any incorrect options

4. **Enhance Test**
   - Add 1 more hard coding problem manually
   - Adjust time limit to 90 minutes
   - Set difficulty to "medium"

5. **Publish**
   - Click "PUBLISH"
   - Activate for students
   - Monitor submissions

---

## FAQ

**Q: Can I edit AI-generated tests?**
A: Yes! All tests are created as drafts. Edit freely before publishing.

**Q: What if AI generates wrong answers?**
A: Always review and correct. AI is a starting point, not final authority.

**Q: Can I use images in questions?**
A: Not yet. Text-only for now. Images coming in future update.

**Q: How long does generation take?**
A: 10-30 seconds depending on file size and complexity.

**Q: Is my uploaded file stored?**
A: No. Files are processed and discarded immediately.

**Q: Can I regenerate if I don't like the result?**
A: Yes! Delete the draft and upload again with better formatting.

**Q: Does it work with handwritten notes?**
A: No. Only typed/digital text. Handwritten PDFs won't work.

**Q: Can I upload Word documents?**
A: Not yet. Convert to PDF first. DOCX support coming soon.

---

## Support

For issues or questions:
- Check this guide first
- Review error messages carefully
- Try reformatting your source file
- Contact admin support if problem persists

---

**Happy Test Building! 🚀**

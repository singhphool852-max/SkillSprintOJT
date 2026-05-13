package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/ledongthuc/pdf"
)

// ──────────────────────────────────────────────
// AI Test Generation Structs
// ──────────────────────────────────────────────

type AITestResponse struct {
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	TopicID         string          `json:"topicId"`
	Difficulty      string          `json:"difficulty"`
	DurationMinutes int             `json:"durationMinutes"`
	MCQQuestions    []AIMCQQuestion `json:"mcqQuestions"`
	CodingQuestions []AICodingQuestion `json:"codingQuestions"`
}

type AIMCQQuestion struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Points      int         `json:"points"`
	Options     []AIOption  `json:"options"`
}

type AIOption struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}

type AICodingQuestion struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Points      int          `json:"points"`
	Constraints string       `json:"constraints"`
	StarterCode string       `json:"starterCode"`
	TimeLimitMs int          `json:"timeLimitMs"`
	TestCases   []AITestCase `json:"testCases"`
}

type AITestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expectedOutput"`
	IsHidden       bool   `json:"isHidden"`
}

// ──────────────────────────────────────────────
// HandleAIBuildTest → POST /api/admin/ai/build-test
// ──────────────────────────────────────────────
func HandleAIBuildTest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get topicId from form (optional)
	topicID := c.PostForm("topicId")

	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Validate file type
	filename := header.Filename
	var extractedText string

	if strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		extractedText, err = extractTextFromPDF(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse PDF: %v", err)})
			return
		}
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		extractedText, err = extractTextFromCSV(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse CSV: %v", err)})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF and CSV files are supported"})
		return
	}

	if len(extractedText) < 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Extracted text is too short. Please upload a file with more content."})
		return
	}

	// Log extracted text for debugging
	log.Printf("[AI_BUILD] Extracted text length=%d, preview=%.500s", len(extractedText), extractedText)

	// Call OpenAI to generate test structure
	aiResponse, err := generateTestWithOpenAI(extractedText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI generation failed: %v", err)})
		return
	}

	// Create draft test in database
	testID, err := createDraftTestFromAI(aiResponse, userID.(string), topicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create test: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test generated successfully",
		"testId":  testID,
		"test":    aiResponse,
	})
}

// ──────────────────────────────────────────────
// Extract text from PDF
// ──────────────────────────────────────────────
func extractTextFromPDF(file io.Reader) (string, error) {
	// Read file into memory
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", err
	}

	// Parse PDF
	reader, err := pdf.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return "", err
	}

	var text strings.Builder
	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		text.WriteString(pageText)
		text.WriteString("\n")
	}

	return text.String(), nil
}

// ──────────────────────────────────────────────
// Extract text from CSV
// ──────────────────────────────────────────────
func extractTextFromCSV(file io.Reader) (string, error) {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	var text strings.Builder
	for _, record := range records {
		text.WriteString(strings.Join(record, " | "))
		text.WriteString("\n")
	}

	return text.String(), nil
}

// ──────────────────────────────────────────────
// Generate test structure using OpenAI
// ──────────────────────────────────────────────
func generateTestWithOpenAI(content string) (*AITestResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_MODEL")
	
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}
	
	if model == "" {
		model = "gpt-4o-mini" // Default model
	}

	// Construct prompt - PRECISE EXTRACTION, not generation
	prompt := fmt.Sprintf(`You are a precise test extractor AI.
You are given text extracted from a PDF document.
Your job is to EXTRACT questions EXACTLY as written.

DO NOT invent new questions.
DO NOT add questions that are not in the document.
ONLY extract what is explicitly present.

EXTRACTION RULES:

1. MCQ QUESTIONS
Extract ONLY MCQ questions that are explicitly written in the document with answer choices (A), (B), (C), (D).
For each MCQ identify the correct answer from context or from any answer key present.

2. CODING QUESTIONS
Extract ONLY coding problems explicitly written in the document.
Include the full problem statement, constraints, and function signature exactly as written.

3. TESTCASES — CRITICAL RULE
If the document contains a table of testcases with columns like: #, nums input, target, expected output —
extract ALL of them. Every single row.
Do NOT skip any testcase.
Do NOT limit to 2 or 5 testcases.
If there are 50 testcases in the table, extract all 50.

For two-parameter problems like Two Sum where input has both an array and a target, format the input as:
"nums = [2, 7, 11, 15]\ntarget = 9"

And expected output as: "[0, 1]"

4. COUNT
Generate exactly as many questions as exist in the document.
If document has 1 MCQ and 1 coding question, return exactly 2.
If document has 3 MCQs, return exactly 3.
Do NOT add or remove any questions.

Return ONLY this JSON object, no markdown, no extra text:

{
  "title": "short title based on document content",
  "description": "2 sentence description of what this test covers",
  "difficulty": "easy|medium|hard",
  "durationMinutes": 60,
  "mcqQuestions": [
    {
      "title": "exact question text from document",
      "description": "full question with all details",
      "points": 10,
      "options": [
        {"text": "exact option A", "isCorrect": false},
        {"text": "exact option B", "isCorrect": true},
        {"text": "exact option C", "isCorrect": false},
        {"text": "exact option D", "isCorrect": false}
      ]
    }
  ],
  "codingQuestions": [
    {
      "title": "problem title exactly as written",
      "description": "full problem statement exactly as written",
      "points": 20,
      "constraints": "constraints exactly as written",
      "starterCode": "def two_sum(nums, target):\n    pass",
      "timeLimitMs": 2000,
      "testCases": [
        {"input": "nums = [2, 7, 11, 15]\ntarget = 9", "expectedOutput": "[0, 1]", "isHidden": false},
        {"input": "nums = [3, 2, 4]\ntarget = 6", "expectedOutput": "[1, 2]", "isHidden": false},
        {"input": "nums = [3, 3]\ntarget = 6", "expectedOutput": "[0, 1]", "isHidden": true}
      ]
    }
  ]
}

IMPORTANT:
- First 2-3 testcases should have isHidden: false (sample testcases)
- ALL remaining testcases should have isHidden: true (hidden testcases)
- Extract ALL testcases from any table in the document
- Ensure all JSON is valid and properly escaped
- Return ONLY the JSON, no markdown code blocks

DOCUMENT CONTENT:
%s`, content)

	// Call OpenAI API with increased token limit for large testcase lists
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a precise test extraction assistant. Extract questions and testcases EXACTLY as written in the document. Return ONLY valid JSON, no markdown."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3, // Lower temperature for more precise extraction
		"max_tokens":  8000, // Increased to handle 50+ testcases
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse AI response
	content = openAIResp.Choices[0].Message.Content
	
	// Log AI response for debugging
	log.Printf("[AI_BUILD] AI response length=%d", len(content))
	if len(content) < 500 {
		log.Printf("[AI_BUILD] WARNING: Response may be truncated: %s", content)
	}
	
	// Remove markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var aiTest AITestResponse
	if err := json.Unmarshal([]byte(content), &aiTest); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v - Content: %s", err, content)
	}

	// Log what was extracted
	log.Printf("[AI_BUILD] Extracted %d MCQ questions, %d coding questions", 
		len(aiTest.MCQQuestions), len(aiTest.CodingQuestions))
	for i, coding := range aiTest.CodingQuestions {
		log.Printf("[AI_BUILD] Coding question %d has %d testcases", i+1, len(coding.TestCases))
	}

	return &aiTest, nil
}

// ──────────────────────────────────────────────
// Create draft test from AI response
// ──────────────────────────────────────────────
func createDraftTestFromAI(aiTest *AITestResponse, createdBy string, requestedTopicID string) (string, error) {
	testID := uuid.New().String()
	now := time.Now()
	startTime := now.Add(24 * time.Hour) // Default: start tomorrow

	// Determine valid topicId
	var topicID string
	
	// Priority 1: Use topicId from request if provided
	if requestedTopicID != "" {
		var topic models.Topic
		if err := database.DB.Where("id = ?", requestedTopicID).First(&topic).Error; err != nil {
			return "", fmt.Errorf("selected topic not found: %v", err)
		}
		topicID = topic.ID
	} else {
		// Priority 2: Get first available topic as fallback
		var defaultTopic models.Topic
		if err := database.DB.First(&defaultTopic).Error; err != nil {
			return "", fmt.Errorf("no topics found in database. Please create a topic first")
		}
		topicID = defaultTopic.ID
	}

	// Create test
	test := models.Test{
		ID:              testID,
		Title:           aiTest.Title,
		Description:     aiTest.Description + " (AI Generated - Review before publishing)",
		TopicID:         topicID, // Use verified topicID
		Difficulty:      aiTest.Difficulty,
		StartTime:       &startTime,
		DurationSeconds: aiTest.DurationMinutes * 60,
		IsPublished:     false, // Always draft
		IsActive:        false,
		CreatedBy:       createdBy,
	}

	tx := database.DB.Begin()

	if err := tx.Create(&test).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	position := 1

	// Create MCQ questions
	for _, mcq := range aiTest.MCQQuestions {
		questionID := uuid.New().String()
		question := models.TestQuestion{
			ID:          questionID,
			TestID:      testID,
			Type:        "mcq",
			Position:    position,
			Title:       mcq.Title,
			Description: mcq.Description,
			Points:      mcq.Points,
		}

		if err := tx.Create(&question).Error; err != nil {
			tx.Rollback()
			return "", err
		}

		// Create options
		for _, opt := range mcq.Options {
			option := models.TestMCQOption{
				ID:         uuid.New().String(),
				QuestionID: questionID,
				OptionText: opt.Text,
				IsCorrect:  opt.IsCorrect,
			}
			if err := tx.Create(&option).Error; err != nil {
				tx.Rollback()
				return "", err
			}
		}

		position++
	}

	// Create coding questions
	for _, coding := range aiTest.CodingQuestions {
		questionID := uuid.New().String()
		question := models.TestQuestion{
			ID:          questionID,
			TestID:      testID,
			Type:        "coding",
			Position:    position,
			Title:       coding.Title,
			Description: coding.Description,
			Points:      coding.Points,
		}

		if err := tx.Create(&question).Error; err != nil {
			tx.Rollback()
			return "", err
		}

		// Create coding detail
		detail := models.TestCodingDetail{
			ID:          uuid.New().String(),
			QuestionID:  questionID,
			Constraints: coding.Constraints,
			StarterCode: coding.StarterCode,
			TimeLimitMs: coding.TimeLimitMs,
		}
		if err := tx.Create(&detail).Error; err != nil {
			tx.Rollback()
			return "", err
		}

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
		
		log.Printf("[AI_BUILD] Saved %d testcases for coding question '%s'", testcaseCount, coding.Title)

		position++
	}

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	return testID, nil
}

package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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

	// Construct prompt
	prompt := fmt.Sprintf(`You are an expert test creator for a coding assessment platform called SkillSprint.

Analyze the following content and generate a complete test with MCQ and coding questions.

Content:
%s

Generate a JSON response with this EXACT structure (no markdown, no explanations, ONLY valid JSON):

{
  "title": "Test title (concise, under 50 chars)",
  "description": "Brief description (1-2 sentences)",
  "topicId": "",
  "difficulty": "easy|medium|hard",
  "durationMinutes": 60,
  "mcqQuestions": [
    {
      "title": "Question title",
      "description": "Question text with details",
      "points": 10,
      "options": [
        {"text": "Option A", "isCorrect": false},
        {"text": "Option B", "isCorrect": true},
        {"text": "Option C", "isCorrect": false},
        {"text": "Option D", "isCorrect": false}
      ]
    }
  ],
  "codingQuestions": [
    {
      "title": "Problem title",
      "description": "Problem statement with examples",
      "points": 20,
      "constraints": "1 <= n <= 10^5\n1 <= arr[i] <= 10^9",
      "starterCode": "def solution(arr):\n    pass",
      "timeLimitMs": 2000,
      "testCases": [
        {"input": "5\n1 2 3 4 5", "expectedOutput": "15", "isHidden": false},
        {"input": "3\n10 20 30", "expectedOutput": "60", "isHidden": true}
      ]
    }
  ]
}

Rules:
- Generate 3-5 MCQ questions if content has theory/concepts
- Generate 1-3 coding questions if content has algorithms/problems
- If content is only notes, create questions FROM the notes
- If content lacks testcases, GENERATE them
- Each MCQ must have exactly 4 options with ONE correct answer
- Each coding question must have at least 2 testcases (1 visible, 1 hidden)
- Points: MCQ=10, Easy Coding=20, Medium=30, Hard=50
- Ensure all JSON is valid and properly escaped
- Return ONLY the JSON, no markdown code blocks`, content)

	// Call OpenAI API
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a test generation assistant. Return ONLY valid JSON, no markdown."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  4000,
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
	
	// Remove markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var aiTest AITestResponse
	if err := json.Unmarshal([]byte(content), &aiTest); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v - Content: %s", err, content)
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
		}

		position++
	}

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	return testID, nil
}

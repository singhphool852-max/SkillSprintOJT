package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// ErrGeminiRateLimit is returned when Gemini responds with 429 RESOURCE_EXHAUSTED.
var ErrGeminiRateLimit = errors.New("Gemini free-tier quota exceeded. Please wait and try again.")

// isGeminiRateLimited checks if a Gemini response indicates rate limiting.
func isGeminiRateLimited(statusCode int, body []byte) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	// Gemini may also return 429 info inside a non-429 status body.
	var apiErr geminiAPIError
	if err := json.Unmarshal(body, &apiErr); err == nil {
		if strings.Contains(strings.ToUpper(apiErr.Error.Status), "RESOURCE_EXHAUSTED") {
			return true
		}
	}
	return false
}

type geminiAPIError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// AIResponse is used by the answer evaluation endpoint.
type AIResponse struct {
	Score                 int    `json:"score"`
	IsCorrect             bool   `json:"isCorrect"`
	Feedback              string `json:"feedback"`
	Explanation           string `json:"explanation"`
	ImprovementSuggestion string `json:"improvementSuggestion"`
}

// EvaluateAnswer scores a user's answer against the correct answer using Gemini or OpenAI.
func EvaluateAnswer(question, correctAnswer, userAnswer string, maxScore int) (*AIResponse, error) {
	// 1. Try Gemini first (as it's the primary provider for this project)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" {
		return evaluateWithGemini(question, correctAnswer, userAnswer, maxScore, apiKey)
	}

	// 2. Fallback to OpenAI if Gemini key is missing
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey != "" {
		return evaluateWithOpenAI(question, correctAnswer, userAnswer, maxScore, openAIKey)
	}

	// 3. Last resort: Mock response
	score := 0
	if len(strings.TrimSpace(userAnswer)) > 20 {
		score = maxScore - 2
	}
	return &AIResponse{
		Score:                 score,
		IsCorrect:             score > maxScore/2,
		Feedback:              "Evaluated by SkillSprint AI (System Mock). Technical connectivity to neural models is pending.",
		Explanation:           fmt.Sprintf("Reference Answer: %s", correctAnswer),
		ImprovementSuggestion: "Ensure your answer contains specific technical terminology relevant to the prompt.",
	}, nil
}

func evaluateWithGemini(question, correctAnswer, userAnswer string, maxScore int, apiKey string) (*AIResponse, error) {
	prompt := fmt.Sprintf(`You are a technical examiner.
Evaluate the User Answer against the Correct Answer for the given Question.

Question: %s
Correct Answer: %s
User Answer: %s

Task:
1. Assign a score out of %d.
2. Determine if it is fundamentally correct (isCorrect).
3. Provide concise feedback.
4. Provide the correct technical explanation.
5. Suggest one specific improvement.

Return ONLY a JSON object with this structure:
{
  "score": integer,
  "isCorrect": boolean,
  "feedback": "string",
  "explanation": "string",
  "improvementSuggestion": "string"
}

No markdown. No extra text.`, question, correctAnswer, userAnswer, maxScore)

	type Part struct{ Text string `json:"text"` }
	type Content struct{ Parts []Part `json:"parts"` }
	type RequestBody struct{ Contents []Content `json:"contents"` }

	reqBody := RequestBody{Contents: []Content{{Parts: []Part{{Text: prompt}}}}}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("gemini network error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if isGeminiRateLimited(resp.StatusCode, respBody) {
			return nil, ErrGeminiRateLimit
		}
		return nil, fmt.Errorf("gemini evaluation error: %d", resp.StatusCode)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text
	cleaned := cleanGeminiResponse(rawText)

	var eval AIResponse
	if err := json.Unmarshal([]byte(cleaned), &eval); err != nil {
		log.Printf("[AI_EVAL_ERROR] failed to parse JSON: %s", cleaned)
		return nil, fmt.Errorf("failed to parse AI evaluation result")
	}

	return &eval, nil
}

func evaluateWithOpenAI(question, correctAnswer, userAnswer string, maxScore int, apiKey string) (*AIResponse, error) {
	url := "https://api.openai.com/v1/chat/completions"
	prompt := fmt.Sprintf(`You are an evaluator.
Question: %s
Correct Answer: %s
User Answer: %s

Task: Evaluate the User Answer against the Correct Answer.
Return a JSON object with:
- score: out of %d
- isCorrect: boolean
- feedback: short text on quality
- explanation: the correct reasoning
- improvementSuggestion: how to get a better score

Respond ONLY with JSON.`, question, correctAnswer, userAnswer, maxScore)

	payload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI service error: %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var eval AIResponse
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &eval); err != nil {
		return nil, err
	}

	return &eval, nil
}

// GenerateSimilarQuestions creates variations of a specific question to help students master a concept they struggled with.
func GenerateSimilarQuestions(originalPrompt string, topic string, difficulty string, count int) ([]GeneratedQuestion, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not found")
	}

	prompt := fmt.Sprintf(`You are an adaptive learning assistant.
Original Question the student failed: "%s"
Topic: %s
Difficulty: %s

Task: Generate EXACTLY %d similar but unique MCQ questions that test the SAME concept or a closely related concept.
Return ONLY a JSON array of objects with structure:
[
  {
    "type": "mcq",
    "prompt": "new unique question text",
    "options": ["A", "B", "C", "D"],
    "answer": "correct option text",
    "explanation": "clear technical explanation",
    "difficulty": "%s"
  }
]`, originalPrompt, topic, difficulty, count, difficulty)

	type Part struct{ Text string `json:"text"` }
	type Content struct{ Parts []Part `json:"parts"` }
	type RequestBody struct{ Contents []Content `json:"contents"` }

	reqBody := RequestBody{Contents: []Content{{Parts: []Part{{Text: prompt}}}}}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error: %d", resp.StatusCode)
	}

	type GeminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	var geminiResp GeminiResponse
	json.Unmarshal(respBody, &geminiResp)
	
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text
	cleaned := cleanGeminiResponse(rawText)

	var questions []GeneratedQuestion
	if err := json.Unmarshal([]byte(cleaned), &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// ---------- Gemini Question Generation ----------

// GeneratedQuestion is the struct returned by GenerateQuestions.
type GeneratedQuestion struct {
	Type          string   `json:"type"`              // mcq, debug_code, fix_code, logic_explanation
	Prompt        string   `json:"prompt"`
	Options       []string `json:"options,omitempty"` // for mcq
	Answer        string   `json:"answer,omitempty"`  // from training
	CorrectAnswer string   `json:"correctAnswer,omitempty"`
	Explanation   string   `json:"explanation"`
	MaxScore      int      `json:"maxScore,omitempty"`
	Difficulty    string   `json:"difficulty,omitempty"`
}

// GenerateQuestions calls the Gemini 1.5 Flash API to produce MCQ questions.
func GenerateQuestions(topic, difficulty string, count int, excludePrompts []string) ([]GeneratedQuestion, error) {
	// 1. Read API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("[AI] CRITICAL: GEMINI_API_KEY is not set in environment")
		return nil, fmt.Errorf("GEMINI_API_KEY not found in environment")
	}

	log.Printf("[AI] Start: topic=%s difficulty=%s count=%d", topic, difficulty, count)

	// 2. Format subtopic variety guidance
	topicLower := strings.ToLower(topic)
	varietyGuidance := ""
	switch {
	case strings.Contains(topicLower, "dbms") || strings.Contains(topicLower, "database"):
		varietyGuidance = "Cover: normalization, indexing, joins, transactions, ACID, keys, clustered/non-clustered index, query output, isolation levels."
	case strings.Contains(topicLower, "dsa") || strings.Contains(topicLower, "algorithm") || strings.Contains(topicLower, "data structure"):
		varietyGuidance = "Cover: arrays, stacks, queues, trees, graphs, sorting, recursion, DP, hashing, complexity."
	case strings.Contains(topicLower, "os") || strings.Contains(topicLower, "operating"):
		varietyGuidance = "Cover: processes, threads, scheduling, deadlock, paging, memory, synchronization, semaphores."
	case strings.Contains(topicLower, "js") || strings.Contains(topicLower, "javascript"):
		varietyGuidance = "Cover: closures, async/await, event loop, hoisting, promises, scope, DOM, prototypes."
	case strings.Contains(topicLower, "aptitude") || strings.Contains(topicLower, "math"):
		varietyGuidance = "Cover: probability, percentages, ratios, time-speed-distance, permutations, logic."
	}

	excludeText := ""
	if len(excludePrompts) > 0 {
		excludeText = fmt.Sprintf("\nIMPORTANT: DO NOT generate questions similar to these existing prompts: %s", strings.Join(excludePrompts, " | "))
	}

	// 3. Build the strict prompt
	prompt := fmt.Sprintf(`Generate EXACTLY %d MCQ questions for topic: %s, difficulty: %s.
%s

Return ONLY JSON array:
[
  {
    "type": "mcq",
    "prompt": "Question text here (clear and technical)",
    "options": ["Option A", "Option B", "Option C", "Option D"],
    "answer": "Exact text of the correct option (must match one of the options)",
    "explanation": "Brief technical explanation",
    "difficulty": "%s"
  }
]

Rules:
- Generate diverse questions covering different subtopics.
- Do not repeat the same concept phrased differently.
- No markdown formatting.
- No text outside JSON.
- No `+"`"+`json code fences.%s`, count, topic, difficulty, varietyGuidance, difficulty, excludeText)

	// 3. Build the Gemini API request payload
	type Part struct {
		Text string `json:"text"`
	}
	type Content struct {
		Parts []Part `json:"parts"`
	}
	type RequestBody struct {
		Contents []Content `json:"contents"`
	}

	reqBody := RequestBody{
		Contents: []Content{
			{Parts: []Part{{Text: prompt}}},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 4. POST to gemini-2.5-flash with 30s timeout
	// Verifying availability: gemini-2.5-flash is the stable multimodal model in this environment.
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s",
		apiKey,
	)

	// Diagnostic masking: log first 4 and last 4 of key
	maskedKey := "EMPTY"
	if len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "...." + apiKey[len(apiKey)-4:]
	}
	log.Printf("[AI] Calling Gemini v1beta: model=gemini-2.5-flash target_count=%d key=%s", count, maskedKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[AI] ERROR: Network request failed: %v", err)
		return nil, fmt.Errorf("gemini network error: %w", err)
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read gemini response: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		if isGeminiRateLimited(resp.StatusCode, respBody) {
			log.Printf("[GEMINI_RATE_LIMIT] 429 on GenerateQuestions topic=%s", topic)
			return nil, ErrGeminiRateLimit
		}
		log.Printf("[AI] ERROR: Gemini API returned status %d. URL used: https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent", resp.StatusCode)
		return nil, fmt.Errorf("gemini API error: status %d", resp.StatusCode)
	}

	// 5. Decode the Gemini response envelope
	type GeminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		log.Printf("[AI] ERROR: Failed to decode response body: %v", err)
		return nil, fmt.Errorf("gemini decode error: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		log.Println("[AI] ERROR: Empty candidates in Gemini response")
		return nil, fmt.Errorf("gemini returned empty response")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text
	log.Printf("[AI] Raw response received. Length: %d bytes", len(rawText))

	// 6. Clean and extract JSON from the raw text
	cleaned := cleanGeminiResponse(rawText)
	log.Printf("[AI] Cleaned JSON length: %d bytes", len(cleaned))

	// 7. Parse into []GeneratedQuestion
	var rawQuestions []GeneratedQuestion
	if err := json.Unmarshal([]byte(cleaned), &rawQuestions); err != nil {
		log.Printf("[AI ERROR] Failed to unmarshal JSON: %v | Raw start: %.100s", err, cleaned)
		return nil, fmt.Errorf("AI generation failed: could not parse response")
	}

	// 8. Safely Filter & Validate
	var questions []GeneratedQuestion
	for _, q := range rawQuestions {
		if q.Prompt == "" || len(q.Options) < 2 || q.Answer == "" {
			log.Printf("[AI SKIP] Dropping malformed question: prompt_len=%d options=%d", len(q.Prompt), len(q.Options))
			continue
		}
		questions = append(questions, q)
	}

	log.Printf("[AI] Generation Summary | Requested: %d | AI Returned: %d | Validated: %d", count, len(rawQuestions), len(questions))
	return questions, nil
}

// SummarizeNotes generates a concise technical summary of the provided text.
func SummarizeNotes(text string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not found")
	}

	inputLen := len(text)
	log.Printf("[GEMINI_SUMMARY] input_text_len=%d", inputLen)
	if len(text) > 8000 {
		text = text[:8000]
	}
	log.Printf("[GEMINI_SUMMARY] truncated_text_len=%d", len(text))

	prompt := fmt.Sprintf("Summarize the following notes into concise bullet points covering only the key concepts.\n\nNOTES:\n%s", text)

	type Part struct {
		Text string `json:"text"`
	}
	type Content struct {
		Parts []Part `json:"parts"`
	}
	type RequestBody struct {
		Contents []Content `json:"contents"`
	}

	reqBody := RequestBody{Contents: []Content{{Parts: []Part{{Text: prompt}}}}}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read Gemini summary response: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		if isGeminiRateLimited(resp.StatusCode, responseBody) {
			log.Printf("[GEMINI_RATE_LIMIT] 429 on SummarizeNotes")
			return "", ErrGeminiRateLimit
		}
		log.Printf("[AI] Gemini Summary Error: %d | Body: %s", resp.StatusCode, string(responseBody))
		var apiErr geminiAPIError
		if err := json.Unmarshal(responseBody, &apiErr); err == nil {
			parts := []string{fmt.Sprintf("gemini API error: %d", resp.StatusCode)}
			if strings.TrimSpace(apiErr.Error.Status) != "" {
				parts = append(parts, apiErr.Error.Status)
			}
			if strings.TrimSpace(apiErr.Error.Message) != "" {
				parts = append(parts, apiErr.Error.Message)
			}
			return "", errors.New(strings.Join(parts, " - "))
		}
		return "", fmt.Errorf("gemini API error: %d", resp.StatusCode)
	}

	log.Printf("[GEMINI_SUMMARY_PREVIEW] %.300s", string(responseBody))
	log.Println("[GEMINI_SUMMARY_RAW]", string(responseBody))

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse Gemini summary response: %w", err)
	}

	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Gemini")
	}

	if len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned from Gemini")
	}

	summaryText := strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text)
	if summaryText == "" {
		return "", fmt.Errorf("empty summary returned from Gemini")
	}
	log.Printf("[GEMINI_SUMMARY] summary_len=%d", len(summaryText))

	return summaryText, nil
}

// GenerateQuestionsFromNotes derives MCQ questions from a technical summary.
func GenerateQuestionsFromNotes(summary string, count int, difficulty string) ([]GeneratedQuestion, error) {
	return GenerateQuestionsFromNotesWithMinMatches(summary, count, difficulty, 0)
}

// GenerateQuestionsFromNotesWithMinMatches derives MCQ questions from a technical summary.
// minKeywordMatches=0 enables adaptive grounding strictness.
func GenerateQuestionsFromNotesWithMinMatches(summary string, count int, difficulty string, minKeywordMatches int) ([]GeneratedQuestion, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not found")
	}

	prompt := fmt.Sprintf(`You are an expert technical interviewer AI.

You are given a SUMMARY extracted from a user's personal notes.

Your task is to generate EXACTLY %d high-quality MCQ (multiple choice) questions STRICTLY based on the provided summary.

CRITICAL RULES (MUST FOLLOW):

1. GROUNDING (VERY IMPORTANT)
- Every question MUST be directly based on concepts present in the summary.
- DO NOT use external knowledge unless it is clearly implied in the summary.
- DO NOT generate generic textbook questions.
- If a concept is not in the summary, DO NOT use it.

2. DIVERSITY
- Cover DIFFERENT concepts from the summary.
- Do NOT repeat the same idea in multiple questions.
- Maximum ONE question per concept.

3. QUESTION QUALITY
- Questions must test understanding, not just definitions.
- Include logic-based, scenario-based, or application-based questions where possible.
- Avoid trivial or overly obvious questions.

4. OPTIONS
- Each question must have EXACTLY 4 options.
- Only ONE correct answer.
- Options should be realistic and not obviously wrong.

5. EXPLANATION
- Each question MUST include an explanation.
- Explanation must reference the concept from the summary.
- Keep explanation clear and useful for learning.

6. FORMAT (STRICT JSON ONLY)
Return ONLY a JSON array. No markdown. No text.

Each object must follow EXACT structure:

[
  {
    "prompt": "question text",
    "options": ["A", "B", "C", "D"],
    "answer": "correct option text",
    "explanation": "clear explanation based on summary",
    "difficulty": "%s"
  }
]

7. FAILURE CONDITION
- If the summary is too short or unclear, still generate the BEST possible grounded questions.
- DO NOT return fewer than %d questions.

---

SUMMARY:
%s`, count, difficulty, count, summary)

	type Part struct {
		Text string `json:"text"`
	}
	type Content struct {
		Parts []Part `json:"parts"`
	}
	type RequestBody struct {
		Contents []Content `json:"contents"`
	}

	reqBody := RequestBody{Contents: []Content{{Parts: []Part{{Text: prompt}}}}}
	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read AI response body: %w", readErr)
	}
	if resp.StatusCode != http.StatusOK {
		if isGeminiRateLimited(resp.StatusCode, rawBody) {
			log.Printf("[GEMINI_RATE_LIMIT] 429 on GenerateQuestionsFromNotes")
			return nil, ErrGeminiRateLimit
		}
		log.Printf("[NOTES][generate] Gemini notes generation status=%d preview=%.200s", resp.StatusCode, string(rawBody))
		return nil, fmt.Errorf("gemini API error: status %d", resp.StatusCode)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(rawBody, &result); err != nil {
		log.Printf("[NOTES][generate] malformed gemini envelope preview=%.200s", string(rawBody))
		return nil, fmt.Errorf("failed to parse AI-generated notes questions")
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		log.Printf("[NOTES][generate] empty gemini candidates preview=%.200s", string(rawBody))
		return nil, fmt.Errorf("empty response from gemini")
	}

	rawText := result.Candidates[0].Content.Parts[0].Text
	if strings.TrimSpace(rawText) == "" {
		log.Printf("[NOTES][generate] empty candidate text preview=%.200s", string(rawBody))
		return nil, fmt.Errorf("empty response from gemini")
	}
	cleaned := cleanGeminiResponse(rawText)

	var rawQuestions []GeneratedQuestion
	if err := json.Unmarshal([]byte(cleaned), &rawQuestions); err != nil {
		log.Printf("[NOTES][generate] failed parsing questions cleaned_preview=%.200s", cleaned)
		return nil, fmt.Errorf("failed to parse AI-generated notes questions")
	}

	summaryKeywords := buildSummaryKeywordSet(summary)
	requiredMatches := minKeywordMatches
	if requiredMatches <= 0 {
		requiredMatches = adaptiveMinKeywordMatches(summaryKeywords)
	}
	valid := make([]GeneratedQuestion, 0, len(rawQuestions))
	for _, q := range rawQuestions {
		if strings.TrimSpace(q.Prompt) == "" ||
			len(q.Options) < 4 ||
			strings.TrimSpace(q.Answer) == "" ||
			strings.TrimSpace(q.Explanation) == "" {
			continue
		}
		if !isGroundedInSummary(q.Prompt, q.Explanation, summaryKeywords, requiredMatches) {
			continue
		}
		if strings.TrimSpace(q.Difficulty) == "" {
			q.Difficulty = difficulty
		}
		valid = append(valid, q)
	}

	if len(valid) == 0 {
		log.Printf("[NOTES][generate] validation produced zero questions raw_count=%d", len(rawQuestions))
		return nil, fmt.Errorf("no valid questions generated from notes")
	}

	return valid, nil
}

func buildSummaryKeywordSet(summary string) map[string]struct{} {
	re := regexp.MustCompile(`[a-zA-Z]{4,}`)
	words := re.FindAllString(strings.ToLower(summary), -1)
	keywords := make(map[string]struct{}, len(words))
	stop := map[string]struct{}{
		"that": {}, "with": {}, "from": {}, "this": {}, "these": {}, "those": {},
		"have": {}, "will": {}, "into": {}, "about": {}, "such": {}, "their": {},
		"there": {}, "which": {}, "where": {}, "when": {}, "while": {}, "using": {},
		"used": {}, "also": {}, "only": {}, "below": {}, "based": {}, "content": {},
		"summary": {}, "concept": {}, "concepts": {}, "technical": {}, "should": {},
	}
	for _, w := range words {
		if _, blocked := stop[w]; blocked {
			continue
		}
		keywords[w] = struct{}{}
	}
	return keywords
}

func adaptiveMinKeywordMatches(summaryKeywords map[string]struct{}) int {
	// Lighter default acceptance, stricter only when summary is rich.
	if len(summaryKeywords) >= 60 {
		return 2
	}
	return 1
}

func isGroundedInSummary(prompt, explanation string, summaryKeywords map[string]struct{}, requiredMatches int) bool {
	if requiredMatches < 1 {
		requiredMatches = 1
	}
	re := regexp.MustCompile(`[a-zA-Z]{4,}`)
	text := strings.ToLower(prompt + " " + explanation)
	words := re.FindAllString(text, -1)
	if len(words) == 0 {
		return false
	}

	// Keep a simple guard against unrelated generic filler.
	genericPhrases := []string{
		"in general", "best practice", "typically", "commonly", "always", "never",
	}
	hasGenericSignal := false
	for _, phrase := range genericPhrases {
		if strings.Contains(text, phrase) {
			hasGenericSignal = true
			break
		}
	}

	matches := 0
	for _, w := range words {
		if _, ok := summaryKeywords[w]; ok {
			matches++
			if matches >= requiredMatches {
				return true
			}
		}
	}
	
	// If has generic signal AND not enough matches → not grounded
	if hasGenericSignal && matches < requiredMatches {
		return false
	}
	
	// Return true if we have enough matches, false otherwise
	return matches >= requiredMatches
}

// cleanGeminiResponse strips markdown fences and extracts the JSON array.
func cleanGeminiResponse(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	first := strings.Index(raw, "[")
	last := strings.LastIndex(raw, "]")
	if first != -1 && last != -1 && last > first {
		return raw[first : last+1]
	}
	return raw
}

// ---------- Batched Generation Wrappers ----------

const batchSize = 5

// GenerateQuestionsBatched splits a large question request into batches of 5.
// It accumulates exclude prompts across batches to avoid duplicates.
// If a batch fails, it retries once. Rate-limit errors stop further batches
// but preserve already-collected questions.
func GenerateQuestionsBatched(topic, difficulty string, count int, excludePrompts []string) ([]GeneratedQuestion, error) {
	if count <= batchSize {
		return GenerateQuestions(topic, difficulty, count, excludePrompts)
	}

	log.Printf("[AI_BATCH] total_requested=%d batch_size=%d", count, batchSize)

	var allQuestions []GeneratedQuestion
	// Accumulate exclude prompts so subsequent batches avoid earlier questions
	excludeAcc := make([]string, len(excludePrompts))
	copy(excludeAcc, excludePrompts)

	remaining := count
	batchNum := 0
	for remaining > 0 {
		batchNum++
		batchCount := batchSize
		if remaining < batchSize {
			batchCount = remaining
		}

		questions, err := GenerateQuestions(topic, difficulty, batchCount, excludeAcc)
		if err != nil {
			log.Printf("[AI_BATCH] batch_%d FAILED: %v", batchNum, err)

			// If rate-limited, stop immediately but keep what we have
			if errors.Is(err, ErrGeminiRateLimit) {
				log.Printf("[AI_BATCH] rate-limited at batch_%d, returning %d questions collected so far", batchNum, len(allQuestions))
				if len(allQuestions) == 0 {
					return nil, ErrGeminiRateLimit
				}
				return allQuestions, nil
			}

			// Retry this batch once
			log.Printf("[AI_BATCH] retrying batch_%d", batchNum)
			questions, err = GenerateQuestions(topic, difficulty, batchCount, excludeAcc)
			if err != nil {
				log.Printf("[AI_BATCH] batch_%d retry FAILED: %v", batchNum, err)
				if errors.Is(err, ErrGeminiRateLimit) && len(allQuestions) > 0 {
					return allQuestions, nil
				}
				// Skip this batch, continue with remaining
				remaining -= batchCount
				continue
			}
		}

		// Add prompts to exclude for next batches
		for _, q := range questions {
			excludeAcc = append(excludeAcc, q.Prompt)
		}

		allQuestions = append(allQuestions, questions...)
		log.Printf("[AI_BATCH] batch_%d_requested=%d returned=%d", batchNum, batchCount, len(questions))
		remaining -= batchCount
	}

	log.Printf("[AI_BATCH] final_valid=%d", len(allQuestions))
	return allQuestions, nil
}

// GenerateQuestionsFromNotesBatched splits notes-based generation into batches of 5.
// Same retry and rate-limit semantics as GenerateQuestionsBatched.
func GenerateQuestionsFromNotesBatched(summary string, count int, difficulty string) ([]GeneratedQuestion, error) {
	if count <= batchSize {
		return GenerateQuestionsFromNotes(summary, count, difficulty)
	}

	log.Printf("[AI_BATCH] notes total_requested=%d batch_size=%d", count, batchSize)

	var allQuestions []GeneratedQuestion
	remaining := count
	batchNum := 0

	for remaining > 0 {
		batchNum++
		batchCount := batchSize
		if remaining < batchSize {
			batchCount = remaining
		}

		questions, err := GenerateQuestionsFromNotes(summary, batchCount, difficulty)
		if err != nil {
			log.Printf("[AI_BATCH] notes batch_%d FAILED: %v", batchNum, err)

			if errors.Is(err, ErrGeminiRateLimit) {
				log.Printf("[AI_BATCH] notes rate-limited at batch_%d, returning %d questions collected so far", batchNum, len(allQuestions))
				if len(allQuestions) == 0 {
					return nil, ErrGeminiRateLimit
				}
				return allQuestions, nil
			}

			// Retry once
			log.Printf("[AI_BATCH] notes retrying batch_%d", batchNum)
			questions, err = GenerateQuestionsFromNotes(summary, batchCount, difficulty)
			if err != nil {
				log.Printf("[AI_BATCH] notes batch_%d retry FAILED: %v", batchNum, err)
				if errors.Is(err, ErrGeminiRateLimit) && len(allQuestions) > 0 {
					return allQuestions, nil
				}
				remaining -= batchCount
				continue
			}
		}

		allQuestions = append(allQuestions, questions...)
		log.Printf("[AI_BATCH] notes batch_%d_requested=%d returned=%d", batchNum, batchCount, len(questions))
		remaining -= batchCount
	}

	log.Printf("[AI_BATCH] notes final_valid=%d", len(allQuestions))
	return allQuestions, nil
}

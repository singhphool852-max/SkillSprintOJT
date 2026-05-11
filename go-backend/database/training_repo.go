package database

import (
	"backend/models"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// CountQuestions returns how many TrainingQuestions exist for a given topic+difficulty.
func CountQuestions(topic, difficulty string) int64 {
	var total int64
	query := DB.Model(&models.TrainingQuestion{}).Where("topic = ?", topic)
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	query.Count(&total)
	return total
}

// GetQuestions fetches random TrainingQuestions from the DB filtered by topic, difficulty, and count.
// If difficulty is empty, all difficulties are returned.
// Uses ORDER BY RANDOM() for SQLite randomisation.
// Deduplicates by question ID to prevent repeats.
func GetQuestions(topic, difficulty string, count int) ([]models.TrainingQuestion, error) {
	total := CountQuestions(topic, difficulty)
	log.Printf("[TrainingRepo] Available questions: topic=%s difficulty=%s total=%d requested=%d", topic, difficulty, total, count)

	if total == 0 {
		return nil, fmt.Errorf("no questions found for topic=%s difficulty=%s", topic, difficulty)
	}

	if int64(count) > total {
		log.Printf("[TrainingRepo] WARNING: Only %d unique questions available, requested %d", total, count)
	}

	var questions []models.TrainingQuestion
	query := DB.Where("topic = ?", topic)
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	query = query.Order("RAND()").Limit(count)

	if err := query.Find(&questions).Error; err != nil {
		log.Printf("[TrainingRepo] ERROR fetching questions: topic=%s difficulty=%s count=%d err=%v", topic, difficulty, count, err)
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}

	// Deduplicate by ID
	initialCount := len(questions)
	questions = dedupeByID(questions)
	idDedupeCount := initialCount - len(questions)

	// Deduplicate by normalized prompt text
	prePromptDedupe := len(questions)
	questions = dedupeByPrompt(questions)
	promptDedupeCount := prePromptDedupe - len(questions)

	// Log selected question IDs and prompt previews
	ids := make([]uint, len(questions))
	for i, q := range questions {
		ids[i] = q.ID
		preview := q.Prompt
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		log.Printf("[TrainingRepo] Selected: id=%d prompt=%q", q.ID, preview)
	}
	log.Printf("[TrainingRepo] Final summary | unique_questions=%d | repeated_ids_skipped=%d | repeated_prompts_skipped=%d | total_available=%d", len(questions), idDedupeCount, promptDedupeCount, total)

	return questions, nil
}

// dedupeByID removes duplicate questions by ID.
func dedupeByID(questions []models.TrainingQuestion) []models.TrainingQuestion {
	seen := make(map[uint]bool)
	result := make([]models.TrainingQuestion, 0, len(questions))
	for _, q := range questions {
		if !seen[q.ID] {
			seen[q.ID] = true
			result = append(result, q)
		}
	}
	return result
}

// NormalizePrompt cleans and standardizes a prompt string for deduplication.
func NormalizePrompt(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Collapse repeated whitespace
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// dedupeByPrompt removes duplicate questions by normalized prompt text.
func dedupeByPrompt(questions []models.TrainingQuestion) []models.TrainingQuestion {
	seen := make(map[string]bool)
	result := make([]models.TrainingQuestion, 0, len(questions))
	for _, q := range questions {
		key := NormalizePrompt(q.Prompt)
		if !seen[key] {
			seen[key] = true
			result = append(result, q)
		}
	}
	return result
}

// CheckPromptExists returns true if a question with the same normalized prompt exists for the topic.
func CheckPromptExists(topic, prompt string) bool {
	var count int64
	// In a real scenario, we'd use a unique hash or normalized field, but for simplicity:
	normalized := NormalizePrompt(prompt)
	
	// We check for fuzzy matches in the DB using LIKE or just exact match for now
	// Since we don't have a normalized_prompt column, we'll do our best with strings.ToLower
	DB.Model(&models.TrainingQuestion{}).
		Where("topic = ? AND LOWER(TRIM(prompt)) = ?", topic, normalized).
		Count(&count)
	
	return count > 0
}

// SaveQuestions inserts AI-generated TrainingQuestions into the DB.
// Deduplication: skips any question whose prompt already exists for the same topic.
func SaveQuestions(questions []models.TrainingQuestion) (int, error) {
	inserted := 0

	for i := range questions {
		// Check for duplicate by (topic, prompt)
		var existing int64
		DB.Model(&models.TrainingQuestion{}).
			Where("topic = ? AND prompt = ?", questions[i].Topic, questions[i].Prompt).
			Count(&existing)

		if existing > 0 {
			log.Printf("[TrainingRepo] Skipped duplicate question: topic=%s prompt=%.60s...", questions[i].Topic, questions[i].Prompt)
			continue
		}

		if err := DB.Create(&questions[i]).Error; err != nil {
			log.Printf("[TrainingRepo] ERROR inserting question: %v", err)
			continue // Defensive: skip malformed rows
		}
		inserted++
	}

	log.Printf("[TrainingRepo] Saved %d/%d questions (duplicates skipped: %d)", inserted, len(questions), len(questions)-inserted)
	return inserted, nil
}

// CreateSession stores a new TrainingSession in the DB.
// questionIDs is a slice of TrainingQuestion IDs that gets serialised to JSON.
func CreateSession(session models.TrainingSession) error {
	if err := DB.Create(&session).Error; err != nil {
		log.Printf("[TrainingRepo] ERROR creating session: session_id=%s err=%v", session.SessionID, err)
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("[TrainingRepo] Session created | session_id=%s topic=%s status=%s", session.SessionID, session.Topic, session.Status)
	return nil
}

// GetSession retrieves a TrainingSession by session_id and hydrates the associated questions.
// Returns the session and the full question objects referenced in question_ids.
func GetSession(sessionID string) (*models.TrainingSession, []models.TrainingQuestion, error) {
	var session models.TrainingSession
	if err := DB.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		log.Printf("[TrainingRepo] ERROR fetching session: session_id=%s err=%v", sessionID, err)
		return nil, nil, fmt.Errorf("session not found: %w", err)
	}

	// Parse question_ids JSON array
	var questionIDs []uint
	if err := json.Unmarshal([]byte(session.QuestionIDs), &questionIDs); err != nil {
		log.Printf("[TrainingRepo] ERROR parsing question_ids for session %s: %v", sessionID, err)
		return &session, nil, fmt.Errorf("failed to parse question IDs: %w", err)
	}

	var questions []models.TrainingQuestion
	if len(questionIDs) > 0 {
		if err := DB.Where("id IN ?", questionIDs).Find(&questions).Error; err != nil {
			log.Printf("[TrainingRepo] ERROR fetching questions for session %s: %v", sessionID, err)
			return &session, nil, fmt.Errorf("failed to fetch session questions: %w", err)
		}
	}

	log.Printf("[TrainingRepo] Session retrieved | session_id=%s topic=%s questions=%d", sessionID, session.Topic, len(questions))
	return &session, questions, nil
}

package handlers

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"log"
	"time"
)

// StartAutoSubmitWatcher runs a background goroutine that periodically
// checks for expired test attempts and auto-submits them.
// This protects against users who close the browser without submitting.
func StartAutoSubmitWatcher() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			autoSubmitExpiredAttempts()
		}
	}()
	log.Println("Auto-submit watcher started (checks every 30s)")
}

// autoSubmitExpiredAttempts finds all un-submitted attempts whose test
// window has closed, grades them, and marks them as auto-submitted.
//
// CRITICAL: This runs in a background goroutine. Every DB write must use
// a transaction with a re-check to prevent racing with SubmitTestAttempt.
func autoSubmitExpiredAttempts() {
	// Find all active tests that have ended
	var tests []models.Test
	database.DB.Where("isPublished = ?", true).Find(&tests)

	now := time.Now()
	for _, test := range tests {
		if test.StartTime == nil {
			continue
		}
		elapsed := now.Sub(*test.StartTime)
		if int(elapsed.Seconds()) < test.DurationSeconds {
			continue // test still running
		}

		// Find un-submitted attempts for this ended test
		// FIX: Exclude Go zero-time values that SQLite stores as real timestamps
		var attempts []models.TestAttempt
		database.DB.Where(
			"testId = ? AND (submittedAt IS NULL OR submittedAt = '' OR submittedAt < '0001-01-02')",
			test.ID,
		).Find(&attempts)

		for _, attempt := range attempts {
			// Double-check in Go (belt + suspenders)
			if !attempt.SubmittedAt.IsZero() {
				continue
			}

			autoSubmitSingleAttempt(attempt, test)
		}
	}
}

// autoSubmitSingleAttempt grades and submits one attempt inside a transaction.
// If the attempt was already submitted (by the user or a previous watcher tick),
// the transaction rolls back harmlessly.
func autoSubmitSingleAttempt(attempt models.TestAttempt, test models.Test) {
	// Use a transaction to prevent racing with SubmitTestAttempt
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("[AUTO-SUBMIT] failed to begin tx for attempt %s: %v", attempt.ID, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("[AUTO-SUBMIT] panic recovered for attempt %s: %v", attempt.ID, r)
		}
	}()

	// Re-read the attempt inside the transaction to check if it was submitted
	// between our initial query and now (race with SubmitTestAttempt)
	var freshAttempt models.TestAttempt
	if err := tx.Where("id = ?", attempt.ID).First(&freshAttempt).Error; err != nil {
		tx.Rollback()
		log.Printf("[AUTO-SUBMIT] attempt %s not found in tx: %v", attempt.ID, err)
		return
	}

	// If already submitted, skip — someone else got there first
	if !freshAttempt.SubmittedAt.IsZero() {
		tx.Rollback()
		log.Printf("[AUTO-SUBMIT] attempt %s already submitted, skipping", attempt.ID)
		return
	}

	// Grade MCQs
	var questions []models.TestQuestion
	tx.Preload("MCQOptions").Where("testId = ?", freshAttempt.TestID).Find(&questions)

	var submissions []models.TestSubmission
	tx.Where("attemptId = ?", freshAttempt.ID).Find(&submissions)

	subMap := make(map[string]*models.TestSubmission)
	for i := range submissions {
		subMap[submissions[i].QuestionID] = &submissions[i]
	}

	totalScore := 0
	for _, q := range questions {
		sub, exists := subMap[q.ID]
		if !exists {
			continue
		}

		if q.Type == "mcq" {
			for _, opt := range q.MCQOptions {
				if opt.ID == sub.SelectedOptionID && opt.IsCorrect {
					sub.Score = q.Points
					sub.Verdict = "accepted"
					break
				}
			}
			if sub.Verdict != "accepted" {
				sub.Score = 0
				sub.Verdict = "wrong_answer"
			}
			// FIX: Use targeted Updates() instead of Save() to avoid overwriting unrelated columns
			tx.Model(sub).Updates(map[string]interface{}{
				"score":   sub.Score,
				"verdict": sub.Verdict,
			})
		}
		totalScore += sub.Score
	}

	// FIX: Use targeted Updates() instead of Save() to avoid overwriting columns with zero values
	submittedAt := time.Now()
	result := tx.Model(&models.TestAttempt{}).Where("id = ?", freshAttempt.ID).Updates(map[string]interface{}{
		"score":           totalScore,
		"totalQuestions":  len(questions),
		"timeTaken":       int(time.Since(freshAttempt.StartedAt).Seconds()),
		"submittedAt":     submittedAt,
		"isAutoSubmitted": true,
	})
	if result.Error != nil {
		tx.Rollback()
		log.Printf("[AUTO-SUBMIT] failed to update attempt %s: %v", freshAttempt.ID, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("[AUTO-SUBMIT] attempt %s: 0 rows affected, possible race", freshAttempt.ID)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[AUTO-SUBMIT] commit failed for attempt %s: %v", freshAttempt.ID, err)
		return
	}

	log.Printf("[AUTO-SUBMIT] SUCCESS: attempt=%s test=%s score=%d", freshAttempt.ID, test.ID, totalScore)

	// Refresh attempt with committed values for post-commit side effects
	freshAttempt.Score = totalScore
	freshAttempt.TotalQuestions = len(questions)
	freshAttempt.TimeTaken = int(time.Since(freshAttempt.StartedAt).Seconds())
	freshAttempt.SubmittedAt = &submittedAt
	freshAttempt.IsAutoSubmitted = true

	// Post-commit side effects: persist result + track wrong answers
	computeAndSaveResult(freshAttempt)
	extractWrongQuestions(freshAttempt)

	// Notify via WebSocket (after commit, non-blocking)
	if ArenaSessionHub != nil {
		ArenaSessionHub.BroadcastAutoSubmit(freshAttempt.ID, totalScore)
	}

	// Broadcast leaderboard update (after commit)
	broadcastLeaderboard(freshAttempt.TestID)
}

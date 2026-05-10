package handlers

import (
	"backend/database"
	"backend/models"
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
func autoSubmitExpiredAttempts() {
	// Find all active tests that have ended
	var tests []models.Test
	database.DB.Where("isPublished = ?", true).Find(&tests)

	now := time.Now()
	for _, test := range tests {
		elapsed := now.Sub(test.StartTime)
		if int(elapsed.Seconds()) < test.DurationSeconds {
			continue // test still running
		}

		// Find un-submitted attempts for this ended test
		var attempts []models.TestAttempt
		database.DB.Where("testId = ? AND (submittedAt IS NULL OR submittedAt = '')", test.ID).Find(&attempts)

		for _, attempt := range attempts {
			if !attempt.SubmittedAt.IsZero() {
				continue
			}

			// Grade MCQs
			var questions []models.TestQuestion
			database.DB.Preload("MCQOptions").Where("testId = ?", attempt.TestID).Find(&questions)

			var submissions []models.TestSubmission
			database.DB.Where("attemptId = ?", attempt.ID).Find(&submissions)

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
					database.DB.Save(sub)
				}
				totalScore += sub.Score
			}

			attempt.Score = totalScore
			attempt.TotalQuestions = len(questions)
			attempt.TimeTaken = int(time.Since(attempt.StartedAt).Seconds())
			attempt.SubmittedAt = time.Now()
			attempt.IsAutoSubmitted = true
			database.DB.Save(&attempt)

			log.Printf("Auto-submitted attempt %s for test %s (score: %d)", attempt.ID, test.ID, totalScore)

			// Notify via WebSocket
			if ArenaSessionHub != nil {
				ArenaSessionHub.BroadcastAutoSubmit(attempt.ID, totalScore)
			}

			// Broadcast leaderboard update
			broadcastLeaderboard(attempt.TestID)
		}
	}
}

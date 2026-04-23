package database

import (
	"backend/models"
	"log"
	"time"

	"github.com/google/uuid"
)

func SeedDB() {
	var count int64
	DB.Model(&models.Arena{}).Count(&count)
	if count > 4 {
		return // Data already seeded
	}

	log.Println("Seeding training modules data...")

	// Categories
	catSoftware := models.QuizCategory{ID: uuid.New().String(), Name: "Software Engineering", Slug: "software-engineering"}
	catAptitude := models.QuizCategory{ID: uuid.New().String(), Name: "Aptitude", Slug: "aptitude"}
	DB.Create(&catSoftware)
	DB.Create(&catAptitude)

	arenas := []models.Arena{
		{ID: "dsa_arena", Title: "DSA Speed Relay", Slug: "dsa-arena", CategoryID: catSoftware.ID, Difficulty: "Medium", Status: "live"},
		{ID: "dbms_arena", Title: "DBMS Optimizer", Slug: "dbms-arena", CategoryID: catSoftware.ID, Difficulty: "Medium", Status: "live"},
		{ID: "os_arena", Title: "OS Deadlock Escape", Slug: "os-arena", CategoryID: catSoftware.ID, Difficulty: "Hard", Status: "live"},
		{ID: "js_arena", Title: "JavaScript Syntax", Slug: "js-arena", CategoryID: catSoftware.ID, Difficulty: "Medium", Status: "live"},
		{ID: "aptitude_arena", Title: "Cognitive Relay", Slug: "aptitude-arena", CategoryID: catAptitude.ID, Difficulty: "Intermediate", Status: "live"},
	}

	for _, a := range arenas {
		DB.Create(&a)
		// Give each arena an active quiz
		q := models.Quiz{
			ID:         uuid.New().String(),
			Title:      a.Title + " Module",
			ArenaID:    a.ID,
			CategoryID: a.CategoryID,
			Difficulty: a.Difficulty,
			IsActive:   true,
		}
		DB.Create(&q)

		// Seed a quick dummy question so backend doesn't return empty array if queried directly
		qn := models.Question{
			ID:            uuid.New().String(),
			QuizID:        q.ID,
			Prompt:        "Neural diagnostic complete. Start primary sync?",
			Type:          "mcq",
			CorrectAnswer: "",
			Explanation:   "System primed.",
			MaxScore:      10,
		}
		DB.Create(&qn)
		DB.Create(&models.Option{ID: uuid.New().String(), QuestionID: qn.ID, Text: "Yes", IsCorrect: true})
		DB.Create(&models.Option{ID: uuid.New().String(), QuestionID: qn.ID, Text: "No", IsCorrect: false})
	}

	// Make sure a dummy user exists
	var userCount int64
	DB.Model(&models.User{}).Where("username = ?", "RIVAL_X").Count(&userCount)
	if userCount == 0 {
		dummyUser := models.User{
			ID:        uuid.New().String(),
			Email:     "dummy@skillsprint.io",
			Password:  "password",
			Username:  "RIVAL_X",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		DB.Create(&dummyUser)
	}

	log.Println("Seed complete.")
}

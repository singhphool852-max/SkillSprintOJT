package database

import (
	"log"

	"backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Pointing to the Next.js SQLite database
	dsn := "../dev.db"
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!", err)
	}

	err = database.AutoMigrate(
		// Existing models
		&models.User{},
		&models.QuizCategory{},
		&models.Arena{},
		&models.Quiz{},
		&models.Question{},
		&models.Option{},
		&models.Attempt{},
		&models.AttemptAnswer{},
		// Test module models
		&models.Topic{},
		&models.Test{},
		&models.TestQuestion{},
		&models.TestMCQOption{},
		&models.TestCodingDetail{},
		&models.TestCase{},
		&models.TestAttempt{},
		&models.TestSubmission{},
		&models.TestResult{},
		&models.TestViolation{},

		// Training module models
		&models.TrainingQuestion{},
		&models.TrainingSession{},
		&models.Upload{},
	)
	if err != nil {
		log.Println("Database migration error (ignoring if table already populated):", err)
	}

	sqlDB, err := database.DB()
	if err == nil {
		// SQLite standard for avoiding locks in concurrent access
		sqlDB.SetMaxOpenConns(1)
		database.Exec("PRAGMA journal_mode=WAL;")
		database.Exec("PRAGMA synchronous=NORMAL;")
	}

	DB = database
	log.Println("Database connection established (WAL mode enabled)")

	// Basic check to seed data if empty
	SeedDB()
	// Seed training questions (runs only once)
	SeedTrainingQuestions()
}

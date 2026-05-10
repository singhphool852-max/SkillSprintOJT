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

	// ── Explicit schema fixes ──
	// SQLite AutoMigrate cannot add columns/indexes to existing tables.
	// These are idempotent: they silently fail if the column/index already exists.
	migrations := []string{
		// Missing columns on test_attempts
		"ALTER TABLE test_attempts ADD COLUMN totalQuestions integer DEFAULT 0",
		"ALTER TABLE test_attempts ADD COLUMN timeTaken integer DEFAULT 0",
		"ALTER TABLE test_attempts ADD COLUMN violationCount integer DEFAULT 0",
		// Composite unique index — prevents duplicate (userId, testId) pairs
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_test ON test_attempts(userId, testId)",
	}
	for _, m := range migrations {
		if err := database.Exec(m).Error; err != nil {
			// Expected to fail if column/index already exists — not an error
			log.Printf("[MIGRATE] Skipped (already exists): %s", m)
		} else {
			log.Printf("[MIGRATE] Applied: %s", m)
		}
	}
	log.Println("[MIGRATE] Schema sync complete")

	// Basic check to seed data if empty
	SeedDB()
	// Seed training questions (runs only once)
	SeedTrainingQuestions()
}

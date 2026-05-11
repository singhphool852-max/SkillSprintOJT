package database

import (
	"log"
	"os"

	"backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var database *gorm.DB
	var err error

	// Check for PostgreSQL connection string (Standard for Render/Heroku)
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL != "" {
		log.Println("[DB] DATABASE_URL found. Connecting to PostgreSQL...")
		database, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	} else {
		// Fallback to SQLite for local development
		log.Println("[DB] No DATABASE_URL found. Falling back to local SQLite (dev.db)...")
		dsn := "../dev.db"
		if _, err := os.Stat("dev.db"); err == nil {
			dsn = "dev.db"
		}
		database, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}

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

		// Wrong question tracking & analytics
		&models.UserWrongQuestion{},
		&models.UserTopicStats{},
	)
	if err != nil {
		log.Println("Database migration error (ignoring if table already populated):", err)
	}

	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
		database.Exec("PRAGMA journal_mode=WAL;")
		database.Exec("PRAGMA synchronous=NORMAL;")
	}

	DB = database
	log.Println("Database connection established (WAL mode enabled)")

	// ── Explicit schema fixes ──
	migrations := []string{
		"ALTER TABLE test_attempts ADD COLUMN totalQuestions integer DEFAULT 0",
		"ALTER TABLE test_attempts ADD COLUMN timeTaken integer DEFAULT 0",
		"ALTER TABLE test_attempts ADD COLUMN violationCount integer DEFAULT 0",
		"ALTER TABLE test_attempts ADD COLUMN mode text DEFAULT 'arena'",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user_test ON test_attempts(userId, testId)",
		"ALTER TABLE tests ADD COLUMN deletedAt datetime",
		"ALTER TABLE tests ADD COLUMN deletedBy text DEFAULT ''",
		"ALTER TABLE test_results ADD COLUMN topicId text DEFAULT ''",
		"ALTER TABLE test_results ADD COLUMN topicName text DEFAULT ''",
		"ALTER TABLE test_results ADD COLUMN timeTaken integer DEFAULT 0",
		"CREATE INDEX IF NOT EXISTS idx_attempt_test_score ON test_attempts(testId, score DESC)",
		"CREATE INDEX IF NOT EXISTS idx_attempt_user ON test_attempts(userId)",
		"CREATE INDEX IF NOT EXISTS idx_submissions_attempt ON test_submissions(attemptId)",
		"CREATE INDEX IF NOT EXISTS idx_wrong_questions_user ON user_wrong_questions(userId)",
		"CREATE INDEX IF NOT EXISTS idx_wrong_questions_topic ON user_wrong_questions(userId, topicId)",
		"CREATE INDEX IF NOT EXISTS idx_test_results_user ON test_results(userId)",
		"CREATE INDEX IF NOT EXISTS idx_tests_deleted ON tests(deletedAt)",
	}
	for _, m := range migrations {
		database.Exec(m)
	}

	SeedDB()
	SeedTrainingQuestions()
	
	// NEW: Sync old categories to new topics so Admin Panel has data immediately
	SyncCategoriesToTopics()
}

// SyncCategoriesToTopics ensures that any old QuizCategory is available as a Topic
func SyncCategoriesToTopics() {
	var categories []models.QuizCategory
	DB.Find(&categories)

	for _, cat := range categories {
		var exists int64
		DB.Model(&models.Topic{}).Where("id = ? OR slug = ?", cat.ID, cat.Slug).Count(&exists)
		if exists == 0 {
			log.Printf("[SYNC] Migrating category %s to topics table", cat.Name)
			DB.Create(&models.Topic{
				ID:   cat.ID,
				Name: cat.Name,
				Slug: cat.Slug,
			})
		}
	}
}

package database

import (
	"log"
	"os"
	"time"

	"backend/models"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Get MySQL DSN from environment
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("[DB] MYSQL_DSN environment variable is not set")
	}

	log.Println("[DB] Connecting to MySQL...")

	// Open MySQL connection with GORM
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("[DB] Failed to connect to MySQL:", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("[DB] Failed to get database instance:", err)
	}

	// Configure connection pool for high concurrency
	sqlDB.SetMaxOpenConns(100)                  // Maximum 100 concurrent connections
	sqlDB.SetMaxIdleConns(20)                   // Keep 20 idle connections ready
	sqlDB.SetConnMaxLifetime(5 * time.Minute)   // Recycle connections every 5 minutes
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)   // Close idle connections after 2 minutes

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("[DB] MySQL ping failed:", err)
	}

	log.Println("[DB] MySQL connected successfully")

	// Auto-migrate all models
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
		log.Println("[DB] Database migration error (ignoring if table already populated):", err)
	}

	DB = database
	log.Println("[DB] Database migration completed")

	// Run seed data
	SeedDB()
	SeedTrainingQuestions()
	
	// Sync old categories to new topics
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

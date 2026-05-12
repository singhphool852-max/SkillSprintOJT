package database

import (
	"log"
	"os"
	"time"

	"backend/models"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB initializes the MySQL connection using MYSQL_DSN environment variable.
func ConnectDB() {
	// Get MySQL DSN from environment
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = os.Getenv("MYSQL_URL")
	}
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	if dsn == "" {
		log.Fatal("[DB] FATAL: MYSQL_DSN environment variable is not set. Please set MYSQL_DSN in your environment variables.")
	}

	log.Println("[DB] MYSQL_DSN found, connecting to MySQL...")
	log.Printf("[DB] DSN format check: %d characters", len(dsn))

	// Open connection with detailed logging in development
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("[DB] Failed to connect to MySQL database:", err)
	}

	// Configure Connection Pool
	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	DB = database
	log.Println("[DB] MySQL connection established successfully")

	// Run Auto-Migrations for all models to ensure schema parity
	MigrateModels()

	// Bootstrap essential data
	SeedTrainingQuestions()
	SyncCategoriesToTopics()
}

// MigrateModels ensures all tables exist with correct relationships and indexes
func MigrateModels() {
	log.Println("[DB] Running auto-migrations...")
	err := DB.AutoMigrate(
		&models.User{},
		&models.QuizCategory{},
		&models.Arena{},
		&models.Quiz{},
		&models.Question{},
		&models.Option{},
		&models.Attempt{},
		&models.Test{},
		&models.TestQuestion{},
		&models.TestMCQOption{},
		&models.TestCodingDetail{},
		&models.TestCase{},
		&models.TestAttempt{},
		&models.TestSubmission{},
		&models.TestResult{},
		&models.Topic{},
		&models.TestViolation{},
		&models.TrainingQuestion{},
		&models.TrainingSession{},
		&models.Upload{},
		&models.UserWrongQuestion{},
		&models.UserTopicStats{},
	)

	if err != nil {
		log.Println("[DB] Migration warning:", err)
	}
}

func SyncCategoriesToTopics() {
	var categories []models.QuizCategory
	DB.Find(&categories)

	for _, cat := range categories {
		var exists int64
		DB.Model(&models.Topic{}).Where("id = ? OR slug = ?", cat.ID, cat.Slug).Count(&exists)
		if exists == 0 {
			log.Printf("[SYNC] Mapping category %s to topics table", cat.Name)
			DB.Create(&models.Topic{
				ID:   cat.ID,
				Name: cat.Name,
				Slug: cat.Slug,
			})
		}
	}
}

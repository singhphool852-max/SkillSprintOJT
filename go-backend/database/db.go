package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB initializes the MySQL connection using environment variables.
func ConnectDB() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	// Fallback to a single DSN string if provided (convenient for some cloud providers)
	dsn := os.Getenv("MYSQL_DSN")

	if dsn == "" {
		// Build DSN from individual components
		// Example: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		if user == "" || host == "" || name == "" {
			log.Fatal("[DB] Critical environment variables missing (DB_USER, DB_HOST, DB_NAME)")
		}
		if port == "" {
			port = "3306"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
			user, pass, host, port, name)
	}

	log.Printf("[DB] Connecting to MySQL at %s:%s/%s...", host, port, name)

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
		sqlDB.SetMaxIdleConns(25)
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
		&models.Result{},
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

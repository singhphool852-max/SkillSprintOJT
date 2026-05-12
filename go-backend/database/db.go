package database

import (
	"log"
	"os"
	"time"

	"backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var database *gorm.DB
	var err error

	// 1. Detect Environment & Connection String
	postgresDSN := os.Getenv("DATABASE_URL")
	mysqlDSN := os.Getenv("MYSQL_DSN")

	if postgresDSN != "" {
		log.Println("[DB] Connecting to PostgreSQL (Production)...")
		database, err = gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	} else if mysqlDSN != "" {
		log.Println("[DB] Connecting to MySQL...")
		database, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	} else {
		log.Println("[DB] No cloud DB found. Falling back to local SQLite (dev.db)...")
		database, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})
	}

	if err != nil {
		log.Fatal("[DB] Failed to connect to database:", err)
	}

	// 2. Configure Connection Pool
	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		
		// SQLite specific performance tweaks
		if postgresDSN == "" && mysqlDSN == "" {
			database.Exec("PRAGMA journal_mode=WAL;")
			database.Exec("PRAGMA synchronous=NORMAL;")
		}
	}

	DB = database
	log.Println("[DB] Database connection established")

	// 3. Auto-Migrate Models
	err = database.AutoMigrate(
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
		log.Println("[DB] Migration error (non-fatal):", err)
	}

	// Seed required data
	SeedTrainingQuestions()
	SyncCategoriesToTopics()
}

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

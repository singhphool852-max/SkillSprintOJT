package database

import (
	"log"
	"os"
	"time"

	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"

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
		log.Fatal("[DB] FATAL: No database connection string found (MYSQL_DSN, MYSQL_URL, or DATABASE_URL).")
	}

	log.Println("[DB] Connecting to MySQL...")

	// Open connection with detailed logging in development
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("[DB] FATAL: Failed to connect to MySQL: %v", err)
	}

	// Configure Connection Pool
	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	DB = database
	log.Println("[DB] ✅ MySQL connection established successfully")

	// 1. Run Auto-Migrations (Strict Order)
	MigrateModels()

	// 2. Sync Categories to Topics (Ensure parent topics exist)
	SyncCategoriesToTopics()

	// 3. Bootstrap essential data (Seeding)
	SeedTrainingQuestions()
}

// MigrateModels ensures all tables exist with correct relationships and indexes.
// Strategic Order: Parent tables must be migrated before child tables for FK integrity.
func MigrateModels() {
	log.Println("[DB] Starting schema auto-migrations...")

	// Phase 1: Core/Parent Tables
	log.Println("[DB] Migrating Core tables (User, Category, Topic)...")
	err := DB.AutoMigrate(
		&models.User{},
		&models.QuizCategory{},
		&models.Topic{},
	)
	if err != nil {
		log.Fatalf("[DB] FATAL: Core migration failed: %v", err)
	}

	// Phase 2: Secondary Parent Tables (Arenas, Tests, Training)
	log.Println("[DB] Migrating Secondary tables (Arena, Quiz, Test, Training)...")
	err = DB.AutoMigrate(
		&models.Arena{},
		&models.Quiz{},
		&models.Test{},
		&models.TrainingQuestion{},
		&models.TrainingSession{},
		&models.Upload{},
	)
	if err != nil {
		log.Fatalf("[DB] FATAL: Secondary migration failed: %v", err)
	}

	// Phase 3: Content/Child Tables (Questions, Options, TestCases)
	log.Println("[DB] Migrating Content tables (Questions, MCQOptions, TestCases)...")
	err = DB.AutoMigrate(
		&models.Question{},
		&models.Option{},
		&models.TestQuestion{},
		&models.TestMCQOption{},
		&models.TestCodingDetail{},
		&models.TestCase{},
	)
	if err != nil {
		log.Fatalf("[DB] FATAL: Content migration failed: %v", err)
	}

	// Phase 4: Activity/Result Tables (Attempts, Submissions, Analytics)
	log.Println("[DB] Migrating Activity tables (Attempts, Submissions, Analytics)...")
	err = DB.AutoMigrate(
		&models.Attempt{},
		&models.TestAttempt{},
		&models.TestSubmission{},
		&models.TestResult{},
		&models.TestViolation{},
		&models.UserWrongQuestion{},
		&models.UserTopicStats{},
		&models.ChatMessage{},
	)
	if err != nil {
		log.Fatalf("[DB] FATAL: Activity migration failed: %v", err)
	}

	log.Println("[DB] ✅ Schema auto-migrations completed successfully.")
}

func SyncCategoriesToTopics() {
	log.Println("[DB] Syncing categories to topics table...")
	var categories []models.QuizCategory
	if err := DB.Find(&categories).Error; err != nil {
		log.Printf("[DB] Sync warning: failed to fetch categories: %v", err)
		return
	}

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

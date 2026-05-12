package main

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 1. Load env
	godotenv.Load()
	dsn := os.Getenv("MYSQL_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}

	if dsn == "" {
		log.Fatal("ERROR: No database connection string found in environment (MYSQL_URL, DATABASE_URL, or MYSQL_DSN)")
	}

	log.Println("Connecting to Railway MySQL...")

	// 2. Connect
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	database.DB = db

	log.Println("Connection successful! Running migrations...")

	// 3. Migrate all tables
	err = db.AutoMigrate(
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
		log.Fatal("Migration failed:", err)
	}
	log.Println("✅ All tables created/updated successfully.")

	// 4. Seed Training Questions (Vault)
	log.Println("Seeding training questions (Vault)...")
	database.SeedTrainingQuestions()

	// 5. Sync Categories to Topics
	log.Println("Syncing categories to topics...")
	database.SyncCategoriesToTopics()

	log.Println("\n✨ DATABASE INITIALIZATION COMPLETE!")
	log.Println("You can now login and start using SkillSprint with the new Railway DB.")
}

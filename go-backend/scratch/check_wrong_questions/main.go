package main

import (
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	godotenv.Load("../../.env")
	
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = os.Getenv("MYSQL_URL")
	}
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	database.ConnectDB()

	var count int64
	database.DB.Model(&models.UserWrongQuestion{}).Count(&count)
	log.Printf("[REPORT] Total UserWrongQuestions in DB: %d", count)

	var topics []models.UserTopicStats
	database.DB.Find(&topics)
	for _, t := range topics {
		log.Printf("[REPORT] User %s Topic %s: Accuracy %.2f%%, Wrong %d", t.UserID, t.TopicName, t.AccuracyPercent, t.TotalWrong)
	}
}

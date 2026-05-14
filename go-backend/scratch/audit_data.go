package main

import (
	"fmt"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	database.ConnectDB()

	fmt.Println("--- MISTAKE AUDIT ---")
	var allMistakes []models.UserWrongQuestion
	database.DB.Find(&allMistakes)
	fmt.Printf("Total Mistakes in DB: %d\n", len(allMistakes))
	for _, m := range allMistakes {
		fmt.Printf(" - User: %s, Topic: %s, Question: %s\n", m.UserID, m.TopicID, m.QuestionTitle)
	}

	fmt.Println("\n--- TOPIC STATS AUDIT ---")
	var allStats []models.UserTopicStats
	database.DB.Find(&allStats)
	fmt.Printf("Total Topic Stats in DB: %d\n", len(allStats))
	for _, s := range allStats {
		fmt.Printf(" - User: %s, Topic: %s, Accuracy: %.2f%%\n", s.UserID, s.TopicID, s.AccuracyPercent)
	}

	fmt.Println("\n--- VAULT AUDIT ---")
	var count int64
	database.DB.Model(&models.TrainingQuestion{}).Count(&count)
	fmt.Printf("Total questions in Training Vault: %d\n", count)
	
	var topics []string
	database.DB.Model(&models.TrainingQuestion{}).Distinct("topic").Pluck("topic", &topics)
	fmt.Printf("Topics available in vault: %v\n", topics)
}

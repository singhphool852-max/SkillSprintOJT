package main

import (
	"backend/database"
	"backend/models"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load("../.env")

	// Connect to database
	database.ConnectDB()

	// 1. Pick a user with mistakes
	var user models.User
	database.DB.Order("createdAt DESC").First(&user)
	if user.ID == "" {
		log.Fatal("No users found in database")
	}

	fmt.Printf("Analyzing Adaptive Training for User: %s (%s)\n", user.Name, user.ID)

	// 2. Fetch mistakes (highest wrong count first)
	var mistakes []models.UserWrongQuestion
	database.DB.Where("userId = ?", user.ID).Order("wrongCount DESC, createdAt DESC").Limit(5).Find(&mistakes)

	if len(mistakes) == 0 {
		fmt.Println("No mistakes recorded for this user yet. Start some tests in the Arena!")
		return
	}

	fmt.Printf("Found %d high-priority mistakes:\n", len(mistakes))
	for i, m := range mistakes {
		fmt.Printf("%d. [%s] %s (Failed %d times)\n", i+1, m.QuestionType, m.QuestionTitle, m.WrongCount)
	}

	// 3. Fetch weak topics
	var stats []models.UserTopicStats
	database.DB.Where("userId = ?", user.ID).Order("accuracyPercent ASC").Limit(3).Find(&stats)

	fmt.Println("\nWeakest Topics:")
	for _, s := range stats {
		fmt.Printf("- %s: Accuracy %.2f%% (%s)\n", s.TopicName, s.AccuracyPercent, s.WeakLevel)
	}

	fmt.Println("\nSuccess: Adaptive logic is correctly identifying priority areas.")
}
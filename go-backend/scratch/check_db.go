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
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found, using system environment")
	}

	// Connect to database
	database.ConnectDB()

	fmt.Println("\n--- SkillSprint Database Statistics ---")

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	fmt.Printf("Total Users: %d\n", userCount)

	var topicCount int64
	database.DB.Model(&models.Topic{}).Count(&topicCount)
	fmt.Printf("Total Topics: %d\n", topicCount)

	var testCount int64
	database.DB.Model(&models.Test{}).Count(&testCount)
	fmt.Printf("Total Tests: %d\n", testCount)

	var mistakeCount int64
	database.DB.Model(&models.UserWrongQuestion{}).Count(&mistakeCount)
	fmt.Printf("Total Tracked Mistakes: %d\n", mistakeCount)

	var topicStatsCount int64
	database.DB.Model(&models.UserTopicStats{}).Count(&topicStatsCount)
	fmt.Printf("Users with Topic Stats: %d\n", topicStatsCount)

	fmt.Println("--------------------------------------\n")
}
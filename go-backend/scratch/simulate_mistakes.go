package main

import (
	"backend/database"
	"backend/models"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load("../.env")

	// Connect to database
	database.ConnectDB()

	// 1. Get or Create a test user
	var user models.User
	result := database.DB.Where("email = ?", "test_student@skillsprint.com").First(&user)
	if result.Error != nil {
		user = models.User{
			ID:    uuid.New().String(),
			Name:  "Test Student",
			Email: "test_student@skillsprint.com",
			Role:  "student",
		}
		database.DB.Create(&user)
		fmt.Printf("Created test user: %s\n", user.Email)
	}

	// 2. Get a Topic to fail in
	var topic models.Topic
	database.DB.First(&topic)
	if topic.ID == "" {
		topic = models.Topic{ID: uuid.New().String(), Name: "System Design", Slug: "system-design"}
		database.DB.Create(&topic)
	}

	// 3. Create a Dummy Test and Question
	testID := uuid.New().String()
	questionID := uuid.New().String()

	fmt.Printf("Simulating mistake for User: %s in Topic: %s\n", user.Name, topic.Name)

	// 4. Record a Wrong Question (Upsert Logic)
	mistake := models.UserWrongQuestion{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		QuestionID:    questionID,
		TestID:        testID,
		TopicID:       topic.ID,
		QuestionType:  "mcq",
		QuestionTitle: "What is horizontal scaling?",
		Verdict:       "wrong_answer",
		WrongCount:    1,
		CreatedAt:     time.Now(),
	}

	// Check if this user+question combo exists
	var existing models.UserWrongQuestion
	database.DB.Where("userId = ? AND questionId = ?", user.ID, questionID).First(&existing)

	if existing.ID != "" {
		database.DB.Model(&existing).Update("wrongCount", existing.WrongCount+1)
		fmt.Printf("Increased mistake count for existing question. New count: %d\n", existing.WrongCount+1)
	} else {
		database.DB.Create(&mistake)
		fmt.Println("Recorded new unique mistake in database.")
	}

	// 5. Update Topic Stats
	var stats models.UserTopicStats
	database.DB.Where("userId = ? AND topicId = ?", user.ID, topic.ID).First(&stats)

	if stats.ID == "" {
		stats = models.UserTopicStats{
			ID:             uuid.New().String(),
			UserID:         user.ID,
			TopicID:        topic.ID,
			TopicName:      topic.Name,
			TotalAttempted: 1,
			TotalWrong:     1,
			AccuracyPercent: 0,
			WeakLevel:      "critical",
		}
		database.DB.Create(&stats)
		fmt.Println("Initialized topic stats for user.")
	} else {
		newAttempted := stats.TotalAttempted + 1
		newWrong := stats.TotalWrong + 1
		newAccuracy := (float64(stats.TotalCorrect) / float64(newAttempted)) * 100
		
		database.DB.Model(&stats).Updates(map[string]interface{}{
			"totalAttempted":  newAttempted,
			"totalWrong":      newWrong,
			"accuracyPercent": newAccuracy,
			"weakLevel":       "critical",
		})
		fmt.Printf("Updated topic stats. Accuracy now: %.2f%%\n", newAccuracy)
	}

	fmt.Println("Simulation Complete.")
}
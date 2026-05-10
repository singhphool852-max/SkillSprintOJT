//go:build ignore

package main

import (
	"fmt"
	"log"

	"backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var questions []models.TestQuestion
	if err := db.Find(&questions).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total questions: %d\n", len(questions))
	for _, q := range questions {
		fmt.Printf("Q: %s, testId: %s\n", q.ID, q.TestID)
	}
}

package models

import (
	"time"
)

type QuizCategory struct {
	ID   string `gorm:"primaryKey;column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
	Slug string `gorm:"unique;column:slug" json:"slug"`
}

type Arena struct {
	ID              string    `gorm:"primaryKey;column:id" json:"id"`
	Title           string    `gorm:"column:title" json:"title"`
	Slug            string    `gorm:"unique;column:slug" json:"slug"`
	CategoryID      string    `gorm:"column:categoryId" json:"categoryId"`
	Difficulty      string    `gorm:"column:difficulty" json:"difficulty"`
	Status          string    `gorm:"column:status" json:"status"` // live/open/closed
	MaxPlayers      int       `gorm:"column:maxPlayers" json:"maxPlayers"`
	CurrentPlayers  int       `gorm:"column:currentPlayers" json:"currentPlayers"`
	DurationSeconds int       `gorm:"column:durationSeconds" json:"durationSeconds"`
	Description     string    `gorm:"column:description" json:"description"`
	CreatedAt       time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	Category QuizCategory `gorm:"foreignKey:CategoryID" json:"category"`
}

type Quiz struct {
	ID         string    `gorm:"primaryKey;column:id" json:"id"`
	Title      string    `gorm:"column:title" json:"title"`
	ArenaID    string    `gorm:"column:arenaId" json:"arenaId"`
	CategoryID string    `gorm:"column:categoryId" json:"categoryId"`
	Difficulty string    `gorm:"column:difficulty" json:"difficulty"`
	IsActive   bool      `gorm:"column:isActive" json:"isActive"`
	CreatedAt  time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

type Question struct {
	ID            string    `gorm:"primaryKey;column:id" json:"id"`
	QuizID        string    `gorm:"column:quizId" json:"quizId"`
	Prompt        string    `gorm:"column:prompt" json:"prompt"`
	Type          string    `gorm:"column:type" json:"type"` // mcq, subjective
	CorrectAnswer string    `gorm:"column:correctAnswer" json:"correctAnswer"`
	Explanation   string    `gorm:"column:explanation" json:"explanation"`
	MaxScore      int       `gorm:"column:maxScore" json:"maxScore"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	Options []Option `gorm:"foreignKey:QuestionID" json:"options"`
}

type Option struct {
	ID         string `gorm:"primaryKey;column:id" json:"id"`
	QuestionID string `gorm:"column:questionId" json:"questionId"`
	Text       string `gorm:"column:text" json:"text"`
	IsCorrect  bool   `gorm:"column:isCorrect" json:"isCorrect"`
}

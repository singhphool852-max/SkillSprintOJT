package models

import (
	"time"
)

type Attempt struct {
main
	ID             string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	UserID         string    `gorm:"type:varchar(191);index;column:userId" json:"userId"`
	QuizID         string    `gorm:"type:varchar(191);index;column:quizId" json:"quizId"`

	ID             string    `gorm:"primaryKey;column:id;type:varchar(191)" json:"id"`
	UserID         string    `gorm:"column:userId;type:varchar(191)" json:"userId"`
	QuizID         string    `gorm:"column:quizId;type:varchar(191)" json:"quizId"`
main
	Score          int       `gorm:"column:score" json:"score"`
	TotalQuestions int       `gorm:"column:totalQuestions" json:"totalQuestions"`
	StartedAt      time.Time `gorm:"column:startedAt" json:"startedAt"`
	CompletedAt    time.Time `gorm:"column:completedAt" json:"completedAt"`

main
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Quiz Quiz `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE" json:"quiz,omitempty"`
}

type AttemptAnswer struct {
	ID               string `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	AttemptID        string `gorm:"type:varchar(191);index;column:attemptId" json:"attemptId"`
	QuestionID       string `gorm:"type:varchar(191);index;column:questionId" json:"questionId"`
	SelectedOptionID string `gorm:"type:varchar(191);index;column:selectedOptionId" json:"selectedOptionId"`

	User User `gorm:"-" json:"user,omitempty"`
	Quiz Quiz `gorm:"-" json:"quiz,omitempty"`
}

type AttemptAnswer struct {
	ID               string `gorm:"primaryKey;column:id;type:varchar(191)" json:"id"`
	AttemptID        string `gorm:"column:attemptId;type:varchar(191)" json:"attemptId"`
	QuestionID       string `gorm:"column:questionId;type:varchar(191)" json:"questionId"`
	SelectedOptionID string `gorm:"column:selectedOptionId;type:varchar(191)" json:"selectedOptionId"`
main
	WrittenAnswer    string `gorm:"column:writtenAnswer" json:"writtenAnswer"`
	IsCorrect        bool   `gorm:"column:isCorrect" json:"isCorrect"`
	Score            int    `gorm:"column:score" json:"score"`
	Feedback         string `gorm:"column:feedback" json:"feedback"`
	Explanation      string `gorm:"column:explanation" json:"explanation"`
	EvaluatedBy      string `gorm:"column:evaluatedBy" json:"evaluatedBy"` // AI or system
}

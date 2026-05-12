package models

import (
	"time"
)

// ──────────────────────────────────────────────
// UserWrongQuestion — tracks every wrong/skipped answer
// from arena tests. Foundation for personalized training.
// ──────────────────────────────────────────────
type UserWrongQuestion struct {
	ID             string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	UserID         string    `gorm:"type:varchar(191);uniqueIndex:idx_user_question;column:userId" json:"userId"`
	AttemptID      string    `gorm:"type:varchar(191);index;column:attemptId" json:"attemptId"`
	QuestionID     string    `gorm:"type:varchar(191);uniqueIndex:idx_user_question;column:questionId" json:"questionId"`
	TestID         string    `gorm:"type:varchar(191);index;column:testId" json:"testId"`
	TopicID        string    `gorm:"type:varchar(191);index;column:topicId" json:"topicId"`
	ID             string    `gorm:"primaryKey;column:id;type:varchar(191)" json:"id"`
	UserID         string    `gorm:"column:userId;type:varchar(191);uniqueIndex:idx_user_question" json:"userId"`
	AttemptID      string    `gorm:"column:attemptId;type:varchar(191)" json:"attemptId"`
	QuestionID     string    `gorm:"column:questionId;type:varchar(191);uniqueIndex:idx_user_question" json:"questionId"`
	TestID         string    `gorm:"column:testId;type:varchar(191)" json:"testId"`
	TopicID        string    `gorm:"column:topicId;type:varchar(191);index" json:"topicId"`
	QuestionType   string    `gorm:"column:questionType" json:"questionType"`       // "mcq" or "coding"
	Difficulty     string    `gorm:"column:difficulty" json:"difficulty"`            // test-level difficulty
	QuestionTitle  string    `gorm:"column:questionTitle" json:"questionTitle"`      // snapshot of question title
	UserAnswer     string    `gorm:"column:userAnswer;type:text" json:"userAnswer"`  // selected option text or code
	CorrectAnswer  string    `gorm:"column:correctAnswer;type:text" json:"correctAnswer"`
	Verdict        string    `gorm:"column:verdict" json:"verdict"`                 // "wrong_answer", "time_limit", "skipped", "compile_error"
	PointsLost     int       `gorm:"column:pointsLost" json:"pointsLost"`
	PointsPossible int       `gorm:"column:pointsPossible" json:"pointsPossible"`
	ReviewCount    int       `gorm:"column:reviewCount;default:0" json:"reviewCount"`
	CorrectStreak  int       `gorm:"column:correctStreak;default:0" json:"correctStreak"` // Number of times answered correctly in training
	WrongCount     int       `gorm:"column:wrongCount;default:1" json:"wrongCount"`       // Total times failed across all tests
	LastReviewedAt time.Time `gorm:"column:lastReviewedAt" json:"lastReviewedAt"`
	MasteredAt     time.Time `gorm:"column:masteredAt" json:"masteredAt"`
	CreatedAt      time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`

	User     User         `gorm:"-" json:"user,omitempty"`
	Question TestQuestion `gorm:"-" json:"question,omitempty"`
	Test     Test         `gorm:"-" json:"test,omitempty"`
}

func (UserWrongQuestion) TableName() string {
	return "user_wrong_questions"
}

// ──────────────────────────────────────────────
// UserTopicStats — pre-computed weak-topic analysis
// per user. Updated after each test submission.
// ──────────────────────────────────────────────
type UserTopicStats struct {
	ID              string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	UserID          string    `gorm:"type:varchar(191);uniqueIndex:idx_user_topic_stats;column:userId" json:"userId"`
	TopicID         string    `gorm:"type:varchar(191);uniqueIndex:idx_user_topic_stats;column:topicId" json:"topicId"`
	ID              string    `gorm:"primaryKey;column:id;type:varchar(191)" json:"id"`
	UserID          string    `gorm:"column:userId;type:varchar(191);uniqueIndex:idx_user_topic_stats" json:"userId"`
	TopicID         string    `gorm:"column:topicId;type:varchar(191);uniqueIndex:idx_user_topic_stats" json:"topicId"`
	TopicName       string    `gorm:"column:topicName" json:"topicName"`
	TotalAttempted  int       `gorm:"column:totalAttempted" json:"totalAttempted"`
	TotalCorrect    int       `gorm:"column:totalCorrect" json:"totalCorrect"`
	TotalWrong      int       `gorm:"column:totalWrong" json:"totalWrong"`
	TotalSkipped    int       `gorm:"column:totalSkipped" json:"totalSkipped"`
	AccuracyPercent float64   `gorm:"column:accuracyPercent" json:"accuracyPercent"`
	WeakLevel       string    `gorm:"column:weakLevel" json:"weakLevel"` // "strong", "moderate", "weak", "critical"
	LastAttemptedAt time.Time `gorm:"column:lastAttemptedAt" json:"lastAttemptedAt"`
	UpdatedAt       time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	User  User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Topic *Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE" json:"topic,omitempty"`
	User  User   `gorm:"-" json:"user,omitempty"`
	Topic *Topic `gorm:"-" json:"topic,omitempty"`
}

func (UserTopicStats) TableName() string {
	return "user_topic_stats"
}

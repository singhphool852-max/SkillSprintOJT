package models

import (
	"time"
)

// ──────────────────────────────────────────────
// Test — a timed assessment created by an admin
// ──────────────────────────────────────────────
type Test struct {
	ID              string     `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	Title           string     `gorm:"column:title" json:"title"`
	Description     string     `gorm:"column:description" json:"description"`
	TopicID         string     `gorm:"type:varchar(191);index;column:topicId" json:"topicId,omitempty"`
	StartTime       *time.Time `gorm:"column:startTime" json:"startTime"`
	DurationSeconds int        `gorm:"column:durationSeconds" json:"durationSeconds"`
	Difficulty      string     `gorm:"column:difficulty" json:"difficulty"` // "easy", "medium", "hard"
	MaxScore        int        `gorm:"column:maxScore" json:"maxScore"`
	IsPublished     bool       `gorm:"column:isPublished;default:false" json:"isPublished"`
	IsActive        bool       `gorm:"column:isActive;default:false" json:"isActive"`
	CreatedBy       string     `gorm:"type:varchar(191);index;column:createdBy" json:"createdBy"`
	CreatedAt       time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	DeletedAt       *time.Time `gorm:"column:deletedAt;index" json:"deletedAt,omitempty"`
	DeletedBy       string     `gorm:"type:varchar(191);column:deletedBy" json:"deletedBy,omitempty"`

	Creator   User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Topic     *Topic         `gorm:"foreignKey:TopicID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"topic,omitempty"`
	Questions []TestQuestion `gorm:"foreignKey:TestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"questions,omitempty"`
}

func (Test) TableName() string {
	return "tests"
}

// ──────────────────────────────────────────────
// TestQuestion — one question inside a test
// ──────────────────────────────────────────────
type TestQuestion struct {
	ID          string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	TestID      string    `gorm:"type:varchar(191);index;column:testId" json:"testId"`
	Type        string    `gorm:"column:type" json:"type"` // "mcq" or "coding"
	Position    int       `gorm:"column:position" json:"position"`
	Title       string    `gorm:"column:title" json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	Points      int       `gorm:"column:points" json:"points"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	Test         Test              `gorm:"foreignKey:TestID" json:"-"`
	MCQOptions   []TestMCQOption   `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"mcqOptions,omitempty"`
	CodingDetail *TestCodingDetail `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"codingDetail,omitempty"`
	TestCases    []TestCase        `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"testCases,omitempty"`
}

func (TestQuestion) TableName() string {
	return "test_questions"
}

// ──────────────────────────────────────────────
// TestMCQOption — one option for an MCQ question
// ──────────────────────────────────────────────
type TestMCQOption struct {
	ID         string `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	QuestionID string `gorm:"type:varchar(191);index;column:questionId" json:"questionId"`
	OptionText string `gorm:"column:optionText" json:"optionText"`
	IsCorrect  bool   `gorm:"column:isCorrect" json:"isCorrect"`
}

func (TestMCQOption) TableName() string {
	return "test_mcq_options"
}

// ──────────────────────────────────────────────
// TestCodingDetail — metadata for a coding question
// ──────────────────────────────────────────────
type TestCodingDetail struct {
	ID          string `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	QuestionID  string `gorm:"type:varchar(191);unique;index;column:questionId" json:"questionId"`
	Constraints string `gorm:"column:constraints" json:"constraints"`
	StarterCode string `gorm:"column:starterCode" json:"starterCode"`
	TimeLimitMs int    `gorm:"column:timeLimitMs" json:"timeLimitMs"`
}

func (TestCodingDetail) TableName() string {
	return "test_coding_details"
}

// ──────────────────────────────────────────────
// TestCase — input/output pair for code evaluation
// ──────────────────────────────────────────────
type TestCase struct {
	ID             string `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	QuestionID     string `gorm:"type:varchar(191);index;column:questionId" json:"questionId"`
	Input          string `gorm:"column:input" json:"input"`
	ExpectedOutput string `gorm:"column:expectedOutput" json:"expectedOutput"`
	IsHidden       bool   `gorm:"column:isHidden" json:"isHidden"`
}

func (TestCase) TableName() string {
	return "test_cases"
}

// ──────────────────────────────────────────────
// TestAttempt — a user's attempt at a test
// ──────────────────────────────────────────────
type TestAttempt struct {
	ID              string     `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	UserID          string     `gorm:"type:varchar(191);uniqueIndex:idx_user_test;column:userId" json:"userId"`
	TestID          string     `gorm:"type:varchar(191);uniqueIndex:idx_user_test;column:testId" json:"testId"`
	Mode            string     `gorm:"column:mode;default:arena" json:"mode"` // "arena" (ranked, single) | "practice" | "train"
	StartedAt       time.Time  `gorm:"column:startedAt" json:"startedAt"`
	SubmittedAt     *time.Time `gorm:"column:submittedAt" json:"submittedAt"`
	Score           int       `gorm:"column:score" json:"score"`
	TotalQuestions  int       `gorm:"column:totalQuestions" json:"totalQuestions"`
	TimeTaken       int       `gorm:"column:timeTaken" json:"timeTaken"` // seconds
	ViolationCount  int       `gorm:"column:violationCount;default:0" json:"violationCount"`
	IsAutoSubmitted bool      `gorm:"column:isAutoSubmitted;default:false" json:"isAutoSubmitted"`

	User        User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Test        Test             `gorm:"foreignKey:TestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"test,omitempty"`
	Submissions []TestSubmission `gorm:"foreignKey:AttemptID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"submissions,omitempty"`
}

func (TestAttempt) TableName() string {
	return "test_attempts"
}

// ──────────────────────────────────────────────
// TestSubmission — answer to one question within an attempt
// ──────────────────────────────────────────────
type TestSubmission struct {
	ID               string `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	AttemptID        string `gorm:"type:varchar(191);index;column:attemptId" json:"attemptId"`
	QuestionID       string `gorm:"type:varchar(191);index;column:questionId" json:"questionId"`
	Type             string `gorm:"column:type" json:"type"` // "mcq" or "coding"
	SelectedOptionID string `gorm:"type:varchar(191);index;column:selectedOptionId" json:"selectedOptionId"`
	Code             string `gorm:"column:code" json:"code"`
	Language         string `gorm:"column:language" json:"language"`
	Verdict          string `gorm:"column:verdict" json:"verdict"` // "accepted", "wrong_answer", "time_limit", "pending"
	PassedCount      int    `gorm:"column:passedCount" json:"passedCount"`
	TotalCount       int    `gorm:"column:totalCount" json:"totalCount"`
	Score            int    `gorm:"column:score" json:"score"`

	Attempt  TestAttempt  `gorm:"foreignKey:AttemptID" json:"-"`
	Question TestQuestion `gorm:"foreignKey:QuestionID" json:"-"`
}

func (TestSubmission) TableName() string {
	return "test_submissions"
}

// ──────────────────────────────────────────────
// TestResult — persisted result summary after grading
// ──────────────────────────────────────────────
type TestResult struct {
	ID              string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	AttemptID       string    `gorm:"type:varchar(191);unique;index;column:attemptId" json:"attemptId"`
	UserID          string    `gorm:"type:varchar(191);index;column:userId" json:"userId"`
	TestID          string    `gorm:"type:varchar(191);index;column:testId" json:"testId"`
	TotalScore      int       `gorm:"column:totalScore" json:"totalScore"`
	MaxPossible     int       `gorm:"column:maxPossible" json:"maxPossible"`
	Percentage      float64   `gorm:"column:percentage" json:"percentage"`
	Rank            int       `gorm:"column:rank" json:"rank"`
	MCQCorrect      int       `gorm:"column:mcqCorrect" json:"mcqCorrect"`
	MCQTotal        int       `gorm:"column:mcqTotal" json:"mcqTotal"`
	CodingPassed    int       `gorm:"column:codingPassed" json:"codingPassed"`
	CodingTotal     int       `gorm:"column:codingTotal" json:"codingTotal"`
	IsAutoSubmitted bool      `gorm:"column:isAutoSubmitted" json:"isAutoSubmitted"`
	CalculatedAt    time.Time `gorm:"column:calculatedAt;autoCreateTime" json:"calculatedAt"`

	User    User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Test    Test        `gorm:"foreignKey:TestID" json:"test,omitempty"`
	Attempt TestAttempt `gorm:"foreignKey:AttemptID" json:"attempt,omitempty"`
}

func (TestResult) TableName() string {
	return "test_results"
}

// ──────────────────────────────────────────────
// TestViolation — anti-cheat violation log entry
// ──────────────────────────────────────────────
type TestViolation struct {
	ID            string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	AttemptID     string    `gorm:"type:varchar(191);column:attemptId;index" json:"attemptId"`
	UserID        string    `gorm:"type:varchar(191);index;column:userId" json:"userId"`
	TestID        string    `gorm:"type:varchar(191);index;column:testId" json:"testId"`
	ViolationType string    `gorm:"column:violationType" json:"violationType"`
	Timestamp     time.Time `gorm:"column:timestamp" json:"timestamp"`
	RemainingTime int       `gorm:"column:remainingTime" json:"remainingTime"`
	CreatedAt     time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

func (TestViolation) TableName() string {
	return "test_violations"
}

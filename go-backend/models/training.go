package models

import "time"

// TrainingQuestion holds seeded / AI-generated questions used by the training module.
// It is intentionally separate from the arena Question model.
type TrainingQuestion struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Topic       string    `gorm:"type:varchar(191);index;column:topic" json:"topic"`
	Type        string    `gorm:"column:type" json:"type"`        // mcq | debug | code_output | short_answer | logic
	Difficulty  string    `gorm:"column:difficulty" json:"difficulty"` // easy | medium | hard
	Prompt      string    `gorm:"column:prompt;type:text" json:"prompt"`
	Options     string    `gorm:"column:options;type:text" json:"options"`     // JSON array string e.g. ["A","B","C","D"]
	Answer      string    `gorm:"column:answer;type:text" json:"answer"`
	Explanation string    `gorm:"column:explanation;type:text" json:"explanation"`
	Source      string    `gorm:"column:source" json:"source"` // seeded | ai | notes
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (TrainingQuestion) TableName() string { return "training_questions" }

// TrainingSession tracks a user's practice session.
type TrainingSession struct {
	SessionID   string    `gorm:"type:varchar(191);primaryKey;column:session_id" json:"session_id"` // UUID
	Topic       string    `gorm:"type:varchar(191);index;column:topic" json:"topic"`
	QuestionIDs string    `gorm:"column:question_ids;type:text" json:"question_ids"` // JSON array e.g. [1,2,3]
	Status      string    `gorm:"column:status" json:"status"`                       // pending | completed
	Score       int       `gorm:"column:score" json:"score"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (TrainingSession) TableName() string { return "training_sessions" }

// Upload tracks note files uploaded by users for AI question generation.
type Upload struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Filename           string    `gorm:"column:filename" json:"filename"`
	Topic              string    `gorm:"type:varchar(191);index;column:topic" json:"topic"`
	Status             string    `gorm:"column:status" json:"status"` // pending | processing | done | failed
	ExtractedText      string    `gorm:"column:extracted_text;type:text" json:"extracted_text"`
	Summary            string    `gorm:"column:summary;type:text" json:"summary"`
	QuestionsGenerated int       `gorm:"column:questions_generated" json:"questions_generated"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Upload) TableName() string { return "uploads" }

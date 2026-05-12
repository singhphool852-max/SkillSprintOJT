package models

import (
	"time"
)

// ChatMessage represents a message in the global community chat
type ChatMessage struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;index" json:"userId"`
	Username    string    `gorm:"type:varchar(255);not null" json:"username"`
	Avatar      string    `gorm:"type:varchar(255)" json:"avatar"`
	MessageType string    `gorm:"type:varchar(20);default:'text'" json:"messageType"` // "text", "note", "image", "pdf"
	Content     string    `gorm:"type:text" json:"content"`                           // text content or file URL
	FileName    string    `gorm:"type:varchar(255)" json:"fileName"`                  // original filename for files
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

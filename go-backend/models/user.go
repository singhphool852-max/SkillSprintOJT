package models

import (
	"time"
)

type User struct {
	ID           string    `gorm:"type:varchar(191);primaryKey;column:id" json:"id"`
	Email        string    `gorm:"type:varchar(191);unique;column:email" json:"email"`
	Password     string    `gorm:"column:password" json:"-"`
	Username     string    `gorm:"column:username" json:"username"`
	Role         string    `gorm:"column:role;default:user" json:"role"`
	AuthProvider string    `gorm:"column:authProvider;default:local" json:"authProvider"` // "local" or "google"
	GoogleID     string    `gorm:"column:googleId" json:"-"`                              // Google sub claim
	AvatarURL    string    `gorm:"column:avatarUrl" json:"avatarUrl,omitempty"`
	CreatedAt    time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

// Ensure the table name exactly matches what's in dev.db
func (User) TableName() string {
	return "user"
}

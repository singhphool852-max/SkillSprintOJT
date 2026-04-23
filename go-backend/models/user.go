package models

import (
	"time"
)

type User struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Email     string    `gorm:"unique;column:email" json:"email"`
	Password  string    `gorm:"column:password" json:"-"`
	Username  string    `gorm:"column:username" json:"username"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

// Ensure the table name exactly matches what's in dev.db
func (User) TableName() string {
	return "user"
}

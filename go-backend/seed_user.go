//go:build ignore

package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID           string    `gorm:"primaryKey;column:id" json:"id"`
	Email        string    `gorm:"unique;column:email" json:"email"`
	Password     string    `gorm:"column:password" json:"-"`
	Username     string    `gorm:"column:username" json:"username"`
	Role         string    `gorm:"column:role;default:user" json:"role"`
	AuthProvider string    `gorm:"column:authProvider;default:local" json:"authProvider"`
	GoogleID     string    `gorm:"column:googleId" json:"-"`
	AvatarURL    string    `gorm:"column:avatarUrl" json:"avatarUrl,omitempty"`
	CreatedAt    time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
}

func (User) TableName() string {
	return "user"
}

func main() {
	db, err := gorm.Open(sqlite.Open("../dev.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		panic("failed to hash password")
	}

	var existingUser User
	result := db.Where("email = ?", "singhphool852@gmail.com").First(&existingUser)
	if result.Error == nil {
		// User exists, update password and role
		existingUser.Password = string(hashed)
		existingUser.Role = "admin"
		db.Save(&existingUser)
		fmt.Println("User updated successfully to admin with password: password123")
	} else {
		user := User{
			ID:           uuid.New().String(),
			Email:        "singhphool852@gmail.com",
			Username:     "nishant",
			Password:     string(hashed),
			Role:         "admin",
			AuthProvider: "local",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		createResult := db.Create(&user)
		if createResult.Error != nil {
			fmt.Println("Error creating user:", createResult.Error)
		} else {
			fmt.Println("User created successfully with password: password123")
		}
	}
}

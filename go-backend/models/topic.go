package models

import (
	"time"
)

// ──────────────────────────────────────────────
// Topic — a category/subject grouping for tests
// Created by admins to organise tests.
// ──────────────────────────────────────────────
type Topic struct {
	ID          string    `gorm:"primaryKey;column:id;type:varchar(191)" json:"id"`
	Name        string    `gorm:"unique;column:name;type:varchar(191)" json:"name"`
	Slug        string    `gorm:"unique;column:slug;type:varchar(191)" json:"slug"`
	Description string    `gorm:"column:description" json:"description"`
	CreatedBy   string    `gorm:"column:createdBy;type:varchar(191)" json:"createdBy"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	Creator User   `gorm:"-" json:"creator,omitempty"`
	Tests   []Test `gorm:"-" json:"tests,omitempty"`
}

func (Topic) TableName() string {
	return "topics"
}

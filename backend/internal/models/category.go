package models

import "time"

// Category represents a blog post category.
type Category struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null;size:100" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}
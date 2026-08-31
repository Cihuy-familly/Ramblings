package models

import "time"

// Creator represents a blog content creator/author.
type Creator struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"`
	DisplayName  string    `gorm:"not null;size:100" json:"display_name"`
	AvatarURL    string    `gorm:"type:text;default:''" json:"avatar_url"`
	Bio          string    `gorm:"type:text;default:''" json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
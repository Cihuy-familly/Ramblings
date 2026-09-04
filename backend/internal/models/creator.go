package models

import "time"

// Creator represents a blog content creator/author.
type Creator struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"size:255;default:''" json:"-"`
	GoogleID     string    `gorm:"uniqueIndex;size:255;default:''" json:"-"`
	Slug         string    `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	DisplayName  string    `gorm:"not null;size:100" json:"display_name"`
	AvatarURL    string    `gorm:"type:text;default:''" json:"avatar_url"`
	Bio          string    `gorm:"type:text;default:''" json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Follow represents a follower relationship between creators.
type Follow struct {
	FollowerID  string    `gorm:"type:uuid;primaryKey" json:"follower_id"`
	CreatorID   string    `gorm:"type:uuid;primaryKey" json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`

	// Relations (not serialized to keep response clean)
	Follower    Creator   `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE" json:"-"`
	Creator     Creator   `gorm:"foreignKey:CreatorID;constraint:OnDelete:CASCADE" json:"-"`
}
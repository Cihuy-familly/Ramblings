package models

import "time"

// Post represents a blog post.
type Post struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatorID    string    `gorm:"type:uuid;not null;index" json:"creator_id"`
	Title        string    `gorm:"not null;size:255" json:"title"`
	Slug         string    `gorm:"uniqueIndex;not null;size:255" json:"slug"`
	Content      string    `gorm:"not null;type:text" json:"content"`
	Excerpt      string    `gorm:"type:text;default:''" json:"excerpt"`
	ThumbnailURL string    `gorm:"type:text;default:''" json:"thumbnail_url"`
	Published    bool      `gorm:"default:false" json:"published"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Creator    Creator    `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"creator"`
	Categories []Category `gorm:"many2many:post_categories;" json:"categories"`
}
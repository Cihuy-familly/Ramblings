package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog-platform-backend/internal/config"
	"blog-platform-backend/internal/models"
)

// ChannelHandler handles channel-related endpoints.
type ChannelHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewChannelHandler creates a new ChannelHandler.
func NewChannelHandler(db *gorm.DB, cfg *config.Config) *ChannelHandler {
	return &ChannelHandler{db: db, cfg: cfg}
}

// channelResponse represents a public channel profile.
type channelResponse struct {
	ID            string              `json:"id"`
	Slug          string              `json:"slug"`
	DisplayName   string              `json:"display_name"`
	AvatarURL     string              `json:"avatar_url"`
	Bio           string              `json:"bio"`
	FollowerCount int64               `json:"follower_count"`
	Posts         []channelPostItem   `json:"posts"`
	CreatedAt     string              `json:"created_at"`
}

type channelPostItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Excerpt      string `json:"excerpt"`
	ThumbnailURL string `json:"thumbnail_url"`
	CreatedAt    string `json:"created_at"`
}

// GetChannel returns a creator's public channel profile and their published posts.
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	slug := c.Param("slug")

	var creator models.Creator
	if err := h.db.Where("slug = ?", slug).First(&creator).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	// Count followers
	var followerCount int64
	h.db.Model(&models.Follow{}).Where("creator_id = ?", creator.ID).Count(&followerCount)

	// Get published posts
	var posts []models.Post
	h.db.Where("creator_id = ? AND published = ?", creator.ID, true).
		Order("created_at DESC").
		Limit(50).
		Find(&posts)

	items := make([]channelPostItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, channelPostItem{
			ID:           p.ID,
			Title:        p.Title,
			Slug:         p.Slug,
			Excerpt:      p.Excerpt,
			ThumbnailURL: p.ThumbnailURL,
			CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"channel": channelResponse{
			ID:            creator.ID,
			Slug:          creator.Slug,
			DisplayName:   creator.DisplayName,
			AvatarURL:     creator.AvatarURL,
			Bio:           creator.Bio,
			FollowerCount: followerCount,
			Posts:         items,
			CreatedAt:     creator.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// Follow follows a creator.
func (h *ChannelHandler) Follow(c *gin.Context) {
	creatorID := c.GetString("creator_id")
	slug := c.Param("slug")

	// Find the target creator
	var target models.Creator
	if err := h.db.Where("slug = ?", slug).First(&target).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	// Cannot follow yourself
	if target.ID == creatorID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	// Check if already following
	var existing models.Follow
	result := h.db.Where("follower_id = ? AND creator_id = ?", creatorID, target.ID).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already following this creator"})
		return
	}

	follow := models.Follow{
		FollowerID: creatorID,
		CreatorID:  target.ID,
	}
	if err := h.db.Create(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow"})
		return
	}

	// Return updated follower count
	var followerCount int64
	h.db.Model(&models.Follow{}).Where("creator_id = ?", target.ID).Count(&followerCount)

	c.JSON(http.StatusOK, gin.H{
		"message":        "followed successfully",
		"follower_count": followerCount,
	})
}

// Unfollow unfollows a creator.
func (h *ChannelHandler) Unfollow(c *gin.Context) {
	creatorID := c.GetString("creator_id")
	slug := c.Param("slug")

	var target models.Creator
	if err := h.db.Where("slug = ?", slug).First(&target).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	result := h.db.Where("follower_id = ? AND creator_id = ?", creatorID, target.ID).Delete(&models.Follow{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not following this creator"})
		return
	}

	var followerCount int64
	h.db.Model(&models.Follow{}).Where("creator_id = ?", target.ID).Count(&followerCount)

	c.JSON(http.StatusOK, gin.H{
		"message":        "unfollowed successfully",
		"follower_count": followerCount,
	})
}

// IsFollowing checks if the authenticated user follows a creator.
func (h *ChannelHandler) IsFollowing(c *gin.Context) {
	creatorID := c.GetString("creator_id")
	slug := c.Param("slug")

	var target models.Creator
	if err := h.db.Where("slug = ?", slug).First(&target).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var count int64
	h.db.Model(&models.Follow{}).Where("follower_id = ? AND creator_id = ?", creatorID, target.ID).Count(&count)

	var followerCount int64
	h.db.Model(&models.Follow{}).Where("creator_id = ?", target.ID).Count(&followerCount)

	c.JSON(http.StatusOK, gin.H{
		"is_following":   count > 0,
		"follower_count": followerCount,
	})
}
// StudioStats returns the authenticated creator's dashboard stats.
// Shows post counts and follower count — information without pressure.
func (h *ChannelHandler) StudioStats(c *gin.Context) {
	creatorID := c.GetString("creator_id")

	// Total posts
	var totalPosts int64
	h.db.Model(&models.Post{}).Where("creator_id = ?", creatorID).Count(&totalPosts)

	// Published vs draft
	var publishedPosts int64
	h.db.Model(&models.Post{}).Where("creator_id = ? AND published = ?", creatorID, true).Count(&publishedPosts)
	draftPosts := totalPosts - publishedPosts

	// Follower count
	var followerCount int64
	h.db.Model(&models.Follow{}).Where("creator_id = ?", creatorID).Count(&followerCount)

	// Following count
	var followingCount int64
	h.db.Model(&models.Follow{}).Where("follower_id = ?", creatorID).Count(&followingCount)

	// Recent posts (last 10)
	var recentPosts []models.Post
	h.db.Where("creator_id = ?", creatorID).
		Order("created_at DESC").
		Limit(10).
		Find(&recentPosts)

	type recentPostItem struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Slug      string `json:"slug"`
		Published bool   `json:"published"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	items := make([]recentPostItem, 0, len(recentPosts))
	for _, p := range recentPosts {
		items = append(items, recentPostItem{
			ID:        p.ID,
			Title:     p.Title,
			Slug:      p.Slug,
			Published: p.Published,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_posts":      totalPosts,
			"published_posts":  publishedPosts,
			"draft_posts":      draftPosts,
			"follower_count":   followerCount,
			"following_count":  followingCount,
		},
		"recent_posts": items,
	})
}

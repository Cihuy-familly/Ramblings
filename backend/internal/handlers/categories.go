package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"blog-platform-backend/internal/models"
)

// CategoryHandler handles category endpoints.
type CategoryHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(db *gorm.DB, rdb *redis.Client) *CategoryHandler {
	return &CategoryHandler{db: db, redis: rdb}
}

// CategoryWithCount represents a category with its published post count.
type CategoryWithCount struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"post_count"`
}

// ListCategories returns all categories with their published post counts.
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	ctx := context.Background()

	// Try cache
	cacheKey := "categories:list"
	cached, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
		return
	}

	// Query categories with published post count
	var results []CategoryWithCount
	if err := h.db.Model(&models.Category{}).
		Select("categories.id, categories.name, categories.slug, COUNT(posts.id) AS post_count").
		Joins("LEFT JOIN posts ON posts.category_id = categories.id AND posts.published = ?", true).
		Group("categories.id, categories.name, categories.slug").
		Order("categories.name ASC").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}

	if results == nil {
		results = []CategoryWithCount{}
	}

	response := gin.H{"categories": results}

	// Cache for 5 minutes
	jsonData, err := json.Marshal(response)
	if err == nil {
		h.redis.Set(ctx, cacheKey, jsonData, 5*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}
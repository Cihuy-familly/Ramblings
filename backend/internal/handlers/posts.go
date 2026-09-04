package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"blog-platform-backend/internal/models"
	"blog-platform-backend/internal/search"
)

// PostHandler handles blog post endpoints.
type PostHandler struct {
	db        *gorm.DB
	redis     *redis.Client
	searchSvc *search.Service
}

// NewPostHandler creates a new PostHandler.
func NewPostHandler(db *gorm.DB, rdb *redis.Client, svc *search.Service) *PostHandler {
	return &PostHandler{db: db, redis: rdb, searchSvc: svc}
}

// --- Response types ---

type creatorBrief struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type categoryBrief struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type postListItem struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Slug         string         `json:"slug"`
	Content      string         `json:"content"`
	Excerpt      string         `json:"excerpt"`
	ThumbnailURL string         `json:"thumbnail_url"`
	Published    bool           `json:"published"`
	Categories   []categoryBrief `json:"categories"`
	Creator      creatorBrief   `json:"creator"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type postDetail struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Slug         string         `json:"slug"`
	Content      string         `json:"content"`
	Excerpt      string         `json:"excerpt"`
	ThumbnailURL string         `json:"thumbnail_url"`
	Published    bool           `json:"published"`
	Categories   []categoryBrief `json:"categories"`
	Creator      creatorBrief   `json:"creator"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// --- Request types ---

type createPostRequest struct {
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content" binding:"required"`
	Excerpt      string `json:"excerpt"`
	CategoryIDs  []int  `json:"category_ids"`
	ThumbnailURL string `json:"thumbnail_url"`
	Published    bool   `json:"published"`
}

type updatePostRequest struct {
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content" binding:"required"`
	Excerpt      string `json:"excerpt"`
	CategoryIDs  []int  `json:"category_ids"`
	ThumbnailURL string `json:"thumbnail_url"`
	Published    bool   `json:"published"`
}

// --- Public endpoints ---

// ListPosts returns paginated published posts.
func (h *PostHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	categorySlug := c.Query("category")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}

	offset := (page - 1) * limit

	// Try cache
	cacheKey := fmt.Sprintf("posts:page:%d:limit:%d:category:%s", page, limit, categorySlug)
	ctx := context.Background()
	cached, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
		return
	}

	// Build query
	query := h.db.Where("published = ?", true)
	if categorySlug != "" {
		query = query.Joins("JOIN post_categories ON post_categories.post_id = posts.id").
			Joins("JOIN categories ON categories.id = post_categories.category_id").
			Where("categories.slug = ?", categorySlug).
			Distinct()
	}

	// Count total
	var total int64
	if err := query.Model(&models.Post{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count posts"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	// Fetch posts
	var posts []models.Post
	if err := query.
		Preload("Creator").
		Preload("Categories").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
		return
	}

	// Build response
	items := make([]postListItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, postToListItem(p))
	}

	response := gin.H{
		"posts":      items,
		"total":      total,
		"page":       page,
		"totalPages": totalPages,
	}

	// Cache for 2 minutes
	jsonData, err := json.Marshal(response)
	if err == nil {
		h.redis.Set(ctx, cacheKey, jsonData, 2*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}

// GetPost returns a single published post by slug.
func (h *PostHandler) GetPost(c *gin.Context) {
	slug := c.Param("slug")

	var post models.Post
	if err := h.db.
		Where("slug = ? AND published = ?", slug, true).
		Preload("Creator").
		Preload("Categories").
		First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": postToDetail(post)})
}

// --- Creator (auth) endpoints ---

// CreatorListPosts returns all posts for the authenticated creator (including drafts).
func (h *PostHandler) CreatorListPosts(c *gin.Context) {
	creatorID := c.GetString("creator_id")

	var posts []models.Post
	if err := h.db.
		Where("creator_id = ?", creatorID).
		Preload("Creator").
		Preload("Categories").
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
		return
	}

	items := make([]postListItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, postToListItem(p))
	}

	c.JSON(http.StatusOK, gin.H{"posts": items})
}

// CreatorCreatePost creates a new post for the authenticated creator.
func (h *PostHandler) CreatorCreatePost(c *gin.Context) {
	creatorID := c.GetString("creator_id")

	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: title and content are required"})
		return
	}

	// Minimum content length (200 chars) to prevent one-liners
	if len(strings.TrimSpace(req.Content)) < 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must be at least 200 characters"})
		return
	}

	slug := h.generateUniqueSlug(req.Title)

	// Look up categories
	var cats []models.Category
	if len(req.CategoryIDs) > 0 {
		h.db.Where("id IN ?", req.CategoryIDs).Find(&cats)
	}

	post := models.Post{
		CreatorID:    creatorID,
		Title:        req.Title,
		Slug:         slug,
		Content:      req.Content,
		Excerpt:      req.Excerpt,
		ThumbnailURL: req.ThumbnailURL,
		Published:    req.Published,
		Categories:   cats,
	}

	if err := h.db.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}

	// Reload with relationships
	h.db.Preload("Creator").Preload("Categories").First(&post, "id = ?", post.ID)

	// Index in Meilisearch
	go h.indexPost(post)

	// Invalidate caches
	h.invalidatePostCaches(ctxForCache())

	c.JSON(http.StatusCreated, gin.H{"post": postToDetail(post)})
}

// CreatorUpdatePost updates an existing post. Only the owning creator can update.
func (h *PostHandler) CreatorUpdatePost(c *gin.Context) {
	creatorID := c.GetString("creator_id")
	postID := c.Param("id")

	var post models.Post
	if err := h.db.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	if post.CreatorID != creatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own posts"})
		return
	}

	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: title and content are required"})
		return
	}

		// Minimum content length check
		if len(strings.TrimSpace(req.Content)) < 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content must be at least 200 characters"})
			return
		}

	// Regenerate slug only if title changed
	newSlug := post.Slug
	if req.Title != post.Title {
		newSlug = h.generateUniqueSlug(req.Title)
	}

	updates := map[string]interface{}{
		"title":         req.Title,
		"slug":          newSlug,
		"content":       req.Content,
		"excerpt":       req.Excerpt,
		"thumbnail_url": req.ThumbnailURL,
		"published":     req.Published,
	}

	if err := h.db.Model(&post).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update post"})
		return
	}

	// Replace categories via association
	var cats []models.Category
	if len(req.CategoryIDs) > 0 {
		h.db.Where("id IN ?", req.CategoryIDs).Find(&cats)
	}
	h.db.Model(&post).Association("Categories").Replace(cats)

	// Reload with relationships
	h.db.Preload("Creator").Preload("Categories").First(&post, "id = ?", post.ID)

	// Index in Meilisearch
	go h.indexPost(post)

	// Invalidate caches
	h.invalidatePostCaches(ctxForCache())

	c.JSON(http.StatusOK, gin.H{"post": postToDetail(post)})
}

// CreatorDeletePost deletes a post. Only the owning creator can delete.
func (h *PostHandler) CreatorDeletePost(c *gin.Context) {
	creatorID := c.GetString("creator_id")
	postID := c.Param("id")

	var post models.Post
	if err := h.db.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	if post.CreatorID != creatorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own posts"})
		return
	}

	if err := h.db.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}

	// Remove from Meilisearch
	go h.searchSvc.RemovePost(post.ID)

	// Invalidate caches
	h.invalidatePostCaches(ctxForCache())

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Helpers ---

// indexPost indexes a post in Meilisearch, loading creator and category names.
func (h *PostHandler) indexPost(post models.Post) {
	// Load relationships if not already loaded
	if post.Creator.ID == "" {
		h.db.Preload("Creator").Preload("Categories").First(&post, "id = ?", post.ID)
	}
	creatorName := ""
	if post.Creator.DisplayName != "" {
		creatorName = post.Creator.DisplayName
	}
	categoryNames := []string{}
	for _, cat := range post.Categories {
		categoryNames = append(categoryNames, cat.Name)
	}
	h.searchSvc.IndexPost(post, creatorName, categoryNames)
}

func postToListItem(p models.Post) postListItem {
	item := postListItem{
		ID:           p.ID,
		Title:        p.Title,
		Slug:         p.Slug,
		Content:      p.Content,
		Excerpt:      p.Excerpt,
		ThumbnailURL: p.ThumbnailURL,
		Published:    p.Published,
		Creator: creatorBrief{
			ID:          p.Creator.ID,
			DisplayName: p.Creator.DisplayName,
			AvatarURL:   p.Creator.AvatarURL,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	for _, cat := range p.Categories {
		item.Categories = append(item.Categories, categoryBrief{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
		})
	}
	return item
}

func postToDetail(p models.Post) postDetail {
	d := postDetail{
		ID:           p.ID,
		Title:        p.Title,
		Slug:         p.Slug,
		Content:      p.Content,
		Excerpt:      p.Excerpt,
		ThumbnailURL: p.ThumbnailURL,
		Published:    p.Published,
		Creator: creatorBrief{
			ID:          p.Creator.ID,
			DisplayName: p.Creator.DisplayName,
			AvatarURL:   p.Creator.AvatarURL,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	for _, cat := range p.Categories {
		d.Categories = append(d.Categories, categoryBrief{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
		})
	}
	return d
}

// generateUniqueSlug creates a URL-safe slug from a title, appending a random
// suffix if the slug already exists in the database.
func (h *PostHandler) generateUniqueSlug(title string) string {
	slug := slugify(title)
	if slug == "" {
		slug = "post"
	}

	var count int64
	h.db.Model(&models.Post{}).Where("slug = ?", slug).Count(&count)
	if count == 0 {
		return slug
	}

	// Append random suffix
	suffix := randomHex(4)
	return fmt.Sprintf("%s-%s", slug, suffix)
}

func slugify(s string) string {
	// Lowercase
	slug := strings.ToLower(s)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove anything that's not alphanumeric or hyphen
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "")
	// Collapse consecutive hyphens
	re = regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")
	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use timestamp-based suffix
		return fmt.Sprintf("%x", time.Now().UnixNano())[:n*2]
	}
	return hex.EncodeToString(b)
}

// invalidatePostCaches removes all cached post list and category data.
func (h *PostHandler) invalidatePostCaches(ctx context.Context) {
	// Delete all keys matching posts:* and categories:*
	keys, err := h.redis.Keys(ctx, "posts:*").Result()
	if err == nil && len(keys) > 0 {
		h.redis.Del(ctx, keys...)
	}
	keys, err = h.redis.Keys(ctx, "categories:*").Result()
	if err == nil && len(keys) > 0 {
		h.redis.Del(ctx, keys...)
	}
}

func ctxForCache() context.Context {
	return context.Background()
}
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blog-platform-backend/internal/search"
)

// SearchHandler handles search endpoints.
type SearchHandler struct {
	searchSvc *search.Service
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(svc *search.Service) *SearchHandler {
	return &SearchHandler{searchSvc: svc}
}

// Search handles GET /api/v1/search?q=...&page=1&limit=12
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}

	result, err := h.searchSvc.SearchAll(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":    result.Posts,
		"creators": result.Creators,
		"total":    result.Total,
		"page":     result.Page,
	})
}
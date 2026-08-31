package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"blog-platform-backend/internal/config"
	"blog-platform-backend/internal/models"
)

// AuthHandler handles authentication-related endpoints.
type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a creator and returns a JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: email and password are required"})
		return
	}

	var creator models.Creator
	if err := h.db.Where("email = ?", req.Email).First(&creator).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creator.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.generateToken(&creator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"creator": gin.H{
			"id":           creator.ID,
			"email":        creator.Email,
			"display_name": creator.DisplayName,
			"avatar_url":   creator.AvatarURL,
		},
	})
}

// Me returns the currently authenticated creator's profile.
func (h *AuthHandler) Me(c *gin.Context) {
	creatorID := c.GetString("creator_id")

	var creator models.Creator
	if err := h.db.First(&creator, "id = ?", creatorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "creator not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"creator": gin.H{
			"id":           creator.ID,
			"email":        creator.Email,
			"display_name": creator.DisplayName,
			"avatar_url":   creator.AvatarURL,
			"bio":          creator.Bio,
		},
	})
}

// generateToken creates a JWT token for the given creator.
func (h *AuthHandler) generateToken(creator *models.Creator) (string, error) {
	claims := jwt.MapClaims{
		"creator_id": creator.ID,
		"email":      creator.Email,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
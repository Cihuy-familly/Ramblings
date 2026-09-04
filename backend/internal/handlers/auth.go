package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

	if creator.PasswordHash == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "this account uses Google Sign-In, please login with Google"})
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
			"slug":         creator.Slug,
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
			"slug":         creator.Slug,
			"bio":          creator.Bio,
		},
	})
}

// GoogleLogin redirects the user to Google's OAuth consent screen.
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if h.cfg.GoogleClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth is not configured"})
		return
	}

	// Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	// Store state in cookie (HttpOnly, 10 min expiry)
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	// Build Google OAuth URL
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?"+
			"client_id=%s"+
			"&redirect_uri=%s"+
			"&response_type=code"+
			"&scope=email+profile"+
			"&state=%s"+
			"&access_type=online",
		url.QueryEscape(h.cfg.GoogleClientID),
		url.QueryEscape(h.cfg.GoogleCallbackURL),
		url.QueryEscape(state),
	)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles the OAuth callback from Google.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Verify state parameter (CSRF protection)
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid state cookie"})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	stateParam := c.Query("state")
	if stateParam == "" || stateParam != stateCookie {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch, possible CSRF attack"})
		return
	}

	// Check for error from Google
	if errMsg := c.Query("error"); errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google OAuth error: " + errMsg})
		return
	}

	// Get the authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	// Exchange authorization code for access token
	tokenResp, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code: " + err.Error()})
		return
	}

	// Fetch user info from Google
	userInfo, err := h.fetchGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user info: " + err.Error()})
		return
	}

	// Google must have verified the email
	if !userInfo.VerifiedEmail {
		c.JSON(http.StatusForbidden, gin.H{"error": "Google account email is not verified"})
		return
	}

	// Find or create creator by email
	creator, err := h.findOrCreateGoogleCreator(userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create/find user: " + err.Error()})
		return
	}

	// Generate JWT
	token, err := h.generateToken(creator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Redirect to frontend with token
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", h.cfg.FrontendURL, url.QueryEscape(token))
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// --- OAuth helpers ---

// generateState creates a random hex string for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// googleTokenResponse represents the token exchange response from Google.
type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// googleUserInfo represents the user info response from Google.
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// exchangeCodeForToken exchanges the authorization code for an access token.
func (h *AuthHandler) exchangeCodeForToken(code string) (*googleTokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.GoogleClientID},
		"client_secret": {h.cfg.GoogleClientSecret},
		"redirect_uri":  {h.cfg.GoogleCallbackURL},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// fetchGoogleUserInfo fetches the user's profile from Google using the access token.
func (h *AuthHandler) fetchGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request returned status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo googleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	if userInfo.Email == "" {
		return nil, fmt.Errorf("Google did not return an email address")
	}

	return &userInfo, nil
}

// findOrCreateGoogleCreator finds a creator by email or creates a new one.
// If the email exists but has no GoogleID, it links the Google account.
func (h *AuthHandler) findOrCreateGoogleCreator(info *googleUserInfo) (*models.Creator, error) {
	var creator models.Creator

	// Try to find by GoogleID first
	result := h.db.Where("google_id = ?", info.ID).First(&creator)
	if result.Error == nil {
		return &creator, nil
	}

	// Try to find by email
	result = h.db.Where("email = ?", info.Email).First(&creator)
	if result.Error == nil {
		// Link Google account to existing creator
		creator.GoogleID = info.ID
		if info.Name != "" && (creator.DisplayName == "" || creator.DisplayName == "Test Creator") {
			creator.DisplayName = info.Name
		}
		if info.Picture != "" {
			creator.AvatarURL = info.Picture
		}
		if err := h.db.Save(&creator).Error; err != nil {
			return nil, fmt.Errorf("failed to link Google account: %w", err)
		}
		return &creator, nil
	}

	// Create new creator
	displayName := info.Name
	if displayName == "" {
		displayName = info.Email[:len(info.Email)-10] // Use part before @ as name
	}

	newCreator := models.Creator{
		Email:       info.Email,
		GoogleID:    info.ID,
		Slug:        h.generateUniqueSlug(displayName),
			DisplayName: displayName,
		AvatarURL:   info.Picture,
	}

	if err := h.db.Create(&newCreator).Error; err != nil {
		return nil, fmt.Errorf("failed to create creator: %w", err)
	}

	return &newCreator, nil
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
// generateUniqueSlug creates a unique URL-safe slug from a display name.
// It appends a random suffix if the slug is already taken.
func (h *AuthHandler) generateUniqueSlug(displayName string) string {
	slug := models.GenerateSlug(displayName)
	for {
		var count int64
		h.db.Model(&models.Creator{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			return slug
		}
		slug = models.AppendRandomSuffix(slug)
	}
}

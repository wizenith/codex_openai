package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"taskqueue/internal/auth"
	"taskqueue/internal/database"
	"taskqueue/pkg/logger"
)

// AuthHandler handles authentication routes
type AuthHandler struct {
	DB           *pgxpool.Pool
	OAuthProvider *auth.OAuthProvider
	JWTSecret    string
}

// GoogleLogin initiates Google OAuth flow
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state, err := h.OAuthProvider.GenerateStateToken()
	if err != nil {
		logger.Error("failed to generate state token", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Store state in session cookie
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	authURL := h.OAuthProvider.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	ctx := context.Background()

	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state == "" || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Exchange code for token
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.OAuthProvider.Exchange(ctx, code)
	if err != nil {
		logger.Error("failed to exchange code", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
		return
	}

	// Get user info
	userInfo, err := h.OAuthProvider.GetUserInfo(ctx, token)
	if err != nil {
		logger.Error("failed to get user info", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Create or update user
	userID, err := database.CreateUser(ctx, h.DB, userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		logger.Error("failed to create user", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate JWT
	jwtToken, err := auth.GenerateToken(h.JWTSecret, userID)
	if err != nil {
		logger.Error("failed to generate token", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Set auth cookie
	c.SetCookie("auth_token", jwtToken, 86400, "/", "", false, true)

	// Redirect to dashboard
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusTemporaryRedirect, "/login")
}

// GetCurrentUser returns current user info
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	ctx := context.Background()
	var user struct {
		ID      int64     `json:"id"`
		Email   string    `json:"email"`
		Name    string    `json:"name"`
		Picture string    `json:"picture"`
		CreatedAt time.Time `json:"created_at"`
	}

	err := h.DB.QueryRow(ctx, `
		SELECT id, email, name, picture, created_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture, &user.CreatedAt,
	)

	if err != nil {
		logger.Error("failed to get user", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a signed JWT for the given user ID.
func GenerateToken(secret string, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

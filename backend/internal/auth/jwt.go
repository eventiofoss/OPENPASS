package auth

import (
	"errors"
	"os"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// DefaultTokenTTL is the current login session duration.
	DefaultTokenTTL = 24 * time.Hour
	// SessionCookieName is the cookie name used for organizer sessions.
	SessionCookieName = "eventio_jwt"
)

// Claims holds the organizer identity stored in the session token.
type Claims struct {
	Role models.OrganizerRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken signs a JWT for the given organizer.
func GenerateToken(organizer models.Organizer, ttl time.Duration) (string, error) {
	claims := Claims{
		Role: organizer.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   organizer.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret()))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ParseToken validates a signed organizer session token.
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(jwtSecret()), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func jwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "super-secret-development-key"
	}

	return secret
}

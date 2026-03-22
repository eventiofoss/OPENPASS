package middleware

import (
	"errors"
	"os"
	"slices"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// DefaultTokenTTL is the current login session duration.
	DefaultTokenTTL = 24 * time.Hour
	// SessionCookieName is the cookie name used for organizer sessions.
	SessionCookieName = "eventio_jwt"
	claimsContextKey  = "auth_claims"
)

var jwtSigningSecret = mustJWTSecret()

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

// ClaimsFromContext returns organizer claims stored by auth middleware.
func ClaimsFromContext(c *fiber.Ctx) (*Claims, bool) {
	claims, ok := c.Locals(claimsContextKey).(*Claims)
	return claims, ok
}

// RequireAuth ensures the request carries a valid organizer session.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies(SessionCookieName)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{"error": "Authentication required"},
			)
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{"error": "Invalid or expired session"},
			)
		}

		c.Locals(claimsContextKey, claims)
		return c.Next()
	}
}

// RequireRoles restricts access to the provided organizer roles.
func RequireRoles(roles ...models.OrganizerRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{"error": "Authentication required"},
			)
		}

		if !slices.Contains(roles, claims.Role) {
			return c.Status(fiber.StatusForbidden).JSON(
				fiber.Map{"error": "Insufficient permissions"},
			)
		}

		return c.Next()
	}
}

func jwtSecret() string {
	return jwtSigningSecret
}

func mustJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	return secret
}

package auth

import (
	"slices"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

const claimsContextKey = "auth_claims"

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

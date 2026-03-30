package api

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

// RegisterRequest is the expected signup payload.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Register creates a user account.
func (h *Handler) Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	user, err := h.Auth.Register(
		c.Context(), req.Name, req.Email, req.Password,
		req.Role,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		if errors.Is(err, service.ErrDuplicateEmail) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not create account"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Account created successfully",
		"user":      user,
		"organizer": user,
	})
}

// LoginRequest is the expected login payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user and issues a session cookie.
func (h *Handler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	token, err := h.Auth.Login(
		c.Context(), req.Email, req.Password,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{"error": "Invalid credentials"},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not login"},
		)
	}

	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Expires:  time.Now().Add(middleware.DefaultTokenTTL),
		HTTPOnly: true,
		Secure:   isProductionEnv(),
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	return c.JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}

// Me resolves the currently authenticated user.
func (h *Handler) Me(c *fiber.Ctx) error {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok || claims.Subject == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	user, err := h.Auth.GetUser(
		c.Context(), claims.Subject,
	)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Invalid or expired session"},
		)
	}

	return c.JSON(fiber.Map{
		"user":      user,
		"organizer": user,
	})
}

// Logout clears the user session cookie.
func (h *Handler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   isProductionEnv(),
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func isProductionEnv() bool {
	return strings.EqualFold(os.Getenv("APP_ENV"), "production")
}

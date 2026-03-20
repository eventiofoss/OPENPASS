package api

import (
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Handler holds dependencies for request handlers.
type Handler struct {
	DB *gorm.DB
}

// RegisterRequest is the expected organizer signup payload.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register creates an organizer account.
func (h *Handler) Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	if strings.TrimSpace(req.Name) == "" ||
		strings.TrimSpace(req.Email) == "" ||
		len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, email, and a password (min 8 chars) are required",
		})
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		12,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to secure password"},
		)
	}

	organizer := models.Organizer{
		Name:         strings.TrimSpace(req.Name),
		Email:        cleanEmail,
		PasswordHash: string(hashedPassword),
		Role:         models.OrganizerRoleOrganizer,
	}

	result := h.DB.Create(&organizer)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "An account with this email already exists",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not create account"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Account created successfully",
		"organizer": organizer,
	})
}

// LoginRequest is the expected organizer login payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates an organizer and issues a session cookie.
func (h *Handler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	var organizer models.Organizer
	result := h.DB.Where("email = ?", cleanEmail).First(&organizer)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Invalid credentials"},
		)
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(organizer.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Invalid credentials"},
		)
	}

	signedToken, err := middleware.GenerateToken(
		organizer,
		middleware.DefaultTokenTTL,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not login"},
		)
	}

	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    signedToken,
		Expires:  time.Now().Add(middleware.DefaultTokenTTL),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	return c.JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}

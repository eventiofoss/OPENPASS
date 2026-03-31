package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	guestSessionCookieName = "guest_session"
	guestSessionCookieTTL  = 180 * 24 * time.Hour
)

// RegisterAttendeeRequest is the public registration payload DTO.
type RegisterAttendeeRequest struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	FormData json.RawMessage `json:"form_data"`
}

// RegisterAttendee handles POST /api/events/:id/register (public route).
func (h *Handler) RegisterAttendee(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	req := new(RegisterAttendeeRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	// Set stateless guest session JWT using parsed credentials before processing
	// so it persists even if this registration is a duplicate
	setGuestSessionCookie(c, req.Email, req.Name)

	attendee, err := h.Registration.RegisterAttendee(
		c.Context(),
		service.RegisterAttendeeInput{
			EventID:             eventID,
			Name:                req.Name,
			Email:               req.Email,
			FormData:            req.FormData,
			AuthenticatedUserID: authenticatedUserIDFromSession(c),
		},
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRegistrationInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		if errors.Is(err, service.ErrRegistrationEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		if errors.Is(err, service.ErrRegistrationEventFull) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": "Event is full"},
			)
		}

		if errors.Is(err, service.ErrRegistrationRequiresCheckout) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": "This event requires checkout"},
			)
		}

		if errors.Is(err, service.ErrAlreadyRegistered) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": "You are already registered for this event"},
			)
		}

		slog.Error("attendee registration failed", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not register attendee"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registration successful",
		"attendee": fiber.Map{
			"id": attendee.ID,
		},
	})
}

func authenticatedUserIDFromSession(c *fiber.Ctx) *uuid.UUID {
	token := strings.TrimSpace(c.Cookies(middleware.SessionCookieName))
	if token == "" {
		return nil
	}

	claims, err := middleware.ParseToken(token)
	if err != nil || claims == nil || claims.Subject == "" {
		return nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil
	}

	return &userID
}

func setGuestSessionCookie(c *fiber.Ctx, email, name string) {
	// Generate stateless JWT holding guest ID
	token, err := middleware.GenerateGuestToken(strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name), guestSessionCookieTTL)
	if err != nil {
		return // Silently ignore guest token issuance failures
	}

	c.Cookie(&fiber.Cookie{
		Name:     guestSessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(guestSessionCookieTTL),
		HTTPOnly: true,
		Secure:   cookieSecure(c),
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func cookieSecure(c *fiber.Ctx) bool {
	if isProductionEnv() {
		return true
	}

	return strings.EqualFold(c.Protocol(), "https")
}

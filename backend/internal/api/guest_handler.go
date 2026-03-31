package api

import (
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// GuestSession reads the stateless guest_session JWT returning identity variables.
// This is called by SvelteKit's +layout.server.ts during SSR to hydrate guest
// state so the UI can lock forms and prevent duplicate registration attempts
// without needing a successful database execution footprint.
func (h *Handler) GuestSession(c *fiber.Ctx) error {
	tokenString := strings.TrimSpace(c.Cookies(guestSessionCookieName))
	if tokenString == "" {
		return c.JSON(fiber.Map{
			"guest":         nil,
			"registrations": []interface{}{},
		})
	}

	claims, err := middleware.ParseGuestToken(tokenString)
	if err != nil {
		// Logically treat expired/invalid guest sessions as nonexistent.
		return c.JSON(fiber.Map{
			"guest":         nil,
			"registrations": []interface{}{},
		})
	}

	var registeredEventIDs []string
	// Query the attendees table to find all events this email has registered for
	h.DB.Model(&models.Attendee{}).Where("email = ?", claims.Email).Pluck("event_id", &registeredEventIDs)

	return c.JSON(fiber.Map{
		"guest": fiber.Map{
			"email":                claims.Email,
			"name":                 claims.Name,
			"registered_event_ids": registeredEventIDs,
		},
		"registrations": []interface{}{},
	})
}

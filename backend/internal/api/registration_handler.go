package api

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	attendee, err := h.Registration.RegisterAttendee(
		c.Context(),
		service.RegisterAttendeeInput{
			EventID:  eventID,
			Name:     req.Name,
			Email:    req.Email,
			FormData: req.FormData,
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

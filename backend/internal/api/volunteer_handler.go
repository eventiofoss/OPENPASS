package api

import (
	"errors"

	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AssignVolunteerRequest is the expected payload for volunteer assignment.
type AssignVolunteerRequest struct {
	Email string `json:"email"`
}

// AssignVolunteer handles POST /api/events/:id/volunteers.
func (h *Handler) AssignVolunteer(c *fiber.Ctx) error {
	orgID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	// Verify the event belongs to this organizer.
	event, err := h.Events.GetEvent(c.Context(), orgID, eventID)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not verify event ownership"},
		)
	}
	_ = event

	req := new(AssignVolunteerRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	record, err := h.Volunteer.AssignVolunteer(
		c.Context(), eventID, req.Email,
	)
	if err != nil {
		if errors.Is(err, service.ErrVolunteerUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "No account found with that email"},
			)
		}
		if errors.Is(err, service.ErrVolunteerAlreadyAssigned) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": "User is already a volunteer for this event"},
			)
		}
		if errors.Is(err, service.ErrInvalidInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not assign volunteer"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Volunteer assigned successfully",
		"volunteer": fiber.Map{"id": record.ID},
	})
}

// ListVolunteers handles GET /api/events/:id/volunteers.
func (h *Handler) ListVolunteers(c *fiber.Ctx) error {
	orgID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	// Verify ownership.
	if _, err := h.Events.GetEvent(c.Context(), orgID, eventID); err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not verify event ownership"},
		)
	}

	volunteers, err := h.Volunteer.ListVolunteers(c.Context(), eventID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not list volunteers"},
		)
	}

	result := make([]fiber.Map, 0, len(volunteers))
	for _, v := range volunteers {
		result = append(result, fiber.Map{
			"id":         v.ID,
			"user_id":    v.UserID,
			"user_name":  v.User.Name,
			"user_email": v.User.Email,
			"created_at": v.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{"volunteers": result})
}

// RemoveVolunteer handles DELETE /api/events/:id/volunteers.
type RemoveVolunteerRequest struct {
	Email string `json:"email"`
}

func (h *Handler) RemoveVolunteer(c *fiber.Ctx) error {
	orgID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	if _, err := h.Events.GetEvent(c.Context(), orgID, eventID); err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not verify event ownership"},
		)
	}

	req := new(RemoveVolunteerRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	if err := h.Volunteer.RemoveVolunteer(c.Context(), eventID, req.Email); err != nil {
		if errors.Is(err, service.ErrVolunteerUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "No account found with that email"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not remove volunteer"},
		)
	}

	return c.JSON(fiber.Map{"message": "Volunteer removed successfully"})
}

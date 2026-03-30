package api

import (
	"errors"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateEventRequest is the accepted payload for creating events.
type CreateEventRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	StartDate   string  `json:"start_date"`
	Venue       string  `json:"venue"`
	Capacity    int     `json:"capacity"`
	Price       float64 `json:"price"`
	IsPublic    bool    `json:"is_public"`
}

// CreateEvent creates a new organizer-owned event.
func (h *Handler) CreateEvent(c *fiber.Ctx) error {
	req := new(CreateEventRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	organizerID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	startDate, err := parseTime(req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid start_date format"},
		)
	}

	event, err := h.Events.CreateEvent(
		c.Context(),
		organizerID,
		service.CreateEventInput{
			Title:       req.Title,
			Description: req.Description,
			StartDate:   startDate,
			Venue:       req.Venue,
			Capacity:    req.Capacity,
			Price:       req.Price,
			IsPublic:    req.IsPublic,
		},
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEventInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not create event"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Event created successfully",
		"event": fiber.Map{
			"id":   event.ID,
			"slug": event.Slug,
		},
	})
}

// ListEvents returns organizing and/or attending event views for the user.
func (h *Handler) ListEvents(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	view := strings.ToLower(strings.TrimSpace(c.Query("view")))
	eventLists, err := h.Events.ListEventsByView(c.Context(), organizerID, view)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEventInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not fetch events"},
		)
	}

	return c.JSON(fiber.Map{
		"view":              eventLists.View,
		"events":            eventLists.Events,
		"organizing_events": eventLists.OrganizingEvents,
		"attending_events":  eventLists.AttendingEvents,
	})
}

// GetEvent returns a single organizer-owned event by ID.
func (h *Handler) GetEvent(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
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

	event, err := h.Events.GetEvent(
		c.Context(), organizerID, eventID,
	)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not fetch event"},
		)
	}

	return c.JSON(fiber.Map{"event": event})
}

// UpdateEventRequest allows partial updates on mutable fields.
type UpdateEventRequest struct {
	Capacity *int                `json:"capacity"`
	Price    *float64            `json:"price"`
	Status   *models.EventStatus `json:"status"`
}

// UpdateEvent partially updates an organizer-owned event.
func (h *Handler) UpdateEvent(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
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

	req := new(UpdateEventRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	event, err := h.Events.UpdateEvent(
		c.Context(),
		organizerID,
		eventID,
		service.UpdateEventInput{
			Capacity: req.Capacity,
			Price:    req.Price,
			Status:   req.Status,
		},
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEventInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not update event"},
		)
	}

	return c.JSON(fiber.Map{
		"message": "Event updated successfully",
		"event":   event,
	})
}

// DeleteEvent removes an organizer-owned event.
func (h *Handler) DeleteEvent(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
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

	err = h.Events.DeleteEvent(
		c.Context(), organizerID, eventID,
	)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not delete event"},
		)
	}

	return c.JSON(
		fiber.Map{"message": "Event deleted successfully"},
	)
}

func organizerIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok || claims.Subject == "" {
		return uuid.Nil, errors.New("missing auth claims")
	}

	organizerID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return organizerID, nil
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// GetPublicEvent returns a single event by slug for public display.
func (h *Handler) GetPublicEvent(c *fiber.Ctx) error {
	slug := c.Params("slug")

	event, err := h.Events.GetPublicEvent(
		c.Context(), slug,
	)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		if errors.Is(err, service.ErrInvalidEventInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not fetch event"},
		)
	}

	organizerName := ""
	if event.Organizer.Name != "" {
		organizerName = event.Organizer.Name
	}

	return c.JSON(fiber.Map{
		"event": fiber.Map{
			"id":               event.ID,
			"slug":             event.Slug,
			"title":            event.Title,
			"description":      event.Description,
			"poster_url":       event.PosterURL,
			"venue":            event.Venue,
			"start_date":       event.StartDate,
			"capacity":         event.Capacity,
			"total_registered": event.TotalRegistered,
			"price":            event.Price,
			"status":           event.Status,
			"organizer_name":   organizerName,
			"organizer_email":  event.Organizer.Email,
		},
	})
}

package api

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateEventRequest is the accepted payload for creating events.
type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	Venue       string    `json:"venue"`
	Capacity    int       `json:"capacity"`
	IsPublic    bool      `json:"is_public"`
}

// CreateEvent creates a new organizer-owned event.
func (h *Handler) CreateEvent(c *fiber.Ctx) error {
	req := new(CreateEventRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	claims, ok := middleware.ClaimsFromContext(c)
	if !ok || claims.Subject == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	organizerID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Invalid or expired session"},
		)
	}

	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	venue := strings.TrimSpace(req.Venue)
	startDate := req.StartDate.UTC()

	if title == "" || venue == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Title and venue are required"},
		)
	}

	if req.Capacity < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Capacity cannot be negative"},
		)
	}

	if startDate.IsZero() || !startDate.After(time.Now().UTC()) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Start date must be in the future"},
		)
	}

	var event models.Event
	var createErr error

	for i := 0; i < 3; i++ {
		slug, slugErr := generateEventSlug(title)
		if slugErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{"error": "Could not create event"},
			)
		}

		event = models.Event{
			Slug:        slug,
			OrganizerID: organizerID,
			Title:       title,
			Description: description,
			StartDate:   startDate,
			Venue:       venue,
			Capacity:    req.Capacity,
			IsPublic:    req.IsPublic,
		}

		result := h.DB.Create(&event)
		if result.Error == nil {
			createErr = nil
			break
		}

		createErr = result.Error
		if strings.Contains(createErr.Error(), "duplicate key value") && strings.Contains(createErr.Error(), "slug") {
			continue
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not create event"},
		)
	}

	if createErr != nil {
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

// ListEvents returns all events owned by the authenticated organizer.
func (h *Handler) ListEvents(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	var events []models.Event
	result := h.DB.Where("organizer_id = ?", organizerID).
		Order("created_at DESC").
		Find(&events)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not fetch events"},
		)
	}

	return c.JSON(fiber.Map{"events": events})
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

	var event models.Event
	result := h.DB.Where("id = ? AND organizer_id = ?", eventID, organizerID).First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
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

// UpdateEventRequest allows partial updates on mutable event fields.
type UpdateEventRequest struct {
	Capacity *int               `json:"capacity"`
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

	updates := map[string]interface{}{}

	if req.Capacity != nil {
		if *req.Capacity < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": "Capacity cannot be negative"},
			)
		}
		updates["capacity"] = *req.Capacity
	}

	if req.Status != nil {
		if !isValidEventStatus(*req.Status) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": "Invalid event status"},
			)
		}
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "No valid fields provided for update"},
		)
	}

	result := h.DB.Model(&models.Event{}).
		Where("id = ? AND organizer_id = ?", eventID, organizerID).
		Updates(updates)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not update event"},
		)
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(
			fiber.Map{"error": "Event not found"},
		)
	}

	var event models.Event
	fetchResult := h.DB.Where("id = ? AND organizer_id = ?", eventID, organizerID).First(&event)
	if fetchResult.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not fetch updated event"},
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

	result := h.DB.Where("id = ? AND organizer_id = ?", eventID, organizerID).Delete(&models.Event{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not delete event"},
		)
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(
			fiber.Map{"error": "Event not found"},
		)
	}

	return c.JSON(fiber.Map{"message": "Event deleted successfully"})
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

func isValidEventStatus(status models.EventStatus) bool {
	switch status {
	case models.EventStatusDraft, models.EventStatusActive, models.EventStatusFull, models.EventStatusCancelled:
		return true
	default:
		return false
	}
}

func generateEventSlug(title string) (string, error) {
	base := slugify(title)

	suffix, err := randomAlphaNumeric(6)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", base, strings.ToLower(suffix)), nil
}

func slugify(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))

	var b strings.Builder
	prevHyphen := false

	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}

		if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "event"
	}

	return out
}

func randomAlphaNumeric(n int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, n)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	for i := range buf {
		buf[i] = chars[int(buf[i])%len(chars)]
	}

	return string(buf), nil
}

package api

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const attendeeExportTimeout = 5 * time.Minute

// GetAnalytics returns the dashboard JSON for an event.
func (h *Handler) GetAnalytics(c *fiber.Ctx) error {
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

	data, err := h.Analytics.GetDashboardData(
		c.Context(), organizerID, eventID,
	)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "Could not fetch analytics",
			},
		)
	}

	return c.JSON(fiber.Map{"analytics": data})
}

// ExportAttendees streams a CSV of all attendees for
// the given event using SetBodyStreamWriter.
func (h *Handler) ExportAttendees(c *fiber.Ctx) error {
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

	// Verify ownership before streaming.
	_, checkErr := h.Analytics.GetDashboardData(
		c.Context(), organizerID, eventID,
	)
	if checkErr != nil {
		if errors.Is(
			checkErr, service.ErrEventNotFound,
		) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "Could not verify event",
			},
		)
	}

	filename := fmt.Sprintf(
		"attendees-%s.csv", eventID.String()[:8],
	)
	exportOrganizerID := organizerID
	exportEventID := eventID
	analyticsSvc := h.Analytics

	c.Set("Content-Type", "text/csv")
	c.Set(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=\"%s\"", filename,
		),
	)

	c.Context().SetBodyStreamWriter(
		func(w *bufio.Writer) {
			// Never pass Fiber's RequestCtx into asynchronous stream work.
			// RequestCtx is pooled and may be recycled before DB rows finish.
			streamCtx, cancel := context.WithTimeout(
				context.Background(),
				attendeeExportTimeout,
			)
			defer cancel()

			_ = analyticsSvc.ExportAttendeesCSV(
				streamCtx,
				exportOrganizerID,
				exportEventID,
				w,
			)
			_ = w.Flush()
		},
	)

	return nil
}

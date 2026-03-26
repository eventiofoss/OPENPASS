package api

import (
	"errors"
	"log/slog"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ScanCheckInRequest is the public scan payload DTO.
type ScanCheckInRequest struct {
	QRCode string `json:"qr_code"`
}

// ScanCheckIn handles POST /api/checkins/scan.
func (h *Handler) ScanCheckIn(c *fiber.Ctx) error {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"code":  "ERR_UNAUTHORIZED",
				"error": "Authentication required",
			},
		)
	}

	scannerID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"code":  "ERR_UNAUTHORIZED",
				"error": "Invalid session",
			},
		)
	}

	req := new(ScanCheckInRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"code":  "ERR_INVALID_REQUEST",
				"error": "Invalid JSON payload",
			},
		)
	}

	checkin, err := h.QR.ScanCheckIn(
		c.Context(), req.QRCode, scannerID,
	)
	if err != nil {
		return handleScanError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "CHECKIN_OK",
		"message": "Check-in successful",
		"checkin": fiber.Map{
			"id":         checkin.ID,
			"scanned_at": checkin.ScannedAt,
		},
	})
}

func handleScanError(
	c *fiber.Ctx, err error,
) error {
	if errors.Is(err, service.ErrInvalidQR) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"code":  "ERR_INVALID_QR",
				"error": err.Error(),
			},
		)
	}

	if errors.Is(err, service.ErrAttendeeNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(
			fiber.Map{
				"code":  "ERR_ATTENDEE_NOT_FOUND",
				"error": "Attendee not found for this event",
			},
		)
	}

	var alreadyErr *service.AlreadyCheckedInError
	if errors.As(err, &alreadyErr) {
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"code":             "ERR_ALREADY_SCANNED",
				"error":            "Attendee already checked in",
				"first_scanned_at": alreadyErr.FirstScannedAt,
			},
		)
	}

	if errors.Is(err, service.ErrVenueFull) {
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"code":  "ERR_VENUE_FULL",
				"error": "Venue has reached full capacity",
			},
		)
	}

	if errors.Is(err, service.ErrNotEventOrganizer) {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"code":  "ERR_NOT_EVENT_ORGANIZER",
				"error": "Not authorized to scan for this event",
			},
		)
	}

	slog.Error(
		"check-in scan failed",
		slog.String("error", err.Error()),
	)
	return c.Status(fiber.StatusInternalServerError).JSON(
		fiber.Map{
			"code":  "ERR_INTERNAL",
			"error": "Could not process check-in",
		},
	)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
)

const qrScheme = "eventpass://"

var (
	// ErrInvalidQR is returned for malformed QR strings.
	ErrInvalidQR = errors.New("invalid QR code")
	// ErrAttendeeNotFound is returned when the hash
	// does not match any attendee for the event.
	ErrAttendeeNotFound = errors.New(
		"attendee not found for event",
	)
	// ErrVenueFull is returned when venue capacity
	// has been reached.
	ErrVenueFull = errors.New("venue is at full capacity")
	// ErrNotEventOrganizer is returned when the scanner
	// is not the organizer of the event.
	ErrNotEventOrganizer = errors.New(
		"not authorized for this event",
	)
)

// AlreadyCheckedInError carries the first scan timestamp.
type AlreadyCheckedInError struct {
	FirstScannedAt time.Time
}

func (e *AlreadyCheckedInError) Error() string {
	return fmt.Sprintf(
		"attendee already checked in at %s",
		e.FirstScannedAt.Format(time.RFC3339),
	)
}

// CheckInRepo defines required storage operations.
type CheckInRepo interface {
	FindAttendeeByHashAndEvent(
		ctx context.Context,
		qrHash string,
		eventID uuid.UUID,
	) (*models.Attendee, error)

	ProcessCheckIn(
		ctx context.Context,
		eventID uuid.UUID,
		attendeeID uuid.UUID,
		scannedByID uuid.UUID,
	) (repository.CheckInOutcome, *models.CheckIn, error)
}

// QRService contains QR scan check-in business logic.
type QRService struct {
	repo CheckInRepo
}

// NewQRService returns a service backed by concrete repo.
func NewQRService(
	repo *repository.CheckInRepository,
) *QRService {
	return &QRService{repo: repo}
}

// NewQRServiceWithRepo returns a service using any
// CheckInRepo (for testing).
func NewQRServiceWithRepo(repo CheckInRepo) *QRService {
	return &QRService{repo: repo}
}

// ParseQR validates and extracts event_id and
// attendee_hash from a QR URI string.
//
// Expected format: eventpass://v1/{event_id}/{hash}
func ParseQR(
	raw string,
) (uuid.UUID, string, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, qrScheme) {
		return uuid.Nil, "", fmt.Errorf(
			"%w: missing eventpass:// scheme", ErrInvalidQR,
		)
	}

	body := strings.TrimPrefix(raw, qrScheme)
	parts := strings.Split(body, "/")
	if len(parts) != 3 {
		return uuid.Nil, "", fmt.Errorf(
			"%w: expected 3 path segments", ErrInvalidQR,
		)
	}

	version := parts[0]
	if version != "v1" {
		return uuid.Nil, "", fmt.Errorf(
			"%w: unsupported version %q", ErrInvalidQR,
			version,
		)
	}

	eventID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, "", fmt.Errorf(
			"%w: invalid event_id", ErrInvalidQR,
		)
	}

	hash := parts[2]
	if hash == "" {
		return uuid.Nil, "", fmt.Errorf(
			"%w: empty attendee hash", ErrInvalidQR,
		)
	}

	return eventID, hash, nil
}

// ScanCheckIn parses a QR code, looks up the attendee,
// and performs the transactional check-in.
func (s *QRService) ScanCheckIn(
	ctx context.Context,
	rawQR string,
	scannedByID uuid.UUID,
) (*models.CheckIn, error) {
	eventID, hash, err := ParseQR(rawQR)
	if err != nil {
		return nil, err
	}

	attendee, err := s.repo.FindAttendeeByHashAndEvent(
		ctx, hash, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"finding attendee: %w", err,
		)
	}
	if attendee == nil {
		return nil, ErrAttendeeNotFound
	}

	outcome, checkin, err := s.repo.ProcessCheckIn(
		ctx, eventID, attendee.ID, scannedByID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"processing check-in: %w", err,
		)
	}

	switch outcome {
	case repository.CheckInOutcomeSuccess:
		return checkin, nil
	case repository.CheckInOutcomeAlreadyCheckedIn:
		scannedAt := time.Time{}
		if checkin != nil {
			scannedAt = checkin.ScannedAt
		}
		return nil, &AlreadyCheckedInError{
			FirstScannedAt: scannedAt,
		}
	case repository.CheckInOutcomeVenueFull:
		return nil, ErrVenueFull
	case repository.CheckInOutcomeNotEventOrganizer:
		return nil, ErrNotEventOrganizer
	default:
		return nil, errors.New("check-in failed")
	}
}

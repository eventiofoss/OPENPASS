package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CheckInOutcome represents the result of a check-in attempt.
type CheckInOutcome int

const (
	CheckInOutcomeUnknown CheckInOutcome = iota
	CheckInOutcomeSuccess
	CheckInOutcomeVenueFull
	CheckInOutcomeAlreadyCheckedIn
	CheckInOutcomeNotEventOrganizer
)

// CheckInRepository handles check-in database queries.
type CheckInRepository struct {
	db *gorm.DB
}

// NewCheckInRepository returns a ready repository.
func NewCheckInRepository(
	db *gorm.DB,
) *CheckInRepository {
	return &CheckInRepository{db: db}
}

// FindAttendeeByHashAndEvent looks up an attendee
// by QR hash scoped to a specific event.
func (r *CheckInRepository) FindAttendeeByHashAndEvent(
	ctx context.Context,
	qrHash string,
	eventID uuid.UUID,
) (*models.Attendee, error) {
	var attendee models.Attendee

	err := r.db.WithContext(ctx).
		Where("qr_hash = ? AND event_id = ?", qrHash, eventID).
		First(&attendee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &attendee, nil
}

// ProcessCheckIn performs an insert-first transactional
// check-in with minimal lock duration.
//
// Transaction flow:
//  1. INSERT check-in (catches duplicates via unique index).
//  2. SELECT … FOR UPDATE on event (capacity + ownership).
//  3. Increment check_in_count.
func (r *CheckInRepository) ProcessCheckIn(
	ctx context.Context,
	eventID uuid.UUID,
	attendeeID uuid.UUID,
	scannedByID uuid.UUID,
) (CheckInOutcome, *models.CheckIn, error) {
	outcome := CheckInOutcomeUnknown
	var result *models.CheckIn

	err := r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// Step 1: Attempt INSERT first (fast duplicate check).
			checkin := models.CheckIn{
				AttendeeID:  attendeeID,
				ScannedByID: scannedByID,
				ScannedAt:   time.Now().UTC(),
			}

			if err := tx.Create(&checkin).Error; err != nil {
				if isDuplicateCheckInErr(err) {
					outcome = CheckInOutcomeAlreadyCheckedIn
					// Fetch existing record for scanned_at.
					var existing models.CheckIn
					if fetchErr := r.db.
						WithContext(ctx).
						Where("attendee_id = ?", attendeeID).
						First(&existing).Error; fetchErr == nil {
						result = &existing
					}

					return nil
				}

				return err
			}

			// Step 2: Lock event row (shortest possible window).
			var event models.Event
			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Select(
					"id",
					"organizer_id",
					"capacity",
					"check_in_count",
				).
				Where("id = ?", eventID).
				First(&event).Error; err != nil {
				return err
			}

			// Step 3: Verify organizer ownership.
			if event.OrganizerID != scannedByID {
				outcome = CheckInOutcomeNotEventOrganizer
				return nil
			}

			// Step 4: Check venue capacity.
			if event.CheckInCount >= event.Capacity {
				outcome = CheckInOutcomeVenueFull
				return nil
			}

			// Step 5: Increment check-in counter.
			if err := tx.
				Model(&models.Event{}).
				Where("id = ?", eventID).
				UpdateColumn(
					"check_in_count",
					gorm.Expr("check_in_count + 1"),
				).Error; err != nil {
				return err
			}

			outcome = CheckInOutcomeSuccess
			result = &checkin
			return nil
		},
	)
	if err != nil {
		return CheckInOutcomeUnknown, nil, err
	}

	return outcome, result, nil
}

func isDuplicateCheckInErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") &&
		strings.Contains(msg, "attendee_id")
}

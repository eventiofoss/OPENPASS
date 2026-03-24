package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistrationOutcome represents transactional registration result.
type RegistrationOutcome int

const (
	RegistrationOutcomeUnknown RegistrationOutcome = iota
	RegistrationOutcomeCreated
	RegistrationOutcomeEventNotFound
	RegistrationOutcomeEventFull
	RegistrationOutcomeDuplicateAttendee
)

// AttendeeRepository handles attendee database queries.
type AttendeeRepository struct {
	db *gorm.DB
}

// NewAttendeeRepository returns a ready repository.
func NewAttendeeRepository(db *gorm.DB) *AttendeeRepository {
	return &AttendeeRepository{db: db}
}

// CreateForEvent registers an attendee and increments total_registered atomically.
func (r *AttendeeRepository) CreateForEvent(
	ctx context.Context,
	eventID uuid.UUID,
	attendee *models.Attendee,
) (RegistrationOutcome, error) {
	outcome := RegistrationOutcomeUnknown

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var event models.Event
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "capacity", "total_registered").
			Where("id = ?", eventID).
			First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				outcome = RegistrationOutcomeEventNotFound
				return nil
			}
			return err
		}

		if event.TotalRegistered >= event.Capacity {
			outcome = RegistrationOutcomeEventFull
			return nil
		}

		attendee.EventID = eventID
		if err := tx.Create(attendee).Error; err != nil {
			if isDuplicateAttendeeErr(err) {
				outcome = RegistrationOutcomeDuplicateAttendee
				return nil
			}
			return err
		}

		if err := tx.
			Model(&models.Event{}).
			Where("id = ?", eventID).
			UpdateColumn("total_registered", gorm.Expr("total_registered + 1")).
			Error; err != nil {
			return err
		}

		outcome = RegistrationOutcomeCreated
		return nil
	})
	if err != nil {
		return RegistrationOutcomeUnknown, err
	}

	return outcome, nil
}

func isDuplicateAttendeeErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") &&
		(strings.Contains(msg, "idx_attendees_event_email") ||
			(strings.Contains(msg, "event_id") && strings.Contains(msg, "email")))
}

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

// RegistrationOutcome represents transactional registration result.
type RegistrationOutcome int

const (
	RegistrationOutcomeUnknown RegistrationOutcome = iota
	RegistrationOutcomeCreated
	RegistrationOutcomeEventNotFound
	RegistrationOutcomeEventFull
	RegistrationOutcomePaymentRequired
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
			Select("id", "capacity", "price", "total_registered", "tickets_sold").
			Where("id = ? AND status = ?", eventID, models.EventStatusActive).
			First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				outcome = RegistrationOutcomeEventNotFound
				return nil
			}
			return err
		}

		if event.Price > 0 {
			outcome = RegistrationOutcomePaymentRequired
			return nil
		}

		occupied := occupiedInventory(
			event.TotalRegistered,
			event.TicketsSold,
		)
		if occupied >= event.Capacity {
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

		nextOccupied := occupied + 1
		if err := tx.
			Model(&models.Event{}).
			Where("id = ?", eventID).
			Updates(map[string]interface{}{
				"total_registered": nextOccupied,
				"tickets_sold":     nextOccupied,
				"updated_at":       time.Now().UTC(),
			}).
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

func occupiedInventory(totalRegistered, ticketsSold int) int {
	if ticketsSold > totalRegistered {
		return ticketsSold
	}

	return totalRegistered
}

func isDuplicateAttendeeErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") &&
		(strings.Contains(msg, "idx_attendees_event_email") ||
			(strings.Contains(msg, "event_id") && strings.Contains(msg, "email")))
}

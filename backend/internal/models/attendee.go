package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendeeStatus string

const (
	AttendeeStatusPending    AttendeeStatus = "pending"
	AttendeeStatusRegistered AttendeeStatus = "registered"
	AttendeeStatusPaid       AttendeeStatus = "paid"
	AttendeeStatusCheckedIn  AttendeeStatus = "checked_in"
	AttendeeStatusCancelled  AttendeeStatus = "cancelled"
)

// Attendee stores passwordless event registrations and pass metadata.
type Attendee struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	EventID     uuid.UUID       `gorm:"type:uuid;not null;index;uniqueIndex:idx_attendees_event_email" json:"event_id"`
	Email       string          `gorm:"not null;index;uniqueIndex:idx_attendees_event_email" json:"email"`
	Name        string          `gorm:"not null" json:"name"`
	FormData    json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"form_data"`
	QRHash      string          `gorm:"column:qr_hash;size:64;not null;uniqueIndex" json:"-"`
	Status      AttendeeStatus  `gorm:"type:varchar(20);not null;default:pending;check:attendees_status,status IN ('pending','registered','paid','checked_in','cancelled')" json:"status"`
	CheckedInAt *time.Time      `json:"checked_in_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Event   Event    `gorm:"foreignKey:EventID" json:"-"`
	Payment *Payment `gorm:"foreignKey:AttendeeID" json:"-"`
	CheckIn *CheckIn `gorm:"foreignKey:AttendeeID" json:"-"`
}

func (a *Attendee) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&a.ID)

	if a.QRHash == "" {
		token, err := generateSecureToken(32)
		if err != nil {
			return err
		}
		a.QRHash = token
	}

	return nil
}

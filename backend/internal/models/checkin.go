package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckIn stores the successful venue scan for an attendee.
type CheckIn struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AttendeeID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"attendee_id"`
	ScannedByID uuid.UUID `gorm:"column:scanned_by;type:uuid;not null;index" json:"scanned_by"`
	ScannedAt   time.Time `gorm:"not null" json:"scanned_at"`
	CreatedAt   time.Time `json:"created_at"`

	Attendee  Attendee `gorm:"foreignKey:AttendeeID" json:"-"`
	ScannedBy User     `gorm:"foreignKey:ScannedByID" json:"-"`
}

func (CheckIn) TableName() string {
	return "checkins"
}

func (c *CheckIn) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&c.ID)

	if c.ScannedAt.IsZero() {
		c.ScannedAt = time.Now().UTC()
	}

	return nil
}

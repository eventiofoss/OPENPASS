package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventVolunteer maps a user to an event as a volunteer (event-scoped role).
// Organizers assign volunteers via their dashboard; volunteers cannot self-assign.
type EventVolunteer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_volunteer" json:"event_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_volunteer" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

func (ev *EventVolunteer) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&ev.ID)
	return nil
}

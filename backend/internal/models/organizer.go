package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizerRole string

const (
	OrganizerRoleOrganizer   OrganizerRole = "organizer"
	OrganizerRoleAdmin       OrganizerRole = "admin"
	OrganizerRoleVolunteer   OrganizerRole = "volunteer"
	OrganizerRoleParticipant OrganizerRole = "participant"
)

// Organizer is the only authenticated account type in the system.
type Organizer struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string        `gorm:"not null" json:"name"`
	Email        string        `gorm:"not null;uniqueIndex" json:"email"`
	PasswordHash string        `gorm:"not null" json:"-"`
	Role         OrganizerRole `gorm:"type:varchar(20);not null;default:organizer;check:organizers_role,role IN ('organizer','admin','volunteer','participant')" json:"role"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	Events   []Event   `gorm:"foreignKey:OrganizerID" json:"-"`
	CheckIns []CheckIn `gorm:"foreignKey:ScannedByID" json:"-"`
	Exports  []Export  `gorm:"foreignKey:OrganizerID" json:"-"`
}

func (o *Organizer) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&o.ID)
	return nil
}

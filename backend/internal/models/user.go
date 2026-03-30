package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleOrganizer   UserRole = "organizer"
	UserRoleAdmin       UserRole = "admin"
	UserRoleVolunteer   UserRole = "volunteer"
	UserRoleParticipant UserRole = "participant"
)

// OrganizerRole is preserved as a compatibility alias while API layers migrate.
type OrganizerRole = UserRole

const (
	OrganizerRoleOrganizer   OrganizerRole = UserRoleOrganizer
	OrganizerRoleAdmin       OrganizerRole = UserRoleAdmin
	OrganizerRoleVolunteer   OrganizerRole = UserRoleVolunteer
	OrganizerRoleParticipant OrganizerRole = UserRoleParticipant
)

// User is the authenticated account type in the system.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         UserRole  `gorm:"type:varchar(20);not null;default:organizer;check:users_role,role IN ('organizer','admin','volunteer','participant')" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Events   []Event   `gorm:"foreignKey:OrganizerID" json:"-"`
	CheckIns []CheckIn `gorm:"foreignKey:ScannedByID" json:"-"`
	Exports  []Export  `gorm:"foreignKey:OrganizerID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&u.ID)
	return nil
}

// Organizer is preserved as a compatibility alias while services migrate.
type Organizer = User

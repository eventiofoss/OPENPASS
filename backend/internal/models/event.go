package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusActive    EventStatus = "active"
	EventStatusFull      EventStatus = "full"
	EventStatusCancelled EventStatus = "cancelled"
)

// Event stores the core event metadata and visibility controls.
type Event struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Slug            string      `gorm:"not null;uniqueIndex" json:"slug"`
	AccessToken     string      `gorm:"column:access_token;size:64;not null;uniqueIndex" json:"-"`
	OrganizerID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"organizer_id"`
	Title           string      `gorm:"not null" json:"title"`
	Description     string      `gorm:"type:text" json:"description"`
	StartDate       time.Time   `gorm:"not null" json:"start_date"`
	Venue           string      `gorm:"not null" json:"venue"`
	Capacity        int         `gorm:"not null;check:events_capacity_non_negative,capacity >= 0" json:"capacity"`
	Price           float64     `gorm:"type:numeric(10,2);not null;default:0;check:events_price_non_negative,price >= 0" json:"price"`
	IsPublic        bool        `gorm:"not null;default:false" json:"is_public"`
	Status          EventStatus `gorm:"type:varchar(20);not null;default:draft;check:events_status,status IN ('draft','active','full','cancelled')" json:"status"`
	TotalRegistered int         `gorm:"not null;default:0;check:events_total_registered_non_negative,total_registered >= 0" json:"total_registered"`
	TicketsSold     int         `gorm:"not null;default:0" json:"tickets_sold"`
	CheckInCount    int         `gorm:"not null;default:0" json:"check_in_count"`
	TotalRevenue    float64     `gorm:"type:numeric(10,2);not null;default:0" json:"total_revenue"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	Organizer  Organizer   `gorm:"foreignKey:OrganizerID" json:"-"`
	FormFields []FormField `gorm:"foreignKey:EventID" json:"-"`
	Attendees  []Attendee  `gorm:"foreignKey:EventID" json:"-"`
	Exports    []Export    `gorm:"foreignKey:EventID" json:"-"`
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&e.ID)

	if e.AccessToken == "" {
		token, err := generateSecureToken(32)
		if err != nil {
			return err
		}
		e.AccessToken = token
	}

	return nil
}

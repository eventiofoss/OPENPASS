package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExportFormat string

const (
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatExcel ExportFormat = "excel"
)

// Export captures an audit log of attendee data exports.
type Export struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	EventID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"event_id"`
	OrganizerID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organizer_id"`
	Format      ExportFormat `gorm:"type:varchar(10);not null;check:exports_format,format IN ('csv','excel')" json:"format"`
	Filename    string       `gorm:"not null" json:"filename"`
	GeneratedAt time.Time    `gorm:"not null;autoCreateTime" json:"generated_at"`

	Event     Event `gorm:"foreignKey:EventID" json:"-"`
	Organizer User  `gorm:"foreignKey:OrganizerID" json:"-"`
}

func (e *Export) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&e.ID)
	return nil
}

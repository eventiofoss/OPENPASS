package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormFieldType string

const (
	FormFieldTypeText     FormFieldType = "text"
	FormFieldTypeEmail    FormFieldType = "email"
	FormFieldTypeSelect   FormFieldType = "select"
	FormFieldTypeCheckbox FormFieldType = "checkbox"
	FormFieldTypeNumber   FormFieldType = "number"
)

// FormField defines a custom registration field for a specific event.
type FormField struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	EventID   uuid.UUID       `gorm:"type:uuid;not null;index;uniqueIndex:idx_form_fields_event_name;uniqueIndex:idx_form_fields_event_position" json:"event_id"`
	Name      string          `gorm:"not null;uniqueIndex:idx_form_fields_event_name" json:"name"`
	Type      FormFieldType   `gorm:"type:varchar(20);not null;check:form_fields_type,type IN ('text','email','select','checkbox','number')" json:"type"`
	Label     string          `gorm:"not null" json:"label"`
	Required  bool            `gorm:"not null;default:false" json:"required"`
	Options   json.RawMessage `gorm:"type:jsonb" json:"options,omitempty"`
	Position  int             `gorm:"not null;default:0;uniqueIndex:idx_form_fields_event_position;check:form_fields_position_non_negative,position >= 0" json:"position"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (f *FormField) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&f.ID)
	return nil
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
)

// Payment stores the current payment state for a registration.
type Payment struct {
	ID                   uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	AttendeeID           uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex" json:"attendee_id"`
	HyperswitchPaymentID string        `gorm:"not null;uniqueIndex" json:"hyperswitch_payment_id"`
	Amount               float64       `gorm:"type:numeric(10,2);not null;check:payments_amount_non_negative,amount >= 0" json:"amount"`
	Currency             string        `gorm:"type:char(3);not null" json:"currency"`
	Status               PaymentStatus `gorm:"type:varchar(20);not null;default:pending;check:payments_status,status IN ('pending','succeeded','failed')" json:"status"`
	PaidAt               *time.Time    `json:"paid_at,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`

	Attendee Attendee `gorm:"foreignKey:AttendeeID" json:"-"`
}

func (p *Payment) BeforeCreate(_ *gorm.DB) error {
	ensureUUID(&p.ID)
	return nil
}

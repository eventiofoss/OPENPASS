package models

import (
	"time"
	"github.com/google/uuid"
)

// User represents a system user with role-based access control.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"` // SECURITY: The "-" guarantees this never leaks into JSON
	Role         string    `gorm:"not null;default:'user'" json:"role"` // user, admin, etc.
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
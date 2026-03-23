package repository

import (
	"context"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FormFieldRepository handles form field database queries.
type FormFieldRepository struct {
	db *gorm.DB
}

// NewFormFieldRepository returns a ready repository.
func NewFormFieldRepository(
	db *gorm.DB,
) *FormFieldRepository {
	return &FormFieldRepository{db: db}
}

// CreateBatch inserts multiple form fields in one transaction.
func (r *FormFieldRepository) CreateBatch(
	ctx context.Context,
	fields []models.FormField,
) error {
	if len(fields) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&fields).Error
}

// FindByEventID returns all form fields for an event
// ordered by position.
func (r *FormFieldRepository) FindByEventID(
	ctx context.Context,
	eventID uuid.UUID,
) ([]models.FormField, error) {
	var fields []models.FormField

	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("position ASC").
		Find(&fields).Error
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// DeleteByEventID removes all form fields for an event.
func (r *FormFieldRepository) DeleteByEventID(
	ctx context.Context,
	eventID uuid.UUID,
) error {
	return r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Delete(&models.FormField{}).Error
}

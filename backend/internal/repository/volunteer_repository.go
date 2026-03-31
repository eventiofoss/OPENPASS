package repository

import (
	"context"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VolunteerRepository handles event_volunteers database queries.
type VolunteerRepository struct {
	db *gorm.DB
}

// NewVolunteerRepository returns a ready repository.
func NewVolunteerRepository(db *gorm.DB) *VolunteerRepository {
	return &VolunteerRepository{db: db}
}

// Assign creates an event-volunteer mapping.
func (r *VolunteerRepository) Assign(
	ctx context.Context,
	eventID, userID uuid.UUID,
) (*models.EventVolunteer, error) {
	record := &models.EventVolunteer{
		EventID: eventID,
		UserID:  userID,
	}

	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

// Remove deletes an event-volunteer mapping.
func (r *VolunteerRepository) Remove(
	ctx context.Context,
	eventID, userID uuid.UUID,
) error {
	return r.db.WithContext(ctx).
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Delete(&models.EventVolunteer{}).Error
}

// FindByEvent returns all volunteers assigned to an event.
func (r *VolunteerRepository) FindByEvent(
	ctx context.Context,
	eventID uuid.UUID,
) ([]models.EventVolunteer, error) {
	var volunteers []models.EventVolunteer

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("event_id = ?", eventID).
		Order("created_at ASC").
		Find(&volunteers).Error
	if err != nil {
		return nil, err
	}

	return volunteers, nil
}

// IsVolunteer checks whether a user is assigned as a volunteer for an event.
func (r *VolunteerRepository) IsVolunteer(
	ctx context.Context,
	eventID, userID uuid.UUID,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.EventVolunteer{}).
		Where("event_id = ? AND user_id = ?", eventID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsDuplicateVolunteerErr checks for unique constraint violation.
func IsDuplicateVolunteerErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") &&
		strings.Contains(msg, "idx_event_volunteer")
}

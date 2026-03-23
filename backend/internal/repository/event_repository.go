package repository

import (
	"context"
	"errors"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventRepository handles event database queries.
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository returns a ready repository.
func NewEventRepository(
	db *gorm.DB,
) *EventRepository {
	return &EventRepository{db: db}
}

// Create inserts a new event record.
func (r *EventRepository) Create(
	ctx context.Context,
	event *models.Event,
) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// FindByIDAndOrganizer returns one event scoped to owner.
func (r *EventRepository) FindByIDAndOrganizer(
	ctx context.Context,
	eventID uuid.UUID,
	organizerID uuid.UUID,
) (*models.Event, error) {
	var event models.Event

	err := r.db.WithContext(ctx).
		Where("id = ? AND organizer_id = ?", eventID, organizerID).
		First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &event, nil
}

// FindAllByOrganizer returns all events owned by the organizer.
func (r *EventRepository) FindAllByOrganizer(
	ctx context.Context,
	organizerID uuid.UUID,
) ([]models.Event, error) {
	var events []models.Event

	err := r.db.WithContext(ctx).
		Where("organizer_id = ?", organizerID).
		Order("created_at DESC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

// Update applies partial updates scoped to the owner.
func (r *EventRepository) Update(
	ctx context.Context,
	eventID uuid.UUID,
	organizerID uuid.UUID,
	updates map[string]interface{},
) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("id = ? AND organizer_id = ?", eventID, organizerID).
		Updates(updates)

	return result.RowsAffected, result.Error
}

// Delete removes an event scoped to the owner.
func (r *EventRepository) Delete(
	ctx context.Context,
	eventID uuid.UUID,
	organizerID uuid.UUID,
) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND organizer_id = ?", eventID, organizerID).
		Delete(&models.Event{})

	return result.RowsAffected, result.Error
}

package repository

import (
	"context"
	"errors"
	"strings"

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

// FindAllByAttendeeUser returns all events where the user is a linked attendee.
func (r *EventRepository) FindAllByAttendeeUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Event, error) {
	var events []models.Event

	err := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Distinct("events.*").
		Joins("JOIN attendees ON attendees.event_id = events.id").
		Where("attendees.user_id = ?", userID).
		Order("events.created_at DESC").
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

// FindBySlug returns one event by its public slug.
func (r *EventRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (*models.Event, error) {
	var event models.Event

	err := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Select("events.*").
		Joins("JOIN users ON users.id = events.organizer_id").
		Preload("Organizer").
		Where("events.slug = ?", slug).
		First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &event, nil
}

// FindPublicPublished returns public events currently visible in the directory.
func (r *EventRepository) FindPublicPublished(
	ctx context.Context,
	search string,
) ([]models.Event, error) {
	var events []models.Event

	query := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Preload("Organizer").
		Where("is_public = ?", true).
		Where("status IN ?", []models.EventStatus{
			models.EventStatusActive,
			models.EventStatusFull,
		}).
		Order("start_date ASC")

	search = strings.TrimSpace(search)
	if search != "" {
		term := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(venue) LIKE ? OR LOWER(description) LIKE ?",
			term,
			term,
			term,
		)
	}

	err := query.Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
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

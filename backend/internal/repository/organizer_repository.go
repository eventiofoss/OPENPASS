package repository

import (
	"context"
	"errors"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"gorm.io/gorm"
)

// OrganizerRepository handles organizer database queries.
type OrganizerRepository struct {
	db *gorm.DB
}

// NewOrganizerRepository returns a ready repository.
func NewOrganizerRepository(
	db *gorm.DB,
) *OrganizerRepository {
	return &OrganizerRepository{db: db}
}

// Create inserts a new organizer record.
func (r *OrganizerRepository) Create(
	ctx context.Context,
	organizer *models.Organizer,
) error {
	return r.db.WithContext(ctx).Create(organizer).Error
}

// FindByEmail returns the organizer matching the email.
func (r *OrganizerRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*models.Organizer, error) {
	var organizer models.Organizer

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&organizer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &organizer, nil
}

// FindByID returns the organizer matching the UUID string.
func (r *OrganizerRepository) FindByID(
	ctx context.Context,
	id string,
) (*models.Organizer, error) {
	var organizer models.Organizer

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&organizer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &organizer, nil
}

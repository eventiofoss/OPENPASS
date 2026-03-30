package repository

import (
	"context"
	"errors"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles authenticated user database queries.
type UserRepository struct {
	db *gorm.DB
}

// OrganizerRepository is kept as a compatibility alias while callers migrate.
type OrganizerRepository = UserRepository

// NewUserRepository returns a ready repository.
func NewUserRepository(
	db *gorm.DB,
) *UserRepository {
	return &UserRepository{db: db}
}

// NewOrganizerRepository is kept as a compatibility constructor alias.
func NewOrganizerRepository(
	db *gorm.DB,
) *UserRepository {
	return NewUserRepository(db)
}

// Create inserts a new user record.
func (r *UserRepository) Create(
	ctx context.Context,
	user *models.User,
) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// CreateAndClaim inserts a new user and links prior guest attendees by email.
func (r *UserRepository) CreateAndClaim(
	ctx context.Context,
	user *models.User,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		_, err := claimUnlinkedAttendeesByEmail(
			tx,
			user.ID,
			user.Email,
		)
		return err
	})
}

// ClaimUnlinkedAttendeesByEmail links guest attendee rows to a user account.
func (r *UserRepository) ClaimUnlinkedAttendeesByEmail(
	ctx context.Context,
	userID uuid.UUID,
	email string,
) (int64, error) {
	return claimUnlinkedAttendeesByEmail(
		r.db.WithContext(ctx),
		userID,
		email,
	)
}

// FindByEmail returns the user matching the email.
func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	var user models.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

// FindByID returns the user matching the UUID string.
func (r *UserRepository) FindByID(
	ctx context.Context,
	id string,
) (*models.User, error) {
	var user models.User

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func claimUnlinkedAttendeesByEmail(
	db *gorm.DB,
	userID uuid.UUID,
	email string,
) (int64, error) {
	result := db.Model(&models.Attendee{}).
		Where("email = ? AND user_id IS NULL", email).
		Updates(map[string]interface{}{
			"user_id":    userID,
			"updated_at": time.Now().UTC(),
		})

	return result.RowsAffected, result.Error
}

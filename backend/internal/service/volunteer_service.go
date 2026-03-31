package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	// ErrVolunteerUserNotFound means the email doesn't match any account.
	ErrVolunteerUserNotFound = errors.New("user not found")
	// ErrVolunteerAlreadyAssigned means the user is already a volunteer for this event.
	ErrVolunteerAlreadyAssigned = errors.New("user is already a volunteer for this event")
	// ErrVolunteerNotAssigned means the user is not a volunteer for this event.
	ErrVolunteerNotAssigned = errors.New("user is not a volunteer for this event")
)

// VolunteerRepo defines required storage operations.
type VolunteerRepo interface {
	Assign(ctx context.Context, eventID, userID uuid.UUID) (*models.EventVolunteer, error)
	Remove(ctx context.Context, eventID, userID uuid.UUID) error
	FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventVolunteer, error)
}

// VolunteerService contains volunteer assignment business logic.
type VolunteerService struct {
	repo     *repository.VolunteerRepository
	userRepo UserRepo
}

// NewVolunteerService returns a service backed by repositories.
func NewVolunteerService(
	repo *repository.VolunteerRepository,
	userRepo *repository.UserRepository,
) *VolunteerService {
	return &VolunteerService{repo: repo, userRepo: userRepo}
}

// AssignVolunteer looks up a user by email and assigns them as a volunteer
// for the given event.
func (s *VolunteerService) AssignVolunteer(
	ctx context.Context,
	eventID uuid.UUID,
	email string,
) (*models.EventVolunteer, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, ErrVolunteerUserNotFound
	}

	record, err := s.repo.Assign(ctx, eventID, user.ID)
	if err != nil {
		if repository.IsDuplicateVolunteerErr(err) {
			return nil, ErrVolunteerAlreadyAssigned
		}
		return nil, fmt.Errorf("assigning volunteer: %w", err)
	}

	return record, nil
}

// RemoveVolunteer removes a volunteer assignment by email.
func (s *VolunteerService) RemoveVolunteer(
	ctx context.Context,
	eventID uuid.UUID,
	email string,
) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return ErrVolunteerUserNotFound
	}

	return s.repo.Remove(ctx, eventID, user.ID)
}

// ListVolunteers returns all volunteers for an event.
func (s *VolunteerService) ListVolunteers(
	ctx context.Context,
	eventID uuid.UUID,
) ([]models.EventVolunteer, error) {
	return s.repo.FindByEvent(ctx, eventID)
}

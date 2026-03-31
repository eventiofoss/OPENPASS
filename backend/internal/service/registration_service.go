package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	// ErrInvalidRegistrationInput is returned for malformed registration data.
	ErrInvalidRegistrationInput = errors.New("invalid registration input")
	// ErrRegistrationEventNotFound is returned when event does not exist.
	ErrRegistrationEventNotFound = errors.New("event not found")
	// ErrRegistrationEventFull is returned when event capacity is reached.
	ErrRegistrationEventFull = errors.New("event is full")
	// ErrRegistrationRequiresCheckout is returned when a paid event is
	// posted to the free registration flow.
	ErrRegistrationRequiresCheckout = errors.New("event requires checkout")
	// ErrAlreadyRegistered is returned for duplicate attendee email per event.
	ErrAlreadyRegistered = errors.New("attendee already registered for this event")
)

// RegistrationRepo defines required storage operations.
type RegistrationRepo interface {
	CreateForEvent(
		ctx context.Context,
		eventID uuid.UUID,
		attendee *models.Attendee,
	) (repository.RegistrationOutcome, error)
}

// RegistrationUserRepo defines user lookups used for attendee linking.
type RegistrationUserRepo interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

// RegisterAttendeeInput carries registration payload.
type RegisterAttendeeInput struct {
	EventID             uuid.UUID
	Name                string
	Email               string
	FormData            json.RawMessage
	AuthenticatedUserID *uuid.UUID
}

// RegistrationService contains registration business logic.
type RegistrationService struct {
	repo     RegistrationRepo
	userRepo RegistrationUserRepo
}

// NewRegistrationService returns a service backed by concrete repository.
func NewRegistrationService(
	repo *repository.AttendeeRepository,
	userRepo ...RegistrationUserRepo,
) *RegistrationService {
	service := &RegistrationService{repo: repo}
	if len(userRepo) > 0 {
		service.userRepo = userRepo[0]
	}

	return service
}

// NewRegistrationServiceWithRepo returns a service using any RegistrationRepo.
func NewRegistrationServiceWithRepo(
	repo RegistrationRepo,
	userRepo ...RegistrationUserRepo,
) *RegistrationService {
	service := &RegistrationService{repo: repo}
	if len(userRepo) > 0 {
		service.userRepo = userRepo[0]
	}

	return service
}

// RegisterAttendee creates attendee and updates event counter atomically.
func (s *RegistrationService) RegisterAttendee(
	ctx context.Context,
	input RegisterAttendeeInput,
) (*models.Attendee, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if input.EventID == uuid.Nil || name == "" || email == "" {
		return nil, fmt.Errorf("%w: name, email, and event_id are required", ErrInvalidRegistrationInput)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("%w: email format is invalid", ErrInvalidRegistrationInput)
	}

	formData := input.FormData
	if len(formData) == 0 {
		formData = json.RawMessage(`{}`)
	}
	if !json.Valid(formData) {
		return nil, fmt.Errorf("%w: form_data must be valid JSON", ErrInvalidRegistrationInput)
	}

	attendee := &models.Attendee{
		Name:           name,
		Email:          email,
		FormData:       formData,
		QRHash:         uuid.NewString(),
		Status:         models.AttendeeStatusPending,
	}

	if input.AuthenticatedUserID != nil {
		attendee.UserID = input.AuthenticatedUserID
	}

	outcome, err := s.repo.CreateForEvent(ctx, input.EventID, attendee)
	if err != nil {
		return nil, fmt.Errorf("registering attendee: %w", err)
	}

	switch outcome {
	case repository.RegistrationOutcomeCreated:
		return attendee, nil
	case repository.RegistrationOutcomeEventNotFound:
		return nil, ErrRegistrationEventNotFound
	case repository.RegistrationOutcomeEventFull:
		return nil, ErrRegistrationEventFull
	case repository.RegistrationOutcomePaymentRequired:
		return nil, ErrRegistrationRequiresCheckout
	case repository.RegistrationOutcomeDuplicateAttendee:
		return nil, ErrAlreadyRegistered
	default:
		return nil, errors.New("registration failed")
	}
}

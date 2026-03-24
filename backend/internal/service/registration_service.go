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

// RegisterAttendeeInput carries registration payload.
type RegisterAttendeeInput struct {
	EventID  uuid.UUID
	Name     string
	Email    string
	FormData json.RawMessage
}

// RegistrationService contains registration business logic.
type RegistrationService struct {
	repo RegistrationRepo
}

// NewRegistrationService returns a service backed by concrete repository.
func NewRegistrationService(repo *repository.AttendeeRepository) *RegistrationService {
	return &RegistrationService{repo: repo}
}

// NewRegistrationServiceWithRepo returns a service using any RegistrationRepo.
func NewRegistrationServiceWithRepo(repo RegistrationRepo) *RegistrationService {
	return &RegistrationService{repo: repo}
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
		Name:     name,
		Email:    email,
		FormData: formData,
		QRHash:   uuid.NewString(),
		Status:   models.AttendeeStatusPending,
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
	case repository.RegistrationOutcomeDuplicateAttendee:
		return nil, ErrAlreadyRegistered
	default:
		return nil, errors.New("registration failed")
	}
}

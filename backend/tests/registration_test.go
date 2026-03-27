package tests

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/google/uuid"
)

// --- mock registration repository -------------------------------------

type mockRegistrationRepo struct {
	outcome      repository.RegistrationOutcome
	err          error
	lastEventID  uuid.UUID
	lastAttendee *models.Attendee
}

func newMockRegistrationRepo() *mockRegistrationRepo {
	return &mockRegistrationRepo{}
}

func (m *mockRegistrationRepo) CreateForEvent(
	_ context.Context,
	eventID uuid.UUID,
	attendee *models.Attendee,
) (repository.RegistrationOutcome, error) {
	m.lastEventID = eventID
	m.lastAttendee = attendee

	if attendee.ID == uuid.Nil {
		attendee.ID = uuid.New()
	}

	if m.err != nil {
		return repository.RegistrationOutcomeUnknown, m.err
	}

	if m.outcome == repository.RegistrationOutcomeUnknown {
		return repository.RegistrationOutcomeCreated, nil
	}

	return m.outcome, nil
}

// --- helpers -----------------------------------------------------------

func validRegistrationInput() service.RegisterAttendeeInput {
	return service.RegisterAttendeeInput{
		EventID:  uuid.New(),
		Name:     " Alice Johnson ",
		Email:    " Alice@Example.COM ",
		FormData: json.RawMessage(`{"company":"Acme"}`),
	}
}

// --- tests -------------------------------------------------------------

func TestRegistrationService_RegisterAttendee_HappyPath(t *testing.T) {
	repo := newMockRegistrationRepo()
	svc := service.NewRegistrationServiceWithRepo(repo)

	attendee, err := svc.RegisterAttendee(
		context.Background(),
		validRegistrationInput(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attendee == nil {
		t.Fatal("expected attendee, got nil")
	}

	if attendee.Name != "Alice Johnson" {
		t.Errorf("expected trimmed name, got %q", attendee.Name)
	}

	if attendee.Email != "alice@example.com" {
		t.Errorf("expected normalised email, got %q", attendee.Email)
	}

	if attendee.QRHash == "" {
		t.Error("expected non-empty qr_hash")
	}

	if attendee.Status != models.AttendeeStatusPending {
		t.Errorf("expected status pending, got %q", attendee.Status)
	}

	if repo.lastEventID == uuid.Nil {
		t.Error("expected event ID to be passed to repository")
	}
}

func TestRegistrationService_RegisterAttendee_EmptyFormDataDefaultsToObject(t *testing.T) {
	repo := newMockRegistrationRepo()
	svc := service.NewRegistrationServiceWithRepo(repo)

	in := validRegistrationInput()
	in.FormData = nil

	attendee, err := svc.RegisterAttendee(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(attendee.FormData) != "{}" {
		t.Errorf("expected form_data default {}, got %s", string(attendee.FormData))
	}
}

func TestRegistrationService_RegisterAttendee_InvalidInput(t *testing.T) {
	repo := newMockRegistrationRepo()
	svc := service.NewRegistrationServiceWithRepo(repo)

	cases := []struct {
		name  string
		input service.RegisterAttendeeInput
	}{
		{
			name: "missing event id",
			input: service.RegisterAttendeeInput{
				EventID: uuid.Nil,
				Name:    "Alice",
				Email:   "alice@example.com",
			},
		},
		{
			name: "empty name",
			input: service.RegisterAttendeeInput{
				EventID: uuid.New(),
				Name:    "   ",
				Email:   "alice@example.com",
			},
		},
		{
			name: "empty email",
			input: service.RegisterAttendeeInput{
				EventID: uuid.New(),
				Name:    "Alice",
				Email:   "   ",
			},
		},
		{
			name: "invalid email format",
			input: service.RegisterAttendeeInput{
				EventID: uuid.New(),
				Name:    "Alice",
				Email:   "not-an-email",
			},
		},
		{
			name: "invalid form_data json",
			input: service.RegisterAttendeeInput{
				EventID:  uuid.New(),
				Name:     "Alice",
				Email:    "alice@example.com",
				FormData: json.RawMessage(`{"broken":`),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.RegisterAttendee(context.Background(), tc.input)
			if !errors.Is(err, service.ErrInvalidRegistrationInput) {
				t.Errorf("expected ErrInvalidRegistrationInput, got: %v", err)
			}
		})
	}
}

func TestRegistrationService_RegisterAttendee_EventNotFound(t *testing.T) {
	repo := newMockRegistrationRepo()
	repo.outcome = repository.RegistrationOutcomeEventNotFound
	svc := service.NewRegistrationServiceWithRepo(repo)

	_, err := svc.RegisterAttendee(context.Background(), validRegistrationInput())
	if !errors.Is(err, service.ErrRegistrationEventNotFound) {
		t.Errorf("expected ErrRegistrationEventNotFound, got: %v", err)
	}
}

func TestRegistrationService_RegisterAttendee_EventFull(t *testing.T) {
	repo := newMockRegistrationRepo()
	repo.outcome = repository.RegistrationOutcomeEventFull
	svc := service.NewRegistrationServiceWithRepo(repo)

	_, err := svc.RegisterAttendee(context.Background(), validRegistrationInput())
	if !errors.Is(err, service.ErrRegistrationEventFull) {
		t.Errorf("expected ErrRegistrationEventFull, got: %v", err)
	}
}

func TestRegistrationService_RegisterAttendee_PaidEventRequiresCheckout(t *testing.T) {
	repo := newMockRegistrationRepo()
	repo.outcome = repository.RegistrationOutcomePaymentRequired
	svc := service.NewRegistrationServiceWithRepo(repo)

	_, err := svc.RegisterAttendee(context.Background(), validRegistrationInput())
	if !errors.Is(err, service.ErrRegistrationRequiresCheckout) {
		t.Errorf("expected ErrRegistrationRequiresCheckout, got: %v", err)
	}
}

func TestRegistrationService_RegisterAttendee_DuplicateAttendee(t *testing.T) {
	repo := newMockRegistrationRepo()
	repo.outcome = repository.RegistrationOutcomeDuplicateAttendee
	svc := service.NewRegistrationServiceWithRepo(repo)

	_, err := svc.RegisterAttendee(context.Background(), validRegistrationInput())
	if !errors.Is(err, service.ErrAlreadyRegistered) {
		t.Errorf("expected ErrAlreadyRegistered, got: %v", err)
	}
}

func TestRegistrationService_RegisterAttendee_RepositoryError(t *testing.T) {
	repo := newMockRegistrationRepo()
	repo.err = errors.New("db down")
	svc := service.NewRegistrationServiceWithRepo(repo)

	_, err := svc.RegisterAttendee(context.Background(), validRegistrationInput())
	if err == nil {
		t.Fatal("expected wrapped repository error")
	}

	if !strings.Contains(err.Error(), "registering attendee") {
		t.Errorf("expected wrapped error message, got: %v", err)
	}
}

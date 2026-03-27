package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/google/uuid"
)

// --- mock check-in repository ----------------------------------------

type mockCheckInRepo struct {
	// FindAttendee behaviour
	attendee *models.Attendee
	findErr  error

	// ProcessCheckIn behaviour
	outcome    repository.CheckInOutcome
	checkin    *models.CheckIn
	processErr error

	// Captured args
	lastEventID    uuid.UUID
	lastAttendeeID uuid.UUID
	lastScannedBy  uuid.UUID
}

func newMockCheckInRepo() *mockCheckInRepo {
	return &mockCheckInRepo{}
}

func (m *mockCheckInRepo) FindAttendeeByHashAndEvent(
	_ context.Context,
	_ string,
	eventID uuid.UUID,
) (*models.Attendee, error) {
	m.lastEventID = eventID
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.attendee, nil
}

func (m *mockCheckInRepo) ProcessCheckIn(
	_ context.Context,
	eventID uuid.UUID,
	attendeeID uuid.UUID,
	scannedByID uuid.UUID,
) (repository.CheckInOutcome, *models.CheckIn, error) {
	m.lastEventID = eventID
	m.lastAttendeeID = attendeeID
	m.lastScannedBy = scannedByID

	if m.processErr != nil {
		return repository.CheckInOutcomeUnknown, nil,
			m.processErr
	}

	if m.outcome == repository.CheckInOutcomeUnknown {
		return repository.CheckInOutcomeSuccess,
			m.checkin, nil
	}

	return m.outcome, m.checkin, nil
}

// --- helpers ---------------------------------------------------------

func validQR(
	eventID uuid.UUID, hash string,
) string {
	return fmt.Sprintf(
		"eventpass://v1/%s/%s", eventID, hash,
	)
}

func defaultAttendee(
	eventID uuid.UUID,
) *models.Attendee {
	return &models.Attendee{
		ID:      uuid.New(),
		EventID: eventID,
		Email:   "alice@example.com",
		Name:    "Alice",
		QRHash:  "abc123hash",
		Status:  models.AttendeeStatusRegistered,
	}
}

func defaultCheckIn(
	attendeeID uuid.UUID,
) *models.CheckIn {
	return &models.CheckIn{
		ID:          uuid.New(),
		AttendeeID:  attendeeID,
		ScannedByID: uuid.New(),
		ScannedAt:   time.Now().UTC(),
	}
}

// --- ParseQR tests ---------------------------------------------------

func TestQRService_ParseQR_ValidV1(t *testing.T) {
	eventID := uuid.New()
	hash := "abc123hash"
	raw := validQR(eventID, hash)

	gotEvent, gotHash, err := service.ParseQR(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEvent != eventID {
		t.Errorf(
			"event ID mismatch: got %s, want %s",
			gotEvent, eventID,
		)
	}
	if gotHash != hash {
		t.Errorf(
			"hash mismatch: got %q, want %q",
			gotHash, hash,
		)
	}
}

func TestQRService_ParseQR_InvalidVersion(t *testing.T) {
	raw := fmt.Sprintf(
		"eventpass://v2/%s/somehash", uuid.New(),
	)
	_, _, err := service.ParseQR(raw)
	if !errors.Is(err, service.ErrInvalidQR) {
		t.Errorf(
			"expected ErrInvalidQR, got: %v", err,
		)
	}
}

func TestQRService_ParseQR_MissingScheme(t *testing.T) {
	raw := "https://example.com/v1/id/hash"
	_, _, err := service.ParseQR(raw)
	if !errors.Is(err, service.ErrInvalidQR) {
		t.Errorf(
			"expected ErrInvalidQR, got: %v", err,
		)
	}
}

func TestQRService_ParseQR_InvalidEventID(t *testing.T) {
	raw := "eventpass://v1/not-a-uuid/hash123"
	_, _, err := service.ParseQR(raw)
	if !errors.Is(err, service.ErrInvalidQR) {
		t.Errorf(
			"expected ErrInvalidQR, got: %v", err,
		)
	}
}

func TestQRService_ParseQR_TooFewSegments(t *testing.T) {
	raw := "eventpass://v1/onlyone"
	_, _, err := service.ParseQR(raw)
	if !errors.Is(err, service.ErrInvalidQR) {
		t.Errorf(
			"expected ErrInvalidQR, got: %v", err,
		)
	}
}

// --- ScanCheckIn tests -----------------------------------------------

func TestQRService_ScanCheckIn_HappyPath(t *testing.T) {
	eventID := uuid.New()
	attendee := defaultAttendee(eventID)
	checkin := defaultCheckIn(attendee.ID)

	repo := newMockCheckInRepo()
	repo.attendee = attendee
	repo.checkin = checkin
	svc := service.NewQRServiceWithRepo(repo)

	raw := validQR(eventID, attendee.QRHash)
	scannerID := uuid.New()

	got, err := svc.ScanCheckIn(
		context.Background(), raw, scannerID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected check-in, got nil")
	}

	if got.ID != checkin.ID {
		t.Errorf(
			"check-in ID mismatch: got %s, want %s",
			got.ID, checkin.ID,
		)
	}

	if repo.lastAttendeeID != attendee.ID {
		t.Errorf(
			"attendee ID not passed to repo: got %s",
			repo.lastAttendeeID,
		)
	}

	if repo.lastScannedBy != scannerID {
		t.Errorf(
			"scanned_by not passed: got %s",
			repo.lastScannedBy,
		)
	}
}

func TestQRService_ScanCheckIn_VenueFull(t *testing.T) {
	eventID := uuid.New()
	attendee := defaultAttendee(eventID)

	repo := newMockCheckInRepo()
	repo.attendee = attendee
	repo.outcome = repository.CheckInOutcomeVenueFull
	svc := service.NewQRServiceWithRepo(repo)

	raw := validQR(eventID, attendee.QRHash)

	_, err := svc.ScanCheckIn(
		context.Background(), raw, uuid.New(),
	)
	if !errors.Is(err, service.ErrVenueFull) {
		t.Errorf(
			"expected ErrVenueFull, got: %v", err,
		)
	}
}

func TestQRService_ScanCheckIn_AlreadyCheckedIn(
	t *testing.T,
) {
	eventID := uuid.New()
	attendee := defaultAttendee(eventID)

	firstScan := time.Date(
		2026, 3, 26, 9, 0, 0, 0, time.UTC,
	)
	existing := &models.CheckIn{
		ID:          uuid.New(),
		AttendeeID:  attendee.ID,
		ScannedByID: uuid.New(),
		ScannedAt:   firstScan,
	}

	repo := newMockCheckInRepo()
	repo.attendee = attendee
	repo.outcome = repository.CheckInOutcomeAlreadyCheckedIn
	repo.checkin = existing
	svc := service.NewQRServiceWithRepo(repo)

	raw := validQR(eventID, attendee.QRHash)

	_, err := svc.ScanCheckIn(
		context.Background(), raw, uuid.New(),
	)

	var alreadyErr *service.AlreadyCheckedInError
	if !errors.As(err, &alreadyErr) {
		t.Fatalf(
			"expected AlreadyCheckedInError, got: %v",
			err,
		)
	}

	if !alreadyErr.FirstScannedAt.Equal(firstScan) {
		t.Errorf(
			"first_scanned_at mismatch: got %v, want %v",
			alreadyErr.FirstScannedAt, firstScan,
		)
	}
}

func TestQRService_ScanCheckIn_InvalidQR(t *testing.T) {
	repo := newMockCheckInRepo()
	svc := service.NewQRServiceWithRepo(repo)

	_, err := svc.ScanCheckIn(
		context.Background(),
		"garbage://data",
		uuid.New(),
	)
	if !errors.Is(err, service.ErrInvalidQR) {
		t.Errorf(
			"expected ErrInvalidQR, got: %v", err,
		)
	}
}

func TestQRService_ScanCheckIn_AttendeeNotFound(
	t *testing.T,
) {
	eventID := uuid.New()

	repo := newMockCheckInRepo()
	repo.attendee = nil // not found
	svc := service.NewQRServiceWithRepo(repo)

	raw := validQR(eventID, "nonexistent_hash")

	_, err := svc.ScanCheckIn(
		context.Background(), raw, uuid.New(),
	)
	if !errors.Is(err, service.ErrAttendeeNotFound) {
		t.Errorf(
			"expected ErrAttendeeNotFound, got: %v", err,
		)
	}
}

func TestQRService_ScanCheckIn_NotEventOrganizer(
	t *testing.T,
) {
	eventID := uuid.New()
	attendee := defaultAttendee(eventID)

	repo := newMockCheckInRepo()
	repo.attendee = attendee
	repo.outcome = repository.CheckInOutcomeNotEventOrganizer
	svc := service.NewQRServiceWithRepo(repo)

	raw := validQR(eventID, attendee.QRHash)

	_, err := svc.ScanCheckIn(
		context.Background(), raw, uuid.New(),
	)
	if !errors.Is(err, service.ErrNotEventOrganizer) {
		t.Errorf(
			"expected ErrNotEventOrganizer, got: %v", err,
		)
	}
}

func TestQRService_ScanCheckIn_RepositoryError(
	t *testing.T,
) {
	eventID := uuid.New()
	attendee := defaultAttendee(eventID)

	repo := newMockCheckInRepo()
	repo.attendee = attendee
	repo.processErr = errors.New("db connection lost")
	svc := service.NewQRServiceWithRepo(repo)

	raw := validQR(eventID, attendee.QRHash)

	_, err := svc.ScanCheckIn(
		context.Background(), raw, uuid.New(),
	)
	if err == nil {
		t.Fatal("expected wrapped repository error")
	}

	errMsg := err.Error()
	if !contains(errMsg, "processing check-in") {
		t.Errorf(
			"expected wrapped error, got: %v", err,
		)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		containsAt(s, substr)
}

func containsAt(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

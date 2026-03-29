package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/google/uuid"
)

// --- mock event repository --------------------------------------------

type mockEventRepo struct {
	events    map[uuid.UUID]*models.Event
	createErr error
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events: make(map[uuid.UUID]*models.Event),
	}
}

func (m *mockEventRepo) Create(
	_ context.Context,
	event *models.Event,
) error {
	if m.createErr != nil {
		return m.createErr
	}

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	m.events[event.ID] = event
	return nil
}

func (m *mockEventRepo) FindByIDAndOrganizer(
	_ context.Context,
	eventID uuid.UUID,
	organizerID uuid.UUID,
) (*models.Event, error) {
	ev, ok := m.events[eventID]
	if !ok || ev.OrganizerID != organizerID {
		return nil, nil
	}

	return ev, nil
}

func (m *mockEventRepo) FindAllByOrganizer(
	_ context.Context,
	organizerID uuid.UUID,
) ([]models.Event, error) {
	var result []models.Event
	for _, ev := range m.events {
		if ev.OrganizerID == organizerID {
			result = append(result, *ev)
		}
	}

	return result, nil
}

func (m *mockEventRepo) Update(
	_ context.Context,
	eventID uuid.UUID,
	organizerID uuid.UUID,
	updates map[string]interface{},
) (int64, error) {
	ev, ok := m.events[eventID]
	if !ok || ev.OrganizerID != organizerID {
		return 0, nil
	}

	if cap, ok := updates["capacity"]; ok {
		ev.Capacity = cap.(int)
	}
	if price, ok := updates["price"]; ok {
		ev.Price = price.(float64)
	}

	if status, ok := updates["status"]; ok {
		ev.Status = status.(models.EventStatus)
	}

	return 1, nil
}

func (m *mockEventRepo) Delete(
	_ context.Context,
	eventID uuid.UUID,
	organizerID uuid.UUID,
) (int64, error) {
	ev, ok := m.events[eventID]
	if !ok || ev.OrganizerID != organizerID {
		return 0, nil
	}

	delete(m.events, eventID)
	return 1, nil
}

func (m *mockEventRepo) FindBySlug(
	_ context.Context,
	slug string,
) (*models.Event, error) {
	for _, ev := range m.events {
		if ev.Slug == slug {
			return ev, nil
		}
	}

	return nil, nil
}

// --- tests ------------------------------------------------------------

var testOrgID = uuid.New()

func validInput() service.CreateEventInput {
	return service.CreateEventInput{
		Title:     "Go Conference 2026",
		Venue:     "Convention Center",
		StartDate: time.Now().Add(48 * time.Hour),
		Capacity:  200,
		Price:     1499.99,
		IsPublic:  true,
	}
}

func TestEventService_CreateEvent_HappyPath(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	event, err := svc.CreateEvent(
		context.Background(), testOrgID, validInput(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.Title != "Go Conference 2026" {
		t.Errorf("unexpected title: %s", event.Title)
	}

	if event.Slug == "" {
		t.Error("expected non-empty slug")
	}

	if event.OrganizerID != testOrgID {
		t.Error("organizer ID mismatch")
	}

	if event.Price != 1499.99 {
		t.Errorf("expected price to be persisted, got %.2f", event.Price)
	}
}

func TestEventService_CreateEvent_MissingFields(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	cases := []struct {
		name  string
		input service.CreateEventInput
	}{
		{
			"empty title",
			service.CreateEventInput{
				Title:     "",
				Venue:     "Place",
				StartDate: time.Now().Add(48 * time.Hour),
			},
		},
		{
			"empty venue",
			service.CreateEventInput{
				Title:     "Event",
				Venue:     "",
				StartDate: time.Now().Add(48 * time.Hour),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateEvent(
				context.Background(), testOrgID, tc.input,
			)
			if !errors.Is(err, service.ErrInvalidEventInput) {
				t.Errorf(
					"expected ErrInvalidEventInput, got: %v",
					err,
				)
			}
		})
	}
}

func TestEventService_CreateEvent_NegativeCapacity(
	t *testing.T,
) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	input := validInput()
	input.Capacity = -5

	_, err := svc.CreateEvent(
		context.Background(), testOrgID, input,
	)
	if !errors.Is(err, service.ErrInvalidEventInput) {
		t.Errorf(
			"expected ErrInvalidEventInput, got: %v", err,
		)
	}
}

func TestEventService_CreateEvent_NegativePrice(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	input := validInput()
	input.Price = -1

	_, err := svc.CreateEvent(
		context.Background(), testOrgID, input,
	)
	if !errors.Is(err, service.ErrInvalidEventInput) {
		t.Errorf(
			"expected ErrInvalidEventInput, got: %v", err,
		)
	}
}

func TestEventService_CreateEvent_PastDate(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	input := validInput()
	input.StartDate = time.Now().Add(-24 * time.Hour)

	_, err := svc.CreateEvent(
		context.Background(), testOrgID, input,
	)
	if !errors.Is(err, service.ErrInvalidEventInput) {
		t.Errorf(
			"expected ErrInvalidEventInput, got: %v", err,
		)
	}
}

func TestEventService_ListEvents_Empty(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	events, err := svc.ListEvents(
		context.Background(), testOrgID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestEventService_GetEvent_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	_, err := svc.GetEvent(
		context.Background(), testOrgID, uuid.New(),
	)
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Errorf(
			"expected ErrEventNotFound, got: %v", err,
		)
	}
}

func TestEventService_UpdateEvent_InvalidStatus(
	t *testing.T,
) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	event, _ := svc.CreateEvent(
		context.Background(), testOrgID, validInput(),
	)

	badStatus := models.EventStatus("invalid")
	_, err := svc.UpdateEvent(
		context.Background(),
		testOrgID,
		event.ID,
		service.UpdateEventInput{Status: &badStatus},
	)
	if !errors.Is(err, service.ErrInvalidEventInput) {
		t.Errorf(
			"expected ErrInvalidEventInput, got: %v", err,
		)
	}
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	cap := 100
	_, err := svc.UpdateEvent(
		context.Background(),
		testOrgID,
		uuid.New(),
		service.UpdateEventInput{Capacity: &cap},
	)
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Errorf(
			"expected ErrEventNotFound, got: %v", err,
		)
	}
}

func TestEventService_UpdateEvent_Price(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	event, _ := svc.CreateEvent(
		context.Background(), testOrgID, validInput(),
	)

	price := 2499.50
	updated, err := svc.UpdateEvent(
		context.Background(),
		testOrgID,
		event.ID,
		service.UpdateEventInput{Price: &price},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Price != price {
		t.Fatalf("expected price %.2f, got %.2f", price, updated.Price)
	}
}

func TestEventService_DeleteEvent_HappyPath(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	event, _ := svc.CreateEvent(
		context.Background(), testOrgID, validInput(),
	)

	err := svc.DeleteEvent(
		context.Background(), testOrgID, event.ID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetEvent(
		context.Background(), testOrgID, event.ID,
	)
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Error("expected event to be deleted")
	}
}

func TestEventService_DeleteEvent_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventServiceWithRepo(repo)

	err := svc.DeleteEvent(
		context.Background(), testOrgID, uuid.New(),
	)
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Errorf(
			"expected ErrEventNotFound, got: %v", err,
		)
	}
}

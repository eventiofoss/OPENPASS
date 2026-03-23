package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/google/uuid"
)

// --- mock form field repository ----------------------------------------

type mockFormFieldRepo struct {
	fields map[uuid.UUID][]models.FormField
}

func newMockFormFieldRepo() *mockFormFieldRepo {
	return &mockFormFieldRepo{
		fields: make(map[uuid.UUID][]models.FormField),
	}
}

func (m *mockFormFieldRepo) CreateBatch(
	_ context.Context,
	fields []models.FormField,
) error {
	if len(fields) == 0 {
		return nil
	}

	eventID := fields[0].EventID
	for i := range fields {
		if fields[i].ID == uuid.Nil {
			fields[i].ID = uuid.New()
		}
	}

	m.fields[eventID] = append(
		m.fields[eventID], fields...,
	)
	return nil
}

func (m *mockFormFieldRepo) FindByEventID(
	_ context.Context,
	eventID uuid.UUID,
) ([]models.FormField, error) {
	return m.fields[eventID], nil
}

func (m *mockFormFieldRepo) DeleteByEventID(
	_ context.Context,
	eventID uuid.UUID,
) error {
	delete(m.fields, eventID)
	return nil
}

// --- helpers -----------------------------------------------------------

func createTestEvent(
	repo *mockEventRepo,
	orgID uuid.UUID,
) *models.Event {
	event := &models.Event{
		ID:          uuid.New(),
		OrganizerID: orgID,
		Title:       "Test Event",
	}
	repo.events[event.ID] = event
	return event
}

func textField(name string) service.FormFieldInput {
	return service.FormFieldInput{
		Name:     name,
		Type:     "text",
		Label:    name,
		Required: true,
	}
}

func selectField(
	name string, opts []string,
) service.FormFieldInput {
	raw, _ := json.Marshal(opts)
	return service.FormFieldInput{
		Name:     name,
		Type:     "select",
		Label:    name,
		Required: true,
		Options:  raw,
	}
}

// --- tests -------------------------------------------------------------

func TestFormService_SetFormFields_HappyPath(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	event := createTestEvent(eventRepo, testOrgID)

	fields, err := svc.SetFormFields(
		context.Background(),
		testOrgID,
		event.ID,
		[]service.FormFieldInput{
			textField("Full Name"),
			selectField(
				"T-Shirt Size", []string{"S", "M", "L"},
			),
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	if fields[0].Name != "Full Name" {
		t.Errorf("expected 'Full Name', got %q", fields[0].Name)
	}

	if fields[1].Type != models.FormFieldTypeSelect {
		t.Errorf(
			"expected select type, got %q", fields[1].Type,
		)
	}
}

func TestFormService_SetFormFields_InvalidType(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	event := createTestEvent(eventRepo, testOrgID)

	_, err := svc.SetFormFields(
		context.Background(),
		testOrgID,
		event.ID,
		[]service.FormFieldInput{
			{Name: "Bad", Type: "unknown", Label: "Bad"},
		},
	)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestFormService_SetFormFields_SelectNoOptions(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	event := createTestEvent(eventRepo, testOrgID)

	_, err := svc.SetFormFields(
		context.Background(),
		testOrgID,
		event.ID,
		[]service.FormFieldInput{
			{
				Name:  "Color",
				Type:  "select",
				Label: "Color",
			},
		},
	)
	if err == nil {
		t.Fatal("expected error for select without options")
	}
}

func TestFormService_SetFormFields_EmptyName(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	event := createTestEvent(eventRepo, testOrgID)

	_, err := svc.SetFormFields(
		context.Background(),
		testOrgID,
		event.ID,
		[]service.FormFieldInput{
			{Name: "", Type: "text", Label: ""},
		},
	)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestFormService_SetFormFields_EventNotFound(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	_, err := svc.SetFormFields(
		context.Background(),
		testOrgID,
		uuid.New(),
		[]service.FormFieldInput{textField("Name")},
	)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
}

func TestFormService_GetFormFields_HappyPath(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	event := createTestEvent(eventRepo, testOrgID)

	_, err := svc.SetFormFields(
		context.Background(),
		testOrgID,
		event.ID,
		[]service.FormFieldInput{
			textField("Email"),
			textField("Phone"),
		},
	)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	fields, err := svc.GetFormFields(
		context.Background(), testOrgID, event.ID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
}

func TestFormService_GetFormFields_EventNotFound(
	t *testing.T,
) {
	eventRepo := newMockEventRepo()
	formRepo := newMockFormFieldRepo()
	svc := service.NewFormServiceWithRepo(
		formRepo, eventRepo,
	)

	_, err := svc.GetFormFields(
		context.Background(), testOrgID, uuid.New(),
	)
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
}

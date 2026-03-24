package tests

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/google/uuid"
)

// --- mock analytics repository ------------------------------------

type mockAnalyticsRepo struct {
	stats          *repository.EventStats
	statsErr       error
	attendeeRows   []repository.AttendeeRow
	streamErr      error
}

func (m *mockAnalyticsRepo) GetEventStats(
	_ context.Context,
	_ uuid.UUID,
) (*repository.EventStats, error) {
	if m.statsErr != nil {
		return nil, m.statsErr
	}

	return m.stats, nil
}

func (m *mockAnalyticsRepo) StreamAttendees(
	_ context.Context,
	_ uuid.UUID,
	fn func(repository.AttendeeRow) error,
) error {
	if m.streamErr != nil {
		return m.streamErr
	}

	for _, row := range m.attendeeRows {
		if err := fn(row); err != nil {
			return err
		}
	}

	return nil
}

// --- mock event repository for analytics tests --------------------

type mockAnalyticsEventRepo struct {
	event    *models.Event
	findErr  error
}

func (m *mockAnalyticsEventRepo) Create(
	_ context.Context,
	_ *models.Event,
) error {
	return nil
}

func (m *mockAnalyticsEventRepo) FindByIDAndOrganizer(
	_ context.Context,
	_ uuid.UUID,
	orgID uuid.UUID,
) (*models.Event, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}

	if m.event == nil {
		return nil, nil
	}

	if m.event.OrganizerID != orgID {
		return nil, nil
	}

	return m.event, nil
}

func (m *mockAnalyticsEventRepo) FindAllByOrganizer(
	_ context.Context,
	_ uuid.UUID,
) ([]models.Event, error) {
	return nil, nil
}

func (m *mockAnalyticsEventRepo) Update(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ map[string]interface{},
) (int64, error) {
	return 0, nil
}

func (m *mockAnalyticsEventRepo) Delete(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) (int64, error) {
	return 0, nil
}

// --- analytics tests ------------------------------------------

var analyticsOrgID = uuid.New()
var analyticsEventID = uuid.New()

func makeAnalyticsEvent() *models.Event {
	return &models.Event{
		ID:          analyticsEventID,
		OrganizerID: analyticsOrgID,
		Capacity:    200,
	}
}

func TestAnalyticsService_Dashboard_HappyPath(
	t *testing.T,
) {
	analyticsRepo := &mockAnalyticsRepo{
		stats: &repository.EventStats{
			Capacity:     200,
			TicketsSold:  150,
			CheckInCount: 80,
			TotalRevenue: 7500.00,
		},
	}
	eventRepo := &mockAnalyticsEventRepo{
		event: makeAnalyticsEvent(),
	}

	svc := service.NewAnalyticsServiceWithRepo(
		analyticsRepo, eventRepo,
	)

	data, err := svc.GetDashboardData(
		context.Background(),
		analyticsOrgID,
		analyticsEventID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.TicketsSold != 150 {
		t.Errorf(
			"expected 150 tickets, got %d",
			data.TicketsSold,
		)
	}

	if data.CheckInCount != 80 {
		t.Errorf(
			"expected 80 check-ins, got %d",
			data.CheckInCount,
		)
	}

	if data.TotalRevenue != 7500.00 {
		t.Errorf(
			"expected 7500.00, got %.2f",
			data.TotalRevenue,
		)
	}

	expectedPct := 40.0 // 80/200 * 100
	if data.CapacityPct != expectedPct {
		t.Errorf(
			"expected %.1f%%, got %.1f%%",
			expectedPct, data.CapacityPct,
		)
	}
}

func TestAnalyticsService_Dashboard_ZeroCapacity(
	t *testing.T,
) {
	analyticsRepo := &mockAnalyticsRepo{
		stats: &repository.EventStats{
			Capacity:     0,
			TicketsSold:  0,
			CheckInCount: 0,
			TotalRevenue: 0,
		},
	}
	event := makeAnalyticsEvent()
	event.Capacity = 0
	eventRepo := &mockAnalyticsEventRepo{event: event}

	svc := service.NewAnalyticsServiceWithRepo(
		analyticsRepo, eventRepo,
	)

	data, err := svc.GetDashboardData(
		context.Background(),
		analyticsOrgID,
		analyticsEventID,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.CapacityPct != 0 {
		t.Errorf(
			"expected 0%% for zero capacity, got %.1f%%",
			data.CapacityPct,
		)
	}
}

func TestAnalyticsService_Dashboard_EventNotFound(
	t *testing.T,
) {
	analyticsRepo := &mockAnalyticsRepo{}
	eventRepo := &mockAnalyticsEventRepo{
		event: nil,
	}

	svc := service.NewAnalyticsServiceWithRepo(
		analyticsRepo, eventRepo,
	)

	_, err := svc.GetDashboardData(
		context.Background(),
		analyticsOrgID,
		uuid.New(),
	)
	if !errors.Is(err, service.ErrEventNotFound) {
		t.Errorf(
			"expected ErrEventNotFound, got: %v", err,
		)
	}
}

func TestAnalyticsService_ExportCSV(t *testing.T) {
	checkedIn := "2026-03-24T10:00:00Z"
	analyticsRepo := &mockAnalyticsRepo{
		attendeeRows: []repository.AttendeeRow{
			{
				Name:        "Alice",
				Email:       "alice@example.com",
				Status:      "checked_in",
				CheckedInAt: &checkedIn,
				CreatedAt:   "2026-03-20T09:00:00Z",
			},
			{
				Name:        "Bob",
				Email:       "bob@example.com",
				Status:      "registered",
				CheckedInAt: nil,
				CreatedAt:   "2026-03-21T09:00:00Z",
			},
		},
	}
	eventRepo := &mockAnalyticsEventRepo{
		event: makeAnalyticsEvent(),
	}

	svc := service.NewAnalyticsServiceWithRepo(
		analyticsRepo, eventRepo,
	)

	var buf bytes.Buffer
	err := svc.ExportAttendeesCSV(
		context.Background(),
		analyticsOrgID,
		analyticsEventID,
		&buf,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(
		buf.String(),
	))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv: %v", err)
	}

	// Header + 2 data rows = 3 records
	if len(records) != 3 {
		t.Fatalf(
			"expected 3 csv rows, got %d",
			len(records),
		)
	}

	header := records[0]
	expectedHeader := []string{
		"Name", "Email", "Status",
		"Checked In At", "Registered At",
	}
	for i, col := range expectedHeader {
		if header[i] != col {
			t.Errorf(
				"header[%d] = %q, want %q",
				i, header[i], col,
			)
		}
	}

	if records[1][0] != "Alice" {
		t.Errorf(
			"expected Alice, got %s", records[1][0],
		)
	}

	if records[2][3] != "" {
		t.Errorf(
			"expected empty checked_in_at for Bob, got %s",
			records[2][3],
		)
	}
}

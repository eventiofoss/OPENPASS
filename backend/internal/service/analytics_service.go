package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
)

// AnalyticsRepo defines analytics storage operations.
type AnalyticsRepo interface {
	GetEventStats(
		ctx context.Context,
		eventID uuid.UUID,
	) (*repository.EventStats, error)
	StreamAttendees(
		ctx context.Context,
		eventID uuid.UUID,
		fn func(repository.AttendeeRow) error,
	) error
}

// DashboardData is the analytics JSON response payload.
type DashboardData struct {
	TicketsSold   int     `json:"tickets_sold"`
	CheckInCount  int     `json:"check_in_count"`
	TotalRevenue  float64 `json:"total_revenue"`
	Capacity      int     `json:"capacity"`
	CapacityPct   float64 `json:"capacity_pct"`
}

// AnalyticsService contains analytics business logic.
type AnalyticsService struct {
	analytics AnalyticsRepo
	events    EventRepo
}

// NewAnalyticsService returns a service backed by
// concrete repositories.
func NewAnalyticsService(
	analytics *repository.AnalyticsRepository,
	events *repository.EventRepository,
) *AnalyticsService {
	return &AnalyticsService{
		analytics: analytics,
		events:    events,
	}
}

// NewAnalyticsServiceWithRepo returns a service using
// any AnalyticsRepo and EventRepo for testing.
func NewAnalyticsServiceWithRepo(
	analytics AnalyticsRepo,
	events EventRepo,
) *AnalyticsService {
	return &AnalyticsService{
		analytics: analytics,
		events:    events,
	}
}

// GetDashboardData fetches cached counters and computes
// the capacity percentage for the given event.
func (s *AnalyticsService) GetDashboardData(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
) (*DashboardData, error) {
	event, err := s.events.FindByIDAndOrganizer(
		ctx, eventID, organizerID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"analytics: fetching event: %w", err,
		)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	stats, err := s.analytics.GetEventStats(
		ctx, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"analytics: fetching stats: %w", err,
		)
	}

	var capacityPct float64
	if stats.Capacity > 0 {
		capacityPct = float64(stats.CheckInCount) /
			float64(stats.Capacity) * 100
	}

	return &DashboardData{
		TicketsSold:  stats.TicketsSold,
		CheckInCount: stats.CheckInCount,
		TotalRevenue: stats.TotalRevenue,
		Capacity:     stats.Capacity,
		CapacityPct:  capacityPct,
	}, nil
}

// ExportAttendeesCSV streams attendee data as CSV rows
// into the provided writer without loading all records.
func (s *AnalyticsService) ExportAttendeesCSV(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
	w io.Writer,
) error {
	event, err := s.events.FindByIDAndOrganizer(
		ctx, eventID, organizerID,
	)
	if err != nil {
		return fmt.Errorf(
			"analytics: fetching event: %w", err,
		)
	}

	if event == nil {
		return ErrEventNotFound
	}

	csvW := csv.NewWriter(w)
	defer csvW.Flush()

	header := []string{
		"Name", "Email", "Status",
		"Checked In At", "Registered At",
	}
	if err := csvW.Write(header); err != nil {
		return fmt.Errorf(
			"analytics: writing csv header: %w", err,
		)
	}

	streamErr := s.analytics.StreamAttendees(
		ctx,
		eventID,
		func(row repository.AttendeeRow) error {
			checkedIn := ""
			if row.CheckedInAt != nil {
				checkedIn = *row.CheckedInAt
			}

			return csvW.Write([]string{
				row.Name,
				row.Email,
				row.Status,
				checkedIn,
				row.CreatedAt,
			})
		},
	)
	if streamErr != nil {
		if errors.Is(streamErr, context.Canceled) {
			return nil
		}

		return fmt.Errorf(
			"analytics: streaming csv: %w", streamErr,
		)
	}

	return nil
}

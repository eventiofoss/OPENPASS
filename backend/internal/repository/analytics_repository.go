package repository

import (
	"context"
	"fmt"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventStats holds the pre-calculated analytics counters.
type EventStats struct {
	Capacity     int     `json:"capacity"`
	TicketsSold  int     `json:"tickets_sold"`
	CheckInCount int     `json:"check_in_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

// AttendeeRow holds one row for CSV streaming.
type AttendeeRow struct {
	Name        string
	Email       string
	Status      string
	CheckedInAt *string
	CreatedAt   string
}

// AnalyticsRepository handles analytics database queries.
type AnalyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository returns a ready repository.
func NewAnalyticsRepository(
	db *gorm.DB,
) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetEventStats fetches pre-calculated counters from events.
func (r *AnalyticsRepository) GetEventStats(
	ctx context.Context,
	eventID uuid.UUID,
) (*EventStats, error) {
	var stats EventStats

	err := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Select(
			"capacity",
			"tickets_sold",
			"check_in_count",
			"total_revenue",
		).
		Where("id = ?", eventID).
		Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf(
			"analytics: fetching stats: %w", err,
		)
	}

	return &stats, nil
}

// IncrementTicketsSold atomically bumps tickets_sold
// and total_revenue on the events row.
func (r *AnalyticsRepository) IncrementTicketsSold(
	ctx context.Context,
	eventID uuid.UUID,
	revenue float64,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"tickets_sold": gorm.Expr(
				"tickets_sold + ?", 1,
			),
			"total_revenue": gorm.Expr(
				"total_revenue + ?", revenue,
			),
		})
	if result.Error != nil {
		return fmt.Errorf(
			"analytics: incrementing tickets: %w",
			result.Error,
		)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"analytics: event %s not found", eventID,
		)
	}

	return nil
}

// IncrementCheckIn atomically bumps check_in_count.
func (r *AnalyticsRepository) IncrementCheckIn(
	ctx context.Context,
	eventID uuid.UUID,
) error {
	result := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("id = ?", eventID).
		Update(
			"check_in_count",
			gorm.Expr("check_in_count + ?", 1),
		)
	if result.Error != nil {
		return fmt.Errorf(
			"analytics: incrementing check-in: %w",
			result.Error,
		)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"analytics: event %s not found", eventID,
		)
	}

	return nil
}

// StreamAttendees iterates attendee rows one at a time
// and calls fn for each, keeping memory constant.
func (r *AnalyticsRepository) StreamAttendees(
	ctx context.Context,
	eventID uuid.UUID,
	fn func(AttendeeRow) error,
) error {
	rows, err := r.db.WithContext(ctx).
		Model(&models.Attendee{}).
		Select(
			"name",
			"email",
			"status",
			"checked_in_at",
			"created_at",
		).
		Where("event_id = ?", eventID).
		Order("created_at ASC").
		Rows()
	if err != nil {
		return fmt.Errorf(
			"analytics: streaming attendees: %w", err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var row AttendeeRow
		if err := rows.Scan(
			&row.Name,
			&row.Email,
			&row.Status,
			&row.CheckedInAt,
			&row.CreatedAt,
		); err != nil {
			return fmt.Errorf(
				"analytics: scanning row: %w", err,
			)
		}

		if err := fn(row); err != nil {
			return fmt.Errorf(
				"analytics: callback error: %w", err,
			)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf(
			"analytics: row iteration: %w", err,
		)
	}

	return nil
}

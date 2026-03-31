package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
)

const maxSlugRetries = 3

var (
	// ErrEventNotFound is returned when the event does not exist.
	ErrEventNotFound = errors.New("event not found")
	// ErrInvalidEventInput is returned on bad event payloads.
	ErrInvalidEventInput = errors.New("invalid event input")
)

// EventRepo defines the database operations used by EventService.
type EventRepo interface {
	Create(ctx context.Context, event *models.Event) error
	FindByIDAndOrganizer(
		ctx context.Context,
		eventID uuid.UUID,
		organizerID uuid.UUID,
	) (*models.Event, error)
	FindBySlug(
		ctx context.Context,
		slug string,
	) (*models.Event, error)
	FindAllByOrganizer(
		ctx context.Context,
		organizerID uuid.UUID,
	) ([]models.Event, error)
	Update(
		ctx context.Context,
		eventID uuid.UUID,
		organizerID uuid.UUID,
		updates map[string]interface{},
	) (int64, error)
	Delete(
		ctx context.Context,
		eventID uuid.UUID,
		organizerID uuid.UUID,
	) (int64, error)
}

// CreateEventInput carries validated create-event fields.
type CreateEventInput struct {
	Title       string
	Description string
	StartDate   time.Time
	Venue       string
	Capacity    int
	Price       float64
	IsPublic    bool
}

// UpdateEventInput carries optional update fields.
type UpdateEventInput struct {
	Capacity *int
	Price    *float64
	Status   *models.EventStatus
}

// EventService contains event business logic.
type EventService struct {
	repo EventRepo
}

type attendeeEventRepo interface {
	FindAllByAttendeeUser(
		ctx context.Context,
		userID uuid.UUID,
	) ([]models.Event, error)
}

type publicEventRepo interface {
	FindPublicPublished(
		ctx context.Context,
		search string,
	) ([]models.Event, error)
}

const (
	EventListViewOrganizing = "organizing"
	EventListViewAttending  = "attending"
	EventListViewAll        = "all"
)

// EventListResult supports dashboard list views for owned and attended events.
type EventListResult struct {
	View             string
	Events           []models.Event
	OrganizingEvents []models.Event
	AttendingEvents  []models.Event
}

// NewEventService returns a service backed by the concrete repo.
func NewEventService(
	repo *repository.EventRepository,
) *EventService {
	return &EventService{repo: repo}
}

// NewEventServiceWithRepo returns a service using any EventRepo.
func NewEventServiceWithRepo(repo EventRepo) *EventService {
	return &EventService{repo: repo}
}

// CreateEvent validates input and persists a new event.
func (s *EventService) CreateEvent(
	ctx context.Context,
	organizerID uuid.UUID,
	input CreateEventInput,
) (*models.Event, error) {
	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	venue := strings.TrimSpace(input.Venue)
	startDate := input.StartDate.UTC()

	if title == "" || venue == "" {
		return nil, fmt.Errorf(
			"%w: title and venue are required",
			ErrInvalidEventInput,
		)
	}

	if input.Capacity < 0 {
		return nil, fmt.Errorf(
			"%w: capacity cannot be negative",
			ErrInvalidEventInput,
		)
	}
	if input.Price < 0 {
		return nil, fmt.Errorf(
			"%w: price cannot be negative",
			ErrInvalidEventInput,
		)
	}

	if startDate.IsZero() || !startDate.After(time.Now().UTC()) {
		return nil, fmt.Errorf(
			"%w: start date must be in the future",
			ErrInvalidEventInput,
		)
	}

	var event models.Event
	var createErr error

	for i := 0; i < maxSlugRetries; i++ {
		slug, slugErr := generateEventSlug(title)
		if slugErr != nil {
			return nil, fmt.Errorf("generating slug: %w", slugErr)
		}

		event = models.Event{
			Slug:        slug,
			OrganizerID: organizerID,
			Title:       title,
			Description: description,
			StartDate:   startDate,
			Venue:       venue,
			Capacity:    input.Capacity,
			Price:       normalizeCurrencyAmount(input.Price),
			IsPublic:    input.IsPublic,
		}

		createErr = s.repo.Create(ctx, &event)
		if createErr == nil {
			return &event, nil
		}

		if strings.Contains(createErr.Error(), "duplicate key value") &&
			strings.Contains(createErr.Error(), "slug") {
			continue
		}

		return nil, fmt.Errorf("creating event: %w", createErr)
	}

	return nil, fmt.Errorf("creating event: %w", createErr)
}

// ListEvents returns all events for the given organizer.
func (s *EventService) ListEvents(
	ctx context.Context,
	organizerID uuid.UUID,
) ([]models.Event, error) {
	events, err := s.repo.FindAllByOrganizer(ctx, organizerID)
	if err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}

	return events, nil
}

// ListEventsByView returns event lists for organizing, attending, or both views.
func (s *EventService) ListEventsByView(
	ctx context.Context,
	userID uuid.UUID,
	view string,
) (*EventListResult, error) {
	resolvedView := normalizeEventListView(view)
	if resolvedView == "" {
		return nil, fmt.Errorf(
			"%w: view must be organizing, attending, or all",
			ErrInvalidEventInput,
		)
	}

	result := &EventListResult{View: resolvedView}

	switch resolvedView {
	case EventListViewOrganizing:
		events, err := s.repo.FindAllByOrganizer(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("listing organizing events: %w", err)
		}
		result.Events = events
		result.OrganizingEvents = events
		return result, nil
	case EventListViewAttending:
		events, err := s.listAttendingEvents(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("listing attending events: %w", err)
		}
		result.Events = events
		result.AttendingEvents = events
		return result, nil
	case EventListViewAll:
		organizingEvents, err := s.repo.FindAllByOrganizer(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("listing organizing events: %w", err)
		}

		attendingEvents, err := s.listAttendingEvents(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("listing attending events: %w", err)
		}

		result.Events = organizingEvents
		result.OrganizingEvents = organizingEvents
		result.AttendingEvents = attendingEvents
		return result, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported list view",
			ErrInvalidEventInput,
		)
	}
}

func (s *EventService) listAttendingEvents(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Event, error) {
	attendeeRepo, ok := s.repo.(attendeeEventRepo)
	if !ok {
		return nil, fmt.Errorf(
			"%w: attending view is not available",
			ErrInvalidEventInput,
		)
	}

	return attendeeRepo.FindAllByAttendeeUser(ctx, userID)
}

// ListPublicEvents returns events visible in the public directory.
func (s *EventService) ListPublicEvents(
	ctx context.Context,
	search string,
) ([]models.Event, error) {
	publicRepo, ok := s.repo.(publicEventRepo)
	if !ok {
		return nil, fmt.Errorf(
			"%w: public listing is not available",
			ErrInvalidEventInput,
		)
	}

	events, err := publicRepo.FindPublicPublished(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("listing public events: %w", err)
	}

	return events, nil
}

// GetEvent returns a single event scoped to the organizer.
func (s *EventService) GetEvent(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
) (*models.Event, error) {
	event, err := s.repo.FindByIDAndOrganizer(
		ctx, eventID, organizerID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching event: %w", err)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	return event, nil
}

// UpdateEvent applies partial updates to an organizer-owned event.
func (s *EventService) UpdateEvent(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
	input UpdateEventInput,
) (*models.Event, error) {
	updates := map[string]interface{}{}

	if input.Capacity != nil {
		if *input.Capacity < 0 {
			return nil, fmt.Errorf(
				"%w: capacity cannot be negative",
				ErrInvalidEventInput,
			)
		}
		updates["capacity"] = *input.Capacity
	}
	if input.Price != nil {
		if *input.Price < 0 {
			return nil, fmt.Errorf(
				"%w: price cannot be negative",
				ErrInvalidEventInput,
			)
		}
		updates["price"] = normalizeCurrencyAmount(*input.Price)
	}

	if input.Status != nil {
		if !isValidEventStatus(*input.Status) {
			return nil, fmt.Errorf(
				"%w: invalid event status",
				ErrInvalidEventInput,
			)
		}
		updates["status"] = *input.Status
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf(
			"%w: no valid fields provided for update",
			ErrInvalidEventInput,
		)
	}

	rows, err := s.repo.Update(
		ctx, eventID, organizerID, updates,
	)
	if err != nil {
		return nil, fmt.Errorf("updating event: %w", err)
	}

	if rows == 0 {
		return nil, ErrEventNotFound
	}

	event, err := s.repo.FindByIDAndOrganizer(
		ctx, eventID, organizerID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"fetching updated event: %w", err,
		)
	}

	return event, nil
}

// DeleteEvent removes an organizer-owned event.
func (s *EventService) DeleteEvent(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
) error {
	rows, err := s.repo.Delete(ctx, eventID, organizerID)
	if err != nil {
		return fmt.Errorf("deleting event: %w", err)
	}

	if rows == 0 {
		return ErrEventNotFound
	}

	return nil
}

// GetPublicEvent returns a single event by slug for public display.
// Only active or full public events are returned.
func (s *EventService) GetPublicEvent(
	ctx context.Context,
	slug string,
) (*models.Event, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf(
			"%w: slug is required", ErrInvalidEventInput,
		)
	}

	event, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("fetching public event: %w", err)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	if !event.IsPublic {
		return nil, ErrEventNotFound
	}

	if event.Status != models.EventStatusActive &&
		event.Status != models.EventStatusFull {
		return nil, ErrEventNotFound
	}

	return event, nil
}

func isValidEventStatus(status models.EventStatus) bool {
	switch status {
	case models.EventStatusDraft,
		models.EventStatusActive,
		models.EventStatusFull,
		models.EventStatusCancelled:
		return true
	default:
		return false
	}
}

func generateEventSlug(title string) (string, error) {
	base := slugify(title)

	suffix, err := randomAlphaNumeric(6)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s-%s", base, strings.ToLower(suffix),
	), nil
}

func slugify(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))

	var b strings.Builder
	prevHyphen := false

	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}

		if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "event"
	}

	return out
}

func normalizeEventListView(view string) string {
	view = strings.ToLower(strings.TrimSpace(view))
	if view == "" {
		return EventListViewOrganizing
	}

	switch view {
	case EventListViewOrganizing, EventListViewAttending, EventListViewAll:
		return view
	default:
		return ""
	}
}

func randomAlphaNumeric(n int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, n)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	for i := range buf {
		buf[i] = chars[int(buf[i])%len(chars)]
	}

	return string(buf), nil
}

func normalizeCurrencyAmount(value float64) float64 {
	return math.Round(value*100) / 100
}

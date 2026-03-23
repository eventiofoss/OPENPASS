package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	// ErrInvalidFormInput is returned on bad form payloads.
	ErrInvalidFormInput = errors.New(
		"invalid form field input",
	)
)

// FormFieldRepo defines database operations for form fields.
type FormFieldRepo interface {
	CreateBatch(
		ctx context.Context,
		fields []models.FormField,
	) error
	FindByEventID(
		ctx context.Context,
		eventID uuid.UUID,
	) ([]models.FormField, error)
	DeleteByEventID(
		ctx context.Context,
		eventID uuid.UUID,
	) error
}

// FormFieldInput carries a single field definition from
// the client request.
type FormFieldInput struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Label    string          `json:"label"`
	Required bool            `json:"required"`
	Options  json.RawMessage `json:"options,omitempty"`
}

// FormService contains form-field business logic.
type FormService struct {
	repo      FormFieldRepo
	eventRepo EventRepo
}

// NewFormService returns a service backed by the concrete repos.
func NewFormService(
	repo *repository.FormFieldRepository,
	eventRepo *repository.EventRepository,
) *FormService {
	return &FormService{
		repo:      repo,
		eventRepo: eventRepo,
	}
}

// NewFormServiceWithRepo returns a service using any repo
// implementations (used for testing).
func NewFormServiceWithRepo(
	repo FormFieldRepo,
	eventRepo EventRepo,
) *FormService {
	return &FormService{
		repo:      repo,
		eventRepo: eventRepo,
	}
}

// SetFormFields replaces all form fields for an event.
func (s *FormService) SetFormFields(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
	inputs []FormFieldInput,
) ([]models.FormField, error) {
	event, err := s.eventRepo.FindByIDAndOrganizer(
		ctx, eventID, organizerID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"checking event ownership: %w", err,
		)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	fields, err := validateFormFields(eventID, inputs)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteByEventID(ctx, eventID); err != nil {
		return nil, fmt.Errorf(
			"clearing existing fields: %w", err,
		)
	}

	if len(fields) == 0 {
		return []models.FormField{}, nil
	}

	if err := s.repo.CreateBatch(ctx, fields); err != nil {
		return nil, fmt.Errorf(
			"saving form fields: %w", err,
		)
	}

	return fields, nil
}

// GetFormFields returns the form schema for an event.
func (s *FormService) GetFormFields(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
) ([]models.FormField, error) {
	event, err := s.eventRepo.FindByIDAndOrganizer(
		ctx, eventID, organizerID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"checking event ownership: %w", err,
		)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	fields, err := s.repo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf(
			"fetching form fields: %w", err,
		)
	}

	return fields, nil
}

// validateFormFields checks each input and converts to models.
func validateFormFields(
	eventID uuid.UUID,
	inputs []FormFieldInput,
) ([]models.FormField, error) {
	seen := make(map[string]bool, len(inputs))
	fields := make([]models.FormField, 0, len(inputs))

	for i, in := range inputs {
		name := strings.TrimSpace(in.Name)
		label := strings.TrimSpace(in.Label)

		if name == "" {
			return nil, fmt.Errorf(
				"%w: field at position %d has empty name",
				ErrInvalidFormInput, i,
			)
		}

		if label == "" {
			label = name
		}

		lower := strings.ToLower(name)
		if seen[lower] {
			return nil, fmt.Errorf(
				"%w: duplicate field name %q",
				ErrInvalidFormInput, name,
			)
		}
		seen[lower] = true

		fieldType := models.FormFieldType(
			strings.ToLower(strings.TrimSpace(in.Type)),
		)
		if !isValidFieldType(fieldType) {
			return nil, fmt.Errorf(
				"%w: invalid type %q for field %q",
				ErrInvalidFormInput, in.Type, name,
			)
		}

		if fieldType == models.FormFieldTypeSelect {
			if len(in.Options) == 0 {
				return nil, fmt.Errorf(
					"%w: select field %q requires options",
					ErrInvalidFormInput, name,
				)
			}

			var opts []string
			if err := json.Unmarshal(
				in.Options, &opts,
			); err != nil || len(opts) == 0 {
				return nil, fmt.Errorf(
					"%w: select field %q options must "+
						"be a non-empty string array",
					ErrInvalidFormInput, name,
				)
			}
		}

		fields = append(fields, models.FormField{
			EventID:  eventID,
			Name:     name,
			Type:     fieldType,
			Label:    label,
			Required: in.Required,
			Options:  in.Options,
			Position: i,
		})
	}

	return fields, nil
}

func isValidFieldType(t models.FormFieldType) bool {
	switch t {
	case models.FormFieldTypeText,
		models.FormFieldTypeEmail,
		models.FormFieldTypeSelect,
		models.FormFieldTypeCheckbox,
		models.FormFieldTypeNumber:
		return true
	default:
		return false
	}
}

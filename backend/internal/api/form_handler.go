package api

import (
	"errors"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SetFormFieldsRequest is the accepted payload for setting
// form fields on an event.
type SetFormFieldsRequest struct {
	Fields []service.FormFieldInput `json:"fields"`
}

// SetFormFields replaces all form fields for an event.
func (h *Handler) SetFormFields(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	req := new(SetFormFieldsRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	fields, err := h.Forms.SetFormFields(
		c.Context(),
		organizerID,
		eventID,
		req.Fields,
	)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		if errors.Is(err, service.ErrInvalidFormInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "Could not save form fields",
			},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Form fields saved successfully",
		"fields":  formatFields(fields),
	})
}

// GetFormFields returns the form schema for an event.
func (h *Handler) GetFormFields(c *fiber.Ctx) error {
	organizerID, err := organizerIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "Authentication required"},
		)
	}

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	fields, err := h.Forms.GetFormFields(
		c.Context(),
		organizerID,
		eventID,
	)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "Could not fetch form fields",
			},
		)
	}

	return c.JSON(fiber.Map{"fields": formatFields(fields)})
}

// formatFields builds a clean response slice without internal
// IDs or foreign keys.
func formatFields(
	fields []models.FormField,
) []fiber.Map {
	result := make([]fiber.Map, 0, len(fields))
	for _, f := range fields {
		entry := fiber.Map{
			"name":     f.Name,
			"type":     f.Type,
			"label":    f.Label,
			"required": f.Required,
			"position": f.Position,
		}

		if len(f.Options) > 0 {
			entry["options"] = f.Options
		}

		result = append(result, entry)
	}

	return result
}

package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateCheckoutRequest is the public payment initiation payload.
type CreateCheckoutRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateCheckout handles POST /api/events/:id/pay.
//
// Security model:
//   - Guest checkout: trust only validated request body.
//   - Authenticated checkout: ignore request name/email and resolve canonical
//     identity from the server-side session.
func (h *Handler) CreateCheckout(c *fiber.Ctx) error {
	if h.Payment == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			fiber.Map{"error": "Payment service unavailable"},
		)
	}

	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid event ID"},
		)
	}

	req := new(CreateCheckoutRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON payload"},
		)
	}

	name, email, checkoutMode, err := h.resolveCheckoutIdentity(c, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCheckoutInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{"error": "Invalid or expired session"},
			)
		}

		slog.Error(
			"checkout identity resolution failed",
			slog.String("event_id", eventID.String()),
			slog.String("error", err.Error()),
		)
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			fiber.Map{"error": "Could not resolve authenticated session"},
		)
	}

	orderID, providerData, err := h.Payment.ReserveTicket(
		eventID,
		name,
		email,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCheckoutInput) {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": err.Error()},
			)
		}

		if errors.Is(err, service.ErrEventNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Event not found"},
			)
		}

		if errors.Is(err, service.ErrEventSoldOut) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": "Event is sold out"},
			)
		}

		if errors.Is(err, service.ErrCheckoutNotRequired) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": "This event does not require checkout"},
			)
		}

		if errors.Is(err, service.ErrCheckoutAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{
					"error": "You already have a checkout for this event",
				},
			)
		}

		if errors.Is(err, service.ErrPaymentServiceNotReady) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{"error": "Payment service unavailable"},
			)
		}

		slog.Error(
			"checkout creation failed",
			slog.String("event_id", eventID.String()),
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Could not create checkout"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "Checkout created",
		"order_id":      orderID,
		"provider_data": providerData,
		"checkout_mode": checkoutMode,
	})
}

// HandleWebhook handles POST /api/webhooks/:gateway.
//
// Operational rule:
//   - Return 200 for successful transitions and idempotent final-state replays.
//   - Return 4xx/5xx for signature failures, unknown orders, or infrastructure
//     errors.
func (h *Handler) HandleWebhook(c *fiber.Ctx) error {
	if h.Payment == nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}

	gateway := normalizedGateway(c.Params("gateway"))
	if gateway == "" || gateway != activePaymentGateway() {
		return c.SendStatus(fiber.StatusNotFound)
	}

	signatureHeader, ok := webhookSignatureHeader(gateway)
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}

	signature := ""
	if signatureHeader != "" {
		signature = strings.TrimSpace(c.Get(signatureHeader))
		if signature == "" {
			slog.Warn(
				"payment webhook missing signature header",
				slog.String("gateway", gateway),
				slog.String("header", signatureHeader),
			)
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	err := h.Payment.ProcessWebhook(c.Body(), signature)
	if err == nil {
		return c.SendStatus(fiber.StatusOK)
	}

	if errors.Is(err, service.ErrWebhookOrderNotFound) {
		slog.Warn(
			"payment webhook rejected because order was not found",
			slog.String("gateway", gateway),
			slog.String("error", err.Error()),
		)
		return c.SendStatus(fiber.StatusNotFound)
	}

	if errors.Is(err, service.ErrReservationState) {
		slog.Warn(
			"payment webhook acknowledged as no-op",
			slog.String("gateway", gateway),
			slog.String("error", err.Error()),
		)
		return c.SendStatus(fiber.StatusOK)
	}

	if errors.Is(err, service.ErrInvalidWebhookSignature) {
		slog.Warn(
			"payment webhook signature verification failed",
			slog.String("gateway", gateway),
			slog.String("error", err.Error()),
		)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if errors.Is(err, service.ErrWebhookAmountMismatch) {
		slog.Warn(
			"payment webhook rejected because payload did not match reservation",
			slog.String("gateway", gateway),
			slog.String("error", err.Error()),
		)
		return c.SendStatus(fiber.StatusConflict)
	}

	if errors.Is(err, service.ErrPaymentServiceNotReady) {
		slog.Error(
			"payment webhook rejected because service is unavailable",
			slog.String("gateway", gateway),
		)
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}

	slog.Error(
		"payment webhook processing failed",
		slog.String("gateway", gateway),
		slog.String("error", err.Error()),
	)
	return c.SendStatus(fiber.StatusInternalServerError)
}

func (h *Handler) resolveCheckoutIdentity(
	c *fiber.Ctx,
	req *CreateCheckoutRequest,
) (string, string, string, error) {
	name, email, ok, err := h.checkoutIdentityFromSession(c)
	if err != nil {
		return "", "", "", err
	}
	if ok {
		return name, email, "authenticated", nil
	}

	name = strings.TrimSpace(req.Name)
	email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := validateGuestCheckoutIdentity(name, email); err != nil {
		return "", "", "", err
	}

	return name, email, "guest", nil
}

func (h *Handler) checkoutIdentityFromSession(
	c *fiber.Ctx,
) (string, string, bool, error) {
	if h.Auth == nil {
		return "", "", false, nil
	}

	if claims, ok := middleware.ClaimsFromContext(c); ok &&
		claims != nil && claims.Subject != "" {
		return h.lookupUserIdentity(c, claims.Subject)
	}

	tokenString := strings.TrimSpace(
		c.Cookies(middleware.SessionCookieName),
	)
	if tokenString == "" {
		return "", "", false, nil
	}

	claims, err := middleware.ParseToken(tokenString)
	if err != nil || claims.Subject == "" {
		expireSessionCookie(c)
		return "", "", false, nil
	}

	return h.lookupUserIdentity(c, claims.Subject)
}

func (h *Handler) lookupUserIdentity(
	c *fiber.Ctx,
	subject string,
) (string, string, bool, error) {
	user, err := h.Auth.GetUser(c.Context(), subject)
	if err != nil {
		return "", "", false, err
	}
	if user == nil {
		return "", "", false, service.ErrInvalidCredentials
	}

	return user.Name, user.Email, true, nil
}

func validateGuestCheckoutIdentity(name, email string) error {
	if name == "" || email == "" {
		return fmt.Errorf(
			"%w: name and email are required",
			service.ErrInvalidCheckoutInput,
		)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf(
			"%w: email format is invalid",
			service.ErrInvalidCheckoutInput,
		)
	}

	return nil
}

func webhookSignatureHeader(gateway string) (string, bool) {
	switch gateway {
	case "razorpay":
		return "X-Razorpay-Signature", true
	case "mock":
		return "", true
	default:
		return "", false
	}
}

func activePaymentGateway() string {
	return normalizedGateway(os.Getenv("ACTIVE_PAYMENT_GATEWAY"))
}

func normalizedGateway(gateway string) string {
	return strings.ToLower(strings.TrimSpace(gateway))
}

func expireSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   isProductionEnv(),
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

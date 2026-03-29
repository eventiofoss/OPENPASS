package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/eventiofoss/eventio/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type paymentTestHarness struct {
	app        *fiber.App
	adminDB    *gorm.DB
	db         *gorm.DB
	authSvc    *service.AuthService
	paymentSvc *service.PaymentService
	schema     string
}

type apiErrorResponse struct {
	Error string `json:"error"`
}

type checkoutResponse struct {
	Message      string                 `json:"message"`
	OrderID      string                 `json:"order_id"`
	ProviderData map[string]interface{} `json:"provider_data"`
	CheckoutMode string                 `json:"checkout_mode"`
	Error        string                 `json:"error"`
}

type mockAmountMismatchProvider struct{}

func (p *mockAmountMismatchProvider) CreateOrder(
	amount int64,
	currency string,
	receiptID string,
) (string, map[string]interface{}, error) {
	orderID := "mock_order_" + receiptID

	return orderID, map[string]interface{}{
		"gateway":  "mock",
		"order_id": orderID,
		"amount":   amount,
		"currency": currency,
		"receipt":  receiptID,
	}, nil
}

func (p *mockAmountMismatchProvider) VerifyPaymentSignature(
	orderID,
	paymentID,
	signature string,
) (bool, error) {
	return true, nil
}

func (p *mockAmountMismatchProvider) VerifyWebhook(
	payload []byte,
	signatureHeader string,
) (*service.PaymentEvent, error) {
	orderID, err := webhookOrderIDFromPayload(payload)
	if err != nil {
		return nil, err
	}

	return &service.PaymentEvent{
		OrderID:    orderID,
		EventType:  "payment.captured",
		Status:     "captured",
		Amount:     1,
		RawEventID: orderID + "-mismatch",
	}, nil
}

func TestPaymentCheckout_RejectsSoldOutEvent(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEvent(t, 1, 1)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+event.ID.String()+"/pay",
		`{"name":"Guest Buyer","email":"guest@example.com"}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", resp.StatusCode)
	}

	var body checkoutResponse
	h.decodeJSON(t, resp, &body)
	if body.Error != "Event is sold out" {
		t.Fatalf("expected sold-out error, got %q", body.Error)
	}

	var attendeeCount int64
	if err := h.db.Model(&models.Attendee{}).
		Where("event_id = ?", event.ID).
		Count(&attendeeCount).Error; err != nil {
		t.Fatalf("count attendees: %v", err)
	}
	if attendeeCount != 0 {
		t.Fatalf("expected no attendee reservation, got %d", attendeeCount)
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 1 {
		t.Fatalf(
			"expected tickets_sold to remain 1, got %d",
			storedEvent.TicketsSold,
		)
	}
}

func TestRegistration_RejectsPaidEventWithoutCheckout(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEvent(t, 3, 0)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+event.ID.String()+"/register",
		`{"name":"Guest Buyer","email":"guest@example.com"}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", resp.StatusCode)
	}

	var body apiErrorResponse
	h.decodeJSON(t, resp, &body)
	if body.Error != "This event requires checkout" {
		t.Fatalf("expected checkout-required error, got %q", body.Error)
	}

	var attendeeCount int64
	if err := h.db.Model(&models.Attendee{}).
		Where("event_id = ?", event.ID).
		Count(&attendeeCount).Error; err != nil {
		t.Fatalf("count attendees: %v", err)
	}
	if attendeeCount != 0 {
		t.Fatalf("expected no attendee rows, got %d", attendeeCount)
	}
}

func TestPaymentCheckout_AuthenticatedIdentityOverridesBody(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEvent(t, 5, 0)
	organizer, sessionCookie := h.createAuthenticatedOrganizer(
		t,
		"Alice Organizer",
		"alice@example.com",
	)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+event.ID.String()+"/pay",
		`{"name":"Mallory","email":"mallory@example.com"}`,
		[]*http.Cookie{sessionCookie},
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var body checkoutResponse
	h.decodeJSON(t, resp, &body)
	if body.CheckoutMode != "authenticated" {
		t.Fatalf(
			"expected checkout_mode authenticated, got %q",
			body.CheckoutMode,
		)
	}
	if !strings.HasPrefix(body.OrderID, "mock_order_") {
		t.Fatalf("expected mock order id, got %q", body.OrderID)
	}
	if gateway, ok := body.ProviderData["gateway"].(string); !ok || gateway != "mock" {
		t.Fatalf("expected mock provider data, got %#v", body.ProviderData)
	}
	expectedAmount := float64(int64(event.Price * 100))
	if amount, ok := body.ProviderData["amount"].(float64); !ok || amount != expectedAmount {
		t.Fatalf(
			"expected provider amount %.0f, got %#v",
			expectedAmount,
			body.ProviderData["amount"],
		)
	}

	attendee := h.mustLoadAttendeeByOrderID(t, body.OrderID)
	if attendee.Name != organizer.Name {
		t.Fatalf("expected canonical organizer name %q, got %q", organizer.Name, attendee.Name)
	}
	if attendee.Email != organizer.Email {
		t.Fatalf("expected canonical organizer email %q, got %q", organizer.Email, attendee.Email)
	}
	if attendee.Status != models.AttendeeStatusReserved {
		t.Fatalf("expected attendee status reserved, got %q", attendee.Status)
	}
	if attendee.ReservedAt == nil {
		t.Fatal("expected reserved_at to be set")
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 1 {
		t.Fatalf("expected tickets_sold to become 1, got %d", storedEvent.TicketsSold)
	}
}

func TestPaymentCheckout_RejectsFreeEvent(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEventWithState(t, 5, 0, 0, 0)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+event.ID.String()+"/pay",
		`{"name":"Guest Buyer","email":"guest@example.com"}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", resp.StatusCode)
	}

	var body apiErrorResponse
	h.decodeJSON(t, resp, &body)
	if body.Error != "This event does not require checkout" {
		t.Fatalf("expected free-event error, got %q", body.Error)
	}
}

func TestPaymentCheckout_UsesHistoricalRegistrationsInCapacity(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEventWithState(t, 1, 0, 1, 1499.50)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+event.ID.String()+"/pay",
		`{"name":"Guest Buyer","email":"guest@example.com"}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", resp.StatusCode)
	}

	var body apiErrorResponse
	h.decodeJSON(t, resp, &body)
	if body.Error != "Event is sold out" {
		t.Fatalf("expected sold-out error, got %q", body.Error)
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 0 {
		t.Fatalf("expected tickets_sold to remain 0, got %d", storedEvent.TicketsSold)
	}
}

func TestPaymentWebhook_IsIdempotentAcrossRepeatedDelivery(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEvent(t, 3, 0)

	checkout := h.createGuestCheckout(
		t,
		event.ID,
		"Retry Safe Guest",
		"retry@example.com",
	)

	payload := `{"order_id":"` + checkout.OrderID + `"}`

	for i := 0; i < 3; i++ {
		resp := h.performJSON(
			t,
			http.MethodPost,
			"/api/webhooks/mock",
			payload,
			nil,
		)

		if resp.StatusCode != fiber.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf(
				"attempt %d: expected 200 OK, got %d body=%s",
				i+1,
				resp.StatusCode,
				string(b),
			)
		}
		resp.Body.Close()
	}

	attendee := h.mustLoadAttendeeByOrderID(t, checkout.OrderID)
	if attendee.Status != models.AttendeeStatusPaid {
		t.Fatalf("expected attendee status paid, got %q", attendee.Status)
	}

	var attendeeCount int64
	if err := h.db.Model(&models.Attendee{}).
		Where("event_id = ?", event.ID).
		Count(&attendeeCount).Error; err != nil {
		t.Fatalf("count attendees: %v", err)
	}
	if attendeeCount != 1 {
		t.Fatalf("expected one attendee row, got %d", attendeeCount)
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 1 {
		t.Fatalf("expected tickets_sold to remain 1, got %d", storedEvent.TicketsSold)
	}
	if storedEvent.TotalRevenue != event.Price {
		t.Fatalf(
			"expected total_revenue %.2f, got %.2f",
			event.Price,
			storedEvent.TotalRevenue,
		)
	}
}

func TestPaymentSweeper_ReleasesExpiredReservationAndLateWebhookIsAcked(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEvent(t, 2, 0)

	checkout := h.createGuestCheckout(
		t,
		event.ID,
		"Timeout Guest",
		"timeout@example.com",
	)

	expiredAt := time.Now().UTC().Add(-11 * time.Minute)
	if err := h.db.Model(&models.Attendee{}).
		Where("payment_order_id = ?", checkout.OrderID).
		Update("reserved_at", expiredAt).Error; err != nil {
		t.Fatalf("backdate reserved_at: %v", err)
	}

	if err := h.paymentSvc.RunSweeper(context.Background()); err != nil {
		t.Fatalf("run sweeper: %v", err)
	}

	attendee := h.mustLoadAttendeeByOrderID(t, checkout.OrderID)
	if attendee.Status != models.AttendeeStatusAbandoned {
		t.Fatalf("expected attendee status abandoned, got %q", attendee.Status)
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 0 {
		t.Fatalf("expected tickets_sold to return to 0, got %d", storedEvent.TicketsSold)
	}

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/webhooks/mock",
		`{"order_id":"`+checkout.OrderID+`"}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf(
			"expected 200 OK for late webhook ack, got %d body=%s",
			resp.StatusCode,
			string(b),
		)
	}

	attendee = h.mustLoadAttendeeByOrderID(t, checkout.OrderID)
	if attendee.Status != models.AttendeeStatusAbandoned {
		t.Fatalf(
			"expected late webhook to leave attendee abandoned, got %q",
			attendee.Status,
		)
	}

	storedEvent = h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 0 {
		t.Fatalf(
			"expected late webhook to leave tickets_sold at 0, got %d",
			storedEvent.TicketsSold,
		)
	}
}

func TestPaymentCheckout_AllowsRetryAfterAbandonedReservation(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEvent(t, 2, 0)

	firstCheckout := h.createGuestCheckout(
		t,
		event.ID,
		"Retry Guest",
		"retry-guest@example.com",
	)

	expiredAt := time.Now().UTC().Add(-11 * time.Minute)
	if err := h.db.Model(&models.Attendee{}).
		Where("payment_order_id = ?", firstCheckout.OrderID).
		Update("reserved_at", expiredAt).Error; err != nil {
		t.Fatalf("backdate reserved_at: %v", err)
	}

	if err := h.paymentSvc.RunSweeper(context.Background()); err != nil {
		t.Fatalf("run sweeper: %v", err)
	}

	secondCheckout := h.createGuestCheckout(
		t,
		event.ID,
		"Retry Guest",
		"retry-guest@example.com",
	)
	if secondCheckout.OrderID == firstCheckout.OrderID {
		t.Fatalf("expected a fresh checkout order id, got %q", secondCheckout.OrderID)
	}

	var attendeeCount int64
	if err := h.db.Model(&models.Attendee{}).
		Where("event_id = ? AND email = ?", event.ID, "retry-guest@example.com").
		Count(&attendeeCount).Error; err != nil {
		t.Fatalf("count attendees: %v", err)
	}
	if attendeeCount != 1 {
		t.Fatalf("expected one attendee row after retry, got %d", attendeeCount)
	}

	attendee := h.mustLoadAttendeeByOrderID(t, secondCheckout.OrderID)
	if attendee.Status != models.AttendeeStatusReserved {
		t.Fatalf("expected attendee status reserved, got %q", attendee.Status)
	}
	if attendee.ReservedAt == nil {
		t.Fatal("expected reserved_at to be reset on retry")
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TicketsSold != 1 {
		t.Fatalf("expected tickets_sold to return to 1, got %d", storedEvent.TicketsSold)
	}
}

func TestRegistration_FreeEventKeepsOccupiedInventoryInSync(t *testing.T) {
	h := newPaymentTestHarness(t)
	event := h.createEventWithState(t, 3, 0, 0, 0)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+event.ID.String()+"/register",
		`{"name":"Free Guest","email":"free@example.com","form_data":{"company":"Acme"}}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TotalRegistered != 1 {
		t.Fatalf("expected total_registered to become 1, got %d", storedEvent.TotalRegistered)
	}
	if storedEvent.TicketsSold != 1 {
		t.Fatalf("expected tickets_sold to become 1, got %d", storedEvent.TicketsSold)
	}
}

func TestPaymentWebhook_RejectsAmountMismatch(t *testing.T) {
	h := newPaymentTestHarnessWithProvider(t, &mockAmountMismatchProvider{})
	event := h.createEvent(t, 2, 0)

	checkout := h.createGuestCheckout(
		t,
		event.ID,
		"Webhook Guest",
		"webhook-guest@example.com",
	)

	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/webhooks/mock",
		`{"order_id":"`+checkout.OrderID+`"}`,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf(
			"expected 409 Conflict for amount mismatch, got %d body=%s",
			resp.StatusCode,
			string(b),
		)
	}

	attendee := h.mustLoadAttendeeByOrderID(t, checkout.OrderID)
	if attendee.Status != models.AttendeeStatusReserved {
		t.Fatalf("expected attendee to remain reserved, got %q", attendee.Status)
	}

	storedEvent := h.mustLoadEvent(t, event.ID)
	if storedEvent.TotalRevenue != 0 {
		t.Fatalf("expected total_revenue to remain 0, got %.2f", storedEvent.TotalRevenue)
	}
}

func newPaymentTestHarness(t *testing.T) *paymentTestHarness {
	return newPaymentTestHarnessWithProvider(t, nil)
}

func newPaymentTestHarnessWithProvider(
	t *testing.T,
	provider service.PaymentProvider,
) *paymentTestHarness {
	t.Helper()

	t.Setenv("ACTIVE_PAYMENT_GATEWAY", "mock")
	t.Setenv("PAYMENT_CURRENCY", "INR")
	t.Setenv("APP_ENV", "test")

	baseDSN := paymentTestDatabaseURL()
	adminDB, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres admin db: %v", err)
	}

	schema := "payment_it_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := adminDB.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create test schema %s: %v", schema, err)
	}

	schemaDSN := dsnWithSearchPath(baseDSN, schema)
	db, err := gorm.Open(postgres.Open(schemaDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres schema db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Organizer{},
		&models.Event{},
		&models.Attendee{},
		&models.Payment{},
		&models.CheckIn{},
	); err != nil {
		t.Fatalf("migrate test schema: %v", err)
	}

	orgRepo := repository.NewOrganizerRepository(db)
	authSvc := service.NewAuthService(orgRepo)
	attendeeRepo := repository.NewAttendeeRepository(db)
	registrationSvc := service.NewRegistrationService(attendeeRepo)

	if provider == nil {
		var err error
		provider, err = service.NewPaymentProvider("mock")
		if err != nil {
			t.Fatalf("create mock payment provider: %v", err)
		}
	}
	paymentSvc := service.NewPaymentService(db, provider)

	h := &Handler{
		DB:           db,
		Auth:         authSvc,
		Registration: registrationSvc,
		Payment:      paymentSvc,
	}

	app := fiber.New()
	app.Post("/api/events/:id/register", h.RegisterAttendee)
	app.Post("/api/events/:id/pay", h.CreateCheckout)
	app.Post("/api/webhooks/:gateway", h.HandleWebhook)

	t.Cleanup(func() {
		_ = app.Shutdown()

		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}

		if err := adminDB.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE").Error; err != nil {
			t.Fatalf("drop test schema %s: %v", schema, err)
		}

		if sqlDB, err := adminDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return &paymentTestHarness{
		app:        app,
		adminDB:    adminDB,
		db:         db,
		authSvc:    authSvc,
		paymentSvc: paymentSvc,
		schema:     schema,
	}
}

func paymentTestDatabaseURL() string {
	if dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")); dsn != "" {
		return dsn
	}

	return "host=localhost user=postgres password=postgres dbname=eventio_test port=5432 sslmode=disable"
}

func dsnWithSearchPath(baseDSN, schema string) string {
	if strings.Contains(baseDSN, "://") {
		parsed, err := url.Parse(baseDSN)
		if err == nil {
			query := parsed.Query()
			query.Set("search_path", schema)
			parsed.RawQuery = query.Encode()
			return parsed.String()
		}
	}

	return strings.TrimSpace(baseDSN) + " search_path=" + schema
}

func (h *paymentTestHarness) createEvent(
	t *testing.T,
	capacity int,
	ticketsSold int,
) *models.Event {
	t.Helper()

	return h.createEventWithState(
		t,
		capacity,
		ticketsSold,
		0,
		1499.50,
	)
}

func (h *paymentTestHarness) createEventWithState(
	t *testing.T,
	capacity int,
	ticketsSold int,
	totalRegistered int,
	price float64,
) *models.Event {
	t.Helper()

	organizer := h.createOrganizerFixture(t)

	event := &models.Event{
		OrganizerID:     organizer.ID,
		Title:           "Payment Test Event",
		Description:     "Integration test fixture",
		StartDate:       time.Now().UTC().Add(24 * time.Hour),
		Venue:           "Test Hall",
		Capacity:        capacity,
		Price:           price,
		Status:          models.EventStatusActive,
		TotalRegistered: totalRegistered,
		TicketsSold:     ticketsSold,
		IsPublic:        true,
	}

	if err := h.db.Create(event).Error; err != nil {
		t.Fatalf("create event: %v", err)
	}

	return event
}

func (h *paymentTestHarness) createOrganizerFixture(
	t *testing.T,
) *models.Organizer {
	t.Helper()

	organizer := &models.Organizer{
		Name:         "Fixture Organizer",
		Email:        "fixture+" + uuid.NewString() + "@example.com",
		PasswordHash: "test-password-hash",
		Role:         models.OrganizerRoleOrganizer,
	}

	if err := h.db.Create(organizer).Error; err != nil {
		t.Fatalf("create organizer fixture: %v", err)
	}

	return organizer
}

func (h *paymentTestHarness) createAuthenticatedOrganizer(
	t *testing.T,
	name string,
	email string,
) (*models.Organizer, *http.Cookie) {
	t.Helper()

	organizer, err := h.authSvc.Register(
		context.Background(),
		name,
		email,
		"secureP@ss123",
		"",
	)
	if err != nil {
		t.Fatalf("register organizer: %v", err)
	}

	token, err := h.authSvc.Login(
		context.Background(),
		email,
		"secureP@ss123",
	)
	if err != nil {
		t.Fatalf("login organizer: %v", err)
	}

	return organizer, &http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: token,
	}
}

func (h *paymentTestHarness) createGuestCheckout(
	t *testing.T,
	eventID uuid.UUID,
	name string,
	email string,
) checkoutResponse {
	t.Helper()

	body := `{"name":"` + name + `","email":"` + email + `"}`
	resp := h.performJSON(
		t,
		http.MethodPost,
		"/api/events/"+eventID.String()+"/pay",
		body,
		nil,
	)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf(
			"expected 201 Created, got %d body=%s",
			resp.StatusCode,
			string(payload),
		)
	}

	var checkout checkoutResponse
	h.decodeJSON(t, resp, &checkout)
	return checkout
}

func (h *paymentTestHarness) performJSON(
	t *testing.T,
	method string,
	target string,
	body string,
	cookies []*http.Cookie,
) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := h.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test %s %s: %v", method, target, err)
	}

	return resp
}

func (h *paymentTestHarness) decodeJSON(
	t *testing.T,
	resp *http.Response,
	out interface{},
) {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode response body %s: %v", string(body), err)
	}
}

func (h *paymentTestHarness) mustLoadAttendeeByOrderID(
	t *testing.T,
	orderID string,
) *models.Attendee {
	t.Helper()

	var attendee models.Attendee
	if err := h.db.Where("payment_order_id = ?", orderID).
		First(&attendee).Error; err != nil {
		t.Fatalf("load attendee by order id %s: %v", orderID, err)
	}

	return &attendee
}

func (h *paymentTestHarness) mustLoadEvent(
	t *testing.T,
	eventID uuid.UUID,
) *models.Event {
	t.Helper()

	var event models.Event
	if err := h.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		t.Fatalf("load event %s: %v", eventID, err)
	}

	return &event
}

func webhookOrderIDFromPayload(payload []byte) (string, error) {
	var body struct {
		OrderID string `json:"order_id"`
	}

	if err := json.Unmarshal(payload, &body); err != nil {
		return "", err
	}
	if strings.TrimSpace(body.OrderID) == "" {
		return "", io.EOF
	}

	return strings.TrimSpace(body.OrderID), nil
}

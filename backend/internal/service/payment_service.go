package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	capturedPaymentStatus   = "captured"
	failedPaymentStatus     = "failed"
	ignoredPaymentStatus    = "ignored"
	defaultReservationTTL   = 10 * time.Minute
	defaultSweeperBatchSize = 100
	defaultPaymentCurrency  = "INR"
	minorUnitsPerMajor      = 100

	// "PAY_SWEE" as a stable Postgres advisory lock namespace.
	defaultSweeperLockKey int64 = 0x5041595F53574545
)

var (
	ErrPaymentServiceNotReady = errors.New("payment service is not ready")
	ErrInvalidCheckoutInput   = errors.New("invalid checkout input")
	ErrCheckoutNotRequired    = errors.New("event does not require checkout")
	ErrEventSoldOut           = errors.New("event is sold out")
	ErrCheckoutAlreadyExists  = errors.New(
		"attendee already registered for this event",
	)
	ErrReservationNotFound   = errors.New("reservation not found")
	ErrReservationState      = errors.New("reservation is not in a payable state")
	ErrWebhookOrderNotFound  = errors.New("payment order not found")
	ErrWebhookAmountMismatch = errors.New(
		"payment webhook amount does not match reservation",
	)
)

// PaymentService owns transactional reservation, sweeping, and webhook flows.
type PaymentService struct {
	db               *gorm.DB
	provider         PaymentProvider
	reservationTTL   time.Duration
	sweeperBatchSize int
	sweeperLockKey   int64
	defaultCurrency  string
}

// NewPaymentService wires the payment orchestration service.
func NewPaymentService(
	db *gorm.DB,
	provider PaymentProvider,
) *PaymentService {
	currency := strings.ToUpper(strings.TrimSpace(os.Getenv("PAYMENT_CURRENCY")))
	if currency == "" {
		currency = defaultPaymentCurrency
	}

	return &PaymentService{
		db:               db,
		provider:         provider,
		reservationTTL:   defaultReservationTTL,
		sweeperBatchSize: defaultSweeperBatchSize,
		sweeperLockKey:   defaultSweeperLockKey,
		defaultCurrency:  currency,
	}
}

type reservedInventoryRow struct {
	ID          uuid.UUID `gorm:"column:id"`
	TicketsSold int       `gorm:"column:tickets_sold"`
	Capacity    int       `gorm:"column:capacity"`
	Price       float64   `gorm:"column:price"`
}

type checkoutEventState struct {
	ID              uuid.UUID `gorm:"column:id"`
	Capacity        int       `gorm:"column:capacity"`
	Price           float64   `gorm:"column:price"`
	TicketsSold     int       `gorm:"column:tickets_sold"`
	TotalRegistered int       `gorm:"column:total_registered"`
}

type releasedInventoryRow struct {
	EventID  uuid.UUID `gorm:"column:event_id"`
	Released int64     `gorm:"column:released"`
}

type advisoryLockRow struct {
	Acquired bool `gorm:"column:acquired"`
}

type checkoutAttendeeRow struct {
	ID     uuid.UUID             `gorm:"column:id"`
	Status models.AttendeeStatus `gorm:"column:status"`
}

type webhookReservationRow struct {
	AttendeeID      uuid.UUID             `gorm:"column:attendee_id"`
	EventID         uuid.UUID             `gorm:"column:event_id"`
	Status          models.AttendeeStatus `gorm:"column:status"`
	ReservedAmount  float64               `gorm:"column:reserved_amount"`
	PaymentRecorded bool                  `gorm:"column:payment_recorded"`
}

type webhookOrderEnvelope struct {
	OrderID string `json:"order_id"`
	Payload struct {
		Payment struct {
			Entity struct {
				OrderID string `json:"order_id"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
	Data struct {
		Object struct {
			OrderID string `json:"order_id"`
		} `json:"object"`
	} `json:"data"`
}

// ReserveTicket atomically reserves one unit of inventory for a payment attempt.
//
// The critical section is kept entirely inside Postgres:
//  1. Atomically increment event inventory with UPDATE ... RETURNING.
//  2. Create or reactivate one reserved attendee row for the buyer.
//
// The provider call is intentionally executed after commit so we never hold row
// locks open across a network boundary. If order creation fails, a compensating
// transaction releases the reservation.
func (s *PaymentService) ReserveTicket(
	eventID uuid.UUID,
	name string,
	email string,
) (string, map[string]interface{}, error) {
	if s.db == nil || s.provider == nil {
		return "", nil, ErrPaymentServiceNotReady
	}
	if eventID == uuid.Nil {
		return "", nil, fmt.Errorf(
			"%w: event_id is required",
			ErrInvalidCheckoutInput,
		)
	}

	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" {
		return "", nil, fmt.Errorf(
			"%w: name and email are required",
			ErrInvalidCheckoutInput,
		)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", nil, fmt.Errorf(
			"%w: email format is invalid",
			ErrInvalidCheckoutInput,
		)
	}

	receiptID := uuid.NewString()
	now := time.Now().UTC()
	var inventory reservedInventoryRow

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Raw(`
			UPDATE events
			SET tickets_sold = GREATEST(tickets_sold, total_registered) + 1,
			    updated_at = NOW()
			WHERE id = ?
			  AND price > 0
			  AND GREATEST(tickets_sold, total_registered) < capacity
			RETURNING id, tickets_sold, capacity, price
		`, eventID).Scan(&inventory)
		if result.Error != nil {
			return fmt.Errorf("reserve inventory: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			state, err := s.loadCheckoutEventState(tx, eventID)
			if err != nil {
				return err
			}
			if state == nil {
				return ErrEventNotFound
			}
			if priceToMinorUnits(state.Price) == 0 {
				return ErrCheckoutNotRequired
			}
			if occupiedInventory(
				state.TotalRegistered,
				state.TicketsSold,
			) >= state.Capacity {
				return ErrEventSoldOut
			}

			return errors.New("reserve inventory: event state changed unexpectedly")
		}

		existing, err := s.findCheckoutAttendee(tx, eventID, email)
		if err != nil {
			return err
		}

		if existing == nil {
			attendee := models.Attendee{
				EventID:        eventID,
				Email:          email,
				Name:           name,
				FormData:       json.RawMessage(`{}`),
				Status:         models.AttendeeStatusReserved,
				ReservedAt:     &now,
				PaymentOrderID: receiptID,
			}

			if err := tx.Create(&attendee).Error; err != nil {
				if isDuplicateCheckoutErr(err) {
					return ErrCheckoutAlreadyExists
				}

				return fmt.Errorf("create reserved attendee: %w", err)
			}

			return nil
		}

		if !canRetryCheckout(existing.Status) {
			return ErrCheckoutAlreadyExists
		}

		result = tx.Model(&models.Attendee{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"name":             name,
				"form_data":        json.RawMessage(`{}`),
				"status":           models.AttendeeStatusReserved,
				"reserved_at":      now,
				"payment_order_id": receiptID,
				"updated_at":       now,
			})
		if result.Error != nil {
			return fmt.Errorf("reactivate reservation: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrReservationNotFound
		}

		return nil
	}); err != nil {
		return "", nil, err
	}

	amount := priceToMinorUnits(inventory.Price)
	orderID, providerData, err := s.provider.CreateOrder(
		amount, s.defaultCurrency, receiptID,
	)
	if err != nil {
		releaseErr := s.releaseReservation(eventID, receiptID)
		if releaseErr != nil {
			return "", nil, fmt.Errorf(
				"create provider order: %w; release reservation: %v",
				err,
				releaseErr,
			)
		}

		return "", nil, fmt.Errorf("create provider order: %w", err)
	}

	if err := s.attachProviderOrderID(receiptID, orderID); err != nil {
		releaseErr := s.releaseReservation(eventID, receiptID)
		if releaseErr != nil {
			return "", nil, fmt.Errorf(
				"persist provider order id: %w; release reservation: %v",
				err,
				releaseErr,
			)
		}

		return "", nil, err
	}
	if err := s.recordPendingPayment(orderID, inventory.Price); err != nil {
		releaseErr := s.releaseReservation(eventID, orderID)
		if releaseErr != nil {
			return "", nil, fmt.Errorf(
				"persist payment ledger: %w; release reservation: %v",
				err,
				releaseErr,
			)
		}

		return "", nil, err
	}

	return orderID, providerData, nil
}

// RunSweeper releases expired reservations in a multi-instance-safe way.
//
// Concurrency strategy:
//  1. Acquire pg_try_advisory_xact_lock so only one API instance sweeps at once.
//  2. Select a batch of expired reservations with FOR UPDATE SKIP LOCKED.
//  3. Mark those attendees abandoned.
//  4. Decrement tickets_sold once per event inside the same transaction.
func (s *PaymentService) RunSweeper(ctx context.Context) error {
	if s.db == nil {
		return ErrPaymentServiceNotReady
	}

	cutoff := time.Now().UTC().Add(-s.reservationTTL)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acquired, err := s.tryAcquireSweeperLock(tx)
		if err != nil {
			return err
		}
		if !acquired {
			return nil
		}

		releasedRows, err := s.releaseExpiredReservations(tx, cutoff)
		if err != nil {
			return err
		}

		for _, row := range releasedRows {
			if row.Released <= 0 {
				continue
			}

			result := tx.Exec(`
				UPDATE events
				SET tickets_sold = GREATEST(tickets_sold - ?, 0),
				    updated_at = NOW()
				WHERE id = ?
			`, row.Released, row.EventID)
			if result.Error != nil {
				return fmt.Errorf(
					"decrement inventory for event %s: %w",
					row.EventID,
					result.Error,
				)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf(
					"decrement inventory for event %s: event not found",
					row.EventID,
				)
			}
		}

		return nil
	})
}

// ProcessWebhook verifies the provider signature and performs an idempotent
// state transition from reserved -> paid.
//
// Duplicate delivery is safe:
//   - reserved -> paid happens once.
//   - already paid is treated as success.
//   - abandoned/cancelled rows are rejected so a late webhook cannot silently
//     re-consume inventory that the sweeper already released.
func (s *PaymentService) ProcessWebhook(
	payload []byte,
	signature string,
) error {
	if s.db == nil || s.provider == nil {
		return ErrPaymentServiceNotReady
	}

	paymentEvent, err := s.provider.VerifyWebhook(payload, signature)
	if err != nil {
		return fmt.Errorf("verify webhook: %w", err)
	}
	if paymentEvent == nil {
		return errors.New("verify webhook: provider returned nil event")
	}

	orderID := strings.TrimSpace(paymentEvent.OrderID)

	switch paymentEvent.Status {
	case capturedPaymentStatus:
	case failedPaymentStatus:
		slog.Warn(
			"payment webhook reported failed payment",
			slog.String("event_type", paymentEvent.EventType),
			slog.String("order_id", orderID),
			slog.String("provider_payment_id", paymentEvent.ProviderPaymentID),
		)
		return nil
	default:
		slog.Info(
			"payment webhook ignored",
			slog.String("event_type", paymentEvent.EventType),
			slog.String("order_id", orderID),
			slog.String("provider_payment_id", paymentEvent.ProviderPaymentID),
		)
		return nil
	}

	if orderID == "" {
		orderID, err = extractWebhookOrderID(payload)
		if err != nil {
			return err
		}
	}

	now := time.Now().UTC()

	return s.db.Transaction(func(tx *gorm.DB) error {
		reservation, err := s.loadWebhookReservation(tx, orderID)
		if err != nil {
			return err
		}
		if reservation == nil {
			return ErrWebhookOrderNotFound
		}
		if !reservation.PaymentRecorded {
			return fmt.Errorf(
				"load webhook reservation: no payment ledger row found for order %q",
				orderID,
			)
		}

		if err := validateCapturedPaymentAmount(
			paymentEvent.Amount,
			reservation.ReservedAmount,
		); err != nil {
			return err
		}

		if reservation.Status == models.AttendeeStatusPaid {
			return nil
		}
		if reservation.Status != models.AttendeeStatusReserved {
			return fmt.Errorf(
				"%w: current status is %q",
				ErrReservationState,
				reservation.Status,
			)
		}

		result := tx.Model(&models.Attendee{}).
			Where(
				"id = ? AND status = ?",
				reservation.AttendeeID,
				models.AttendeeStatusReserved,
			).
			Updates(map[string]interface{}{
				"status":      models.AttendeeStatusPaid,
				"reserved_at": nil,
				"updated_at":  now,
			})
		if result.Error != nil {
			return fmt.Errorf("mark attendee paid: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf(
				"mark attendee paid: reservation %s was not updated",
				reservation.AttendeeID,
			)
		}

		eventResult := tx.Model(&models.Event{}).
			Where("id = ?", reservation.EventID).
			Updates(map[string]interface{}{
				"total_registered": gorm.Expr(
					"total_registered + ?",
					1,
				),
				"total_revenue": gorm.Expr(
					"total_revenue + ?",
					reservation.ReservedAmount,
				),
				"updated_at": now,
			})
		if eventResult.Error != nil {
			return fmt.Errorf("increment event revenue: %w", eventResult.Error)
		}
		if eventResult.RowsAffected != 1 {
			return fmt.Errorf(
				"increment event revenue: event %s not found",
				reservation.EventID,
			)
		}

		paymentResult := tx.Model(&models.Payment{}).
			Where("attendee_id = ?", reservation.AttendeeID).
			Updates(map[string]interface{}{
				"status":     models.PaymentStatusSucceeded,
				"paid_at":    now,
				"updated_at": now,
			})
		if paymentResult.Error != nil {
			return fmt.Errorf("mark payment succeeded: %w", paymentResult.Error)
		}
		if paymentResult.RowsAffected != 1 {
			return fmt.Errorf(
				"mark payment succeeded: attendee %s payment row not found",
				reservation.AttendeeID,
			)
		}
		return nil
	})
}

func (s *PaymentService) attachProviderOrderID(
	receiptID string,
	orderID string,
) error {
	result := s.db.Model(&models.Attendee{}).
		Where(
			"payment_order_id = ? AND status = ?",
			receiptID,
			models.AttendeeStatusReserved,
		).
		Updates(map[string]interface{}{
			"payment_order_id": orderID,
			"updated_at":       time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("persist provider order id: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrReservationNotFound
	}

	return nil
}

func (s *PaymentService) releaseReservation(
	eventID uuid.UUID,
	orderID string,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Attendee{}).
			Where(
				"payment_order_id = ? AND status = ?",
				orderID,
				models.AttendeeStatusReserved,
			).
			Updates(map[string]interface{}{
				"status":      models.AttendeeStatusAbandoned,
				"reserved_at": nil,
				"updated_at":  time.Now().UTC(),
			})
		if result.Error != nil {
			return fmt.Errorf("abandon reservation: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return nil
		}

		eventResult := tx.Exec(`
			UPDATE events
			SET tickets_sold = GREATEST(tickets_sold - 1, 0),
			    updated_at = NOW()
			WHERE id = ?
		`, eventID)
		if eventResult.Error != nil {
			return fmt.Errorf("release inventory: %w", eventResult.Error)
		}
		if eventResult.RowsAffected == 0 {
			return ErrReservationNotFound
		}

		if err := s.markPaymentFailedByOrderID(tx, orderID); err != nil {
			return err
		}

		return nil
	})
}

func (s *PaymentService) tryAcquireSweeperLock(
	tx *gorm.DB,
) (bool, error) {
	var row advisoryLockRow

	if err := tx.Raw(
		"SELECT pg_try_advisory_xact_lock(?) AS acquired",
		s.sweeperLockKey,
	).Scan(&row).Error; err != nil {
		return false, fmt.Errorf("acquire sweeper advisory lock: %w", err)
	}

	return row.Acquired, nil
}

func (s *PaymentService) releaseExpiredReservations(
	tx *gorm.DB,
	cutoff time.Time,
) ([]releasedInventoryRow, error) {
	var rows []releasedInventoryRow

	err := tx.Raw(`
		WITH locked AS (
			SELECT id, event_id
			FROM attendees
			WHERE status = ?
			  AND reserved_at <= ?
			ORDER BY reserved_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT ?
		),
		updated AS (
			UPDATE attendees AS a
			SET status = ?,
			    reserved_at = NULL,
			    updated_at = NOW()
			FROM locked
			WHERE a.id = locked.id
			  AND a.status = ?
			RETURNING a.id, a.event_id
		),
		updated_payments AS (
			UPDATE payments AS p
			SET status = ?,
			    paid_at = NULL,
			    updated_at = NOW()
			FROM updated
			WHERE p.attendee_id = updated.id
		)
		SELECT event_id, COUNT(*) AS released
		FROM updated
		GROUP BY event_id
	`, models.AttendeeStatusReserved, cutoff, s.sweeperBatchSize, models.AttendeeStatusAbandoned, models.AttendeeStatusReserved, models.PaymentStatusFailed).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("release expired reservations: %w", err)
	}

	return rows, nil
}

func (s *PaymentService) loadCheckoutEventState(
	tx *gorm.DB,
	eventID uuid.UUID,
) (*checkoutEventState, error) {
	var state checkoutEventState

	err := tx.Model(&models.Event{}).
		Select(
			"id",
			"capacity",
			"price",
			"tickets_sold",
			"total_registered",
		).
		Where("id = ?", eventID).
		Limit(1).
		Take(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("load checkout event state: %w", err)
	}

	return &state, nil
}

func (s *PaymentService) findCheckoutAttendee(
	tx *gorm.DB,
	eventID uuid.UUID,
	email string,
) (*checkoutAttendeeRow, error) {
	var attendee checkoutAttendeeRow

	err := tx.Model(&models.Attendee{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "status").
		Where("event_id = ? AND email = ?", eventID, email).
		Limit(1).
		Take(&attendee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("load checkout attendee: %w", err)
	}

	return &attendee, nil
}

func (s *PaymentService) loadWebhookReservation(
	tx *gorm.DB,
	orderID string,
) (*webhookReservationRow, error) {
	var attendee struct {
		AttendeeID uuid.UUID             `gorm:"column:attendee_id"`
		EventID    uuid.UUID             `gorm:"column:event_id"`
		Status     models.AttendeeStatus `gorm:"column:status"`
	}

	err := tx.Model(&models.Attendee{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select(
			"attendees.id AS attendee_id",
			"attendees.event_id",
			"attendees.status",
		).
		Where("attendees.payment_order_id = ?", orderID).
		Limit(1).
		Take(&attendee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("load webhook reservation: %w", err)
	}

	reservation := &webhookReservationRow{
		AttendeeID: attendee.AttendeeID,
		EventID:    attendee.EventID,
		Status:     attendee.Status,
	}

	var payment struct {
		Amount float64 `gorm:"column:amount"`
	}

	err = tx.Model(&models.Payment{}).
		Select("amount").
		Where("attendee_id = ?", attendee.AttendeeID).
		Limit(1).
		Take(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return reservation, nil
		}

		return nil, fmt.Errorf("load webhook payment ledger: %w", err)
	}

	reservation.ReservedAmount = payment.Amount
	reservation.PaymentRecorded = true

	return reservation, nil
}

func (s *PaymentService) recordPendingPayment(
	orderID string,
	amount float64,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var attendee struct {
			ID uuid.UUID `gorm:"column:id"`
		}

		err := tx.Model(&models.Attendee{}).
			Select("id").
			Where(
				"payment_order_id = ? AND status = ?",
				orderID,
				models.AttendeeStatusReserved,
			).
			Limit(1).
			Take(&attendee).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReservationNotFound
			}
			return fmt.Errorf("load reserved attendee for payment ledger: %w", err)
		}

		payment := models.Payment{
			AttendeeID:           attendee.ID,
			HyperswitchPaymentID: orderID,
			Amount:               amount,
			Currency:             s.defaultCurrency,
			Status:               models.PaymentStatusPending,
			PaidAt:               nil,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "attendee_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"hyperswitch_payment_id": orderID,
				"amount":                 amount,
				"currency":               s.defaultCurrency,
				"status":                 models.PaymentStatusPending,
				"paid_at":                nil,
				"updated_at":             time.Now().UTC(),
			}),
		}).Create(&payment).Error; err != nil {
			return fmt.Errorf("persist payment ledger: %w", err)
		}

		return nil
	})
}

func (s *PaymentService) markPaymentFailedByOrderID(
	tx *gorm.DB,
	orderID string,
) error {
	result := tx.Model(&models.Payment{}).
		Where("hyperswitch_payment_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     models.PaymentStatusFailed,
			"paid_at":    nil,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("mark payment failed: %w", result.Error)
	}

	return nil
}

func extractWebhookOrderID(payload []byte) (string, error) {
	var envelope webhookOrderEnvelope

	if err := json.Unmarshal(payload, &envelope); err != nil {
		return "", fmt.Errorf("decode webhook payload: %w", err)
	}

	switch {
	case strings.TrimSpace(envelope.Payload.Payment.Entity.OrderID) != "":
		return strings.TrimSpace(envelope.Payload.Payment.Entity.OrderID), nil
	case strings.TrimSpace(envelope.Data.Object.OrderID) != "":
		return strings.TrimSpace(envelope.Data.Object.OrderID), nil
	case strings.TrimSpace(envelope.OrderID) != "":
		return strings.TrimSpace(envelope.OrderID), nil
	default:
		return "", errors.New("payment order id missing from webhook payload")
	}
}

func isDuplicateCheckoutErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") &&
		(strings.Contains(msg, "idx_attendees_event_email") ||
			(strings.Contains(msg, "event_id") &&
				strings.Contains(msg, "email")))
}

func canRetryCheckout(status models.AttendeeStatus) bool {
	switch status {
	case models.AttendeeStatusAbandoned, models.AttendeeStatusCancelled:
		return true
	default:
		return false
	}
}

func validateCapturedPaymentAmount(
	receivedAmount int64,
	eventPrice float64,
) error {
	if receivedAmount <= 0 {
		return nil
	}

	expectedAmount := priceToMinorUnits(eventPrice)
	if receivedAmount != expectedAmount {
		return fmt.Errorf(
			"%w: expected %d, got %d",
			ErrWebhookAmountMismatch,
			expectedAmount,
			receivedAmount,
		)
	}

	return nil
}

func priceToMinorUnits(price float64) int64 {
	return int64(math.Round(price * minorUnitsPerMajor))
}

func occupiedInventory(totalRegistered, ticketsSold int) int {
	if ticketsSold > totalRegistered {
		return ticketsSold
	}

	return totalRegistered
}

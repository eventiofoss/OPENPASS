package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrInvalidPaymentSignature = errors.New("invalid payment signature")
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
)

const (
	defaultRazorpayBaseURL     = "https://api.razorpay.com"
	defaultRazorpayHTTPTimeout = 10 * time.Second
	maxGatewayErrorBodyBytes   = 4096
)

// PaymentEvent is the normalized payment webhook payload used internally.
type PaymentEvent struct {
	ProviderPaymentID string
	OrderID           string
	EventType         string
	Status            string // "captured", "failed", "ignored"
	Amount            int64
	RawEventID        string // For idempotency
}

// PaymentProvider exposes the operations every gateway integration must support.
type PaymentProvider interface {
	CreateOrder(
		amount int64,
		currency string,
		receiptID string,
	) (orderID string, providerData map[string]interface{}, err error)
	// Used for the immediate frontend client callback.
	VerifyPaymentSignature(
		orderID,
		paymentID,
		signature string,
	) (bool, error)
	// Used for the server-to-server async webhook.
	VerifyWebhook(payload []byte, signatureHeader string) (*PaymentEvent, error)
}

type razorpayProvider struct {
	keyID         string
	keySecret     string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

type razorpayWebhookEnvelope struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID      string `json:"id"`
				OrderID string `json:"order_id"`
				Amount  int64  `json:"amount"`
			} `json:"entity"`
		} `json:"payment"`
		Order struct {
			Entity struct {
				ID         string `json:"id"`
				AmountPaid int64  `json:"amount_paid"`
			} `json:"entity"`
		} `json:"order"`
	} `json:"payload"`
}

type razorpayCreateOrderRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
}

type razorpayCreateOrderResponse struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
	Status   string `json:"status"`
}

func (p *razorpayProvider) CreateOrder(
	amount int64,
	currency string,
	receiptID string,
) (string, map[string]interface{}, error) {
	requestBody := razorpayCreateOrderRequest{
		Amount:   amount,
		Currency: currency,
		Receipt:  receiptID,
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("encode razorpay order request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		strings.TrimRight(p.baseURL, "/")+"/v1/orders",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", nil, fmt.Errorf("build razorpay order request: %w", err)
	}
	req.SetBasicAuth(p.keyID, p.keySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("create razorpay order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(
			io.LimitReader(resp.Body, maxGatewayErrorBodyBytes),
		)
		if readErr != nil {
			return "", nil, fmt.Errorf(
				"create razorpay order: status %d; read error response: %w",
				resp.StatusCode,
				readErr,
			)
		}

		errBody := strings.TrimSpace(string(body))
		if errBody == "" {
			return "", nil, fmt.Errorf(
				"create razorpay order: unexpected status %d",
				resp.StatusCode,
			)
		}

		return "", nil, fmt.Errorf(
			"create razorpay order: unexpected status %d: %s",
			resp.StatusCode,
			errBody,
		)
	}

	var order razorpayCreateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return "", nil, fmt.Errorf("decode razorpay order response: %w", err)
	}

	orderID := strings.TrimSpace(order.ID)
	if orderID == "" {
		return "", nil, errors.New("create razorpay order: missing order id")
	}

	return orderID, map[string]interface{}{
		"gateway":  "razorpay",
		"key_id":   p.keyID,
		"order_id": orderID,
		"amount":   order.Amount,
		"currency": strings.ToUpper(strings.TrimSpace(order.Currency)),
		"receipt":  strings.TrimSpace(order.Receipt),
	}, nil
}

func (p *razorpayProvider) VerifyPaymentSignature(
	orderID,
	paymentID,
	signature string,
) (bool, error) {
	expectedSignature := computeHMACSHA256Hex(
		p.keySecret,
		[]byte(orderID+"|"+paymentID),
	)

	if subtle.ConstantTimeCompare(
		[]byte(expectedSignature),
		[]byte(signature),
	) != 1 {
		return false, ErrInvalidPaymentSignature
	}

	return true, nil
}

func (p *razorpayProvider) VerifyWebhook(
	payload []byte,
	signatureHeader string,
) (*PaymentEvent, error) {
	expectedSignature := computeHMACSHA256Hex(p.webhookSecret, payload)

	if subtle.ConstantTimeCompare(
		[]byte(expectedSignature),
		[]byte(signatureHeader),
	) != 1 {
		return nil, ErrInvalidWebhookSignature
	}

	var envelope razorpayWebhookEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, fmt.Errorf("decode razorpay webhook payload: %w", err)
	}

	event := &PaymentEvent{
		ProviderPaymentID: strings.TrimSpace(
			envelope.Payload.Payment.Entity.ID,
		),
		OrderID:   strings.TrimSpace(razorpayWebhookOrderID(envelope)),
		EventType: strings.TrimSpace(envelope.Event),
		Amount:    razorpayWebhookAmount(envelope),
	}

	switch event.EventType {
	case "payment.captured", "order.paid":
		event.Status = capturedPaymentStatus
	case "payment.failed":
		event.Status = failedPaymentStatus
	default:
		event.Status = ignoredPaymentStatus
	}

	switch {
	case event.ProviderPaymentID != "":
		event.RawEventID = event.ProviderPaymentID
	case event.OrderID != "":
		event.RawEventID = event.OrderID
	default:
		event.RawEventID = event.EventType
	}

	return event, nil
}

type mockProvider struct{}

func (p *mockProvider) CreateOrder(
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

func (p *mockProvider) VerifyPaymentSignature(
	orderID,
	paymentID,
	signature string,
) (bool, error) {
	return true, nil
}

func (p *mockProvider) VerifyWebhook(
	payload []byte,
	signatureHeader string,
) (*PaymentEvent, error) {
	orderID, err := extractWebhookOrderID(payload)
	if err != nil {
		return nil, err
	}

	return &PaymentEvent{
		OrderID:    orderID,
		EventType:  "mock.payment.captured",
		Status:     capturedPaymentStatus,
		RawEventID: orderID,
	}, nil
}

// NewPaymentProvider returns the configured gateway adapter.
func NewPaymentProvider(gateway string) (PaymentProvider, error) {
	switch gateway {
	case "razorpay":
		keyID := os.Getenv("RAZORPAY_KEY_ID")
		keySecret := os.Getenv("RAZORPAY_SECRET")
		webhookSecret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
		baseURL := strings.TrimSpace(os.Getenv("RAZORPAY_BASE_URL"))

		if keyID == "" || keySecret == "" || webhookSecret == "" {
			return nil, errors.New("missing Razorpay configuration")
		}
		if baseURL == "" {
			baseURL = defaultRazorpayBaseURL
		}

		return &razorpayProvider{
			keyID:         keyID,
			keySecret:     keySecret,
			webhookSecret: webhookSecret,
			baseURL:       baseURL,
			httpClient: &http.Client{
				Timeout: defaultRazorpayHTTPTimeout,
			},
		}, nil
	case "mock":
		if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
			return nil, errors.New(
				"mock payment gateway is not allowed in production",
			)
		}
		return &mockProvider{}, nil
	default:
		return nil, errors.New("unsupported payment gateway")
	}
}

func computeHMACSHA256Hex(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func razorpayWebhookOrderID(envelope razorpayWebhookEnvelope) string {
	switch {
	case strings.TrimSpace(envelope.Payload.Payment.Entity.OrderID) != "":
		return strings.TrimSpace(envelope.Payload.Payment.Entity.OrderID)
	case strings.TrimSpace(envelope.Payload.Order.Entity.ID) != "":
		return strings.TrimSpace(envelope.Payload.Order.Entity.ID)
	default:
		return ""
	}
}

func razorpayWebhookAmount(envelope razorpayWebhookEnvelope) int64 {
	if envelope.Payload.Payment.Entity.Amount > 0 {
		return envelope.Payload.Payment.Entity.Amount
	}

	return envelope.Payload.Order.Entity.AmountPaid
}

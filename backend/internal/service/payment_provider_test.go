package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRazorpayProvider_CreateOrder(t *testing.T) {
	t.Parallel()

	provider := &razorpayProvider{
		keyID:         "rzp_test_key",
		keySecret:     "rzp_test_secret",
		webhookSecret: "whsec_test",
		baseURL:       "https://razorpay.test",
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Fatalf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/v1/orders" {
					t.Fatalf("expected /v1/orders, got %s", r.URL.Path)
				}

				username, password, ok := r.BasicAuth()
				if !ok {
					t.Fatal("expected basic auth")
				}
				if username != "rzp_test_key" || password != "rzp_test_secret" {
					t.Fatalf("unexpected basic auth credentials %q/%q", username, password)
				}

				var payload map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request payload: %v", err)
				}

				if payload["amount"] != float64(149950) {
					t.Fatalf("expected amount 149950, got %#v", payload["amount"])
				}
				if payload["currency"] != "INR" {
					t.Fatalf("expected currency INR, got %#v", payload["currency"])
				}
				if payload["receipt"] != "receipt_123" {
					t.Fatalf("expected receipt receipt_123, got %#v", payload["receipt"])
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(
						strings.NewReader(`{"id":"order_live_123","amount":149950,"currency":"INR","receipt":"receipt_123","status":"created"}`),
					),
				}, nil
			}),
		},
	}

	orderID, providerData, err := provider.CreateOrder(149950, "INR", "receipt_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if orderID != "order_live_123" {
		t.Fatalf("expected order_live_123, got %q", orderID)
	}
	if providerData["gateway"] != "razorpay" {
		t.Fatalf("expected razorpay gateway, got %#v", providerData["gateway"])
	}
	if providerData["order_id"] != "order_live_123" {
		t.Fatalf("expected order_id order_live_123, got %#v", providerData["order_id"])
	}
	if providerData["amount"] != int64(149950) {
		t.Fatalf("expected amount 149950, got %#v", providerData["amount"])
	}
}

func TestRazorpayProvider_CreateOrder_PropagatesGatewayFailure(t *testing.T) {
	t.Parallel()

	provider := &razorpayProvider{
		keyID:         "rzp_test_key",
		keySecret:     "rzp_test_secret",
		webhookSecret: "whsec_test",
		baseURL:       "https://razorpay.test",
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(
						strings.NewReader(`{"error":"bad request"}`),
					),
				}, nil
			}),
		},
	}

	_, _, err := provider.CreateOrder(149950, "INR", "receipt_123")
	if err == nil {
		t.Fatal("expected error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

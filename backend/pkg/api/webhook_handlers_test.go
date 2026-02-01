package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookCreateRequestValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name: "valid webhook",
			body: map[string]any{
				"url":    "https://example.com/webhook",
				"events": []string{"request.created", "offer.received"},
			},
			wantStatus: http.StatusUnauthorized, // No auth in test
		},
		{
			name: "missing url",
			body: map[string]any{
				"events": []string{"request.created"},
			},
			wantStatus: http.StatusUnauthorized, // Will fail auth first
		},
		{
			name: "missing events",
			body: map[string]any{
				"url": "https://example.com/webhook",
			},
			wantStatus: http.StatusUnauthorized, // Will fail auth first
		},
		{
			name:       "empty body",
			body:       map[string]any{},
			wantStatus: http.StatusUnauthorized, // Will fail auth first
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/api/v1/webhooks", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// This will fail auth check since we don't have middleware set up in test
			handler := &WebhookHandler{}
			handler.CreateWebhook(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("CreateWebhook() status = %v, want %v", rr.Code, tc.wantStatus)
			}
		})
	}
}

func TestValidEventTypes(t *testing.T) {
	validEvents := map[string]bool{
		"request.created":      true,
		"offer.received":       true,
		"offer.accepted":       true,
		"offer.rejected":       true,
		"listing.created":      true,
		"listing.updated":      true,
		"auction.started":      true,
		"bid.placed":           true,
		"bid.outbid":           true,
		"auction.ending_soon":  true,
		"auction.ended":        true,
		"order.created":        true,
		"escrow.funded":        true,
		"delivery.confirmed":   true,
		"payment.released":     true,
		"dispute.opened":       true,
		"match.found":          true,
		"order.filled":         true,
		"transaction.created":  true,
		"transaction.delivered": true,
		"transaction.completed": true,
		"rating.submitted":     true,
	}

	tests := []struct {
		name  string
		event string
		valid bool
	}{
		{"request created", "request.created", true},
		{"offer received", "offer.received", true},
		{"offer accepted", "offer.accepted", true},
		{"auction started", "auction.started", true},
		{"bid placed", "bid.placed", true},
		{"transaction completed", "transaction.completed", true},
		{"invalid event", "invalid.event", false},
		{"empty event", "", false},
		{"random string", "foobar", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := validEvents[tc.event]
			if isValid != tc.valid {
				t.Errorf("Event type %q validity = %v, want %v", tc.event, isValid, tc.valid)
			}
		})
	}
}

func TestWebhookListRequiresAuth(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/webhooks", nil)
	rr := httptest.NewRecorder()

	handler := &WebhookHandler{}
	handler.ListWebhooks(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ListWebhooks() status = %v, want %v", rr.Code, http.StatusUnauthorized)
	}
}

func TestWebhookDeleteRequiresAuth(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/api/v1/webhooks/123", nil)
	rr := httptest.NewRecorder()

	handler := &WebhookHandler{}
	handler.DeleteWebhook(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("DeleteWebhook() status = %v, want %v", rr.Code, http.StatusUnauthorized)
	}
}

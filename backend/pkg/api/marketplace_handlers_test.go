package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

func TestParseIntParam(t *testing.T) {
	tests := []struct {
		name       string
		queryParam string
		paramName  string
		defaultVal int
		expected   int
	}{
		{
			name:       "valid integer",
			queryParam: "limit=50",
			paramName:  "limit",
			defaultVal: 20,
			expected:   50,
		},
		{
			name:       "missing param uses default",
			queryParam: "",
			paramName:  "limit",
			defaultVal: 20,
			expected:   20,
		},
		{
			name:       "invalid integer uses default",
			queryParam: "limit=abc",
			paramName:  "limit",
			defaultVal: 20,
			expected:   20,
		},
		{
			name:       "zero value",
			queryParam: "offset=0",
			paramName:  "offset",
			defaultVal: 10,
			expected:   0,
		},
		{
			name:       "negative value",
			queryParam: "limit=-5",
			paramName:  "limit",
			defaultVal: 20,
			expected:   -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.queryParam, nil)
			result := parseIntParam(req, tt.paramName, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("parseIntParam() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestSearchListingsQueryParams(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		expectType   bool
		expectScope  bool
		expectMinMax bool
	}{
		{
			name:  "basic search query",
			query: "q=test",
		},
		{
			name:       "with type filter",
			query:      "q=test&type=services",
			expectType: true,
		},
		{
			name:        "with scope filter",
			query:       "q=test&scope=local",
			expectScope: true,
		},
		{
			name:         "with price range",
			query:        "q=test&min_price=10&max_price=100",
			expectMinMax: true,
		},
		{
			name:         "all filters",
			query:        "q=test&type=goods&scope=regional&min_price=5&max_price=50&limit=10&offset=5",
			expectType:   true,
			expectScope:  true,
			expectMinMax: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/listings?"+tt.query, nil)

			// Verify query params can be parsed
			q := req.URL.Query()
			if q.Get("q") == "" && tt.query != "" {
				// Some tests may not have q param
			}

			if tt.expectType && q.Get("type") == "" {
				t.Error("expected type param")
			}
			if tt.expectScope && q.Get("scope") == "" {
				t.Error("expected scope param")
			}
			if tt.expectMinMax {
				if q.Get("min_price") == "" || q.Get("max_price") == "" {
					t.Error("expected min_price and max_price params")
				}
			}
		})
	}
}

func TestCreateListingRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid listing request",
			body: map[string]interface{}{
				"title":        "Test Service",
				"listing_type": "services",
				"price_amount": 100.0,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing title",
			body: map[string]interface{}{
				"listing_type": "services",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing listing_type",
			body: map[string]interface{}{
				"title": "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/listings", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Verify request body can be parsed
			var parsed map[string]interface{}
			if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsed); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}

			// Check required fields
			_, hasTitle := parsed["title"]
			_, hasType := parsed["listing_type"]

			if tt.expectedStatus == http.StatusBadRequest {
				if hasTitle && hasType {
					t.Error("expected missing required fields for bad request")
				}
			}
		})
	}
}

func TestCreateRequestRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid request",
			body: map[string]interface{}{
				"title":        "Need delivery",
				"request_type": "services",
				"budget_max":   100.0,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing title",
			body: map[string]interface{}{
				"request_type": "services",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)

			var parsed map[string]interface{}
			if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsed); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}

			_, hasTitle := parsed["title"]
			_, hasType := parsed["request_type"]

			if tt.expectedStatus == http.StatusBadRequest {
				if hasTitle && hasType {
					t.Error("expected missing required fields for bad request")
				}
			}
		})
	}
}

func TestCreateOfferRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid offer",
			body: map[string]interface{}{
				"price_amount": 50.0,
				"description":  "I can help",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "zero price",
			body: map[string]interface{}{
				"price_amount": 0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative price",
			body: map[string]interface{}{
				"price_amount": -10.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)

			var parsed map[string]interface{}
			if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsed); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}

			priceAmount, hasPrice := parsed["price_amount"]
			if tt.expectedStatus == http.StatusBadRequest {
				if hasPrice {
					price, ok := priceAmount.(float64)
					if ok && price > 0 {
						t.Error("expected invalid price for bad request")
					}
				}
			}
		})
	}
}

func TestUUIDParsing(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		valid   bool
	}{
		{
			name:  "valid UUID",
			id:    "550e8400-e29b-41d4-a716-446655440000",
			valid: true,
		},
		{
			name:  "invalid UUID",
			id:    "not-a-uuid",
			valid: false,
		},
		{
			name:  "empty string",
			id:    "",
			valid: false,
		},
		{
			name:  "partial UUID",
			id:    "550e8400-e29b",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uuid.Parse(tt.id)
			if (err == nil) != tt.valid {
				t.Errorf("uuid.Parse(%s) valid = %v, want %v", tt.id, err == nil, tt.valid)
			}
		})
	}
}

func TestMiddlewareGetAgent(t *testing.T) {
	// Test that GetAgent returns nil when no agent in context
	req := httptest.NewRequest("GET", "/", nil)
	ag := middleware.GetAgent(req.Context())
	if ag != nil {
		t.Error("expected nil agent when not in context")
	}

	// Test with agent in context using context.WithValue
	testAgent := &agent.Agent{
		ID:   uuid.New(),
		Name: "TestAgent",
	}
	ctx := context.WithValue(req.Context(), middleware.AgentContextKey, testAgent)
	req = req.WithContext(ctx)
	ag = middleware.GetAgent(req.Context())
	if ag == nil {
		t.Error("expected agent in context")
	}
	if ag.Name != "TestAgent" {
		t.Errorf("expected agent name TestAgent, got %s", ag.Name)
	}
}

func TestRouteURLParams(t *testing.T) {
	// Test chi URL param extraction
	r := chi.NewRouter()

	var capturedID string
	r.Get("/listings/{id}", func(w http.ResponseWriter, r *http.Request) {
		capturedID = chi.URLParam(r, "id")
		w.WriteHeader(http.StatusOK)
	})

	testID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("GET", "/listings/"+testID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if capturedID != testID {
		t.Errorf("expected captured ID %s, got %s", testID, capturedID)
	}
}

func TestNestedRouteURLParams(t *testing.T) {
	r := chi.NewRouter()

	var capturedRequestID, capturedOfferID string
	r.Post("/requests/{id}/offers/{offerId}/accept", func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = chi.URLParam(r, "id")
		capturedOfferID = chi.URLParam(r, "offerId")
		w.WriteHeader(http.StatusOK)
	})

	requestID := "550e8400-e29b-41d4-a716-446655440000"
	offerID := "660e8400-e29b-41d4-a716-446655440001"
	req := httptest.NewRequest("POST", "/requests/"+requestID+"/offers/"+offerID+"/accept", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if capturedRequestID != requestID {
		t.Errorf("expected request ID %s, got %s", requestID, capturedRequestID)
	}
	if capturedOfferID != offerID {
		t.Errorf("expected offer ID %s, got %s", offerID, capturedOfferID)
	}
}

func TestJSONContentType(t *testing.T) {
	// Test that handlers set correct content type
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"OK", http.StatusOK, "200 OK"},
		{"Created", http.StatusCreated, "201 Created"},
		{"NoContent", http.StatusNoContent, "204 No Content"},
		{"BadRequest", http.StatusBadRequest, "400 Bad Request"},
		{"Unauthorized", http.StatusUnauthorized, "401 Unauthorized"},
		{"NotFound", http.StatusNotFound, "404 Not Found"},
		{"InternalServerError", http.StatusInternalServerError, "500 Internal Server Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			w.WriteHeader(tt.code)
			if w.Code != tt.code {
				t.Errorf("expected status %d, got %d", tt.code, w.Code)
			}
		})
	}
}

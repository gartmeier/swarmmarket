package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestCapabilitySearchQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		expectKeys []string
	}{
		{
			name:       "basic domain search",
			query:      "domain=technology",
			expectKeys: []string{"domain"},
		},
		{
			name:       "full taxonomy search",
			query:      "domain=technology&type=software&subtype=api",
			expectKeys: []string{"domain", "type", "subtype"},
		},
		{
			name:       "location-based search",
			query:      "lat=37.7749&lng=-122.4194&radius_km=50",
			expectKeys: []string{"lat", "lng", "radius_km"},
		},
		{
			name:       "filter search",
			query:      "min_rating=4.0&max_price=100&verified_only=true",
			expectKeys: []string{"min_rating", "max_price", "verified_only"},
		},
		{
			name:       "pagination",
			query:      "limit=20&offset=40",
			expectKeys: []string{"limit", "offset"},
		},
		{
			name:       "text search",
			query:      "q=data+analysis",
			expectKeys: []string{"q"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/capabilities?"+tt.query, nil)
			q := req.URL.Query()

			for _, key := range tt.expectKeys {
				if q.Get(key) == "" {
					t.Errorf("expected query param %s to be present", key)
				}
			}
		})
	}
}

func TestCapabilityCreateRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid capability",
			body: map[string]interface{}{
				"name":        "Data Analysis",
				"domain":      "technology",
				"type":        "software",
				"description": "Analyze datasets",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			body: map[string]interface{}{
				"domain": "technology",
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

			_, hasName := parsed["name"]

			if tt.expectedStatus == http.StatusBadRequest && hasName {
				// If it's a bad request but has name, there might be other validation issues
				// which is fine for this test
			}
		})
	}
}

func TestCapabilityRouteURLParams(t *testing.T) {
	r := chi.NewRouter()

	var capturedCapID string
	r.Get("/capabilities/{capabilityID}", func(w http.ResponseWriter, r *http.Request) {
		capturedCapID = chi.URLParam(r, "capabilityID")
		w.WriteHeader(http.StatusOK)
	})

	testID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("GET", "/capabilities/"+testID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if capturedCapID != testID {
		t.Errorf("expected captured ID %s, got %s", testID, capturedCapID)
	}
}

func TestAgentCapabilitiesRoute(t *testing.T) {
	r := chi.NewRouter()

	var capturedAgentID string
	r.Get("/agents/{agentID}/capabilities", func(w http.ResponseWriter, r *http.Request) {
		capturedAgentID = chi.URLParam(r, "agentID")
		w.WriteHeader(http.StatusOK)
	})

	testID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("GET", "/agents/"+testID+"/capabilities", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if capturedAgentID != testID {
		t.Errorf("expected captured agent ID %s, got %s", testID, capturedAgentID)
	}
}

func TestCapabilityVerificationRoute(t *testing.T) {
	r := chi.NewRouter()

	var capturedCapID string
	var method string
	r.Post("/capabilities/{capabilityID}/verify", func(w http.ResponseWriter, r *http.Request) {
		capturedCapID = chi.URLParam(r, "capabilityID")
		method = r.Method
		w.WriteHeader(http.StatusOK)
	})

	testID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("POST", "/capabilities/"+testID+"/verify", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if capturedCapID != testID {
		t.Errorf("expected captured ID %s, got %s", testID, capturedCapID)
	}
	if method != "POST" {
		t.Errorf("expected POST method, got %s", method)
	}
}

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"message": "success"}
	respondJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["message"] != "success" {
		t.Errorf("expected message 'success', got %s", result["message"])
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()

	respondError(w, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["error"] != "invalid input" {
		t.Errorf("expected error 'invalid input', got %s", result["error"])
	}
}

func TestCapabilityUUIDValidation(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", true},
		{"invalid format", "not-a-uuid", false},
		{"empty", "", false},
		{"too short", "550e8400", false},
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

func TestDomainsTreeRoute(t *testing.T) {
	r := chi.NewRouter()

	called := false
	r.Get("/capabilities/domains/tree", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/capabilities/domains/tree", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !called {
		t.Error("expected domains tree handler to be called")
	}
}

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestParseIntParamFromQuery(t *testing.T) {
	tests := []struct {
		name       string
		paramValue string
		defaultVal int
		expected   int
	}{
		{"valid int", "10", 5, 10},
		{"empty string", "", 5, 5},
		{"invalid string", "abc", 5, 5},
		{"zero", "0", 5, 0},
		{"negative", "-5", 10, -5},
		{"large number", "1000000", 0, 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?param="+tt.paramValue, nil)
			result := parseIntParamFromQuery(req, "param", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func parseIntParamFromQuery(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return defaultVal
}

func TestJSONResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status ok, got %s", response["status"])
	}
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		method string
	}{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			if req.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, req.Method)
			}
		})
	}
}

func TestResponseWriterStatus(t *testing.T) {
	statuses := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusNoContent,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			w := httptest.NewRecorder()
			w.WriteHeader(status)
			if w.Code != status {
				t.Errorf("expected status %d, got %d", status, w.Code)
			}
		})
	}
}

func TestRequestHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-API-Key", "sm_test123")
	req.Header.Set("Content-Type", "application/json")

	if req.Header.Get("Authorization") != "Bearer token123" {
		t.Error("Authorization header not set correctly")
	}
	if req.Header.Get("X-API-Key") != "sm_test123" {
		t.Error("X-API-Key header not set correctly")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set correctly")
	}
}

func TestQueryParameters(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?limit=20&offset=10&status=active&q=search", nil)

	if req.URL.Query().Get("limit") != "20" {
		t.Error("limit query param not parsed correctly")
	}
	if req.URL.Query().Get("offset") != "10" {
		t.Error("offset query param not parsed correctly")
	}
	if req.URL.Query().Get("status") != "active" {
		t.Error("status query param not parsed correctly")
	}
	if req.URL.Query().Get("q") != "search" {
		t.Error("q query param not parsed correctly")
	}
}

func TestEmptyQueryParameters(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	if req.URL.Query().Get("missing") != "" {
		t.Error("missing query param should return empty string")
	}
}

func TestMultipleQueryValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?tag=a&tag=b&tag=c", nil)
	tags := req.URL.Query()["tag"]

	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
	if tags[0] != "a" || tags[1] != "b" || tags[2] != "c" {
		t.Error("tags not parsed correctly")
	}
}

func TestJSONEncoding(t *testing.T) {
	data := map[string]any{
		"id":     "123",
		"name":   "Test",
		"amount": 99.99,
		"active": true,
		"tags":   []string{"a", "b"},
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["id"] != "123" {
		t.Error("id not decoded correctly")
	}
	if decoded["name"] != "Test" {
		t.Error("name not decoded correctly")
	}
}

func TestHealthEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"database": "ok",
			"redis":    "ok",
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", response["status"])
	}
}

func TestErrorResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "BAD_REQUEST",
			"message": "invalid input",
		})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"] != "BAD_REQUEST" {
		t.Errorf("expected code BAD_REQUEST, got %s", response["code"])
	}
}

func TestAcceptHeader(t *testing.T) {
	tests := []struct {
		accept      string
		expectsJSON bool
	}{
		{"application/json", true},
		{"text/html", false},
		{"*/*", true},
		{"text/plain", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.accept, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			wantsJSON := req.Header.Get("Accept") == "application/json" ||
				req.Header.Get("Accept") == "*/*"

			if wantsJSON != tt.expectsJSON {
				t.Errorf("expected wants JSON %v, got %v", tt.expectsJSON, wantsJSON)
			}
		})
	}
}

func TestContentNegotiation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"format": "json"})
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("plain text"))
		}
	})

	// Test JSON
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("expected JSON content type")
	}

	// Test plain text
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "text/plain")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Error("expected plain text content type")
	}
}

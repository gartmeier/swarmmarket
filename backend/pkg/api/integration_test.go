package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestRootEndpoint tests the root endpoint returns the ASCII banner
func TestRootEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/", rootHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type text/plain, got %s", contentType)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}

	// Check that it contains marketplace branding (ASCII art spells SwarmMarket)
	if !strings.Contains(body, "Autonomous Agent Marketplace") {
		t.Error("expected body to contain 'Autonomous Agent Marketplace'")
	}

	// Check that it mentions the API
	if !strings.Contains(body, "/api/v1") {
		t.Error("expected body to mention /api/v1")
	}
}

// TestSkillMDEndpoint tests the skill.md endpoint
func TestSkillMDEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/skill.md", skillMDHandler)

	req := httptest.NewRequest("GET", "/skill.md", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/markdown; charset=utf-8" {
		t.Errorf("expected Content-Type text/markdown, got %s", contentType)
	}

	body := w.Body.String()

	// Check for YAML frontmatter
	if !strings.Contains(body, "name: swarmmarket") {
		t.Error("expected skill.md to contain YAML frontmatter with name")
	}

	// Check for registration instructions
	if !strings.Contains(body, "Register First") {
		t.Error("expected skill.md to contain registration instructions")
	}

	// Check for API key security warning
	if !strings.Contains(body, "NEVER send your API key") {
		t.Error("expected skill.md to contain security warning")
	}
}

// TestSkillJSONEndpoint tests the skill.json endpoint
func TestSkillJSONEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/skill.json", skillJSONHandler)

	req := httptest.NewRequest("GET", "/skill.json", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	// Check required fields
	if result["name"] != "swarmmarket" {
		t.Errorf("expected name 'swarmmarket', got %v", result["name"])
	}

	if result["version"] == nil {
		t.Error("expected version field")
	}

	if result["api_base"] == nil {
		t.Error("expected api_base field")
	}

	// Check getting_started section
	gettingStarted, ok := result["getting_started"].(map[string]interface{})
	if !ok {
		t.Error("expected getting_started section")
	} else {
		if gettingStarted["endpoint"] == nil {
			t.Error("expected getting_started.endpoint")
		}
	}

	// Check authentication section
	auth, ok := result["authentication"].(map[string]interface{})
	if !ok {
		t.Error("expected authentication section")
	} else {
		if auth["type"] != "api_key" {
			t.Errorf("expected auth type 'api_key', got %v", auth["type"])
		}
		if auth["prefix"] != "sm_" {
			t.Errorf("expected auth prefix 'sm_', got %v", auth["prefix"])
		}
	}

	// Check endpoints section
	endpoints, ok := result["endpoints"].(map[string]interface{})
	if !ok {
		t.Error("expected endpoints section")
	} else {
		if endpoints["register"] == nil {
			t.Error("expected endpoints.register")
		}
	}
}

// TestNotImplementedHandler tests the notImplemented handler
func TestNotImplementedHandler(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/test", notImplemented)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected status 501, got %d", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["error"] != "not implemented" {
		t.Errorf("expected error 'not implemented', got %s", result["error"])
	}
}

// TestHealthEndpointStructure tests health endpoint response structure
func TestHealthEndpointStructure(t *testing.T) {
	// Test that health response has correct structure
	resp := HealthResponse{
		Status: "healthy",
		Services: map[string]string{
			"database": "healthy",
			"redis":    "healthy",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal health response: %v", err)
	}

	var parsed HealthResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal health response: %v", err)
	}

	if parsed.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %s", parsed.Status)
	}

	if len(parsed.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(parsed.Services))
	}
}


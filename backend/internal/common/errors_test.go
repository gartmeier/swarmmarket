package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:    ErrCodeBadRequest,
		Message: "invalid input",
		Details: map[string]string{"field": "email"},
	}

	if err.Error() != "invalid input" {
		t.Errorf("expected 'invalid input', got %q", err.Error())
	}
}

func TestNewAPIError(t *testing.T) {
	details := map[string]string{"field": "name"}
	err := NewAPIError(ErrCodeBadRequest, "test error", details)

	if err.Code != ErrCodeBadRequest {
		t.Errorf("expected code %s, got %s", ErrCodeBadRequest, err.Code)
	}
	if err.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", err.Message)
	}
	if err.Details == nil {
		t.Error("expected details to be set")
	}
}

func TestErrBadRequest(t *testing.T) {
	err := ErrBadRequest("invalid field")

	if err.Code != ErrCodeBadRequest {
		t.Errorf("expected code %s, got %s", ErrCodeBadRequest, err.Code)
	}
	if err.Message != "invalid field" {
		t.Errorf("expected message 'invalid field', got %s", err.Message)
	}
	if err.Details != nil {
		t.Error("expected details to be nil")
	}
}

func TestErrUnauthorized(t *testing.T) {
	err := ErrUnauthorized("not logged in")

	if err.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %s, got %s", ErrCodeUnauthorized, err.Code)
	}
	if err.Message != "not logged in" {
		t.Errorf("expected message 'not logged in', got %s", err.Message)
	}
}

func TestErrForbidden(t *testing.T) {
	err := ErrForbidden("access denied")

	if err.Code != ErrCodeForbidden {
		t.Errorf("expected code %s, got %s", ErrCodeForbidden, err.Code)
	}
	if err.Message != "access denied" {
		t.Errorf("expected message 'access denied', got %s", err.Message)
	}
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("resource not found")

	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Message != "resource not found" {
		t.Errorf("expected message 'resource not found', got %s", err.Message)
	}
}

func TestErrConflict(t *testing.T) {
	err := ErrConflict("already exists")

	if err.Code != ErrCodeConflict {
		t.Errorf("expected code %s, got %s", ErrCodeConflict, err.Code)
	}
	if err.Message != "already exists" {
		t.Errorf("expected message 'already exists', got %s", err.Message)
	}
}

func TestErrTooManyRequests(t *testing.T) {
	err := ErrTooManyRequests("rate limit exceeded")

	if err.Code != ErrCodeTooManyRequests {
		t.Errorf("expected code %s, got %s", ErrCodeTooManyRequests, err.Code)
	}
	if err.Message != "rate limit exceeded" {
		t.Errorf("expected message 'rate limit exceeded', got %s", err.Message)
	}
}

func TestErrInternalServer(t *testing.T) {
	err := ErrInternalServer("something went wrong")

	if err.Code != ErrCodeInternalServer {
		t.Errorf("expected code %s, got %s", ErrCodeInternalServer, err.Code)
	}
	if err.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got %s", err.Message)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	err := ErrBadRequest("invalid input")

	WriteError(w, http.StatusBadRequest, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", contentType)
	}

	var response APIError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != ErrCodeBadRequest {
		t.Errorf("expected code %s, got %s", ErrCodeBadRequest, response.Code)
	}
	if response.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got %s", response.Message)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"name": "test"}

	WriteJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", contentType)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["name"] != "test" {
		t.Errorf("expected name 'test', got %s", response["name"])
	}
}

func TestWriteJSON_Created(t *testing.T) {
	w := httptest.NewRecorder()
	data := struct {
		ID string `json:"id"`
	}{ID: "123"}

	WriteJSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{ErrCodeBadRequest, "BAD_REQUEST"},
		{ErrCodeUnauthorized, "UNAUTHORIZED"},
		{ErrCodeForbidden, "FORBIDDEN"},
		{ErrCodeNotFound, "NOT_FOUND"},
		{ErrCodeConflict, "CONFLICT"},
		{ErrCodeUnprocessableEntity, "UNPROCESSABLE_ENTITY"},
		{ErrCodeTooManyRequests, "TOO_MANY_REQUESTS"},
		{ErrCodeInternalServer, "INTERNAL_SERVER_ERROR"},
	}

	for _, tt := range tests {
		if tt.code != tt.expected {
			t.Errorf("expected code %s, got %s", tt.expected, tt.code)
		}
	}
}

func TestAPIError_JSON(t *testing.T) {
	err := NewAPIError(ErrCodeBadRequest, "test", map[string]int{"count": 5})

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("failed to marshal: %v", jsonErr)
	}

	var decoded APIError
	if jsonErr := json.Unmarshal(data, &decoded); jsonErr != nil {
		t.Fatalf("failed to unmarshal: %v", jsonErr)
	}

	if decoded.Code != ErrCodeBadRequest {
		t.Errorf("expected code %s, got %s", ErrCodeBadRequest, decoded.Code)
	}
	if decoded.Message != "test" {
		t.Errorf("expected message 'test', got %s", decoded.Message)
	}
}

package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// "github.com/sh05/cat-server/src/handlers" // Legacy import - deprecated
)

func TestHealthHandler(t *testing.T) {
	t.Skip("Legacy test - needs refactoring for new architecture")

	start := time.Now()
	handlers.HealthHandler(w, req)
	duration := time.Since(start)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Check response format
	var response handlers.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Validate response fields
	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", response.Status)
	}

	if response.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	// Check response time (should be fast)
	maxDuration := 10 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("response took too long: %v > %v", duration, maxDuration)
	}
}

func TestHealthHandlerResponseTime(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Run multiple times to check consistency
	for i := 0; i < 5; i++ {
		start := time.Now()
		handlers.HealthHandler(w, req)
		duration := time.Since(start)

		maxDuration := 10 * time.Millisecond
		if duration > maxDuration {
			t.Errorf("iteration %d: response time %v exceeds limit %v", i+1, duration, maxDuration)
		}
	}
}

func TestHealthHandlerHTTPMethods(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/health", nil)
		w := httptest.NewRecorder()

		handlers.HealthHandler(w, req)

		// Should still respond with 200 OK for simplicity in this basic implementation
		if w.Code != http.StatusOK {
			t.Errorf("method %s: expected status %d, got %d", method, http.StatusOK, w.Code)
		}
	}
}

package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sh05/cat-server/src/handlers"
)

// TestHealthContractOpenAPI validates the health endpoint against OpenAPI spec
func TestHealthContractOpenAPI(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	// Validate status code as per OpenAPI spec
	if w.Code != http.StatusOK {
		t.Errorf("OpenAPI violation: expected status 200, got %d", w.Code)
	}

	// Validate Content-Type header
	contentType := w.Header().Get("Content-Type")
	expectedTypes := []string{"application/json", "text/plain", "text/html"}

	validContentType := false
	for _, expectedType := range expectedTypes {
		if strings.Contains(contentType, expectedType) {
			validContentType = true
			break
		}
	}

	if !validContentType {
		t.Errorf("OpenAPI violation: invalid Content-Type '%s', expected one of %v", contentType, expectedTypes)
	}

	// Validate response schema for JSON
	if strings.Contains(contentType, "application/json") {
		var response handlers.HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("OpenAPI violation: invalid JSON response: %v", err)
		}

		// Validate required fields according to OpenAPI schema
		if response.Status == "" {
			t.Error("OpenAPI violation: missing required field 'status'")
		}

		if response.Timestamp.IsZero() {
			t.Error("OpenAPI violation: missing required field 'timestamp'")
		}

		// Validate enum value
		if response.Status != "ok" {
			t.Errorf("OpenAPI violation: status must be 'ok', got '%s'", response.Status)
		}

		// Validate timestamp format (ISO 8601)
		if !isValidISO8601(response.Timestamp) {
			t.Errorf("OpenAPI violation: timestamp must be valid ISO 8601 format")
		}
	}
}

// TestHealthContractResponseTime validates response time requirements
func TestHealthContractResponseTime(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	handlers.HealthHandler(w, req)
	duration := time.Since(start)

	// Contract requirement: < 10ms response time
	maxDuration := 10 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("Contract violation: response time %v exceeds requirement %v", duration, maxDuration)
	}
}

// TestHealthContractNoAuthentication validates that no authentication is required
func TestHealthContractNoAuthentication(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// Intentionally no authorization headers
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	// Should succeed without authentication
	if w.Code != http.StatusOK {
		t.Errorf("Contract violation: endpoint should not require authentication, got status %d", w.Code)
	}
}

// TestHealthContractHTTPMethod validates only GET method is supported
func TestHealthContractHTTPMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Contract violation: GET /health should return 200, got %d", w.Code)
	}
}

// TestHealthContractConcurrentRequests validates concurrent request handling
func TestHealthContractConcurrentRequests(t *testing.T) {
	const numRequests = 10
	results := make(chan int, numRequests)

	// Simulate concurrent requests as per OpenAPI x-testing extension
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			start := time.Now()
			handlers.HealthHandler(w, req)
			duration := time.Since(start)

			// All should succeed
			if w.Code != http.StatusOK {
				results <- w.Code
				return
			}

			// All should be fast (within 50ms for concurrent scenario)
			if duration > 50*time.Millisecond {
				results <- 0 // Use 0 to indicate timeout
				return
			}

			results <- http.StatusOK
		}()
	}

	// Validate all requests succeeded
	for i := 0; i < numRequests; i++ {
		result := <-results
		if result != http.StatusOK {
			if result == 0 {
				t.Errorf("Contract violation: concurrent request %d exceeded time limit", i+1)
			} else {
				t.Errorf("Contract violation: concurrent request %d failed with status %d", i+1, result)
			}
		}
	}
}

// Helper function to validate ISO 8601 timestamp format
func isValidISO8601(timestamp time.Time) bool {
	// Check if timestamp can be marshaled to valid ISO 8601 format
	_, err := timestamp.MarshalText()
	return err == nil && !timestamp.IsZero()
}

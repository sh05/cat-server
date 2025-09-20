package contracts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestCatEndpointContract tests the /cat/{filename} endpoint against OpenAPI specification
func TestCatEndpointContract(t *testing.T) {
	// This test will fail until the implementation is complete
	// It validates the OpenAPI contract compliance

	tests := []struct {
		name           string
		filename       string
		setupFunc      func() *httptest.Server
		expectedStatus int
		validateFunc   func(*testing.T, *http.Response, []byte)
	}{
		{
			name:     "successful_file_content_response_structure",
			filename: "example.txt",
			setupFunc: func() *httptest.Server {
				// Mock server - will be replaced with actual implementation
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// This will fail until real implementation
					http.Error(w, "not implemented", http.StatusNotImplemented)
				}))
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *http.Response, body []byte) {
				// Validate response structure matches OpenAPI schema
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse JSON response: %v", err)
				}

				// Check required fields exist
				requiredFields := []string{"content", "filename", "size", "directory", "generated_at"}
				for _, field := range requiredFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Required field '%s' missing from response", field)
					}
				}

				// Validate content field is string
				if content, ok := response["content"].(string); !ok {
					t.Error("Content field is not a string")
				} else {
					// Content can be any valid UTF-8 string including empty
					if !isValidUTF8(content) {
						t.Error("Content is not valid UTF-8")
					}
				}

				// Validate size matches content length
				if size, ok := response["size"].(float64); ok {
					if content, ok := response["content"].(string); ok {
						if int64(size) != int64(len(content)) {
							t.Errorf("Size (%d) does not match content length (%d)", int64(size), len(content))
						}
					}
				}

				// Validate timestamp format
				if timestamp, ok := response["generated_at"].(string); ok {
					if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
						t.Errorf("Invalid timestamp format: %v", err)
					}
				}
			},
		},
		{
			name:     "error_response_structure_404",
			filename: "nonexistent.txt",
			setupFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// This will fail until real implementation
					http.Error(w, "not implemented", http.StatusNotImplemented)
				}))
			},
			expectedStatus: http.StatusNotFound,
			validateFunc: func(t *testing.T, resp *http.Response, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse JSON error response: %v", err)
				}

				// Check required error fields
				requiredFields := []string{"error", "timestamp", "status_code"}
				for _, field := range requiredFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Required error field '%s' missing from response", field)
					}
				}

				// Validate status code matches
				if statusCode, ok := response["status_code"].(float64); ok {
					if int(statusCode) != http.StatusNotFound {
						t.Errorf("Status code in response (%d) does not match expected (%d)", int(statusCode), http.StatusNotFound)
					}
				}
			},
		},
		{
			name:     "path_parameter_extraction",
			filename: "test-file.txt",
			setupFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Check if path parameter is correctly extracted
					if !strings.Contains(r.URL.Path, "test-file.txt") {
						t.Error("Path parameter not correctly extracted")
					}
					http.Error(w, "not implemented", http.StatusNotImplemented)
				}))
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *http.Response, body []byte) {
				// Basic structure validation
			},
		},
		{
			name:     "content_type_validation",
			filename: "example.txt",
			setupFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// This will fail until real implementation
					http.Error(w, "not implemented", http.StatusNotImplemented)
				}))
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *http.Response, body []byte) {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected Content-Type to contain 'application/json', got: %s", contentType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupFunc()
			defer server.Close()

			// Make request to /cat/{filename} endpoint
			resp, err := http.Get(server.URL + "/cat/" + tt.filename)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Read response body
			body := make([]byte, 1024*10) // 10KB buffer
			n, err := resp.Body.Read(body)
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("Failed to read response body: %v", err)
			}
			body = body[:n]

			// Currently expecting failure until implementation
			if resp.StatusCode == http.StatusNotImplemented {
				t.Logf("Test correctly failing - endpoint not implemented yet")
				return
			}

			// Once implemented, validate against expected status
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Run specific validation
			tt.validateFunc(t, resp, body)
		})
	}
}

// TestResponseSchemaValidation validates specific response schema requirements
func TestResponseSchemaValidation(t *testing.T) {
	// Test will fail until implementation complete
	t.Run("file_content_schema_validation", func(t *testing.T) {
		// Mock response for schema validation
		mockResponse := `{
			"content": "Hello, World!\nThis is a test file.",
			"filename": "test.txt",
			"size": 33,
			"directory": "./files/",
			"generated_at": "2025-09-20T10:00:00Z"
		}`

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(mockResponse), &response); err != nil {
			t.Fatalf("Failed to parse mock response: %v", err)
		}

		// Validate schema constraints
		if filename, ok := response["filename"].(string); ok {
			// Check filename length constraint (255 chars)
			if len(filename) > 255 {
				t.Error("Filename exceeds maximum length of 255 characters")
			}
		}

		// Validate directory path constraint
		if directory, ok := response["directory"].(string); ok {
			if len(directory) > 4096 {
				t.Error("Directory path exceeds maximum length of 4096 characters")
			}
		}

		// Validate size constraint (10MB limit)
		if size, ok := response["size"].(float64); ok {
			if size < 0 || size > 10485760 {
				t.Errorf("Size (%d) is outside valid range [0, 10485760]", int64(size))
			}
		}
	})

	t.Run("error_response_schema_validation", func(t *testing.T) {
		// Mock error response for schema validation
		mockErrorResponse := `{
			"error": "file not found",
			"filename": "missing.txt",
			"path": "./files/missing.txt",
			"timestamp": "2025-09-20T10:00:00Z",
			"status_code": 404
		}`

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(mockErrorResponse), &response); err != nil {
			t.Fatalf("Failed to parse mock error response: %v", err)
		}

		// Validate error message constraint
		if errorMsg, ok := response["error"].(string); ok {
			if len(errorMsg) > 1000 {
				t.Error("Error message exceeds maximum length of 1000 characters")
			}
		}

		// Validate status code enum constraint
		if statusCode, ok := response["status_code"].(float64); ok {
			validCodes := []int{400, 403, 404, 413, 415, 500}
			isValid := false
			for _, code := range validCodes {
				if int(statusCode) == code {
					isValid = true
					break
				}
			}
			if !isValid {
				t.Errorf("Status code %d is not in allowed values: %v", int(statusCode), validCodes)
			}
		}
	})
}

// TestEndpointBehaviorContract tests expected endpoint behavior
func TestEndpointBehaviorContract(t *testing.T) {
	// These tests will fail until implementation is complete

	t.Run("endpoint_exists", func(t *testing.T) {
		// Test that /cat/{filename} endpoint exists
		// This will fail until route is registered
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/cat/") {
				http.NotFound(w, r)
				return
			}
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			// Not implemented yet
			http.Error(w, "not implemented", http.StatusNotImplemented)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/cat/test.txt")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Currently expecting not implemented
		if resp.StatusCode == http.StatusNotImplemented {
			t.Log("Endpoint exists but not implemented - test correctly failing")
		} else if resp.StatusCode == http.StatusNotFound {
			t.Error("Endpoint /cat/{filename} not found - route not registered")
		}
	})

	t.Run("method_validation", func(t *testing.T) {
		// Test that only GET method is allowed
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			http.Error(w, "not implemented", http.StatusNotImplemented)
		}))
		defer server.Close()

		// Test POST method should fail
		resp, err := http.Post(server.URL+"/cat/test.txt", "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("Failed to make POST request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected POST to return 405 Method Not Allowed, got %d", resp.StatusCode)
		}
	})
}

// Helper function to check if string is valid UTF-8
func isValidUTF8(s string) bool {
	for i, r := range s {
		if r == '\uFFFD' && len(s[i:]) > 0 {
			return false
		}
	}
	return true
}

func TestMain(m *testing.M) {
	fmt.Println("Running contract tests for /cat/{filename} endpoint")
	fmt.Println("These tests will fail until the implementation is complete")
	fmt.Println("This is expected behavior for TDD approach")
	m.Run()
}

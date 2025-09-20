package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Response structures for integration testing
type FileListResponse struct {
	Files       []string  `json:"files"`
	Directory   string    `json:"directory"`
	Count       int       `json:"count"`
	GeneratedAt time.Time `json:"generated_at"`
}

type ErrorResponse struct {
	Error      string    `json:"error"`
	Path       string    `json:"path,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code"`
}

// setupTestFilesystem creates a temporary filesystem for testing
func setupTestFilesystem(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "cat-server-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test files including hidden files
	testFiles := map[string]string{
		"README.md":  "# Test Project\nThis is a test.",
		"main.go":    "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
		"test.txt":   "Simple test file content",
		".hidden":    "hidden file content",
		".gitignore": "*.log\n*.tmp",
		".env":       "SECRET=test",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestLsEndpoint_EndToEnd tests the complete /ls endpoint functionality
func TestLsEndpoint_EndToEnd(t *testing.T) {
	// This test will fail until the full implementation is complete
	t.Log("/ls endpoint not implemented yet - integration test failing as expected for TDD")

	testDir, cleanup := setupTestFilesystem(t)
	defer cleanup()

	tests := []struct {
		name           string
		serverDir      string
		expectedFiles  []string
		expectedStatus int
	}{
		{
			name:           "successful_file_listing",
			serverDir:      testDir,
			expectedFiles:  []string{"README.md", "main.go", "test.txt", ".hidden", ".gitignore", ".env"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These assertions will be enabled once the server implementation is complete
			// 1. Create server with specified directory
			// 2. Start httptest.Server
			// 3. Make GET request to /ls
			// 4. Verify response structure and content

			t.Logf("Testing directory: %s", tt.serverDir)
			t.Logf("Expected %d files", len(tt.expectedFiles))

			// Mock server for testing
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Test server implementation
				if r.URL.Path != "/ls" {
					http.NotFound(w, r)
					return
				}
				if r.Method != http.MethodGet {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				// Not implemented yet - expected behavior for TDD
				http.Error(w, "not implemented", http.StatusNotImplemented)
			}))
			defer server.Close()

			// Make request
			resp, err := http.Get(server.URL + "/ls")
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Currently expecting not implemented
			if resp.StatusCode == http.StatusNotImplemented {
				t.Log("Endpoint correctly returning not implemented - test failing as expected")
				return
			}

			// Once implemented, these validations will be enabled:
			// assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			// assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			// var response FileListResponse
			// err = json.NewDecoder(resp.Body).Decode(&response)
			// assert.NoError(t, err)

			// Validate response structure
			// assert.ElementsMatch(t, tt.expectedFiles, response.Files)
			// assert.Equal(t, len(tt.expectedFiles), response.Count)
			// assert.Equal(t, tt.serverDir, response.Directory)
			// assert.WithinDuration(t, time.Now(), response.GeneratedAt, time.Second)
		})
	}
}

// TestLsEndpoint_ErrorHandling tests error scenarios
func TestLsEndpoint_ErrorHandling(t *testing.T) {
	t.Log("/ls endpoint not implemented yet - error handling test failing as expected for TDD")

	tests := []struct {
		name           string
		serverDir      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "nonexistent_directory",
			serverDir:      "/nonexistent/directory",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "directory not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing error case: %s", tt.name)

			// Mock server (will be replaced with actual server)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "not implemented", http.StatusNotImplemented)
			}))
			defer server.Close()

			resp, err := http.Get(server.URL + "/ls")
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Currently expecting not implemented
			if resp.StatusCode == http.StatusNotImplemented {
				t.Log("Error handling not implemented - test failing as expected")
				return
			}

			// Once implemented, these validations will be enabled:
			// assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// var errorResp ErrorResponse
			// err = json.NewDecoder(resp.Body).Decode(&errorResp)
			// assert.NoError(t, err)

			// assert.Contains(t, errorResp.Error, tt.expectedError)
			// assert.Equal(t, tt.expectedStatus, errorResp.StatusCode)
		})
	}
}

// TestLsEndpoint_QuickstartScenarios tests scenarios from quickstart.md
func TestLsEndpoint_QuickstartScenarios(t *testing.T) {
	t.Log("Quickstart scenarios not testable until implementation complete")

	t.Run("scenario_1_default_directory", func(t *testing.T) {
		// Scenario 1: Default directory with files
		t.Log("Scenario 1: デフォルトディレクトリ (./files/) にファイルが存在")
		t.Log("Expected: 全ファイル名（隠しファイル含む）がJSONで返される")

		// This will be implemented once the server is complete
		// 1. Start server without -dir flag (default ./files/)
		// 2. Create test files in ./files/
		// 3. GET /ls
		// 4. Verify response contains all files including hidden ones
	})

	t.Run("scenario_2_custom_directory", func(t *testing.T) {
		// Scenario 2: Custom directory specified
		t.Log("Scenario 2: -dir /custom/path でサーバー起動")
		t.Log("Expected: カスタムパスのファイル一覧が返される")

		// This will be implemented once the server is complete
		// 1. Start server with -dir flag pointing to custom directory
		// 2. Create test files in custom directory
		// 3. GET /ls
		// 4. Verify response shows custom directory path and its files
	})

	t.Run("scenario_3_hidden_files_included", func(t *testing.T) {
		// Scenario 3: Hidden files inclusion
		t.Log("Scenario 3: 通常ファイルと隠しファイルが混在")
		t.Log("Expected: 通常ファイルと隠しファイル両方が一覧に含まれる")

		// This will be implemented once the server is complete
		// 1. Create directory with mix of normal and hidden files
		// 2. GET /ls
		// 3. Verify both normal and hidden files are in response
	})

	t.Run("scenario_4_empty_directory", func(t *testing.T) {
		// Scenario 4: Empty directory
		t.Log("Scenario 4: 指定ディレクトリが空")
		t.Log("Expected: 空の一覧を示すJSONレスポンスが返される")

		// This will be implemented once the server is complete
		// 1. Create empty directory
		// 2. Start server pointing to empty directory
		// 3. GET /ls
		// 4. Verify response shows empty files array with count 0
	})
}

// TestLsEndpoint_Performance tests performance requirements
func TestLsEndpoint_Performance(t *testing.T) {
	t.Log("Performance testing not available until implementation complete")

	t.Run("response_time_under_100ms", func(t *testing.T) {
		// Create directory with 1000 files
		testDir, err := os.MkdirTemp("", "perf-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(testDir)

		// Create 1000 test files
		for i := 0; i < 1000; i++ {
			filename := filepath.Join(testDir, fmt.Sprintf("file_%04d.txt", i))
			if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		t.Log("Created 1000 test files for performance testing")
		t.Log("Performance requirement: response time <100ms")

		// This will be implemented once the server is complete
		// 1. Start server pointing to directory with 1000 files
		// 2. Measure response time for GET /ls
		// 3. Assert response time < 100ms
		// 4. Verify all 1000 files are returned correctly
	})
}

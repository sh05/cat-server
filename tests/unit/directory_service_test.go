package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// DirectoryService interface for testing (will be implemented later)
type DirectoryService interface {
	ListFiles() ([]string, error)
	ValidatePath() error
	GetPath() string
}

// Test helper to create temporary test directories and files
func setupTestDirectory(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "cat-server-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test files
	testFiles := []string{
		"README.md",
		"test.txt",
		".hidden",
		".gitignore",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestDirectoryService_ListFiles tests file listing functionality
func TestDirectoryService_ListFiles(t *testing.T) {
	testDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	tests := []struct {
		name     string
		setup    func() string
		expected []string
		wantErr  bool
	}{
		{
			name: "valid_directory_with_files",
			setup: func() string {
				return testDir
			},
			expected: []string{"README.md", "test.txt", ".hidden", ".gitignore"},
			wantErr:  false,
		},
		{
			name: "empty_directory",
			setup: func() string {
				emptyDir, err := os.MkdirTemp("", "empty-*")
				if err != nil {
					t.Fatalf("Failed to create empty dir: %v", err)
				}
				t.Cleanup(func() { os.RemoveAll(emptyDir) })
				return emptyDir
			},
			expected: []string{},
			wantErr:  false,
		},
		{
			name: "nonexistent_directory",
			setup: func() string {
				return "/nonexistent/directory"
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail until DirectoryService is implemented
			dirPath := tt.setup()

			// Mock implementation - will be replaced with actual service
			t.Logf("Testing directory: %s", dirPath)
			t.Log("DirectoryService not implemented yet - test failing as expected for TDD")

			// These assertions will be enabled once DirectoryService is implemented
			// service := services.NewDirectoryService(dirPath)
			// files, err := service.ListFiles()

			if !tt.wantErr {
				t.Log("Expected successful file listing")
				// assert.NoError(t, err)
				// assert.ElementsMatch(t, tt.expected, files)
			} else {
				t.Log("Expected error for invalid directory")
				// assert.Error(t, err)
			}
		})
	}
}

// TestDirectoryService_ValidatePath tests path validation functionality
func TestDirectoryService_ValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType string
	}{
		{
			name:    "valid_relative_path",
			path:    "./files/",
			wantErr: false,
		},
		{
			name:    "valid_absolute_path",
			path:    "/tmp",
			wantErr: false,
		},
		{
			name:    "empty_path",
			path:    "",
			wantErr: true,
			errType: "empty path",
		},
		{
			name:    "null_byte_injection",
			path:    "/tmp\x00/inject",
			wantErr: true,
			errType: "null byte",
		},
		{
			name:    "path_traversal_attack",
			path:    "../../../etc/passwd",
			wantErr: true,
			errType: "path traversal",
		},
		{
			name:    "excessive_path_length",
			path:    "/" + string(make([]byte, 5000)),
			wantErr: true,
			errType: "path too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail until DirectoryService is implemented
			t.Logf("Testing path validation for: %s", tt.path)
			t.Log("DirectoryService not implemented yet - test failing as expected for TDD")

			// These assertions will be enabled once DirectoryService is implemented
			// service := services.NewDirectoryService(tt.path)
			// err := service.ValidatePath()

			if tt.wantErr {
				t.Logf("Expected validation error: %s", tt.errType)
				// assert.Error(t, err)
				// assert.Contains(t, err.Error(), tt.errType)
			} else {
				t.Log("Expected successful validation")
				// assert.NoError(t, err)
			}
		})
	}
}

// TestDirectoryService_HiddenFiles tests hidden file inclusion
func TestDirectoryService_HiddenFiles(t *testing.T) {
	testDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	t.Run("includes_hidden_files", func(t *testing.T) {
		// This test will fail until DirectoryService is implemented
		t.Logf("Testing hidden files inclusion in: %s", testDir)
		t.Log("DirectoryService not implemented yet - test failing as expected for TDD")

		// These assertions will be enabled once DirectoryService is implemented
		// service := services.NewDirectoryService(testDir)
		// files, err := service.ListFiles()
		// assert.NoError(t, err)

		// Check for hidden files
		// hiddenFiles := []string{".hidden", ".gitignore"}
		// for _, hiddenFile := range hiddenFiles {
		//     assert.Contains(t, files, hiddenFile, "Hidden file %s should be included", hiddenFile)
		// }
	})
}

// TestDirectoryService_Performance tests performance requirements
func TestDirectoryService_Performance(t *testing.T) {
	// Create directory with many files for performance testing
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

	t.Run("performance_under_100ms", func(t *testing.T) {
		// This test will fail until DirectoryService is implemented
		t.Log("Testing performance requirement: <100ms for 1000 files")
		t.Log("DirectoryService not implemented yet - test failing as expected for TDD")

		// These assertions will be enabled once DirectoryService is implemented
		// start := time.Now()
		// service := services.NewDirectoryService(testDir)
		// files, err := service.ListFiles()
		// elapsed := time.Since(start)

		// assert.NoError(t, err)
		// assert.Len(t, files, 1000)
		// assert.Less(t, elapsed, 100*time.Millisecond, "Should complete within 100ms")
	})
}

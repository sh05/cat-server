package valueobjects

import (
	"testing"
)

func TestFilePath_NewFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid absolute path",
			path:    "/valid/absolute/path",
			wantErr: false,
		},
		{
			name:    "valid file path",
			path:    "/home/user/document.txt",
			wantErr: false,
		},
		{
			name:    "path traversal should fail",
			path:    "/home/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "relative path traversal should fail",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "double dots in filename should be allowed",
			path:    "/home/user/file..txt",
			wantErr: false,
		},
		{
			name:    "empty path should fail",
			path:    "",
			wantErr: true,
		},
		{
			name:    "null byte should fail",
			path:    "/path/with\x00null",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp, err := NewFilePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if fp.String() != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, fp.String())
			}
		})
	}
}

func TestFilePath_IsSecure(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "safe absolute path",
			path:     "/safe/path/file.txt",
			expected: true,
		},
		{
			name:     "path with double dots in parent",
			path:     "/safe/../unsafe/file.txt",
			expected: false,
		},
		{
			name:     "path with consecutive dots",
			path:     "/safe/path/../file.txt",
			expected: false,
		},
		{
			name:     "safe relative path within bounds",
			path:     "safe/file.txt",
			expected: true,
		},
		{
			name:     "hidden file should be safe",
			path:     "/safe/path/.hidden",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp, err := NewFilePath(tt.path)
			if err != nil && tt.expected {
				t.Errorf("Unexpected error for path %s: %v", tt.path, err)
				return
			}

			if err == nil && fp.IsSecure() != tt.expected {
				t.Errorf("Expected IsSecure() = %v for path %s, got %v", tt.expected, tt.path, fp.IsSecure())
			}
		})
	}
}

func TestFilePath_Join(t *testing.T) {
	tests := []struct {
		name         string
		basePath     string
		relativePath string
		expected     string
		wantErr      bool
	}{
		{
			name:         "simple join",
			basePath:     "/base/path",
			relativePath: "file.txt",
			expected:     "/base/path/file.txt",
			wantErr:      false,
		},
		{
			name:         "join with subdirectory",
			basePath:     "/base/path",
			relativePath: "subdir/file.txt",
			expected:     "/base/path/subdir/file.txt",
			wantErr:      false,
		},
		{
			name:         "join with path traversal should fail",
			basePath:     "/base/path",
			relativePath: "../../../etc/passwd",
			expected:     "",
			wantErr:      true,
		},
		{
			name:         "join with empty relative path",
			basePath:     "/base/path",
			relativePath: "",
			expected:     "/base/path",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseFP, err := NewFilePath(tt.basePath)
			if err != nil {
				t.Fatalf("Failed to create base path: %v", err)
			}

			result, err := baseFP.Join(tt.relativePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.String() != tt.expected {
				t.Errorf("Expected joined path %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestFilePath_Base(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "file in root",
			path:     "/file.txt",
			expected: "file.txt",
		},
		{
			name:     "file in subdirectory",
			path:     "/path/to/file.txt",
			expected: "file.txt",
		},
		{
			name:     "directory path",
			path:     "/path/to/directory",
			expected: "directory",
		},
		{
			name:     "path with trailing slash",
			path:     "/path/to/directory/",
			expected: "directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp, err := NewFilePath(tt.path)
			if err != nil {
				t.Fatalf("Failed to create path: %v", err)
			}

			base := fp.Base()
			if base != tt.expected {
				t.Errorf("Expected base %s, got %s", tt.expected, base)
			}
		})
	}
}

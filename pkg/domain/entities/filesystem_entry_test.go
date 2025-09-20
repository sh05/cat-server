package entities

import (
	"os"
	"testing"
	"time"
)

func TestFileSystemEntry_NewFileSystemEntry(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		path        string
		size        int64
		modTime     time.Time
		isDir       bool
		permissions os.FileMode
		wantErr     bool
	}{
		{
			name:        "valid file entry",
			filename:    "test.txt",
			path:        "/path/to/test.txt",
			size:        1024,
			modTime:     time.Now(),
			isDir:       false,
			permissions: 0644,
			wantErr:     false,
		},
		{
			name:        "valid directory entry",
			filename:    "testdir",
			path:        "/path/to/testdir",
			size:        0,
			modTime:     time.Now(),
			isDir:       true,
			permissions: 0755,
			wantErr:     false,
		},
		{
			name:        "empty filename should fail",
			filename:    "",
			path:        "/path/to/",
			size:        0,
			modTime:     time.Now(),
			isDir:       false,
			permissions: 0644,
			wantErr:     true,
		},
		{
			name:        "negative size should fail",
			filename:    "test.txt",
			path:        "/path/to/test.txt",
			size:        -1,
			modTime:     time.Now(),
			isDir:       false,
			permissions: 0644,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := NewFileSystemEntry(tt.filename, tt.path, tt.size, tt.modTime, tt.isDir, tt.permissions)

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

			if entry.Name() != tt.filename {
				t.Errorf("Expected name %s, got %s", tt.filename, entry.Name())
			}

			if entry.Path() != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, entry.Path())
			}

			if entry.Size() != tt.size {
				t.Errorf("Expected size %d, got %d", tt.size, entry.Size())
			}

			if entry.IsDir() != tt.isDir {
				t.Errorf("Expected isDir %v, got %v", tt.isDir, entry.IsDir())
			}
		})
	}
}

func TestFileSystemEntry_IsSecure(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "safe path",
			path:     "/safe/path/file.txt",
			expected: true,
		},
		{
			name:     "path traversal attack",
			path:     "/safe/../../../etc/passwd",
			expected: false,
		},
		{
			name:     "relative path traversal",
			path:     "./../../etc/passwd",
			expected: false,
		},
		{
			name:     "hidden file should be allowed",
			path:     "/safe/path/.hidden",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, _ := NewFileSystemEntry("test", tt.path, 0, time.Now(), false, 0644)
			if entry != nil && entry.IsSecure() != tt.expected {
				t.Errorf("Expected IsSecure() = %v, got %v", tt.expected, entry.IsSecure())
			}
		})
	}
}

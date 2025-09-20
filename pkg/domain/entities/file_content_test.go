package entities

import (
	"strings"
	"testing"
	"time"
)

func TestFileContent_NewFileContent(t *testing.T) {
	testTime := time.Now()
	entry, _ := NewFileSystemEntry("test.txt", "/path/test.txt", 100, testTime, false, 0644)

	tests := []struct {
		name     string
		entry    *FileSystemEntry
		content  []byte
		encoding string
		wantErr  bool
	}{
		{
			name:     "valid file content",
			entry:    entry,
			content:  []byte("Hello, World!"),
			encoding: "utf-8",
			wantErr:  false,
		},
		{
			name:     "empty content should be allowed",
			entry:    entry,
			content:  []byte{},
			encoding: "utf-8",
			wantErr:  false,
		},
		{
			name:     "nil entry should fail",
			entry:    nil,
			content:  []byte("content"),
			encoding: "utf-8",
			wantErr:  true,
		},
		{
			name: "directory entry should fail",
			entry: func() *FileSystemEntry {
				dir, _ := NewFileSystemEntry("dir", "/path/dir", 0, testTime, true, 0755)
				return dir
			}(),
			content:  []byte("content"),
			encoding: "utf-8",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewFileContent(tt.entry, tt.content, tt.encoding)

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

			if string(content.Content()) != string(tt.content) {
				t.Errorf("Expected content %s, got %s", string(tt.content), string(content.Content()))
			}

			if content.Encoding() != tt.encoding {
				t.Errorf("Expected encoding %s, got %s", tt.encoding, content.Encoding())
			}

			if content.Entry().Name() != tt.entry.Name() {
				t.Errorf("Expected entry name %s, got %s", tt.entry.Name(), content.Entry().Name())
			}
		})
	}
}

func TestFileContent_IsTextContent(t *testing.T) {
	entry, _ := NewFileSystemEntry("test.txt", "/path/test.txt", 100, time.Now(), false, 0644)

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "plain text content",
			content:  []byte("Hello, World!"),
			expected: true,
		},
		{
			name:     "text with newlines",
			content:  []byte("Line 1\nLine 2\nLine 3"),
			expected: true,
		},
		{
			name:     "binary content",
			content:  []byte{0x00, 0x01, 0x02, 0xFF},
			expected: false,
		},
		{
			name:     "mixed content with null bytes",
			content:  []byte("Text\x00with\x00nulls"),
			expected: false,
		},
		{
			name:     "empty content",
			content:  []byte{},
			expected: true, // empty content is considered text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileContent, _ := NewFileContent(entry, tt.content, "utf-8")
			if fileContent.IsTextContent() != tt.expected {
				t.Errorf("Expected IsTextContent() = %v, got %v", tt.expected, fileContent.IsTextContent())
			}
		})
	}
}

func TestFileContent_GetContentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  []byte
		expected string
	}{
		{
			name:     "text file",
			filename: "test.txt",
			content:  []byte("Hello, World!"),
			expected: "text/plain",
		},
		{
			name:     "json file",
			filename: "config.json",
			content:  []byte(`{"key": "value"}`),
			expected: "application/json",
		},
		{
			name:     "html file",
			filename: "index.html",
			content:  []byte("<html><body>Hello</body></html>"),
			expected: "text/html",
		},
		{
			name:     "binary content",
			filename: "data.bin",
			content:  []byte{0x00, 0x01, 0x02, 0xFF},
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create entry with specific filename
			testEntry, _ := NewFileSystemEntry(tt.filename, "/path/"+tt.filename, int64(len(tt.content)), time.Now(), false, 0644)
			fileContent, _ := NewFileContent(testEntry, tt.content, "utf-8")

			contentType := fileContent.GetContentType()
			if !strings.HasPrefix(contentType, tt.expected) {
				t.Errorf("Expected content type to start with %s, got %s", tt.expected, contentType)
			}
		})
	}
}

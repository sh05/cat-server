package entities

import (
	"bytes"
	"errors"
	"mime"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// FileContent represents the content and metadata of a file
type FileContent struct {
	entry    *FileSystemEntry
	content  []byte
	encoding string
	readAt   time.Time
}

// NewFileContent creates a new FileContent with validation
func NewFileContent(entry *FileSystemEntry, content []byte, encoding string) (*FileContent, error) {
	if entry == nil {
		return nil, errors.New("file entry cannot be nil")
	}

	if entry.IsDir() {
		return nil, errors.New("cannot create file content for directory")
	}

	if encoding == "" {
		encoding = "utf-8" // Default encoding
	}

	return &FileContent{
		entry:    entry,
		content:  content,
		encoding: encoding,
		readAt:   time.Now(),
	}, nil
}

// Entry returns the associated FileSystemEntry
func (f *FileContent) Entry() *FileSystemEntry {
	return f.entry
}

// Content returns the raw content bytes
func (f *FileContent) Content() []byte {
	// Return a copy to prevent external modification
	contentCopy := make([]byte, len(f.content))
	copy(contentCopy, f.content)
	return contentCopy
}

// ContentAsString returns the content as a string
func (f *FileContent) ContentAsString() string {
	return string(f.content)
}

// Encoding returns the character encoding
func (f *FileContent) Encoding() string {
	return f.encoding
}

// ReadAt returns when the content was read
func (f *FileContent) ReadAt() time.Time {
	return f.readAt
}

// Size returns the content size in bytes
func (f *FileContent) Size() int64 {
	return int64(len(f.content))
}

// IsTextContent determines if the content is text (not binary)
func (f *FileContent) IsTextContent() bool {
	// Empty content is considered text
	if len(f.content) == 0 {
		return true
	}

	// Check for null bytes which typically indicate binary content
	if bytes.Contains(f.content, []byte{0}) {
		return false
	}

	// Check if content is valid UTF-8
	if !utf8.Valid(f.content) {
		return false
	}

	// Additional heuristics can be added here
	return true
}

// GetContentType determines the MIME content type
func (f *FileContent) GetContentType() string {
	// Get MIME type from file extension
	ext := filepath.Ext(f.entry.Name())
	mimeType := mime.TypeByExtension(ext)

	if mimeType != "" {
		return mimeType
	}

	// Fallback based on content analysis
	if f.IsTextContent() {
		// Check for specific text formats
		content := string(f.content)
		if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
			return "application/json"
		}
		if strings.HasPrefix(content, "<") && strings.Contains(content, ">") {
			return "text/html"
		}
		return "text/plain"
	}

	return "application/octet-stream"
}

// GetLines returns the content split into lines
func (f *FileContent) GetLines() []string {
	if !f.IsTextContent() {
		return nil
	}

	content := strings.ReplaceAll(string(f.content), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.Split(content, "\n")
}

// GetLineCount returns the number of lines in the file
func (f *FileContent) GetLineCount() int {
	if !f.IsTextContent() {
		return 0
	}
	lines := f.GetLines()
	return len(lines)
}

// IsEmpty returns true if the content is empty
func (f *FileContent) IsEmpty() bool {
	return len(f.content) == 0
}

// GetPreview returns the first n characters of the content as a preview
func (f *FileContent) GetPreview(maxChars int) string {
	if !f.IsTextContent() {
		return "[Binary content]"
	}

	content := string(f.content)
	if len(content) <= maxChars {
		return content
	}

	return content[:maxChars] + "..."
}

// ValidateSize checks if the content size is within acceptable limits
func (f *FileContent) ValidateSize(maxSize int64) error {
	if f.Size() > maxSize {
		return errors.New("file content exceeds maximum allowed size")
	}
	return nil
}

// GetContentHash returns a simple hash of the content for comparison
func (f *FileContent) GetContentHash() uint32 {
	// Simple hash function for content comparison
	var hash uint32 = 2166136261 // FNV offset basis
	for _, b := range f.content {
		hash ^= uint32(b)
		hash *= 16777619 // FNV prime
	}
	return hash
}

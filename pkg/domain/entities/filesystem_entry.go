package entities

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileSystemEntry represents a file or directory in the filesystem
type FileSystemEntry struct {
	name        string
	path        string
	size        int64
	modTime     time.Time
	isDir       bool
	permissions os.FileMode
}

// NewFileSystemEntry creates a new FileSystemEntry with validation
func NewFileSystemEntry(name, path string, size int64, modTime time.Time, isDir bool, permissions os.FileMode) (*FileSystemEntry, error) {
	if name == "" {
		return nil, errors.New("filename cannot be empty")
	}

	if size < 0 {
		return nil, errors.New("file size cannot be negative")
	}

	// Store the original path for security checking
	originalPath := path

	// Clean the path
	cleanPath := filepath.Clean(path)

	entry := &FileSystemEntry{
		name:        name,
		path:        cleanPath,
		size:        size,
		modTime:     modTime,
		isDir:       isDir,
		permissions: permissions,
	}

	// Check security on the original path
	if strings.Contains(originalPath, "../") || strings.Contains(originalPath, "..\\") {
		entry.path = originalPath // Keep original for security check
	}

	return entry, nil
}

// Name returns the filename
func (f *FileSystemEntry) Name() string {
	return f.name
}

// Path returns the full path
func (f *FileSystemEntry) Path() string {
	return f.path
}

// Size returns the file size in bytes
func (f *FileSystemEntry) Size() int64 {
	return f.size
}

// ModTime returns the modification time
func (f *FileSystemEntry) ModTime() time.Time {
	return f.modTime
}

// IsDir returns true if this is a directory
func (f *FileSystemEntry) IsDir() bool {
	return f.isDir
}

// Permissions returns the file permissions
func (f *FileSystemEntry) Permissions() os.FileMode {
	return f.permissions
}

// IsSecure checks if the path is safe from directory traversal attacks
func (f *FileSystemEntry) IsSecure() bool {
	// Check for path traversal patterns
	if strings.Contains(f.path, "../") || strings.Contains(f.path, "..\\") {
		return false
	}

	// Check for null bytes
	if strings.Contains(f.path, "\x00") {
		return false
	}

	// Additional security checks can be added here
	return true
}

// IsHidden returns true if the file is hidden (starts with .)
func (f *FileSystemEntry) IsHidden() bool {
	return strings.HasPrefix(f.name, ".")
}

// HumanReadableSize returns the size in human-readable format
func (f *FileSystemEntry) HumanReadableSize() string {
	if f.isDir {
		return "-"
	}

	const unit = 1024
	if f.size < unit {
		return "< 1KB"
	}

	div, exp := int64(unit), 0
	for n := f.size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f%s", float64(f.size)/float64(div), units[exp])
}

// IsExecutable returns true if the file has execute permissions
func (f *FileSystemEntry) IsExecutable() bool {
	return f.permissions&0111 != 0
}

// IsReadable returns true if the file has read permissions
func (f *FileSystemEntry) IsReadable() bool {
	return f.permissions&0444 != 0
}

// IsWritable returns true if the file has write permissions
func (f *FileSystemEntry) IsWritable() bool {
	return f.permissions&0222 != 0
}

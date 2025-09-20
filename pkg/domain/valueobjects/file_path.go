package valueobjects

import (
	"errors"
	"path/filepath"
	"strings"
)

// FilePath represents a secure file path value object
type FilePath struct {
	value string
}

// NewFilePath creates a new FilePath with validation
func NewFilePath(path string) (*FilePath, error) {
	if path == "" {
		return nil, errors.New("file path cannot be empty")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return nil, errors.New("file path cannot contain null bytes")
	}

	// Check for path traversal BEFORE cleaning
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return nil, errors.New("insecure file path detected")
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	fp := &FilePath{
		value: cleanPath,
	}

	return fp, nil
}

// String returns the path as a string
func (fp *FilePath) String() string {
	return fp.value
}

// IsSecure checks if the path is safe from directory traversal attacks
func (fp *FilePath) IsSecure() bool {
	// Check for path traversal patterns
	if strings.Contains(fp.value, "../") || strings.Contains(fp.value, "..\\") {
		return false
	}

	// Check for absolute path traversal attempts
	if strings.Contains(fp.value, "/..") || strings.Contains(fp.value, "\\..") {
		return false
	}

	// Check for null bytes
	if strings.Contains(fp.value, "\x00") {
		return false
	}

	return true
}

// Join safely joins this path with a relative path
func (fp *FilePath) Join(relativePath string) (*FilePath, error) {
	if relativePath == "" {
		return fp, nil
	}

	// Check for path traversal in the relative path
	if strings.Contains(relativePath, "../") || strings.Contains(relativePath, "..\\") {
		return nil, errors.New("relative path contains path traversal attempt")
	}

	joined := filepath.Join(fp.value, relativePath)
	return NewFilePath(joined)
}

// Base returns the base name of the path
func (fp *FilePath) Base() string {
	base := filepath.Base(fp.value)
	// Handle trailing slashes
	if base == "." && strings.HasSuffix(fp.value, "/") {
		// Get the directory name before the trailing slash
		trimmed := strings.TrimSuffix(fp.value, "/")
		if trimmed != "" {
			base = filepath.Base(trimmed)
		}
	}
	return base
}

// Dir returns the directory portion of the path
func (fp *FilePath) Dir() string {
	return filepath.Dir(fp.value)
}

// Ext returns the file extension
func (fp *FilePath) Ext() string {
	return filepath.Ext(fp.value)
}

// IsAbsolute returns true if the path is absolute
func (fp *FilePath) IsAbsolute() bool {
	return filepath.IsAbs(fp.value)
}

// IsRoot returns true if the path represents the root directory
func (fp *FilePath) IsRoot() bool {
	return fp.value == "/" || fp.value == "\\"
}

// Contains checks if this path contains the given subpath
func (fp *FilePath) Contains(subpath string) bool {
	return strings.Contains(fp.value, subpath)
}

// HasPrefix checks if the path starts with the given prefix
func (fp *FilePath) HasPrefix(prefix string) bool {
	return strings.HasPrefix(fp.value, prefix)
}

// HasSuffix checks if the path ends with the given suffix
func (fp *FilePath) HasSuffix(suffix string) bool {
	return strings.HasSuffix(fp.value, suffix)
}

// Equals checks if this path equals another FilePath
func (fp *FilePath) Equals(other *FilePath) bool {
	if other == nil {
		return false
	}
	return fp.value == other.value
}

// IsWithinDirectory checks if this path is within the given directory
func (fp *FilePath) IsWithinDirectory(dir string) bool {
	cleanDir := filepath.Clean(dir)

	// Ensure the directory path ends with a separator for proper checking
	if !strings.HasSuffix(cleanDir, "/") && !strings.HasSuffix(cleanDir, "\\") {
		cleanDir += string(filepath.Separator)
	}

	return strings.HasPrefix(fp.value, cleanDir)
}

// Normalize returns a normalized version of the path
func (fp *FilePath) Normalize() *FilePath {
	normalized := filepath.Clean(fp.value)
	// Note: We can safely create this without validation since the original was valid
	return &FilePath{value: normalized}
}

// Split splits the path into directory and filename
func (fp *FilePath) Split() (dir, file string) {
	return filepath.Split(fp.value)
}

// Validate performs additional validation checks
func (fp *FilePath) Validate() error {
	if !fp.IsSecure() {
		return errors.New("path failed security validation")
	}

	// Additional validation rules can be added here
	// For example, checking against a whitelist of allowed paths
	// or validating against specific naming conventions

	return nil
}

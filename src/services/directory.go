package services

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// DirectoryService handles directory operations for file listing
type DirectoryService struct {
	basePath string
}

// NewDirectoryService creates a new DirectoryService with the specified path
func NewDirectoryService(path string) (*DirectoryService, error) {
	service := &DirectoryService{
		basePath: path,
	}

	if err := service.ValidatePath(); err != nil {
		return nil, err
	}

	return service, nil
}

// GetPath returns the configured directory path
func (ds *DirectoryService) GetPath() string {
	return ds.basePath
}

// ValidatePath validates the directory path for security and correctness
func (ds *DirectoryService) ValidatePath() error {
	// Check for empty path
	if ds.basePath == "" {
		return fmt.Errorf("empty path not allowed")
	}

	// Check for null bytes (security vulnerability)
	if strings.Contains(ds.basePath, "\x00") {
		return fmt.Errorf("null byte in path not allowed")
	}

	// Check path length (Unix/Linux standard)
	if len(ds.basePath) > 4096 {
		return fmt.Errorf("path too long (max 4096 characters)")
	}

	// Basic path traversal protection
	cleanPath := filepath.Clean(ds.basePath)
	if strings.Contains(cleanPath, "..") && strings.Count(cleanPath, "..") > 1 {
		return fmt.Errorf("excessive path traversal detected")
	}

	return nil
}

// ListFiles returns all files in the directory, including hidden files
func (ds *DirectoryService) ListFiles() ([]string, error) {
	// First validate the path
	if err := ds.ValidatePath(); err != nil {
		return nil, err
	}

	// Check if directory exists and is accessible
	info, err := os.Stat(ds.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", ds.basePath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied: %s", ds.basePath)
		}
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	// Ensure it's actually a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", ds.basePath)
	}

	// Read directory contents
	entries, err := os.ReadDir(ds.basePath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading directory: %s", ds.basePath)
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Extract file names (including hidden files, excluding subdirectories)
	var files []string
	for _, entry := range entries {
		// Only include files, not directories
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ValidateForRead performs additional read-specific validation
func (ds *DirectoryService) ValidateForRead() error {
	// Check if directory is readable
	file, err := os.Open(ds.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist")
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied")
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}
	defer file.Close()

	// Try to read at least one entry to confirm read access
	_, err = file.Readdir(1)
	if err != nil && err.Error() != "EOF" {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied reading directory")
		}
		return fmt.Errorf("failed to read directory: %w", err)
	}

	return nil
}

// ReadFile reads a file from the configured directory with security validation
func (ds *DirectoryService) ReadFile(filename string) ([]byte, error) {
	// Validate the filename
	if err := ds.ValidateFilename(filename); err != nil {
		return nil, err
	}

	// Construct safe file path
	cleanFilename := ds.sanitizeFilename(filename)
	filePath := filepath.Join(ds.basePath, cleanFilename)

	// Ensure the resolved path is still within the base directory
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(ds.basePath)) {
		return nil, fmt.Errorf("invalid path: file outside base directory")
	}

	// Check file existence and get info
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied: %s", filename)
		}
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	// Ensure it's a file, not a directory
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", filename)
	}

	// Check file size limit (10MB)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileSize)
	}

	// Read file content
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading file: %s", filename)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Validate content is text (UTF-8)
	if len(content) > 0 && !utf8.Valid(content) {
		return nil, fmt.Errorf("binary file not supported: %s", filename)
	}

	return content, nil
}

// ValidateFilename validates a filename for security and correctness
func (ds *DirectoryService) ValidateFilename(filename string) error {
	// Check for empty filename
	if filename == "" {
		return fmt.Errorf("empty filename not allowed")
	}

	// Check for null bytes (security vulnerability)
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("null byte in filename not allowed")
	}

	// Check filename length (Unix/Linux standard)
	if len(filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}

	// Check for path traversal patterns
	if ds.isPathTraversal(filename) {
		return fmt.Errorf("invalid filename: path traversal detected")
	}

	// Check for reserved names (current and parent directory)
	if filename == "." || filename == ".." {
		return fmt.Errorf("invalid filename: reserved name")
	}

	return nil
}

// isPathTraversal checks if a filename contains path traversal patterns
func (ds *DirectoryService) isPathTraversal(filename string) bool {
	// Decode URL encoding first
	decoded, err := url.QueryUnescape(filename)
	if err != nil {
		// If decoding fails, use original but still check patterns
		decoded = filename
	}

	// Check for obvious path traversal patterns
	dangerous := []string{
		"../", "..\\",
		"%2e%2e%2f", "%2e%2e%5c",
		"%252e%252e%252f", "%252e%252e%255c",
		"..%2f", "..%5c",
		"....//", "....\\\\",
	}

	lowerFilename := strings.ToLower(filename)
	lowerDecoded := strings.ToLower(decoded)

	for _, pattern := range dangerous {
		if strings.Contains(lowerFilename, pattern) || strings.Contains(lowerDecoded, pattern) {
			return true
		}
	}

	// Check after path cleaning
	cleanDecoded := filepath.Clean(decoded)
	if strings.Contains(cleanDecoded, "..") {
		return true
	}

	return false
}

// sanitizeFilename performs basic filename sanitization
func (ds *DirectoryService) sanitizeFilename(filename string) string {
	// URL decode if needed
	decoded, err := url.QueryUnescape(filename)
	if err != nil {
		decoded = filename
	}

	// Remove any null bytes
	cleaned := strings.ReplaceAll(decoded, "\x00", "")

	// Return the cleaned filename
	return cleaned
}

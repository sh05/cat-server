package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

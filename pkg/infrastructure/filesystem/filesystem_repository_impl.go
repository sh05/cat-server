package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sh05/cat-server/pkg/domain/entities"
	"github.com/sh05/cat-server/pkg/domain/repositories"
	"github.com/sh05/cat-server/pkg/domain/valueobjects"
)

// FileSystemRepositoryImpl implements the FileSystemRepository interface
type FileSystemRepositoryImpl struct {
	basePath    string
	maxFileSize int64
}

// NewFileSystemRepository creates a new filesystem repository implementation
func NewFileSystemRepository(basePath string, maxFileSize int64) *FileSystemRepositoryImpl {
	return &FileSystemRepositoryImpl{
		basePath:    basePath,
		maxFileSize: maxFileSize,
	}
}

// ListDirectory returns a directory listing for the given path
func (r *FileSystemRepositoryImpl) ListDirectory(path *valueobjects.FilePath) (*entities.DirectoryListing, error) {
	fullPath := filepath.Join(r.basePath, path.String())

	// Validate path security
	if err := r.ValidatePath(path); err != nil {
		return nil, err
	}

	// Check if directory exists and is readable
	if !r.IsDirectory(path) {
		return nil, repositories.NewFileSystemError(
			"ListDirectory",
			path.String(),
			"path is not a directory",
			repositories.ErrorInvalidPath,
		)
	}

	// Read directory entries
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, repositories.NewFileSystemError(
			"ListDirectory",
			path.String(),
			err.Error(),
			repositories.ErrorPermissionDenied,
		)
	}

	// Convert to domain entities
	var fileEntries []entities.FileSystemEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't read
		}

		relativeEntryPath := filepath.Join(path.String(), entry.Name())

		fileEntry, err := entities.NewFileSystemEntry(
			entry.Name(),
			relativeEntryPath,
			info.Size(),
			info.ModTime(),
			entry.IsDir(),
			info.Mode(),
		)
		if err != nil {
			continue // Skip invalid entries
		}

		fileEntries = append(fileEntries, *fileEntry)
	}

	// Create directory listing
	listing, err := entities.NewDirectoryListing(path.String(), fileEntries)
	if err != nil {
		return nil, repositories.NewFileSystemError(
			"ListDirectory",
			path.String(),
			err.Error(),
			repositories.ErrorUnknown,
		)
	}

	return listing, nil
}

// ReadFile returns the content of a file at the given path
func (r *FileSystemRepositoryImpl) ReadFile(path *valueobjects.FilePath) (*entities.FileContent, error) {
	fullPath := filepath.Join(r.basePath, path.String())

	// Validate path security
	if err := r.ValidatePath(path); err != nil {
		return nil, err
	}

	// Check if file exists and is readable
	if !r.Exists(path) {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			"file not found",
			repositories.ErrorNotFound,
		)
	}

	if !r.IsReadable(path) {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			"file not readable",
			repositories.ErrorPermissionDenied,
		)
	}

	// Get file info
	fileEntry, err := r.GetFileInfo(path)
	if err != nil {
		return nil, err
	}

	if fileEntry.IsDir() {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			"path is a directory",
			repositories.ErrorInvalidPath,
		)
	}

	// Check file size limit
	if r.maxFileSize > 0 && fileEntry.Size() > r.maxFileSize {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			"file too large",
			repositories.ErrorFileTooLarge,
		)
	}

	// Read file content
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			err.Error(),
			repositories.ErrorPermissionDenied,
		)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			err.Error(),
			repositories.ErrorUnknown,
		)
	}

	// Create file content entity
	fileContent, err := entities.NewFileContent(fileEntry, content, "utf-8")
	if err != nil {
		return nil, repositories.NewFileSystemError(
			"ReadFile",
			path.String(),
			err.Error(),
			repositories.ErrorUnknown,
		)
	}

	return fileContent, nil
}

// Exists checks if a file or directory exists at the given path
func (r *FileSystemRepositoryImpl) Exists(path *valueobjects.FilePath) bool {
	fullPath := filepath.Join(r.basePath, path.String())
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// IsReadable checks if the file/directory at the given path is readable
func (r *FileSystemRepositoryImpl) IsReadable(path *valueobjects.FilePath) bool {
	fullPath := filepath.Join(r.basePath, path.String())
	file, err := os.Open(fullPath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsDirectory checks if the path points to a directory
func (r *FileSystemRepositoryImpl) IsDirectory(path *valueobjects.FilePath) bool {
	fullPath := filepath.Join(r.basePath, path.String())
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileInfo returns basic information about a file/directory
func (r *FileSystemRepositoryImpl) GetFileInfo(path *valueobjects.FilePath) (*entities.FileSystemEntry, error) {
	fullPath := filepath.Join(r.basePath, path.String())

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, repositories.NewFileSystemError(
				"GetFileInfo",
				path.String(),
				"file not found",
				repositories.ErrorNotFound,
			)
		}
		return nil, repositories.NewFileSystemError(
			"GetFileInfo",
			path.String(),
			err.Error(),
			repositories.ErrorUnknown,
		)
	}

	entry, err := entities.NewFileSystemEntry(
		path.Base(),
		path.String(),
		info.Size(),
		info.ModTime(),
		info.IsDir(),
		info.Mode(),
	)
	if err != nil {
		return nil, repositories.NewFileSystemError(
			"GetFileInfo",
			path.String(),
			err.Error(),
			repositories.ErrorUnknown,
		)
	}

	return entry, nil
}

// ValidatePath performs security and accessibility checks on the path
func (r *FileSystemRepositoryImpl) ValidatePath(path *valueobjects.FilePath) error {
	// Check path security
	if !path.IsSecure() {
		return repositories.NewFileSystemError(
			"ValidatePath",
			path.String(),
			"insecure path detected",
			repositories.ErrorPathTraversal,
		)
	}

	// Ensure path is within base directory bounds
	fullPath := filepath.Join(r.basePath, path.String())
	cleanFullPath := filepath.Clean(fullPath)
	cleanBasePath := filepath.Clean(r.basePath)

	// Check if the resolved path is still within the base directory
	relPath, err := filepath.Rel(cleanBasePath, cleanFullPath)
	if err != nil || filepath.IsAbs(relPath) || (len(relPath) >= 2 && relPath[0:2] == "..") {
		return repositories.NewFileSystemError(
			"ValidatePath",
			path.String(),
			"path outside allowed directory",
			repositories.ErrorPathTraversal,
		)
	}

	return nil
}

// GetDirectoryStats returns statistics about a directory
func (r *FileSystemRepositoryImpl) GetDirectoryStats(path *valueobjects.FilePath) (*repositories.DirectoryStats, error) {
	listing, err := r.ListDirectory(path)
	if err != nil {
		return nil, err
	}

	stats := &repositories.DirectoryStats{
		TotalFiles:       listing.GetFileCount(),
		TotalDirectories: listing.GetDirectoryCount(),
		TotalSize:        listing.GetTotalSize(),
	}

	// Find largest, newest, and oldest files
	entries := listing.Entries()
	if len(entries) > 0 {
		var largestFile, newestFile, oldestFile *entities.FileSystemEntry

		for i, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// Check for largest file
			if largestFile == nil || entry.Size() > largestFile.Size() {
				largestFile = &entries[i]
			}

			// Check for newest file
			if newestFile == nil || entry.ModTime().After(newestFile.ModTime()) {
				newestFile = &entries[i]
			}

			// Check for oldest file
			if oldestFile == nil || entry.ModTime().Before(oldestFile.ModTime()) {
				oldestFile = &entries[i]
			}
		}

		stats.LargestFile = largestFile
		stats.NewestFile = newestFile
		stats.OldestFile = oldestFile
	}

	return stats, nil
}

// GetBasePath returns the base path for this repository
func (r *FileSystemRepositoryImpl) GetBasePath() string {
	return r.basePath
}

// SetMaxFileSize sets the maximum file size limit
func (r *FileSystemRepositoryImpl) SetMaxFileSize(maxSize int64) {
	r.maxFileSize = maxSize
}

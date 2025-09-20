package repositories

import (
	"fmt"

	"github.com/sh05/cat-server/pkg/domain/entities"
	"github.com/sh05/cat-server/pkg/domain/valueobjects"
)

// FileSystemRepository defines the interface for filesystem operations
type FileSystemRepository interface {
	// ListDirectory returns a directory listing for the given path
	ListDirectory(path *valueobjects.FilePath) (*entities.DirectoryListing, error)

	// ReadFile returns the content of a file at the given path
	ReadFile(path *valueobjects.FilePath) (*entities.FileContent, error)

	// Exists checks if a file or directory exists at the given path
	Exists(path *valueobjects.FilePath) bool

	// IsReadable checks if the file/directory at the given path is readable
	IsReadable(path *valueobjects.FilePath) bool

	// IsDirectory checks if the path points to a directory
	IsDirectory(path *valueobjects.FilePath) bool

	// GetFileInfo returns basic information about a file/directory
	GetFileInfo(path *valueobjects.FilePath) (*entities.FileSystemEntry, error)

	// ValidatePath performs security and accessibility checks on the path
	ValidatePath(path *valueobjects.FilePath) error

	// GetDirectoryStats returns statistics about a directory
	GetDirectoryStats(path *valueobjects.FilePath) (*DirectoryStats, error)
}

// DirectoryStats represents statistics about a directory
type DirectoryStats struct {
	TotalFiles       int
	TotalDirectories int
	TotalSize        int64
	LargestFile      *entities.FileSystemEntry
	NewestFile       *entities.FileSystemEntry
	OldestFile       *entities.FileSystemEntry
}

// FileFilter defines criteria for filtering files
type FileFilter struct {
	IncludeHidden  bool
	IncludeSystem  bool
	FileExtensions []string
	MaxSize        int64
	MinSize        int64
	SortBy         SortCriteria
	SortOrder      SortOrder
}

// SortCriteria defines how to sort directory listings
type SortCriteria int

const (
	SortByName SortCriteria = iota
	SortBySize
	SortByModTime
	SortByType
)

// SortOrder defines the sort direction
type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

// SecurityPolicy defines security constraints for file operations
type SecurityPolicy struct {
	AllowedDirectories []string
	BlockedDirectories []string
	AllowedExtensions  []string
	BlockedExtensions  []string
	MaxFileSize        int64
	AllowHidden        bool
	AllowSymlinks      bool
}

// OperationContext provides context for filesystem operations
type OperationContext struct {
	UserID         string
	RequestID      string
	SecurityPolicy *SecurityPolicy
	Filter         *FileFilter
	Timeout        int // seconds
}

// FileSystemError represents errors specific to filesystem operations
type FileSystemError struct {
	Operation string
	Path      string
	Reason    string
	Code      ErrorCode
}

// ErrorCode represents different types of filesystem errors
type ErrorCode int

const (
	ErrorNotFound ErrorCode = iota
	ErrorPermissionDenied
	ErrorPathTraversal
	ErrorInvalidPath
	ErrorFileTooLarge
	ErrorDirectoryNotEmpty
	ErrorDiskFull
	ErrorTimeout
	ErrorUnknown
)

// Error implements the error interface
func (e *FileSystemError) Error() string {
	return fmt.Sprintf("filesystem error in %s for path '%s': %s", e.Operation, e.Path, e.Reason)
}

// NewFileSystemError creates a new FileSystemError
func NewFileSystemError(operation, path, reason string, code ErrorCode) *FileSystemError {
	return &FileSystemError{
		Operation: operation,
		Path:      path,
		Reason:    reason,
		Code:      code,
	}
}

package valueobjects

import (
	"errors"
	"fmt"
)

// FileSize represents a file size value object
type FileSize struct {
	bytes int64
}

const (
	Byte = 1
	KB   = 1024 * Byte
	MB   = 1024 * KB
	GB   = 1024 * MB
	TB   = 1024 * GB
)

// NewFileSize creates a new FileSize with validation
func NewFileSize(bytes int64) (*FileSize, error) {
	if bytes < 0 {
		return nil, errors.New("file size cannot be negative")
	}

	return &FileSize{
		bytes: bytes,
	}, nil
}

// Bytes returns the size in bytes
func (fs *FileSize) Bytes() int64 {
	return fs.bytes
}

// IsEmpty returns true if the file size is zero
func (fs *FileSize) IsEmpty() bool {
	return fs.bytes == 0
}

// IsLarge returns true if the file size exceeds the given threshold
func (fs *FileSize) IsLarge(threshold int64) bool {
	return fs.bytes > threshold
}

// HumanReadable returns the size in human-readable format
func (fs *FileSize) HumanReadable() string {
	if fs.bytes == 0 {
		return "0 B"
	}

	if fs.bytes < KB {
		return fmt.Sprintf("%d B", fs.bytes)
	}

	if fs.bytes < MB {
		return fmt.Sprintf("%.1f KB", float64(fs.bytes)/float64(KB))
	}

	if fs.bytes < GB {
		return fmt.Sprintf("%.1f MB", float64(fs.bytes)/float64(MB))
	}

	if fs.bytes < TB {
		return fmt.Sprintf("%.1f GB", float64(fs.bytes)/float64(GB))
	}

	return fmt.Sprintf("%.1f TB", float64(fs.bytes)/float64(TB))
}

// Add returns a new FileSize representing the sum of this and another size
func (fs *FileSize) Add(other *FileSize) (*FileSize, error) {
	if other == nil {
		return fs, nil
	}

	newSize := fs.bytes + other.bytes
	if newSize < 0 { // Check for overflow
		return nil, errors.New("file size overflow")
	}

	return NewFileSize(newSize)
}

// Subtract returns a new FileSize representing the difference
func (fs *FileSize) Subtract(other *FileSize) (*FileSize, error) {
	if other == nil {
		return fs, nil
	}

	newSize := fs.bytes - other.bytes
	if newSize < 0 {
		return nil, errors.New("file size cannot be negative")
	}

	return NewFileSize(newSize)
}

// Compare compares this size with another FileSize
// Returns: -1 if smaller, 0 if equal, 1 if larger
func (fs *FileSize) Compare(other *FileSize) int {
	if other == nil {
		return 1
	}

	if fs.bytes < other.bytes {
		return -1
	}
	if fs.bytes > other.bytes {
		return 1
	}
	return 0
}

// Equals checks if this size equals another FileSize
func (fs *FileSize) Equals(other *FileSize) bool {
	return fs.Compare(other) == 0
}

// IsGreaterThan checks if this size is greater than another
func (fs *FileSize) IsGreaterThan(other *FileSize) bool {
	return fs.Compare(other) > 0
}

// IsLessThan checks if this size is less than another
func (fs *FileSize) IsLessThan(other *FileSize) bool {
	return fs.Compare(other) < 0
}

// ToKB returns the size in kilobytes (rounded)
func (fs *FileSize) ToKB() float64 {
	return float64(fs.bytes) / float64(KB)
}

// ToMB returns the size in megabytes (rounded)
func (fs *FileSize) ToMB() float64 {
	return float64(fs.bytes) / float64(MB)
}

// ToGB returns the size in gigabytes (rounded)
func (fs *FileSize) ToGB() float64 {
	return float64(fs.bytes) / float64(GB)
}

// IsWithinLimit checks if the size is within the given limit
func (fs *FileSize) IsWithinLimit(limit int64) bool {
	return fs.bytes <= limit
}

// ExceedsLimit checks if the size exceeds the given limit
func (fs *FileSize) ExceedsLimit(limit int64) bool {
	return fs.bytes > limit
}

// String returns a string representation of the file size
func (fs *FileSize) String() string {
	return fs.HumanReadable()
}

// Percentage calculates what percentage this size is of the total
func (fs *FileSize) Percentage(total *FileSize) float64 {
	if total == nil || total.bytes == 0 {
		return 0.0
	}

	return float64(fs.bytes) / float64(total.bytes) * 100.0
}

// Validate performs validation checks on the file size
func (fs *FileSize) Validate(maxSize int64) error {
	if fs.bytes < 0 {
		return errors.New("file size cannot be negative")
	}

	if maxSize > 0 && fs.bytes > maxSize {
		return fmt.Errorf("file size %s exceeds maximum allowed size", fs.HumanReadable())
	}

	return nil
}

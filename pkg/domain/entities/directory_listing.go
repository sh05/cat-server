package entities

import (
	"errors"
	"sort"
	"time"
)

// DirectoryListing represents a collection of filesystem entries in a directory
type DirectoryListing struct {
	path       string
	entries    []FileSystemEntry
	totalCount int
	scannedAt  time.Time
}

// NewDirectoryListing creates a new DirectoryListing with validation
func NewDirectoryListing(path string, entries []FileSystemEntry) (*DirectoryListing, error) {
	if path == "" {
		return nil, errors.New("directory path cannot be empty")
	}

	if entries == nil {
		return nil, errors.New("entries cannot be nil")
	}

	return &DirectoryListing{
		path:       path,
		entries:    entries,
		totalCount: len(entries),
		scannedAt:  time.Now(),
	}, nil
}

// Path returns the directory path
func (d *DirectoryListing) Path() string {
	return d.path
}

// Entries returns all filesystem entries
func (d *DirectoryListing) Entries() []FileSystemEntry {
	// Return a copy to prevent external modification
	entriesCopy := make([]FileSystemEntry, len(d.entries))
	copy(entriesCopy, d.entries)
	return entriesCopy
}

// TotalCount returns the total number of entries
func (d *DirectoryListing) TotalCount() int {
	return d.totalCount
}

// ScannedAt returns when the directory was scanned
func (d *DirectoryListing) ScannedAt() time.Time {
	return d.scannedAt
}

// FilterByType returns entries filtered by type (file or directory)
func (d *DirectoryListing) FilterByType(isDir bool) []FileSystemEntry {
	var filtered []FileSystemEntry
	for _, entry := range d.entries {
		if entry.IsDir() == isDir {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterByPattern returns entries whose names match the given pattern
func (d *DirectoryListing) FilterByPattern(pattern string) []FileSystemEntry {
	var filtered []FileSystemEntry
	for _, entry := range d.entries {
		// Simple pattern matching - can be enhanced with regex
		if pattern == "*" || entry.Name() == pattern {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterHidden returns entries that are not hidden (don't start with .)
func (d *DirectoryListing) FilterHidden() []FileSystemEntry {
	var filtered []FileSystemEntry
	for _, entry := range d.entries {
		if !entry.IsHidden() {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// SortByName returns entries sorted alphabetically by name
func (d *DirectoryListing) SortByName() []FileSystemEntry {
	entries := d.Entries() // Get a copy
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	return entries
}

// SortBySize returns entries sorted by size (ascending)
func (d *DirectoryListing) SortBySize() []FileSystemEntry {
	entries := d.Entries() // Get a copy
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size() < entries[j].Size()
	})
	return entries
}

// SortByModTime returns entries sorted by modification time (newest first)
func (d *DirectoryListing) SortByModTime() []FileSystemEntry {
	entries := d.Entries() // Get a copy
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ModTime().After(entries[j].ModTime())
	})
	return entries
}

// GetFileCount returns the number of files (non-directories)
func (d *DirectoryListing) GetFileCount() int {
	count := 0
	for _, entry := range d.entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}

// GetDirectoryCount returns the number of directories
func (d *DirectoryListing) GetDirectoryCount() int {
	count := 0
	for _, entry := range d.entries {
		if entry.IsDir() {
			count++
		}
	}
	return count
}

// GetTotalSize returns the total size of all files in bytes
func (d *DirectoryListing) GetTotalSize() int64 {
	var total int64
	for _, entry := range d.entries {
		if !entry.IsDir() {
			total += entry.Size()
		}
	}
	return total
}

// IsEmpty returns true if the directory contains no entries
func (d *DirectoryListing) IsEmpty() bool {
	return d.totalCount == 0
}

package services

import (
	"fmt"
	"time"

	"github.com/sh05/cat-server/pkg/domain/entities"
	"github.com/sh05/cat-server/pkg/domain/repositories"
	"github.com/sh05/cat-server/pkg/domain/valueobjects"
	"github.com/sh05/cat-server/pkg/infrastructure/logging"
)

// DirectoryService provides use cases for directory operations
type DirectoryService struct {
	fileSystemRepo repositories.FileSystemRepository
	logger         *logging.Logger
}

// NewDirectoryService creates a new DirectoryService
func NewDirectoryService(fileSystemRepo repositories.FileSystemRepository, logger *logging.Logger) *DirectoryService {
	return &DirectoryService{
		fileSystemRepo: fileSystemRepo,
		logger:         logger,
	}
}

// ListDirectoryRequest represents a request to list directory contents
type ListDirectoryRequest struct {
	Path          string
	IncludeHidden bool
	SortBy        string // "name", "size", "modtime"
	SortOrder     string // "asc", "desc"
	FilterType    string // "all", "files", "directories"
}

// ListDirectoryResponse represents the response from listing directory contents
type ListDirectoryResponse struct {
	Path       string                  `json:"path"`
	Files      []FileEntryDTO          `json:"files"`
	TotalCount int                     `json:"totalCount"`
	FileCount  int                     `json:"fileCount"`
	DirCount   int                     `json:"dirCount"`
	TotalSize  int64                   `json:"totalSize"`
	ScannedAt  time.Time               `json:"scannedAt"`
	Statistics *DirectoryStatisticsDTO `json:"statistics,omitempty"`
}

// FileEntryDTO represents a file entry for API responses
type FileEntryDTO struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	SizeHuman    string    `json:"sizeHuman"`
	ModTime      time.Time `json:"modTime"`
	IsDir        bool      `json:"isDir"`
	Permissions  string    `json:"permissions"`
	IsHidden     bool      `json:"isHidden"`
	IsExecutable bool      `json:"isExecutable"`
	IsReadable   bool      `json:"isReadable"`
	IsWritable   bool      `json:"isWritable"`
}

// DirectoryStatisticsDTO represents directory statistics
type DirectoryStatisticsDTO struct {
	LargestFile *FileEntryDTO `json:"largestFile,omitempty"`
	NewestFile  *FileEntryDTO `json:"newestFile,omitempty"`
	OldestFile  *FileEntryDTO `json:"oldestFile,omitempty"`
}

// ListDirectory lists the contents of a directory
func (s *DirectoryService) ListDirectory(request *ListDirectoryRequest) (*ListDirectoryResponse, error) {
	start := time.Now()

	// Validate and create file path
	filePath, err := valueobjects.NewFilePath(request.Path)
	if err != nil {
		s.logger.LogFileSystemOperation("list_directory", request.Path, false, time.Since(start), 0)
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Log the operation
	s.logger.LogFileSystemOperation("list_directory", request.Path, true, 0, 0)

	// Get directory listing from repository
	listing, err := s.fileSystemRepo.ListDirectory(filePath)
	if err != nil {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("list_directory", request.Path, false, duration, 0)
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	// Apply filters and sorting
	entries := listing.Entries()

	// Filter hidden files if requested
	if !request.IncludeHidden {
		entries = s.filterHiddenFiles(entries)
	}

	// Filter by type if requested
	switch request.FilterType {
	case "files":
		entries = s.filterByType(entries, false)
	case "directories":
		entries = s.filterByType(entries, true)
		// "all" or default: no additional filtering
	}

	// Sort entries
	entries = s.sortEntries(entries, request.SortBy, request.SortOrder)

	// Convert to DTOs
	fileEntries := make([]FileEntryDTO, len(entries))
	for i, entry := range entries {
		fileEntries[i] = s.convertToFileEntryDTO(entry)
	}

	// Calculate statistics
	stats, err := s.fileSystemRepo.GetDirectoryStats(filePath)
	var statisticsDTO *DirectoryStatisticsDTO
	if err == nil && stats != nil {
		statisticsDTO = s.convertToDirectoryStatisticsDTO(stats)
	}

	response := &ListDirectoryResponse{
		Path:       request.Path,
		Files:      fileEntries,
		TotalCount: len(fileEntries),
		FileCount:  s.countFilesByType(fileEntries, false),
		DirCount:   s.countFilesByType(fileEntries, true),
		TotalSize:  s.calculateTotalSize(fileEntries),
		ScannedAt:  listing.ScannedAt(),
		Statistics: statisticsDTO,
	}

	duration := time.Since(start)
	s.logger.LogFileSystemOperation("list_directory", request.Path, true, duration, response.TotalSize)

	return response, nil
}

// ValidateDirectoryAccess validates if a directory can be accessed
func (s *DirectoryService) ValidateDirectoryAccess(path string) error {
	filePath, err := valueobjects.NewFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return s.fileSystemRepo.ValidatePath(filePath)
}

// GetDirectoryInfo returns basic information about a directory
func (s *DirectoryService) GetDirectoryInfo(path string) (*DirectoryInfoDTO, error) {
	filePath, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if !s.fileSystemRepo.IsDirectory(filePath) {
		return nil, fmt.Errorf("path is not a directory")
	}

	info, err := s.fileSystemRepo.GetFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory info: %w", err)
	}

	stats, err := s.fileSystemRepo.GetDirectoryStats(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory stats: %w", err)
	}

	return &DirectoryInfoDTO{
		Path:             path,
		ModTime:          info.ModTime(),
		Permissions:      info.Permissions().String(),
		TotalFiles:       stats.TotalFiles,
		TotalDirectories: stats.TotalDirectories,
		TotalSize:        stats.TotalSize,
	}, nil
}

// DirectoryInfoDTO represents basic directory information
type DirectoryInfoDTO struct {
	Path             string    `json:"path"`
	ModTime          time.Time `json:"modTime"`
	Permissions      string    `json:"permissions"`
	TotalFiles       int       `json:"totalFiles"`
	TotalDirectories int       `json:"totalDirectories"`
	TotalSize        int64     `json:"totalSize"`
}

// Helper methods

func (s *DirectoryService) filterHiddenFiles(entries []entities.FileSystemEntry) []entities.FileSystemEntry {
	var filtered []entities.FileSystemEntry
	for _, entry := range entries {
		if !entry.IsHidden() {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (s *DirectoryService) filterByType(entries []entities.FileSystemEntry, isDir bool) []entities.FileSystemEntry {
	var filtered []entities.FileSystemEntry
	for _, entry := range entries {
		if entry.IsDir() == isDir {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (s *DirectoryService) sortEntries(entries []entities.FileSystemEntry, sortBy, sortOrder string) []entities.FileSystemEntry {
	// Create a temporary DirectoryListing to use its sorting methods
	listing, err := entities.NewDirectoryListing("temp", entries)
	if err != nil {
		return entries // Return unsorted on error
	}

	var sorted []entities.FileSystemEntry
	switch sortBy {
	case "size":
		sorted = listing.SortBySize()
	case "modtime":
		sorted = listing.SortByModTime()
	case "name":
		fallthrough
	default:
		sorted = listing.SortByName()
	}

	// Reverse if descending order
	if sortOrder == "desc" {
		for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
			sorted[i], sorted[j] = sorted[j], sorted[i]
		}
	}

	return sorted
}

func (s *DirectoryService) convertToFileEntryDTO(entry entities.FileSystemEntry) FileEntryDTO {
	return FileEntryDTO{
		Name:         entry.Name(),
		Size:         entry.Size(),
		SizeHuman:    entry.HumanReadableSize(),
		ModTime:      entry.ModTime(),
		IsDir:        entry.IsDir(),
		Permissions:  entry.Permissions().String(),
		IsHidden:     entry.IsHidden(),
		IsExecutable: entry.IsExecutable(),
		IsReadable:   entry.IsReadable(),
		IsWritable:   entry.IsWritable(),
	}
}

func (s *DirectoryService) convertToDirectoryStatisticsDTO(stats *repositories.DirectoryStats) *DirectoryStatisticsDTO {
	dto := &DirectoryStatisticsDTO{}

	if stats.LargestFile != nil {
		largestDTO := s.convertToFileEntryDTO(*stats.LargestFile)
		dto.LargestFile = &largestDTO
	}

	if stats.NewestFile != nil {
		newestDTO := s.convertToFileEntryDTO(*stats.NewestFile)
		dto.NewestFile = &newestDTO
	}

	if stats.OldestFile != nil {
		oldestDTO := s.convertToFileEntryDTO(*stats.OldestFile)
		dto.OldestFile = &oldestDTO
	}

	return dto
}

func (s *DirectoryService) countFilesByType(entries []FileEntryDTO, isDir bool) int {
	count := 0
	for _, entry := range entries {
		if entry.IsDir == isDir {
			count++
		}
	}
	return count
}

func (s *DirectoryService) calculateTotalSize(entries []FileEntryDTO) int64 {
	var total int64
	for _, entry := range entries {
		if !entry.IsDir {
			total += entry.Size
		}
	}
	return total
}

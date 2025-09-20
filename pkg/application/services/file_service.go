package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/sh05/cat-server/pkg/domain/repositories"
	"github.com/sh05/cat-server/pkg/domain/valueobjects"
	"github.com/sh05/cat-server/pkg/infrastructure/logging"
)

// FileService provides use cases for file operations
type FileService struct {
	fileSystemRepo repositories.FileSystemRepository
	logger         *logging.Logger
}

// NewFileService creates a new FileService
func NewFileService(fileSystemRepo repositories.FileSystemRepository, logger *logging.Logger) *FileService {
	return &FileService{
		fileSystemRepo: fileSystemRepo,
		logger:         logger,
	}
}

// ReadFileRequest represents a request to read a file
type ReadFileRequest struct {
	Filename    string
	MaxSize     int64
	PreviewOnly bool
	PreviewSize int
}

// ReadFileResponse represents the response from reading a file
type ReadFileResponse struct {
	Filename    string    `json:"filename"`
	Content     string    `json:"content"`
	Size        int64     `json:"size"`
	SizeHuman   string    `json:"sizeHuman"`
	ContentType string    `json:"contentType"`
	Encoding    string    `json:"encoding"`
	IsText      bool      `json:"isText"`
	LineCount   int       `json:"lineCount,omitempty"`
	ModTime     time.Time `json:"modTime"`
	ReadAt      time.Time `json:"readAt"`
	IsPreview   bool      `json:"isPreview,omitempty"`
	Hash        uint32    `json:"hash,omitempty"`
}

// FileInfoRequest represents a request for file information
type FileInfoRequest struct {
	Filename string
}

// FileInfoResponse represents file information response
type FileInfoResponse struct {
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	SizeHuman    string    `json:"sizeHuman"`
	ModTime      time.Time `json:"modTime"`
	IsDir        bool      `json:"isDir"`
	Permissions  string    `json:"permissions"`
	IsHidden     bool      `json:"isHidden"`
	IsExecutable bool      `json:"isExecutable"`
	IsReadable   bool      `json:"isReadable"`
	IsWritable   bool      `json:"isWritable"`
	Exists       bool      `json:"exists"`
}

// ReadFile reads the content of a file
func (s *FileService) ReadFile(request *ReadFileRequest) (*ReadFileResponse, error) {
	start := time.Now()

	// Validate and create file path
	filePath, err := valueobjects.NewFilePath(request.Filename)
	if err != nil {
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, time.Since(start), 0)
		s.logger.LogSecurityEvent("invalid_path", request.Filename, "", "", true)
		return nil, fmt.Errorf("invalid filename: %w", err)
	}

	// Log the operation
	s.logger.LogFileSystemOperation("read_file", request.Filename, true, 0, 0)

	// Validate file access
	if err := s.ValidateFileAccess(request.Filename); err != nil {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, duration, 0)
		s.logger.LogSecurityEvent("access_denied", request.Filename, "", "", true)
		return nil, fmt.Errorf("file access validation failed: %w", err)
	}

	// Check if file exists
	if !s.fileSystemRepo.Exists(filePath) {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, duration, 0)
		return nil, fmt.Errorf("file not found: %s", request.Filename)
	}

	// Check if it's actually a file (not a directory)
	if s.fileSystemRepo.IsDirectory(filePath) {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, duration, 0)
		return nil, fmt.Errorf("path is a directory, not a file: %s", request.Filename)
	}

	// Get file information first
	fileInfo, err := s.fileSystemRepo.GetFileInfo(filePath)
	if err != nil {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, duration, 0)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Check file size limits
	if request.MaxSize > 0 && fileInfo.Size() > request.MaxSize {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, duration, fileInfo.Size())
		return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", fileInfo.Size(), request.MaxSize)
	}

	// Read file content
	fileContent, err := s.fileSystemRepo.ReadFile(filePath)
	if err != nil {
		duration := time.Since(start)
		s.logger.LogFileSystemOperation("read_file", request.Filename, false, duration, fileInfo.Size())
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create file size value object
	fileSize, err := valueobjects.NewFileSize(fileContent.Size())
	if err != nil {
		fileSize, _ = valueobjects.NewFileSize(0)
	}

	// Prepare response
	response := &ReadFileResponse{
		Filename:    request.Filename,
		Size:        fileContent.Size(),
		SizeHuman:   fileSize.HumanReadable(),
		ContentType: fileContent.GetContentType(),
		Encoding:    fileContent.Encoding(),
		IsText:      fileContent.IsTextContent(),
		ModTime:     fileContent.Entry().ModTime(),
		ReadAt:      fileContent.ReadAt(),
		Hash:        fileContent.GetContentHash(),
	}

	// Handle content based on request type
	if request.PreviewOnly && request.PreviewSize > 0 {
		response.Content = fileContent.GetPreview(request.PreviewSize)
		response.IsPreview = true
	} else {
		response.Content = fileContent.ContentAsString()
	}

	// Add line count for text files
	if response.IsText {
		response.LineCount = fileContent.GetLineCount()
	}

	duration := time.Since(start)
	s.logger.LogFileSystemOperation("read_file", request.Filename, true, duration, fileContent.Size())

	return response, nil
}

// ValidateFileAccess validates if a file can be accessed safely
func (s *FileService) ValidateFileAccess(filename string) error {
	filePath, err := valueobjects.NewFilePath(filename)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Validate path security
	if err := s.fileSystemRepo.ValidatePath(filePath); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	// Additional security checks
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("filename contains null bytes")
	}

	// Check for potentially dangerous file extensions (optional, based on security policy)
	if s.isDangerousFileType(filename) {
		s.logger.LogSecurityEvent("dangerous_file_access", filename, "", "", true)
		return fmt.Errorf("access to this file type is restricted")
	}

	return nil
}

// GetFileInfo returns information about a file
func (s *FileService) GetFileInfo(request *FileInfoRequest) (*FileInfoResponse, error) {
	start := time.Now()

	// Validate and create file path
	filePath, err := valueobjects.NewFilePath(request.Filename)
	if err != nil {
		s.logger.LogFileSystemOperation("get_file_info", request.Filename, false, time.Since(start), 0)
		return nil, fmt.Errorf("invalid filename: %w", err)
	}

	response := &FileInfoResponse{
		Filename: request.Filename,
		Exists:   s.fileSystemRepo.Exists(filePath),
	}

	if !response.Exists {
		s.logger.LogFileSystemOperation("get_file_info", request.Filename, false, time.Since(start), 0)
		return response, nil
	}

	// Get file information
	fileInfo, err := s.fileSystemRepo.GetFileInfo(filePath)
	if err != nil {
		s.logger.LogFileSystemOperation("get_file_info", request.Filename, false, time.Since(start), 0)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create file size value object
	fileSize, err := valueobjects.NewFileSize(fileInfo.Size())
	if err != nil {
		fileSize, _ = valueobjects.NewFileSize(0)
	}

	// Fill response
	response.Size = fileInfo.Size()
	response.SizeHuman = fileSize.HumanReadable()
	response.ModTime = fileInfo.ModTime()
	response.IsDir = fileInfo.IsDir()
	response.Permissions = fileInfo.Permissions().String()
	response.IsHidden = fileInfo.IsHidden()
	response.IsExecutable = fileInfo.IsExecutable()
	response.IsReadable = fileInfo.IsReadable()
	response.IsWritable = fileInfo.IsWritable()

	duration := time.Since(start)
	s.logger.LogFileSystemOperation("get_file_info", request.Filename, true, duration, fileInfo.Size())

	return response, nil
}

// CheckFileExists checks if a file exists
func (s *FileService) CheckFileExists(filename string) (bool, error) {
	filePath, err := valueobjects.NewFilePath(filename)
	if err != nil {
		return false, fmt.Errorf("invalid filename: %w", err)
	}

	return s.fileSystemRepo.Exists(filePath), nil
}

// GetContentType determines the content type of a file
func (s *FileService) GetContentType(filename string) (string, error) {
	filePath, err := valueobjects.NewFilePath(filename)
	if err != nil {
		return "", fmt.Errorf("invalid filename: %w", err)
	}

	fileContent, err := s.fileSystemRepo.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fileContent.GetContentType(), nil
}

// ValidateFileSize checks if a file size is within limits
func (s *FileService) ValidateFileSize(filename string, maxSize int64) error {
	filePath, err := valueobjects.NewFilePath(filename)
	if err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	fileInfo, err := s.fileSystemRepo.GetFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize, err := valueobjects.NewFileSize(fileInfo.Size())
	if err != nil {
		return fmt.Errorf("invalid file size: %w", err)
	}

	if err := fileSize.Validate(maxSize); err != nil {
		return fmt.Errorf("file size validation failed: %w", err)
	}

	return nil
}

// Helper methods

func (s *FileService) isDangerousFileType(filename string) bool {
	// Define potentially dangerous file extensions
	dangerousExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".scr", ".pif",
		".vbs", ".vbe", ".js", ".jse", ".wsf", ".wsh",
		".msi", ".reg", ".ps1", ".psm1",
	}

	lowerFilename := strings.ToLower(filename)
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}

	return false
}

// GetFilePreview returns a preview of a file's content
func (s *FileService) GetFilePreview(filename string, maxChars int) (string, bool, error) {
	filePath, err := valueobjects.NewFilePath(filename)
	if err != nil {
		return "", false, fmt.Errorf("invalid filename: %w", err)
	}

	fileContent, err := s.fileSystemRepo.ReadFile(filePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read file: %w", err)
	}

	isText := fileContent.IsTextContent()
	preview := fileContent.GetPreview(maxChars)

	return preview, isText, nil
}

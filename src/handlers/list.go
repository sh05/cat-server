package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sh05/cat-server/src/services"
)

// FileListResponse represents the successful response structure
type FileListResponse struct {
	Files       []string  `json:"files"`
	Directory   string    `json:"directory"`
	Count       int       `json:"count"`
	GeneratedAt time.Time `json:"generated_at"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error      string    `json:"error"`
	Path       string    `json:"path,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code"`
}

// ListHandler handles GET requests to the /ls endpoint
func ListHandler(directoryService *services.DirectoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the request
		slog.Info("file list requested",
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"directory", directoryService.GetPath())

		// Set Content-Type header
		w.Header().Set("Content-Type", "application/json")

		// Get file list from directory service
		files, err := directoryService.ListFiles()
		if err != nil {
			handleListError(w, r, err, directoryService.GetPath())
			return
		}

		// Create successful response
		response := FileListResponse{
			Files:       files,
			Directory:   directoryService.GetPath(),
			Count:       len(files),
			GeneratedAt: time.Now().UTC(),
		}

		// Encode and send response
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode JSON response", "error", err)
			handleListError(w, r, err, directoryService.GetPath())
			return
		}

		// Log completion
		duration := time.Since(start)
		slog.Info("file list completed",
			"duration", duration,
			"file_count", len(files),
			"directory", directoryService.GetPath())
	}
}

// handleListError handles errors and sends appropriate HTTP responses
func handleListError(w http.ResponseWriter, r *http.Request, err error, path string) {
	var statusCode int
	var errorMessage string

	// Determine status code based on error type
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist"):
		statusCode = http.StatusBadRequest
		errorMessage = "directory not found"
	case strings.Contains(errStr, "permission denied"):
		statusCode = http.StatusForbidden
		errorMessage = "permission denied"
	case strings.Contains(errStr, "not a directory"):
		statusCode = http.StatusBadRequest
		errorMessage = "path is not a directory"
	case strings.Contains(errStr, "null byte") || strings.Contains(errStr, "path traversal"):
		statusCode = http.StatusBadRequest
		errorMessage = "invalid path"
	case strings.Contains(errStr, "path too long"):
		statusCode = http.StatusBadRequest
		errorMessage = "path too long"
	case strings.Contains(errStr, "empty path"):
		statusCode = http.StatusBadRequest
		errorMessage = "empty path not allowed"
	default:
		statusCode = http.StatusInternalServerError
		errorMessage = "internal server error"
	}

	// Create error response
	errorResponse := ErrorResponse{
		Error:      errorMessage,
		Path:       path,
		Timestamp:  time.Now().UTC(),
		StatusCode: statusCode,
	}

	// Log the error
	slog.Error("file list error",
		"error", err,
		"status_code", statusCode,
		"directory", path,
		"remote_addr", r.RemoteAddr)

	// Set status code and send error response
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		slog.Error("failed to encode error response", "error", encodeErr)
		// Fallback to plain text error
		http.Error(w, errorMessage, statusCode)
	}
}

// Additional helper for checking if error is related to file system permissions
func isPermissionError(err error) bool {
	if pathErr, ok := err.(*os.PathError); ok {
		return pathErr.Err == os.ErrPermission
	}
	return false
}

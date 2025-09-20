package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/sh05/cat-server/src/services"
)

// CatSuccessResponse represents the successful cat response structure
type CatSuccessResponse struct {
	Content     string    `json:"content"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Directory   string    `json:"directory"`
	GeneratedAt time.Time `json:"generated_at"`
}

// CatHandler handles GET requests to the /cat/{filename} endpoint
func CatHandler(directoryService *services.DirectoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract filename from URL path parameter
		filename := r.PathValue("filename")
		if filename == "" {
			handleCatError(w, r, fmt.Errorf("filename parameter required"), "", directoryService.GetPath())
			return
		}

		// Log the request
		slog.Info("file content requested",
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"filename", filename,
			"directory", directoryService.GetPath())

		// Set Content-Type header
		w.Header().Set("Content-Type", "application/json")

		// Read file content from directory service
		content, err := directoryService.ReadFile(filename)
		if err != nil {
			handleCatError(w, r, err, filename, directoryService.GetPath())
			return
		}

		// Create successful response
		response := CatSuccessResponse{
			Content:     string(content),
			Filename:    filename,
			Size:        int64(len(content)),
			Directory:   directoryService.GetPath(),
			GeneratedAt: time.Now().UTC(),
		}

		// Encode and send response
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode JSON response", "error", err)
			handleCatError(w, r, err, filename, directoryService.GetPath())
			return
		}

		// Log completion
		duration := time.Since(start)
		slog.Info("file content completed",
			"duration", duration,
			"filename", filename,
			"size", len(content),
			"directory", directoryService.GetPath())
	}
}

// handleCatError handles errors and sends appropriate HTTP responses
func handleCatError(w http.ResponseWriter, r *http.Request, err error, filename, directory string) {
	var statusCode int
	var errorMessage string

	// Determine status code based on error type
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "filename parameter required"):
		statusCode = http.StatusBadRequest
		errorMessage = "filename parameter required"
	case strings.Contains(errStr, "file not found") || strings.Contains(errStr, "does not exist"):
		statusCode = http.StatusNotFound
		errorMessage = "file not found"
	case strings.Contains(errStr, "permission denied"):
		statusCode = http.StatusForbidden
		errorMessage = "permission denied"
	case strings.Contains(errStr, "file too large"):
		statusCode = http.StatusRequestEntityTooLarge
		errorMessage = "file too large"
	case strings.Contains(errStr, "binary file not supported"):
		statusCode = http.StatusUnsupportedMediaType
		errorMessage = "binary file not supported"
	case strings.Contains(errStr, "path traversal") || strings.Contains(errStr, "invalid filename") ||
		 strings.Contains(errStr, "null byte") || strings.Contains(errStr, "outside base directory") ||
		 strings.Contains(errStr, "reserved name"):
		statusCode = http.StatusBadRequest
		errorMessage = "invalid filename"
	case strings.Contains(errStr, "path is a directory"):
		statusCode = http.StatusBadRequest
		errorMessage = "path is a directory, not a file"
	case strings.Contains(errStr, "filename too long"):
		statusCode = http.StatusBadRequest
		errorMessage = "filename too long"
	case strings.Contains(errStr, "empty filename"):
		statusCode = http.StatusBadRequest
		errorMessage = "empty filename not allowed"
	default:
		statusCode = http.StatusInternalServerError
		errorMessage = "internal server error"
	}

	// Create error response (reusing existing ErrorResponse from list.go)
	errorResponse := ErrorResponse{
		Error:      errorMessage,
		Path:       filepath.Join(directory, filename),
		Timestamp:  time.Now().UTC(),
		StatusCode: statusCode,
	}

	// Add filename to response if provided
	if filename != "" {
		// We need to extend ErrorResponse to include filename
		// For now, we'll create a cat-specific error response
		catErrorResponse := struct {
			Error      string    `json:"error"`
			Filename   string    `json:"filename,omitempty"`
			Path       string    `json:"path,omitempty"`
			Timestamp  time.Time `json:"timestamp"`
			StatusCode int       `json:"status_code"`
		}{
			Error:      errorMessage,
			Filename:   filename,
			Path:       filepath.Join(directory, filename),
			Timestamp:  time.Now().UTC(),
			StatusCode: statusCode,
		}

		// Log the error
		slog.Error("file content error",
			"error", err,
			"status_code", statusCode,
			"filename", filename,
			"directory", directory,
			"remote_addr", r.RemoteAddr)

		// Set status code and send error response
		w.WriteHeader(statusCode)
		if encodeErr := json.NewEncoder(w).Encode(catErrorResponse); encodeErr != nil {
			slog.Error("failed to encode error response", "error", encodeErr)
			// Fallback to plain text error
			http.Error(w, errorMessage, statusCode)
		}
		return
	}

	// Log the error
	slog.Error("file content error",
		"error", err,
		"status_code", statusCode,
		"directory", directory,
		"remote_addr", r.RemoteAddr)

	// Set status code and send error response
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		slog.Error("failed to encode error response", "error", encodeErr)
		// Fallback to plain text error
		http.Error(w, errorMessage, statusCode)
	}
}
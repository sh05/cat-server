package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthHandler handles GET /health requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Log the request
	slog.Info("health check requested",
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent())

	// Determine response format based on Accept header
	acceptHeader := r.Header.Get("Accept")

	// Default to JSON if no specific preference
	if acceptHeader == "" || containsJSON(acceptHeader) {
		sendJSONResponse(w, r)
	} else if containsHTML(acceptHeader) {
		sendHTMLResponse(w, r)
	} else if containsText(acceptHeader) {
		sendTextResponse(w, r)
	} else {
		// Default to JSON for unknown Accept headers
		sendJSONResponse(w, r)
	}

	// Log completion
	duration := time.Since(start)
	slog.Info("health check completed",
		"duration", duration,
		"status", "ok")
}

// sendJSONResponse sends JSON formatted health response
func sendJSONResponse(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode JSON health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// sendHTMLResponse sends HTML formatted health response
func sendHTMLResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Health Check</title>
</head>
<body>
    <h1>Server Status: OK</h1>
    <p>Timestamp: ` + time.Now().UTC().Format(time.RFC3339) + `</p>
</body>
</html>`

	if _, err := w.Write([]byte(html)); err != nil {
		slog.Error("failed to write HTML health response", "error", err)
	}
}

// sendTextResponse sends plain text health response
func sendTextResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	text := "OK"
	if _, err := w.Write([]byte(text)); err != nil {
		slog.Error("failed to write text health response", "error", err)
	}
}

// Helper functions to check Accept header
func containsJSON(accept string) bool {
	return contains(accept, "application/json") || contains(accept, "*/*")
}

func containsHTML(accept string) bool {
	return contains(accept, "text/html")
}

func containsText(accept string) bool {
	return contains(accept, "text/plain")
}

func contains(s, substr string) bool {
	// Simple substring search - more reliable for Accept headers
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			// Check if it's a word boundary (comma, semicolon, or start/end)
			if i == 0 || s[i-1] == ',' || s[i-1] == ';' || s[i-1] == ' ' {
				nextPos := i + len(substr)
				if nextPos == len(s) || s[nextPos] == ',' || s[nextPos] == ';' || s[nextPos] == ' ' {
					return true
				}
			}
		}
	}
	return false
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

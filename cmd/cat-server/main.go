package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sh05/cat-server/internal/config"
	"github.com/sh05/cat-server/pkg/application/services"
	"github.com/sh05/cat-server/pkg/infrastructure/filesystem"
	"github.com/sh05/cat-server/pkg/infrastructure/logging"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logLevel logging.LogLevel
	switch cfg.Logging.Level {
	case "debug":
		logLevel = logging.LevelDebug
	case "warn":
		logLevel = logging.LevelWarn
	case "error":
		logLevel = logging.LevelError
	default:
		logLevel = logging.LevelInfo
	}

	logger := logging.NewLogger(logLevel, cfg.Logging.Format)
	logger.SetAsDefault()

	// Log startup
	logger.LogStartup("cat-server", "1.0.0", cfg.Server.Port, "production")

	// Initialize filesystem repository
	fsRepo := filesystem.NewFileSystemRepository(cfg.FileSystem.BaseDirectory, cfg.FileSystem.MaxFileSize)

	// Initialize services
	healthService := services.NewHealthService(fsRepo, logger, "1.0.0")
	directoryService := services.NewDirectoryService(fsRepo, logger)
	fileService := services.NewFileService(fsRepo, logger)

	// Create HTTP server
	mux := http.NewServeMux()

	// Register handlers
	registerHealthHandler(mux, healthService, logger)
	registerListHandler(mux, directoryService, logger)
	registerCatHandler(mux, fileService, logger)

	// Apply middleware
	handler := addMiddleware(mux, logger)

	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server in goroutine
	go func() {
		logger.Info("server started successfully", "addr", cfg.GetServerAddr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(err, "server failed to start", "addr", cfg.GetServerAddr())
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Shutdown server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.LogError(err, "server shutdown failed")
		os.Exit(1)
	}

	logger.LogShutdown("cat-server", healthService.GetUptime())
}

// registerHealthHandler registers the health check handler
func registerHealthHandler(mux *http.ServeMux, healthService *services.HealthService, logger *logging.Logger) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		health, err := healthService.GetSystemHealth()
		if err != nil {
			logger.LogError(err, "health check failed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set content type based on Accept header
		acceptHeader := r.Header.Get("Accept")
		if acceptHeader == "text/html" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><body><h1>Health Status: %s</h1><p>Uptime: %s</p><p>Version: %s</p></body></html>",
				health.Status, health.Uptime, health.Version)
			return
		} else if acceptHeader == "text/plain" {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Status: %s\nUptime: %s\nVersion: %s\n",
				health.Status, health.Uptime, health.Version)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})
}

// registerListHandler registers the file list handler
func registerListHandler(mux *http.ServeMux, directoryService *services.DirectoryService, logger *logging.Logger) {
	mux.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		request := &services.ListDirectoryRequest{
			Path:          ".",
			IncludeHidden: false,
			SortBy:        "name",
			SortOrder:     "asc",
			FilterType:    "all",
		}

		listing, err := directoryService.ListDirectory(request)
		if err != nil {
			logger.LogError(err, "failed to list directory")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(listing)
	})
}

// registerCatHandler registers the file content handler
func registerCatHandler(mux *http.ServeMux, fileService *services.FileService, logger *logging.Logger) {
	mux.HandleFunc("/cat/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract filename from path
		filename := r.URL.Path[5:] // Remove "/cat/" prefix
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		request := &services.ReadFileRequest{
			Filename:    filename,
			MaxSize:     10 * 1024 * 1024, // 10MB limit
			PreviewOnly: false,
		}

		fileContent, err := fileService.ReadFile(request)
		if err != nil {
			logger.LogError(err, "failed to read file", "filename", filename)
			if err.Error() == "file not found: "+filename {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fileContent)
	})
}

// addMiddleware adds common middleware to the handler
func addMiddleware(handler http.Handler, logger *logging.Logger) http.Handler {
	// Add security headers
	securityHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		handler.ServeHTTP(w, r)
	})

	// Add logging middleware
	loggingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.LogHTTPRequest(r.Method, r.URL.Path, r.UserAgent(), r.RemoteAddr)

		// Wrap response writer to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		securityHandler.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		logger.LogHTTPResponse(r.Method, r.URL.Path, wrapper.statusCode, duration, 0)
	})

	return loggingHandler
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

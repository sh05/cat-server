package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sh05/cat-server/pkg/infrastructure/logging"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	logger     *logging.Logger
	mux        *http.ServeMux
	addr       string
}

// NewServer creates a new HTTP server
func NewServer(addr string, logger *logging.Logger) *Server {
	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		mux:        mux,
		addr:       addr,
	}
}

// RegisterHandler registers a handler for the given pattern
func (s *Server) RegisterHandler(pattern string, handler http.Handler) {
	// Wrap the handler with logging middleware
	wrappedHandler := s.loggingMiddleware(handler)
	s.mux.Handle(pattern, wrappedHandler)
}

// RegisterHandlerFunc registers a handler function for the given pattern
func (s *Server) RegisterHandlerFunc(pattern string, handlerFunc http.HandlerFunc) {
	s.RegisterHandler(pattern, handlerFunc)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.LogStartup("cat-server", "1.0.0", s.addr, "production")
	s.logger.Info("starting HTTP server", "addr", s.addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.LogError(err, "server startup failed", "addr", s.addr)
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server", "addr", s.addr)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.LogError(err, "server shutdown failed", "addr", s.addr)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.LogShutdown("cat-server", time.Since(time.Now()))
	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.addr
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// loggingMiddleware wraps handlers with request/response logging
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the incoming request
		s.logger.LogHTTPRequest(
			r.Method,
			r.URL.Path,
			r.UserAgent(),
			r.RemoteAddr,
		)

		// Create a response writer wrapper to capture status and size
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			responseSize:   0,
		}

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// Log the response
		duration := time.Since(start)
		s.logger.LogHTTPResponse(
			r.Method,
			r.URL.Path,
			wrapper.statusCode,
			duration,
			wrapper.responseSize,
		)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture response details
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

// WriteHeader captures the status code
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response size
func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.responseSize += int64(n)
	return n, err
}

// SecurityMiddleware provides basic security headers and validations
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Basic path validation
		if len(r.URL.Path) > 1000 {
			http.Error(w, "Path too long", http.StatusRequestURITooLong)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"remote_addr", r.RemoteAddr,
					)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// MethodMiddleware ensures only specified HTTP methods are allowed
func MethodMiddleware(allowedMethods ...string) func(http.Handler) http.Handler {
	methodMap := make(map[string]bool)
	for _, method := range allowedMethods {
		methodMap[method] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !methodMap[r.Method] {
				w.Header().Set("Allow", fmt.Sprintf("%v", allowedMethods))
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware sets the appropriate content type
func ContentTypeMiddleware(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware adds timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request Timeout")
	}
}

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

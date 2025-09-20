package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/sh05/cat-server/src/handlers"
	"github.com/sh05/cat-server/src/services"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	addr       string
}

// New creates a new server instance
func New(addr string, directoryService *services.DirectoryService) *Server {
	mux := http.NewServeMux()

	// Register health endpoint
	mux.HandleFunc("GET /health", handlers.HealthHandler)

	// Register ls endpoint
	mux.HandleFunc("GET /ls", handlers.ListHandler(directoryService))

	// Register cat endpoint
	mux.HandleFunc("GET /cat/{filename}", handlers.CatHandler(directoryService))

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		addr:       addr,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	slog.Info("starting server", "addr", s.addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down server")
	return s.httpServer.Shutdown(ctx)
}

package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sh05/cat-server/src/server"
	"github.com/sh05/cat-server/src/services"
)

func main() {
	// Parse command line flags
	dirFlag := flag.String("dir", "./files/", "Directory to list files from")
	flag.Parse()

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Create directory service
	directoryService, err := services.NewDirectoryService(*dirFlag)
	if err != nil {
		slog.Error("failed to create directory service", "error", err, "directory", *dirFlag)
		os.Exit(1)
	}

	// Log startup configuration
	slog.Info("starting cat-server", "directory", *dirFlag)

	// Create server with directory service
	srv := server.New(":8080", directoryService)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started successfully", "addr", ":8080")

	// Wait for interrupt signal
	<-ctx.Done()

	// Shutdown server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server shutdown completed")
}

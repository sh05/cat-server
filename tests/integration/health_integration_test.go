package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/sh05/cat-server/src/handlers"
	"github.com/sh05/cat-server/src/server"
	"github.com/sh05/cat-server/src/services"
)

func TestHealthEndpointIntegration(t *testing.T) {
	// Create a dummy directory service for health endpoint test
	dummyService, err := services.NewDirectoryService("./files/")
	if err != nil {
		t.Fatalf("Failed to create directory service: %v", err)
	}

	// Start test server on fixed test port
	srv := server.New(":8081", dummyService)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("server failed to start: %v", err)
		}
	}()

	// Cleanup: shutdown server after test
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Make HTTP request to health endpoint
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check response body
	var healthResp handlers.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if healthResp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", healthResp.Status)
	}

	if healthResp.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	dummyService, err := services.NewDirectoryService("./files/")
	if err != nil {
		t.Fatalf("Failed to create directory service: %v", err)
	}
	srv := server.New(":8082", dummyService)

	// Start server
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err = srv.Shutdown(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("graceful shutdown failed: %v", err)
	}

	// Should shutdown quickly since no active connections
	maxDuration := 1 * time.Second
	if duration > maxDuration {
		t.Errorf("shutdown took too long: %v > %v", duration, maxDuration)
	}
}

func TestConcurrentHealthRequests(t *testing.T) {
	dummyService, err := services.NewDirectoryService("./files/")
	if err != nil {
		t.Fatalf("Failed to create directory service: %v", err)
	}
	srv := server.New(":8083", dummyService)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("server failed to start: %v", err)
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Make concurrent requests
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get("http://localhost:8083/health")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- err
				return
			}

			results <- nil
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("concurrent request %d failed: %v", i+1, err)
		}
	}
}

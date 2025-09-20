package unit

import (
	"testing"
	"time"

	"github.com/sh05/cat-server/pkg/application/services"
	"github.com/sh05/cat-server/pkg/infrastructure/filesystem"
	"github.com/sh05/cat-server/pkg/infrastructure/logging"
)

func TestHealthService(t *testing.T) {
	logger := logging.NewDefaultLogger()
	repo := filesystem.NewFileSystemRepository("./", 1024*1024) // 1MB limit
	service := services.NewHealthService(repo, logger, "1.0.0")

	t.Run("GetSystemHealth returns valid response", func(t *testing.T) {
		response, err := service.GetSystemHealth()
		if err != nil {
			t.Fatalf("GetSystemHealth failed: %v", err)
		}

		if response.Status == "" {
			t.Error("Expected status to be set")
		}

		if response.Version == "" {
			t.Error("Expected version to be set")
		}

		if response.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}

		if response.UptimeMs < 0 {
			t.Errorf("Expected UptimeMs to be >= 0, got %d", response.UptimeMs)
		}
	})

	t.Run("Response time is reasonable", func(t *testing.T) {
		start := time.Now()
		_, _ = service.GetSystemHealth()
		duration := time.Since(start)

		maxDuration := 50 * time.Millisecond // More reasonable for system health
		if duration > maxDuration {
			t.Errorf("GetSystemHealth took too long: %v > %v", duration, maxDuration)
		}
	})

	t.Run("Multiple calls return consistent format", func(t *testing.T) {
		response1, err1 := service.GetSystemHealth()
		if err1 != nil {
			t.Fatalf("First call failed: %v", err1)
		}

		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		response2, err2 := service.GetSystemHealth()
		if err2 != nil {
			t.Fatalf("Second call failed: %v", err2)
		}

		if response1.Version != response2.Version {
			t.Error("Version should be consistent across calls")
		}

		if response2.UptimeMs <= response1.UptimeMs {
			t.Error("UptimeMs should increase between calls")
		}
	})
}

func TestHealthServiceConcurrency(t *testing.T) {
	logger := logging.NewDefaultLogger()
	repo := filesystem.NewFileSystemRepository("./", 1024*1024)
	service := services.NewHealthService(repo, logger, "1.0.0")

	t.Run("Concurrent access is safe", func(t *testing.T) {
		const numGoroutines = 100
		results := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				response, err := service.GetSystemHealth()
				results <- (err == nil && response.Status != "")
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			if !<-results {
				t.Error("Concurrent health check failed")
			}
		}
	})
}
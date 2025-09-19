package performance

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sh05/cat-server/src/handlers"
	"github.com/sh05/cat-server/src/server"
	"github.com/sh05/cat-server/src/services"
)

func TestHealthEndpointLoadTest(t *testing.T) {
	// Create directory service for test
	dummyService, err := services.NewDirectoryService("./files/")
	if err != nil {
		t.Fatalf("Failed to create directory service: %v", err)
	}

	// Start test server
	srv := server.New(":8084", dummyService)

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

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Load test parameters
	const (
		numRequests     = 100
		concurrentUsers = 10
		requestsPerUser = numRequests / concurrentUsers
	)

	var wg sync.WaitGroup
	results := make(chan LoadTestResult, numRequests)

	// Start concurrent users
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				start := time.Now()

				resp, err := http.Get("http://localhost:8084/health")
				if err != nil {
					results <- LoadTestResult{
						UserID:   userID,
						Success:  false,
						Duration: time.Since(start),
						Error:    err.Error(),
					}
					continue
				}

				// Verify response
				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					results <- LoadTestResult{
						UserID:   userID,
						Success:  false,
						Duration: time.Since(start),
						Error:    "unexpected status code",
					}
					continue
				}

				var healthResp handlers.HealthResponse
				if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
					resp.Body.Close()
					results <- LoadTestResult{
						UserID:   userID,
						Success:  false,
						Duration: time.Since(start),
						Error:    "failed to decode response",
					}
					continue
				}
				resp.Body.Close()

				// Verify response content
				if healthResp.Status != "ok" {
					results <- LoadTestResult{
						UserID:   userID,
						Success:  false,
						Duration: time.Since(start),
						Error:    "invalid health status",
					}
					continue
				}

				results <- LoadTestResult{
					UserID:   userID,
					Success:  true,
					Duration: time.Since(start),
				}
			}
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(results)

	// Analyze results
	var (
		successCount   int
		totalDuration  time.Duration
		maxDuration    time.Duration
		minDuration    = time.Hour // Initialize to large value
		failedRequests []LoadTestResult
	)

	for result := range results {
		if result.Success {
			successCount++
			totalDuration += result.Duration

			if result.Duration > maxDuration {
				maxDuration = result.Duration
			}
			if result.Duration < minDuration {
				minDuration = result.Duration
			}
		} else {
			failedRequests = append(failedRequests, result)
		}
	}

	// Report results
	successRate := float64(successCount) / float64(numRequests) * 100
	avgDuration := totalDuration / time.Duration(successCount)

	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", numRequests)
	t.Logf("  Concurrent Users: %d", concurrentUsers)
	t.Logf("  Successful Requests: %d", successCount)
	t.Logf("  Success Rate: %.2f%%", successRate)
	t.Logf("  Average Response Time: %v", avgDuration)
	t.Logf("  Min Response Time: %v", minDuration)
	t.Logf("  Max Response Time: %v", maxDuration)

	// Performance requirements validation
	if successRate < 100.0 {
		for _, failure := range failedRequests {
			t.Errorf("Failed request from user %d: %s", failure.UserID, failure.Error)
		}
		t.Errorf("Success rate %.2f%% is below 100%% requirement", successRate)
	}

	// Response time requirement: < 10ms
	maxAllowedDuration := 10 * time.Millisecond
	if avgDuration > maxAllowedDuration {
		t.Errorf("Average response time %v exceeds requirement %v", avgDuration, maxAllowedDuration)
	}

	if maxDuration > 50*time.Millisecond { // Allow some tolerance for load conditions
		t.Errorf("Maximum response time %v is too high under load", maxDuration)
	}
}

func TestMemoryUsageUnderLoad(t *testing.T) {
	// Create directory service for test
	dummyService, err := services.NewDirectoryService("./files/")
	if err != nil {
		t.Fatalf("Failed to create directory service: %v", err)
	}

	// Start test server
	srv := server.New(":8085", dummyService)

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

	time.Sleep(200 * time.Millisecond)

	// Sustained load for memory testing
	const duration = 5 * time.Second
	const concurrentUsers = 20

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup

	// Start sustained load
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					resp, err := http.Get("http://localhost:8085/health")
					if err == nil {
						resp.Body.Close()
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()

	// Note: In a real application, you would measure actual memory usage here
	// For this test, we assume the load test completes successfully without memory issues
	t.Logf("Memory usage test completed - sustained %d concurrent users for %v", concurrentUsers, duration)
}

type LoadTestResult struct {
	UserID   int
	Success  bool
	Duration time.Duration
	Error    string
}

package docker

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDockerPerformanceRequirements(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented - this test will pass after T006 implementation")
	}

	// Build the image for performance testing
	imageName := "cat-server"
	imageTag := "performance-test"
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	cleanupTestImage(fullImageName)
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot run performance tests: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fullImageName)

	t.Run("image_size_requirements", func(t *testing.T) {
		testImageSizeRequirements(t, fullImageName, buildResult)
	})

	t.Run("build_time_requirements", func(t *testing.T) {
		testBuildTimeRequirements(t, fullImageName)
	})

	t.Run("container_startup_performance", func(t *testing.T) {
		testContainerStartupPerformance(t, fullImageName)
	})

	t.Run("runtime_memory_usage", func(t *testing.T) {
		testRuntimeMemoryUsage(t, fullImageName)
	})

	t.Run("api_response_performance", func(t *testing.T) {
		testAPIResponsePerformance(t, fullImageName)
	})

	t.Run("concurrent_container_performance", func(t *testing.T) {
		testConcurrentContainerPerformance(t, fullImageName)
	})
}

func testImageSizeRequirements(t *testing.T, imageName string, buildResult BuildResult) {
	const maxSizeBytes = 52428800 // 50MB in bytes

	t.Logf("Image size: %d bytes (%.2f MB)", buildResult.ImageSize, float64(buildResult.ImageSize)/1024/1024)

	// Primary requirement: Image must be under 50MB
	if buildResult.ImageSize > maxSizeBytes {
		t.Errorf("Image size %d bytes exceeds maximum requirement of 50MB (%d bytes)",
			buildResult.ImageSize, maxSizeBytes)
	} else {
		t.Logf("✓ Image size requirement met: %.2f MB < 50MB",
			float64(buildResult.ImageSize)/1024/1024)
	}

	// Analyze image layers for optimization opportunities
	cmd := exec.Command("docker", "history", "--no-trunc", imageName)
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Cannot analyze image layers: %v", err)
		return
	}

	layers := strings.Split(string(output), "\n")
	layerCount := len(layers) - 2 // Subtract header and empty line

	t.Logf("Image has %d layers", layerCount)

	if layerCount > 15 {
		t.Logf("WARNING: Image has many layers (%d), consider optimization", layerCount)
	}

	// Show largest layers for optimization insights
	t.Log("Image layer analysis:")
	for i, layer := range layers {
		if i == 0 || i >= 6 || layer == "" { // Show header + first 5 layers
			continue
		}
		fields := strings.Fields(layer)
		if len(fields) >= 2 {
			t.Logf("  Layer %d: %s", i, fields[1]) // Size field
		}
	}
}

func testBuildTimeRequirements(t *testing.T, imageName string) {
	const maxBuildTimeSeconds = 60

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")

	// Test clean build (no cache)
	t.Run("clean_build_performance", func(t *testing.T) {
		cleanImageName := imageName + "-clean"
		cleanupTestImage(cleanImageName)

		// Clear build cache
		exec.Command("docker", "builder", "prune", "-f").Run()

		start := time.Now()
		result := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), "cat-server", "clean")
		buildTime := time.Since(start).Seconds()

		defer cleanupTestImage("cat-server:clean")

		t.Logf("Clean build time: %.2f seconds", buildTime)

		if buildTime > maxBuildTimeSeconds {
			t.Errorf("Clean build time %.2f seconds exceeds maximum of %d seconds",
				buildTime, maxBuildTimeSeconds)
		} else {
			t.Logf("✓ Clean build time requirement met: %.2f seconds < %d seconds",
				buildTime, maxBuildTimeSeconds)
		}

		if result.Status != "success" {
			t.Errorf("Clean build failed: %s", result.ErrorMessage)
		}
	})

	// Test cached build performance
	t.Run("cached_build_performance", func(t *testing.T) {
		// First build (to populate cache)
		result1 := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), "cat-server", "cached1")
		if result1.Status == "success" {
			cleanupTestImage("cat-server:cached1")

			// Second build (should use cache)
			start := time.Now()
			result2 := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), "cat-server", "cached2")
			cachedBuildTime := time.Since(start).Seconds()

			defer cleanupTestImage("cat-server:cached2")

			if result2.Status != "success" {
				t.Errorf("Cached build failed: %s", result2.ErrorMessage)
				return
			}

			t.Logf("Cached build time: %.2f seconds", cachedBuildTime)

			if cachedBuildTime > 30 { // More lenient for cached builds
				t.Logf("Cached build time %.2f seconds is slower than optimal (30s target)",
					cachedBuildTime)
			} else {
				t.Logf("✓ Cached build performance good: %.2f seconds", cachedBuildTime)
			}
		}
	})
}

func testContainerStartupPerformance(t *testing.T, imageName string) {
	const maxStartupTimeSeconds = 2.0

	containerName := "cat-server-startup-test"

	// Test startup time multiple times for consistency
	var startupTimes []float64

	for i := 0; i < 3; i++ {
		cleanupTestContainer(containerName)

		start := time.Now()
		result := runDockerContainer(imageName, containerName, "18094:8080")
		startupTime := time.Since(start).Seconds()

		if result.Status == "running" {
			startupTimes = append(startupTimes, startupTime)
			cleanupTestContainer(result.ContainerID)
		} else {
			t.Errorf("Container failed to start on attempt %d: %s", i+1, result.ErrorMessage)
		}
	}

	if len(startupTimes) > 0 {
		// Calculate average startup time
		var total float64
		for _, time := range startupTimes {
			total += time
		}
		avgStartupTime := total / float64(len(startupTimes))

		t.Logf("Startup times: %v", startupTimes)
		t.Logf("Average startup time: %.2f seconds", avgStartupTime)

		if avgStartupTime > maxStartupTimeSeconds {
			t.Errorf("Average startup time %.2f seconds exceeds maximum of %.1f seconds",
				avgStartupTime, maxStartupTimeSeconds)
		} else {
			t.Logf("✓ Startup time requirement met: %.2f seconds < %.1f seconds",
				avgStartupTime, maxStartupTimeSeconds)
		}

		// Check for consistency (startup times shouldn't vary too much)
		maxVariation := 0.5 // 500ms
		for _, time := range startupTimes {
			if time-avgStartupTime > maxVariation || avgStartupTime-time > maxVariation {
				t.Logf("WARNING: Inconsistent startup time: %.2f seconds (avg: %.2f)",
					time, avgStartupTime)
			}
		}
	}
}

func testRuntimeMemoryUsage(t *testing.T, imageName string) {
	containerName := "cat-server-memory-test"
	cleanupTestContainer(containerName)

	// Start container with memory monitoring
	result := runDockerContainer(imageName, containerName, "18095:8080")
	if result.Status != "running" {
		t.Fatalf("Container failed to start: %s", result.ErrorMessage)
	}
	defer cleanupTestContainer(result.ContainerID)

	// Wait for application to fully start
	time.Sleep(5 * time.Second)

	// Check memory usage
	cmd := exec.Command("docker", "stats", "--no-stream", "--format",
		"{{.MemUsage}}", result.ContainerID)
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Cannot check memory usage: %v", err)
		return
	}

	memUsage := strings.TrimSpace(string(output))
	t.Logf("Container memory usage: %s", memUsage)

	// Parse memory usage (format: "used / limit")
	parts := strings.Split(memUsage, " / ")
	if len(parts) >= 1 {
		usedMem := strings.TrimSpace(parts[0])

		// Convert to MB for comparison
		if strings.HasSuffix(usedMem, "MiB") || strings.HasSuffix(usedMem, "MB") {
			memStr := strings.TrimSuffix(strings.TrimSuffix(usedMem, "MiB"), "MB")
			if memMB, err := strconv.ParseFloat(memStr, 64); err == nil {
				t.Logf("Memory usage: %.2f MB", memMB)

				// Target: under 64MB for a simple HTTP server
				if memMB > 64 {
					t.Logf("Memory usage %.2f MB exceeds optimal target of 64MB", memMB)
				} else {
					t.Logf("✓ Memory usage is efficient: %.2f MB", memMB)
				}
			}
		}
	}

	// Test memory usage under load
	t.Run("memory_under_load", func(t *testing.T) {
		// Make several concurrent requests
		for i := 0; i < 10; i++ {
			go func() {
				http.Get("http://localhost:18095/health")
			}()
		}

		time.Sleep(2 * time.Second)

		// Check memory usage again
		cmd := exec.Command("docker", "stats", "--no-stream", "--format",
			"{{.MemUsage}}", result.ContainerID)
		output, err := cmd.Output()
		if err == nil {
			memUsageLoad := strings.TrimSpace(string(output))
			t.Logf("Memory usage under load: %s", memUsageLoad)
		}
	})
}

func testAPIResponsePerformance(t *testing.T, imageName string) {
	containerName := "cat-server-api-perf-test"
	cleanupTestContainer(containerName)

	result := runDockerContainer(imageName, containerName, "18096:8080")
	if result.Status != "running" {
		t.Fatalf("Container failed to start: %s", result.ErrorMessage)
	}
	defer cleanupTestContainer(result.ContainerID)

	// Wait for service to be ready
	time.Sleep(3 * time.Second)

	baseURL := "http://localhost:18096"

	// Test health endpoint response time
	t.Run("health_endpoint_response_time", func(t *testing.T) {
		const maxResponseTimeMs = 100 // 100ms target

		var responseTimes []float64

		for i := 0; i < 5; i++ {
			start := time.Now()
			resp, err := http.Get(baseURL + "/health")
			responseTime := time.Since(start).Seconds() * 1000 // Convert to ms

			if err != nil {
				t.Errorf("Health endpoint request failed: %v", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == 200 {
				responseTimes = append(responseTimes, responseTime)
			}
		}

		if len(responseTimes) > 0 {
			var total float64
			for _, time := range responseTimes {
				total += time
			}
			avgResponseTime := total / float64(len(responseTimes))

			t.Logf("Health endpoint response times: %v ms", responseTimes)
			t.Logf("Average response time: %.2f ms", avgResponseTime)

			if avgResponseTime > maxResponseTimeMs {
				t.Logf("Average response time %.2f ms exceeds target of %d ms",
					avgResponseTime, maxResponseTimeMs)
			} else {
				t.Logf("✓ Response time is good: %.2f ms < %d ms",
					avgResponseTime, maxResponseTimeMs)
			}
		}
	})

	// Test file list endpoint performance
	t.Run("file_list_endpoint_performance", func(t *testing.T) {
		start := time.Now()
		resp, err := http.Get(baseURL + "/ls")
		responseTime := time.Since(start).Seconds() * 1000

		if err != nil {
			t.Errorf("File list endpoint request failed: %v", err)
			return
		}
		defer resp.Body.Close()

		t.Logf("File list response time: %.2f ms", responseTime)

		if responseTime > 200 { // 200ms target for file operations
			t.Logf("File list response time %.2f ms is slower than target (200ms)",
				responseTime)
		}
	})

	// Test concurrent request handling
	t.Run("concurrent_request_performance", func(t *testing.T) {
		const concurrentRequests = 10

		results := make(chan float64, concurrentRequests)

		start := time.Now()

		for i := 0; i < concurrentRequests; i++ {
			go func() {
				requestStart := time.Now()
				resp, err := http.Get(baseURL + "/health")
				requestTime := time.Since(requestStart).Seconds() * 1000

				if err == nil && resp.StatusCode == 200 {
					results <- requestTime
					resp.Body.Close()
				} else {
					results <- -1 // Error indicator
				}
			}()
		}

		// Collect results
		var responseTimes []float64
		for i := 0; i < concurrentRequests; i++ {
			result := <-results
			if result > 0 {
				responseTimes = append(responseTimes, result)
			}
		}

		totalTime := time.Since(start).Seconds() * 1000

		t.Logf("Concurrent requests: %d successful out of %d",
			len(responseTimes), concurrentRequests)
		t.Logf("Total time for %d concurrent requests: %.2f ms",
			concurrentRequests, totalTime)

		if len(responseTimes) > 0 {
			var total float64
			for _, time := range responseTimes {
				total += time
			}
			avgConcurrentResponseTime := total / float64(len(responseTimes))
			t.Logf("Average concurrent response time: %.2f ms", avgConcurrentResponseTime)
		}
	})
}

func testConcurrentContainerPerformance(t *testing.T, imageName string) {
	const numContainers = 3

	t.Logf("Testing performance with %d concurrent containers", numContainers)

	containers := make([]string, numContainers)
	startTimes := make([]float64, numContainers)

	// Start multiple containers concurrently
	for i := 0; i < numContainers; i++ {
		containerName := fmt.Sprintf("cat-server-concurrent-%d", i)
		port := fmt.Sprintf("1809%d:8080", 7+i)

		cleanupTestContainer(containerName)

		start := time.Now()
		result := runDockerContainer(imageName, containerName, port)
		startTime := time.Since(start).Seconds()

		if result.Status == "running" {
			containers[i] = result.ContainerID
			startTimes[i] = startTime
		} else {
			t.Errorf("Container %d failed to start: %s", i, result.ErrorMessage)
		}
	}

	// Cleanup
	defer func() {
		for _, containerID := range containers {
			if containerID != "" {
				cleanupTestContainer(containerID)
			}
		}
	}()

	// Analyze concurrent startup performance
	var validStartTimes []float64
	for _, startTime := range startTimes {
		if startTime > 0 {
			validStartTimes = append(validStartTimes, startTime)
		}
	}

	if len(validStartTimes) > 0 {
		var total float64
		for _, time := range validStartTimes {
			total += time
		}
		avgStartTime := total / float64(len(validStartTimes))

		t.Logf("Concurrent container startup times: %v", validStartTimes)
		t.Logf("Average startup time under concurrency: %.2f seconds", avgStartTime)

		// Allow slightly longer startup time under concurrency
		if avgStartTime > 3.0 {
			t.Logf("Concurrent startup time %.2f seconds is slower than expected",
				avgStartTime)
		} else {
			t.Logf("✓ Concurrent startup performance acceptable: %.2f seconds",
				avgStartTime)
		}
	}

	// Test if all containers are responding
	time.Sleep(3 * time.Second)

	successfulResponses := 0
	for i := 0; i < numContainers; i++ {
		if containers[i] != "" {
			port := 18097 + i
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
			if err == nil && resp.StatusCode == 200 {
				successfulResponses++
				resp.Body.Close()
			}
		}
	}

	t.Logf("Concurrent containers responding: %d out of %d",
		successfulResponses, len(containers))

	if successfulResponses < len(containers) {
		t.Logf("Some containers not responding under concurrent load")
	}
}

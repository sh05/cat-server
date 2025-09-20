package docker

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDockerRunContract(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	// Skip if Dockerfile doesn't exist yet
	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented - this test will pass after T006 implementation")
	}

	// Ensure we have an image to test with
	imageName := "cat-server"
	imageTag := "run-test"
	cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))

	// Build image first
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot test run: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))

	tests := []struct {
		name          string
		imageName     string
		containerName string
		portMapping   string
		wantErr       bool
	}{
		{
			name:          "successful container run",
			imageName:     fmt.Sprintf("%s:%s", imageName, imageTag),
			containerName: "cat-server-run-test",
			portMapping:   "18080:8080", // Use different port to avoid conflicts
			wantErr:       false,
		},
		{
			name:          "run non-existent image",
			imageName:     "non-existent:latest",
			containerName: "test-fail",
			portMapping:   "18081:8080",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing container
			cleanupTestContainer(tt.containerName)

			result := runDockerContainer(tt.imageName, tt.containerName, tt.portMapping)

			if tt.wantErr {
				if result.Status != "failed" {
					t.Errorf("Expected container run to fail, but got status: %s", result.Status)
					cleanupTestContainer(result.ContainerID)
				}
				return
			}

			// Verify successful run
			if result.Status != "running" {
				t.Errorf("Expected container to be running, but got status: %s, error: %s", result.Status, result.ErrorMessage)
				return
			}

			// Contract requirements verification
			if result.StartupTime > 2 {
				t.Errorf("Startup time %.2f seconds exceeds maximum of 2 seconds", result.StartupTime)
			}

			if result.ContainerID == "" {
				t.Error("Expected valid container ID, got empty string")
			}

			// Wait a moment for the application to fully start
			time.Sleep(2 * time.Second)

			// Test health endpoint if container is running
			if result.Status == "running" {
				port := strings.Split(tt.portMapping, ":")[0]
				healthURL := fmt.Sprintf("http://localhost:%s/health", port)

				// Give the service a moment to start
				for i := 0; i < 10; i++ {
					resp, err := http.Get(healthURL)
					if err == nil && resp.StatusCode == 200 {
						resp.Body.Close()
						t.Logf("Health check successful on attempt %d", i+1)
						break
					}
					if resp != nil {
						resp.Body.Close()
					}
					if i == 9 {
						t.Logf("Health check failed after 10 attempts, last error: %v", err)
					}
					time.Sleep(500 * time.Millisecond)
				}
			}

			// Clean up
			cleanupTestContainer(result.ContainerID)
		})
	}
}

func TestDockerRunPerformance(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented")
	}

	imageName := "cat-server"
	imageTag := "perf-run-test"
	containerName := "cat-server-perf-test"

	// Clean up
	cleanupTestContainer(containerName)
	cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))

	// Build image
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Fatalf("Build failed: %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))

	// Test startup performance
	result := runDockerContainer(fmt.Sprintf("%s:%s", imageName, imageTag), containerName, "18082:8080")
	defer cleanupTestContainer(result.ContainerID)

	if result.Status != "running" {
		t.Fatalf("Container failed to start: %s", result.ErrorMessage)
	}

	t.Logf("Container startup time: %.2f seconds", result.StartupTime)

	if result.StartupTime > 2 {
		t.Errorf("Startup time %.2f seconds exceeds 2 second requirement", result.StartupTime)
	}

	// Test health check response time
	healthURL := "http://localhost:18082/health"
	start := time.Now()

	var healthCheckSuccess bool
	for i := 0; i < 20; i++ { // Wait up to 10 seconds
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			healthCheckSuccess = true
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	healthResponseTime := time.Since(start).Seconds()
	t.Logf("Health check response time: %.2f seconds", healthResponseTime)

	if !healthCheckSuccess {
		t.Error("Health check failed to respond within 10 seconds")
	} else if healthResponseTime > 3 {
		t.Errorf("Health check response time %.2f seconds exceeds 3 second target", healthResponseTime)
	}
}

// RunResult represents the result of a Docker run operation
type RunResult struct {
	Status        string  `json:"status"`
	ContainerID   string  `json:"container_id,omitempty"`
	StartupTime   float64 `json:"startup_time,omitempty"`
	HealthStatus  string  `json:"health_status,omitempty"`
	ErrorMessage  string  `json:"error_message,omitempty"`
	ContainerLogs string  `json:"container_logs,omitempty"`
}

// runDockerContainer executes docker run and returns structured result
func runDockerContainer(imageName, containerName, portMapping string) RunResult {
	start := time.Now()

	cmd := exec.Command("docker", "run", "-d", "--rm", "--name", containerName, "-p", portMapping, imageName)
	output, err := cmd.CombinedOutput()

	startupTime := time.Since(start).Seconds()

	if err != nil {
		return RunResult{
			Status:        "failed",
			ErrorMessage:  err.Error(),
			ContainerLogs: string(output),
			StartupTime:   startupTime,
		}
	}

	containerID := strings.TrimSpace(string(output))

	// Verify container is actually running
	cmd = exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
	statusOutput, err := cmd.Output()
	if err != nil || len(statusOutput) == 0 {
		return RunResult{
			Status:       "failed",
			ErrorMessage: "Container not found after start",
			ContainerID:  containerID,
			StartupTime:  startupTime,
		}
	}

	return RunResult{
		Status:       "running",
		ContainerID:  containerID,
		StartupTime:  startupTime,
		HealthStatus: "starting",
	}
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	cmd := exec.Command("test", "-f", filename)
	return cmd.Run() == nil
}

// cleanupTestContainer removes test containers
func cleanupTestContainer(containerName string) {
	if containerName == "" {
		return
	}

	// Stop container
	cmd := exec.Command("docker", "stop", containerName)
	cmd.Run() // Ignore errors

	// Remove container (--rm should handle this, but just in case)
	cmd = exec.Command("docker", "rm", "-f", containerName)
	cmd.Run() // Ignore errors
}

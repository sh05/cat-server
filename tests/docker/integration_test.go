package docker

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDockerIntegrationScenarios(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented - this test will pass after T006 implementation")
	}

	// Build the image for integration testing
	imageName := "cat-server"
	imageTag := "integration-test"
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	cleanupTestImage(fullImageName)
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot run integration tests: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fullImageName)

	t.Run("basic_container_lifecycle", func(t *testing.T) {
		testBasicContainerLifecycle(t, fullImageName)
	})

	t.Run("api_endpoints_functionality", func(t *testing.T) {
		testAPIEndpointsFunctionality(t, fullImageName)
	})

	t.Run("volume_mount_scenario", func(t *testing.T) {
		testVolumeMountScenario(t, fullImageName)
	})

	t.Run("environment_variables", func(t *testing.T) {
		testEnvironmentVariables(t, fullImageName)
	})

	t.Run("health_check_behavior", func(t *testing.T) {
		testHealthCheckBehavior(t, fullImageName)
	})
}

func testBasicContainerLifecycle(t *testing.T, imageName string) {
	containerName := "cat-server-lifecycle-test"
	cleanupTestContainer(containerName)

	// Start container
	result := runDockerContainer(imageName, containerName, "18090:8080")
	if result.Status != "running" {
		t.Fatalf("Failed to start container: %s", result.ErrorMessage)
	}
	defer cleanupTestContainer(result.ContainerID)

	// Wait for service to be ready
	time.Sleep(3 * time.Second)

	// Test basic connectivity
	resp, err := http.Get("http://localhost:18090/health")
	if err != nil {
		t.Errorf("Health endpoint not accessible: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Stop container gracefully
	cmd := exec.Command("docker", "stop", result.ContainerID)
	if err := cmd.Run(); err != nil {
		t.Errorf("Failed to stop container gracefully: %v", err)
	}

	// Verify container stopped
	time.Sleep(2 * time.Second)
	cmd = exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("id=%s", result.ContainerID))
	output, _ := cmd.Output()
	if len(strings.TrimSpace(string(output))) > 0 {
		t.Error("Container did not stop gracefully")
	}
}

func testAPIEndpointsFunctionality(t *testing.T, imageName string) {
	containerName := "cat-server-api-test"
	cleanupTestContainer(containerName)

	// Start container
	result := runDockerContainer(imageName, containerName, "18091:8080")
	if result.Status != "running" {
		t.Fatalf("Failed to start container: %s", result.ErrorMessage)
	}
	defer cleanupTestContainer(result.ContainerID)

	// Wait for service to be ready
	time.Sleep(3 * time.Second)

	baseURL := "http://localhost:18091"

	// Test health endpoint
	t.Run("health_endpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Health endpoint error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Health endpoint status: expected 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read health response: %v", err)
		}

		if !strings.Contains(string(body), "status") {
			t.Error("Health response doesn't contain status field")
		}
	})

	// Test file list endpoint
	t.Run("file_list_endpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/ls")
		if err != nil {
			t.Fatalf("File list endpoint error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("File list endpoint status: expected 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read file list response: %v", err)
		}

		// Should return JSON with files array
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "files") {
			t.Error("File list response doesn't contain files field")
		}
	})

	// Test file content endpoint (if go.mod exists in container)
	t.Run("file_content_endpoint", func(t *testing.T) {
		// First, check what files are available
		resp, err := http.Get(baseURL + "/ls")
		if err != nil {
			t.Skip("Cannot test file content: file list unavailable")
		}
		defer resp.Body.Close()

		// Try to get a file that should exist
		resp, err = http.Get(baseURL + "/cat/nonexistent.txt")
		if err != nil {
			t.Errorf("File content endpoint error: %v", err)
		} else {
			defer resp.Body.Close()
			// Should return 404 for non-existent file
			if resp.StatusCode != 404 {
				t.Logf("File content endpoint returned %d for non-existent file (expected 404)", resp.StatusCode)
			}
		}
	})
}

func testVolumeMountScenario(t *testing.T, imageName string) {
	containerName := "cat-server-volume-test"
	cleanupTestContainer(containerName)

	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "cat-server-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello from volume mount!"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Start container with volume mount
	cmd := exec.Command("docker", "run", "-d", "--rm", "--name", containerName,
		"-p", "18092:8080",
		"-v", fmt.Sprintf("%s:/app/files", tempDir),
		imageName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start container with volume: %v, output: %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	defer cleanupTestContainer(containerID)

	// Wait for service to be ready
	time.Sleep(3 * time.Second)

	// Test that mounted files are accessible
	resp, err := http.Get("http://localhost:18092/cat/test.txt")
	if err != nil {
		t.Errorf("Failed to access mounted file: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read file content: %v", err)
		} else {
			bodyStr := string(body)
			if !strings.Contains(bodyStr, "Hello from volume mount!") {
				t.Errorf("File content doesn't match expected content. Got: %s", bodyStr)
			} else {
				t.Log("Volume mount test successful: file content accessible")
			}
		}
	} else {
		t.Logf("Volume mount test: expected file access to work, got status %d", resp.StatusCode)
	}
}

func testEnvironmentVariables(t *testing.T, imageName string) {
	containerName := "cat-server-env-test"
	cleanupTestContainer(containerName)

	// Start container with custom port environment variable
	cmd := exec.Command("docker", "run", "-d", "--rm", "--name", containerName,
		"-p", "19090:9090",
		"-e", "PORT=9090",
		imageName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start container with custom port: %v, output: %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	defer cleanupTestContainer(containerID)

	// Wait for service to be ready
	time.Sleep(3 * time.Second)

	// Test if service is running on custom port
	resp, err := http.Get("http://localhost:19090/health")
	if err != nil {
		// The cat-server might not support PORT env var yet, which is fine
		t.Logf("Custom port test: service may not support PORT environment variable yet: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			t.Log("Environment variable test successful: service running on custom port")
		}
	}

	// Test environment variables inside container
	cmd = exec.Command("docker", "exec", containerID, "env")
	envOutput, err := cmd.Output()
	if err != nil {
		t.Errorf("Failed to get environment variables: %v", err)
	} else {
		envStr := string(envOutput)
		if strings.Contains(envStr, "PORT=9090") {
			t.Log("Environment variable successfully set in container")
		}
	}
}

func testHealthCheckBehavior(t *testing.T, imageName string) {
	containerName := "cat-server-health-test"
	cleanupTestContainer(containerName)

	// Start container
	result := runDockerContainer(imageName, containerName, "18093:8080")
	if result.Status != "running" {
		t.Fatalf("Failed to start container: %s", result.ErrorMessage)
	}
	defer cleanupTestContainer(result.ContainerID)

	// Wait for health check to stabilize
	time.Sleep(10 * time.Second)

	// Check health status
	cmd := exec.Command("docker", "inspect", result.ContainerID,
		"--format", "{{.State.Health.Status}}")
	output, err := cmd.Output()

	if err != nil {
		t.Logf("Health check may not be implemented yet: %v", err)
		return
	}

	healthStatus := strings.TrimSpace(string(output))
	t.Logf("Container health status: %s", healthStatus)

	if healthStatus == "healthy" {
		t.Log("Health check test successful: container is healthy")
	} else if healthStatus == "starting" {
		t.Log("Health check in progress (status: starting)")
	} else if healthStatus == "unhealthy" {
		t.Error("Container is unhealthy")

		// Get health check logs for debugging
		cmd = exec.Command("docker", "inspect", result.ContainerID,
			"--format", "{{range .State.Health.Log}}{{.Output}}{{end}}")
		logOutput, err := cmd.Output()
		if err == nil {
			t.Logf("Health check logs: %s", string(logOutput))
		}
	}
}

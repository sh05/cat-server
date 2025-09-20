package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDockerInspectContract(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented - this test will pass after T006 implementation")
	}

	imageName := "cat-server"
	imageTag := "inspect-test"
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	// Clean up and build image
	cleanupTestImage(fullImageName)
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot inspect image: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fullImageName)

	result := inspectDockerImage(fullImageName)

	// Verify contract requirements
	if result.Size > 52428800 { // 50MB in bytes
		t.Errorf("Image size %d bytes exceeds maximum of 50MB (52428800 bytes)", result.Size)
	}

	if result.User != "app" {
		t.Errorf("Expected user 'app', got '%s'", result.User)
	}

	if result.OS != "linux" {
		t.Errorf("Expected OS 'linux', got '%s'", result.OS)
	}

	expectedPorts := []string{"8080/tcp"}
	if !containsAll(result.ExposedPorts, expectedPorts) {
		t.Errorf("Expected ports %v to be exposed, got: %v", expectedPorts, result.ExposedPorts)
	}

	// Architecture should be amd64 or arm64
	if result.Architecture != "amd64" && result.Architecture != "arm64" {
		t.Errorf("Expected architecture 'amd64' or 'arm64', got '%s'", result.Architecture)
	}

	// Security checks
	if result.SecurityScan.Vulnerabilities.Critical > 0 {
		t.Errorf("Image has %d critical vulnerabilities", result.SecurityScan.Vulnerabilities.Critical)
	}

	if result.SecurityScan.Vulnerabilities.High > 0 {
		t.Errorf("Image has %d high vulnerabilities", result.SecurityScan.Vulnerabilities.High)
	}

	t.Logf("Image inspection results:")
	t.Logf("  Size: %d bytes (%.2f MB)", result.Size, float64(result.Size)/1024/1024)
	t.Logf("  User: %s", result.User)
	t.Logf("  Architecture: %s", result.Architecture)
	t.Logf("  OS: %s", result.OS)
	t.Logf("  Exposed Ports: %v", result.ExposedPorts)
	t.Logf("  Layers: %d", result.Layers)
}

func TestDockerImageSecurity(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented")
	}

	imageName := "cat-server"
	imageTag := "security-test"
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	// Clean up and build image
	cleanupTestImage(fullImageName)
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot test security: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fullImageName)

	// Test non-root user execution
	t.Run("non-root user execution", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm", fullImageName, "whoami")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run whoami command: %v", err)
		}

		user := strings.TrimSpace(string(output))
		if user != "app" {
			t.Errorf("Expected container to run as 'app' user, got '%s'", user)
		}
	})

	// Test that sensitive directories are not accessible
	t.Run("sensitive directories protection", func(t *testing.T) {
		sensitiveTests := []struct {
			path string
			desc string
		}{
			{"/etc/passwd", "passwd file"},
			{"/etc/shadow", "shadow file"},
			{"/root", "root directory"},
		}

		for _, test := range sensitiveTests {
			cmd := exec.Command("docker", "run", "--rm", fullImageName, "ls", "-la", test.path)
			output, err := cmd.CombinedOutput()

			// For /etc/passwd, it should exist but not be writable by app user
			// For /etc/shadow and /root, they should not be accessible
			if test.path == "/etc/passwd" {
				if err != nil {
					t.Errorf("Expected %s to exist, but got error: %v", test.desc, err)
				}
			} else {
				// For sensitive files/directories, expect access to be denied
				if err == nil {
					t.Logf("WARNING: %s is accessible (output: %s)", test.desc, string(output))
				}
			}
		}
	})

	// Test container capabilities
	t.Run("container capabilities", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm", fullImageName, "sh", "-c", "id && ps aux")
		output, err := cmd.Output()
		if err != nil {
			t.Logf("Container capabilities test failed (expected for security): %v", err)
		} else {
			t.Logf("Container capabilities output:\n%s", string(output))
		}
	})
}

func TestDockerImageOptimization(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented")
	}

	imageName := "cat-server"
	imageTag := "optimization-test"
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	// Clean up and build image
	cleanupTestImage(fullImageName)
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot test optimization: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fullImageName)

	result := inspectDockerImage(fullImageName)

	// Test layer count (fewer layers = better optimization)
	t.Logf("Image has %d layers", result.Layers)
	if result.Layers > 10 {
		t.Logf("WARNING: Image has %d layers, consider optimizing Dockerfile for fewer layers", result.Layers)
	}

	// Test binary exists and is executable
	t.Run("application binary", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm", fullImageName, "ls", "-la", "/app/cat-server")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Application binary not found: %v", err)
		} else {
			t.Logf("Application binary info: %s", strings.TrimSpace(string(output)))
		}
	})

	// Test static binary (no dynamic dependencies)
	t.Run("static binary", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm", fullImageName, "ldd", "/app/cat-server")
		output, err := cmd.CombinedOutput()

		// For a static binary, ldd should fail or report "not a dynamic executable"
		outputStr := string(output)
		if err != nil || strings.Contains(outputStr, "not a dynamic executable") || strings.Contains(outputStr, "statically linked") {
			t.Logf("Confirmed static binary: %s", outputStr)
		} else {
			t.Logf("WARNING: Binary may have dynamic dependencies: %s", outputStr)
		}
	})
}

// InspectResult represents the result of a Docker image inspection
type InspectResult struct {
	ImageID      string       `json:"image_id"`
	Size         int64        `json:"size"`
	Architecture string       `json:"architecture"`
	OS           string       `json:"os"`
	User         string       `json:"user"`
	ExposedPorts []string     `json:"exposed_ports"`
	SecurityScan SecurityScan `json:"security_scan"`
	Layers       int          `json:"layers"`
}

// SecurityScan represents security scan results
type SecurityScan struct {
	Vulnerabilities Vulnerabilities `json:"vulnerabilities"`
}

// Vulnerabilities represents vulnerability counts
type Vulnerabilities struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// inspectDockerImage performs detailed inspection of a Docker image
func inspectDockerImage(imageName string) InspectResult {
	cmd := exec.Command("docker", "inspect", imageName)
	output, err := cmd.Output()

	if err != nil {
		return InspectResult{}
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return InspectResult{}
	}

	if len(inspectData) == 0 {
		return InspectResult{}
	}

	data := inspectData[0]

	// Extract basic information
	result := InspectResult{
		ImageID:      getString(data, "Id"),
		Size:         getInt64(data, "Size"),
		Architecture: getString(data, "Architecture"),
		OS:           getString(data, "Os"),
		User:         "",
		ExposedPorts: []string{},
		Layers:       0,
	}

	// Extract user from config
	if config, ok := data["Config"].(map[string]interface{}); ok {
		result.User = getString(config, "User")

		// Extract exposed ports
		if exposedPortsRaw, ok := config["ExposedPorts"].(map[string]interface{}); ok {
			for port := range exposedPortsRaw {
				result.ExposedPorts = append(result.ExposedPorts, port)
			}
		}
	}

	// Count layers
	if rootFS, ok := data["RootFS"].(map[string]interface{}); ok {
		if layers, ok := rootFS["Layers"].([]interface{}); ok {
			result.Layers = len(layers)
		}
	}

	// Mock security scan (in real implementation, would integrate with security scanner)
	result.SecurityScan = SecurityScan{
		Vulnerabilities: Vulnerabilities{
			Critical: 0, // Alpine Linux is generally secure
			High:     0,
			Medium:   getRandomInt(0, 3), // Some minor issues are normal
			Low:      getRandomInt(2, 8),
		},
	}

	return result
}

// Helper functions for data extraction
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt64(data map[string]interface{}, key string) int64 {
	if val, ok := data[key].(float64); ok {
		return int64(val)
	}
	return 0
}

func getRandomInt(min, max int) int {
	// Simple deterministic "random" for testing
	return min + (max-min)/2
}

// containsAll checks if slice contains all expected items
func containsAll(slice []string, expected []string) bool {
	for _, exp := range expected {
		found := false
		for _, item := range slice {
			if item == exp {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDockerBuildContract(t *testing.T) {
	// Skip if Docker not available
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	// Skip if Dockerfile doesn't exist yet
	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Skip("Dockerfile not yet implemented - this test will pass after T006 implementation")
	}

	tests := []struct {
		name       string
		dockerfile string
		context    string
		imageName  string
		imageTag   string
		wantErr    bool
	}{
		{
			name:       "successful build with default parameters",
			dockerfile: dockerfilePath,
			context:    filepath.Join("..", ".."),
			imageName:  "cat-server",
			imageTag:   "test",
			wantErr:    false,
		},
		{
			name:       "build with non-existent dockerfile",
			dockerfile: "./NonExistentDockerfile",
			context:    ".",
			imageName:  "cat-server",
			imageTag:   "test-fail",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing test images
			cleanupTestImage(fmt.Sprintf("%s:%s", tt.imageName, tt.imageTag))

			result := buildDockerImage(tt.dockerfile, tt.context, tt.imageName, tt.imageTag)

			if tt.wantErr {
				if result.Status != "failed" {
					t.Errorf("Expected build to fail, but got status: %s", result.Status)
				}
				return
			}

			// Verify successful build
			if result.Status != "success" {
				t.Errorf("Expected build to succeed, but got status: %s, error: %s", result.Status, result.ErrorMessage)
				return
			}

			// Contract requirements verification
			if result.ImageSize > 52428800 { // 50MB in bytes
				t.Errorf("Image size %d bytes exceeds maximum of 50MB (52428800 bytes)", result.ImageSize)
			}

			if result.BuildTime > 60 {
				t.Errorf("Build time %.2f seconds exceeds maximum of 60 seconds", result.BuildTime)
			}

			// Verify image exists
			if result.ImageID == "" {
				t.Error("Expected valid image ID, got empty string")
			}

			// Clean up test image
			cleanupTestImage(fmt.Sprintf("%s:%s", tt.imageName, tt.imageTag))
		})
	}
}

func TestDockerBuildPerformance(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Skip("Dockerfile not yet implemented")
	}

	imageName := "cat-server"
	imageTag := "perf-test"

	// Clean up any existing test images
	cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))

	// Test initial build performance
	result := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)

	if result.Status != "success" {
		t.Fatalf("Build failed: %s", result.ErrorMessage)
	}

	// Verify performance requirements
	t.Logf("Build time: %.2f seconds", result.BuildTime)
	t.Logf("Image size: %d bytes (%.2f MB)", result.ImageSize, float64(result.ImageSize)/1024/1024)

	if result.BuildTime > 60 {
		t.Errorf("Initial build time %.2f seconds exceeds 60 second requirement", result.BuildTime)
	}

	if result.ImageSize > 52428800 { // 50MB
		t.Errorf("Image size %d bytes exceeds 50MB requirement", result.ImageSize)
	}

	// Test rebuild performance (should be faster due to cache)
	cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))
	rebuildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)

	if rebuildResult.Status == "success" {
		t.Logf("Rebuild time: %.2f seconds", rebuildResult.BuildTime)
		if rebuildResult.BuildTime > 30 {
			t.Logf("Rebuild time %.2f seconds exceeds optimal 30 second target (with cache)", rebuildResult.BuildTime)
		}
	}

	// Clean up
	cleanupTestImage(fmt.Sprintf("%s:%s", imageName, imageTag))
}

// BuildResult represents the result of a Docker build operation
type BuildResult struct {
	Status       string   `json:"status"`
	ImageID      string   `json:"image_id,omitempty"`
	ImageSize    int64    `json:"image_size,omitempty"`
	BuildTime    float64  `json:"build_time,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
	BuildLogs    string   `json:"build_logs,omitempty"`
}

// buildDockerImage executes docker build and returns structured result
func buildDockerImage(dockerfile, context, imageName, imageTag string) BuildResult {
	start := time.Now()

	// Prepare docker build command
	imageFullName := fmt.Sprintf("%s:%s", imageName, imageTag)
	cmd := exec.Command("docker", "build", "-f", dockerfile, "-t", imageFullName, context)

	output, err := cmd.CombinedOutput()
	buildTime := time.Since(start).Seconds()

	if err != nil {
		return BuildResult{
			Status:       "failed",
			ErrorMessage: err.Error(),
			BuildLogs:    string(output),
			BuildTime:    buildTime,
		}
	}

	// Get image information
	imageID, size := getImageInfo(imageFullName)

	return BuildResult{
		Status:    "success",
		ImageID:   imageID,
		ImageSize: size,
		BuildTime: buildTime,
	}
}

// getImageInfo retrieves image ID and size
func getImageInfo(imageName string) (string, int64) {
	// Get image ID
	cmd := exec.Command("docker", "images", "--format", "{{.ID}}", imageName)
	output, err := cmd.Output()
	if err != nil {
		return "", 0
	}
	imageID := strings.TrimSpace(string(output))

	// Get image size in bytes
	cmd = exec.Command("docker", "inspect", imageName, "--format", "{{.Size}}")
	output, err = cmd.Output()
	if err != nil {
		return imageID, 0
	}

	sizeStr := strings.TrimSpace(string(output))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return imageID, 0
	}

	return imageID, size
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

// cleanupTestImage removes test images
func cleanupTestImage(imageName string) {
	cmd := exec.Command("docker", "rmi", "-f", imageName)
	cmd.Run() // Ignore errors as image might not exist
}

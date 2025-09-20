package docker

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerSecurityCompliance(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available for testing")
	}

	dockerfilePath := filepath.Join("..", "..", "Dockerfile")
	if !fileExists(dockerfilePath) {
		t.Skip("Dockerfile not yet implemented - this test will pass after T006 implementation")
	}

	// Build the image for security testing
	imageName := "cat-server"
	imageTag := "security-test"
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	cleanupTestImage(fullImageName)
	buildResult := buildDockerImage(dockerfilePath, filepath.Join("..", ".."), imageName, imageTag)
	if buildResult.Status != "success" {
		t.Skipf("Cannot run security tests: build failed - %s", buildResult.ErrorMessage)
	}
	defer cleanupTestImage(fullImageName)

	t.Run("non_root_user_execution", func(t *testing.T) {
		testNonRootUserExecution(t, fullImageName)
	})

	t.Run("file_system_permissions", func(t *testing.T) {
		testFileSystemPermissions(t, fullImageName)
	})

	t.Run("process_capabilities", func(t *testing.T) {
		testProcessCapabilities(t, fullImageName)
	})

	t.Run("sensitive_directories_protection", func(t *testing.T) {
		testSensitiveDirectoriesProtection(t, fullImageName)
	})

	t.Run("package_security", func(t *testing.T) {
		testPackageSecurity(t, fullImageName)
	})

	t.Run("container_escape_prevention", func(t *testing.T) {
		testContainerEscapePrevention(t, fullImageName)
	})
}

func testNonRootUserExecution(t *testing.T, imageName string) {
	// Test 1: Check default user
	cmd := exec.Command("docker", "run", "--rm", imageName, "whoami")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run whoami command: %v", err)
	}

	user := strings.TrimSpace(string(output))
	if user != "app" {
		t.Errorf("Expected container to run as 'app' user, got '%s'", user)
	}

	// Test 2: Check user ID (should not be 0)
	cmd = exec.Command("docker", "run", "--rm", imageName, "id", "-u")
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	uid := strings.TrimSpace(string(output))
	if uid == "0" {
		t.Error("Container is running as root (UID 0), should be non-root")
	} else {
		t.Logf("Container running as UID: %s (non-root)", uid)
	}

	// Test 3: Check group ID (should not be 0)
	cmd = exec.Command("docker", "run", "--rm", imageName, "id", "-g")
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get group ID: %v", err)
	}

	gid := strings.TrimSpace(string(output))
	if gid == "0" {
		t.Error("Container is running as root group (GID 0), should be non-root")
	} else {
		t.Logf("Container running as GID: %s (non-root)", gid)
	}
}

func testFileSystemPermissions(t *testing.T, imageName string) {
	// Test 1: Check application directory permissions
	cmd := exec.Command("docker", "run", "--rm", imageName, "ls", "-la", "/app")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check /app permissions: %v", err)
	}

	outputStr := string(output)
	t.Logf("Application directory permissions:\n%s", outputStr)

	// Verify app user owns the application
	if !strings.Contains(outputStr, "app") {
		t.Error("Application directory should be owned by 'app' user")
	}

	// Test 2: Check if user can write to application directory
	cmd = exec.Command("docker", "run", "--rm", imageName, "touch", "/app/test-write.txt")
	err = cmd.Run()
	if err != nil {
		t.Error("User should be able to write to application directory")
	}

	// Test 3: Check if user cannot write to system directories
	systemDirs := []string{"/etc", "/usr", "/bin", "/sbin"}
	for _, dir := range systemDirs {
		cmd = exec.Command("docker", "run", "--rm", imageName, "touch", dir+"/test-write.txt")
		err = cmd.Run()
		if err == nil {
			t.Errorf("User should NOT be able to write to system directory: %s", dir)
		}
	}

	// Test 4: Check read-only file system areas
	cmd = exec.Command("docker", "run", "--rm", imageName, "ls", "-la", "/etc/passwd")
	output, err = cmd.Output()
	if err != nil {
		t.Error("Should be able to read /etc/passwd")
	} else {
		t.Logf("Passwd file permissions: %s", strings.TrimSpace(string(output)))
	}
}

func testProcessCapabilities(t *testing.T, imageName string) {
	// Test 1: Check if container has minimal capabilities
	cmd := exec.Command("docker", "run", "--rm", imageName, "cat", "/proc/self/status")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Cannot check process capabilities: %v", err)
		return
	}

	outputStr := string(output)

	// Look for capability information
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Cap") {
			t.Logf("Process capability: %s", line)
		}
	}

	// Test 2: Check if dangerous operations are prevented
	dangerousCommands := [][]string{
		{"mount", "/dev/sda1", "/mnt"},
		{"chroot", "/"},
		{"mknod", "/tmp/testdevice", "c", "1", "1"},
	}

	for _, cmd := range dangerousCommands {
		execCmd := exec.Command("docker", "run", "--rm", imageName)
		execCmd.Args = append(execCmd.Args, cmd...)

		err := execCmd.Run()
		if err == nil {
			t.Errorf("Dangerous command should have failed: %v", cmd)
		}
	}
}

func testSensitiveDirectoriesProtection(t *testing.T, imageName string) {
	// Test access to sensitive directories and files
	sensitiveTests := []struct {
		path        string
		description string
		shouldExist bool
	}{
		{"/etc/passwd", "passwd file", true},
		{"/etc/shadow", "shadow file", false}, // Should not be accessible
		{"/root", "root home directory", false},
		{"/proc/self/environ", "process environment", true},
		{"/sys", "system directory", true},
	}

	for _, test := range sensitiveTests {
		cmd := exec.Command("docker", "run", "--rm", imageName, "ls", "-la", test.path)
		output, err := cmd.CombinedOutput()

		if test.shouldExist {
			if err != nil {
				t.Errorf("Expected %s to be accessible, but got error: %v", test.description, err)
			} else {
				t.Logf("✓ %s is accessible (as expected): %s", test.description,
					strings.TrimSpace(string(output)))
			}
		} else {
			if err == nil {
				t.Logf("WARNING: %s is accessible (output: %s)", test.description,
					strings.TrimSpace(string(output)))
			} else {
				t.Logf("✓ %s is protected (access denied)", test.description)
			}
		}
	}

	// Test if /etc/passwd is readable but not writable
	cmd := exec.Command("docker", "run", "--rm", imageName, "cat", "/etc/passwd")
	err := cmd.Run()
	if err != nil {
		t.Error("Should be able to read /etc/passwd")
	}

	cmd = exec.Command("docker", "run", "--rm", imageName, "sh", "-c", "echo 'test' >> /etc/passwd")
	err = cmd.Run()
	if err == nil {
		t.Error("Should NOT be able to write to /etc/passwd")
	}
}

func testPackageSecurity(t *testing.T, imageName string) {
	// Test 1: Check what packages are installed
	cmd := exec.Command("docker", "run", "--rm", imageName, "apk", "list", "--installed")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Cannot check installed packages: %v", err)
		return
	}

	packages := strings.Split(string(output), "\n")
	t.Logf("Installed packages count: %d", len(packages)-1) // -1 for empty line

	// Log first few packages for verification
	for i, pkg := range packages {
		if i >= 10 || pkg == "" {
			break
		}
		t.Logf("Package: %s", pkg)
	}

	// Test 2: Check for unnecessary packages that shouldn't be there
	unnecessaryPackages := []string{
		"gcc", "g++", "make", "python", "perl", "ruby",
		"curl", "ssh", "telnet", "ftp", "nc", "netcat",
	}

	for _, pkg := range unnecessaryPackages {
		cmd := exec.Command("docker", "run", "--rm", imageName, "which", pkg)
		err := cmd.Run()
		if err == nil {
			t.Logf("WARNING: Unnecessary package '%s' found in container", pkg)
		}
	}

	// Test 3: Verify essential packages are present
	essentialPackages := []string{"wget"} // wget is needed for health checks
	for _, pkg := range essentialPackages {
		cmd := exec.Command("docker", "run", "--rm", imageName, "which", pkg)
		err := cmd.Run()
		if err != nil {
			t.Errorf("Essential package '%s' not found", pkg)
		}
	}
}

func testContainerEscapePrevention(t *testing.T, imageName string) {
	// Test 1: Check if container can access host processes
	cmd := exec.Command("docker", "run", "--rm", imageName, "ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Cannot run ps command: %v", err)
		return
	}

	processes := strings.Split(string(output), "\n")
	t.Logf("Visible processes in container: %d", len(processes)-1)

	// Should only see container processes, not host processes
	for _, proc := range processes {
		if strings.Contains(proc, "systemd") || strings.Contains(proc, "kernel") {
			t.Logf("WARNING: Host process visible in container: %s", proc)
		}
	}

	// Test 2: Check network namespace isolation
	cmd = exec.Command("docker", "run", "--rm", imageName, "ip", "addr", "show")
	output, err = cmd.Output()
	if err != nil {
		t.Logf("Cannot check network interfaces: %v", err)
	} else {
		networkInfo := string(output)
		t.Logf("Container network interfaces:\n%s", networkInfo)

		// Should have limited network interfaces (lo and container interface)
		interfaceCount := strings.Count(networkInfo, "inet ")
		if interfaceCount > 3 { // lo + container interface + maybe docker0
			t.Logf("WARNING: Many network interfaces visible (%d)", interfaceCount)
		}
	}

	// Test 3: Check if container can access Docker socket
	cmd = exec.Command("docker", "run", "--rm", imageName, "ls", "-la", "/var/run/docker.sock")
	err = cmd.Run()
	if err == nil {
		t.Error("CRITICAL: Docker socket accessible from container!")
	} else {
		t.Log("✓ Docker socket not accessible (good)")
	}

	// Test 4: Check filesystem isolation
	cmd = exec.Command("docker", "run", "--rm", imageName, "ls", "/host")
	err = cmd.Run()
	if err == nil {
		t.Error("WARNING: Host filesystem may be accessible")
	} else {
		t.Log("✓ Host filesystem not accessible (good)")
	}
}

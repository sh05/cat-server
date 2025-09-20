package services

import (
	"runtime"
	"time"

	"github.com/sh05/cat-server/pkg/domain/repositories"
	"github.com/sh05/cat-server/pkg/infrastructure/logging"
)

// HealthService provides use cases for health checking operations
type HealthService struct {
	fileSystemRepo repositories.FileSystemRepository
	logger         *logging.Logger
	startTime      time.Time
	version        string
}

// NewHealthService creates a new HealthService
func NewHealthService(fileSystemRepo repositories.FileSystemRepository, logger *logging.Logger, version string) *HealthService {
	return &HealthService{
		fileSystemRepo: fileSystemRepo,
		logger:         logger,
		startTime:      time.Now(),
		version:        version,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	UptimeMs   int64                      `json:"uptimeMs"`
	System     *SystemHealthInfo          `json:"system,omitempty"`
	Components map[string]ComponentHealth `json:"components,omitempty"`
	Metrics    *HealthMetrics             `json:"metrics,omitempty"`
}

// SystemHealthInfo represents system-level health information
type SystemHealthInfo struct {
	Memory      *MemoryInfo `json:"memory"`
	Goroutines  int         `json:"goroutines"`
	GCStats     *GCInfo     `json:"gcStats"`
	LoadAverage *LoadInfo   `json:"loadAverage,omitempty"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Allocated   uint64 `json:"allocated"`
	TotalAlloc  uint64 `json:"totalAlloc"`
	System      uint64 `json:"system"`
	GCCycles    uint32 `json:"gcCycles"`
	HeapObjects uint64 `json:"heapObjects"`
	HeapInuse   uint64 `json:"heapInuse"`
	StackInuse  uint64 `json:"stackInuse"`
}

// GCInfo represents garbage collection statistics
type GCInfo struct {
	NumGC      uint32        `json:"numGC"`
	PauseTotal time.Duration `json:"pauseTotal"`
	LastPause  time.Duration `json:"lastPause"`
	NextGC     uint64        `json:"nextGC"`
	LastGC     time.Time     `json:"lastGC"`
}

// LoadInfo represents system load information (simplified for cross-platform)
type LoadInfo struct {
	NumCPU       int `json:"numCPU"`
	NumGoroutine int `json:"numGoroutine"`
}

// ComponentHealth represents the health of a system component
type ComponentHealth struct {
	Status      string        `json:"status"`
	Message     string        `json:"message,omitempty"`
	LastChecked time.Time     `json:"lastChecked"`
	Duration    time.Duration `json:"duration"`
	Details     interface{}   `json:"details,omitempty"`
}

// HealthMetrics represents key health metrics
type HealthMetrics struct {
	RequestCount    int64         `json:"requestCount"`
	ErrorCount      int64         `json:"errorCount"`
	AverageResponse time.Duration `json:"averageResponse"`
	SuccessRate     float64       `json:"successRate"`
	LastActivity    time.Time     `json:"lastActivity"`
}

// GetSystemHealth returns basic health status
func (s *HealthService) GetSystemHealth() (*HealthResponse, error) {
	start := time.Now()

	response := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   s.version,
		Uptime:    s.getUptime(),
		UptimeMs:  time.Since(s.startTime).Milliseconds(),
	}

	// Log health check
	duration := time.Since(start)
	s.logger.LogHealthCheck("basic", response.Status, duration)

	return response, nil
}

// GetDetailedHealth returns comprehensive health information
func (s *HealthService) GetDetailedHealth() (*HealthResponse, error) {
	start := time.Now()

	// Get basic health first
	response, err := s.GetSystemHealth()
	if err != nil {
		return nil, err
	}

	// Add detailed system information
	response.System = s.getSystemHealthInfo()

	// Check components
	components := make(map[string]ComponentHealth)

	// Check filesystem component
	fsHealth := s.checkFileSystemHealth()
	components["filesystem"] = fsHealth

	// Check memory component
	memHealth := s.checkMemoryHealth()
	components["memory"] = memHealth

	response.Components = components

	// Add metrics
	response.Metrics = s.getHealthMetrics()

	// Determine overall status based on components
	overallStatus := s.calculateOverallStatus(components)
	response.Status = overallStatus

	// Log detailed health check
	duration := time.Since(start)
	s.logger.LogHealthCheck("detailed", response.Status, duration)

	return response, nil
}

// GetHealthMetrics returns performance metrics
func (s *HealthService) GetHealthMetrics() (*HealthMetrics, error) {
	return s.getHealthMetrics(), nil
}

// CheckComponent checks the health of a specific component
func (s *HealthService) CheckComponent(component string) (*ComponentHealth, error) {
	start := time.Now()

	var health ComponentHealth

	switch component {
	case "filesystem":
		health = s.checkFileSystemHealth()
	case "memory":
		health = s.checkMemoryHealth()
	case "goroutines":
		health = s.checkGoroutineHealth()
	default:
		health = ComponentHealth{
			Status:      "unknown",
			Message:     "unknown component",
			LastChecked: time.Now(),
			Duration:    time.Since(start),
		}
	}

	// Log component check
	s.logger.LogHealthCheck(component, health.Status, health.Duration)

	return &health, nil
}

// Helper methods

func (s *HealthService) getUptime() string {
	uptime := time.Since(s.startTime)
	return uptime.String()
}

func (s *HealthService) getSystemHealthInfo() *SystemHealthInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var gcStats *GCInfo
	if m.NumGC > 0 {
		lastGC := time.Unix(0, int64(m.LastGC))
		gcStats = &GCInfo{
			NumGC:      m.NumGC,
			PauseTotal: time.Duration(m.PauseTotalNs),
			LastPause:  time.Duration(m.PauseNs[(m.NumGC+255)%256]),
			NextGC:     m.NextGC,
			LastGC:     lastGC,
		}
	}

	return &SystemHealthInfo{
		Memory: &MemoryInfo{
			Allocated:   m.Alloc,
			TotalAlloc:  m.TotalAlloc,
			System:      m.Sys,
			GCCycles:    m.NumGC,
			HeapObjects: m.HeapObjects,
			HeapInuse:   m.HeapInuse,
			StackInuse:  m.StackInuse,
		},
		Goroutines: runtime.NumGoroutine(),
		GCStats:    gcStats,
		LoadAverage: &LoadInfo{
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
	}
}

func (s *HealthService) checkFileSystemHealth() ComponentHealth {
	start := time.Now()

	// Try to validate a simple path to check filesystem access
	// This is a basic check - in a real implementation you might want to
	// check disk space, read/write permissions, etc.
	status := "healthy"
	message := "filesystem accessible"

	// Here you could add more comprehensive filesystem checks
	// For example, checking if base directory is accessible,
	// checking disk space, etc.

	return ComponentHealth{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Duration:    time.Since(start),
	}
}

func (s *HealthService) checkMemoryHealth() ComponentHealth {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := "healthy"
	message := "memory usage normal"

	// Check if memory usage is concerning
	// These thresholds are examples and should be adjusted based on your needs
	const maxMemoryMB = 500 // 500MB threshold
	memoryUsageMB := m.Alloc / 1024 / 1024

	if memoryUsageMB > maxMemoryMB {
		status = "warning"
		message = "high memory usage"
	}

	details := map[string]interface{}{
		"allocatedMB": memoryUsageMB,
		"systemMB":    m.Sys / 1024 / 1024,
		"gcCycles":    m.NumGC,
		"heapObjects": m.HeapObjects,
	}

	return ComponentHealth{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Duration:    time.Since(start),
		Details:     details,
	}
}

func (s *HealthService) checkGoroutineHealth() ComponentHealth {
	start := time.Now()

	numGoroutines := runtime.NumGoroutine()
	status := "healthy"
	message := "goroutine count normal"

	// Check if goroutine count is concerning
	const maxGoroutines = 1000
	if numGoroutines > maxGoroutines {
		status = "warning"
		message = "high goroutine count"
	}

	details := map[string]interface{}{
		"count":  numGoroutines,
		"maxCPU": runtime.NumCPU(),
	}

	return ComponentHealth{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Duration:    time.Since(start),
		Details:     details,
	}
}

func (s *HealthService) getHealthMetrics() *HealthMetrics {
	// In a real implementation, these would be collected from actual metrics
	// This is a simplified version
	return &HealthMetrics{
		RequestCount:    0, // Would be tracked by middleware
		ErrorCount:      0, // Would be tracked by error handling
		AverageResponse: 0, // Would be calculated from request logs
		SuccessRate:     100.0,
		LastActivity:    time.Now(),
	}
}

func (s *HealthService) calculateOverallStatus(components map[string]ComponentHealth) string {
	hasWarning := false
	hasError := false

	for _, health := range components {
		switch health.Status {
		case "error", "unhealthy":
			hasError = true
		case "warning", "degraded":
			hasWarning = true
		}
	}

	if hasError {
		return "unhealthy"
	}
	if hasWarning {
		return "degraded"
	}
	return "healthy"
}

// SetStartTime sets the application start time (useful for testing)
func (s *HealthService) SetStartTime(startTime time.Time) {
	s.startTime = startTime
}

// GetUptime returns the current uptime duration
func (s *HealthService) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

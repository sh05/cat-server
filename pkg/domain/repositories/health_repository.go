package repositories

import (
	"time"
)

// HealthRepository defines the interface for health check operations
type HealthRepository interface {
	// GetSystemHealth returns the overall system health status
	GetSystemHealth() (*HealthStatus, error)

	// GetDetailedHealth returns detailed health information
	GetDetailedHealth() (*DetailedHealthStatus, error)

	// CheckDependencies verifies the health of external dependencies
	CheckDependencies() ([]DependencyStatus, error)

	// RecordHealthCheck records a health check event
	RecordHealthCheck(result *HealthCheckResult) error

	// GetHealthHistory returns historical health check data
	GetHealthHistory(since time.Time) ([]HealthCheckResult, error)
}

// HealthStatus represents the basic health status
type HealthStatus struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    Duration  `json:"uptime"`
}

// DetailedHealthStatus provides comprehensive health information
type DetailedHealthStatus struct {
	*HealthStatus
	System       *SystemHealth       `json:"system"`
	Dependencies []DependencyStatus  `json:"dependencies"`
	Metrics      *HealthMetrics      `json:"metrics"`
	Checks       []HealthCheckResult `json:"checks"`
}

// Status represents the health status levels
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// SystemHealth represents system-level health information
type SystemHealth struct {
	Memory      *MemoryInfo  `json:"memory"`
	Disk        *DiskInfo    `json:"disk"`
	Network     *NetworkInfo `json:"network"`
	LoadAverage *LoadInfo    `json:"loadAverage"`
	Goroutines  int          `json:"goroutines"`
	GCStats     *GCInfo      `json:"gcStats"`
	StartTime   time.Time    `json:"startTime"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Allocated    uint64  `json:"allocated"`
	TotalAlloc   uint64  `json:"totalAlloc"`
	System       uint64  `json:"system"`
	GCCycles     uint32  `json:"gcCycles"`
	UsagePercent float64 `json:"usagePercent"`
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Total        uint64  `json:"total"`
	Available    uint64  `json:"available"`
	Used         uint64  `json:"used"`
	UsagePercent float64 `json:"usagePercent"`
	Path         string  `json:"path"`
}

// NetworkInfo represents network connectivity information
type NetworkInfo struct {
	InterfaceCount int                `json:"interfaceCount"`
	Interfaces     []NetworkInterface `json:"interfaces"`
	Connectivity   ConnectivityStatus `json:"connectivity"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	IsUp      bool     `json:"isUp"`
}

// ConnectivityStatus represents network connectivity status
type ConnectivityStatus struct {
	Internet bool `json:"internet"`
	DNS      bool `json:"dns"`
}

// LoadInfo represents system load information
type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// GCInfo represents garbage collection statistics
type GCInfo struct {
	NumGC      uint32        `json:"numGC"`
	PauseTotal time.Duration `json:"pauseTotal"`
	LastPause  time.Duration `json:"lastPause"`
	NextGC     uint64        `json:"nextGC"`
}

// DependencyStatus represents the status of an external dependency
type DependencyStatus struct {
	Name         string        `json:"name"`
	Status       Status        `json:"status"`
	ResponseTime time.Duration `json:"responseTime"`
	LastChecked  time.Time     `json:"lastChecked"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
	Version      string        `json:"version,omitempty"`
	Endpoint     string        `json:"endpoint,omitempty"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	CheckID   string        `json:"checkId"`
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Message   string        `json:"message,omitempty"`
	Details   interface{}   `json:"details,omitempty"`
}

// HealthMetrics represents key health metrics
type HealthMetrics struct {
	RequestCount    int64         `json:"requestCount"`
	ErrorCount      int64         `json:"errorCount"`
	AverageResponse time.Duration `json:"averageResponse"`
	SuccessRate     float64       `json:"successRate"`
	LastActivity    time.Time     `json:"lastActivity"`
}

// Duration is a custom type for JSON serialization of time.Duration
type Duration time.Duration

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Duration(d).String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	str := string(data[1 : len(data)-1]) // Remove quotes
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

// HealthChecker defines a function type for individual health checks
type HealthChecker func() HealthCheckResult

// HealthCheckConfig represents configuration for health checks
type HealthCheckConfig struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	FailureThreshold int           `json:"failureThreshold"`
	SuccessThreshold int           `json:"successThreshold"`
}

// IsHealthy returns true if the status indicates the system is healthy
func (s Status) IsHealthy() bool {
	return s == StatusHealthy
}

// IsUnhealthy returns true if the status indicates the system is unhealthy
func (s Status) IsUnhealthy() bool {
	return s == StatusUnhealthy
}

// GetOverallStatus determines the overall status from multiple dependency statuses
func GetOverallStatus(dependencies []DependencyStatus) Status {
	if len(dependencies) == 0 {
		return StatusUnknown
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, dep := range dependencies {
		switch dep.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		case StatusUnknown:
			return StatusUnknown
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

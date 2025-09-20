package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the cat-server application
type Config struct {
	Server     ServerConfig     `json:"server"`
	FileSystem FileSystemConfig `json:"filesystem"`
	Logging    LoggingConfig    `json:"logging"`
	Security   SecurityConfig   `json:"security"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// FileSystemConfig holds filesystem-related configuration
type FileSystemConfig struct {
	BaseDirectory string `json:"base_directory"`
	MaxFileSize   int64  `json:"max_file_size"`
	AllowHidden   bool   `json:"allow_hidden"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableCORS            bool `json:"enable_cors"`
	EnableSecurityHeaders bool `json:"enable_security_headers"`
	EnableRateLimit       bool `json:"enable_rate_limit"`
	MaxPathLength         int  `json:"max_path_length"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			Host:         "",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		FileSystem: FileSystemConfig{
			BaseDirectory: "./files/",
			MaxFileSize:   10 * 1024 * 1024, // 10MB
			AllowHidden:   false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Security: SecurityConfig{
			EnableCORS:            true,
			EnableSecurityHeaders: true,
			EnableRateLimit:       false,
			MaxPathLength:         1000,
		},
	}
}

// LoadFromFlags loads configuration from command line flags
func LoadFromFlags() (*Config, error) {
	config := DefaultConfig()

	// Define command line flags
	var (
		port         = flag.String("port", config.Server.Port, "HTTP server port")
		host         = flag.String("host", config.Server.Host, "HTTP server host")
		dir          = flag.String("dir", config.FileSystem.BaseDirectory, "Base directory to serve files from")
		maxFileSize  = flag.Int64("max-file-size", config.FileSystem.MaxFileSize, "Maximum file size in bytes")
		allowHidden  = flag.Bool("allow-hidden", config.FileSystem.AllowHidden, "Allow access to hidden files")
		logLevel     = flag.String("log-level", config.Logging.Level, "Logging level (debug, info, warn, error)")
		logFormat    = flag.String("log-format", config.Logging.Format, "Logging format (json, text)")
		enableCORS   = flag.Bool("enable-cors", config.Security.EnableCORS, "Enable CORS headers")
		readTimeout  = flag.Duration("read-timeout", config.Server.ReadTimeout, "HTTP read timeout")
		writeTimeout = flag.Duration("write-timeout", config.Server.WriteTimeout, "HTTP write timeout")
		idleTimeout  = flag.Duration("idle-timeout", config.Server.IdleTimeout, "HTTP idle timeout")
	)

	flag.Parse()

	// Apply flag values to config
	config.Server.Port = *port
	config.Server.Host = *host
	config.Server.ReadTimeout = *readTimeout
	config.Server.WriteTimeout = *writeTimeout
	config.Server.IdleTimeout = *idleTimeout

	config.FileSystem.BaseDirectory = *dir
	config.FileSystem.MaxFileSize = *maxFileSize
	config.FileSystem.AllowHidden = *allowHidden

	config.Logging.Level = *logLevel
	config.Logging.Format = *logFormat

	config.Security.EnableCORS = *enableCORS

	// Load additional configuration from environment variables
	if err := config.LoadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() error {
	// Server configuration
	if port := os.Getenv("CAT_SERVER_PORT"); port != "" {
		c.Server.Port = port
	}

	if host := os.Getenv("CAT_SERVER_HOST"); host != "" {
		c.Server.Host = host
	}

	// FileSystem configuration
	if dir := os.Getenv("CAT_SERVER_DIR"); dir != "" {
		c.FileSystem.BaseDirectory = dir
	}

	if maxSizeStr := os.Getenv("CAT_SERVER_MAX_FILE_SIZE"); maxSizeStr != "" {
		maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid CAT_SERVER_MAX_FILE_SIZE: %w", err)
		}
		c.FileSystem.MaxFileSize = maxSize
	}

	if allowHiddenStr := os.Getenv("CAT_SERVER_ALLOW_HIDDEN"); allowHiddenStr != "" {
		allowHidden, err := strconv.ParseBool(allowHiddenStr)
		if err != nil {
			return fmt.Errorf("invalid CAT_SERVER_ALLOW_HIDDEN: %w", err)
		}
		c.FileSystem.AllowHidden = allowHidden
	}

	// Logging configuration
	if level := os.Getenv("CAT_SERVER_LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}

	if format := os.Getenv("CAT_SERVER_LOG_FORMAT"); format != "" {
		c.Logging.Format = format
	}

	// Security configuration
	if corsStr := os.Getenv("CAT_SERVER_ENABLE_CORS"); corsStr != "" {
		enableCORS, err := strconv.ParseBool(corsStr)
		if err != nil {
			return fmt.Errorf("invalid CAT_SERVER_ENABLE_CORS: %w", err)
		}
		c.Security.EnableCORS = enableCORS
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	if _, err := strconv.Atoi(c.Server.Port); err != nil {
		return fmt.Errorf("invalid server port: %w", err)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	// Validate filesystem configuration
	if c.FileSystem.BaseDirectory == "" {
		return fmt.Errorf("base directory cannot be empty")
	}

	if c.FileSystem.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive")
	}

	// Check if base directory exists
	if info, err := os.Stat(c.FileSystem.BaseDirectory); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("base directory does not exist: %s", c.FileSystem.BaseDirectory)
		}
		return fmt.Errorf("cannot access base directory: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("base directory is not a directory: %s", c.FileSystem.BaseDirectory)
	}

	// Validate logging configuration
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	// Validate security configuration
	if c.Security.MaxPathLength <= 0 {
		return fmt.Errorf("max path length must be positive")
	}

	return nil
}

// GetServerAddr returns the complete server address
func (c *Config) GetServerAddr() string {
	if c.Server.Host == "" {
		return ":" + c.Server.Port
	}
	return c.Server.Host + ":" + c.Server.Port
}

// IsDebugMode returns true if debug logging is enabled
func (c *Config) IsDebugMode() bool {
	return c.Logging.Level == "debug"
}

// GetMaxFileSize returns the maximum file size in bytes
func (c *Config) GetMaxFileSize() int64 {
	return c.FileSystem.MaxFileSize
}

// GetBaseDirectory returns the base directory path
func (c *Config) GetBaseDirectory() string {
	return c.FileSystem.BaseDirectory
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf("Config{Server: %+v, FileSystem: %+v, Logging: %+v, Security: %+v}",
		c.Server, c.FileSystem, c.Logging, c.Security)
}

// PrintConfig prints the configuration (excluding sensitive information)
func (c *Config) PrintConfig() {
	fmt.Printf("Server Configuration:\n")
	fmt.Printf("  Address: %s\n", c.GetServerAddr())
	fmt.Printf("  Read Timeout: %v\n", c.Server.ReadTimeout)
	fmt.Printf("  Write Timeout: %v\n", c.Server.WriteTimeout)
	fmt.Printf("  Idle Timeout: %v\n", c.Server.IdleTimeout)

	fmt.Printf("FileSystem Configuration:\n")
	fmt.Printf("  Base Directory: %s\n", c.FileSystem.BaseDirectory)
	fmt.Printf("  Max File Size: %d bytes\n", c.FileSystem.MaxFileSize)
	fmt.Printf("  Allow Hidden: %v\n", c.FileSystem.AllowHidden)

	fmt.Printf("Logging Configuration:\n")
	fmt.Printf("  Level: %s\n", c.Logging.Level)
	fmt.Printf("  Format: %s\n", c.Logging.Format)

	fmt.Printf("Security Configuration:\n")
	fmt.Printf("  Enable CORS: %v\n", c.Security.EnableCORS)
	fmt.Printf("  Enable Security Headers: %v\n", c.Security.EnableSecurityHeaders)
	fmt.Printf("  Max Path Length: %d\n", c.Security.MaxPathLength)
}

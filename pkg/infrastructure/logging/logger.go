package logging

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// Logger wraps slog.Logger to provide domain-specific logging functionality
type Logger struct {
	logger *slog.Logger
}

// LogLevel represents logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// NewLogger creates a new logger with the specified configuration
func NewLogger(level LogLevel, format string) *Logger {
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// NewDefaultLogger creates a logger with default settings (INFO level, JSON format)
func NewDefaultLogger() *Logger {
	return NewLogger(LevelInfo, "json")
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// With returns a new logger with the provided key-value pairs added to the context
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// WithContext returns a new logger that includes context information
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract common context values
	var contextArgs []interface{}

	// Add request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		contextArgs = append(contextArgs, "request_id", requestID)
	}

	// Add user ID if available
	if userID := ctx.Value("user_id"); userID != nil {
		contextArgs = append(contextArgs, "user_id", userID)
	}

	// Add correlation ID if available
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		contextArgs = append(contextArgs, "correlation_id", correlationID)
	}

	if len(contextArgs) > 0 {
		return l.With(contextArgs...)
	}

	return l
}

// LogHTTPRequest logs HTTP request information
func (l *Logger) LogHTTPRequest(method, path, userAgent, remoteAddr string) {
	l.Info("http request",
		"method", method,
		"path", path,
		"user_agent", userAgent,
		"remote_addr", remoteAddr,
		"timestamp", time.Now(),
	)
}

// LogHTTPResponse logs HTTP response information
func (l *Logger) LogHTTPResponse(method, path string, statusCode int, duration time.Duration, responseSize int64) {
	l.Info("http response",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration", duration,
		"response_size", responseSize,
		"timestamp", time.Now(),
	)
}

// LogFileSystemOperation logs filesystem operation information
func (l *Logger) LogFileSystemOperation(operation, path string, success bool, duration time.Duration, size int64) {
	level := "info"
	if !success {
		level = "error"
	}

	args := []interface{}{
		"operation", operation,
		"path", path,
		"success", success,
		"duration", duration,
		"timestamp", time.Now(),
	}

	if size > 0 {
		args = append(args, "size", size)
	}

	if level == "error" {
		l.Error("filesystem operation", args...)
	} else {
		l.Info("filesystem operation", args...)
	}
}

// LogHealthCheck logs health check information
func (l *Logger) LogHealthCheck(component string, status string, duration time.Duration) {
	l.Info("health check",
		"component", component,
		"status", status,
		"duration", duration,
		"timestamp", time.Now(),
	)
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(event, path, remoteAddr, userAgent string, blocked bool) {
	level := "warn"
	if blocked {
		level = "error"
	}

	args := []interface{}{
		"security_event", event,
		"path", path,
		"remote_addr", remoteAddr,
		"user_agent", userAgent,
		"blocked", blocked,
		"timestamp", time.Now(),
	}

	if level == "error" {
		l.Error("security event", args...)
	} else {
		l.Warn("security event", args...)
	}
}

// LogError logs an error with additional context
func (l *Logger) LogError(err error, context string, args ...interface{}) {
	logArgs := []interface{}{
		"error", err.Error(),
		"context", context,
		"timestamp", time.Now(),
	}
	logArgs = append(logArgs, args...)

	l.Error("error occurred", logArgs...)
}

// LogStartup logs application startup information
func (l *Logger) LogStartup(service string, version string, port string, env string) {
	l.Info("service starting",
		"service", service,
		"version", version,
		"port", port,
		"environment", env,
		"timestamp", time.Now(),
	)
}

// LogShutdown logs application shutdown information
func (l *Logger) LogShutdown(service string, duration time.Duration) {
	l.Info("service shutdown",
		"service", service,
		"uptime", duration,
		"timestamp", time.Now(),
	)
}

// LogMetrics logs performance metrics
func (l *Logger) LogMetrics(metrics map[string]interface{}) {
	args := []interface{}{
		"timestamp", time.Now(),
	}

	for key, value := range metrics {
		args = append(args, key, value)
	}

	l.Info("metrics", args...)
}

// SetAsDefault sets this logger as the default slog logger
func (l *Logger) SetAsDefault() {
	slog.SetDefault(l.logger)
}

// GetSlogLogger returns the underlying slog.Logger for advanced usage
func (l *Logger) GetSlogLogger() *slog.Logger {
	return l.logger
}

// LogLevel returns the current log level
func (l *Logger) LogLevel() slog.Level {
	// This is a simplified implementation - in practice you might want to track this
	return slog.LevelInfo
}

// IsDebugEnabled returns true if debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelDebug)
}

// IsInfoEnabled returns true if info logging is enabled
func (l *Logger) IsInfoEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelInfo)
}

// LogPerformance logs performance information for an operation
func (l *Logger) LogPerformance(operation string, duration time.Duration, success bool, metadata map[string]interface{}) {
	args := []interface{}{
		"operation", operation,
		"duration", duration,
		"success", success,
		"timestamp", time.Now(),
	}

	for key, value := range metadata {
		args = append(args, key, value)
	}

	if success {
		l.Info("performance", args...)
	} else {
		l.Warn("performance", args...)
	}
}

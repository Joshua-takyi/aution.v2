package utils

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger wraps slog.Logger for structured logging
type Logger struct {
	slog *slog.Logger
}

// NewLogger creates a new logger instance from slog.Logger
func NewLogger(slogger *slog.Logger) *Logger {
	return &Logger{slog: slogger}
}

// Info logs an informational message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	args := l.mapToArgs(fields)
	l.slog.Info(message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	args := l.mapToArgs(fields)
	l.slog.Warn(message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	args := l.mapToArgs(fields)
	l.slog.Error(message, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	args := l.mapToArgs(fields)
	l.slog.Debug(message, args...)
}

// mapToArgs converts map to slog attributes
func (l *Logger) mapToArgs(fields map[string]interface{}) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}
	return args
}

// LogRequest logs HTTP request information
func LogRequest(c *gin.Context, logger *Logger) {
	logger.Info("Incoming request", map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"ip":     c.ClientIP(),
	})
}

// LogResponse logs HTTP response information
func LogResponse(c *gin.Context, logger *Logger, statusCode int, message string) {
	logger.Info("Response sent", map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"status": statusCode,
		"msg":    message,
	})
}

// LogRequestWithTiming logs request completion with duration
func LogRequestWithTiming(c *gin.Context, logger *Logger, duration time.Duration) {
	logger.Info("Request completed", map[string]interface{}{
		"method":   c.Request.Method,
		"path":     c.Request.URL.Path,
		"status":   c.Writer.Status(),
		"duration": duration.Milliseconds(),
		"ip":       c.ClientIP(),
	})
}

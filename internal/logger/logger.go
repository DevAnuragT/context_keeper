package logger

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// New creates a new logger instance
func New(levelStr string) *Logger {
	level := parseLogLevel(levelStr)
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// parseLogLevel parses a log level string
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DEBUG, message, fields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(INFO, message, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WARN, message, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(ERROR, message, fields...)
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
	}

	// Merge all field maps
	if len(fields) > 0 {
		entry.Fields = make(map[string]interface{})
		for _, fieldMap := range fields {
			for k, v := range fieldMap {
				entry.Fields[k] = v
			}
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.logger.Printf("[%s] %s: %s", level.String(), message, err.Error())
		return
	}

	l.logger.Println(string(jsonBytes))
}

// WithFields creates a logger with default fields
func (l *Logger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// FieldLogger is a logger with default fields
type FieldLogger struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug logs a debug message with default fields
func (fl *FieldLogger) Debug(message string, additionalFields ...map[string]interface{}) {
	fields := []map[string]interface{}{fl.fields}
	fields = append(fields, additionalFields...)
	fl.logger.Debug(message, fields...)
}

// Info logs an info message with default fields
func (fl *FieldLogger) Info(message string, additionalFields ...map[string]interface{}) {
	fields := []map[string]interface{}{fl.fields}
	fields = append(fields, additionalFields...)
	fl.logger.Info(message, fields...)
}

// Warn logs a warning message with default fields
func (fl *FieldLogger) Warn(message string, additionalFields ...map[string]interface{}) {
	fields := []map[string]interface{}{fl.fields}
	fields = append(fields, additionalFields...)
	fl.logger.Warn(message, fields...)
}

// Error logs an error message with default fields
func (fl *FieldLogger) Error(message string, additionalFields ...map[string]interface{}) {
	fields := []map[string]interface{}{fl.fields}
	fields = append(fields, additionalFields...)
	fl.logger.Error(message, fields...)
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger
func Init(level string) {
	globalLogger = New(level)
}

// Global logging functions
func Debug(message string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(message, fields...)
	}
}

func Info(message string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(message, fields...)
	}
}

func Warn(message string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(message, fields...)
	}
}

func Error(message string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(message, fields...)
	}
}

func WithFields(fields map[string]interface{}) *FieldLogger {
	if globalLogger != nil {
		return globalLogger.WithFields(fields)
	}
	return nil
}

// HTTPMiddleware creates a logging middleware for HTTP requests
func HTTPMiddleware(logger *Logger) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(wrapper, r)

			// Log request
			duration := time.Since(start)
			logger.Info("HTTP request", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status_code": wrapper.statusCode,
				"duration_ms": duration.Milliseconds(),
				"user_agent":  r.UserAgent(),
				"remote_addr": r.RemoteAddr,
			})
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

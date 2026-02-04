package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  INFO,
		logger: log.New(&buf, "", 0),
	}

	// Test info logging
	logger.Info("test message", map[string]interface{}{
		"key": "value",
	})

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected log message not found")
	}

	// Verify JSON structure
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Errorf("Failed to parse log entry as JSON: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}

	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}

	if entry.Fields["key"] != "value" {
		t.Errorf("Expected field key=value, got %v", entry.Fields["key"])
	}
}

func TestLogLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  WARN,
		logger: log.New(&buf, "", 0),
	}

	// Debug should not be logged
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Debug message should not be logged at WARN level")
	}

	// Warn should be logged
	logger.Warn("warn message")
	if buf.Len() == 0 {
		t.Error("Warn message should be logged at WARN level")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"debug", DEBUG},
		{"INFO", INFO},
		{"info", INFO},
		{"WARN", WARN},
		{"warn", WARN},
		{"WARNING", WARN},
		{"ERROR", ERROR},
		{"error", ERROR},
		{"invalid", INFO}, // default
	}

	for _, test := range tests {
		result := parseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLogLevel(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestGlobalLogger(t *testing.T) {
	// Initialize global logger
	Init("debug")

	// Test that global functions work
	Info("test global info")
	Debug("test global debug")
	Warn("test global warn")
	Error("test global error")

	// Test WithFields
	fieldLogger := WithFields(map[string]interface{}{
		"component": "test",
	})

	if fieldLogger == nil {
		t.Error("WithFields should return a FieldLogger")
	}

	fieldLogger.Info("test with fields")
}

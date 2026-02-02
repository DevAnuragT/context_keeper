package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DevAnuragT/context_keeper/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a test server without database (nil)
	cfg := &config.Config{
		Port: 8080,
	}

	server := &Server{
		mux:    http.NewServeMux(),
		config: cfg,
	}

	// Add health endpoint
	server.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Create test request with allowed origin
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000") // Set allowed origin
	w := httptest.NewRecorder()

	// Test the handler
	server.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := `{"status":"ok"}`
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}

	// Check CORS headers (should match the origin for allowed origins)
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected CORS header 'http://localhost:3000', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Check security headers
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Missing X-Content-Type-Options header")
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Missing X-Frame-Options header")
	}
}

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

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
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

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Missing CORS header")
	}
}

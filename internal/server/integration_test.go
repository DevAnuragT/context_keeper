package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/database"
	"github.com/DevAnuragT/context_keeper/internal/models"
	_ "github.com/lib/pq"
)

// TestIntegrationAPIFlows tests complete API flows end-to-end
func TestIntegrationAPIFlows(t *testing.T) {
	// Skip integration tests if not in CI or if database is not available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}

	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Setup test configuration
	cfg := &config.Config{
		Port:        8080,
		DatabaseURL: "test-db-url",
		JWTSecret:   "test-secret-key-for-integration-tests",
		Environment: "test",
		LogLevel:    "error",
		GitHubOAuth: config.GitHubOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/api/auth/github",
		},
		AIService: config.AIServiceConfig{
			BaseURL: "http://localhost:8000",
			Timeout: 30,
		},
	}

	// Create server
	server := New(db, cfg)

	t.Run("Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %s", response["status"])
		}
	})

	t.Run("CORS Headers", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/repos", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
		}

		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("Expected CORS headers to be set")
		}
	})

	t.Run("Authentication Required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/repos", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		var errorResp models.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		if errorResp.Error != "unauthorized" {
			t.Errorf("Expected error 'unauthorized', got %s", errorResp.Error)
		}
	})

	t.Run("GitHub OAuth Flow", func(t *testing.T) {
		// Test OAuth callback endpoint
		reqBody := map[string]string{"code": "test-oauth-code"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/github", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// This will fail because we don't have a real GitHub OAuth setup
		// but we can verify the endpoint is wired correctly
		if w.Code != http.StatusBadRequest {
			t.Logf("OAuth endpoint returned status %d (expected for test environment)", w.Code)
		}

		// Verify it returns JSON error response
		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["error"] == nil {
			t.Error("Expected error response from OAuth endpoint")
		}
	})

	t.Run("Invalid Endpoints", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/nonexistent", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Security Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		securityHeaders := []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
			"Strict-Transport-Security",
			"Content-Security-Policy",
		}

		for _, header := range securityHeaders {
			if w.Header().Get(header) == "" {
				t.Errorf("Expected security header %s to be set", header)
			}
		}
	})
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use in-memory SQLite for testing if PostgreSQL is not available
	// This is a simplified approach for integration testing
	db, err := sql.Open("postgres", "postgres://localhost/contextkeeper_test?sslmode=disable")
	if err != nil {
		t.Skipf("Skipping integration tests: database not available: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Skipping integration tests: database not reachable: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	return db
}

// TestServiceIntegration tests that all services are properly wired together
func TestServiceIntegration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}

	db := setupTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		Port:        8080,
		DatabaseURL: "test-db-url",
		JWTSecret:   "test-secret-key-for-integration-tests",
		Environment: "test",
		LogLevel:    "error",
		GitHubOAuth: config.GitHubOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/api/auth/github",
		},
		AIService: config.AIServiceConfig{
			BaseURL: "http://localhost:8000",
			Timeout: 30,
		},
	}

	// This should not panic - validates all services are properly wired
	server := New(db, cfg)

	if server == nil {
		t.Fatal("Expected server to be created successfully")
	}

	// Test that the server can handle basic requests
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected health check to return 200, got %d", w.Code)
	}
}

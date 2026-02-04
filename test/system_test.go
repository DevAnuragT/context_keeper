package test

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
	"github.com/DevAnuragT/context_keeper/internal/logger"
	"github.com/DevAnuragT/context_keeper/internal/models"
	"github.com/DevAnuragT/context_keeper/internal/server"
	_ "github.com/lib/pq"
)

// TestSystemEndToEnd tests the complete system behavior
func TestSystemEndToEnd(t *testing.T) {
	if os.Getenv("SKIP_SYSTEM_TESTS") == "true" {
		t.Skip("Skipping system tests")
	}

	// Setup test environment
	db := setupSystemTestDB(t)
	defer db.Close()

	cfg := getSystemTestConfig()
	logger.Init("error") // Reduce log noise during tests

	// Create server
	srv := server.New(db, cfg)

	t.Run("System Health Checks", func(t *testing.T) {
		testSystemHealthChecks(t, srv)
	})

	t.Run("Authentication Flow", func(t *testing.T) {
		testAuthenticationFlow(t, srv)
	})

	t.Run("Error Handling", func(t *testing.T) {
		testErrorHandling(t, srv)
	})

	t.Run("Security Headers", func(t *testing.T) {
		testSecurityHeaders(t, srv)
	})

	t.Run("CORS Configuration", func(t *testing.T) {
		testCORSConfiguration(t, srv)
	})
}

func testSystemHealthChecks(t *testing.T, srv *server.Server) {
	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "Health Check",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse health response: %v", err)
				}
				if response["status"] != "ok" {
					t.Errorf("Expected status 'ok', got %s", response["status"])
				}
			},
		},
		{
			name:           "Readiness Check",
			endpoint:       "/ready",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse readiness response: %v", err)
				}
				if response["status"] != "ready" {
					t.Errorf("Expected status 'ready', got %v", response["status"])
				}

				checks, ok := response["checks"].(map[string]interface{})
				if !ok {
					t.Error("Expected checks object in readiness response")
					return
				}

				db, ok := checks["database"].(map[string]interface{})
				if !ok {
					t.Error("Expected database check in readiness response")
					return
				}

				if db["status"] != "healthy" {
					t.Errorf("Expected database status 'healthy', got %v", db["status"])
				}
			},
		},
		{
			name:           "Metrics",
			endpoint:       "/metrics",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse metrics response: %v", err)
				}

				expectedFields := []string{"uptime_seconds", "version", "environment"}
				for _, field := range expectedFields {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field %s in metrics response", field)
					}
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", test.endpoint, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			if test.checkResponse != nil {
				test.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func testAuthenticationFlow(t *testing.T, srv *server.Server) {
	// Test OAuth endpoint (should fail without real GitHub setup)
	t.Run("OAuth Callback", func(t *testing.T) {
		reqBody := map[string]string{"code": "test-oauth-code"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/auth/github", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)

		// Should return error due to invalid OAuth setup
		if w.Code != http.StatusBadRequest {
			t.Logf("OAuth endpoint returned %d (expected for test environment)", w.Code)
		}

		// Should return JSON error
		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode OAuth response: %v", err)
		}

		if response["error"] == nil {
			t.Error("Expected error in OAuth response")
		}
	})

	// Test protected endpoints without authentication
	protectedEndpoints := []string{
		"/api/repos",
		"/api/repos/ingest",
		"/api/repos/1/status",
		"/api/context/query",
	}

	for _, endpoint := range protectedEndpoints {
		t.Run("Protected "+endpoint, func(t *testing.T) {
			method := "GET"
			if endpoint == "/api/repos/ingest" || endpoint == "/api/context/query" {
				method = "POST"
			}

			req := httptest.NewRequest(method, endpoint, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected 401 for protected endpoint %s, got %d", endpoint, w.Code)
			}

			var errorResp models.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errorResp.Error != "unauthorized" {
				t.Errorf("Expected error 'unauthorized', got %s", errorResp.Error)
			}
		})
	}
}

func testErrorHandling(t *testing.T, srv *server.Server) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           []byte
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid JSON",
			method:         "POST",
			endpoint:       "/api/auth/github",
			body:           []byte(`{"invalid": json}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
		},
		{
			name:           "Missing Fields",
			method:         "POST",
			endpoint:       "/api/auth/github",
			body:           []byte(`{}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
		},
		{
			name:           "Invalid Method",
			method:         "PUT",
			endpoint:       "/api/auth/github",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method_not_allowed",
		},
		{
			name:           "Not Found",
			method:         "GET",
			endpoint:       "/api/nonexistent",
			body:           nil,
			expectedStatus: http.StatusNotFound,
			expectedError:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var req *http.Request
			if test.body != nil {
				req = httptest.NewRequest(test.method, test.endpoint, bytes.NewBuffer(test.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(test.method, test.endpoint, nil)
			}

			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			if test.expectedError != "" && w.Code >= 400 {
				var errorResp models.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if errorResp.Error != test.expectedError {
					t.Errorf("Expected error '%s', got '%s'", test.expectedError, errorResp.Error)
				}
			}
		})
	}
}

func testSecurityHeaders(t *testing.T, srv *server.Server) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	securityHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for header, expectedValue := range securityHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", header, expectedValue, actualValue)
		}
	}
}

func testCORSConfiguration(t *testing.T, srv *server.Server) {
	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/api/repos", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for OPTIONS request, got %d", w.Code)
	}

	corsHeaders := []string{
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Credentials",
	}

	for _, header := range corsHeaders {
		if w.Header().Get(header) == "" {
			t.Errorf("Expected CORS header %s to be set", header)
		}
	}

	// Test allowed origin
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected allowed origin to be set for localhost:3000")
	}
}

func setupSystemTestDB(t *testing.T) *sql.DB {
	// Try to connect to test database
	db, err := sql.Open("postgres", "postgres://localhost/contextkeeper_test?sslmode=disable")
	if err != nil {
		t.Skipf("Skipping system tests: database not available: %v", err)
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Skipping system tests: database not reachable: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to run system test migrations: %v", err)
	}

	return db
}

func getSystemTestConfig() *config.Config {
	return &config.Config{
		Port:        8080,
		DatabaseURL: "postgres://localhost/contextkeeper_test?sslmode=disable",
		JWTSecret:   "system-test-jwt-secret-key",
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
}

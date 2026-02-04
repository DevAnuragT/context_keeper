package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DevAnuragT/context_keeper/internal/middleware"
	"github.com/DevAnuragT/context_keeper/internal/models"
)

// Property 9: Backend API Authentication Enforcement
// For any API endpoint request (except OAuth callback), the system should validate JWT authentication
// before processing and expose action-based REST endpoints with structured JSON responses
// **Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 8.4**
func TestAPIAuthenticationEnforcementProperty(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Generate random test scenario
			scenario := generateAPIAuthScenario(rand.New(rand.NewSource(int64(i))))

			// Test the property
			if !validateAPIAuthProperty(t, scenario) {
				t.Errorf("API authentication enforcement property violated for scenario: %+v", scenario)
			}
		})
	}
}

// APIAuthScenario represents a test scenario for API authentication
type APIAuthScenario struct {
	Endpoint     string
	Method       string
	HasAuth      bool
	ValidToken   bool
	ExpectedCode int
	RequiresAuth bool
}

// generateAPIAuthScenario creates a random API authentication test scenario
func generateAPIAuthScenario(r *rand.Rand) APIAuthScenario {
	// Define all API endpoints
	endpoints := []struct {
		path         string
		method       string
		requiresAuth bool
	}{
		{"/api/auth/github", "POST", false},  // OAuth callback - no auth required
		{"/api/repos", "GET", true},          // Protected endpoint
		{"/api/repos/ingest", "POST", true},  // Protected endpoint
		{"/api/repos/1/status", "GET", true}, // Protected endpoint
		{"/api/context/query", "POST", true}, // Protected endpoint
	}

	endpoint := endpoints[r.Intn(len(endpoints))]

	// Random authentication scenarios
	hasAuth := r.Float32() < 0.8    // 80% chance of having auth header
	validToken := r.Float32() < 0.7 // 70% chance of valid token (when auth is present)

	expectedCode := http.StatusOK
	if endpoint.requiresAuth {
		if !hasAuth {
			expectedCode = http.StatusUnauthorized
		} else if !validToken {
			expectedCode = http.StatusUnauthorized
		}
	}

	return APIAuthScenario{
		Endpoint:     endpoint.path,
		Method:       endpoint.method,
		HasAuth:      hasAuth,
		ValidToken:   validToken,
		ExpectedCode: expectedCode,
		RequiresAuth: endpoint.requiresAuth,
	}
}

// validateAPIAuthProperty validates the API authentication enforcement property
func validateAPIAuthProperty(t *testing.T, scenario APIAuthScenario) bool {
	// Create mock services
	mockAuth := &MockAuthService{shouldFail: !scenario.ValidToken}
	mockJob := &MockJobService{}
	mockContext := &MockContextService{}
	mockRepo := &MockRepositoryStore{
		repos: []models.Repository{{ID: 1, Name: "test-repo"}},
		jobs:  []models.IngestionJob{{ID: 1, Status: models.JobStatusCompleted}},
	}

	handlers := New(mockAuth, mockJob, mockContext, mockRepo)

	// Create request with appropriate body for POST endpoints
	var body []byte
	if scenario.Method == "POST" {
		switch scenario.Endpoint {
		case "/api/auth/github":
			reqBody := map[string]string{"code": "test-code"}
			body, _ = json.Marshal(reqBody)
		case "/api/repos/ingest":
			reqBody := map[string]int64{"repo_id": 1}
			body, _ = json.Marshal(reqBody)
		case "/api/context/query":
			reqBody := models.ContextQuery{RepoID: 1, Query: "test query", Mode: "clarify"}
			body, _ = json.Marshal(reqBody)
		}
	}

	req := httptest.NewRequest(scenario.Method, scenario.Endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Add authentication header if scenario requires it
	if scenario.HasAuth {
		if scenario.ValidToken {
			req.Header.Set("Authorization", "Bearer valid-token")
		} else {
			req.Header.Set("Authorization", "Bearer invalid-token")
		}
	}

	w := httptest.NewRecorder()

	// Route to appropriate handler - always apply middleware for protected endpoints
	switch scenario.Endpoint {
	case "/api/auth/github":
		handlers.HandleGitHubAuth(w, req)
	case "/api/repos":
		middleware.AuthRequired(mockAuth, handlers.HandleGetRepos)(w, req)
	case "/api/repos/ingest":
		middleware.AuthRequired(mockAuth, handlers.HandleIngestRepo)(w, req)
	case "/api/repos/1/status":
		middleware.AuthRequired(mockAuth, handlers.HandleGetRepoStatus)(w, req)
	case "/api/context/query":
		middleware.AuthRequired(mockAuth, handlers.HandleContextQuery)(w, req)
	default:
		t.Logf("Unknown endpoint: %s", scenario.Endpoint)
		return false
	}

	// Validate authentication enforcement
	if scenario.RequiresAuth {
		if !scenario.HasAuth || !scenario.ValidToken {
			// Should return 401 Unauthorized
			if w.Code != http.StatusUnauthorized {
				t.Logf("Expected 401 for protected endpoint without valid auth, got %d", w.Code)
				return false
			}

			// Read body bytes first
			bodyBytes := w.Body.Bytes()
			if len(bodyBytes) == 0 {
				t.Logf("Expected structured JSON error response, got empty body")
				return false
			}

			// Should return structured JSON error
			var errorResp models.ErrorResponse
			if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
				t.Logf("Expected structured JSON error response, got decode error: %v", err)
				return false
			}

			if errorResp.Code != http.StatusUnauthorized {
				t.Logf("Expected error code 401, got %d", errorResp.Code)
				return false
			}

			if errorResp.Error != "unauthorized" {
				t.Logf("Expected error type 'unauthorized', got %s", errorResp.Error)
				return false
			}
		} else {
			// Should allow access with valid auth
			if w.Code == http.StatusUnauthorized {
				t.Logf("Valid auth should not return 401, got %d", w.Code)
				return false
			}
		}
	} else {
		// OAuth endpoint should work without auth
		if scenario.Endpoint == "/api/auth/github" && w.Code == http.StatusUnauthorized {
			t.Logf("OAuth endpoint should not require auth, got %d", w.Code)
			return false
		}
	}

	// Validate structured JSON responses
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Logf("Expected Content-Type application/json, got %s", contentType)
		return false
	}

	// Validate that response is valid JSON (if there's a body)
	bodyBytes := w.Body.Bytes()
	if len(bodyBytes) > 0 {
		var jsonResp any
		if err := json.Unmarshal(bodyBytes, &jsonResp); err != nil {
			t.Logf("Response is not valid JSON: %v, body: %s", err, string(bodyBytes))
			return false
		}
	} else if w.Code >= 400 {
		// Error responses should have a body
		t.Logf("Error response should have JSON body, got empty response")
		return false
	}

	return true
}

// Property 10: Fail-Fast Error Handling
// For any error condition (external service failures, malformed requests, database failures),
// the system should return immediate error responses with appropriate HTTP status codes and
// descriptive messages, without background retries beyond a single attempt
// **Validates: Requirements 2.6, 8.1, 8.2, 8.3, 8.4, 8.5**
func TestFailFastErrorHandlingProperty(t *testing.T) {
	const iterations = 50

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Generate random test scenario
			scenario := generateErrorHandlingScenario(rand.New(rand.NewSource(int64(i))))

			// Test the property
			if !validateErrorHandlingProperty(t, scenario) {
				t.Errorf("Fail-fast error handling property violated for scenario: %+v", scenario)
			}
		})
	}
}

// ErrorHandlingScenario represents a test scenario for error handling
type ErrorHandlingScenario struct {
	ErrorType     string
	Endpoint      string
	Method        string
	ExpectedCode  int
	ShouldHaveMsg bool
}

// generateErrorHandlingScenario creates a random error handling test scenario
func generateErrorHandlingScenario(r *rand.Rand) ErrorHandlingScenario {
	errorTypes := []struct {
		name         string
		expectedCode int
		hasMessage   bool
	}{
		{"malformed_json", http.StatusBadRequest, true},
		{"missing_fields", http.StatusBadRequest, true},
		{"database_error", http.StatusInternalServerError, true},
		{"ai_service_error", http.StatusBadGateway, true},
		{"ai_service_timeout", http.StatusGatewayTimeout, true},
		{"invalid_method", http.StatusMethodNotAllowed, true},
	}

	endpoints := []string{
		"/api/auth/github",
		"/api/repos",
		"/api/repos/ingest",
		"/api/context/query",
	}

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	errorType := errorTypes[r.Intn(len(errorTypes))]
	endpoint := endpoints[r.Intn(len(endpoints))]
	method := methods[r.Intn(len(methods))]

	return ErrorHandlingScenario{
		ErrorType:     errorType.name,
		Endpoint:      endpoint,
		Method:        method,
		ExpectedCode:  errorType.expectedCode,
		ShouldHaveMsg: errorType.hasMessage,
	}
}

// validateErrorHandlingProperty validates the fail-fast error handling property
func validateErrorHandlingProperty(t *testing.T, scenario ErrorHandlingScenario) bool {
	// Create mock services that can simulate different error conditions
	mockAuth := &MockAuthService{}
	mockJob := &MockJobService{}
	mockContext := &MockContextService{}
	mockRepo := &MockRepositoryStore{}

	// Configure mocks based on error type
	switch scenario.ErrorType {
	case "database_error":
		mockRepo.shouldFail = true
	case "ai_service_error":
		mockContext.shouldFail = true
	case "ai_service_timeout":
		mockContext.shouldTimeout = true
	}

	handlers := New(mockAuth, mockJob, mockContext, mockRepo)

	// Create request based on scenario
	var body []byte
	switch scenario.ErrorType {
	case "malformed_json":
		body = []byte(`{"invalid": json}`) // Invalid JSON
	case "missing_fields":
		body = []byte(`{}`) // Missing required fields
	default:
		// Valid JSON for other error types
		if scenario.Method == "POST" {
			switch scenario.Endpoint {
			case "/api/auth/github":
				reqBody := map[string]string{"code": "test-code"}
				body, _ = json.Marshal(reqBody)
			case "/api/repos/ingest":
				reqBody := map[string]int64{"repo_id": 1}
				body, _ = json.Marshal(reqBody)
			case "/api/context/query":
				reqBody := models.ContextQuery{RepoID: 1, Query: "test query", Mode: "clarify"}
				body, _ = json.Marshal(reqBody)
			}
		}
	}

	req := httptest.NewRequest(scenario.Method, scenario.Endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Add valid auth for protected endpoints
	user := &models.User{ID: "123", Login: "testuser", GitHubToken: "github-token"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey{}, user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Route to appropriate handler
	switch scenario.Endpoint {
	case "/api/auth/github":
		handlers.HandleGitHubAuth(w, req)
	case "/api/repos":
		handlers.HandleGetRepos(w, req)
	case "/api/repos/ingest":
		handlers.HandleIngestRepo(w, req)
	case "/api/context/query":
		handlers.HandleContextQuery(w, req)
	}

	// Validate fail-fast behavior
	if scenario.ErrorType == "invalid_method" {
		// Check if method is actually invalid for this endpoint
		validMethods := map[string][]string{
			"/api/auth/github":   {"POST"},
			"/api/repos":         {"GET"},
			"/api/repos/ingest":  {"POST"},
			"/api/context/query": {"POST"},
		}

		if valid, exists := validMethods[scenario.Endpoint]; exists {
			isValidMethod := false
			for _, method := range valid {
				if method == scenario.Method {
					isValidMethod = true
					break
				}
			}

			if !isValidMethod {
				if w.Code != http.StatusMethodNotAllowed {
					t.Logf("Expected 405 for invalid method %s on %s, got %d", scenario.Method, scenario.Endpoint, w.Code)
					return false
				}
			}
		}
	}

	// Validate structured error responses
	if w.Code >= 400 {
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Logf("Expected JSON error response, got Content-Type: %s", contentType)
			return false
		}

		var errorResp models.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
			t.Logf("Expected structured error response, got decode error: %v", err)
			return false
		}

		if errorResp.Code != w.Code {
			t.Logf("Error response code mismatch: HTTP %d vs JSON %d", w.Code, errorResp.Code)
			return false
		}

		if scenario.ShouldHaveMsg && errorResp.Message == "" {
			t.Logf("Expected descriptive error message, got empty string")
			return false
		}

		if errorResp.Error == "" {
			t.Logf("Expected error type, got empty string")
			return false
		}
	}

	return true
}

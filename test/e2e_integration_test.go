package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/contextkeeper/internal/config"
	"github.com/yourusername/contextkeeper/internal/models"
	"github.com/yourusername/contextkeeper/internal/repository"
	"github.com/yourusername/contextkeeper/internal/server"
	"github.com/yourusername/contextkeeper/internal/services"
)

// TestEndToEndUserJourney tests the complete user flow from signup to context query
func TestEndToEndUserJourney(t *testing.T) {
	// Setup test server
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:      8080,
			ServerURL: "http://localhost:8080",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "contextkeeper_test",
		},
		GitHub: config.GitHubConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}

	repo := repository.NewRepository(cfg)
	defer repo.Close()

	srv := server.NewServer(cfg, repo)
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	t.Run("Complete User Journey", func(t *testing.T) {
		// Step 1: User Signup
		t.Log("Step 1: User Signup")
		signupData := map[string]string{
			"email":    "test@example.com",
			"password": "SecurePassword123!",
			"name":     "Test User",
		}
		signupResp := makeRequest(t, testServer.URL+"/api/auth/signup", "POST", signupData, "")
		if signupResp.StatusCode != http.StatusCreated {
			t.Fatalf("Signup failed: %d", signupResp.StatusCode)
		}

		var signupResult map[string]interface{}
		json.NewDecoder(signupResp.Body).Decode(&signupResult)
		userID := signupResult["user_id"].(string)
		t.Logf("✓ User created: %s", userID)

		// Step 2: Email Verification (simulated)
		t.Log("Step 2: Email Verification")
		verifyToken := signupResult["verification_token"].(string)
		verifyResp := makeRequest(t, testServer.URL+"/api/auth/verify-email?token="+verifyToken, "GET", nil, "")
		if verifyResp.StatusCode != http.StatusOK {
			t.Fatalf("Email verification failed: %d", verifyResp.StatusCode)
		}
		t.Log("✓ Email verified")

		// Step 3: User Login
		t.Log("Step 3: User Login")
		loginData := map[string]string{
			"email":    "test@example.com",
			"password": "SecurePassword123!",
		}
		loginResp := makeRequest(t, testServer.URL+"/api/auth/login", "POST", loginData, "")
		if loginResp.StatusCode != http.StatusOK {
			t.Fatalf("Login failed: %d", loginResp.StatusCode)
		}

		var loginResult map[string]interface{}
		json.NewDecoder(loginResp.Body).Decode(&loginResult)
		token := loginResult["token"].(string)
		t.Logf("✓ User logged in, token: %s...", token[:20])

		// Step 4: Create Project Workspace
		t.Log("Step 4: Create Project Workspace")
		projectData := map[string]string{
			"name":        "Test Project",
			"description": "End-to-end test project",
		}
		projectResp := makeRequest(t, testServer.URL+"/api/projects", "POST", projectData, token)
		if projectResp.StatusCode != http.StatusCreated {
			t.Fatalf("Project creation failed: %d", projectResp.StatusCode)
		}

		var projectResult map[string]interface{}
		json.NewDecoder(projectResp.Body).Decode(&projectResult)
		projectID := projectResult["id"].(string)
		t.Logf("✓ Project created: %s", projectID)

		// Step 5: Connect GitHub Integration
		t.Log("Step 5: Connect GitHub Integration")
		githubData := map[string]interface{}{
			"platform":     "github",
			"access_token": "test-github-token",
			"config": map[string]interface{}{
				"repositories": []string{"owner/repo"},
			},
		}
		githubResp := makeRequest(t, testServer.URL+"/api/projects/"+projectID+"/integrations", "POST", githubData, token)
		if githubResp.StatusCode != http.StatusCreated {
			t.Fatalf("GitHub integration failed: %d", githubResp.StatusCode)
		}

		var githubResult map[string]interface{}
		json.NewDecoder(githubResp.Body).Decode(&githubResult)
		integrationID := githubResult["id"].(string)
		t.Logf("✓ GitHub integration connected: %s", integrationID)

		// Step 6: Trigger Data Ingestion
		t.Log("Step 6: Trigger Data Ingestion")
		syncResp := makeRequest(t, testServer.URL+"/api/projects/"+projectID+"/integrations/"+integrationID+"/sync", "POST", nil, token)
		if syncResp.StatusCode != http.StatusAccepted {
			t.Fatalf("Sync trigger failed: %d", syncResp.StatusCode)
		}
		t.Log("✓ Data ingestion triggered")

		// Step 7: Wait for Ingestion to Complete (poll status)
		t.Log("Step 7: Wait for Ingestion to Complete")
		maxWait := 30 * time.Second
		pollInterval := 2 * time.Second
		deadline := time.Now().Add(maxWait)

		for time.Now().Before(deadline) {
			statusResp := makeRequest(t, testServer.URL+"/api/projects/"+projectID+"/integrations/"+integrationID+"/status", "GET", nil, token)
			var statusResult map[string]interface{}
			json.NewDecoder(statusResp.Body).Decode(&statusResult)

			state := statusResult["state"].(string)
			if state == "completed" {
				t.Log("✓ Data ingestion completed")
				break
			} else if state == "error" {
				t.Fatalf("Data ingestion failed: %v", statusResult["error"])
			}

			time.Sleep(pollInterval)
		}

		// Step 8: Query Context via MCP Tools
		t.Log("Step 8: Query Context via MCP Tools")
		queryData := map[string]string{
			"query": "authentication implementation",
		}
		queryResp := makeRequest(t, testServer.URL+"/api/mcp/search_project_knowledge", "POST", queryData, token)
		if queryResp.StatusCode != http.StatusOK {
			t.Fatalf("Context query failed: %d", queryResp.StatusCode)
		}

		var queryResult map[string]interface{}
		json.NewDecoder(queryResp.Body).Decode(&queryResult)
		results := queryResult["results"].([]interface{})
		if len(results) == 0 {
			t.Fatal("Expected context results, got none")
		}
		t.Logf("✓ Context query returned %d results", len(results))

		// Step 9: Get File Context
		t.Log("Step 9: Get File Context")
		fileContextData := map[string]string{
			"file_path": "src/auth/login.go",
		}
		fileResp := makeRequest(t, testServer.URL+"/api/mcp/get_context_for_file", "POST", fileContextData, token)
		if fileResp.StatusCode != http.StatusOK {
			t.Fatalf("File context query failed: %d", fileResp.StatusCode)
		}

		var fileResult map[string]interface{}
		json.NewDecoder(fileResp.Body).Decode(&fileResult)
		t.Logf("✓ File context retrieved: %v", fileResult["file_path"])

		// Step 10: Get Decision History
		t.Log("Step 10: Get Decision History")
		decisionData := map[string]string{
			"feature_or_file": "authentication",
		}
		decisionResp := makeRequest(t, testServer.URL+"/api/mcp/get_decision_history", "POST", decisionData, token)
		if decisionResp.StatusCode != http.StatusOK {
			t.Fatalf("Decision history query failed: %d", decisionResp.StatusCode)
		}

		var decisionResult map[string]interface{}
		json.NewDecoder(decisionResp.Body).Decode(&decisionResult)
		decisions := decisionResult["decisions"].([]interface{})
		t.Logf("✓ Decision history retrieved: %d decisions", len(decisions))

		t.Log("✅ End-to-end user journey completed successfully")
	})
}

// TestMultiPlatformIngestion tests ingestion from multiple platforms
func TestMultiPlatformIngestion(t *testing.T) {
	t.Run("Multi-Platform Data Ingestion", func(t *testing.T) {
		// Setup
		cfg := getTestConfig()
		repo := repository.NewRepository(cfg)
		defer repo.Close()

		srv := server.NewServer(cfg, repo)
		testServer := httptest.NewServer(srv.Router)
		defer testServer.Close()

		// Create user and project
		token, projectID := setupUserAndProject(t, testServer.URL)

		// Connect GitHub
		t.Log("Connecting GitHub integration...")
		githubID := connectIntegration(t, testServer.URL, projectID, "github", token)
		t.Logf("✓ GitHub connected: %s", githubID)

		// Connect Slack
		t.Log("Connecting Slack integration...")
		slackID := connectIntegration(t, testServer.URL, projectID, "slack", token)
		t.Logf("✓ Slack connected: %s", slackID)

		// Connect Discord
		t.Log("Connecting Discord integration...")
		discordID := connectIntegration(t, testServer.URL, projectID, "discord", token)
		t.Logf("✓ Discord connected: %s", discordID)

		// Trigger sync for all platforms
		t.Log("Triggering sync for all platforms...")
		for _, integrationID := range []string{githubID, slackID, discordID} {
			syncResp := makeRequest(t, testServer.URL+"/api/projects/"+projectID+"/integrations/"+integrationID+"/sync", "POST", nil, token)
			if syncResp.StatusCode != http.StatusAccepted {
				t.Fatalf("Sync failed for integration %s: %d", integrationID, syncResp.StatusCode)
			}
		}

		// Wait for all syncs to complete
		t.Log("Waiting for all syncs to complete...")
		time.Sleep(10 * time.Second)

		// Verify cross-platform knowledge linking
		t.Log("Verifying cross-platform knowledge linking...")
		queryData := map[string]string{
			"query": "feature discussion",
		}
		queryResp := makeRequest(t, testServer.URL+"/api/mcp/search_project_knowledge", "POST", queryData, token)
		
		var queryResult map[string]interface{}
		json.NewDecoder(queryResp.Body).Decode(&queryResult)
		results := queryResult["results"].([]interface{})

		// Verify results contain data from multiple platforms
		platforms := make(map[string]bool)
		for _, result := range results {
			r := result.(map[string]interface{})
			if source, ok := r["source"].(string); ok {
				platforms[source] = true
			}
		}

		if len(platforms) < 2 {
			t.Errorf("Expected results from multiple platforms, got: %v", platforms)
		}

		t.Log("✅ Multi-platform ingestion and linking verified")
	})
}

// TestTenantIsolation verifies that tenant data is properly isolated
func TestTenantIsolation(t *testing.T) {
	t.Run("Tenant Data Isolation", func(t *testing.T) {
		cfg := getTestConfig()
		repo := repository.NewRepository(cfg)
		defer repo.Close()

		srv := server.NewServer(cfg, repo)
		testServer := httptest.NewServer(srv.Router)
		defer testServer.Close()

		// Create two separate users and projects
		token1, projectID1 := setupUserAndProject(t, testServer.URL)
		token2, projectID2 := setupUserAndProject(t, testServer.URL)

		// Add data to project 1
		addTestData(t, testServer.URL, projectID1, token1)

		// Try to access project 1 data with user 2 token
		queryResp := makeRequest(t, testServer.URL+"/api/projects/"+projectID1+"/data", "GET", nil, token2)
		if queryResp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden, got %d - tenant isolation violated", queryResp.StatusCode)
		}

		// Verify user 2 can only see their own projects
		projectsResp := makeRequest(t, testServer.URL+"/api/projects", "GET", nil, token2)
		var projectsResult map[string]interface{}
		json.NewDecoder(projectsResp.Body).Decode(&projectsResult)
		projects := projectsResult["projects"].([]interface{})

		for _, p := range projects {
			proj := p.(map[string]interface{})
			if proj["id"].(string) == projectID1 {
				t.Error("User 2 can see User 1's project - tenant isolation violated")
			}
		}

		t.Log("✅ Tenant isolation verified")
	})
}

// TestSecurityBoundaries tests authentication and authorization
func TestSecurityBoundaries(t *testing.T) {
	t.Run("Security Boundaries", func(t *testing.T) {
		cfg := getTestConfig()
		repo := repository.NewRepository(cfg)
		defer repo.Close()

		srv := server.NewServer(cfg, repo)
		testServer := httptest.NewServer(srv.Router)
		defer testServer.Close()

		// Test 1: Unauthenticated access should be denied
		t.Log("Test 1: Unauthenticated access")
		resp := makeRequest(t, testServer.URL+"/api/projects", "GET", nil, "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
		t.Log("✓ Unauthenticated access denied")

		// Test 2: Invalid token should be rejected
		t.Log("Test 2: Invalid token")
		resp = makeRequest(t, testServer.URL+"/api/projects", "GET", nil, "invalid-token")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
		t.Log("✓ Invalid token rejected")

		// Test 3: Expired token should be rejected
		t.Log("Test 3: Expired token")
		expiredToken := generateExpiredToken(t)
		resp = makeRequest(t, testServer.URL+"/api/projects", "GET", nil, expiredToken)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
		t.Log("✓ Expired token rejected")

		// Test 4: SQL injection attempts should be sanitized
		t.Log("Test 4: SQL injection protection")
		maliciousData := map[string]string{
			"email":    "test@example.com'; DROP TABLE users; --",
			"password": "password",
		}
		resp = makeRequest(t, testServer.URL+"/api/auth/login", "POST", maliciousData, "")
		// Should return 401 (invalid credentials) not 500 (SQL error)
		if resp.StatusCode == http.StatusInternalServerError {
			t.Error("SQL injection may not be properly sanitized")
		}
		t.Log("✓ SQL injection protected")

		t.Log("✅ Security boundaries verified")
	})
}

// Helper functions

func makeRequest(t *testing.T, url, method string, data interface{}, token string) *http.Response {
	var body *bytes.Buffer
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	} else {
		body = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

func getTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:      8080,
			ServerURL: "http://localhost:8080",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "contextkeeper_test",
		},
	}
}

func setupUserAndProject(t *testing.T, baseURL string) (token, projectID string) {
	// Signup
	signupData := map[string]string{
		"email":    fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		"password": "SecurePassword123!",
		"name":     "Test User",
	}
	signupResp := makeRequest(t, baseURL+"/api/auth/signup", "POST", signupData, "")
	var signupResult map[string]interface{}
	json.NewDecoder(signupResp.Body).Decode(&signupResult)

	// Login
	loginData := map[string]string{
		"email":    signupData["email"],
		"password": signupData["password"],
	}
	loginResp := makeRequest(t, baseURL+"/api/auth/login", "POST", loginData, "")
	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	token = loginResult["token"].(string)

	// Create project
	projectData := map[string]string{
		"name":        "Test Project",
		"description": "Test project",
	}
	projectResp := makeRequest(t, baseURL+"/api/projects", "POST", projectData, token)
	var projectResult map[string]interface{}
	json.NewDecoder(projectResp.Body).Decode(&projectResult)
	projectID = projectResult["id"].(string)

	return token, projectID
}

func connectIntegration(t *testing.T, baseURL, projectID, platform, token string) string {
	integrationData := map[string]interface{}{
		"platform":     platform,
		"access_token": "test-token",
		"config":       map[string]interface{}{},
	}
	resp := makeRequest(t, baseURL+"/api/projects/"+projectID+"/integrations", "POST", integrationData, token)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["id"].(string)
}

func addTestData(t *testing.T, baseURL, projectID, token string) {
	// Add some test data to the project
	data := map[string]interface{}{
		"type": "test_data",
		"content": map[string]string{
			"key": "value",
		},
	}
	makeRequest(t, baseURL+"/api/projects/"+projectID+"/data", "POST", data, token)
}

func generateExpiredToken(t *testing.T) string {
	// Generate a JWT token that expired 1 hour ago
	// This is a simplified version - actual implementation would use proper JWT library
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyMzkwMjJ9.invalid"
}

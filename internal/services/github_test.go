package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGitHubService_GetUserRepos(t *testing.T) {
	// Create mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path != "/user/repos" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Return mock repositories as JSON
		response := `[
			{
				"id": 12345,
				"name": "test-repo",
				"full_name": "testuser/test-repo",
				"owner": {"login": "testuser"},
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-02T00:00:00Z"
			}
		]`

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create service with mock server
	service := &GitHubServiceImpl{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL,
	}

	// Test successful repo retrieval
	repos, err := service.GetUserRepos(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("GetUserRepos failed: %v", err)
	}

	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}

	// Verify repo data
	if repos[0].ID != 12345 {
		t.Errorf("Expected repo ID 12345, got %d", repos[0].ID)
	}

	if repos[0].FullName != "testuser/test-repo" {
		t.Errorf("Expected full name 'testuser/test-repo', got '%s'", repos[0].FullName)
	}

	if repos[0].Owner != "testuser" {
		t.Errorf("Expected owner 'testuser', got '%s'", repos[0].Owner)
	}
}

func TestGitHubService_RateLimitHandling(t *testing.T) {
	// Create mock server that returns rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(1*time.Hour).Unix(), 10))
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	service := &GitHubServiceImpl{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL,
	}

	// Test rate limit handling
	_, err := service.GetUserRepos(context.Background(), "test-token")
	if err == nil {
		t.Error("Expected rate limit error, got nil")
	}

	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("Expected rate limit error message, got: %v", err)
	}
}

func TestGitHubService_RetryLogic(t *testing.T) {
	attempts := 0

	// Create mock server that fails first request, succeeds second
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts == 1 {
			// First attempt fails with server error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Second attempt succeeds
		if r.URL.Path == "/user/repos" {
			response := `[
				{
					"id": 12345,
					"name": "test-repo",
					"full_name": "testuser/test-repo",
					"owner": {"login": "testuser"},
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-02T00:00:00Z"
				}
			]`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		}
	}))
	defer server.Close()

	service := &GitHubServiceImpl{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL,
	}

	// Test retry logic
	repos, err := service.GetUserRepos(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

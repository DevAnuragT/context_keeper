package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/models"
)

func TestContextService_FilterRepoData(t *testing.T) {
	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, "http://localhost:8080")

	// Set up test data
	repoID := int64(1)

	// Mock repository methods to return test data
	mockRepo.prs = []models.PullRequest{
		{ID: 1, Number: 1, Title: "Test PR 1", Author: "user1"},
		{ID: 2, Number: 2, Title: "Test PR 2", Author: "user2"},
	}
	mockRepo.issues = []models.Issue{
		{ID: 1, Title: "Test Issue 1", Author: "user1"},
		{ID: 2, Title: "Test Issue 2", Author: "user2"},
	}
	mockRepo.commits = []models.Commit{
		{SHA: "abc123", Message: "Test commit 1", Author: "user1"},
		{SHA: "def456", Message: "Test commit 2", Author: "user2"},
	}

	ctx := context.Background()
	repoData, err := contextService.FilterRepoData(ctx, repoID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(repoData.PullRequests) != 2 {
		t.Errorf("Expected 2 pull requests, got %d", len(repoData.PullRequests))
	}

	if len(repoData.Issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(repoData.Issues))
	}

	if len(repoData.Commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(repoData.Commits))
	}
}

func TestContextService_FilterRepoData_Limits(t *testing.T) {
	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, "http://localhost:8080")

	ctx := context.Background()
	repoID := int64(1)

	// Test that the service requests the correct limits
	_, err := contextService.FilterRepoData(ctx, repoID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that the correct limits were requested
	if mockRepo.lastPRLimit != 10 {
		t.Errorf("Expected PR limit 10, got %d", mockRepo.lastPRLimit)
	}

	if mockRepo.lastIssueLimit != 10 {
		t.Errorf("Expected issue limit 10, got %d", mockRepo.lastIssueLimit)
	}

	if mockRepo.lastCommitLimit != 20 {
		t.Errorf("Expected commit limit 20, got %d", mockRepo.lastCommitLimit)
	}
}

func TestContextService_ProcessQuery_Success(t *testing.T) {
	// Create a mock AI service server
	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return mock response
		response := models.ContextResponse{
			ClarifiedGoal: "Test clarified goal",
			Tasks: []models.Task{
				{Title: "Task 1", Acceptance: "Acceptance 1"},
			},
			Questions: []string{"Question 1", "Question 2"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAIServer.Close()

	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, mockAIServer.URL)

	// Set up test data
	repoID := int64(1)
	testRepo := &models.Repository{
		ID:       repoID,
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	mockRepo.repos[repoID] = testRepo

	ctx := context.Background()
	response, err := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.ClarifiedGoal != "Test clarified goal" {
		t.Errorf("Expected clarified goal 'Test clarified goal', got %s", response.ClarifiedGoal)
	}

	if len(response.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(response.Tasks))
	}

	if len(response.Questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(response.Questions))
	}
}

func TestContextService_ProcessQuery_Timeout(t *testing.T) {
	// Create a mock AI service server that delays response
	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than the client timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockAIServer.Close()

	mockRepo := NewMockRepositoryStore()
	// Create a context service with a very short timeout for testing
	contextService := &ContextServiceImpl{
		repo:      mockRepo,
		aiBaseURL: mockAIServer.URL,
		client: &http.Client{
			Timeout: 1 * time.Second, // Short timeout for testing
		},
	}

	// Set up test data
	repoID := int64(1)
	testRepo := &models.Repository{
		ID:       repoID,
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	mockRepo.repos[repoID] = testRepo

	ctx := context.Background()
	_, err := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !contains(err.Error(), "timeout") && !contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestContextService_ProcessQuery_AIServiceError(t *testing.T) {
	// Create a mock AI service server that returns an error
	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer mockAIServer.Close()

	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, mockAIServer.URL)

	// Set up test data
	repoID := int64(1)
	testRepo := &models.Repository{
		ID:       repoID,
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	mockRepo.repos[repoID] = testRepo

	ctx := context.Background()
	_, err := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "AI service returned status 500") {
		t.Errorf("Expected AI service error, got %v", err)
	}
}

func TestContextService_GetAIEndpoint(t *testing.T) {
	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, "http://localhost:8080")

	// Access the implementation to test the private method
	impl := contextService.(*ContextServiceImpl)

	tests := []struct {
		mode     string
		expected string
	}{
		{"restore", "http://localhost:8080/context/restore"},
		{"clarify", "http://localhost:8080/context/clarify"},
		{"unknown", "http://localhost:8080/context/query"},
		{"", "http://localhost:8080/context/query"},
	}

	for _, tt := range tests {
		result := impl.getAIEndpoint(tt.mode)
		if result != tt.expected {
			t.Errorf("getAIEndpoint(%q) = %q, want %q", tt.mode, result, tt.expected)
		}
	}
}

func TestAIServiceError(t *testing.T) {
	// Test regular error
	err := NewAIServiceError(500, "Internal server error", false)
	expected := "AI service error (status 500): Internal server error"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}

	// Test timeout error
	timeoutErr := NewAIServiceError(0, "Request timeout", true)
	expectedTimeout := "AI service timeout: Request timeout"
	if timeoutErr.Error() != expectedTimeout {
		t.Errorf("Expected timeout error message %q, got %q", expectedTimeout, timeoutErr.Error())
	}
}
func TestContextService_ProcessQuery_NoSemanticFiltering(t *testing.T) {
	// Create a mock AI service server that verifies the request payload
	var receivedRequest models.FilteredRepoData
	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and parse the request body
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedRequest)

		// Return mock response
		response := models.ContextResponse{
			ClarifiedGoal: "Test response",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAIServer.Close()

	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, mockAIServer.URL)

	// Set up test data with more items than the limits
	repoID := int64(1)
	testRepo := &models.Repository{
		ID:       repoID,
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	mockRepo.repos[repoID] = testRepo

	// Set up mock data that exceeds limits
	mockRepo.prs = make([]models.PullRequest, 15) // More than 10
	mockRepo.issues = make([]models.Issue, 15)    // More than 10
	mockRepo.commits = make([]models.Commit, 25)  // More than 20

	for i := 0; i < 15; i++ {
		mockRepo.prs[i] = models.PullRequest{ID: int64(i + 1), Number: i + 1, Title: fmt.Sprintf("PR %d", i+1)}
		mockRepo.issues[i] = models.Issue{ID: int64(i + 1), Title: fmt.Sprintf("Issue %d", i+1)}
	}
	for i := 0; i < 25; i++ {
		mockRepo.commits[i] = models.Commit{SHA: fmt.Sprintf("sha%d", i+1), Message: fmt.Sprintf("Commit %d", i+1)}
	}

	ctx := context.Background()
	_, err := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that the request sent to AI service contains filtered data (10/10/20 limits)
	if len(receivedRequest.Context.PullRequests) != 10 {
		t.Errorf("Expected 10 PRs to be sent to AI service (filtered), got %d", len(receivedRequest.Context.PullRequests))
	}

	if len(receivedRequest.Context.Issues) != 10 {
		t.Errorf("Expected 10 issues to be sent to AI service (filtered), got %d", len(receivedRequest.Context.Issues))
	}

	if len(receivedRequest.Context.Commits) != 20 {
		t.Errorf("Expected 20 commits to be sent to AI service (filtered), got %d", len(receivedRequest.Context.Commits))
	}

	// Verify that the query is passed through unchanged
	if receivedRequest.Query != "test query" {
		t.Errorf("Expected query 'test query', got %s", receivedRequest.Query)
	}

	// Verify that the repo name is included
	if receivedRequest.Repo != "testuser/test-repo" {
		t.Errorf("Expected repo 'testuser/test-repo', got %s", receivedRequest.Repo)
	}
}

func TestContextService_ProcessQuery_NoCaching(t *testing.T) {
	callCount := 0
	// Create a mock AI service server that counts calls
	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := models.ContextResponse{
			ClarifiedGoal: fmt.Sprintf("Response %d", callCount),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAIServer.Close()

	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, mockAIServer.URL)

	// Set up test data
	repoID := int64(1)
	testRepo := &models.Repository{
		ID:       repoID,
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	mockRepo.repos[repoID] = testRepo

	ctx := context.Background()

	// Make the same query twice
	response1, err1 := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err1 != nil {
		t.Fatalf("First query failed: %v", err1)
	}

	response2, err2 := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err2 != nil {
		t.Fatalf("Second query failed: %v", err2)
	}

	// Verify that both calls were made to the AI service (no caching)
	if callCount != 2 {
		t.Errorf("Expected 2 calls to AI service, got %d", callCount)
	}

	// Verify that responses are different (proving no caching)
	if response1.ClarifiedGoal == response2.ClarifiedGoal {
		t.Errorf("Expected different responses due to no caching, got same response: %s", response1.ClarifiedGoal)
	}
}

func TestContextService_ProcessQuery_ForwardsResponseDirectly(t *testing.T) {
	// Create a mock AI service server with a complex response
	expectedResponse := models.ContextResponse{
		ClarifiedGoal: "Complex clarified goal",
		Tasks: []models.Task{
			{Title: "Task 1", Acceptance: "Acceptance 1"},
			{Title: "Task 2", Acceptance: "Acceptance 2"},
		},
		Questions: []string{"Question 1", "Question 2", "Question 3"},
		PRScaffold: &models.PRScaffold{
			Branch: "feature/test",
			Title:  "Test PR",
			Body:   "Test PR body",
		},
		Context: map[string]interface{}{
			"complexity": "high",
			"priority":   1,
			"tags":       []string{"backend", "api"},
		},
	}

	mockAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer mockAIServer.Close()

	mockRepo := NewMockRepositoryStore()
	contextService := NewContextService(mockRepo, mockAIServer.URL)

	// Set up test data
	repoID := int64(1)
	testRepo := &models.Repository{
		ID:       repoID,
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	mockRepo.repos[repoID] = testRepo

	ctx := context.Background()
	actualResponse, err := contextService.ProcessQuery(ctx, repoID, "test query", "clarify")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that the response is forwarded exactly as received
	if actualResponse.ClarifiedGoal != expectedResponse.ClarifiedGoal {
		t.Errorf("Expected clarified goal %s, got %s", expectedResponse.ClarifiedGoal, actualResponse.ClarifiedGoal)
	}

	if len(actualResponse.Tasks) != len(expectedResponse.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(expectedResponse.Tasks), len(actualResponse.Tasks))
	}

	if len(actualResponse.Questions) != len(expectedResponse.Questions) {
		t.Errorf("Expected %d questions, got %d", len(expectedResponse.Questions), len(actualResponse.Questions))
	}

	if actualResponse.PRScaffold == nil || actualResponse.PRScaffold.Branch != expectedResponse.PRScaffold.Branch {
		t.Errorf("PR scaffold not forwarded correctly")
	}

	// Verify context map is forwarded
	if actualResponse.Context == nil {
		t.Error("Context map not forwarded")
	} else {
		if actualResponse.Context["complexity"] != "high" {
			t.Errorf("Context complexity not forwarded correctly")
		}
	}
}

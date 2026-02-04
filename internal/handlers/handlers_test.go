package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DevAnuragT/context_keeper/internal/middleware"
	"github.com/DevAnuragT/context_keeper/internal/models"
)

// Mock services for testing
type MockAuthService struct {
	shouldFail bool
}

func (m *MockAuthService) HandleGitHubCallback(ctx context.Context, code string) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("oauth failed")
	}
	return &models.AuthResponse{
		Token: "test-jwt-token",
		User: models.User{
			ID:    "123",
			Login: "testuser",
			Email: "test@example.com",
		},
	}, nil
}

func (m *MockAuthService) ValidateJWT(token string) (*models.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("invalid token")
	}
	return &models.User{
		ID:          "123",
		Login:       "testuser",
		Email:       "test@example.com",
		GitHubToken: "github-token",
	}, nil
}

func (m *MockAuthService) GenerateJWT(user *models.User) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("jwt generation failed")
	}
	return "test-jwt-token", nil
}

type MockJobService struct {
	shouldFail bool
}

func (m *MockJobService) CreateIngestionJob(ctx context.Context, repoID int64, userID string) (*models.IngestionJob, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("job creation failed")
	}
	return &models.IngestionJob{
		ID:     1,
		RepoID: repoID,
		Status: models.JobStatusPending,
	}, nil
}

func (m *MockJobService) GetJobStatus(ctx context.Context, jobID int64) (*models.IngestionJob, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("job not found")
	}
	return &models.IngestionJob{
		ID:     jobID,
		RepoID: 1,
		Status: models.JobStatusCompleted,
	}, nil
}

func (m *MockJobService) ProcessJob(ctx context.Context, job *models.IngestionJob, githubToken string) error {
	if m.shouldFail {
		return fmt.Errorf("job processing failed")
	}
	return nil
}

type MockContextService struct {
	shouldFail    bool
	shouldTimeout bool
}

func (m *MockContextService) ProcessQuery(ctx context.Context, repoID int64, query, mode string) (*models.ContextResponse, error) {
	if m.shouldTimeout {
		return nil, fmt.Errorf("AI service timeout: request exceeded 30 seconds")
	}
	if m.shouldFail {
		return nil, fmt.Errorf("AI service error: internal error")
	}
	return &models.ContextResponse{
		ClarifiedGoal: "Test response",
	}, nil
}

func (m *MockContextService) FilterRepoData(ctx context.Context, repoID int64) (*models.RepoContext, error) {
	return &models.RepoContext{}, nil
}

type MockRepositoryStore struct {
	shouldFail bool
	repos      []models.Repository
	jobs       []models.IngestionJob
}

func (m *MockRepositoryStore) GetReposByUser(ctx context.Context, userID string) ([]models.Repository, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("database error")
	}
	return m.repos, nil
}

func (m *MockRepositoryStore) GetJobsByRepo(ctx context.Context, repoID int64) ([]models.IngestionJob, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("database error")
	}
	return m.jobs, nil
}

// Implement other required methods (not used in handler tests)
func (m *MockRepositoryStore) CreateRepo(ctx context.Context, repo *models.Repository) error {
	return nil
}
func (m *MockRepositoryStore) GetRepoByID(ctx context.Context, repoID int64) (*models.Repository, error) {
	return nil, nil
}
func (m *MockRepositoryStore) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	return nil
}
func (m *MockRepositoryStore) GetRecentPRs(ctx context.Context, repoID int64, limit int) ([]models.PullRequest, error) {
	return nil, nil
}
func (m *MockRepositoryStore) CreateIssue(ctx context.Context, issue *models.Issue) error { return nil }
func (m *MockRepositoryStore) GetRecentIssues(ctx context.Context, repoID int64, limit int) ([]models.Issue, error) {
	return nil, nil
}
func (m *MockRepositoryStore) CreateCommit(ctx context.Context, commit *models.Commit) error {
	return nil
}
func (m *MockRepositoryStore) GetRecentCommits(ctx context.Context, repoID int64, limit int) ([]models.Commit, error) {
	return nil, nil
}
func (m *MockRepositoryStore) CreateJob(ctx context.Context, job *models.IngestionJob) error {
	return nil
}
func (m *MockRepositoryStore) UpdateJobStatus(ctx context.Context, jobID int64, status models.JobStatus, errorMsg *string) error {
	return nil
}
func (m *MockRepositoryStore) GetJobByID(ctx context.Context, jobID int64) (*models.IngestionJob, error) {
	return nil, nil
}

func TestHandleGitHubAuth_Success(t *testing.T) {
	mockAuth := &MockAuthService{}
	handlers := New(mockAuth, &MockJobService{}, &MockContextService{}, &MockRepositoryStore{})

	reqBody := map[string]string{"code": "test-code"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/auth/github", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleGitHubAuth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.AuthResponse
	json.NewDecoder(w.Body).Decode(&response)
	if response.Token != "test-jwt-token" {
		t.Errorf("Expected token 'test-jwt-token', got %s", response.Token)
	}
}

func TestHandleGitHubAuth_InvalidMethod(t *testing.T) {
	handlers := New(&MockAuthService{}, &MockJobService{}, &MockContextService{}, &MockRepositoryStore{})

	req := httptest.NewRequest("GET", "/api/auth/github", nil)
	w := httptest.NewRecorder()

	handlers.HandleGitHubAuth(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleGitHubAuth_MissingCode(t *testing.T) {
	handlers := New(&MockAuthService{}, &MockJobService{}, &MockContextService{}, &MockRepositoryStore{})

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/auth/github", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleGitHubAuth(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleGetRepos_Success(t *testing.T) {
	mockRepo := &MockRepositoryStore{
		repos: []models.Repository{
			{ID: 1, Name: "test-repo", FullName: "user/test-repo"},
		},
	}
	handlers := New(&MockAuthService{}, &MockJobService{}, &MockContextService{}, mockRepo)

	req := httptest.NewRequest("GET", "/api/repos", nil)
	// Add user to context
	user := &models.User{ID: "123", Login: "testuser"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey{}, user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.HandleGetRepos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	repos := response["repositories"].([]interface{})
	if len(repos) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(repos))
	}
}

func TestHandleIngestRepo_Success(t *testing.T) {
	handlers := New(&MockAuthService{}, &MockJobService{}, &MockContextService{}, &MockRepositoryStore{})

	reqBody := map[string]int64{"repo_id": 1}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/repos/ingest", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user to context
	user := &models.User{ID: "123", Login: "testuser", GitHubToken: "github-token"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey{}, user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.HandleIngestRepo(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	if response["job_id"] != float64(1) {
		t.Errorf("Expected job_id 1, got %v", response["job_id"])
	}
}

func TestHandleContextQuery_Success(t *testing.T) {
	handlers := New(&MockAuthService{}, &MockJobService{}, &MockContextService{}, &MockRepositoryStore{})

	reqBody := models.ContextQuery{
		RepoID: 1,
		Query:  "test query",
		Mode:   "clarify",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/context/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleContextQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.ContextResponse
	json.NewDecoder(w.Body).Decode(&response)
	if response.ClarifiedGoal != "Test response" {
		t.Errorf("Expected clarified goal 'Test response', got %s", response.ClarifiedGoal)
	}
}

func TestHandleContextQuery_Timeout(t *testing.T) {
	mockContext := &MockContextService{shouldTimeout: true}
	handlers := New(&MockAuthService{}, &MockJobService{}, mockContext, &MockRepositoryStore{})

	reqBody := models.ContextQuery{
		RepoID: 1,
		Query:  "test query",
		Mode:   "clarify",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/context/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleContextQuery(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("Expected status 504, got %d", w.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	panicHandler := func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}

	handler := RecoveryMiddleware(panicHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := CORSMiddleware(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS headers not set correctly")
	}
}

func TestCORSMiddleware_Options(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := CORSMiddleware(testHandler)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}
}

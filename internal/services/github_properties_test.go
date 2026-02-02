package services

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// Feature: contextkeeper-go-backend, Property 3: Repository Data Extraction Limits
// **Validates: Requirements 2.2, 2.3, 2.4, 10.1, 10.2, 10.3**
func TestRepositoryDataExtractionLimitsProperty(t *testing.T) {
	// Property: For any repository ingestion operation, the system should extract
	// at most 50 pull requests, 50 issues, and 100 commits, ordered by most recent timestamp,
	// regardless of how many items exist in the repository

	rand.Seed(time.Now().UnixNano())

	// Run property test with 50 iterations
	for i := 0; i < 50; i++ {
		// Generate random repository data sizes
		totalPRs := rand.Intn(200) + 1     // 1-200 PRs available
		totalIssues := rand.Intn(200) + 1  // 1-200 issues available
		totalCommits := rand.Intn(300) + 1 // 1-300 commits available

		// Test data extraction limits
		if !testDataExtractionLimits(t, totalPRs, totalIssues, totalCommits, i) {
			t.Errorf("Data extraction limits failed for iteration %d with PRs=%d, Issues=%d, Commits=%d",
				i, totalPRs, totalIssues, totalCommits)
		}
	}
}

// Feature: contextkeeper-go-backend, Property 4: Repository Metadata Field Extraction
// **Validates: Requirements 3.1, 3.2, 3.3**
func TestRepositoryMetadataFieldExtractionProperty(t *testing.T) {
	// Property: For any repository data item, only the specified metadata fields should be
	// extracted and stored: PRs (id, number, title, body, author, state, created_at, merged_at,
	// files_changed, labels), Issues (id, title, body, author, state, created_at, closed_at, labels),
	// Commits (sha, message, author, created_at, files_changed)

	rand.Seed(time.Now().UnixNano())

	// Run property test with 30 iterations
	for i := 0; i < 30; i++ {
		// Test metadata field extraction
		if !testMetadataFieldExtraction(t, i) {
			t.Errorf("Metadata field extraction failed for iteration %d", i)
		}
	}
}

// testDataExtractionLimits tests that data extraction respects the specified limits
func testDataExtractionLimits(t *testing.T, totalPRs, totalIssues, totalCommits, iteration int) bool {
	// Create mock server with specified amounts of data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/repos/test/repo/pulls":
			// Parse per_page parameter
			perPage := 50 // Default
			if pp := r.URL.Query().Get("per_page"); pp != "" {
				if parsed, err := strconv.Atoi(pp); err == nil {
					perPage = parsed
				}
			}

			// Generate mock PRs up to the requested limit
			prs := generateMockPRs(min(perPage, totalPRs))
			w.Write([]byte(prs))

		case r.URL.Path == "/repos/test/repo/issues":
			// Parse per_page parameter
			perPage := 50 // Default
			if pp := r.URL.Query().Get("per_page"); pp != "" {
				if parsed, err := strconv.Atoi(pp); err == nil {
					perPage = parsed
				}
			}

			// Generate mock issues up to the requested limit
			issues := generateMockIssues(min(perPage, totalIssues))
			w.Write([]byte(issues))

		case r.URL.Path == "/repos/test/repo/commits":
			// Parse per_page parameter
			perPage := 100 // Default for commits
			if pp := r.URL.Query().Get("per_page"); pp != "" {
				if parsed, err := strconv.Atoi(pp); err == nil {
					perPage = parsed
				}
			}

			// Generate mock commits up to the requested limit
			commits := generateMockCommits(min(perPage, totalCommits))
			w.Write([]byte(commits))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := &GitHubServiceImpl{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL,
	}

	// Test PR extraction limit (should be max 50)
	prs, err := service.GetPullRequests(context.Background(), "test-token", "test", "repo", 50)
	if err != nil {
		t.Logf("Iteration %d: GetPullRequests failed: %v", iteration, err)
		return false
	}

	expectedPRs := min(50, totalPRs)
	if len(prs) != expectedPRs {
		t.Logf("Iteration %d: Expected %d PRs, got %d", iteration, expectedPRs, len(prs))
		return false
	}

	// Test issue extraction limit (should be max 50)
	issues, err := service.GetIssues(context.Background(), "test-token", "test", "repo", 50)
	if err != nil {
		t.Logf("Iteration %d: GetIssues failed: %v", iteration, err)
		return false
	}

	expectedIssues := min(50, totalIssues)
	if len(issues) != expectedIssues {
		t.Logf("Iteration %d: Expected %d issues, got %d", iteration, expectedIssues, len(issues))
		return false
	}

	// Test commit extraction limit (should be max 100)
	commits, err := service.GetCommits(context.Background(), "test-token", "test", "repo", 100)
	if err != nil {
		t.Logf("Iteration %d: GetCommits failed: %v", iteration, err)
		return false
	}

	expectedCommits := min(100, totalCommits)
	if len(commits) != expectedCommits {
		t.Logf("Iteration %d: Expected %d commits, got %d", iteration, expectedCommits, len(commits))
		return false
	}

	return true
}

// testMetadataFieldExtraction tests that only specified fields are extracted
func testMetadataFieldExtraction(t *testing.T, iteration int) bool {
	// Create mock server with comprehensive data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/test/repo/pulls":
			// Return PR with extra fields that should be ignored
			response := `[{
				"id": 1,
				"number": 1,
				"title": "Test PR",
				"body": "Test body",
				"user": {"login": "testuser"},
				"state": "open",
				"created_at": "2023-01-01T00:00:00Z",
				"merged_at": null,
				"labels": [{"name": "bug"}],
				"extra_field": "should_be_ignored",
				"another_field": 12345
			}]`
			w.Write([]byte(response))

		case "/repos/test/repo/pulls/1/files":
			response := `[{"filename": "test.go"}]`
			w.Write([]byte(response))

		case "/repos/test/repo/issues":
			// Return issue with extra fields that should be ignored
			response := `[{
				"id": 1,
				"title": "Test Issue",
				"body": "Test body",
				"user": {"login": "testuser"},
				"state": "open",
				"created_at": "2023-01-01T00:00:00Z",
				"closed_at": null,
				"labels": [{"name": "bug"}],
				"extra_field": "should_be_ignored",
				"assignees": ["user1", "user2"]
			}]`
			w.Write([]byte(response))

		case "/repos/test/repo/commits":
			// Return commit with extra fields that should be ignored
			response := `[{
				"sha": "abc123",
				"commit": {
					"message": "Test commit",
					"author": {
						"name": "testuser",
						"date": "2023-01-01T00:00:00Z"
					}
				},
				"files": [{"filename": "test.go"}],
				"stats": {"additions": 10, "deletions": 5},
				"extra_field": "should_be_ignored"
			}]`
			w.Write([]byte(response))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := &GitHubServiceImpl{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL,
	}

	// Test PR field extraction
	prs, err := service.GetPullRequests(context.Background(), "test-token", "test", "repo", 10)
	if err != nil {
		t.Logf("Iteration %d: GetPullRequests failed: %v", iteration, err)
		return false
	}

	if len(prs) != 1 {
		t.Logf("Iteration %d: Expected 1 PR, got %d", iteration, len(prs))
		return false
	}

	pr := prs[0]
	// Verify only expected fields are populated
	if pr.ID != 1 || pr.Number != 1 || pr.Title != "Test PR" || pr.Body != "Test body" ||
		pr.Author != "testuser" || pr.State != "open" || len(pr.Labels) != 1 ||
		len(pr.FilesChanged) != 1 {
		t.Logf("Iteration %d: PR fields not correctly extracted: %+v", iteration, pr)
		return false
	}

	// Test issue field extraction
	issues, err := service.GetIssues(context.Background(), "test-token", "test", "repo", 10)
	if err != nil {
		t.Logf("Iteration %d: GetIssues failed: %v", iteration, err)
		return false
	}

	if len(issues) != 1 {
		t.Logf("Iteration %d: Expected 1 issue, got %d", iteration, len(issues))
		return false
	}

	issue := issues[0]
	// Verify only expected fields are populated
	if issue.ID != 1 || issue.Title != "Test Issue" || issue.Body != "Test body" ||
		issue.Author != "testuser" || issue.State != "open" || len(issue.Labels) != 1 {
		t.Logf("Iteration %d: Issue fields not correctly extracted: %+v", iteration, issue)
		return false
	}

	// Test commit field extraction
	commits, err := service.GetCommits(context.Background(), "test-token", "test", "repo", 10)
	if err != nil {
		t.Logf("Iteration %d: GetCommits failed: %v", iteration, err)
		return false
	}

	if len(commits) != 1 {
		t.Logf("Iteration %d: Expected 1 commit, got %d", iteration, len(commits))
		return false
	}

	commit := commits[0]
	// Verify only expected fields are populated
	if commit.SHA != "abc123" || commit.Message != "Test commit" || commit.Author != "testuser" ||
		len(commit.FilesChanged) != 1 {
		t.Logf("Iteration %d: Commit fields not correctly extracted: %+v", iteration, commit)
		return false
	}

	return true
}

// Helper functions for generating mock data

func generateMockPRs(count int) string {
	if count == 0 {
		return "[]"
	}

	result := "["
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{
			"id": %d,
			"number": %d,
			"title": "PR %d",
			"body": "Body %d",
			"user": {"login": "user%d"},
			"state": "open",
			"created_at": "2023-01-01T00:00:00Z",
			"merged_at": null,
			"labels": []
		}`, i+1, i+1, i+1, i+1, i+1)
	}
	result += "]"
	return result
}

func generateMockIssues(count int) string {
	if count == 0 {
		return "[]"
	}

	result := "["
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{
			"id": %d,
			"title": "Issue %d",
			"body": "Body %d",
			"user": {"login": "user%d"},
			"state": "open",
			"created_at": "2023-01-01T00:00:00Z",
			"closed_at": null,
			"labels": []
		}`, i+1, i+1, i+1, i+1)
	}
	result += "]"
	return result
}

func generateMockCommits(count int) string {
	if count == 0 {
		return "[]"
	}

	result := "["
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{
			"sha": "sha%d",
			"commit": {
				"message": "Commit %d",
				"author": {
					"name": "user%d",
					"date": "2023-01-01T00:00:00Z"
				}
			},
			"files": []
		}`, i+1, i+1, i+1)
	}
	result += "]"
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

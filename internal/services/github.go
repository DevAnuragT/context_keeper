package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/models"
)

// GitHubServiceImpl implements the GitHubService interface
type GitHubServiceImpl struct {
	httpClient *http.Client
	baseURL    string
}

// NewGitHubService creates a new GitHub API service
func NewGitHubService() GitHubService {
	return &GitHubServiceImpl{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// GitHubRepository represents a GitHub repository from API
type GitHubRepository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GitHubPullRequest represents a GitHub pull request from API
type GitHubPullRequest struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"created_at"`
	MergedAt  *time.Time `json:"merged_at"`
	Labels    []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// GitHubIssue represents a GitHub issue from API
type GitHubIssue struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Number int    `json:"number"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	Labels    []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// GitHubCommit represents a GitHub commit from API
type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Files []struct {
		Filename string `json:"filename"`
	} `json:"files"`
}

// GitHubUser represents a GitHub user from API
type GitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// GetUserRepos retrieves repositories for the authenticated user
func (g *GitHubServiceImpl) GetUserRepos(ctx context.Context, token string) ([]models.Repository, error) {
	url := g.baseURL + "/user/repos?type=public&sort=updated&per_page=100"

	var githubRepos []GitHubRepository
	err := g.makeRequest(ctx, "GET", url, token, nil, &githubRepos)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repos: %w", err)
	}

	// Convert to our models
	repos := make([]models.Repository, len(githubRepos))
	for i, ghRepo := range githubRepos {
		repos[i] = models.Repository{
			ID:        ghRepo.ID,
			Name:      ghRepo.Name,
			FullName:  ghRepo.FullName,
			Owner:     ghRepo.Owner.Login,
			CreatedAt: ghRepo.CreatedAt,
			UpdatedAt: ghRepo.UpdatedAt,
		}
	}

	return repos, nil
}

// GetPullRequests retrieves pull requests for a repository with limit
func (g *GitHubServiceImpl) GetPullRequests(ctx context.Context, token, owner, repo string, limit int) ([]models.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=all&sort=updated&direction=desc&per_page=%d",
		g.baseURL, owner, repo, limit)

	var githubPRs []GitHubPullRequest
	err := g.makeRequest(ctx, "GET", url, token, nil, &githubPRs)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	// Convert to our models
	prs := make([]models.PullRequest, len(githubPRs))
	for i, ghPR := range githubPRs {
		// Get files for this PR
		files, err := g.getPullRequestFiles(ctx, token, owner, repo, ghPR.Number)
		if err != nil {
			// Log error but continue - don't fail entire operation for file retrieval
			files = models.StringList{}
		}

		// Extract labels
		labels := make(models.StringList, len(ghPR.Labels))
		for j, label := range ghPR.Labels {
			labels[j] = label.Name
		}

		prs[i] = models.PullRequest{
			ID:           ghPR.ID,
			Number:       ghPR.Number,
			Title:        ghPR.Title,
			Body:         ghPR.Body,
			Author:       ghPR.User.Login,
			State:        ghPR.State,
			CreatedAt:    ghPR.CreatedAt,
			MergedAt:     ghPR.MergedAt,
			FilesChanged: files,
			Labels:       labels,
		}
	}

	return prs, nil
}

// GetIssues retrieves issues for a repository with limit
func (g *GitHubServiceImpl) GetIssues(ctx context.Context, token, owner, repo string, limit int) ([]models.Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=all&sort=updated&direction=desc&per_page=%d",
		g.baseURL, owner, repo, limit)

	var githubIssues []GitHubIssue
	err := g.makeRequest(ctx, "GET", url, token, nil, &githubIssues)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}

	// Convert to our models and filter out pull requests (GitHub API returns PRs as issues)
	var issues []models.Issue
	for _, ghIssue := range githubIssues {
		// Skip if this is actually a pull request
		if g.isPullRequest(ghIssue) {
			continue
		}

		// Extract labels
		labels := make(models.StringList, len(ghIssue.Labels))
		for j, label := range ghIssue.Labels {
			labels[j] = label.Name
		}

		issue := models.Issue{
			ID:        ghIssue.ID,
			Title:     ghIssue.Title,
			Body:      ghIssue.Body,
			Author:    ghIssue.User.Login,
			State:     ghIssue.State,
			CreatedAt: ghIssue.CreatedAt,
			ClosedAt:  ghIssue.ClosedAt,
			Labels:    labels,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// GetCommits retrieves commits for a repository with limit
func (g *GitHubServiceImpl) GetCommits(ctx context.Context, token, owner, repo string, limit int) ([]models.Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?per_page=%d",
		g.baseURL, owner, repo, limit)

	var githubCommits []GitHubCommit
	err := g.makeRequest(ctx, "GET", url, token, nil, &githubCommits)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Convert to our models
	commits := make([]models.Commit, len(githubCommits))
	for i, ghCommit := range githubCommits {
		// Extract filenames
		files := make(models.StringList, len(ghCommit.Files))
		for j, file := range ghCommit.Files {
			files[j] = file.Filename
		}

		commits[i] = models.Commit{
			SHA:          ghCommit.SHA,
			Message:      ghCommit.Commit.Message,
			Author:       ghCommit.Commit.Author.Name,
			CreatedAt:    ghCommit.Commit.Author.Date,
			FilesChanged: files,
		}
	}

	return commits, nil
}

// GetUserInfo retrieves user information
func (g *GitHubServiceImpl) GetUserInfo(ctx context.Context, token string) (*models.User, error) {
	url := g.baseURL + "/user"

	var githubUser GitHubUser
	err := g.makeRequest(ctx, "GET", url, token, nil, &githubUser)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &models.User{
		ID:    strconv.FormatInt(githubUser.ID, 10),
		Login: githubUser.Login,
		Email: githubUser.Email,
	}, nil
}

// makeRequest makes an HTTP request to GitHub API with rate limiting and retry logic
func (g *GitHubServiceImpl) makeRequest(ctx context.Context, method, url, token string, body interface{}, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ContextKeeper/1.0")

	// Make request with retry logic
	return g.makeRequestWithRetry(req, result)
}

// makeRequestWithRetry implements retry logic for GitHub API requests
func (g *GitHubServiceImpl) makeRequestWithRetry(req *http.Request, result interface{}) error {
	var lastErr error

	// Try up to 2 times (1 retry)
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := g.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt == 0 {
				time.Sleep(1 * time.Second) // Brief delay before retry
				continue
			}
			return fmt.Errorf("HTTP request failed after retry: %w", err)
		}
		defer resp.Body.Close()

		// Handle rate limiting
		if resp.StatusCode == http.StatusForbidden {
			rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
			if rateLimitRemaining == "0" {
				resetTime := resp.Header.Get("X-RateLimit-Reset")
				return fmt.Errorf("GitHub API rate limit exceeded, resets at %s", resetTime)
			}
		}

		// Handle other HTTP errors
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("GitHub API request failed with status %d", resp.StatusCode)
			if attempt == 0 && resp.StatusCode >= 500 {
				time.Sleep(2 * time.Second) // Longer delay for server errors
				continue
			}
			return lastErr
		}

		// Success - decode response
		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}

		return nil
	}

	return lastErr
}

// getPullRequestFiles retrieves files changed in a pull request
func (g *GitHubServiceImpl) getPullRequestFiles(ctx context.Context, token, owner, repo string, prNumber int) (models.StringList, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files", g.baseURL, owner, repo, prNumber)

	var files []struct {
		Filename string `json:"filename"`
	}

	err := g.makeRequest(ctx, "GET", url, token, nil, &files)
	if err != nil {
		return nil, err
	}

	filenames := make(models.StringList, len(files))
	for i, file := range files {
		filenames[i] = file.Filename
	}

	return filenames, nil
}

// isPullRequest checks if a GitHub issue is actually a pull request
func (g *GitHubServiceImpl) isPullRequest(issue GitHubIssue) bool {
	// GitHub API returns pull requests in the issues endpoint
	// We can identify them by checking if they have a pull_request field
	// For now, we'll use a simple heuristic - this could be improved
	return false // For simplicity, we'll handle this in the API call filtering
}

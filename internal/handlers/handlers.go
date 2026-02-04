package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DevAnuragT/context_keeper/internal/middleware"
	"github.com/DevAnuragT/context_keeper/internal/models"
	"github.com/DevAnuragT/context_keeper/internal/services"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	authSvc    services.AuthService
	jobSvc     services.JobService
	contextSvc services.ContextService
	repoStore  services.RepositoryStore
}

// New creates a new handlers instance
func New(authSvc services.AuthService, jobSvc services.JobService, contextSvc services.ContextService, repoStore services.RepositoryStore) *Handlers {
	return &Handlers{
		authSvc:    authSvc,
		jobSvc:     jobSvc,
		contextSvc: contextSvc,
		repoStore:  repoStore,
	}
}

// HandleGitHubAuth handles GitHub OAuth callback
// POST /api/auth/github
func (h *Handlers) HandleGitHubAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Parse request body
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Authorization code required")
		return
	}

	// Handle GitHub OAuth
	authResponse, err := h.authSvc.HandleGitHubCallback(r.Context(), req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "oauth_error", fmt.Sprintf("OAuth failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, authResponse)
}

// HandleGetRepos returns the authenticated user's ingested repositories
// GET /api/repos
func (h *Handlers) HandleGetRepos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Get user from context (set by auth middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "User not found in context")
		return
	}

	// Get repositories for user
	repos, err := h.repoStore.GetReposByUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database_error", "Failed to get repositories")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"repositories": repos,
	})
}

// HandleIngestRepo triggers repository ingestion
// POST /api/repos/ingest
func (h *Handlers) HandleIngestRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "User not found in context")
		return
	}

	// Parse request body
	var req struct {
		RepoID int64 `json:"repo_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.RepoID == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "Repository ID required")
		return
	}

	// Create ingestion job
	job, err := h.jobSvc.CreateIngestionJob(r.Context(), req.RepoID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "job_creation_error", fmt.Sprintf("Failed to create ingestion job: %v", err))
		return
	}

	// Start processing the job in background
	err = h.jobSvc.ProcessJob(r.Context(), job, user.GitHubToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "job_processing_error", fmt.Sprintf("Failed to start job processing: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"job_id": job.ID,
		"status": job.Status,
	})
}

// HandleGetRepoStatus returns ingestion job status for a repository
// GET /api/repos/{id}/status
func (h *Handlers) HandleGetRepoStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Extract repo ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/repos/")
	path = strings.TrimSuffix(path, "/status")

	repoID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid repository ID")
		return
	}

	// Get jobs for repository
	jobs, err := h.repoStore.GetJobsByRepo(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database_error", "Failed to get job status")
		return
	}

	if len(jobs) == 0 {
		writeError(w, http.StatusNotFound, "not_found", "No ingestion jobs found for repository")
		return
	}

	// Return the most recent job
	writeJSON(w, http.StatusOK, jobs[0])
}

// HandleContextQuery processes unified context queries
// POST /api/context/query
func (h *Handlers) HandleContextQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Parse request body
	var req models.ContextQuery

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.RepoID == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "Repository ID required")
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Query required")
		return
	}

	// Default mode if not specified
	if req.Mode == "" {
		req.Mode = "query"
	}

	// Process context query
	response, err := h.contextSvc.ProcessQuery(r.Context(), req.RepoID, req.Query, req.Mode)
	if err != nil {
		// Check if it's a timeout error
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
			writeError(w, http.StatusGatewayTimeout, "ai_service_timeout", "AI service request timed out")
			return
		}

		// Check if it's an AI service error
		if strings.Contains(err.Error(), "AI service") {
			writeError(w, http.StatusBadGateway, "ai_service_error", fmt.Sprintf("AI service error: %v", err))
			return
		}

		writeError(w, http.StatusInternalServerError, "context_query_error", fmt.Sprintf("Failed to process context query: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := models.ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    status,
	}
	json.NewEncoder(w).Encode(response)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

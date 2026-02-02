package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DevAnuragT/context_keeper/internal/models"
	"github.com/DevAnuragT/context_keeper/internal/services"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	authSvc    services.AuthService
	githubSvc  services.GitHubService
	jobSvc     services.JobService
	contextSvc services.ContextService
	repo       services.RepositoryStore
}

// New creates a new handlers instance
func New(
	authSvc services.AuthService,
	githubSvc services.GitHubService,
	jobSvc services.JobService,
	contextSvc services.ContextService,
	repo services.RepositoryStore,
) *Handlers {
	return &Handlers{
		authSvc:    authSvc,
		githubSvc:  githubSvc,
		jobSvc:     jobSvc,
		contextSvc: contextSvc,
		repo:       repo,
	}
}

// HandleGitHubAuth handles GitHub OAuth callback
func (h *Handlers) HandleGitHubAuth(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in task 3.1
	writeError(w, http.StatusNotImplemented, "not_implemented", "GitHub OAuth not implemented yet")
}

// HandleRepos handles repository listing and status requests
func (h *Handlers) HandleRepos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Handle both /api/repos and /api/repos/{id}/status
		// TODO: Implement in task 8.2
		writeError(w, http.StatusNotImplemented, "not_implemented", "Repository listing not implemented yet")
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	}
}

// HandleIngestRepo handles repository ingestion requests
func (h *Handlers) HandleIngestRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// TODO: Implement in task 8.2
	writeError(w, http.StatusNotImplemented, "not_implemented", "Repository ingestion not implemented yet")
}

// HandleContextQuery handles context restoration and requirement clarification
func (h *Handlers) HandleContextQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// TODO: Implement in task 8.2
	writeError(w, http.StatusNotImplemented, "not_implemented", "Context query not implemented yet")
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, errorType, message string) {
	response := models.ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    status,
	}
	writeJSON(w, status, response)
}
package server

import (
	"database/sql"
	"net/http"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/handlers"
	"github.com/DevAnuragT/context_keeper/internal/middleware"
	"github.com/DevAnuragT/context_keeper/internal/repository"
	"github.com/DevAnuragT/context_keeper/internal/services"
)

// Server represents the HTTP server
type Server struct {
	mux    *http.ServeMux
	config *config.Config
}

// New creates a new server instance
func New(db *sql.DB, cfg *config.Config) *Server {
	// Initialize repository
	repo := repository.New(db)

	// Initialize services (will be implemented in later tasks)
	var authSvc services.AuthService
	var githubSvc services.GitHubService
	var jobSvc services.JobService
	var contextSvc services.ContextService

	// Initialize handlers
	h := handlers.New(authSvc, githubSvc, jobSvc, contextSvc, repo)

	// Create router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes (will be implemented in later tasks)
	mux.HandleFunc("/api/auth/github", h.HandleGitHubAuth)
	mux.HandleFunc("/api/repos", middleware.AuthRequired(authSvc, h.HandleRepos))
	mux.HandleFunc("/api/repos/ingest", middleware.AuthRequired(authSvc, h.HandleIngestRepo))
	mux.HandleFunc("/api/context/query", middleware.AuthRequired(authSvc, h.HandleContextQuery))

	return &Server{
		mux:    mux,
		config: cfg,
	}
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}
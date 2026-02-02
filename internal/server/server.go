package server

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

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

	// Initialize services
	authSvc := services.NewAuthService(cfg)
	githubSvc := services.NewGitHubService()
	var jobSvc services.JobService         // Will be implemented in task 6
	var contextSvc services.ContextService // Will be implemented in task 7

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

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// contains checks if a comma-separated list contains a value
func contains(list, value string) bool {
	items := strings.Split(list, ",")
	for _, item := range items {
		if strings.TrimSpace(item) == value {
			return true
		}
	}
	return false
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")

	// Add CORS headers (restrict in production)
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	origin := r.Header.Get("Origin")
	if origin != "" && contains(allowedOrigins, origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/handlers"
	"github.com/DevAnuragT/context_keeper/internal/middleware"
	"github.com/DevAnuragT/context_keeper/internal/repository"
	"github.com/DevAnuragT/context_keeper/internal/services"
)

// Server represents the HTTP server
type Server struct {
	mux       *http.ServeMux
	config    *config.Config
	startTime time.Time
}

// New creates a new server instance
func New(db *sql.DB, cfg *config.Config) *Server {
	server := &Server{
		config:    cfg,
		startTime: time.Now(),
	}

	// Initialize repository
	repo := repository.New(db)

	// Initialize services
	authSvc := services.NewAuthService(cfg)
	githubSvc := services.NewGitHubService()
	jobSvc := services.NewJobService(repo, githubSvc)
	contextSvc := services.NewContextService(repo, cfg.AIService.BaseURL)

	// Initialize handlers
	h := handlers.New(authSvc, jobSvc, contextSvc, repo)

	// Create router
	mux := http.NewServeMux()

	// Health and monitoring endpoints
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/ready", server.handleReady(db))
	mux.HandleFunc("/metrics", server.handleMetrics)

	// API routes
	mux.HandleFunc("/api/auth/github", h.HandleGitHubAuth)
	mux.HandleFunc("/api/repos", middleware.AuthRequired(authSvc, h.HandleGetRepos))
	mux.HandleFunc("/api/repos/ingest", middleware.AuthRequired(authSvc, h.HandleIngestRepo))
	mux.HandleFunc("/api/context/query", middleware.AuthRequired(authSvc, h.HandleContextQuery))

	// Handle repo status endpoint with pattern matching
	mux.HandleFunc("/api/repos/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			middleware.AuthRequired(authSvc, h.HandleGetRepoStatus)(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	server.mux = mux
	return server
}

// handleHealth handles basic health checks
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleReady handles readiness checks with database connectivity
func (s *Server) handleReady(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check database connection
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			response := map[string]interface{}{
				"status": "not_ready",
				"checks": map[string]interface{}{
					"database": map[string]interface{}{
						"status": "unhealthy",
						"error":  err.Error(),
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status": "ready",
			"checks": map[string]interface{}{
				"database": map[string]interface{}{
					"status": "healthy",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}
}

// handleMetrics handles basic application metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	metrics := map[string]interface{}{
		"uptime_seconds": time.Since(s.startTime).Seconds(),
		"version":        "1.0.0",
		"environment":    s.config.Environment,
	}
	json.NewEncoder(w).Encode(metrics)
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

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

	// Initialize supporting services
	passwordSvc := services.NewPasswordService()
	emailSvc := services.NewEmailService(cfg)
	googleOAuthSvc := services.NewGoogleOAuthService(cfg)

	// Initialize services
	authSvc := services.NewAuthService(cfg, repo, passwordSvc, emailSvc, googleOAuthSvc)
	githubSvc := services.NewGitHubService()
	
	// Initialize permission service
	permissionSvc := services.NewPermissionService(repo)
	
	jobSvc := services.NewJobService(repo, githubSvc)
	contextSvc := services.NewContextService(repo, permissionSvc, cfg.AIService.BaseURL)
	
	// Initialize context processor and knowledge graph services
	logger := &services.SimpleLogger{}
	mockAI := &services.ProductionMockAIService{} // Use production mock AI service
	contextProcessor := services.NewContextProcessor(mockAI, logger)
	knowledgeGraphSvc := services.NewKnowledgeGraphService(repo, permissionSvc, contextProcessor, logger)
	
	// Initialize encryption service
	encryptSvc := services.NewEncryptionService(cfg)
	
	// Initialize GitHub integration service
	githubIntegrationSvc := services.NewGitHubIntegrationService(cfg, repo, encryptSvc, logger)
	
	// Initialize Slack integration service
	slackIntegrationSvc := services.NewSlackIntegrationService(cfg, repo, encryptSvc, logger)
	
	// Initialize Discord integration service
	discordIntegrationSvc := services.NewDiscordIntegrationService(cfg, repo, encryptSvc, logger)
	
	// Initialize MCP server
	mcpSvc := services.NewMCPServer(knowledgeGraphSvc, contextSvc, logger)

	// Initialize handlers
	h := handlers.New(authSvc, jobSvc, contextSvc, repo, permissionSvc)
	githubIntegrationHandlers := handlers.NewGitHubIntegrationHandlers(authSvc, githubIntegrationSvc, permissionSvc)
	slackIntegrationHandlers := handlers.NewSlackIntegrationHandlers(authSvc, slackIntegrationSvc, permissionSvc)
	discordIntegrationHandlers := handlers.NewDiscordIntegrationHandlers(authSvc, discordIntegrationSvc, permissionSvc)

	// Create router
	mux := http.NewServeMux()

	// Health and monitoring endpoints
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/ready", server.handleReady(db))
	mux.HandleFunc("/metrics", server.handleMetrics)

	// API routes
	// Authentication routes
	mux.HandleFunc("/api/auth/github", h.HandleGitHubAuth)
	mux.HandleFunc("/api/auth/register", h.HandleRegister)
	mux.HandleFunc("/api/auth/login", h.HandleLogin)
	mux.HandleFunc("/api/auth/password-reset", h.HandlePasswordResetRequest)
	mux.HandleFunc("/api/auth/password-reset/confirm", h.HandlePasswordResetConfirm)
	mux.HandleFunc("/api/auth/verify-email", h.HandleEmailVerification)
	mux.HandleFunc("/api/auth/resend-verification", h.HandleResendEmailVerification)
	
	// OAuth routes
	mux.HandleFunc("/api/auth/oauth/", h.HandleOAuth)
	
	// Protected routes
	mux.HandleFunc("/api/repos", middleware.AuthRequired(authSvc, h.HandleGetRepos))
	mux.HandleFunc("/api/repos/ingest", middleware.AuthRequired(authSvc, h.HandleIngestRepo))
	mux.HandleFunc("/api/context/query", middleware.AuthRequired(authSvc, h.HandleContextQuery))
	
	// GitHub integration routes
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		// GitHub App installation: POST /api/projects/{project_id}/integrations/github/app/install
		if strings.Contains(path, "/integrations/github/app/install") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleGitHubAppInstallation)(w, r)
			return
		}
		
		// GitHub OAuth installation: POST /api/projects/{project_id}/integrations/github/oauth/install
		if strings.Contains(path, "/integrations/github/oauth/install") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleGitHubOAuthInstallation)(w, r)
			return
		}
		
		// Get available repositories: GET /api/projects/{project_id}/integrations/github/{integration_id}/repositories
		if strings.Contains(path, "/integrations/github/") && strings.HasSuffix(path, "/repositories") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleGetAvailableRepositories)(w, r)
			return
		}
		
		// Select repositories: POST /api/projects/{project_id}/integrations/github/{integration_id}/repositories/select
		if strings.Contains(path, "/integrations/github/") && strings.HasSuffix(path, "/repositories/select") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleSelectRepositories)(w, r)
			return
		}
		
		// Update configuration: PUT /api/projects/{project_id}/integrations/github/{integration_id}/configuration
		if strings.Contains(path, "/integrations/github/") && strings.HasSuffix(path, "/configuration") && r.Method == http.MethodPut {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleUpdateIntegrationConfiguration)(w, r)
			return
		}
		
		// Get integration status: GET /api/projects/{project_id}/integrations/github/{integration_id}/status
		if strings.Contains(path, "/integrations/github/") && strings.HasSuffix(path, "/status") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleGetIntegrationStatus)(w, r)
			return
		}
		
		// Delete integration: DELETE /api/projects/{project_id}/integrations/github/{integration_id}
		if strings.Contains(path, "/integrations/github/") && !strings.Contains(path, "/repositories") && !strings.Contains(path, "/configuration") && !strings.Contains(path, "/status") && r.Method == http.MethodDelete {
			middleware.AuthRequired(authSvc, githubIntegrationHandlers.HandleDeleteIntegration)(w, r)
			return
		}
		
		// Slack OAuth installation: POST /api/projects/{project_id}/integrations/slack/oauth/install
		if strings.Contains(path, "/integrations/slack/oauth/install") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleSlackOAuthInstallation)(w, r)
			return
		}
		
		// Get available channels: GET /api/projects/{project_id}/integrations/slack/{integration_id}/channels
		if strings.Contains(path, "/integrations/slack/") && strings.HasSuffix(path, "/channels") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleGetAvailableChannels)(w, r)
			return
		}
		
		// Select channels: POST /api/projects/{project_id}/integrations/slack/{integration_id}/channels/select
		if strings.Contains(path, "/integrations/slack/") && strings.HasSuffix(path, "/channels/select") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleSelectChannels)(w, r)
			return
		}
		
		// Update configuration: PUT /api/projects/{project_id}/integrations/slack/{integration_id}/configuration
		if strings.Contains(path, "/integrations/slack/") && strings.HasSuffix(path, "/configuration") && r.Method == http.MethodPut {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleUpdateSlackConfiguration)(w, r)
			return
		}
		
		// Get integration status: GET /api/projects/{project_id}/integrations/slack/{integration_id}/status
		if strings.Contains(path, "/integrations/slack/") && strings.HasSuffix(path, "/status") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleGetSlackIntegrationStatus)(w, r)
			return
		}
		
		// Validate workspace connection: POST /api/projects/{project_id}/integrations/slack/{integration_id}/validate
		if strings.Contains(path, "/integrations/slack/") && strings.HasSuffix(path, "/validate") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleValidateWorkspaceConnection)(w, r)
			return
		}
		
		// Delete integration: DELETE /api/projects/{project_id}/integrations/slack/{integration_id}
		if strings.Contains(path, "/integrations/slack/") && !strings.Contains(path, "/channels") && !strings.Contains(path, "/configuration") && !strings.Contains(path, "/status") && !strings.Contains(path, "/validate") && r.Method == http.MethodDelete {
			middleware.AuthRequired(authSvc, slackIntegrationHandlers.HandleDeleteSlackIntegration)(w, r)
			return
		}
		
		// Discord bot installation: POST /api/projects/{project_id}/integrations/discord/bot/install
		if strings.Contains(path, "/integrations/discord/bot/install") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleDiscordBotInstallation)(w, r)
			return
		}
		
		// Get available servers: GET /api/projects/{project_id}/integrations/discord/{integration_id}/servers
		if strings.Contains(path, "/integrations/discord/") && strings.HasSuffix(path, "/servers") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleGetAvailableServers)(w, r)
			return
		}
		
		// Get available channels: GET /api/projects/{project_id}/integrations/discord/{integration_id}/servers/{guild_id}/channels
		if strings.Contains(path, "/integrations/discord/") && strings.Contains(path, "/servers/") && strings.HasSuffix(path, "/channels") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleGetAvailableChannels)(w, r)
			return
		}
		
		// Select channels: POST /api/projects/{project_id}/integrations/discord/{integration_id}/channels/select
		if strings.Contains(path, "/integrations/discord/") && strings.HasSuffix(path, "/channels/select") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleSelectChannels)(w, r)
			return
		}
		
		// Update configuration: PUT /api/projects/{project_id}/integrations/discord/{integration_id}/configuration
		if strings.Contains(path, "/integrations/discord/") && strings.HasSuffix(path, "/configuration") && r.Method == http.MethodPut {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleUpdateDiscordConfiguration)(w, r)
			return
		}
		
		// Get integration status: GET /api/projects/{project_id}/integrations/discord/{integration_id}/status
		if strings.Contains(path, "/integrations/discord/") && strings.HasSuffix(path, "/status") && r.Method == http.MethodGet {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleGetDiscordIntegrationStatus)(w, r)
			return
		}
		
		// Validate bot connection: POST /api/projects/{project_id}/integrations/discord/{integration_id}/validate
		if strings.Contains(path, "/integrations/discord/") && strings.HasSuffix(path, "/validate") && r.Method == http.MethodPost {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleValidateBotConnection)(w, r)
			return
		}
		
		// Delete integration: DELETE /api/projects/{project_id}/integrations/discord/{integration_id}
		if strings.Contains(path, "/integrations/discord/") && !strings.Contains(path, "/servers") && !strings.Contains(path, "/channels") && !strings.Contains(path, "/configuration") && !strings.Contains(path, "/status") && !strings.Contains(path, "/validate") && r.Method == http.MethodDelete {
			middleware.AuthRequired(authSvc, discordIntegrationHandlers.HandleDeleteDiscordIntegration)(w, r)
			return
		}
		
		http.NotFound(w, r)
	})
	
	// MCP JSON-RPC endpoint
	mux.Handle("/mcp", mcpSvc)

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

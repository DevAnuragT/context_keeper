package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
	GitHubOAuth GitHubOAuthConfig
	AIService   AIServiceConfig
	Environment string
	LogLevel    string
}

// GitHubOAuthConfig holds GitHub OAuth configuration
type GitHubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// AIServiceConfig holds AI service configuration
type AIServiceConfig struct {
	BaseURL string
	Timeout int // seconds
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:        getEnvInt("PORT", 8080),
		DatabaseURL: getEnv("DATABASE_URL", getDefaultDatabaseURL()),
		JWTSecret:   getSecretOrEnv("JWT_SECRET_FILE", "JWT_SECRET", generateSecureSecret()),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		GitHubOAuth: GitHubOAuthConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getSecretOrEnv("GITHUB_CLIENT_SECRET_FILE", "GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/auth/github"),
		},
		AIService: AIServiceConfig{
			BaseURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
			Timeout: getEnvInt("AI_SERVICE_TIMEOUT", 30),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	if c.Port <= 0 || c.Port > 65535 {
		errors = append(errors, "PORT must be between 1 and 65535")
	}

	if c.DatabaseURL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}

	if c.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is required")
	}

	if c.GitHubOAuth.ClientID == "" {
		errors = append(errors, "GITHUB_CLIENT_ID is required")
	}

	if c.GitHubOAuth.ClientSecret == "" {
		errors = append(errors, "GITHUB_CLIENT_SECRET is required")
	}

	if c.AIService.BaseURL == "" {
		errors = append(errors, "AI_SERVICE_URL is required")
	}

	if c.AIService.Timeout <= 0 {
		errors = append(errors, "AI_SERVICE_TIMEOUT must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDefaultDatabaseURL returns the default database URL based on environment
func getDefaultDatabaseURL() string {
	// For local development, use a simpler connection string
	if getEnv("ENVIRONMENT", "development") == "development" {
		return "postgres://localhost/contextkeeper?sslmode=disable"
	}
	// For production, require SSL
	return "postgres://localhost/contextkeeper?sslmode=require"
}

// getSecretOrEnv reads a secret from a file or falls back to environment variable
func getSecretOrEnv(fileEnvKey, envKey, defaultValue string) string {
	// Try to read from file first
	if filePath := os.Getenv(fileEnvKey); filePath != "" {
		if content, err := os.ReadFile(filePath); err == nil {
			return strings.TrimSpace(string(content))
		}
	}

	// Fall back to environment variable
	return getEnv(envKey, defaultValue)
}

// generateSecureSecret generates a cryptographically secure random secret
func generateSecureSecret() string {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Warning: Failed to generate secure secret, using fallback: %v", err)
		return "fallback-secret-key-change-immediately"
	}
	return hex.EncodeToString(bytes)
}

package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
	GitHubOAuth GitHubOAuthConfig
	AIService   AIServiceConfig
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
	return &Config{
		Port:        getEnvInt("PORT", 8080),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/contextkeeper?sslmode=require"),
		JWTSecret:   getEnv("JWT_SECRET", generateSecureSecret()),
		GitHubOAuth: GitHubOAuthConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/auth/github"),
		},
		AIService: AIServiceConfig{
			BaseURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
			Timeout: getEnvInt("AI_SERVICE_TIMEOUT", 30),
		},
	}
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

// generateSecureSecret generates a cryptographically secure random secret
func generateSecureSecret() string {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Warning: Failed to generate secure secret, using fallback: %v", err)
		return "fallback-secret-key-change-immediately"
	}
	return hex.EncodeToString(bytes)
}

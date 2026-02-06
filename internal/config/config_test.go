package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	defer func() {
		os.Unsetenv("GITHUB_CLIENT_ID")
		os.Unsetenv("GITHUB_CLIENT_SECRET")
	}()

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}

	if cfg.Environment != "development" {
		t.Errorf("Expected environment 'development', got %s", cfg.Environment)
	}

	if cfg.GitHubOAuth.ClientID != "test-client-id" {
		t.Errorf("Expected GitHub client ID 'test-client-id', got %s", cfg.GitHubOAuth.ClientID)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:        8080,
				ServerURL:   "http://localhost:8080",
				DatabaseURL: "postgres://localhost/test",
				JWTSecret:   "secret",
				GitHubOAuth: GitHubOAuthConfig{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
				AIService: AIServiceConfig{
					BaseURL: "http://localhost:8000",
					Timeout: 30,
				},
			},
			wantErr: false,
		},
		{
			name: "missing GitHub client ID",
			config: &Config{
				Port:        8080,
				DatabaseURL: "postgres://localhost/test",
				JWTSecret:   "secret",
				GitHubOAuth: GitHubOAuthConfig{
					ClientSecret: "client-secret",
				},
				AIService: AIServiceConfig{
					BaseURL: "http://localhost:8000",
					Timeout: 30,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Port:        0,
				DatabaseURL: "postgres://localhost/test",
				JWTSecret:   "secret",
				GitHubOAuth: GitHubOAuthConfig{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
				AIService: AIServiceConfig{
					BaseURL: "http://localhost:8000",
					Timeout: 30,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	cfg := &Config{Environment: "development"}
	if !cfg.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return true")
	}

	cfg.Environment = "production"
	if cfg.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return false")
	}
}

func TestIsProduction(t *testing.T) {
	cfg := &Config{Environment: "production"}
	if !cfg.IsProduction() {
		t.Error("Expected IsProduction() to return true")
	}

	cfg.Environment = "development"
	if cfg.IsProduction() {
		t.Error("Expected IsProduction() to return false")
	}
}

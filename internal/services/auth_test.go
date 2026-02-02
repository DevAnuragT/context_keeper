package services

import (
	"testing"

	"github.com/DevAnuragT/context_keeper/internal/config"
)

func TestAuthService_verifyScopesPresent(t *testing.T) {
	cfg := &config.Config{
		GitHubOAuth: config.GitHubOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)

	tests := []struct {
		name           string
		grantedScopes  string
		requiredScopes []string
		expected       bool
	}{
		{
			name:           "all scopes present",
			grantedScopes:  "public_repo,read:user,user:email",
			requiredScopes: []string{"public_repo", "read:user", "user:email"},
			expected:       true,
		},
		{
			name:           "missing scope",
			grantedScopes:  "public_repo,read:user",
			requiredScopes: []string{"public_repo", "read:user", "user:email"},
			expected:       false,
		},
		{
			name:           "extra scopes allowed",
			grantedScopes:  "public_repo,read:user,user:email,repo",
			requiredScopes: []string{"public_repo", "read:user", "user:email"},
			expected:       true,
		},
		{
			name:           "empty granted scopes",
			grantedScopes:  "",
			requiredScopes: []string{"public_repo"},
			expected:       false,
		},
		{
			name:           "no required scopes",
			grantedScopes:  "public_repo",
			requiredScopes: []string{},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authSvc.verifyScopesPresent(tt.grantedScopes, tt.requiredScopes)
			if result != tt.expected {
				t.Errorf("verifyScopesPresent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAuthService_getUserInfo(t *testing.T) {
	// For now, skip this test as it requires modifying the service to accept custom URLs
	t.Skip("Requires refactoring service to accept custom GitHub API base URL")
}

func TestAuthService_exchangeCodeForToken(t *testing.T) {
	// For now, skip this test as it requires modifying the service to accept custom URLs
	t.Skip("Requires refactoring service to accept custom GitHub OAuth base URL")
}

func TestAuthService_HandleGitHubCallback_Integration(t *testing.T) {
	// This test would require both OAuth and user API endpoints
	// For now, we'll test the individual components
	t.Skip("Integration test requires full mock server setup")
}

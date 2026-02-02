package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/models"
)

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	config     *config.Config
	httpClient *http.Client
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.Config) AuthService {
	return &AuthServiceImpl{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HandleGitHubCallback handles the GitHub OAuth callback
func (a *AuthServiceImpl) HandleGitHubCallback(ctx context.Context, code string) (*models.AuthResponse, error) {
	// Exchange code for access token
	token, err := a.exchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user information
	user, err := a.getUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Store GitHub token in user
	user.GitHubToken = token

	// Generate JWT token
	jwtToken, err := a.GenerateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &models.AuthResponse{
		Token: jwtToken,
		User:  *user,
	}, nil
}

// exchangeCodeForToken exchanges authorization code for access token
func (a *AuthServiceImpl) exchangeCodeForToken(ctx context.Context, code string) (string, error) {
	// Prepare request data
	data := url.Values{}
	data.Set("client_id", a.config.GitHubOAuth.ClientID)
	data.Set("client_secret", a.config.GitHubOAuth.ClientSecret)
	data.Set("code", code)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub OAuth token exchange failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("GitHub OAuth error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token received")
	}

	// Verify scopes (Requirements 1.1)
	requiredScopes := []string{"public_repo", "read:user", "user:email"}
	if !a.verifyScopesPresent(tokenResp.Scope, requiredScopes) {
		return "", fmt.Errorf("insufficient scopes granted: got %s, need %v", tokenResp.Scope, requiredScopes)
	}

	return tokenResp.AccessToken, nil
}

// verifyScopesPresent checks if all required scopes are present
func (a *AuthServiceImpl) verifyScopesPresent(grantedScopes string, requiredScopes []string) bool {
	scopes := strings.Split(grantedScopes, ",")
	scopeMap := make(map[string]bool)
	for _, scope := range scopes {
		scopeMap[strings.TrimSpace(scope)] = true
	}

	for _, required := range requiredScopes {
		if !scopeMap[required] {
			return false
		}
	}
	return true
}

// getUserInfo gets user information from GitHub API
func (a *AuthServiceImpl) getUserInfo(ctx context.Context, token string) (*models.User, error) {
	// Get user profile
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %d", resp.StatusCode)
	}

	var githubUser struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	// Get user email if not in profile (private email)
	email := githubUser.Email
	if email == "" {
		email, _ = a.getUserEmail(ctx, token) // Best effort, don't fail if email is private
	}

	return &models.User{
		ID:    fmt.Sprintf("%d", githubUser.ID),
		Login: githubUser.Login,
		Email: email,
	}, nil
}

// getUserEmail gets user email from GitHub API
func (a *AuthServiceImpl) getUserEmail(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API email request failed with status: %d", resp.StatusCode)
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find primary email
	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	// Fallback to first email
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

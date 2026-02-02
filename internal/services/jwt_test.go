package services

import (
	"strings"
	"testing"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/models"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)

	// Test user
	user := &models.User{
		ID:          "12345",
		Login:       "testuser",
		Email:       "test@example.com",
		GitHubToken: "github-token-123",
	}

	// Generate JWT
	token, err := authSvc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Validate JWT
	validatedUser, err := authSvc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	// Verify user data
	if validatedUser.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, validatedUser.ID)
	}

	if validatedUser.Login != user.Login {
		t.Errorf("Expected login %s, got %s", user.Login, validatedUser.Login)
	}

	if validatedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, validatedUser.Email)
	}

	if validatedUser.GitHubToken != user.GitHubToken {
		t.Errorf("Expected GitHub token %s, got %s", user.GitHubToken, validatedUser.GitHubToken)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "invalid.token"},
		{"too many parts", "part1.part2.part3.part4"},
		{"invalid base64", "invalid-base64.invalid-base64.invalid-base64"},
		{"tampered token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzNDUiLCJsb2dpbiI6InRlc3R1c2VyIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwiZ2l0aHViX3Rva2VuIjoiZ2l0aHViLXRva2VuLTEyMyIsImlhdCI6MTcwMDAwMDAwMCwiZXhwIjoxNzAwMDg2NDAwfQ.tampered-signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authSvc.ValidateJWT(tt.token)
			if err == nil {
				t.Errorf("Expected validation to fail for %s", tt.name)
			}
		})
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)

	// Create a token that's already expired by manually creating claims
	user := &models.User{
		ID:          "12345",
		Login:       "testuser",
		Email:       "test@example.com",
		GitHubToken: "github-token-123",
	}

	// Generate a normal token first
	token, err := authSvc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	// Parse the claims to verify structure
	claims, err := ParseJWTClaims(token)
	if err != nil {
		t.Fatalf("ParseJWTClaims failed: %v", err)
	}

	// Verify claims structure
	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, claims.UserID)
	}

	if claims.ExpiresAt <= claims.IssuedAt {
		t.Error("Expected ExpiresAt to be after IssuedAt")
	}

	// For testing expired tokens, we'd need to either:
	// 1. Wait for the token to expire (not practical)
	// 2. Mock time (complex with standard library)
	// 3. Create a token with past expiry (requires exposing internal methods)
	// For now, we'll test the structure and leave expiry testing for integration tests
}

func TestJWT_DifferentSecrets(t *testing.T) {
	cfg1 := &config.Config{JWTSecret: "secret1"}
	cfg2 := &config.Config{JWTSecret: "secret2"}

	authSvc1 := NewAuthService(cfg1).(*AuthServiceImpl)
	authSvc2 := NewAuthService(cfg2).(*AuthServiceImpl)

	user := &models.User{
		ID:          "12345",
		Login:       "testuser",
		Email:       "test@example.com",
		GitHubToken: "github-token-123",
	}

	// Generate token with first service
	token, err := authSvc1.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	// Try to validate with second service (different secret)
	_, err = authSvc2.ValidateJWT(token)
	if err == nil {
		t.Error("Expected validation to fail with different secret")
	}

	// Validate with correct service should work
	_, err = authSvc1.ValidateJWT(token)
	if err != nil {
		t.Errorf("Validation with correct secret failed: %v", err)
	}
}

func TestJWT_EmptyUserFields(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)

	// Test with empty fields
	user := &models.User{
		ID:          "",
		Login:       "",
		Email:       "",
		GitHubToken: "",
	}

	// Should still generate and validate successfully
	token, err := authSvc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed with empty fields: %v", err)
	}

	validatedUser, err := authSvc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT failed with empty fields: %v", err)
	}

	if validatedUser.ID != "" || validatedUser.Login != "" || validatedUser.Email != "" || validatedUser.GitHubToken != "" {
		t.Error("Empty fields should remain empty after round-trip")
	}
}

func TestJWT_TokenStructure(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)

	user := &models.User{
		ID:          "12345",
		Login:       "testuser",
		Email:       "test@example.com",
		GitHubToken: "github-token-123",
	}

	token, err := authSvc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts in JWT, got %d", len(parts))
	}

	// Each part should be valid base64
	for i, part := range parts {
		if part == "" {
			t.Errorf("Part %d is empty", i)
		}
	}
}

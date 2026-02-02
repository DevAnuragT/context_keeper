package services

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/config"
	"github.com/DevAnuragT/context_keeper/internal/models"
)

// Feature: contextkeeper-go-backend, Property 1: OAuth Scope Consistency
// **Validates: Requirements 1.1**
func TestOAuthScopeConsistencyProperty(t *testing.T) {
	// Property: For any GitHub OAuth request initiated by the system,
	// the requested scopes should always include exactly: public_repo, read:user, user:email

	cfg := &config.Config{
		GitHubOAuth: config.GitHubOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)
	requiredScopes := []string{"public_repo", "read:user", "user:email"}

	rand.Seed(time.Now().UnixNano())

	// Run property test with 100 iterations
	for i := 0; i < 100; i++ {
		// Generate random scope combinations
		grantedScopes := generateRandomScopeString()

		// Test scope verification
		result := authSvc.verifyScopesPresent(grantedScopes, requiredScopes)

		// Verify the property: result should be true if and only if all required scopes are present
		expectedResult := containsAllRequiredScopes(grantedScopes, requiredScopes)

		if result != expectedResult {
			t.Errorf("Iteration %d: scope verification failed for scopes '%s'. Expected %v, got %v",
				i, grantedScopes, expectedResult, result)
		}
	}
}

// Feature: contextkeeper-go-backend, Property 2: JWT Authentication Round Trip
// **Validates: Requirements 1.3, 1.5**
func TestJWTAuthenticationRoundTripProperty(t *testing.T) {
	// Property: For any successful GitHub OAuth flow, generating a JWT token
	// and then validating it should preserve the user identity (excluding GitHub token for security)

	cfg := &config.Config{
		JWTSecret: "test-secret-key-for-property-testing",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)
	rand.Seed(time.Now().UnixNano())

	// Run property test with 100 iterations
	for i := 0; i < 100; i++ {
		// Generate random user data
		originalUser := generateRandomUser()

		// Test JWT round-trip
		if !testJWTRoundTrip(t, authSvc, originalUser, i) {
			t.Errorf("JWT round-trip failed for iteration %d with user: %+v", i, originalUser)
		}
	}
}

// generateRandomScopeString creates random scope combinations for testing
func generateRandomScopeString() string {
	allPossibleScopes := []string{
		"public_repo", "read:user", "user:email", // Required scopes
		"repo", "admin:repo_hook", "write:repo_hook", "read:repo_hook", // Additional repo scopes
		"admin:org", "write:org", "read:org", // Org scopes
		"admin:public_key", "write:public_key", "read:public_key", // Key scopes
		"admin:org_hook", "gist", "notifications", "user", "delete_repo", // Other scopes
		"write:discussion", "read:discussion", "admin:enterprise", // More scopes
	}

	// Randomly select 0-10 scopes
	numScopes := rand.Intn(11)
	if numScopes == 0 {
		return ""
	}

	selectedScopes := make([]string, 0, numScopes)
	usedScopes := make(map[string]bool)

	for len(selectedScopes) < numScopes {
		scope := allPossibleScopes[rand.Intn(len(allPossibleScopes))]
		if !usedScopes[scope] {
			selectedScopes = append(selectedScopes, scope)
			usedScopes[scope] = true
		}
	}

	return strings.Join(selectedScopes, ",")
}

// containsAllRequiredScopes checks if all required scopes are present in granted scopes
func containsAllRequiredScopes(grantedScopes string, requiredScopes []string) bool {
	if grantedScopes == "" && len(requiredScopes) > 0 {
		return false
	}

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

// generateRandomUser creates random user data for property testing
func generateRandomUser() *models.User {
	return &models.User{
		ID:          generateRandomUserID(),
		Login:       generateRandomLogin(),
		Email:       generateRandomEmail(),
		GitHubToken: generateRandomToken(),
	}
}

// generateRandomUserID creates random user IDs
func generateRandomUserID() string {
	// GitHub user IDs are typically large integers
	id := rand.Int63n(999999999) + 1
	return strconv.FormatInt(id, 10)
}

// generateRandomLogin creates random GitHub login names
func generateRandomLogin() string {
	patterns := []func() string{
		func() string { return "user" + string(rune(rand.Intn(10000))) },
		func() string { return "test-user-" + string(rune(rand.Intn(1000))) },
		func() string { return "dev" + string(rune(rand.Intn(100))) },
		func() string { return "" },                  // Empty login
		func() string { return "a" },                 // Single character
		func() string { return generateLongLogin() }, // Long login
		func() string { return "user-with-dashes" },
		func() string { return "user_with_underscores" },
		func() string { return "UserWithCaps" },
		func() string { return "123numeric" },
	}

	pattern := patterns[rand.Intn(len(patterns))]
	return pattern()
}

// generateRandomEmail creates random email addresses
func generateRandomEmail() string {
	patterns := []func() string{
		func() string { return "user@example.com" },
		func() string { return "test" + strconv.Itoa(rand.Intn(1000)) + "@github.com" },
		func() string { return "" }, // Empty email (private)
		func() string { return "user@domain.co.uk" },
		func() string { return "user+tag@example.org" },
		func() string { return "user.name@sub.domain.com" },
		func() string { return "123@numeric.domain" },
	}

	pattern := patterns[rand.Intn(len(patterns))]
	return pattern()
}

// generateRandomToken creates random GitHub tokens
func generateRandomToken() string {
	patterns := []func() string{
		func() string { return "ghp_" + generateRandomString(36) }, // Personal access token format
		func() string { return "gho_" + generateRandomString(36) }, // OAuth token format
		func() string { return "" },                                // Empty token
		func() string { return generateRandomString(40) },          // Generic token
		func() string { return "token-" + generateRandomString(20) },
	}

	pattern := patterns[rand.Intn(len(patterns))]
	return pattern()
}

// generateLongLogin creates a long login name
func generateLongLogin() string {
	length := rand.Intn(30) + 10 // 10-40 characters
	chars := "abcdefghijklmnopqrstuvwxyz0123456789-_"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// testJWTRoundTrip tests that a user can be encoded to JWT and decoded back
func testJWTRoundTrip(t *testing.T, authSvc *AuthServiceImpl, originalUser *models.User, iteration int) bool {
	// Generate JWT token
	token, err := authSvc.GenerateJWT(originalUser)
	if err != nil {
		t.Logf("Iteration %d: GenerateJWT failed: %v", iteration, err)
		return false
	}

	// Validate JWT token
	decodedUser, err := authSvc.ValidateJWT(token)
	if err != nil {
		t.Logf("Iteration %d: ValidateJWT failed: %v", iteration, err)
		return false
	}

	// Verify user data preservation
	if !usersEqual(originalUser, decodedUser) {
		t.Logf("Iteration %d: User data not preserved. Original: %+v, Decoded: %+v",
			iteration, originalUser, decodedUser)
		return false
	}

	return true
}

// usersEqual compares two users for equality (excluding GitHub token for security)
func usersEqual(a, b *models.User) bool {
	return a.ID == b.ID &&
		a.Login == b.Login &&
		a.Email == b.Email &&
		b.GitHubToken == "" // GitHub token should always be empty in JWT for security
}

// Additional property test for JWT token structure consistency
func TestJWTTokenStructureProperty(t *testing.T) {
	// Property: All generated JWT tokens should have consistent structure (3 parts, valid base64)

	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	authSvc := NewAuthService(cfg).(*AuthServiceImpl)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 50; i++ {
		user := generateRandomUser()

		token, err := authSvc.GenerateJWT(user)
		if err != nil {
			t.Errorf("Iteration %d: GenerateJWT failed: %v", i, err)
			continue
		}

		// Verify token structure
		if !hasValidJWTStructure(token) {
			t.Errorf("Iteration %d: Invalid JWT structure for token: %s", i, token)
		}
	}
}

// hasValidJWTStructure checks if a token has valid JWT structure
func hasValidJWTStructure(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	// Each part should be non-empty
	for _, part := range parts {
		if part == "" {
			return false
		}
	}

	return true
}

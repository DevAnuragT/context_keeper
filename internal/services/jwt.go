package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/models"
)

// JWT header structure
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWT claims structure
type jwtClaims struct {
	UserID      string `json:"user_id"`
	Login       string `json:"login"`
	Email       string `json:"email"`
	GitHubToken string `json:"github_token"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
}

// GenerateJWT generates a JWT token for a user
func (a *AuthServiceImpl) GenerateJWT(user *models.User) (string, error) {
	// Create header
	header := jwtHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	// Create claims
	now := time.Now()
	claims := jwtClaims{
		UserID:      user.ID,
		Login:       user.Login,
		Email:       user.Email,
		GitHubToken: user.GitHubToken,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(24 * time.Hour).Unix(), // 24 hour expiry
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Encode header and claims
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := a.createSignature(message)

	// Combine all parts
	token := message + "." + signature

	return token, nil
}

// ValidateJWT validates a JWT token and returns user claims
func (a *AuthServiceImpl) ValidateJWT(token string) (*models.User, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	message := headerEncoded + "." + claimsEncoded
	expectedSignature := a.createSignature(message)
	if signatureEncoded != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}

	// Decode and validate header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	if header.Alg != "HS256" || header.Typ != "JWT" {
		return nil, fmt.Errorf("unsupported token type or algorithm")
	}

	// Decode and validate claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	now := time.Now().Unix()
	if claims.ExpiresAt < now {
		return nil, fmt.Errorf("token has expired")
	}

	// Check issued at (not in future)
	if claims.IssuedAt > now+60 { // Allow 60 seconds clock skew
		return nil, fmt.Errorf("token issued in the future")
	}

	// Return user
	return &models.User{
		ID:          claims.UserID,
		Login:       claims.Login,
		Email:       claims.Email,
		GitHubToken: claims.GitHubToken,
	}, nil
}

// createSignature creates HMAC-SHA256 signature for JWT
func (a *AuthServiceImpl) createSignature(message string) string {
	h := hmac.New(sha256.New, []byte(a.config.JWTSecret))
	h.Write([]byte(message))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}

// ParseJWTClaims parses JWT claims without validation (for testing)
func ParseJWTClaims(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return &claims, nil
}

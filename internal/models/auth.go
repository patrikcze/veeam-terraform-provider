package models

import (
	"time"
)

// AuthRequest represents the authentication request payload
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain,omitempty"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// TokenRefreshRequest represents the token refresh request payload
type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type"`
}

// TokenInfo holds token information and expiration details
type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scope        string
}

// IsExpired checks if the token has expired
func (t *TokenInfo) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// WillExpireSoon checks if the token will expire within the given duration
func (t *TokenInfo) WillExpireSoon(buffer time.Duration) bool {
	return time.Now().Add(buffer).After(t.ExpiresAt)
}

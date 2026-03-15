package models

import (
	"fmt"
	"time"
)

// TokenModel represents the V13 OAuth2 token response.
// Matches the Veeam V13 swagger TokenModel schema.
// POST /api/oauth2/token returns this.
type TokenModel struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Issued       string `json:".issued"`
	Expires      string `json:".expires"`
}

// TokenInfo holds the active token state for the client.
// This is the internal representation — never serialized or logged.
type TokenInfo struct {
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenType    string    `json:"-"`
	ExpiresAt    time.Time `json:"-"`
}

// IsExpired checks if the access token has expired.
func (t *TokenInfo) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// WillExpireSoon checks if the access token will expire within the given buffer.
func (t *TokenInfo) WillExpireSoon(buffer time.Duration) bool {
	return time.Now().Add(buffer).After(t.ExpiresAt)
}

// String redacts sensitive fields to prevent accidental logging.
func (t *TokenInfo) String() string {
	return fmt.Sprintf("TokenInfo{ExpiresAt: %s, Expired: %v}", t.ExpiresAt.Format(time.RFC3339), t.IsExpired())
}

// APIError represents the Veeam V13 error response body.
// The API returns this on 4xx/5xx responses.
type APIError struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Details   string `json:"details"`
}

// Error implements the error interface with an actionable message.
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.ErrorCode, e.Message, e.Details)
	}
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
	}
	return e.ErrorCode
}

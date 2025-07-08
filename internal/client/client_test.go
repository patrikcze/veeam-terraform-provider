package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestNewVeeamClient(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/token" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "test-token",
				"token_type": "Bearer",
				"refresh_token": "test-refresh-token",
				"expires_in": 3600
			}`))
		}
	}))
	defer server.Close()

	// Test successful client creation
	client, err := NewVeeamClient(context.Background(), server.URL, "testuser", "testpass", false)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, server.URL, client.BaseURL)
	assert.Equal(t, "test-token", client.TokenInfo.AccessToken)
	assert.Equal(t, "Bearer", client.TokenInfo.TokenType)
	assert.Equal(t, "test-refresh-token", client.TokenInfo.RefreshToken)
}

func TestVeeamClient_authenticate(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/token" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "test-token",
				"token_type": "Bearer",
				"refresh_token": "test-refresh-token",
				"expires_in": 3600
			}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Test successful authentication
	err := client.authenticate(context.Background(), "testuser", "testpass")

	require.NoError(t, err)
	assert.Equal(t, "test-token", client.TokenInfo.AccessToken)
	assert.Equal(t, "Bearer", client.TokenInfo.TokenType)
	assert.Equal(t, "test-refresh-token", client.TokenInfo.RefreshToken)
}

func TestVeeamClient_authenticate_Failure(t *testing.T) {
	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/token" && r.Method == "POST" {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Test authentication failure
	err := client.authenticate(context.Background(), "testuser", "wrongpass")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed with status")
}

func TestVeeamClient_RefreshToken(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/refresh" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "new-test-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken:  "old-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().Add(1 * time.Minute), // Will expire soon
		},
	}

	// Test token refresh
	err := client.RefreshToken(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "new-test-token", client.TokenInfo.AccessToken)
}

func TestVeeamClient_RefreshToken_NotNeeded(t *testing.T) {
	client := &VeeamClient{
		BaseURL:    "http://test.com",
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken:  "valid-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().Add(10 * time.Minute), // Not expiring soon
		},
	}

	// Test that refresh is not needed
	err := client.RefreshToken(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "valid-token", client.TokenInfo.AccessToken) // Should remain unchanged
}

func TestVeeamClient_RefreshToken_Failure(t *testing.T) {
	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/refresh" && r.Method == "POST" {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken:  "old-token",
			RefreshToken: "invalid-refresh-token",
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().Add(1 * time.Minute), // Will expire soon
		},
	}

	// Test token refresh failure
	err := client.RefreshToken(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to refresh token")
}

func TestTokenInfo_IsExpired(t *testing.T) {
	// Test expired token
	expiredToken := models.TokenInfo{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, expiredToken.IsExpired())

	// Test valid token
	validToken := models.TokenInfo{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, validToken.IsExpired())
}

func TestTokenInfo_WillExpireSoon(t *testing.T) {
	// Test token that will expire soon
	soonToExpireToken := models.TokenInfo{
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}
	assert.True(t, soonToExpireToken.WillExpireSoon(5*time.Minute))

	// Test token that won't expire soon
	validToken := models.TokenInfo{
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	assert.False(t, validToken.WillExpireSoon(5*time.Minute))
}

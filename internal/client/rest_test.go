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

func TestVeeamClient_GET(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	// Test GET request
	resp, err := client.GET(context.Background(), "/api/v1/test")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestVeeamClient_POST(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "123", "status": "created"}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	payload := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	// Test POST request
	resp, err := client.POST(context.Background(), "/api/v1/test", payload)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
}

func TestVeeamClient_PUT(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test/123" && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "123", "status": "updated"}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	payload := map[string]interface{}{
		"name":  "updated_test",
		"value": 456,
	}

	// Test PUT request
	resp, err := client.PUT(context.Background(), "/api/v1/test/123", payload)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestVeeamClient_DELETE(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test/123" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	// Test DELETE request
	resp, err := client.DELETE(context.Background(), "/api/v1/test/123")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestVeeamClient_GetJSON(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success", "value": 42}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	var result map[string]interface{}

	// Test GetJSON request
	err := client.GetJSON(context.Background(), "/api/v1/test", &result)

	require.NoError(t, err)
	assert.Equal(t, "success", result["message"])
	assert.Equal(t, float64(42), result["value"])
}

func TestVeeamClient_PostJSON(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "123", "status": "created"}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	payload := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	var result map[string]interface{}

	// Test PostJSON request
	err := client.PostJSON(context.Background(), "/api/v1/test", payload, &result)

	require.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "created", result["status"])
}

func TestVeeamClient_PutJSON(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test/123" && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "123", "status": "updated"}`))
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	payload := map[string]interface{}{
		"name":  "updated_test",
		"value": 456,
	}

	var result map[string]interface{}

	// Test PutJSON request
	err := client.PutJSON(context.Background(), "/api/v1/test/123", payload, &result)

	require.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "updated", result["status"])
}

func TestVeeamClient_DeleteJSON(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test/123" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	// Test DeleteJSON request
	err := client.DeleteJSON(context.Background(), "/api/v1/test/123")

	require.NoError(t, err)
}

func TestVeeamClient_GetJSON_ErrorResponse(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test" && r.Method == "GET" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	var result map[string]interface{}

	// Test GetJSON request with error
	err := client.GetJSON(context.Background(), "/api/v1/test", &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500")
}

func TestVeeamClient_PostJSON_ErrorResponse(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/test" && r.Method == "POST" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := &VeeamClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		TokenInfo: models.TokenInfo{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	payload := map[string]interface{}{
		"name": "test",
	}

	var result map[string]interface{}

	// Test PostJSON request with error
	err := client.PostJSON(context.Background(), "/api/v1/test", payload, &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}

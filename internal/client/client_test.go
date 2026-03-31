package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestTokenResponse returns a valid V13 TokenModel JSON response.
func newTestTokenResponse(accessToken, refreshToken string, expiresIn int) []byte {
	resp := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"refresh_token": refreshToken,
		"expires_in":    expiresIn,
		".issued":       time.Now().UTC().Format(time.RFC3339),
		".expires":      time.Now().Add(time.Duration(expiresIn) * time.Second).UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(resp)
	return b
}

// newTokenServer creates an httptest TLS server that validates OAuth2 token requests.
func newTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate endpoint
		assert.Equal(t, "/api/oauth2/token", r.URL.Path, "wrong token endpoint")
		assert.Equal(t, http.MethodPost, r.Method, "wrong HTTP method")
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"), "wrong content type")
		assert.Equal(t, APIVersion, r.Header.Get("x-api-version"), "missing x-api-version header")

		// Parse form body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		values, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		grantType := values.Get("grant_type")

		switch grantType {
		case "password":
			assert.NotEmpty(t, values.Get("username"), "missing username")
			assert.NotEmpty(t, values.Get("password"), "missing password")
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-access-token", "test-refresh-token", 900))

		case "refresh_token":
			assert.NotEmpty(t, values.Get("refresh_token"), "missing refresh_token")
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("refreshed-access-token", "new-refresh-token", 900))

		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": "InvalidGrantType",
				"message":   "unsupported grant_type: " + grantType,
			})
		}
	}))
}

func TestNewVeeamClientWithHTTPClient_Success(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())

	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "test-access-token", c.TokenInfo.AccessToken)
	assert.Equal(t, "test-refresh-token", c.TokenInfo.RefreshToken)
	assert.Equal(t, "bearer", c.TokenInfo.TokenType)
	assert.False(t, c.TokenInfo.IsExpired())
}

func TestNewVeeamClientWithHTTPClient_AuthFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": "AuthenticationFailed",
			"message":   "Invalid credentials",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "bad-user", "bad-pass", server.Client())

	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestNewVeeamClientWithHTTPClient_EmptyTokenResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "",
			"token_type":   "bearer",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "pass", server.Client())

	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "missing access_token")
}

func TestRefreshToken_NotNeeded(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// Token is fresh (expires in 900s), refresh should be a no-op
	originalToken := c.TokenInfo.AccessToken
	err = c.RefreshToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, originalToken, c.TokenInfo.AccessToken, "token should not have changed")
}

func TestRefreshToken_ExpiringSoon(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// Simulate token about to expire
	c.TokenInfo.ExpiresAt = time.Now().Add(30 * time.Second)

	err = c.RefreshToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "refreshed-access-token", c.TokenInfo.AccessToken)
	assert.Equal(t, "new-refresh-token", c.TokenInfo.RefreshToken)
}

func TestRefreshToken_FallbackToReauth(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		callCount++

		if values.Get("grant_type") == "refresh_token" && callCount > 1 {
			// Fail the refresh
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": "InvalidGrant",
				"message":   "refresh token expired",
			})
			return
		}

		// Password auth always succeeds
		w.WriteHeader(http.StatusOK)
		w.Write(newTestTokenResponse("reauth-token", "reauth-refresh", 900))
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// Force token to expire soon
	c.TokenInfo.ExpiresAt = time.Now().Add(30 * time.Second)

	// This should try refresh, fail, then re-authenticate
	err = c.RefreshToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "reauth-token", c.TokenInfo.AccessToken)
}

func TestTokenInfo_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"not expired", time.Now().Add(10 * time.Minute), false},
		{"expired", time.Now().Add(-1 * time.Minute), true},
		{"just expired", time.Now().Add(-1 * time.Second), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &VeeamClient{}
			c.TokenInfo.ExpiresAt = tt.expiresAt
			assert.Equal(t, tt.want, c.TokenInfo.IsExpired())
		})
	}
}

func TestTokenInfo_WillExpireSoon(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		buffer    time.Duration
		want      bool
	}{
		{"expiring soon", time.Now().Add(2 * time.Minute), 5 * time.Minute, true},
		{"not expiring soon", time.Now().Add(10 * time.Minute), 5 * time.Minute, false},
		{"already expired", time.Now().Add(-1 * time.Minute), 5 * time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &VeeamClient{}
			c.TokenInfo.ExpiresAt = tt.expiresAt
			assert.Equal(t, tt.want, c.TokenInfo.WillExpireSoon(tt.buffer))
		})
	}
}

// --- NewVeeamClient tests ---

func TestNewVeeamClient_EmptyHost(t *testing.T) {
	ctx := context.Background()
	c, err := NewVeeamClient(ctx, "", 9419, "admin", "secret", false)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "host is empty")
}

func TestNewVeeamClient_InsecureTrue(t *testing.T) {
	// httptest.NewTLSServer uses a self-signed certificate. NewVeeamClient with
	// insecure=true builds a transport with InsecureSkipVerify=true, which will
	// accept that certificate. This exercises the insecure warning-log branch
	// and the full client construction + authentication path.
	server := newTokenServer(t)
	defer server.Close()

	// Parse "https://127.0.0.1:<port>" into separate host and port.
	// net/url gives us the host:port string; we split it manually.
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	hostOnly := parsedURL.Hostname()
	portStr := parsedURL.Port()
	var port int
	_, scanErr := fmt.Sscan(portStr, &port)
	require.NoError(t, scanErr)

	ctx := context.Background()
	c, err := NewVeeamClient(ctx, hostOnly, port, "admin", "secret", true)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "test-access-token", c.TokenInfo.AccessToken)
}

func TestNewVeeamClient_AuthFailure(t *testing.T) {
	// Spin up a server that always rejects authentication.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": "Unauthorized",
			"message":   "bad credentials",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	// Use WithHTTPClient so we can supply the test server TLS client.
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "wrong", "creds", server.Client())
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "authentication failed")
}

// --- RefreshToken failure: refresh fails AND re-auth fails ---

func TestRefreshToken_RefreshAndReauthBothFail(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		callCount++

		if callCount == 1 && values.Get("grant_type") == "password" {
			// First call: initial auth succeeds
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("initial-token", "initial-refresh", 900))
			return
		}

		// All subsequent calls fail (both refresh_token and fallback password grant)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": "Unauthorized",
			"message":   "credentials revoked",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// Force token to expire soon so RefreshToken actually attempts a refresh
	c.TokenInfo.ExpiresAt = time.Now().Add(30 * time.Second)

	err = c.RefreshToken(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

// --- postTokenRequest: malformed JSON response ---

func TestPostTokenRequest_MalformedJSONResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return non-JSON to trigger unmarshal error in postTokenRequest
		w.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "failed to parse token response")
}

// --- doRequest: marshal failure ---

func TestDoRequest_MarshalFailure(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// A channel is not JSON-serializable — json.Marshal will return an error.
	// PostJSON passes the payload through doRequest → json.Marshal, so this
	// exercises the marshal-failure branch in doRequest.
	err = c.PostJSON(ctx, "/api/v1/test", make(chan int), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal request payload")
}

// --- readAndClose error paths in HTTP verb methods ---
//
// We inject a custom RoundTripper that (a) handles the initial auth request
// normally so the client can be constructed, then (b) returns a response whose
// body always errors on Read for subsequent requests.

// brokenBodyTransport wraps an inner RoundTripper. The first request (token
// auth) is delegated normally; subsequent requests return a response with a
// body that errors on Read.
type brokenBodyTransport struct {
	inner     http.RoundTripper
	callCount int
	readErr   error
}

func (t *brokenBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.callCount++
	if req.URL.Path == PathOAuth2Token {
		// Let auth succeed normally so the client can initialise.
		return t.inner.RoundTrip(req)
	}
	// Return a 200 with a body that always fails on Read.
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       errReader{err: t.readErr},
		Header:     make(http.Header),
	}, nil
}

func newBrokenBodyClient(t *testing.T, tokenServer *httptest.Server) *VeeamClient {
	t.Helper()
	transport := &brokenBodyTransport{
		inner:   tokenServer.Client().Transport,
		readErr: errors.New("simulated network read failure"),
	}
	httpClient := &http.Client{Transport: transport}
	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, tokenServer.URL, "admin", "secret", httpClient)
	require.NoError(t, err)
	return c
}

func TestGetJSON_ReadBodyError(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()
	c := newBrokenBodyClient(t, server)

	var result map[string]string
	err := c.GetJSON(context.Background(), "/api/v1/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read response body")
}

func TestPostJSON_ReadBodyError(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()
	c := newBrokenBodyClient(t, server)

	err := c.PostJSON(context.Background(), "/api/v1/test", map[string]string{"k": "v"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read response body")
}

func TestPutJSON_ReadBodyError(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()
	c := newBrokenBodyClient(t, server)

	err := c.PutJSON(context.Background(), "/api/v1/test", map[string]string{"k": "v"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read response body")
}

func TestDeleteJSON_ReadBodyError(t *testing.T) {
	server := newTokenServer(t)
	defer server.Close()
	c := newBrokenBodyClient(t, server)

	err := c.DeleteJSON(context.Background(), "/api/v1/test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read response body")
}

// --- postTokenRequest: 4xx with no JSON message field ---

func TestPostTokenRequest_4xxNoMessage(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		// Body is valid JSON but has no "message" field — hits the plain-status error path.
		w.Write([]byte(`{"code":401}`))
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "401")
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"plain hostname", "veeam.example.com", 9419, "https://veeam.example.com:9419"},
		{"with https scheme", "https://veeam.example.com", 9419, "https://veeam.example.com:9419"},
		{"with http scheme", "http://veeam.example.com", 9419, "https://veeam.example.com:9419"},
		{"with existing port", "veeam.example.com:1234", 9419, "https://veeam.example.com:1234"},
		{"IP address", "192.168.1.100", 9419, "https://192.168.1.100:9419"},
		{"trailing slash", "veeam.example.com/", 9419, "https://veeam.example.com:9419"},
		{"empty", "", 9419, ""},
		{"custom port", "veeam.example.com", 8443, "https://veeam.example.com:8443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeURL(tt.host, tt.port)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetJSON_WithAPIVersionHeader(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		// Verify headers on API requests
		assert.Equal(t, APIVersion, r.Header.Get("x-api-version"), "missing x-api-version")
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"), "wrong auth header")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "test-123"})
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	var result map[string]string
	err = c.GetJSON(ctx, "/api/v1/test", &result)
	require.NoError(t, err)
	assert.Equal(t, "test-123", result["id"])
}

func TestPostJSON_ErrorParsing(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": "AlreadyExists",
			"message":   "Resource already exists",
			"details":   "A repository with name 'Default' already exists.",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.PostJSON(ctx, "/api/v1/repositories", map[string]string{"name": "Default"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AlreadyExists")
	assert.Contains(t, err.Error(), "Resource already exists")
	assert.Contains(t, err.Error(), "409")
}

func TestDeleteJSON_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.DeleteJSON(ctx, "/api/v1/credentials/abc-123")
	assert.NoError(t, err)
}

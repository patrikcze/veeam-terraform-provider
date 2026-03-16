package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

const (
	// APIVersion is the Veeam V13 REST API version header value.
	APIVersion = "1.3-rev1"

	// tokenEndpoint references the centralized path from endpoints.go.
	tokenEndpoint = PathOAuth2Token

	// tokenRefreshBuffer is how far before expiry we proactively refresh.
	tokenRefreshBuffer = 2 * time.Minute

	// defaultTimeout for HTTP requests.
	defaultTimeout = 60 * time.Second
)

// VeeamClient implements APIClient for the Veeam V13 REST API.
type VeeamClient struct {
	BaseURL    string
	HTTPClient *http.Client
	TokenInfo  models.TokenInfo

	// credentials stored for re-authentication if refresh token expires.
	// NEVER logged, serialized, or exposed.
	username string
	password string

	// mu protects token refresh from concurrent goroutines.
	mu sync.Mutex
}

// normalizeURL ensures the URL has a proper https:// scheme.
// Security: defaults to https, never http.
func normalizeURL(host string, port int) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	// Strip any existing scheme for normalization
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")

	// Strip trailing slashes and port if already present
	host = strings.TrimRight(host, "/")
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		// Host already has a port, use as-is
		return fmt.Sprintf("https://%s", host)
	}

	return fmt.Sprintf("https://%s:%d", host, port)
}

// NewVeeamClient initializes and authenticates a new Veeam V13 API client.
func NewVeeamClient(ctx context.Context, host string, port int, username, password string, insecure bool) (*VeeamClient, error) {
	// Log host only, NEVER credentials
	tflog.Debug(ctx, "Initializing Veeam client", map[string]interface{}{"host": host, "port": port})

	if insecure {
		tflog.Warn(ctx, "TLS certificate verification is DISABLED — do not use in production")
	}

	baseURL := normalizeURL(host, port)
	if baseURL == "" {
		return nil, fmt.Errorf("failed to initialize client: host is empty")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure, //nolint:gosec // user-controlled flag with warning
		},
	}

	c := &VeeamClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
		username: username,
		password: password,
	}

	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("failed to authenticate with Veeam server at %s: %w", host, err)
	}

	tflog.Info(ctx, "Veeam client initialized successfully", map[string]interface{}{"host": host})
	return c, nil
}

// NewVeeamClientWithHTTPClient creates a client with a custom http.Client (for testing).
func NewVeeamClientWithHTTPClient(ctx context.Context, baseURL, username, password string, httpClient *http.Client) (*VeeamClient, error) {
	c := &VeeamClient{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		username:   username,
		password:   password,
	}

	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return c, nil
}

// authenticate performs OAuth2 password grant against POST /api/oauth2/token.
// Content-Type: application/x-www-form-urlencoded
// Body: grant_type=password&username=USER&password=PASS
func (c *VeeamClient) authenticate(ctx context.Context) error {
	tflog.Debug(ctx, "Authenticating with Veeam server")

	formData := url.Values{
		"grant_type": {"password"},
		"username":   {c.username},
		"password":   {c.password},
	}

	tokenModel, err := c.postTokenRequest(ctx, formData)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.TokenInfo = models.TokenInfo{
		AccessToken:  tokenModel.AccessToken,
		TokenType:    tokenModel.TokenType,
		RefreshToken: tokenModel.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenModel.ExpiresIn) * time.Second),
	}

	tflog.Debug(ctx, "Authentication successful", map[string]interface{}{
		"expires_in_seconds": tokenModel.ExpiresIn,
	})

	return nil
}

// RefreshToken refreshes the access token before it expires.
// Uses grant_type=refresh_token. If the refresh token itself is expired,
// falls back to full re-authentication with stored credentials.
func (c *VeeamClient) RefreshToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// No refresh needed yet
	if !c.TokenInfo.WillExpireSoon(tokenRefreshBuffer) {
		return nil
	}

	tflog.Debug(ctx, "Access token expiring soon, refreshing")

	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.TokenInfo.RefreshToken},
	}

	tokenModel, err := c.postTokenRequest(ctx, formData)
	if err != nil {
		// Refresh failed — try full re-authentication
		tflog.Warn(ctx, "Token refresh failed, attempting full re-authentication")
		return c.authenticate(ctx)
	}

	// Update tokens (refresh token is single-use in V13, we get a new one)
	c.TokenInfo.AccessToken = tokenModel.AccessToken
	c.TokenInfo.RefreshToken = tokenModel.RefreshToken
	c.TokenInfo.ExpiresAt = time.Now().Add(time.Duration(tokenModel.ExpiresIn) * time.Second)

	tflog.Debug(ctx, "Token refreshed successfully")
	return nil
}

// postTokenRequest sends a form-encoded POST to the token endpoint
// and returns the parsed TokenModel.
func (c *VeeamClient) postTokenRequest(ctx context.Context, formData url.Values) (*models.TokenModel, error) {
	tokenURL := c.BaseURL + tokenEndpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-api-version", APIVersion)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response for actionable message
		var apiErr models.APIError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("token request failed (HTTP %d): %w", resp.StatusCode, &apiErr)
		}
		return nil, fmt.Errorf("token request failed with HTTP %d", resp.StatusCode)
	}

	var tokenModel models.TokenModel
	if err := json.Unmarshal(body, &tokenModel); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenModel.AccessToken == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}

	return &tokenModel, nil
}

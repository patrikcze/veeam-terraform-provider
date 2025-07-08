package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// VeeamClient represents the client structure
type VeeamClient struct {
	BaseURL    string
	HTTPClient *http.Client
	TokenInfo  models.TokenInfo
}

// normalizeURL ensures the URL has a proper protocol scheme
func normalizeURL(url string) string {
	if url == "" {
		return url
	}

	// If URL already has a protocol scheme, return as-is
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	// Default to https:// for security
	return "https://" + url
}

// NewVeeamClient initializes a new Veeam API client
func NewVeeamClient(ctx context.Context, baseURL, username, password string, insecure bool) (*VeeamClient, error) {
	tflog.Debug(ctx, "Initializing Veeam client", map[string]interface{}{"host": baseURL})
	normalizedURL := normalizeURL(baseURL)

	// Configure HTTP client with TLS settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	client := &VeeamClient{
		BaseURL: normalizedURL,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}

	// Authenticate and set token
	if err := client.authenticate(ctx, username, password); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return client, nil
}

// authenticate handles authentication and token retrieval
func (c *VeeamClient) authenticate(ctx context.Context, username, password string) error {
	tflog.Debug(ctx, "Authenticating Veeam client", nil)
	url := c.BaseURL + "/api/v1/token"

	authReq := models.AuthRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to execute authentication request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var authResp models.AuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return err
	}

	c.TokenInfo = models.TokenInfo{
		AccessToken:  authResp.AccessToken,
		TokenType:    authResp.TokenType,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
	}

	return nil
}

// RefreshToken refreshes the access token before it expires
func (c *VeeamClient) RefreshToken(ctx context.Context) error {
	tflog.Debug(ctx, "Refreshing access token", nil)
	if !c.TokenInfo.WillExpireSoon(5 * time.Minute) {
		return nil
	}

	url := c.BaseURL + "/api/v1/refresh"

	refreshReq := models.TokenRefreshRequest{
		RefreshToken: c.TokenInfo.RefreshToken,
		GrantType:    "refresh_token",
	}

	body, err := json.Marshal(refreshReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to refresh token")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var authResp models.AuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return err
	}

	c.TokenInfo.AccessToken = authResp.AccessToken
	c.TokenInfo.ExpiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return nil
}

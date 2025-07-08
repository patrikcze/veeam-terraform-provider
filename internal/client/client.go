package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// VeeamClient represents the client structure
type VeeamClient struct {
	BaseURL    string
	HTTPClient *http.Client
	TokenInfo  models.TokenInfo
}

// NewVeeamClient initializes a new Veeam API client
func NewVeeamClient(baseURL, username, password string) (*VeeamClient, error) {
	client := &VeeamClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Authenticate and set token
	if err := client.authenticate(username, password); err != nil {
		return nil, err
	}

	return client, nil
}

// authenticate handles authentication and token retrieval
func (c *VeeamClient) authenticate(username, password string) error {
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to authenticate")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
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
func (c *VeeamClient) RefreshToken() error {
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to refresh token")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
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

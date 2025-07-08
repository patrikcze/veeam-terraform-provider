package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/patrikcze/terraform-provider-veeam/internal/utils"
)

// GET performs a GET request to the specified endpoint
func (c *VeeamClient) GET(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "GET", endpoint, nil)
}

// POST performs a POST request to the specified endpoint with a JSON payload
func (c *VeeamClient) POST(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "POST", endpoint, payload)
}

// PUT performs a PUT request to the specified endpoint with a JSON payload
func (c *VeeamClient) PUT(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	return c.doRequest(ctx, "PUT", endpoint, payload)
}

// DELETE performs a DELETE request to the specified endpoint
func (c *VeeamClient) DELETE(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "DELETE", endpoint, nil)
}

// doRequest is the internal method that handles all HTTP requests
func (c *VeeamClient) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (*http.Response, error) {
	tflog.Debug(ctx, "Making API request", map[string]interface{}{"method": method, "endpoint": endpoint})
	// Refresh token if needed
	if err := c.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Prepare the URL
	url := c.BaseURL + endpoint
	if !strings.HasPrefix(endpoint, "/") {
		url = c.BaseURL + "/" + endpoint
	}

	// Prepare the request body
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	// Create the request with retry logic
	return utils.RetryRequest(func() (*http.Response, error) {
		req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.TokenInfo.AccessToken))

		// Execute the request
		resp, err := c.HTTPClient.Do(req.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		return resp, nil
	}, 3, utils.DefaultRetryPolicy)
}

// GetJSON performs a GET request and unmarshals the response into the provided interface
func (c *VeeamClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
	resp, err := c.GET(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// PostJSON performs a POST request and unmarshals the response into the provided interface
func (c *VeeamClient) PostJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error {
	resp, err := c.POST(ctx, endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// PutJSON performs a PUT request and unmarshals the response into the provided interface
func (c *VeeamClient) PutJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error {
	resp, err := c.PUT(ctx, endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// DeleteJSON performs a DELETE request and returns any error
func (c *VeeamClient) DeleteJSON(ctx context.Context, endpoint string) error {
	resp, err := c.DELETE(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

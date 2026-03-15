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

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
	"github.com/patrikcze/terraform-provider-veeam/internal/utils"
)

// doRequest is the internal method that handles all authenticated HTTP requests.
// It adds Authorization bearer token, x-api-version header, and handles token refresh.
func (c *VeeamClient) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (*http.Response, error) {
	tflog.Debug(ctx, "Making API request", map[string]interface{}{"method": method, "endpoint": endpoint})

	// Refresh token if expiring soon
	if err := c.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Build full URL
	requestURL := c.BaseURL + endpoint
	if !strings.HasPrefix(endpoint, "/") {
		requestURL = c.BaseURL + "/" + endpoint
	}

	// Marshal payload
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
	}

	// Execute with retry
	return utils.RetryRequest(func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Required headers for all V13 API requests
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("x-api-version", APIVersion)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.TokenInfo.AccessToken))

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		return resp, nil
	}, 3, utils.DefaultRetryPolicy)
}

// parseErrorResponse attempts to parse a V13 API error from a response body.
// Returns an actionable error with the API error details.
func parseErrorResponse(statusCode int, body []byte) error {
	var apiErr models.APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && (apiErr.Message != "" || apiErr.ErrorCode != "") {
		return fmt.Errorf("API request failed (HTTP %d): %w", statusCode, &apiErr)
	}
	return fmt.Errorf("API request failed with HTTP %d: %s", statusCode, truncateBody(body, 200))
}

// truncateBody returns first n bytes of body as string for error messages.
func truncateBody(body []byte, maxLen int) string {
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}

// readAndClose reads the response body and closes it.
func readAndClose(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// GetJSON performs a GET request and unmarshals the JSON response into result.
func (c *VeeamClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	body, err := readAndClose(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, body)
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal GET %s response: %w", endpoint, err)
		}
	}

	return nil
}

// PostJSON performs a POST request with a JSON payload and unmarshals the response.
func (c *VeeamClient) PostJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	body, err := readAndClose(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, body)
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal POST %s response: %w", endpoint, err)
		}
	}

	return nil
}

// PutJSON performs a PUT request with a JSON payload and unmarshals the response.
func (c *VeeamClient) PutJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	body, err := readAndClose(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, body)
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal PUT %s response: %w", endpoint, err)
		}
	}

	return nil
}

// DeleteJSON performs a DELETE request and returns any error.
func (c *VeeamClient) DeleteJSON(ctx context.Context, endpoint string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	body, err := readAndClose(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, body)
	}

	return nil
}

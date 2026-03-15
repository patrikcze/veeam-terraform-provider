package client

import (
	"context"
)

// APIClient defines the interface for interacting with the Veeam V13 REST API.
// Resources and data sources depend on this interface (not the concrete VeeamClient),
// enabling dependency injection and testability with mock implementations.
type APIClient interface {
	// GetJSON performs a GET request and unmarshals the JSON response into result.
	GetJSON(ctx context.Context, endpoint string, result interface{}) error

	// PostJSON performs a POST request with a JSON payload and unmarshals the response.
	// result may be nil if no response body is expected.
	PostJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error

	// PutJSON performs a PUT request with a JSON payload and unmarshals the response.
	// result may be nil if no response body is expected.
	PutJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error

	// DeleteJSON performs a DELETE request.
	DeleteJSON(ctx context.Context, endpoint string) error

	// WaitForTask polls a session until the async operation completes.
	// sessionID is the ID returned by 202 Accepted responses.
	// Returns nil on success, error on failure or timeout.
	WaitForTask(ctx context.Context, sessionID string) error
}

// Compile-time check: VeeamClient must satisfy APIClient.
var _ APIClient = (*VeeamClient)(nil)

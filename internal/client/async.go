package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// defaultPollInterval is the time between polling attempts for async tasks.
	defaultPollInterval = 5 * time.Second

	// defaultTaskTimeout is the maximum time to wait for an async task.
	defaultTaskTimeout = 30 * time.Minute

	// sessionsEndpoint is the V13 sessions API path.
	sessionsEndpoint = "/api/v1/sessions"
)

// SessionState represents the state of a V13 async session.
type SessionState string

const (
	SessionStateWorking SessionState = "Working"
	SessionStateStopped SessionState = "Stopped"
)

// SessionResult represents the result of a completed V13 session.
type SessionResult string

const (
	SessionResultSuccess SessionResult = "Success"
	SessionResultFailed  SessionResult = "Failed"
	SessionResultWarning SessionResult = "Warning"
	SessionResultNone    SessionResult = "None"
)

// SessionModel represents a V13 session response for async task polling.
// Matches GET /api/v1/sessions/{id} response.
type SessionModel struct {
	ID     string        `json:"id"`
	Type   string        `json:"type"`
	State  SessionState  `json:"state"`
	Result SessionResult `json:"result"`
}

// WaitForTask polls GET /api/v1/sessions/{id} until the session reaches
// a terminal state (Stopped). Returns nil on Success, error on Failed/timeout.
func (c *VeeamClient) WaitForTask(ctx context.Context, sessionID string) error {
	return c.WaitForTaskWithOptions(ctx, sessionID, defaultPollInterval, defaultTaskTimeout)
}

// WaitForTaskWithOptions is like WaitForTask but with configurable poll interval and timeout.
func (c *VeeamClient) WaitForTaskWithOptions(ctx context.Context, sessionID string, pollInterval, timeout time.Duration) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is empty")
	}

	endpoint := fmt.Sprintf("%s/%s", sessionsEndpoint, sessionID)

	tflog.Debug(ctx, "Waiting for async task to complete", map[string]interface{}{
		"session_id": sessionID,
		"timeout":    timeout.String(),
	})

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		var session SessionModel
		if err := c.GetJSON(timeoutCtx, endpoint, &session); err != nil {
			return fmt.Errorf("failed to poll session %s: %w", sessionID, err)
		}

		tflog.Debug(ctx, "Task poll result", map[string]interface{}{
			"session_id": sessionID,
			"state":      string(session.State),
			"result":     string(session.Result),
		})

		switch session.State {
		case SessionStateStopped:
			switch session.Result {
			case SessionResultSuccess:
				tflog.Info(ctx, "Async task completed successfully", map[string]interface{}{
					"session_id": sessionID,
				})
				return nil
			case SessionResultWarning:
				tflog.Warn(ctx, "Async task completed with warnings", map[string]interface{}{
					"session_id": sessionID,
				})
				return nil // Treat warnings as success
			case SessionResultFailed:
				return fmt.Errorf("async task %s failed", sessionID)
			default:
				return fmt.Errorf("async task %s stopped with unexpected result: %s", sessionID, session.Result)
			}

		case SessionStateWorking:
			// Continue polling

		default:
			// Unknown state, continue polling
			tflog.Debug(ctx, "Unknown session state, continuing to poll", map[string]interface{}{
				"state": string(session.State),
			})
		}

		// Wait for next poll or context cancellation
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timed out waiting for task %s after %s", sessionID, timeout)
		case <-ticker.C:
			// Continue to next poll
		}
	}
}

// ParseSessionIDFromResponse extracts a session ID from a 202 Accepted response body.
// The V13 API returns session info in the response body for async operations.
func ParseSessionIDFromResponse(body []byte) (string, error) {
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse session ID from response: %w", err)
	}
	if result.ID == "" {
		return "", fmt.Errorf("session ID not found in response body")
	}
	return result.ID, nil
}

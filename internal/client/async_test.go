package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitForTaskWithOptions_Success(t *testing.T) {
	pollCount := int32(0)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		count := atomic.AddInt32(&pollCount, 1)
		session := SessionModel{
			ID:   "session-123",
			Type: "BackupJob",
		}

		if count < 3 {
			session.State = SessionStateWorking
			session.Result = SessionResultNone
		} else {
			session.State = SessionStateStopped
			session.Result = SessionResultSuccess
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.WaitForTaskWithOptions(ctx, "session-123", 50*time.Millisecond, 5*time.Second)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&pollCount), int32(3))
}

func TestWaitForTaskWithOptions_Failed(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		session := SessionModel{
			ID:     "session-fail",
			Type:   "BackupJob",
			State:  SessionStateStopped,
			Result: SessionResultFailed,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.WaitForTaskWithOptions(ctx, "session-fail", 50*time.Millisecond, 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestWaitForTaskWithOptions_Timeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		// Always return Working
		session := SessionModel{
			ID:     "session-slow",
			State:  SessionStateWorking,
			Result: SessionResultNone,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.WaitForTaskWithOptions(ctx, "session-slow", 50*time.Millisecond, 200*time.Millisecond)
	assert.Error(t, err)
	// The error may be our custom "timed out" message or a context deadline exceeded
	// from the HTTP client — both indicate the timeout worked correctly.
	errMsg := err.Error()
	assert.True(t, strings.Contains(errMsg, "timed out") || strings.Contains(errMsg, "context deadline exceeded"),
		"expected timeout-related error, got: %s", errMsg)
}

func TestWaitForTaskWithOptions_EmptySessionID(t *testing.T) {
	c := &VeeamClient{}
	err := c.WaitForTaskWithOptions(context.Background(), "", time.Second, time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session ID is empty")
}

func TestWaitForTaskWithOptions_Warning(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		session := SessionModel{
			ID:     "session-warn",
			State:  SessionStateStopped,
			Result: SessionResultWarning,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// Warnings should be treated as success
	err = c.WaitForTaskWithOptions(ctx, "session-warn", 50*time.Millisecond, 5*time.Second)
	assert.NoError(t, err)
}

func TestWaitForTaskWithOptions_ResultAsObject(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		// Real-world VBR response can return result as an object with nested "result".
		session := map[string]interface{}{
			"id":    "session-obj",
			"state": "Stopped",
			"result": map[string]interface{}{
				"result": "Success",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.WaitForTaskWithOptions(ctx, "session-obj", 50*time.Millisecond, 5*time.Second)
	assert.NoError(t, err)
}

// --- WaitForTask public wrapper ---

func TestWaitForTask_Wrapper(t *testing.T) {
	// WaitForTask delegates to WaitForTaskWithOptions with default intervals.
	// Use a fast server that immediately returns Stopped/Success so the test
	// does not take 5 seconds waiting for the defaultPollInterval.
	// We can't inject poll interval through WaitForTask, so we just verify
	// that calling it on an empty session ID returns the expected error (no
	// server needed — the validation fires before any HTTP call).
	c := &VeeamClient{}
	err := c.WaitForTask(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session ID is empty")
}

// --- normalizeSessionResult branches ---

func TestNormalizeSessionResult(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  SessionResult
	}{
		{"string success", "Success", SessionResultSuccess},
		{"string empty", "", SessionResultNone},
		{"string failed", "Failed", SessionResultFailed},
		{"map with nested result", map[string]interface{}{"result": "Success"}, SessionResultSuccess},
		{"map with empty nested result", map[string]interface{}{"result": ""}, SessionResultNone},
		{"map without result key", map[string]interface{}{"other": "value"}, SessionResultNone},
		{"nil value", nil, SessionResultNone},
		{"int value (unknown type)", 42, SessionResultNone},
		{"bool value (unknown type)", true, SessionResultNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSessionResult(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- WaitForTaskWithOptions unknown-state branch ---

func TestWaitForTaskWithOptions_UnknownStateThenStopped(t *testing.T) {
	pollCount := int32(0)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		count := atomic.AddInt32(&pollCount, 1)

		var session map[string]interface{}
		if count < 3 {
			// Return an unrecognized state — the code should treat this as
			// "continue polling" per the default branch in WaitForTaskWithOptions.
			session = map[string]interface{}{
				"id":     "session-unknown",
				"state":  "Pending", // not "Working" or "Stopped"
				"result": "None",
			}
		} else {
			session = map[string]interface{}{
				"id":     "session-unknown",
				"state":  "Stopped",
				"result": "Success",
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.WaitForTaskWithOptions(ctx, "session-unknown", 50*time.Millisecond, 5*time.Second)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&pollCount), int32(3))
}

// --- WaitForTaskWithOptions Stopped with unexpected result ---

func TestWaitForTaskWithOptions_StoppedUnexpectedResult(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}

		session := SessionModel{
			ID:    "session-odd",
			State: SessionStateStopped,
			// "None" is not Success/Warning/Failed — hits the default case
			Result: "None",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.WaitForTaskWithOptions(ctx, "session-odd", 50*time.Millisecond, 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected result")
}

func TestParseSessionIDFromResponse(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantID  string
		wantErr bool
	}{
		{"valid", `{"id":"abc-123"}`, "abc-123", false},
		{"empty id", `{"id":""}`, "", true},
		{"missing id", `{"name":"test"}`, "", true},
		{"invalid json", `not json`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseSessionIDFromResponse([]byte(tt.body))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}
}

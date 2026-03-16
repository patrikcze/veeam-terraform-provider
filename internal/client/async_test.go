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

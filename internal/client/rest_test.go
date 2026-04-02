package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errReader is an io.Reader that always returns an error on Read.
type errReader struct{ err error }

func (e errReader) Read(p []byte) (int, error) { return 0, e.err }
func (e errReader) Close() error               { return nil }

func TestParseErrorResponse_WithAPIError(t *testing.T) {
	body := []byte(`{"errorCode":"NotFound","message":"Repository not found","details":"No repository with id 'abc'"}`)
	err := parseErrorResponse(404, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "NotFound")
	assert.Contains(t, err.Error(), "Repository not found")
}

func TestParseErrorResponse_WithPlainText(t *testing.T) {
	body := []byte(`Internal Server Error`)
	err := parseErrorResponse(500, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "Internal Server Error")
}

func TestParseErrorResponse_EmptyBody(t *testing.T) {
	err := parseErrorResponse(403, []byte{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		maxLen int
		want   string
	}{
		{"short body", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello..."},
		{"empty", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateBody([]byte(tt.body), tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadAndClose(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"id":"test"}`)),
	}

	body, err := readAndClose(resp)
	require.NoError(t, err)
	assert.Equal(t, `{"id":"test"}`, string(body))
}

func TestReadAndClose_ReadError(t *testing.T) {
	readErr := errors.New("simulated read failure")
	resp := &http.Response{
		Body: errReader{err: readErr},
	}

	body, err := readAndClose(resp)
	require.Error(t, err)
	assert.Nil(t, body)
	assert.Contains(t, err.Error(), "failed to read response body")
}

// newAPIServer creates a TLS test server that handles token auth and delegates
// all other requests to the provided handler function.
func newAPIServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth2/token" {
			w.WriteHeader(http.StatusOK)
			w.Write(newTestTokenResponse("test-token", "test-refresh", 900))
			return
		}
		handler(w, r)
	}))
}

// --- PutJSON tests ---

func TestPutJSON_HappyPathWithResult(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "updated-123", "name": "repo"})
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	var result map[string]string
	err = c.PutJSON(ctx, "/api/v1/repositories/updated-123", map[string]string{"name": "repo"}, &result)
	require.NoError(t, err)
	assert.Equal(t, "updated-123", result["id"])
	assert.Equal(t, "repo", result["name"])
}

func TestPutJSON_NilResult(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.PutJSON(ctx, "/api/v1/repositories/abc-123", map[string]string{"name": "repo"}, nil)
	assert.NoError(t, err)
}

func TestPutJSON_4xxError(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": "InvalidInput",
			"message":   "name must not be empty",
		})
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.PutJSON(ctx, "/api/v1/repositories/abc-123", map[string]string{"name": ""}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "InvalidInput")
}

func TestPutJSON_UnmarshalError(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return invalid JSON to trigger unmarshal error
		w.Write([]byte(`not valid json`))
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	var result map[string]string
	err = c.PutJSON(ctx, "/api/v1/repositories/abc-123", map[string]string{"name": "repo"}, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal PUT")
}

// --- PostJSON with non-nil result ---

func TestPostJSON_WithNonNilResult(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "new-repo-456", "name": "backup-repo"})
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	var result map[string]string
	err = c.PostJSON(ctx, "/api/v1/repositories", map[string]string{"name": "backup-repo"}, &result)
	require.NoError(t, err)
	assert.Equal(t, "new-repo-456", result["id"])
	assert.Equal(t, "backup-repo", result["name"])
}

func TestPostJSON_UnmarshalError(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	var result map[string]string
	err = c.PostJSON(ctx, "/api/v1/repositories", map[string]string{"name": "repo"}, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal POST")
}

// --- DeleteJSON 4xx ---

func TestDeleteJSON_404Error(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": "NotFound",
			"message":   "credential not found",
		})
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.DeleteJSON(ctx, "/api/v1/credentials/missing-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "NotFound")
}

func TestDeleteJSON_500Error(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	err = c.DeleteJSON(ctx, "/api/v1/credentials/abc-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// --- GetJSON unmarshal error ---

func TestGetJSON_UnmarshalError(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	var result map[string]string
	err = c.GetJSON(ctx, "/api/v1/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal GET")
}

// --- doRequest: endpoint without leading slash ---

func TestDoRequest_EndpointWithoutLeadingSlash(t *testing.T) {
	server := newAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		// The request should succeed regardless of whether the endpoint had a leading slash
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "abc"})
	})
	defer server.Close()

	ctx := context.Background()
	c, err := NewVeeamClientWithHTTPClient(ctx, server.URL, "admin", "secret", server.Client())
	require.NoError(t, err)

	// Pass an endpoint without a leading slash — doRequest should normalize it
	var result map[string]string
	err = c.GetJSON(ctx, "api/v1/credentials", &result)
	assert.NoError(t, err)
}

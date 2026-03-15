package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

package utils

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryRequest_Success(t *testing.T) {
	callCount := 0
	
	requestFunc := func() (*http.Response, error) {
		callCount++
		return &http.Response{StatusCode: 200}, nil
	}

	policy := RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: DefaultShouldRetryFunc,
	}

	resp, err := RetryRequest(requestFunc, 3, policy)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 1, callCount) // Should succeed on first try
}

func TestRetryRequest_RetryOnNetworkError(t *testing.T) {
	callCount := 0
	
	requestFunc := func() (*http.Response, error) {
		callCount++
		if callCount < 3 {
			return nil, errors.New("network error")
		}
		return &http.Response{StatusCode: 200}, nil
	}

	policy := RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: DefaultShouldRetryFunc,
	}

	resp, err := RetryRequest(requestFunc, 3, policy)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 3, callCount) // Should retry twice before succeeding
}

func TestRetryRequest_RetryOnRetryableStatusCodes(t *testing.T) {
	callCount := 0
	
	requestFunc := func() (*http.Response, error) {
		callCount++
		if callCount < 3 {
			return &http.Response{StatusCode: 500}, nil
		}
		return &http.Response{StatusCode: 200}, nil
	}

	policy := RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: DefaultShouldRetryFunc,
	}

	resp, err := RetryRequest(requestFunc, 3, policy)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 3, callCount) // Should retry twice before succeeding
}

func TestRetryRequest_ExhaustedRetries(t *testing.T) {
	callCount := 0
	
	requestFunc := func() (*http.Response, error) {
		callCount++
		return nil, errors.New("persistent error")
	}

	policy := RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: DefaultShouldRetryFunc,
	}

	resp, err := RetryRequest(requestFunc, 3, policy)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 4, callCount) // Should try 4 times (initial + 3 retries)
	assert.Contains(t, err.Error(), "persistent error")
}

func TestRetryRequest_NonRetryableError(t *testing.T) {
	callCount := 0
	
	requestFunc := func() (*http.Response, error) {
		callCount++
		return &http.Response{StatusCode: 400}, nil // Bad Request - not retryable
	}

	policy := RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
		ShouldRetry: DefaultShouldRetryFunc,
	}

	resp, err := RetryRequest(requestFunc, 3, policy)

	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, 1, callCount) // Should not retry on non-retryable error
}

func TestDefaultShouldRetryFunc(t *testing.T) {
	tests := []struct {
		name           string
		resp           *http.Response
		err            error
		shouldRetry    bool
	}{
		{
			name:        "Network error",
			resp:        nil,
			err:         errors.New("network error"),
			shouldRetry: true,
		},
		{
			name:        "429 Too Many Requests",
			resp:        &http.Response{StatusCode: 429},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "500 Internal Server Error",
			resp:        &http.Response{StatusCode: 500},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "502 Bad Gateway",
			resp:        &http.Response{StatusCode: 502},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "503 Service Unavailable",
			resp:        &http.Response{StatusCode: 503},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "504 Gateway Timeout",
			resp:        &http.Response{StatusCode: 504},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "200 OK",
			resp:        &http.Response{StatusCode: 200},
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "400 Bad Request",
			resp:        &http.Response{StatusCode: 400},
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "404 Not Found",
			resp:        &http.Response{StatusCode: 404},
			err:         nil,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultShouldRetryFunc(tt.resp, tt.err)
			assert.Equal(t, tt.shouldRetry, result)
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	tests := []struct {
		name        string
		attempt     int
		baseDelay   time.Duration
		maxDelay    time.Duration
		multiplier  float64
		expected    time.Duration
	}{
		{
			name:       "First attempt",
			attempt:    0,
			baseDelay:  1 * time.Second,
			maxDelay:   30 * time.Second,
			multiplier: 2.0,
			expected:   1 * time.Second,
		},
		{
			name:       "Second attempt",
			attempt:    1,
			baseDelay:  1 * time.Second,
			maxDelay:   30 * time.Second,
			multiplier: 2.0,
			expected:   2 * time.Second,
		},
		{
			name:       "Third attempt",
			attempt:    2,
			baseDelay:  1 * time.Second,
			maxDelay:   30 * time.Second,
			multiplier: 2.0,
			expected:   4 * time.Second,
		},
		{
			name:       "Delay exceeds max",
			attempt:    10,
			baseDelay:  1 * time.Second,
			maxDelay:   10 * time.Second,
			multiplier: 2.0,
			expected:   10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDelay(tt.attempt, tt.baseDelay, tt.maxDelay, tt.multiplier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithRetryPolicy(t *testing.T) {
	customShouldRetry := func(resp *http.Response, err error) bool {
		return false // Never retry
	}

	policy := WithRetryPolicy(5, 2*time.Second, 60*time.Second, 3.0, customShouldRetry)

	assert.Equal(t, 5, policy.MaxRetries)
	assert.Equal(t, 2*time.Second, policy.BaseDelay)
	assert.Equal(t, 60*time.Second, policy.MaxDelay)
	assert.Equal(t, 3.0, policy.Multiplier)
	assert.False(t, policy.ShouldRetry(nil, errors.New("test")))
}

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy

	assert.Equal(t, 3, policy.MaxRetries)
	assert.Equal(t, 1*time.Second, policy.BaseDelay)
	assert.Equal(t, 30*time.Second, policy.MaxDelay)
	assert.Equal(t, 2.0, policy.Multiplier)
	assert.NotNil(t, policy.ShouldRetry)
}

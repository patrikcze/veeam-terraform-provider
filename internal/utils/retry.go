package utils

import (
	"math"
	"net/http"
	"time"
)

// RetryPolicy defines the retry policy for HTTP requests
type RetryPolicy struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	ShouldRetry func(*http.Response, error) bool
}

// DefaultRetryPolicy provides a default retry policy
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries:  3,
	BaseDelay:   1 * time.Second,
	MaxDelay:    30 * time.Second,
	Multiplier:  2.0,
	ShouldRetry: DefaultShouldRetryFunc,
}

// DefaultShouldRetryFunc determines if a request should be retried
func DefaultShouldRetryFunc(resp *http.Response, err error) bool {
	// Always retry on network errors
	if err != nil {
		return true
	}

	// Retry on specific HTTP status codes
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			return true
		}
	}

	return false
}

// RetryRequest executes a request with retry logic
func RetryRequest(requestFunc func() (*http.Response, error), maxRetries int, policy RetryPolicy) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Execute the request
		resp, err := requestFunc()

		// If successful or shouldn't retry, return
		if !policy.ShouldRetry(resp, err) {
			return resp, err
		}

		// Store the last response and error
		lastResp = resp
		lastErr = err

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			delay := calculateDelay(attempt, policy.BaseDelay, policy.MaxDelay, policy.Multiplier)
			time.Sleep(delay)
		}
	}

	// Return the last response and error after all retries are exhausted
	return lastResp, lastErr
}

// calculateDelay calculates the delay for exponential backoff
func calculateDelay(attempt int, baseDelay, maxDelay time.Duration, multiplier float64) time.Duration {
	delay := float64(baseDelay) * math.Pow(multiplier, float64(attempt))
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}
	return time.Duration(delay)
}

// WithRetryPolicy creates a new retry policy with custom settings
func WithRetryPolicy(maxRetries int, baseDelay, maxDelay time.Duration, multiplier float64, shouldRetry func(*http.Response, error) bool) RetryPolicy {
	return RetryPolicy{
		MaxRetries:  maxRetries,
		BaseDelay:   baseDelay,
		MaxDelay:    maxDelay,
		Multiplier:  multiplier,
		ShouldRetry: shouldRetry,
	}
}

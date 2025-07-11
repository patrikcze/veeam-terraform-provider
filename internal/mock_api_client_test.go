package internal_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockVeeamClient is a mock implementation of the VeeamClient for testing
type MockVeeamClient struct {
	mock.Mock
}

func (m *MockVeeamClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
	args := m.Called(ctx, endpoint, result)
	return args.Error(0)
}

func (m *MockVeeamClient) PostJSON(endpoint string, payload interface{}, result interface{}) error {
	args := m.Called(endpoint, payload, result)
	return args.Error(0)
}

func (m *MockVeeamClient) PutJSON(endpoint string, payload interface{}, result interface{}) error {
	args := m.Called(endpoint, payload, result)
	return args.Error(0)
}

func (m *MockVeeamClient) DeleteJSON(endpoint string) error {
	args := m.Called(endpoint)
	return args.Error(0)
}

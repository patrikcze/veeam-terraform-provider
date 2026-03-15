package internal_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockVeeamClient is a mock implementation of the APIClient interface for testing.
type MockVeeamClient struct {
	mock.Mock
}

func (m *MockVeeamClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
	args := m.Called(ctx, endpoint, result)
	return args.Error(0)
}

func (m *MockVeeamClient) PostJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error {
	args := m.Called(ctx, endpoint, payload, result)
	return args.Error(0)
}

func (m *MockVeeamClient) PutJSON(ctx context.Context, endpoint string, payload interface{}, result interface{}) error {
	args := m.Called(ctx, endpoint, payload, result)
	return args.Error(0)
}

func (m *MockVeeamClient) DeleteJSON(ctx context.Context, endpoint string) error {
	args := m.Called(ctx, endpoint)
	return args.Error(0)
}

func (m *MockVeeamClient) WaitForTask(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

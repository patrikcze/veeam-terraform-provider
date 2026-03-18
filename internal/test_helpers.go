package internal

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

// TestHelper provides common test utilities
type TestHelper struct {
	MockClient *MockVeeamClient
}

// NewTestHelper creates a new test helper with a mock client
func NewTestHelper() *TestHelper {
	return &TestHelper{
		MockClient: new(MockVeeamClient),
	}
}

// SetupMockResponse sets up a mock response for testing
func (h *TestHelper) SetupMockResponse(method, endpoint string, response interface{}) {
	switch method {
	case "GET":
		h.MockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
			result := args.Get(2)
			switch v := result.(type) {
			case *map[string]interface{}:
				*v = response.(map[string]interface{})
			case *[]map[string]interface{}:
				*v = response.([]map[string]interface{})
			}
		}).Return(nil)
	case "POST":
		h.MockClient.On("PostJSON", mock.Anything, endpoint, mock.Anything, mock.Anything).Return(nil)
	case "PUT":
		h.MockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, mock.Anything).Return(nil)
	case "DELETE":
		h.MockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)
	}
}

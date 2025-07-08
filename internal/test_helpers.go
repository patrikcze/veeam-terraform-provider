package internal

import (
	"github.com/stretchr/testify/mock"
)

// MockVeeamClient is a mock implementation of the VeeamClient for testing
type MockVeeamClient struct {
	mock.Mock
}

func (m *MockVeeamClient) GetJSON(endpoint string, result interface{}) error {
	args := m.Called(endpoint, result)
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
		h.MockClient.On("GetJSON", endpoint, mock.Anything).Run(func(args mock.Arguments) {
			result := args.Get(1)
			switch v := result.(type) {
			case *map[string]interface{}:
				*v = response.(map[string]interface{})
			case *[]map[string]interface{}:
				*v = response.([]map[string]interface{})
			}
		}).Return(nil)
	case "POST":
		h.MockClient.On("PostJSON", endpoint, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			if response != nil {
				result := args.Get(2).(*map[string]interface{})
				*result = response.(map[string]interface{})
			}
		}).Return(nil)
	case "PUT":
		h.MockClient.On("PutJSON", endpoint, mock.Anything, mock.Anything).Return(nil)
	case "DELETE":
		h.MockClient.On("DeleteJSON", endpoint).Return(nil)
	}
}

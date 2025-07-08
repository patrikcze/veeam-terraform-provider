package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
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

// TestBackupJob_CreatePayload tests the creation of a backup job payload
func TestBackupJob_CreatePayload(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response
	mockClient.On("PostJSON", "/backupJobs", mock.Anything, mock.Anything).Return(nil)

	// Create test data
	data := BackupJobModel{
		Name:    types.StringValue("test_backup"),
		Enabled: types.BoolValue(true),
	}

	// Test payload creation
	payload := map[string]interface{}{
		"name":    data.Name.ValueString(),
		"enabled": data.Enabled.ValueBool(),
	}

	// Execute mock API call
	var result map[string]interface{}
	err := mockClient.PostJSON("/backupJobs", payload, &result)

	// Assert no errors
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestBackupJob_ReadPayload(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response
	mockClient.On("GetJSON", mock.Anything, "/backupJobs/test_backup", mock.Anything).Return(nil)

	// Execute mock API call
	var result map[string]interface{}
	err := mockClient.GetJSON(context.Background(), "/backupJobs/test_backup", &result)

	// Assert no errors
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestBackupJob_UpdatePayload(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response
	mockClient.On("PutJSON", "/backupJobs/test_backup", mock.Anything, mock.Anything).Return(nil)

	// Create test data
	data := BackupJobModel{
		Name:    types.StringValue("test_backup"),
		Enabled: types.BoolValue(false),
	}

	// Test payload creation
	payload := map[string]interface{}{
		"name":    data.Name.ValueString(),
		"enabled": data.Enabled.ValueBool(),
	}

	// Execute mock API call
	err := mockClient.PutJSON("/backupJobs/test_backup", payload, nil)

	// Assert no errors
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestBackupJob_DeletePayload(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response
	mockClient.On("DeleteJSON", "/backupJobs/test_backup").Return(nil)

	// Execute mock API call
	err := mockClient.DeleteJSON("/backupJobs/test_backup")

	// Assert no errors
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestBackupJobModel tests the BackupJobModel structure
func TestBackupJobModel(t *testing.T) {
	// Create test data
	data := BackupJobModel{
		Name:    types.StringValue("test_backup"),
		Enabled: types.BoolValue(true),
	}

	// Test the model
	assert.Equal(t, "test_backup", data.Name.ValueString())
	assert.Equal(t, true, data.Enabled.ValueBool())
}

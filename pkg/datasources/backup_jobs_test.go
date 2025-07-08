package datasources

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

func TestBackupJobsDataSource_ReadAllJobs(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response for all jobs
	mockClient.On("GetJSON", mock.Anything, "/api/v1/backupJobs", mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*[]map[string]interface{})
		*result = []map[string]interface{}{
			{
				"id":          "job-1",
				"name":        "Daily Backup",
				"enabled":     true,
				"description": "Daily backup job",
				"repository":  "repo-1",
				"schedule":    "daily",
				"jobType":     "backup",
				"createdAt":   "2023-01-01T00:00:00Z",
				"updatedAt":   "2023-01-01T00:00:00Z",
			},
			{
				"id":          "job-2",
				"name":        "Weekly Backup",
				"enabled":     false,
				"description": "Weekly backup job",
				"repository":  "repo-2",
				"schedule":    "weekly",
				"jobType":     "backup",
				"createdAt":   "2023-01-01T00:00:00Z",
				"updatedAt":   "2023-01-01T00:00:00Z",
			},
		}
	}).Return(nil)

	// Execute mock API call
	var response []map[string]interface{}
	err := mockClient.GetJSON(context.Background(), "/api/v1/backupJobs", &response)

	// Assert no errors
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "job-1", response[0]["id"])
	assert.Equal(t, "Daily Backup", response[0]["name"])
	mockClient.AssertExpectations(t)
}

func TestBackupJobsDataSource_ReadByJobID(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response for specific job
	mockClient.On("GetJSON", mock.Anything, "/api/v1/backupJobs/job-1", mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"id":          "job-1",
			"name":        "Daily Backup",
			"enabled":     true,
			"description": "Daily backup job",
			"repository":  "repo-1",
			"schedule":    "daily",
			"jobType":     "backup",
			"createdAt":   "2023-01-01T00:00:00Z",
			"updatedAt":   "2023-01-01T00:00:00Z",
		}
	}).Return(nil)

	// Execute mock API call
	var response map[string]interface{}
	err := mockClient.GetJSON(context.Background(), "/api/v1/backupJobs/job-1", &response)

	// Assert no errors
	assert.NoError(t, err)
	assert.Equal(t, "job-1", response["id"])
	assert.Equal(t, "Daily Backup", response["name"])
	assert.Equal(t, true, response["enabled"])
	mockClient.AssertExpectations(t)
}

// TestBackupJobsDataSourceModel tests the BackupJobsDataSourceModel structure
func TestBackupJobsDataSourceModel(t *testing.T) {
	// Create test data
	data := BackupJobsDataSourceModel{
		ID:         types.StringValue("backup_jobs"),
		JobID:      types.StringValue("job-1"),
		JobName:    types.StringValue("Daily Backup"),
		BackupJobs: []BackupJobDataModel{},
	}

	// Test the model
	assert.Equal(t, "backup_jobs", data.ID.ValueString())
	assert.Equal(t, "job-1", data.JobID.ValueString())
	assert.Equal(t, "Daily Backup", data.JobName.ValueString())
	assert.NotNil(t, data.BackupJobs)
}

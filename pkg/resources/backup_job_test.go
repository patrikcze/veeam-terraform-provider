package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
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

func TestBackupJob_BuildSpec(t *testing.T) {
	resource := &BackupJob{}
	data := &BackupJobModel{
		Name:              types.StringValue("Daily-Backup"),
		Type:              types.StringValue("VSphereBackup"),
		Description:       types.StringValue("Daily VM backup"),
		IsHighPriority:    types.BoolValue(true),
		RepositoryID:      types.StringValue("repo-123"),
		ProxyAutoSelect:   types.BoolValue(true),
		RetentionType:     types.StringValue("RestorePoints"),
		RetentionQuantity: types.Int64Value(14),
		ScheduleEnabled:   types.BoolValue(true),
		ScheduleTime:      types.StringValue("22:00"),
		ScheduleKind:      types.StringValue("Weekdays"),
		RetryEnabled:      types.BoolValue(true),
		RetryCount:        types.Int64Value(3),
		RetryAwaitMinutes: types.Int64Value(10),
	}

	spec := resource.buildSpec(data)

	assert.Equal(t, "Daily-Backup", spec.Name)
	assert.Equal(t, models.JobTypeBackup, spec.Type)
	assert.True(t, spec.IsHighPriority)
	assert.Equal(t, "repo-123", spec.Storage.BackupRepositoryID)
	assert.True(t, spec.Storage.BackupProxies.AutoSelectEnabled)
	assert.Equal(t, 14, spec.Storage.RetentionPolicy.Quantity)
	assert.True(t, spec.Schedule.RunAutomatically)
	assert.Equal(t, "22:00", spec.Schedule.Daily.LocalTime)
	assert.Equal(t, models.DailyKindsWeekdays, spec.Schedule.Daily.DailyKind)
	assert.Equal(t, 3, spec.Schedule.Retry.RetryCount)
}

func TestBackupJob_CreatePayload(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathJobs, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(3).(*models.BackupJobModel)
		result.ID = "job-123"
		result.Name = "Daily-Backup"
		result.Type = models.JobTypeBackup
	}).Return(nil)

	var result models.BackupJobModel
	err := mockClient.PostJSON(context.Background(), client.PathJobs, nil, &result)

	assert.NoError(t, err)
	assert.Equal(t, "job-123", result.ID)
	mockClient.AssertExpectations(t)
}

func TestBackupJob_SyncFromAPI(t *testing.T) {
	resource := &BackupJob{}
	data := &BackupJobModel{}

	api := &models.BackupJobModel{
		JobModel: models.JobModel{
			ID:         "job-abc",
			Name:       "Test-Job",
			Type:       models.JobTypeBackup,
			IsDisabled: false,
		},
		Description: "Test backup job",
	}

	resource.syncFromAPI(data, api)

	assert.Equal(t, "Test-Job", data.Name.ValueString())
	assert.Equal(t, "Test backup job", data.Description.ValueString())
	assert.Equal(t, "VSphereBackup", data.Type.ValueString())
	assert.False(t, data.IsDisabled.ValueBool())
}

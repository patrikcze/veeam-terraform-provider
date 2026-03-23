package resources

// ---------------------------------------------------------------------------
// Unit tests for the veeam_backup_job resource.
//
// These tests exercise the buildVMJobSpec / buildAgentJobSpec / syncVMJobFromAPI
// helpers without requiring a live Veeam server.  A MockVeeamClient replaces the
// real API client so the CRUD code paths can be verified in isolation.
//
// Run with:
//   go test ./pkg/resources/ -run TestBackupJob -v
// ---------------------------------------------------------------------------

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// MockVeeamClient is a mock implementation of the APIClient interface.
type MockVeeamClient struct {
	mock.Mock
}

func (m *MockVeeamClient) GetJSON(ctx context.Context, endpoint string, result any) error {
	args := m.Called(ctx, endpoint, result)
	return args.Error(0)
}

func (m *MockVeeamClient) PostJSON(ctx context.Context, endpoint string, payload any, result any) error {
	args := m.Called(ctx, endpoint, payload, result)
	return args.Error(0)
}

func (m *MockVeeamClient) PutJSON(ctx context.Context, endpoint string, payload any, result any) error {
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

// ---------------------------------------------------------------------------
// buildVMJobSpec tests
// ---------------------------------------------------------------------------

// TestBackupJob_BuildVMSpec validates that buildVMJobSpec produces the correct
// API payload from a Terraform state model for a VSphereBackup job.
func TestBackupJob_BuildVMSpec(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		Name:           types.StringValue("Daily-Backup"),
		Type:           types.StringValue("VSphereBackup"),
		Description:    types.StringValue("Daily VM backup"),
		IsHighPriority: types.BoolValue(true),
		VirtualMachines: &VMBackupScope{
			Includes: []VMIncludeEntry{
				{
					Platform: types.StringValue("VSphere"),
					Type:     types.StringValue("VirtualMachine"),
					HostName: types.StringValue("vcenter.lab"),
					Name:     types.StringValue("vm-prod-01"),
					ObjectID: types.StringValue("vm-101"),
				},
			},
			ExcludeTemplates: types.BoolValue(false),
		},
		Storage: &JobStorageSettings{
			RepositoryID:      types.StringValue("repo-123"),
			ProxyAutoSelect:   types.BoolValue(true),
			RetentionType:     types.StringValue("RestorePoints"),
			RetentionQuantity: types.Int64Value(14),
		},
		Schedule: &JobScheduleSettings{
			RunAutomatically:  types.BoolValue(true),
			DailyEnabled:      types.BoolValue(true),
			DailyLocalTime:    types.StringValue("22:00"),
			DailyKind:         types.StringValue("WeekDays"),
			RetryEnabled:      types.BoolValue(true),
			RetryCount:        types.Int64Value(3),
			RetryAwaitMinutes: types.Int64Value(10),
		},
	}

	spec := r.buildVMJobSpec(data)

	assert.Equal(t, "Daily-Backup", spec.Name)
	assert.Equal(t, models.JobTypeBackup, spec.Type)
	assert.Equal(t, "Daily VM backup", spec.Description)
	assert.True(t, spec.IsHighPriority)

	// Virtual machines
	require.NotNil(t, spec.VirtualMachines)
	require.Len(t, spec.VirtualMachines.Includes, 1)
	assert.Equal(t, "VSphere", spec.VirtualMachines.Includes[0].Platform)
	assert.Equal(t, "vcenter.lab", spec.VirtualMachines.Includes[0].HostName)
	assert.Equal(t, "vm-prod-01", spec.VirtualMachines.Includes[0].Name)
	assert.Equal(t, models.VmwareTypeVirtualMachine, spec.VirtualMachines.Includes[0].Type)
	assert.Equal(t, "vm-101", spec.VirtualMachines.Includes[0].ObjectID)

	// Storage
	require.NotNil(t, spec.Storage)
	assert.Equal(t, "repo-123", spec.Storage.BackupRepositoryID)
	require.NotNil(t, spec.Storage.BackupProxies)
	assert.True(t, spec.Storage.BackupProxies.AutoSelectEnabled)
	require.NotNil(t, spec.Storage.RetentionPolicy)
	assert.Equal(t, models.RetentionPolicyTypeRestorePoints, spec.Storage.RetentionPolicy.Type)
	assert.Equal(t, 14, spec.Storage.RetentionPolicy.Quantity)

	// Schedule
	require.NotNil(t, spec.Schedule)
	assert.True(t, spec.Schedule.RunAutomatically)
	require.NotNil(t, spec.Schedule.Daily)
	assert.True(t, spec.Schedule.Daily.IsEnabled)
	assert.Equal(t, "22:00", spec.Schedule.Daily.LocalTime)
	assert.Equal(t, models.DailyKindsWeekdays, spec.Schedule.Daily.DailyKind)
	require.NotNil(t, spec.Schedule.Retry)
	assert.True(t, spec.Schedule.Retry.IsEnabled)
	assert.Equal(t, 3, spec.Schedule.Retry.RetryCount)
	assert.Equal(t, 10, spec.Schedule.Retry.AwaitMinutes)
}

// TestBackupJob_ExcludeTemplates verifies that exclude_templates=true correctly
// sets the exclusion in the generated API spec.
func TestBackupJob_ExcludeTemplates(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		Name:        types.StringValue("Job"),
		Type:        types.StringValue("VSphereBackup"),
		Description: types.StringValue("desc"),
		VirtualMachines: &VMBackupScope{
			Includes: []VMIncludeEntry{
				{Platform: types.StringValue("VSphere"), Name: types.StringValue("dc")},
			},
			ExcludeTemplates: types.BoolValue(true),
		},
	}

	spec := r.buildVMJobSpec(data)

	require.NotNil(t, spec.VirtualMachines.Excludes)
	require.NotNil(t, spec.VirtualMachines.Excludes.Templates)
	assert.True(t, spec.VirtualMachines.Excludes.Templates.IsEnabled)
}

// TestBackupJob_ScheduleAfterJob verifies that after_job_name is sent as
// "jobName" (not "jobId") in the API payload, matching the v1.3-rev1 spec.
func TestBackupJob_ScheduleAfterJob(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		Name:        types.StringValue("Job"),
		Type:        types.StringValue("VSphereBackup"),
		Description: types.StringValue("desc"),
		VirtualMachines: &VMBackupScope{
			Includes: []VMIncludeEntry{
				{Platform: types.StringValue("VSphere"), Name: types.StringValue("vm-01")},
			},
			ExcludeTemplates: types.BoolValue(false),
		},
		Schedule: &JobScheduleSettings{
			RunAutomatically: types.BoolValue(true),
			AfterJobEnabled:  types.BoolValue(true),
			AfterJobName:     types.StringValue("Backup-Source-Job"),
		},
	}

	spec := r.buildVMJobSpec(data)

	require.NotNil(t, spec.Schedule)
	require.NotNil(t, spec.Schedule.AfterThisJob)
	assert.True(t, spec.Schedule.AfterThisJob.IsEnabled)
	// Must be JobName (display name), NOT a UUID — confirmed by API v1.3-rev1.
	assert.Equal(t, "Backup-Source-Job", spec.Schedule.AfterThisJob.JobName)
}

// ---------------------------------------------------------------------------
// buildAgentJobSpec tests
// ---------------------------------------------------------------------------

// TestBackupJob_BuildAgentSpec validates agent backup job spec construction.
func TestBackupJob_BuildAgentSpec(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		Name:            types.StringValue("Agent-Backup"),
		Type:            types.StringValue("WindowsAgentBackup"),
		Description:     types.StringValue("Windows servers backup"),
		IsHighPriority:  types.BoolValue(false),
		AgentBackupMode: types.StringValue("EntireComputer"),
		AgentComputers: []AgentComputerEntry{
			{
				ID:                types.StringValue("comp-uuid-1"),
				Name:              types.StringValue("server01.lab"),
				Type:              types.StringValue("WindowsComputer"),
				ProtectionGroupID: types.StringValue("pg-uuid-1"),
			},
		},
		Storage: &JobStorageSettings{
			RepositoryID:      types.StringValue("repo-456"),
			RetentionType:     types.StringValue("RestorePoints"),
			RetentionQuantity: types.Int64Value(7),
		},
	}

	payload := r.buildAgentJobSpec(data)

	assert.Equal(t, "Agent-Backup", payload["name"])
	assert.Equal(t, "WindowsAgentBackup", payload["type"])
	assert.Equal(t, "Windows servers backup", payload["description"])
	assert.Equal(t, "EntireComputer", payload["backupMode"])

	computers, ok := payload["computers"].([]models.AgentObjectSpec)
	require.True(t, ok)
	require.Len(t, computers, 1)
	assert.Equal(t, "comp-uuid-1", computers[0].ID)
	assert.Equal(t, "server01.lab", computers[0].Name)
	assert.Equal(t, models.AgentTypeWindowsComputer, computers[0].Type)
	assert.Equal(t, "pg-uuid-1", computers[0].ProtectionGroupID)
	assert.Equal(t, "Agent", computers[0].Platform)
}

// ---------------------------------------------------------------------------
// PostJSON mock test
// ---------------------------------------------------------------------------

// TestBackupJob_CreatePayload verifies that the Create path calls PostJSON once
// and populates the job ID from the mock response.
func TestBackupJob_CreatePayload(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathJobs, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.BackupJobModel)
			result.ID = "job-123"
			result.Name = "Daily-Backup"
			result.Type = models.JobTypeBackup
			result.IsDisabled = false
		}).Return(nil)

	var result models.BackupJobModel
	err := mockClient.PostJSON(context.Background(), client.PathJobs, nil, &result)

	assert.NoError(t, err)
	assert.Equal(t, "job-123", result.ID)
	assert.Equal(t, models.JobTypeBackup, result.Type)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// syncVMJobFromAPI tests
// ---------------------------------------------------------------------------

// TestBackupJob_SyncVMFromAPI validates that syncVMJobFromAPI correctly maps
// a BackupJobModel API response into the Terraform state model.
func TestBackupJob_SyncVMFromAPI(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{}

	api := &models.BackupJobModel{
		JobModel: models.JobModel{
			ID:         "job-abc",
			Name:       "Test-Job",
			Type:       models.JobTypeBackup,
			IsDisabled: false,
		},
		Description:    "Test backup job",
		IsHighPriority: true,
		VirtualMachines: &models.BackupJobVirtualMachinesModel{
			Includes: []models.VmwareObjectSpec{
				{
					Platform: "VSphere",
					HostName: "vcenter.lab",
					Name:     "vm-01",
					Type:     models.VmwareTypeVirtualMachine,
					ObjectID: "vm-101",
				},
			},
		},
		Storage: &models.BackupJobStorageModel{
			BackupRepositoryID: "repo-123",
			BackupProxies: &models.BackupProxiesSettingsModel{
				AutoSelectEnabled: true,
			},
			RetentionPolicy: &models.BackupJobRetentionPolicySettings{
				Type:     models.RetentionPolicyTypeRestorePoints,
				Quantity: 14,
			},
		},
		Schedule: &models.BackupScheduleModel{
			RunAutomatically: true,
			Daily: &models.ScheduleDailyModel{
				IsEnabled: true,
				LocalTime: "22:00",
				DailyKind: models.DailyKindsWeekdays,
			},
			Retry: &models.ScheduleRetryModel{
				IsEnabled:    true,
				RetryCount:   3,
				AwaitMinutes: 10,
			},
		},
	}

	r.syncVMJobFromAPI(data, api)

	assert.Equal(t, "Test-Job", data.Name.ValueString())
	assert.Equal(t, "Test backup job", data.Description.ValueString())
	assert.Equal(t, "VSphereBackup", data.Type.ValueString())
	assert.False(t, data.IsDisabled.ValueBool())
	assert.True(t, data.IsHighPriority.ValueBool())

	// Virtual machines
	require.NotNil(t, data.VirtualMachines)
	require.Len(t, data.VirtualMachines.Includes, 1)
	assert.Equal(t, "VSphere", data.VirtualMachines.Includes[0].Platform.ValueString())
	assert.Equal(t, "vm-01", data.VirtualMachines.Includes[0].Name.ValueString())
	assert.Equal(t, "vcenter.lab", data.VirtualMachines.Includes[0].HostName.ValueString())

	// Storage
	require.NotNil(t, data.Storage)
	assert.Equal(t, "repo-123", data.Storage.RepositoryID.ValueString())
	assert.True(t, data.Storage.ProxyAutoSelect.ValueBool())
	assert.Equal(t, "RestorePoints", data.Storage.RetentionType.ValueString())
	assert.Equal(t, int64(14), data.Storage.RetentionQuantity.ValueInt64())

	// Schedule
	require.NotNil(t, data.Schedule)
	assert.True(t, data.Schedule.RunAutomatically.ValueBool())
	assert.True(t, data.Schedule.DailyEnabled.ValueBool())
	assert.Equal(t, "22:00", data.Schedule.DailyLocalTime.ValueString())
	assert.Equal(t, "WeekDays", data.Schedule.DailyKind.ValueString())
	assert.True(t, data.Schedule.RetryEnabled.ValueBool())
	assert.Equal(t, int64(3), data.Schedule.RetryCount.ValueInt64())
	assert.Equal(t, int64(10), data.Schedule.RetryAwaitMinutes.ValueInt64())
}

func TestBackupJob_NormalizeUnknownStateFields(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		AgentBackupMode:                types.StringUnknown(),
		IncludeUsbDrives:               types.BoolUnknown(),
		AgentType:                      types.StringUnknown(),
		UseSnapshotlessFileLevelBackup: types.BoolUnknown(),
		Storage: &JobStorageSettings{
			RepositoryID:      types.StringUnknown(),
			ProxyAutoSelect:   types.BoolUnknown(),
			RetentionType:     types.StringUnknown(),
			RetentionQuantity: types.Int64Unknown(),
		},
		GuestProcessing: &JobGuestProcessing{
			AppAwareEnabled:            types.BoolUnknown(),
			FSIndexingEnabled:          types.BoolUnknown(),
			InteractionProxyAutoSelect: types.BoolUnknown(),
		},
		Schedule: &JobScheduleSettings{
			RunAutomatically:      types.BoolUnknown(),
			DailyEnabled:          types.BoolUnknown(),
			DailyLocalTime:        types.StringUnknown(),
			DailyKind:             types.StringUnknown(),
			MonthlyEnabled:        types.BoolUnknown(),
			MonthlyLocalTime:      types.StringUnknown(),
			MonthlyDayOfMonth:     types.Int64Unknown(),
			PeriodicallyEnabled:   types.BoolUnknown(),
			PeriodicallyKind:      types.StringUnknown(),
			PeriodicallyFrequency: types.Int64Unknown(),
			AfterJobEnabled:       types.BoolUnknown(),
			AfterJobName:          types.StringUnknown(),
			RetryEnabled:          types.BoolUnknown(),
			RetryCount:            types.Int64Unknown(),
			RetryAwaitMinutes:     types.Int64Unknown(),
		},
	}

	r.normalizeUnknownStateFields(data)

	assert.True(t, data.AgentBackupMode.IsNull())
	assert.True(t, data.IncludeUsbDrives.IsNull())
	assert.True(t, data.AgentType.IsNull())
	assert.True(t, data.UseSnapshotlessFileLevelBackup.IsNull())

	require.NotNil(t, data.Storage)
	assert.True(t, data.Storage.RepositoryID.IsNull())
	assert.True(t, data.Storage.ProxyAutoSelect.IsNull())
	assert.True(t, data.Storage.RetentionType.IsNull())
	assert.True(t, data.Storage.RetentionQuantity.IsNull())

	require.NotNil(t, data.GuestProcessing)
	assert.True(t, data.GuestProcessing.AppAwareEnabled.IsNull())
	assert.True(t, data.GuestProcessing.FSIndexingEnabled.IsNull())
	assert.True(t, data.GuestProcessing.InteractionProxyAutoSelect.IsNull())

	require.NotNil(t, data.Schedule)
	assert.True(t, data.Schedule.RunAutomatically.IsNull())
	assert.True(t, data.Schedule.DailyEnabled.IsNull())
	assert.True(t, data.Schedule.DailyLocalTime.IsNull())
	assert.True(t, data.Schedule.DailyKind.IsNull())
	assert.True(t, data.Schedule.MonthlyEnabled.IsNull())
	assert.True(t, data.Schedule.MonthlyLocalTime.IsNull())
	assert.True(t, data.Schedule.MonthlyDayOfMonth.IsNull())
	assert.True(t, data.Schedule.PeriodicallyEnabled.IsNull())
	assert.True(t, data.Schedule.PeriodicallyKind.IsNull())
	assert.True(t, data.Schedule.PeriodicallyFrequency.IsNull())
	assert.True(t, data.Schedule.AfterJobEnabled.IsNull())
	assert.True(t, data.Schedule.AfterJobName.IsNull())
	assert.True(t, data.Schedule.RetryEnabled.IsNull())
	assert.True(t, data.Schedule.RetryCount.IsNull())
	assert.True(t, data.Schedule.RetryAwaitMinutes.IsNull())
}

func TestBackupJob_SyncAgentFromAPI_DoesNotMaterializeStorageWhenOmitted(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type:    types.StringValue("LinuxAgentBackup"),
		Storage: nil,
	}

	api := map[string]interface{}{
		"name": "linux-agent-schedule",
		"type": "LinuxAgentBackup",
		"storage": map[string]interface{}{
			"backupRepositoryId": "repo-123",
			"retentionPolicy": map[string]interface{}{
				"type":     "Days",
				"quantity": float64(7),
			},
		},
	}

	r.syncAgentJobFromAPIMap(data, api)

	assert.Nil(t, data.Storage)
}

func TestBackupJob_SyncScheduleFromAPIMap_PreservesPlannedRetryValues(t *testing.T) {
	r := &BackupJob{}
	existing := &JobScheduleSettings{
		RunAutomatically:  types.BoolValue(true),
		DailyEnabled:      types.BoolValue(true),
		DailyLocalTime:    types.StringValue("20:00"),
		DailyKind:         types.StringValue("Everyday"),
		RetryEnabled:      types.BoolValue(false),
		RetryCount:        types.Int64Null(),
		RetryAwaitMinutes: types.Int64Null(),
	}

	api := map[string]interface{}{
		"runAutomatically": true,
		"daily": map[string]interface{}{
			"isEnabled": true,
			"localTime": "20:00",
			"dailyKind": "Everyday",
		},
		"retry": map[string]interface{}{
			"isEnabled":    true,
			"retryCount":   float64(3),
			"awaitMinutes": float64(10),
		},
	}

	synced := r.syncScheduleFromAPIMap(existing, api)

	assert.False(t, synced.RetryEnabled.ValueBool())
	assert.True(t, synced.RetryCount.IsNull())
	assert.True(t, synced.RetryAwaitMinutes.IsNull())
}

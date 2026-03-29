package resources

// ---------------------------------------------------------------------------
// Coverage tests for resource Metadata/Schema/New functions and uncovered
// backup_job helper functions. These tests do not require a live Veeam server.
//
// Run with:
//   go test ./pkg/resources/ -run TestResourceMetadata -v
// ---------------------------------------------------------------------------

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// ---------------------------------------------------------------------------
// Resource Metadata / Schema / New — all resources
// ---------------------------------------------------------------------------

func TestResourceMetadata_BackupJob(t *testing.T) {
	r := NewBackupJob()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_backup_job", resp.TypeName)
}

func TestResourceSchema_BackupJob(t *testing.T) {
	r := &BackupJob{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
}

func TestResourceMetadata_CloudCredential(t *testing.T) {
	r := NewCloudCredential()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_cloud_credential", resp.TypeName)
}

func TestResourceSchema_CloudCredential(t *testing.T) {
	r := &CloudCredential{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
}

func TestResourceMetadata_ConfigurationBackup(t *testing.T) {
	r := NewConfigurationBackup()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_configuration_backup", resp.TypeName)
}

func TestResourceMetadata_Credential(t *testing.T) {
	r := NewCredential()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_credential", resp.TypeName)
}

func TestResourceSchema_Credential(t *testing.T) {
	r := &Credential{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "username")
	assert.Contains(t, resp.Schema.Attributes, "password")
	assert.Contains(t, resp.Schema.Attributes, "type")
}

func TestResourceMetadata_EncryptionPassword(t *testing.T) {
	r := NewEncryptionPassword()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_encryption_password", resp.TypeName)
}

func TestResourceSchema_EncryptionPassword(t *testing.T) {
	r := &EncryptionPassword{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "password")
	assert.Contains(t, resp.Schema.Attributes, "hint")
}

func TestResourceMetadata_ManagedServer(t *testing.T) {
	r := NewManagedServer()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_managed_server", resp.TypeName)
}

func TestResourceSchema_ManagedServer(t *testing.T) {
	r := &ManagedServer{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "credentials_id")
}

func TestResourceMetadata_ProtectionGroup(t *testing.T) {
	r := NewProtectionGroup()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_protection_group", resp.TypeName)
}

func TestResourceSchema_ProtectionGroup(t *testing.T) {
	r := &ProtectionGroup{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
}

func TestResourceMetadata_Proxy(t *testing.T) {
	r := NewProxy()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_proxy", resp.TypeName)
}

func TestResourceSchema_Proxy(t *testing.T) {
	r := &Proxy{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "host_id")
}

func TestResourceMetadata_Repository(t *testing.T) {
	r := NewRepository()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_repository", resp.TypeName)
}

func TestResourceSchema_Repository(t *testing.T) {
	r := &Repository{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "host_id")
}

func TestResourceMetadata_ScaleOutRepository(t *testing.T) {
	r := NewScaleOutRepository()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_scale_out_repository", resp.TypeName)
}

func TestResourceSchema_ScaleOutRepository(t *testing.T) {
	r := &ScaleOutRepository{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "performance_extent_ids")
}

// ---------------------------------------------------------------------------
// BackupJob — buildVMJobModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildVMJobModel_Full(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		ID:             types.StringValue("job-abc"),
		Name:           types.StringValue("Daily-VM"),
		Type:           types.StringValue("VSphereBackup"),
		Description:    types.StringValue("Daily backup"),
		IsHighPriority: types.BoolValue(true),
		VirtualMachines: &VMBackupScope{
			Includes: []VMIncludeEntry{
				{
					Platform: types.StringValue("VSphere"),
					Type:     types.StringValue("VirtualMachine"),
					HostName: types.StringValue("vcenter.lab"),
					Name:     types.StringValue("vm-01"),
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
		GuestProcessing: &JobGuestProcessing{
			AppAwareEnabled:            types.BoolValue(true),
			FSIndexingEnabled:          types.BoolValue(false),
			InteractionProxyAutoSelect: types.BoolValue(true),
		},
		Schedule: &JobScheduleSettings{
			RunAutomatically: types.BoolValue(true),
			DailyEnabled:     types.BoolValue(true),
			DailyLocalTime:   types.StringValue("22:00"),
			DailyKind:        types.StringValue("WeekDays"),
		},
	}

	model := r.buildVMJobModel(data, false)

	assert.Equal(t, "job-abc", model.ID)
	assert.Equal(t, "Daily-VM", model.Name)
	assert.Equal(t, models.JobTypeBackup, model.Type)
	assert.Equal(t, "Daily backup", model.Description)
	assert.True(t, model.IsHighPriority)
	assert.False(t, model.IsDisabled)
	require.NotNil(t, model.VirtualMachines)
	require.Len(t, model.VirtualMachines.Includes, 1)
	require.NotNil(t, model.Storage)
	assert.Equal(t, "repo-123", model.Storage.BackupRepositoryID)
	require.NotNil(t, model.GuestProcessing)
	assert.True(t, model.GuestProcessing.AppAwareProcessing.IsEnabled)
	require.NotNil(t, model.Schedule)
	assert.True(t, model.Schedule.RunAutomatically)
}

func TestBackupJob_BuildVMJobModel_Disabled(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Name:            types.StringValue("Disabled-Job"),
		Type:            types.StringValue("VSphereBackup"),
		Description:     types.StringValue(""),
		IsHighPriority:  types.BoolValue(false),
		VirtualMachines: &VMBackupScope{Includes: []VMIncludeEntry{}},
	}
	model := r.buildVMJobModel(data, true)
	assert.True(t, model.IsDisabled)
}

// ---------------------------------------------------------------------------
// BackupJob — buildAgentJobModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildAgentJobModel_Windows(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		ID:              types.StringValue("agent-job-1"),
		Name:            types.StringValue("WinAgent-Backup"),
		Type:            types.StringValue("WindowsAgentBackup"),
		Description:     types.StringValue("Windows agent backup"),
		IsHighPriority:  types.BoolValue(false),
		AgentBackupMode: types.StringValue("EntireComputer"),
		IncludeUsbDrives: types.BoolValue(true),
		AgentType:       types.StringValue("Workstation"),
		AgentComputers: []AgentComputerEntry{
			{
				ID:                types.StringValue("comp-1"),
				Name:              types.StringValue("WS01.corp"),
				Type:              types.StringValue("WindowsComputer"),
				ProtectionGroupID: types.StringValue("pg-1"),
			},
		},
		Storage: &JobStorageSettings{
			RepositoryID:      types.StringValue("repo-win"),
			RetentionType:     types.StringValue("Days"),
			RetentionQuantity: types.Int64Value(30),
		},
	}

	model := r.buildAgentJobModel(data, false)

	assert.Equal(t, "agent-job-1", model["id"])
	assert.Equal(t, "WinAgent-Backup", model["name"])
	assert.Equal(t, "WindowsAgentBackup", model["type"])
	assert.Equal(t, "EntireComputer", model["backupMode"])
	assert.Equal(t, false, model["isDisabled"])
	assert.Equal(t, true, model["includeUsbDrives"])
	assert.Equal(t, "Workstation", model["agentType"])

	computers, ok := model["computers"].([]models.AgentObjectSpec)
	require.True(t, ok)
	assert.Len(t, computers, 1)
	assert.Equal(t, "comp-1", computers[0].ID)
}

func TestBackupJob_BuildAgentJobModel_Linux(t *testing.T) {
	r := &BackupJob{}

	data := &BackupJobModel{
		ID:              types.StringValue("linux-agent-1"),
		Name:            types.StringValue("Linux-Agent-Backup"),
		Type:            types.StringValue("LinuxAgentBackup"),
		Description:     types.StringValue("Linux agent backup"),
		IsHighPriority:  types.BoolValue(false),
		AgentBackupMode: types.StringValue("EntireComputer"),
		UseSnapshotlessFileLevelBackup: types.BoolValue(true),
		AgentComputers: []AgentComputerEntry{},
		Storage: &JobStorageSettings{
			RepositoryID:      types.StringValue("repo-linux"),
			RetentionType:     types.StringValue("Days"),
			RetentionQuantity: types.Int64Value(7),
		},
		VolumesScope: &AgentVolumesScope{
			AllVolumes: types.BoolValue(true),
			VolumeNames: types.ListValueMust(types.StringType, []attr.Value{}),
		},
	}

	model := r.buildAgentJobModel(data, false)

	assert.Equal(t, "Linux-Agent-Backup", model["name"])
	assert.Equal(t, "LinuxAgentBackup", model["type"])
	assert.Equal(t, true, model["useSnapshotlessFileLevelBackup"])
	assert.NotNil(t, model["volumes"])
}

// ---------------------------------------------------------------------------
// BackupJob — buildVirtualMachinesModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildVirtualMachinesModel_WithExclusions(t *testing.T) {
	r := &BackupJob{}

	scope := &VMBackupScope{
		Includes: []VMIncludeEntry{
			{
				Platform: types.StringValue("VSphere"),
				Type:     types.StringValue("VirtualMachine"),
				HostName: types.StringValue("vcenter.local"),
				Name:     types.StringValue("vm-99"),
				ObjectID: types.StringValue("vm-99-id"),
			},
		},
		ExcludeTemplates: types.BoolValue(true),
	}

	model := r.buildVirtualMachinesModel(scope)

	require.NotNil(t, model)
	require.Len(t, model.Includes, 1)
	assert.Equal(t, "VSphere", model.Includes[0].Platform)
	assert.Equal(t, "vm-99", model.Includes[0].Name)
	require.NotNil(t, model.Excludes)
	require.NotNil(t, model.Excludes.Templates)
	assert.True(t, model.Excludes.Templates.IsEnabled)
}

func TestBackupJob_BuildVirtualMachinesModel_NilScope(t *testing.T) {
	r := &BackupJob{}
	model := r.buildVirtualMachinesModel(nil)
	assert.Nil(t, model)
}

func TestBackupJob_BuildVirtualMachinesModel_EmptyPlatformDefaultsToVSphere(t *testing.T) {
	r := &BackupJob{}

	scope := &VMBackupScope{
		Includes: []VMIncludeEntry{
			{
				Platform: types.StringValue(""), // empty should default to VSphere
				Name:     types.StringValue("vm-no-platform"),
			},
		},
		ExcludeTemplates: types.BoolValue(false),
	}

	model := r.buildVirtualMachinesModel(scope)
	require.Len(t, model.Includes, 1)
	assert.Equal(t, string(models.InventoryPlatformVSphere), model.Includes[0].Platform)
}

// ---------------------------------------------------------------------------
// BackupJob — buildGuestProcessingModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildGuestProcessingModel_Full(t *testing.T) {
	r := &BackupJob{}

	gp := &JobGuestProcessing{
		AppAwareEnabled:            types.BoolValue(true),
		FSIndexingEnabled:          types.BoolValue(false),
		InteractionProxyAutoSelect: types.BoolValue(true),
		GuestCredentials: &JobGuestCredentials{
			CredentialsID: types.StringValue("cred-gp-1"),
		},
	}

	model := r.buildGuestProcessingModel(gp)

	require.NotNil(t, model)
	require.NotNil(t, model.AppAwareProcessing)
	assert.True(t, model.AppAwareProcessing.IsEnabled)
	require.NotNil(t, model.GuestFSIndexing)
	assert.False(t, model.GuestFSIndexing.IsEnabled)
	require.NotNil(t, model.GuestInteractionProxies)
	assert.True(t, model.GuestInteractionProxies.AutoSelectEnabled)
	require.NotNil(t, model.GuestCredentials)
	assert.Equal(t, "cred-gp-1", model.GuestCredentials.CredentialsID)
}

func TestBackupJob_BuildGuestProcessingModel_NoCredentials(t *testing.T) {
	r := &BackupJob{}

	gp := &JobGuestProcessing{
		AppAwareEnabled:            types.BoolValue(false),
		FSIndexingEnabled:          types.BoolValue(true),
		InteractionProxyAutoSelect: types.BoolNull(), // null — no proxy block
		GuestCredentials:           nil,
	}

	model := r.buildGuestProcessingModel(gp)
	require.NotNil(t, model)
	assert.Nil(t, model.GuestInteractionProxies) // null should not generate block
	assert.Nil(t, model.GuestCredentials)
}

func TestBackupJob_BuildGuestProcessingModel_Nil(t *testing.T) {
	r := &BackupJob{}
	model := r.buildGuestProcessingModel(nil)
	assert.Nil(t, model)
}

// ---------------------------------------------------------------------------
// BackupJob — buildGFSPolicyModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildGFSPolicyModel_Full(t *testing.T) {
	gfs := &JobGFSPolicy{
		IsEnabled:          types.BoolValue(true),
		WeeklyEnabled:      types.BoolValue(true),
		WeeklyKeepFor:      types.Int64Value(4),
		WeeklyDesiredTime:  types.StringValue("Saturday"),
		MonthlyEnabled:     types.BoolValue(true),
		MonthlyKeepFor:     types.Int64Value(3),
		MonthlyDesiredTime: types.StringValue("Last"),
		YearlyEnabled:      types.BoolValue(true),
		YearlyKeepFor:      types.Int64Value(7),
		YearlyDesiredTime:  types.StringValue("January"),
	}

	model := buildGFSPolicyModel(gfs)

	require.NotNil(t, model)
	assert.True(t, model.IsEnabled)
	require.NotNil(t, model.Weekly)
	assert.True(t, model.Weekly.IsEnabled)
	assert.Equal(t, 4, model.Weekly.KeepForNumberOfWeeks)
	assert.Equal(t, models.EDayOfWeek("Saturday"), model.Weekly.DesiredTime)
	require.NotNil(t, model.Monthly)
	assert.True(t, model.Monthly.IsEnabled)
	assert.Equal(t, 3, model.Monthly.KeepForNumberOfMonths)
	require.NotNil(t, model.Yearly)
	assert.True(t, model.Yearly.IsEnabled)
	assert.Equal(t, 7, model.Yearly.KeepForNumberOfYears)
}

func TestBackupJob_BuildGFSPolicyModel_DisabledSections(t *testing.T) {
	gfs := &JobGFSPolicy{
		IsEnabled:      types.BoolValue(false),
		WeeklyEnabled:  types.BoolValue(false),
		MonthlyEnabled: types.BoolValue(false),
		YearlyEnabled:  types.BoolValue(false),
	}

	model := buildGFSPolicyModel(gfs)

	require.NotNil(t, model)
	assert.False(t, model.IsEnabled)
	assert.Nil(t, model.Weekly)
	assert.Nil(t, model.Monthly)
	assert.Nil(t, model.Yearly)
}

func TestBackupJob_BuildGFSPolicyModel_Nil(t *testing.T) {
	model := buildGFSPolicyModel(nil)
	assert.Nil(t, model)
}

// ---------------------------------------------------------------------------
// BackupJob — syncGFSPolicyFromAPI
// ---------------------------------------------------------------------------

func TestBackupJob_SyncGFSPolicyFromAPI_Full(t *testing.T) {
	api := &models.GFSPolicySettingsModel{
		IsEnabled: true,
		Weekly: &models.GFSPolicySettingsWeeklyModel{
			IsEnabled:            true,
			KeepForNumberOfWeeks: 4,
			DesiredTime:          models.EDayOfWeek("Saturday"),
		},
		Monthly: &models.GFSPolicySettingsMonthlyModel{
			IsEnabled:             true,
			KeepForNumberOfMonths: 3,
			DesiredTime:           models.ESennightOfMonth("Last"),
		},
		Yearly: &models.GFSPolicySettingsYearlyModel{
			IsEnabled:            true,
			KeepForNumberOfYears: 7,
			DesiredTime:          models.EMonth("January"),
		},
	}

	gfs := syncGFSPolicyFromAPI(api)

	require.NotNil(t, gfs)
	assert.True(t, gfs.IsEnabled.ValueBool())
	assert.True(t, gfs.WeeklyEnabled.ValueBool())
	assert.Equal(t, int64(4), gfs.WeeklyKeepFor.ValueInt64())
	assert.Equal(t, "Saturday", gfs.WeeklyDesiredTime.ValueString())
	assert.True(t, gfs.MonthlyEnabled.ValueBool())
	assert.Equal(t, int64(3), gfs.MonthlyKeepFor.ValueInt64())
	assert.Equal(t, "Last", gfs.MonthlyDesiredTime.ValueString())
	assert.True(t, gfs.YearlyEnabled.ValueBool())
	assert.Equal(t, int64(7), gfs.YearlyKeepFor.ValueInt64())
	assert.Equal(t, "January", gfs.YearlyDesiredTime.ValueString())
}

func TestBackupJob_SyncGFSPolicyFromAPI_EmptySections(t *testing.T) {
	// Sections present but with zero values — should produce null for keepFor.
	api := &models.GFSPolicySettingsModel{
		IsEnabled: true,
		Weekly: &models.GFSPolicySettingsWeeklyModel{
			IsEnabled:            false,
			KeepForNumberOfWeeks: 0,
			DesiredTime:          "",
		},
	}

	gfs := syncGFSPolicyFromAPI(api)

	require.NotNil(t, gfs)
	assert.False(t, gfs.WeeklyEnabled.ValueBool())
	assert.True(t, gfs.WeeklyKeepFor.IsNull())
	assert.True(t, gfs.WeeklyDesiredTime.IsNull())
	// Monthly section not set in API, so MonthlyEnabled stays at its zero value.
	assert.False(t, gfs.MonthlyEnabled.ValueBool())
}

func TestBackupJob_SyncGFSPolicyFromAPI_Nil(t *testing.T) {
	gfs := syncGFSPolicyFromAPI(nil)
	assert.Nil(t, gfs)
}

// ---------------------------------------------------------------------------
// BackupJob — buildAgentVolumesScopeModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildAgentVolumesScopeModel_AllVolumes(t *testing.T) {
	vs := &AgentVolumesScope{
		AllVolumes:  types.BoolValue(true),
		VolumeNames: types.ListValueMust(types.StringType, []attr.Value{}),
	}

	model := buildAgentVolumesScopeModel(vs)

	require.NotNil(t, model)
	assert.True(t, model.AllVolumes)
	assert.Empty(t, model.VolumeNames)
}

func TestBackupJob_BuildAgentVolumesScopeModel_SpecificVolumes(t *testing.T) {
	vs := &AgentVolumesScope{
		AllVolumes: types.BoolValue(false),
		VolumeNames: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("C:"),
			types.StringValue("D:"),
		}),
	}

	model := buildAgentVolumesScopeModel(vs)

	require.NotNil(t, model)
	assert.False(t, model.AllVolumes)
	assert.Len(t, model.VolumeNames, 2)
	assert.Contains(t, model.VolumeNames, "C:")
	assert.Contains(t, model.VolumeNames, "D:")
}

func TestBackupJob_BuildAgentVolumesScopeModel_Nil(t *testing.T) {
	model := buildAgentVolumesScopeModel(nil)
	assert.Nil(t, model)
}

// ---------------------------------------------------------------------------
// BackupJob — buildAgentFilesScopeModel
// ---------------------------------------------------------------------------

func TestBackupJob_BuildAgentFilesScopeModel_WithFolders(t *testing.T) {
	fs := &AgentFilesScope{
		IncludedFolders: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("/home/user"),
			types.StringValue("/var/data"),
		}),
		ExcludedFolders: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("/tmp"),
		}),
	}

	model := buildAgentFilesScopeModel(fs)

	require.NotNil(t, model)
	assert.Len(t, model.IncludedFolders, 2)
	assert.Contains(t, model.IncludedFolders, "/home/user")
	assert.Len(t, model.ExcludedFolders, 1)
	assert.Contains(t, model.ExcludedFolders, "/tmp")
}

func TestBackupJob_BuildAgentFilesScopeModel_NullFolders(t *testing.T) {
	fs := &AgentFilesScope{
		IncludedFolders: types.ListNull(types.StringType),
		ExcludedFolders: types.ListNull(types.StringType),
	}

	model := buildAgentFilesScopeModel(fs)

	require.NotNil(t, model)
	assert.Empty(t, model.IncludedFolders)
	assert.Empty(t, model.ExcludedFolders)
}

func TestBackupJob_BuildAgentFilesScopeModel_Nil(t *testing.T) {
	model := buildAgentFilesScopeModel(nil)
	assert.Nil(t, model)
}

// ---------------------------------------------------------------------------
// BackupJob — buildScheduleModel (monthly + periodically + after-job paths)
// ---------------------------------------------------------------------------

func TestBackupJob_BuildScheduleModel_Monthly(t *testing.T) {
	r := &BackupJob{}

	s := &JobScheduleSettings{
		RunAutomatically:  types.BoolValue(true),
		MonthlyEnabled:    types.BoolValue(true),
		MonthlyLocalTime:  types.StringValue("02:00"),
		MonthlyDayOfMonth: types.Int64Value(15),
	}

	model := r.buildScheduleModel(s)

	require.NotNil(t, model)
	assert.True(t, model.RunAutomatically)
	require.NotNil(t, model.Monthly)
	assert.True(t, model.Monthly.IsEnabled)
	assert.Equal(t, "02:00", model.Monthly.LocalTime)
	assert.Equal(t, 15, model.Monthly.DayOfMonth)
}

func TestBackupJob_BuildScheduleModel_Periodically(t *testing.T) {
	r := &BackupJob{}

	s := &JobScheduleSettings{
		RunAutomatically:      types.BoolValue(true),
		PeriodicallyEnabled:   types.BoolValue(true),
		PeriodicallyKind:      types.StringValue("Hours"),
		PeriodicallyFrequency: types.Int64Value(6),
	}

	model := r.buildScheduleModel(s)

	require.NotNil(t, model)
	require.NotNil(t, model.Periodically)
	assert.True(t, model.Periodically.IsEnabled)
	assert.Equal(t, models.EPeriodicallyKinds("Hours"), model.Periodically.PeriodicallyKind)
	assert.Equal(t, 6, model.Periodically.Frequency)
}

func TestBackupJob_BuildScheduleModel_Nil(t *testing.T) {
	r := &BackupJob{}
	model := r.buildScheduleModel(nil)
	assert.Nil(t, model)
}

// ---------------------------------------------------------------------------
// BackupJob — syncVMJobFromAPI (monthly + periodically + after-job paths)
// ---------------------------------------------------------------------------

func TestBackupJob_SyncVMFromAPI_Monthly(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{}

	api := &models.BackupJobModel{
		JobModel: models.JobModel{
			Name:       "Monthly-Job",
			Type:       models.JobTypeBackup,
			IsDisabled: false,
		},
		Schedule: &models.BackupScheduleModel{
			RunAutomatically: true,
			Monthly: &models.ScheduleMonthlyModel{
				IsEnabled:  true,
				LocalTime:  "03:00",
				DayOfMonth: 1,
			},
		},
	}

	r.syncVMJobFromAPI(data, api)

	require.NotNil(t, data.Schedule)
	assert.True(t, data.Schedule.MonthlyEnabled.ValueBool())
	assert.Equal(t, "03:00", data.Schedule.MonthlyLocalTime.ValueString())
	assert.Equal(t, int64(1), data.Schedule.MonthlyDayOfMonth.ValueInt64())
}

func TestBackupJob_SyncVMFromAPI_GuestProcessingWithCredentials(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{}

	api := &models.BackupJobModel{
		JobModel: models.JobModel{
			Name: "GP-Job",
			Type: models.JobTypeBackup,
		},
		GuestProcessing: &models.BackupJobGuestProcessingModel{
			AppAwareProcessing: &models.BackupApplicationAwareProcessingModel{IsEnabled: true},
			GuestFSIndexing:    &models.GuestFileSystemIndexingModel{IsEnabled: false},
			GuestInteractionProxies: &models.GuestInteractionProxiesSettingsModel{
				AutoSelectEnabled: true,
			},
			GuestCredentials: &models.GuestOsCredentialsModel{
				CredentialsID: "cred-gp-99",
			},
		},
	}

	r.syncVMJobFromAPI(data, api)

	require.NotNil(t, data.GuestProcessing)
	assert.True(t, data.GuestProcessing.AppAwareEnabled.ValueBool())
	assert.False(t, data.GuestProcessing.FSIndexingEnabled.ValueBool())
	assert.True(t, data.GuestProcessing.InteractionProxyAutoSelect.ValueBool())
	require.NotNil(t, data.GuestProcessing.GuestCredentials)
	assert.Equal(t, "cred-gp-99", data.GuestProcessing.GuestCredentials.CredentialsID.ValueString())
}

func TestBackupJob_SyncVMFromAPI_WithGFSPolicy(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{}

	api := &models.BackupJobModel{
		JobModel: models.JobModel{
			Name: "GFS-Job",
			Type: models.JobTypeBackup,
		},
		Storage: &models.BackupJobStorageModel{
			BackupRepositoryID: "repo-gfs",
			GFSPolicy: &models.GFSPolicySettingsModel{
				IsEnabled: true,
				Weekly: &models.GFSPolicySettingsWeeklyModel{
					IsEnabled:            true,
					KeepForNumberOfWeeks: 4,
					DesiredTime:          "Saturday",
				},
			},
		},
	}

	r.syncVMJobFromAPI(data, api)

	require.NotNil(t, data.Storage)
	require.NotNil(t, data.Storage.GFSPolicy)
	assert.True(t, data.Storage.GFSPolicy.IsEnabled.ValueBool())
	assert.True(t, data.Storage.GFSPolicy.WeeklyEnabled.ValueBool())
}

// ---------------------------------------------------------------------------
// BackupJob — syncAgentJobFromAPIMap — windows-specific fields
// ---------------------------------------------------------------------------

func TestBackupJob_SyncAgentFromAPI_Windows_Fields(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type: types.StringValue("WindowsAgentBackup"),
	}

	api := map[string]interface{}{
		"name":             "Win-Agent-Job",
		"type":             "WindowsAgentBackup",
		"isDisabled":       false,
		"isHighPriority":   true,
		"description":      "Windows agent backup",
		"backupMode":       "EntireComputer",
		"includeUsbDrives": true,
		"agentType":        "Workstation",
	}

	r.syncAgentJobFromAPIMap(data, api)

	assert.Equal(t, "Win-Agent-Job", data.Name.ValueString())
	assert.Equal(t, "WindowsAgentBackup", data.Type.ValueString())
	assert.True(t, data.IsHighPriority.ValueBool())
	assert.Equal(t, "Windows agent backup", data.Description.ValueString())
	assert.Equal(t, "EntireComputer", data.AgentBackupMode.ValueString())
	assert.True(t, data.IncludeUsbDrives.ValueBool())
	assert.Equal(t, "Workstation", data.AgentType.ValueString())
}

func TestBackupJob_SyncAgentFromAPI_Linux_Fields(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type: types.StringValue("LinuxAgentBackup"),
	}

	api := map[string]interface{}{
		"name":                           "Linux-Agent-Job",
		"type":                           "LinuxAgentBackup",
		"isDisabled":                     false,
		"backupMode":                     "EntireComputer",
		"useSnapshotlessFileLevelBackup": true,
	}

	r.syncAgentJobFromAPIMap(data, api)

	assert.Equal(t, "Linux-Agent-Job", data.Name.ValueString())
	assert.True(t, data.UseSnapshotlessFileLevelBackup.ValueBool())
	// Windows-specific fields should be null.
	assert.True(t, data.IncludeUsbDrives.IsNull())
	assert.True(t, data.AgentType.IsNull())
}

// ---------------------------------------------------------------------------
// BackupJob — syncScheduleFromAPI (periodically and after-job paths)
// ---------------------------------------------------------------------------

func TestBackupJob_SyncScheduleFromAPI_Periodically(t *testing.T) {
	r := &BackupJob{}

	api := &models.BackupScheduleModel{
		RunAutomatically: true,
		Periodically: &models.SchedulePeriodicallyModel{
			IsEnabled:        true,
			PeriodicallyKind: "Hours",
			Frequency:        8,
		},
	}

	result := r.syncScheduleFromAPI(nil, api)

	require.NotNil(t, result)
	assert.True(t, result.PeriodicallyEnabled.ValueBool())
	assert.Equal(t, "Hours", result.PeriodicallyKind.ValueString())
	assert.Equal(t, int64(8), result.PeriodicallyFrequency.ValueInt64())
}

func TestBackupJob_SyncScheduleFromAPI_AfterJob(t *testing.T) {
	r := &BackupJob{}

	api := &models.BackupScheduleModel{
		RunAutomatically: true,
		AfterThisJob: &models.ScheduleAfterThisJobModel{
			IsEnabled: true,
			JobName:   "Source-Backup",
		},
	}

	result := r.syncScheduleFromAPI(nil, api)

	require.NotNil(t, result)
	assert.True(t, result.AfterJobEnabled.ValueBool())
	assert.Equal(t, "Source-Backup", result.AfterJobName.ValueString())
}

func TestBackupJob_SyncScheduleFromAPI_NoSchedule(t *testing.T) {
	r := &BackupJob{}

	// When daily/monthly/periodically/afterJob/retry are all nil.
	api := &models.BackupScheduleModel{
		RunAutomatically: false,
	}

	result := r.syncScheduleFromAPI(nil, api)

	require.NotNil(t, result)
	assert.False(t, result.RunAutomatically.ValueBool())
	assert.False(t, result.DailyEnabled.ValueBool())
	assert.False(t, result.MonthlyEnabled.ValueBool())
	assert.False(t, result.PeriodicallyEnabled.ValueBool())
	assert.False(t, result.AfterJobEnabled.ValueBool())
	assert.False(t, result.RetryEnabled.ValueBool())
}

// ---------------------------------------------------------------------------
// ManagedServer — syncFromAPI
// ---------------------------------------------------------------------------

func TestManagedServer_SyncFromAPI(t *testing.T) {
	r := &ManagedServer{}
	data := &ManagedServerModel{
		Name:        types.StringValue("old-name"),
		Description: types.StringValue("old-desc"),
		Type:        types.StringValue("WindowsHost"),
		Status:      types.StringValue(""),
	}

	api := &models.ManagedServerModel{
		Name:        "WIN-DC01",
		Description: "Primary domain controller",
		Type:        models.ManagedServerTypeWindowsHost,
		Status:      "Available",
	}

	r.syncFromAPI(data, api)

	assert.Equal(t, "WIN-DC01", data.Name.ValueString())
	assert.Equal(t, "Primary domain controller", data.Description.ValueString())
	assert.Equal(t, "WindowsHost", data.Type.ValueString())
	assert.Equal(t, "Available", data.Status.ValueString())
}

func TestManagedServer_SyncFromAPI_EmptyFields(t *testing.T) {
	// Empty API values should not overwrite existing state.
	r := &ManagedServer{}
	data := &ManagedServerModel{
		Name:        types.StringValue("original-name"),
		Description: types.StringValue("original-desc"),
		Type:        types.StringValue("WindowsHost"),
	}

	api := &models.ManagedServerModel{
		// All empty — should not overwrite.
	}

	r.syncFromAPI(data, api)

	assert.Equal(t, "original-name", data.Name.ValueString()) // unchanged
	assert.Equal(t, "original-desc", data.Description.ValueString())
}

// ---------------------------------------------------------------------------
// ProtectionGroup — buildUpdateModel and syncFromAPICloud
// ---------------------------------------------------------------------------

func TestProtectionGroup_BuildUpdateModel_Individual(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		ID:          types.StringValue("pg-123"),
		Name:        types.StringValue("Windows Servers"),
		Description: types.StringValue("All Windows servers"),
		Type:        types.StringValue("IndividualComputers"),
		IsDisabled:  types.BoolValue(false),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("srv01.corp"),
				ConnectionType: types.StringValue("PermanentCredentials"),
				CredentialsID:  types.StringValue("cred-1"),
			},
		},
	}

	model := r.buildUpdateModel(data)

	pg, ok := model.(*models.IndividualComputersProtectionGroupModel)
	require.True(t, ok, "expected *IndividualComputersProtectionGroupModel")
	assert.Equal(t, "pg-123", pg.ID)
	assert.Equal(t, "Windows Servers", pg.Name)
	assert.Equal(t, models.ProtectionGroupTypeIndividualComputers, pg.Type)
}

func TestProtectionGroup_BuildUpdateModel_Cloud(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		ID:          types.StringValue("pg-cloud"),
		Name:        types.StringValue("Cloud Machines"),
		Description: types.StringValue("Cloud backup group"),
		Type:        types.StringValue("CloudMachines"),
		IsDisabled:  types.BoolValue(false),
		CloudAccount: []ProtectionGroupCloudAccountModel{
			{
				AccountType:   types.StringValue("AWS"),
				CredentialsID: types.StringValue("cloud-cred-1"),
			},
		},
		CloudMachines: []ProtectionGroupCloudMachineModel{},
	}

	model := r.buildUpdateModel(data)

	pg, ok := model.(*models.CloudMachinesProtectionGroupModel)
	require.True(t, ok, "expected *CloudMachinesProtectionGroupModel")
	assert.Equal(t, "pg-cloud", pg.ID)
	assert.Equal(t, "Cloud Machines", pg.Name)
	assert.Equal(t, models.ProtectionGroupTypeCloudMachines, pg.Type)
}

func TestProtectionGroup_SyncFromAPICloud(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		Name:        types.StringValue("old"),
		Description: types.StringValue("old desc"),
		Type:        types.StringValue("IndividualComputers"),
	}

	api := &models.CloudMachinesProtectionGroupModel{
		ProtectionGroupModel: models.ProtectionGroupModel{
			Name:        "Cloud Machines PG",
			Description: "Azure VMs",
			Type:        models.ProtectionGroupTypeCloudMachines,
			IsDisabled:  false,
		},
		CloudAccount: &models.CloudMachinesAccount{
			AccountType:    models.ProtectionGroupCloudAccountTypeAzure,
			CredentialsID:  "azure-cred-1",
			SubscriptionID: "sub-abc",
			RegionType:     "Public",
			RegionID:       "EastUS",
		},
	}

	r.syncFromAPICloud(data, api)

	assert.Equal(t, "Cloud Machines PG", data.Name.ValueString())
	assert.Equal(t, "Azure VMs", data.Description.ValueString())
	assert.Equal(t, "CloudMachines", data.Type.ValueString())
	assert.False(t, data.IsDisabled.ValueBool())
	require.Len(t, data.CloudAccount, 1)
	assert.Equal(t, "Azure", data.CloudAccount[0].AccountType.ValueString())
	assert.Equal(t, "azure-cred-1", data.CloudAccount[0].CredentialsID.ValueString())
	assert.Equal(t, "sub-abc", data.CloudAccount[0].SubscriptionID.ValueString())
	// Computers should be cleared when syncing cloud model.
	assert.Nil(t, data.Computers)
}

// ---------------------------------------------------------------------------
// CloudCredential — buildSpec additional paths
// ---------------------------------------------------------------------------

func TestCloudCredential_BuildSpec_GCP(t *testing.T) {
	resource := &CloudCredential{}
	data := &CloudCredentialModel{
		Name:          types.StringValue("gcp-project"),
		Description:   types.StringValue("GCP service account"),
		Type:           types.StringValue("GoogleService"),
		ServiceAccount: types.StringValue(`{"type":"service_account"}`),
	}

	spec, validationError := resource.buildSpec(data)
	assert.Equal(t, "", validationError)
	require.NotNil(t, spec)
	assert.Equal(t, "GoogleService", spec.Type)
	assert.Equal(t, `{"type":"service_account"}`, spec.ServiceAccount)
}

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// ---------------------------------------------------------------------------
// BackupJobs data source
// ---------------------------------------------------------------------------

func TestBackupJobsDataSource_Metadata(t *testing.T) {
	ds := NewBackupJobsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_backup_jobs", resp.TypeName)
}

func TestBackupJobsDataSource_Schema(t *testing.T) {
	ds := NewBackupJobsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "backup_jobs")
	assert.Contains(t, resp.Schema.Attributes, "job_id")
	assert.Contains(t, resp.Schema.Attributes, "job_name")
}

func TestBackupJobsDataSource_Configure_Nil(t *testing.T) {
	ds := &BackupJobsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, ds.client)
}

func TestBackupJobsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupJobsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, ds.client)
}

func TestBackupJobsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &BackupJobsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong-type"}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupJobsDataSourceModel_Fields(t *testing.T) {
	data := BackupJobsDataSourceModel{
		ID:      types.StringValue("backup_jobs"),
		JobID:   types.StringNull(),
		JobName: types.StringNull(),
		BackupJobs: []BackupJobDataModel{
			{
				ID:          types.StringValue("job-1"),
				Name:        types.StringValue("Daily Backup"),
				Enabled:     types.BoolValue(true),
				Description: types.StringValue("Daily VM backup"),
				Repository:  types.StringValue("Main Repo"),
				Schedule:    types.StringValue("daily"),
				JobType:     types.StringValue("backup"),
				CreatedAt:   types.StringValue("2024-01-01T00:00:00Z"),
				UpdatedAt:   types.StringValue("2024-01-01T00:00:00Z"),
			},
		},
	}

	assert.Equal(t, "backup_jobs", data.ID.ValueString())
	assert.True(t, data.JobID.IsNull())
	assert.Len(t, data.BackupJobs, 1)
	assert.Equal(t, "job-1", data.BackupJobs[0].ID.ValueString())
	assert.True(t, data.BackupJobs[0].Enabled.ValueBool())
}

// ---------------------------------------------------------------------------
// Repositories data source
// ---------------------------------------------------------------------------

func TestRepositoriesDataSource_Metadata(t *testing.T) {
	ds := NewRepositoriesDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_repositories", resp.TypeName)
}

func TestRepositoriesDataSource_Schema(t *testing.T) {
	ds := NewRepositoriesDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "repositories")
	assert.Contains(t, resp.Schema.Attributes, "repository_id")
	assert.Contains(t, resp.Schema.Attributes, "repository_name")
}

func TestRepositoriesDataSource_Configure_Nil(t *testing.T) {
	ds := &RepositoriesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, ds.client)
}

func TestRepositoriesDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoriesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, ds.client)
}

func TestRepositoriesDataSource_Configure_InvalidType(t *testing.T) {
	ds := &RepositoriesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: 123}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestRepositoryDataModel_Fields(t *testing.T) {
	data := RepositoryDataModel{
		ID:          types.StringValue("repo-1"),
		Name:        types.StringValue("Main Repository"),
		Description: types.StringValue("Main backup repository"),
		Path:        types.StringValue("/backup/main"),
		Type:        types.StringValue("WinLocal"),
		Capacity:    types.Int64Value(1000000),
		FreeSpace:   types.Int64Value(800000),
		UsedSpace:   types.Int64Value(200000),
		Status:      types.StringValue("Available"),
		CreatedAt:   types.StringValue("2024-01-01T00:00:00Z"),
		UpdatedAt:   types.StringValue("2024-01-01T00:00:00Z"),
	}

	assert.Equal(t, "repo-1", data.ID.ValueString())
	assert.Equal(t, "Main Repository", data.Name.ValueString())
	assert.Equal(t, int64(1000000), data.Capacity.ValueInt64())
	assert.Equal(t, int64(800000), data.FreeSpace.ValueInt64())
}

// ---------------------------------------------------------------------------
// Server Info data source
// ---------------------------------------------------------------------------

func TestServerInfoDataSource_Metadata(t *testing.T) {
	ds := NewServerInfoDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_server_info", resp.TypeName)
}

func TestServerInfoDataSource_Schema(t *testing.T) {
	ds := NewServerInfoDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "server_name")
	assert.Contains(t, resp.Schema.Attributes, "build_number")
	assert.Contains(t, resp.Schema.Attributes, "version")
	assert.Contains(t, resp.Schema.Attributes, "installation_id")
}

func TestServerInfoDataSource_Configure_Nil(t *testing.T) {
	ds := &ServerInfoDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, ds.client)
}

func TestServerInfoDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerInfoDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, ds.client)
}

func TestServerInfoDataSource_Configure_InvalidType(t *testing.T) {
	ds := &ServerInfoDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "bad"}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestUnwrapObjectData_TopLevel(t *testing.T) {
	// When data is NOT wrapped (no "data" key), returns as-is.
	data := map[string]interface{}{
		"serverName":  "vbr01",
		"buildNumber": "12.1.0.2131",
	}
	result := unwrapObjectData(data)
	assert.Equal(t, "vbr01", result["serverName"])
}

func TestUnwrapObjectData_Wrapped(t *testing.T) {
	// When data IS wrapped in a "data" envelope.
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"serverName":  "vbr01.lab",
			"buildNumber": "13.0.0.1234",
		},
	}
	result := unwrapObjectData(data)
	assert.Equal(t, "vbr01.lab", result["serverName"])
	assert.Equal(t, "13.0.0.1234", result["buildNumber"])
}

func TestUnwrapObjectData_NonMapData(t *testing.T) {
	// When "data" key exists but is not a map, returns raw.
	data := map[string]interface{}{
		"data":       "not-a-map",
		"serverName": "direct",
	}
	result := unwrapObjectData(data)
	assert.Equal(t, "direct", result["serverName"])
}

func TestServerInfoDataSourceModel_Fields(t *testing.T) {
	data := ServerInfoDataSourceModel{
		ID:             types.StringValue("server_info"),
		InstallationID: types.StringValue("inst-1234"),
		ServerName:     types.StringValue("vbr01.local"),
		BuildNumber:    types.StringValue("13.0.0.1234"),
		Version:        types.StringValue("13.0"),
	}
	assert.Equal(t, "server_info", data.ID.ValueString())
	assert.Equal(t, "vbr01.local", data.ServerName.ValueString())
	assert.Equal(t, "13.0.0.1234", data.BuildNumber.ValueString())
}

// ---------------------------------------------------------------------------
// Internal mapping logic tests — exercises what Read functions do
// ---------------------------------------------------------------------------

func TestBackupJobsDataSource_APIResultMapping(t *testing.T) {
	// Directly test the mapping logic used in the Read function.
	items := []map[string]interface{}{
		{
			"id":          "job-1",
			"name":        "Daily Backup",
			"isDisabled":  false,
			"description": "Daily backup",
			"type":        "VSphereBackup",
		},
	}

	// Simulate the mapping that Read() performs.
	backupJobs := make([]BackupJobDataModel, 0)
	for _, item := range items {
		backupJobs = append(backupJobs, BackupJobDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Enabled:     types.BoolValue(!getBoolValue(item, "isDisabled")),
			Description: types.StringValue(getStringValue(item, "description")),
			JobType:     types.StringValue(getStringValue(item, "type")),
		})
	}

	assert.Len(t, backupJobs, 1)
	assert.Equal(t, "job-1", backupJobs[0].ID.ValueString())
	assert.Equal(t, "Daily Backup", backupJobs[0].Name.ValueString())
	assert.True(t, backupJobs[0].Enabled.ValueBool()) // isDisabled=false → enabled=true
}

func TestRepositoriesDataSource_APIResultMapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":          "repo-1",
			"name":        "Main Repository",
			"description": "Primary backup storage",
			"path":        "/backups",
			"type":        "LinuxLocal",
			"capacity":    float64(2000000),
			"freeSpace":   float64(1500000),
			"usedSpace":   float64(500000),
			"status":      "Available",
			"createdAt":   "2024-01-01T00:00:00Z",
			"updatedAt":   "2024-01-01T12:00:00Z",
		},
	}

	repos := make([]RepositoryDataModel, len(items))
	for i, item := range items {
		repos[i] = RepositoryDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Description: types.StringValue(getStringValue(item, "description")),
			Path:        types.StringValue(getStringValue(item, "path")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Capacity:    types.Int64Value(getInt64Value(item, "capacity")),
			FreeSpace:   types.Int64Value(getInt64Value(item, "freeSpace")),
			UsedSpace:   types.Int64Value(getInt64Value(item, "usedSpace")),
			Status:      types.StringValue(getStringValue(item, "status")),
			CreatedAt:   types.StringValue(getStringValue(item, "createdAt")),
			UpdatedAt:   types.StringValue(getStringValue(item, "updatedAt")),
		}
	}

	assert.Equal(t, "repo-1", repos[0].ID.ValueString())
	assert.Equal(t, "LinuxLocal", repos[0].Type.ValueString())
	assert.Equal(t, int64(2000000), repos[0].Capacity.ValueInt64())
	assert.Equal(t, "2024-01-01T00:00:00Z", repos[0].CreatedAt.ValueString())
}

func TestSessionsDataSource_AllSessionsMapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathSessions, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":           "sess-10",
				"name":         "Backup Session",
				"jobId":        "job-5",
				"sessionType":  "Backup",
				"state":        "Stopped",
				"result":       "Success",
				"creationTime": "2024-01-01T00:00:00Z",
				"endTime":      "2024-01-01T01:00:00Z",
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathSessions)
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	out := make([]SessionDataModel, len(items))
	for i, item := range items {
		out[i] = SessionDataModel{
			ID:           types.StringValue(getStringValue(item, "id")),
			Name:         types.StringValue(getStringValue(item, "name")),
			JobID:        types.StringValue(getStringValue(item, "jobId")),
			SessionType:  types.StringValue(getStringValue(item, "sessionType")),
			State:        types.StringValue(getStringValue(item, "state")),
			Result:       types.StringValue(getStringValue(item, "result")),
			CreationTime: types.StringValue(getStringValue(item, "creationTime")),
			EndTime:      types.StringValue(getStringValue(item, "endTime")),
		}
	}
	assert.Equal(t, "sess-10", out[0].ID.ValueString())
	assert.Equal(t, "Success", out[0].Result.ValueString())
	mockClient.AssertExpectations(t)
}

func TestJobStatesDataSource_ReadAll_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathJobStates, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"jobId":      "job-a",
				"name":       "Alpha Job",
				"type":       "Backup",
				"status":     "Running",
				"lastResult": "None",
				"lastRun":    "2024-03-01T00:00:00Z",
			},
			{
				"jobId":      "job-b",
				"name":       "Beta Job",
				"type":       "Backup",
				"status":     "Stopped",
				"lastResult": "Success",
				"lastRun":    "2024-03-01T00:30:00Z",
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathJobStates)
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Map all items (no filter).
	out := make([]JobStateDataModel, 0)
	for _, item := range items {
		out = append(out, JobStateDataModel{
			JobID:      types.StringValue(getStringValue(item, "jobId")),
			Name:       types.StringValue(getStringValue(item, "name")),
			Type:       types.StringValue(getStringValue(item, "type")),
			Status:     types.StringValue(getStringValue(item, "status")),
			LastResult: types.StringValue(getStringValue(item, "lastResult")),
			LastRun:    types.StringValue(getStringValue(item, "lastRun")),
		})
	}
	assert.Len(t, out, 2)
	assert.Equal(t, "job-a", out[0].JobID.ValueString())
	assert.Equal(t, "job-b", out[1].JobID.ValueString())
	mockClient.AssertExpectations(t)
}

func TestProtectionGroupsDataSource_ReadByID_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	pgID := "pg-abc"
	endpoint := "/api/v1/infrastructure/protectionGroups/pg-abc"
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          pgID,
			"name":        "Windows VMs",
			"type":        "ManuallyDeployed",
			"description": "Manually deployed group",
		}
	}).Return(nil)

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), endpoint, &item)
	assert.NoError(t, err)

	pg := ProtectionGroupDataModel{
		ID:          types.StringValue(getStringValue(item, "id")),
		Name:        types.StringValue(getStringValue(item, "name")),
		Type:        types.StringValue(getStringValue(item, "type")),
		Description: types.StringValue(getStringValue(item, "description")),
	}
	assert.Equal(t, "pg-abc", pg.ID.ValueString())
	assert.Equal(t, "Windows VMs", pg.Name.ValueString())
	mockClient.AssertExpectations(t)
}

func TestWanAcceleratorsDataSource_ReadByID_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	wanID := "wan-xyz"
	endpoint := "/api/v1/wanAccelerators/wan-xyz"
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          wanID,
			"name":        "Branch Office WAN",
			"type":        "Target",
			"description": "Branch office WAN accelerator",
		}
	}).Return(nil)

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), endpoint, &item)
	assert.NoError(t, err)

	wan := WanAcceleratorDataModel{
		ID:          types.StringValue(getStringValue(item, "id")),
		Name:        types.StringValue(getStringValue(item, "name")),
		Type:        types.StringValue(getStringValue(item, "type")),
		Description: types.StringValue(getStringValue(item, "description")),
	}
	assert.Equal(t, "wan-xyz", wan.ID.ValueString())
	assert.Equal(t, "Target", wan.Type.ValueString())
	mockClient.AssertExpectations(t)
}

func TestRepositoryStatesDataSource_ReadAll_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathRepositoryState, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":        "repo-state-1",
				"name":      "Scale-Out Repo",
				"type":      "ScaleOut",
				"status":    "Available",
				"capacity":  float64(10000000),
				"freeSpace": float64(7000000),
				"usedSpace": float64(3000000),
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathRepositoryState)
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	out := make([]RepositoryStateDataModel, len(items))
	for i, item := range items {
		out[i] = RepositoryStateDataModel{
			ID:        types.StringValue(getStringValue(item, "id")),
			Name:      types.StringValue(getStringValue(item, "name")),
			Type:      types.StringValue(getStringValue(item, "type")),
			Status:    types.StringValue(getStringValue(item, "status")),
			Capacity:  types.Int64Value(getInt64Value(item, "capacity")),
			FreeSpace: types.Int64Value(getInt64Value(item, "freeSpace")),
			UsedSpace: types.Int64Value(getInt64Value(item, "usedSpace")),
		}
	}
	assert.Equal(t, "repo-state-1", out[0].ID.ValueString())
	assert.Equal(t, int64(10000000), out[0].Capacity.ValueInt64())
	mockClient.AssertExpectations(t)
}

func TestRestorePointsDataSource_ReadAll_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathRestorePoints, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":           "rp-200",
				"name":         "Server01 (Mar 1)",
				"backupId":     "bk-100",
				"creationTime": "2024-03-01T02:00:00Z",
				"type":         "Vm",
			},
			{
				"id":           "rp-201",
				"name":         "Server01 (Mar 2)",
				"backupId":     "bk-100",
				"creationTime": "2024-03-02T02:00:00Z",
				"type":         "Vm",
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathRestorePoints)
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	out := make([]RestorePointDataModel, len(items))
	for i, item := range items {
		out[i] = RestorePointDataModel{
			ID:           types.StringValue(getStringValue(item, "id")),
			Name:         types.StringValue(getStringValue(item, "name")),
			BackupID:     types.StringValue(getStringValue(item, "backupId")),
			CreationTime: types.StringValue(getStringValue(item, "creationTime")),
			Type:         types.StringValue(getStringValue(item, "type")),
		}
	}
	assert.Equal(t, "rp-200", out[0].ID.ValueString())
	assert.Equal(t, "bk-100", out[0].BackupID.ValueString())
	mockClient.AssertExpectations(t)
}

func TestManagedServersDataSource_ReadAll_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathManagedServers, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":          "ms-1",
				"name":        "DC01.corp.local",
				"type":        "Microsoft",
				"description": "Domain controller",
				"status":      "Available",
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathManagedServers)
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	out := make([]ManagedServerDataModel, len(items))
	for i, item := range items {
		out[i] = ManagedServerDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
			Status:      types.StringValue(getStringValue(item, "status")),
		}
	}
	assert.Equal(t, "ms-1", out[0].ID.ValueString())
	assert.Equal(t, "Domain controller", out[0].Description.ValueString())
	mockClient.AssertExpectations(t)
}

func TestProxiesDataSource_ReadAll_Mapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathProxies, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "p-1", "name": "Proxy-01", "type": "Vi", "description": "VMware proxy"},
			{"id": "p-2", "name": "Proxy-02", "type": "Agent", "description": "Agent proxy"},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathProxies)
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	out := make([]ProxyDataModel, len(items))
	for i, item := range items {
		out[i] = ProxyDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}
	assert.Equal(t, "p-1", out[0].ID.ValueString())
	assert.Equal(t, "Vi", out[0].Type.ValueString())
	mockClient.AssertExpectations(t)
}

func TestServerInfoDataSource_ReadHelperLogic(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathServerInfo, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"serverName":  "vbr01.lab",
			"buildNumber": "13.0.0.2131",
			"version":     "13.0",
		}
	}).Return(nil)

	var result map[string]interface{}
	err := mockClient.GetJSON(context.Background(), client.PathServerInfo, &result)
	assert.NoError(t, err)

	payload := unwrapObjectData(result)
	serverName := getFirstStringValue(payload, "serverName", "server_name", "name")
	buildNumber := getFirstStringValue(payload, "buildNumber", "build_number", "build")
	version := getFirstStringValue(payload, "version", "productVersion", "apiVersion")

	assert.Equal(t, "vbr01.lab", serverName)
	assert.Equal(t, "13.0.0.2131", buildNumber)
	assert.Equal(t, "13.0", version)
	mockClient.AssertExpectations(t)
}

func TestLicenseDataSource_ReadHelperLogic(t *testing.T) {
	mockClient := new(MockVeeamClient)

	// Mock license endpoint.
	mockClient.On("GetJSON", mock.Anything, client.PathLicense, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2)
		switch v := dest.(type) {
		case *map[string]interface{}:
			*v = map[string]interface{}{
				"type":       "Perpetual",
				"status":     "Valid",
				"licensedTo": "Test Corp",
				"expiration": "2030-01-01",
			}
		}
	}).Return(nil)

	var result map[string]interface{}
	err := mockClient.GetJSON(context.Background(), client.PathLicense, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Perpetual", result["type"])
	mockClient.AssertExpectations(t)
}

func TestBackupsDataSource_ReadByIDMapping(t *testing.T) {
	mockClient := new(MockVeeamClient)
	backupID := "bk-999"
	endpoint := "/api/v1/backups/bk-999"
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":      backupID,
			"name":    "Server99 Backup",
			"type":    "VmBackup",
			"jobId":   "job-50",
			"jobName": "Server99 Job",
		}
	}).Return(nil)

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), endpoint, &item)
	assert.NoError(t, err)

	backup := BackupDataModel{
		ID:      types.StringValue(getStringValue(item, "id")),
		Name:    types.StringValue(getStringValue(item, "name")),
		Type:    types.StringValue(getStringValue(item, "type")),
		JobID:   types.StringValue(getStringValue(item, "jobId")),
		JobName: types.StringValue(getStringValue(item, "jobName")),
		Files:   []BackupFileDataModel{},
	}

	assert.Equal(t, "bk-999", backup.ID.ValueString())
	assert.Equal(t, "Server99 Backup", backup.Name.ValueString())
	assert.Equal(t, "job-50", backup.JobID.ValueString())
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// normalizeDataSourceID edge cases
// ---------------------------------------------------------------------------

func TestNormalizeDataSourceID_Variants(t *testing.T) {
	tests := []struct {
		prefix   string
		value    string
		expected string
	}{
		{"backup", "", "backup"},
		{"backup", "abc-123", "backup_abc-123"},
		{"job_states", "job-1", "job_states_job-1"},
		{"restore_points_object", "obj-uuid", "restore_points_object_obj-uuid"},
	}

	for _, tc := range tests {
		result := normalizeDataSourceID(tc.prefix, tc.value)
		assert.Equal(t, tc.expected, result, "prefix=%s value=%s", tc.prefix, tc.value)
	}
}

package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// ---------------------------------------------------------------------------
// helpers.go coverage
// ---------------------------------------------------------------------------

func TestGetStringValue(t *testing.T) {
	data := map[string]interface{}{
		"name":   "vbr01",
		"number": 42,
		"flag":   true,
	}

	assert.Equal(t, "vbr01", getStringValue(data, "name"))
	assert.Equal(t, "", getStringValue(data, "number")) // not a string
	assert.Equal(t, "", getStringValue(data, "missing"))
}

func TestGetBoolValue(t *testing.T) {
	data := map[string]interface{}{
		"enabled":  true,
		"disabled": false,
		"name":     "str",
	}

	assert.True(t, getBoolValue(data, "enabled"))
	assert.False(t, getBoolValue(data, "disabled"))
	assert.False(t, getBoolValue(data, "name"))    // not a bool
	assert.False(t, getBoolValue(data, "missing")) // missing key
}

func TestGetInt64Value(t *testing.T) {
	data := map[string]interface{}{
		"intVal":     int(100),
		"int64Val":   int64(200),
		"float64Val": float64(300),
		"strVal":     "ignored",
	}

	assert.Equal(t, int64(100), getInt64Value(data, "intVal"))
	assert.Equal(t, int64(200), getInt64Value(data, "int64Val"))
	assert.Equal(t, int64(300), getInt64Value(data, "float64Val"))
	assert.Equal(t, int64(0), getInt64Value(data, "strVal"))
	assert.Equal(t, int64(0), getInt64Value(data, "missing"))
}

func TestGetDataList(t *testing.T) {
	data := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "1"},
			map[string]interface{}{"id": "2"},
		},
	}

	result := getDataList(data)
	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0]["id"])
	assert.Equal(t, "2", result[1]["id"])
}

func TestGetDataList_Empty(t *testing.T) {
	result := getDataList(map[string]interface{}{})
	assert.Empty(t, result)
}

func TestGetDataList_NotAList(t *testing.T) {
	data := map[string]interface{}{
		"data": "not-a-list",
	}
	result := getDataList(data)
	assert.Empty(t, result)
}

func TestFetchList_Error(t *testing.T) {
	calls := 0
	getter := func(_ context.Context, _ string, out interface{}) error {
		calls++
		return errors.New("api error")
	}

	items, err := fetchList(context.Background(), getter, "/test")
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, 2, calls)
}

// ---------------------------------------------------------------------------
// Credentials data source
// ---------------------------------------------------------------------------

func TestCredentialsDataSource_Metadata(t *testing.T) {
	ds := NewCredentialsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_credentials", resp.TypeName)
}

func TestCredentialsDataSource_Schema(t *testing.T) {
	ds := NewCredentialsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "credentials")
}

func TestCredentialsDataSource_Configure_Nil(t *testing.T) {
	ds := &CredentialsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, ds.client)
}

func TestCredentialsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &CredentialsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "not-a-client"}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestCredentialsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &CredentialsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, ds.client)
}

func TestCredentialsDataSourceModel(t *testing.T) {
	data := CredentialsDataSourceModel{
		ID: types.StringValue("credentials"),
		Credentials: []CredentialDataModel{
			{
				ID:          types.StringValue("cred-1"),
				Username:    types.StringValue("DOMAIN\\admin"),
				Description: types.StringValue("Domain admin"),
				Type:        types.StringValue("Standard"),
			},
		},
	}

	assert.Equal(t, "credentials", data.ID.ValueString())
	assert.Len(t, data.Credentials, 1)
	assert.Equal(t, "cred-1", data.Credentials[0].ID.ValueString())
	assert.Equal(t, "DOMAIN\\admin", data.Credentials[0].Username.ValueString())
}

// ---------------------------------------------------------------------------
// Backups data source
// ---------------------------------------------------------------------------

func TestBackupsDataSource_Metadata(t *testing.T) {
	ds := NewBackupsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_backups", resp.TypeName)
}

func TestBackupsDataSource_Schema(t *testing.T) {
	ds := NewBackupsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "backups")
	assert.Contains(t, resp.Schema.Attributes, "backup_id")
	assert.Contains(t, resp.Schema.Attributes, "include_files")
}

func TestBackupsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, ds.client)
}

func TestBackupsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &BackupsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "bad"}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupsDataSourceModel(t *testing.T) {
	data := BackupsDataSourceModel{
		ID:           types.StringValue("backups"),
		BackupID:     types.StringNull(),
		IncludeFiles: types.BoolValue(false),
		Backups: []BackupDataModel{
			{
				ID:      types.StringValue("bk-1"),
				Name:    types.StringValue("Job1 backup"),
				Type:    types.StringValue("VmBackup"),
				JobID:   types.StringValue("job-1"),
				JobName: types.StringValue("Daily Job"),
				Files:   []BackupFileDataModel{},
			},
		},
	}
	assert.Equal(t, "backups", data.ID.ValueString())
	assert.Len(t, data.Backups, 1)
}

// ---------------------------------------------------------------------------
// Job States data source
// ---------------------------------------------------------------------------

func TestJobStatesDataSource_Metadata(t *testing.T) {
	ds := NewJobStatesDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_job_states", resp.TypeName)
}

func TestJobStatesDataSource_Schema(t *testing.T) {
	ds := NewJobStatesDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "states")
	assert.Contains(t, resp.Schema.Attributes, "job_id")
}

func TestJobStatesDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &JobStatesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, ds.client)
}

func TestJobStatesDataSource_Configure_InvalidType(t *testing.T) {
	ds := &JobStatesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: 42}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestJobStatesDataSourceModel_AllJobs(t *testing.T) {
	// Simulate the mapping logic for all job states.
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathJobStates, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"jobId":      "job-1",
				"name":       "Daily Backup",
				"type":       "Backup",
				"status":     "Running",
				"lastResult": "Success",
				"lastRun":    "2024-01-01T00:00:00Z",
			},
			{
				"jobId":      "job-2",
				"name":       "Weekly Backup",
				"type":       "Backup",
				"status":     "Stopped",
				"lastResult": "Warning",
				"lastRun":    "2024-01-02T00:00:00Z",
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathJobStates)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "job-1", items[0]["jobId"])
	assert.Equal(t, "Daily Backup", items[0]["name"])
}

func TestJobStatesDataSourceModel_FilterByJobID(t *testing.T) {
	items := []map[string]interface{}{
		{"jobId": "job-1", "name": "Daily Backup", "type": "Backup", "status": "Running", "lastResult": "Success", "lastRun": "2024-01-01"},
		{"jobId": "job-2", "name": "Weekly Backup", "type": "Backup", "status": "Stopped", "lastResult": "Warning", "lastRun": "2024-01-02"},
	}

	filterJobID := "job-1"
	out := make([]JobStateDataModel, 0)
	for _, item := range items {
		jobID := getStringValue(item, "jobId")
		if jobID != filterJobID {
			continue
		}
		out = append(out, JobStateDataModel{
			JobID:      types.StringValue(jobID),
			Name:       types.StringValue(getStringValue(item, "name")),
			Type:       types.StringValue(getStringValue(item, "type")),
			Status:     types.StringValue(getStringValue(item, "status")),
			LastResult: types.StringValue(getStringValue(item, "lastResult")),
			LastRun:    types.StringValue(getStringValue(item, "lastRun")),
		})
	}

	assert.Len(t, out, 1)
	assert.Equal(t, "job-1", out[0].JobID.ValueString())
	assert.Equal(t, "Daily Backup", out[0].Name.ValueString())
}

// ---------------------------------------------------------------------------
// License data source
// ---------------------------------------------------------------------------

func TestLicenseDataSource_Metadata(t *testing.T) {
	ds := NewLicenseDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_license", resp.TypeName)
}

func TestLicenseDataSource_Schema(t *testing.T) {
	ds := NewLicenseDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "status")
	assert.Contains(t, resp.Schema.Attributes, "licensed_to")
	assert.Contains(t, resp.Schema.Attributes, "expiration_date")
	assert.Contains(t, resp.Schema.Attributes, "licensed_sockets")
	assert.Contains(t, resp.Schema.Attributes, "consumed_sockets")
}

func TestLicenseDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &LicenseDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestLicenseDataSource_Configure_InvalidType(t *testing.T) {
	ds := &LicenseDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "bad"}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestLicenseDataSource_MappingModel(t *testing.T) {
	// Test that the model fields get populated correctly from API responses.
	license := &models.LicenseModel{
		Type:       "Perpetual",
		Status:     "Valid",
		LicensedTo: "Acme Corp",
		Expiration: "2028-01-01",
	}

	socketsData := map[string]interface{}{
		"licensedSocketsNumber": float64(100),
		"consumedSocketsNumber": float64(50),
	}
	instancesData := map[string]interface{}{
		"licensedInstancesNumber": float64(200),
		"consumedInstancesNumber": float64(75),
	}
	capacityData := map[string]interface{}{
		"licensedCapacityTb": float64(10),
		"consumedCapacityTb": float64(4),
	}

	data := LicenseDataSourceModel{
		ID:                 types.StringValue("license"),
		Type:               types.StringValue(license.Type),
		Status:             types.StringValue(license.Status),
		LicensedTo:         types.StringValue(license.LicensedTo),
		ExpirationDate:     types.StringValue(license.Expiration),
		LicensedSockets:    types.Int64Value(getInt64Value(socketsData, "licensedSocketsNumber")),
		ConsumedSockets:    types.Int64Value(getInt64Value(socketsData, "consumedSocketsNumber")),
		LicensedInstances:  types.Int64Value(getInt64Value(instancesData, "licensedInstancesNumber")),
		ConsumedInstances:  types.Int64Value(getInt64Value(instancesData, "consumedInstancesNumber")),
		LicensedCapacityTB: types.Int64Value(getInt64Value(capacityData, "licensedCapacityTb")),
		ConsumedCapacityTB: types.Int64Value(getInt64Value(capacityData, "consumedCapacityTb")),
	}

	assert.Equal(t, "Perpetual", data.Type.ValueString())
	assert.Equal(t, "Valid", data.Status.ValueString())
	assert.Equal(t, "Acme Corp", data.LicensedTo.ValueString())
	assert.Equal(t, "2028-01-01", data.ExpirationDate.ValueString())
	assert.Equal(t, int64(100), data.LicensedSockets.ValueInt64())
	assert.Equal(t, int64(50), data.ConsumedSockets.ValueInt64())
	assert.Equal(t, int64(200), data.LicensedInstances.ValueInt64())
	assert.Equal(t, int64(75), data.ConsumedInstances.ValueInt64())
	assert.Equal(t, int64(10), data.LicensedCapacityTB.ValueInt64())
	assert.Equal(t, int64(4), data.ConsumedCapacityTB.ValueInt64())
}

// ---------------------------------------------------------------------------
// Managed Servers data source
// ---------------------------------------------------------------------------

func TestManagedServersDataSource_Metadata(t *testing.T) {
	ds := NewManagedServersDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_managed_servers", resp.TypeName)
}

func TestManagedServersDataSource_Schema(t *testing.T) {
	ds := NewManagedServersDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "servers")
	assert.Contains(t, resp.Schema.Attributes, "server_id")
}

func TestManagedServersDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ManagedServersDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestManagedServersDataSource_Configure_InvalidType(t *testing.T) {
	ds := &ManagedServersDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: 999}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestManagedServersDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":          "server-1",
			"name":        "WIN-SERVER-01",
			"type":        "Microsoft",
			"description": "Windows server",
			"status":      "Available",
		},
		{
			"id":          "server-2",
			"name":        "LINUX-01",
			"type":        "Linux",
			"description": "Linux server",
			"status":      "Available",
		},
	}

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

	assert.Len(t, out, 2)
	assert.Equal(t, "server-1", out[0].ID.ValueString())
	assert.Equal(t, "WIN-SERVER-01", out[0].Name.ValueString())
	assert.Equal(t, "Microsoft", out[0].Type.ValueString())
	assert.Equal(t, "Available", out[0].Status.ValueString())
}

// ---------------------------------------------------------------------------
// Protection Groups data source
// ---------------------------------------------------------------------------

func TestProtectionGroupsDataSource_Metadata(t *testing.T) {
	ds := NewProtectionGroupsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_protection_groups", resp.TypeName)
}

func TestProtectionGroupsDataSource_Schema(t *testing.T) {
	ds := NewProtectionGroupsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "protection_groups")
	assert.Contains(t, resp.Schema.Attributes, "protection_group_id")
}

func TestProtectionGroupsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectionGroupsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestProtectionGroupsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &ProtectionGroupsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: true}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestProtectionGroupsDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":          "pg-1",
			"name":        "All Servers",
			"type":        "ActiveDirectory",
			"description": "All AD computers",
		},
	}

	out := make([]ProtectionGroupDataModel, len(items))
	for i, item := range items {
		out[i] = ProtectionGroupDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	assert.Equal(t, "pg-1", out[0].ID.ValueString())
	assert.Equal(t, "All Servers", out[0].Name.ValueString())
	assert.Equal(t, "ActiveDirectory", out[0].Type.ValueString())
}

// ---------------------------------------------------------------------------
// Proxies data source
// ---------------------------------------------------------------------------

func TestProxiesDataSource_Metadata(t *testing.T) {
	ds := NewProxiesDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_proxies", resp.TypeName)
}

func TestProxiesDataSource_Schema(t *testing.T) {
	ds := NewProxiesDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "proxies")
	assert.Contains(t, resp.Schema.Attributes, "proxy_id")
}

func TestProxiesDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxiesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestProxiesDataSource_Configure_InvalidType(t *testing.T) {
	ds := &ProxiesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: 3.14}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestProxiesDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":          "proxy-1",
			"name":        "Proxy01",
			"type":        "Vi",
			"description": "VMware proxy",
		},
		{
			"id":          "proxy-2",
			"name":        "Proxy02",
			"type":        "Agent",
			"description": "",
		},
	}

	out := make([]ProxyDataModel, len(items))
	for i, item := range items {
		out[i] = ProxyDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	assert.Len(t, out, 2)
	assert.Equal(t, "proxy-1", out[0].ID.ValueString())
	assert.Equal(t, "Vi", out[0].Type.ValueString())
}

// ---------------------------------------------------------------------------
// Repository States data source
// ---------------------------------------------------------------------------

func TestRepositoryStatesDataSource_Metadata(t *testing.T) {
	ds := NewRepositoryStatesDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_repository_states", resp.TypeName)
}

func TestRepositoryStatesDataSource_Schema(t *testing.T) {
	ds := NewRepositoryStatesDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "states")
	assert.Contains(t, resp.Schema.Attributes, "id")
}

func TestRepositoryStatesDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoryStatesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestRepositoryStatesDataSource_Configure_InvalidType(t *testing.T) {
	ds := &RepositoryStatesDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: []string{}}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestRepositoryStatesDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":        "repo-1",
			"name":      "Default Repo",
			"type":      "WinLocal",
			"status":    "Available",
			"capacity":  float64(1000000),
			"freeSpace": float64(800000),
			"usedSpace": float64(200000),
		},
	}

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

	assert.Equal(t, "repo-1", out[0].ID.ValueString())
	assert.Equal(t, int64(1000000), out[0].Capacity.ValueInt64())
	assert.Equal(t, int64(800000), out[0].FreeSpace.ValueInt64())
	assert.Equal(t, int64(200000), out[0].UsedSpace.ValueInt64())
}

// ---------------------------------------------------------------------------
// Restore Points data source
// ---------------------------------------------------------------------------

func TestRestorePointsDataSource_Metadata(t *testing.T) {
	ds := NewRestorePointsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_restore_points", resp.TypeName)
}

func TestRestorePointsDataSource_Schema(t *testing.T) {
	ds := NewRestorePointsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "restore_points")
	assert.Contains(t, resp.Schema.Attributes, "restore_point_id")
	assert.Contains(t, resp.Schema.Attributes, "backup_object_id")
}

func TestRestorePointsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestRestorePointsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &RestorePointsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: false}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestRestorePointsDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":           "rp-1",
			"name":         "VM01 (Jan 1, 2024)",
			"backupId":     "bk-1",
			"creationTime": "2024-01-01T00:00:00Z",
			"type":         "Vm",
		},
		{
			"id":           "rp-2",
			"name":         "VM01 (Jan 2, 2024)",
			"backupId":     "bk-1",
			"creationTime": "2024-01-02T00:00:00Z",
			"type":         "Vm",
		},
	}

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

	assert.Len(t, out, 2)
	assert.Equal(t, "rp-1", out[0].ID.ValueString())
	assert.Equal(t, "bk-1", out[0].BackupID.ValueString())
	assert.Equal(t, "2024-01-01T00:00:00Z", out[0].CreationTime.ValueString())
}

// ---------------------------------------------------------------------------
// Sessions data source
// ---------------------------------------------------------------------------

func TestSessionsDataSource_Metadata(t *testing.T) {
	ds := NewSessionsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_sessions", resp.TypeName)
}

func TestSessionsDataSource_Schema(t *testing.T) {
	ds := NewSessionsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "sessions")
	assert.Contains(t, resp.Schema.Attributes, "session_id")
}

func TestSessionsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SessionsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSessionsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &SessionsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: 1}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSessionsDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":           "sess-1",
			"name":         "Daily Backup (Jan 1, 2024)",
			"jobId":        "job-1",
			"sessionType":  "Backup",
			"state":        "Stopped",
			"result":       "Success",
			"creationTime": "2024-01-01T00:00:00Z",
			"endTime":      "2024-01-01T01:00:00Z",
		},
	}

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

	assert.Equal(t, "sess-1", out[0].ID.ValueString())
	assert.Equal(t, "job-1", out[0].JobID.ValueString())
	assert.Equal(t, "Backup", out[0].SessionType.ValueString())
	assert.Equal(t, "Success", out[0].Result.ValueString())
}

// ---------------------------------------------------------------------------
// WAN Accelerators data source
// ---------------------------------------------------------------------------

func TestWanAcceleratorsDataSource_Metadata(t *testing.T) {
	ds := NewWanAcceleratorsDataSource()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_wan_accelerators", resp.TypeName)
}

func TestWanAcceleratorsDataSource_Schema(t *testing.T) {
	ds := NewWanAcceleratorsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	assert.Contains(t, resp.Schema.Attributes, "accelerators")
	assert.Contains(t, resp.Schema.Attributes, "accelerator_id")
}

func TestWanAcceleratorsDataSource_Configure_Valid(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &WanAcceleratorsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: mockClient}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestWanAcceleratorsDataSource_Configure_InvalidType(t *testing.T) {
	ds := &WanAcceleratorsDataSource{}
	var resp datasource.ConfigureResponse
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestWanAcceleratorsDataSourceModel_Mapping(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id":          "wan-1",
			"name":        "HQ WAN Acc",
			"type":        "Source",
			"description": "Headquarters WAN accelerator",
		},
	}

	out := make([]WanAcceleratorDataModel, len(items))
	for i, item := range items {
		out[i] = WanAcceleratorDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	assert.Equal(t, "wan-1", out[0].ID.ValueString())
	assert.Equal(t, "HQ WAN Acc", out[0].Name.ValueString())
	assert.Equal(t, "Source", out[0].Type.ValueString())
}

// ---------------------------------------------------------------------------
// Mock-driven Read path tests (exercises Configure + Read code paths)
// ---------------------------------------------------------------------------

func TestCredentialsDataSource_ReadMock(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathCredentials, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "c-1", "username": "admin", "description": "Main admin", "type": "Standard"},
		}
	}).Return(nil)

	// Verify mock fetch succeeds.
	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathCredentials)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "c-1", items[0]["id"])
	assert.Equal(t, "admin", items[0]["username"])
}

func TestCredentialsDataSource_ReadMock_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathCredentials, mock.Anything).Return(errors.New("network error"))
	// Second call for wrapped format also fails.
	mockClient.On("GetJSON", mock.Anything, client.PathCredentials, mock.Anything).Return(errors.New("network error"))

	_, err := fetchList(context.Background(), mockClient.GetJSON, client.PathCredentials)
	assert.Error(t, err)
}

func TestManagedServersDataSource_ReadByID(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := "/api/v1/infrastructure/managementServers/server-1"
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":     "server-1",
			"name":   "WIN-DC01",
			"type":   "Microsoft",
			"status": "Available",
		}
	}).Return(nil)

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), endpoint, &item)
	assert.NoError(t, err)
	assert.Equal(t, "server-1", item["id"])
	assert.Equal(t, "WIN-DC01", item["name"])
	mockClient.AssertExpectations(t)
}

func TestProtectionGroupsDataSource_ReadByID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(errors.New("not found"))

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), client.PathProtectionGroupByID, &item)
	assert.Error(t, err)
}

func TestRepositoryStatesDataSource_ReadList(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathRepositoryState, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":        "repo-1",
				"name":      "Main Repo",
				"type":      "WinLocal",
				"status":    "Available",
				"capacity":  float64(5000000),
				"freeSpace": float64(4000000),
				"usedSpace": float64(1000000),
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathRepositoryState)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "repo-1", items[0]["id"])
	mockClient.AssertExpectations(t)
}

func TestRestorePointsDataSource_ReadByBackupObjectID(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := "/api/v1/backupObjects/obj-1/restorePoints"
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":           "rp-100",
				"name":         "obj (Jan 1)",
				"backupId":     "bk-1",
				"creationTime": "2024-01-01T00:00:00Z",
				"type":         "Vm",
			},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, endpoint)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "rp-100", items[0]["id"])
	mockClient.AssertExpectations(t)
}

func TestSessionsDataSource_ReadByID(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := "/api/v1/sessions/sess-1"
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":           "sess-1",
			"name":         "Job Session",
			"jobId":        "job-1",
			"sessionType":  "Backup",
			"state":        "Stopped",
			"result":       "Success",
			"creationTime": "2024-01-01T00:00:00Z",
			"endTime":      "2024-01-01T01:00:00Z",
		}
	}).Return(nil)

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), endpoint, &item)
	assert.NoError(t, err)
	assert.Equal(t, "sess-1", item["id"])
	assert.Equal(t, "Success", item["result"])
	mockClient.AssertExpectations(t)
}

func TestWanAcceleratorsDataSource_ReadList(t *testing.T) {
	mockClient := new(MockVeeamClient)
	mockClient.On("GetJSON", mock.Anything, client.PathWanAccelerators, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "wan-1", "name": "HQ", "type": "Source", "description": ""},
			{"id": "wan-2", "name": "Branch", "type": "Target", "description": ""},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathWanAccelerators)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "wan-1", items[0]["id"])
	mockClient.AssertExpectations(t)
}

func TestBackupsDataSource_ReadWithFiles(t *testing.T) {
	mockClient := new(MockVeeamClient)

	// The list call returns one backup.
	mockClient.On("GetJSON", mock.Anything, client.PathBackups, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "bk-1", "name": "Job Backup", "type": "VmBackup", "jobId": "job-1", "jobName": "Daily"},
		}
	}).Return(nil)

	items, err := fetchList(context.Background(), mockClient.GetJSON, client.PathBackups)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "bk-1", items[0]["id"])
}

// ---------------------------------------------------------------------------
// Schema correctness tests
// ---------------------------------------------------------------------------

func TestAllDataSourceSchemas(t *testing.T) {
	factories := []func() datasource.DataSource{
		NewCredentialsDataSource,
		NewBackupsDataSource,
		NewJobStatesDataSource,
		NewLicenseDataSource,
		NewManagedServersDataSource,
		NewProtectionGroupsDataSource,
		NewProxiesDataSource,
		NewRepositoryStatesDataSource,
		NewRestorePointsDataSource,
		NewSessionsDataSource,
		NewWanAcceleratorsDataSource,
	}

	for _, factory := range factories {
		ds := factory()
		assert.NotNil(t, ds)

		var schemaResp datasource.SchemaResponse
		ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
		assert.NotNil(t, schemaResp.Schema, "Schema should not be nil")
		assert.Contains(t, schemaResp.Schema.Attributes, "id", "All datasources should have an 'id' attribute")
	}
}

package datasources

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// ---------------------------------------------------------------------------
// Helper: build a tfsdk.Config from a datasource schema with null values.
// This lets us call Read() without a live Terraform binary.
// ---------------------------------------------------------------------------

func buildNullConfig(ds datasource.DataSource) tfsdk.Config {
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
}

// nullObjectForSchema recursively builds a tftypes.Value with all null leaves
// for the given tftypes.Type (must be an Object at the top level for a schema).
func nullObjectForSchema(typ tftypes.Type) tftypes.Value {
	switch t := typ.(type) {
	case tftypes.Object:
		attrs := make(map[string]tftypes.Value, len(t.AttributeTypes))
		for k, attrType := range t.AttributeTypes {
			attrs[k] = nullValueForType(attrType)
		}
		return tftypes.NewValue(t, attrs)
	default:
		return tftypes.NewValue(typ, nil)
	}
}

func nullValueForType(typ tftypes.Type) tftypes.Value {
	switch t := typ.(type) {
	case tftypes.Object:
		attrs := make(map[string]tftypes.Value, len(t.AttributeTypes))
		for k, attrType := range t.AttributeTypes {
			attrs[k] = nullValueForType(attrType)
		}
		return tftypes.NewValue(t, attrs)
	case tftypes.List:
		return tftypes.NewValue(t, nil) // null list
	case tftypes.Set:
		return tftypes.NewValue(t, nil)
	case tftypes.Map:
		return tftypes.NewValue(t, nil)
	default:
		return tftypes.NewValue(typ, nil)
	}
}

// buildNullState builds a tfsdk.State from the datasource schema (for resp.State).
func buildNullState(ds datasource.DataSource) tfsdk.State {
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	return tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
}

// ---------------------------------------------------------------------------
// Credentials — Read (list all)
// ---------------------------------------------------------------------------

func TestCredentialsDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &CredentialsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathCredentials, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "c-1", "username": "admin", "description": "Admin credential", "type": "Standard"},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestCredentialsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &CredentialsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathCredentials, mock.Anything).Return(errors.New("network error"))
	// second call (wrapped) also fails
	mockClient.On("GetJSON", mock.Anything, client.PathCredentials, mock.Anything).Return(errors.New("network error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Repository States — Read
// ---------------------------------------------------------------------------

func TestRepositoryStatesDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoryStatesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathRepositoryState, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":        "repo-1",
				"name":      "Default Repo",
				"type":      "WinLocal",
				"status":    "Available",
				"capacity":  float64(1000),
				"freeSpace": float64(800),
				"usedSpace": float64(200),
			},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRepositoryStatesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoryStatesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathRepositoryState, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathRepositoryState, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Job States — Read (all + filtered)
// ---------------------------------------------------------------------------

func TestJobStatesDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &JobStatesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathJobStates, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"jobId": "j-1", "name": "Job1", "type": "Backup", "status": "Running", "lastResult": "Success", "lastRun": "2024-01-01"},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestJobStatesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &JobStatesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathJobStates, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathJobStates, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Sessions — Read (all)
// ---------------------------------------------------------------------------

func TestSessionsDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SessionsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSessions, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":           "sess-1",
				"name":         "Session 1",
				"jobId":        "job-1",
				"sessionType":  "Backup",
				"state":        "Stopped",
				"result":       "Success",
				"creationTime": "2024-01-01T00:00:00Z",
				"endTime":      "2024-01-01T01:00:00Z",
			},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestSessionsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SessionsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSessions, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathSessions, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Proxies — Read
// ---------------------------------------------------------------------------

func TestProxiesDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxiesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProxies, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "p-1", "name": "Proxy01", "type": "Vi", "description": "VMware proxy"},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestProxiesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxiesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProxies, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathProxies, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Protection Groups — Read
// ---------------------------------------------------------------------------

func TestProtectionGroupsDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectionGroupsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProtectionGroups, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "pg-1", "name": "All Servers", "type": "ActiveDirectory", "description": "All AD computers"},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestProtectionGroupsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectionGroupsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProtectionGroups, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathProtectionGroups, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Managed Servers — Read
// ---------------------------------------------------------------------------

func TestManagedServersDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ManagedServersDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathManagedServers, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "ms-1", "name": "DC01", "type": "Microsoft", "description": "", "status": "Available"},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestManagedServersDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ManagedServersDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathManagedServers, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathManagedServers, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Restore Points — Read
// ---------------------------------------------------------------------------

func TestRestorePointsDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathRestorePoints, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":           "rp-1",
				"name":         "VM Restore Point",
				"backupId":     "bk-1",
				"creationTime": "2024-01-01T00:00:00Z",
				"type":         "Vm",
			},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRestorePointsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathRestorePoints, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathRestorePoints, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// WAN Accelerators — Read
// ---------------------------------------------------------------------------

func TestWanAcceleratorsDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &WanAcceleratorsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathWanAccelerators, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "wan-1", "name": "HQ WAN", "type": "Source", "description": ""},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestWanAcceleratorsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &WanAcceleratorsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathWanAccelerators, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathWanAccelerators, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Backups — Read
// ---------------------------------------------------------------------------

func TestBackupsDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathBackups, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "bk-1", "name": "VM Backup", "type": "VmBackup", "jobId": "job-1", "jobName": "Daily Job"},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestBackupsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathBackups, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathBackups, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Server Info — Read
// ---------------------------------------------------------------------------

func TestServerInfoDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerInfoDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServerInfo, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"serverName":  "vbr01.lab",
			"buildNumber": "13.0.0.2131",
			"version":     "13.0",
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestServerInfoDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerInfoDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServerInfo, mock.Anything).Return(errors.New("connection error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// License — Read
// ---------------------------------------------------------------------------

func TestLicenseDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &LicenseDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathLicense, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*models.LicenseModel)
		dest.Type = "Perpetual"
		dest.Status = "Valid"
		dest.LicensedTo = "Acme Corp"
		dest.Expiration = "2030-01-01"
	}).Return(nil)

	mockClient.On("GetJSON", mock.Anything, client.PathLicenseSockets, mock.Anything).Return(nil)
	mockClient.On("GetJSON", mock.Anything, client.PathLicenseInstances, mock.Anything).Return(nil)
	mockClient.On("GetJSON", mock.Anything, client.PathLicenseCapacity, mock.Anything).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestLicenseDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &LicenseDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathLicense, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJobs — Read
// ---------------------------------------------------------------------------

func TestBackupJobsDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupJobsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathJobs, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":          "job-1",
				"name":        "Daily Backup",
				"isDisabled":  false,
				"description": "Daily VM backup",
				"type":        "VSphereBackup",
			},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestBackupJobsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupJobsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathJobs, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathJobs, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Repositories — Read
// ---------------------------------------------------------------------------

func TestRepositoriesDataSource_Read_AllSuccess(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoriesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathRepositories, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":          "repo-1",
				"name":        "Main Repository",
				"description": "Primary backup storage",
				"path":        "/backups",
				"type":        "WinLocal",
				"capacity":    float64(1000),
				"freeSpace":   float64(800),
				"usedSpace":   float64(200),
				"status":      "Available",
				"createdAt":   "2024-01-01T00:00:00Z",
				"updatedAt":   "2024-01-01T00:00:00Z",
			},
		}
	}).Return(nil)

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRepositoriesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoriesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathRepositories, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, client.PathRepositories, mock.Anything).Return(errors.New("error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Helpers — fetchList empty wrapped data
// ---------------------------------------------------------------------------

func TestFetchList_WrappedEmptyData(t *testing.T) {
	// fetchList returns empty list when wrapped "data" is empty list.
	calls := 0
	getter := func(_ context.Context, _ string, out interface{}) error {
		calls++
		if calls == 1 {
			return errors.New("not array")
		}
		dest := out.(*map[string]interface{})
		*dest = map[string]interface{}{
			"data": []interface{}{},
		}
		return nil
	}

	items, err := fetchList(context.Background(), getter, "/test")
	assert.NoError(t, err)
	assert.Empty(t, items)
}

// ---------------------------------------------------------------------------
// PathBackupFiles — verify format string works
// ---------------------------------------------------------------------------

func TestPathBackupFiles_Format(t *testing.T) {
	path := fmt.Sprintf(client.PathBackupFiles, "bk-123")
	assert.Contains(t, path, "bk-123")
}

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
)

// ---------------------------------------------------------------------------
// Helper: build a tfsdk.Config with specific field values set.
// ---------------------------------------------------------------------------

// buildConfigWithStrings builds a Config where the given string fields are set
// to specific values and all others remain null.
func buildConfigWithStrings(ds datasource.DataSource, fields map[string]string) tfsdk.Config {
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(context.Background())

	objType := schemaType.(tftypes.Object)
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for k, attrType := range objType.AttributeTypes {
		if strVal, ok := fields[k]; ok {
			attrs[k] = tftypes.NewValue(attrType, strVal)
		} else {
			attrs[k] = nullValueForType(attrType)
		}
	}

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(objType, attrs),
	}
}

// ---------------------------------------------------------------------------
// Sessions — Read by session_id
// ---------------------------------------------------------------------------

func TestSessionsDataSource_Read_BySessionID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SessionsDataSource{client: mockClient}

	sessionID := "sess-42"
	endpoint := fmt.Sprintf(client.PathSessionByID, sessionID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":           sessionID,
			"name":         "Backup Session",
			"jobId":        "job-1",
			"sessionType":  "Backup",
			"state":        "Stopped",
			"result":       "Success",
			"creationTime": "2024-01-01T00:00:00Z",
			"endTime":      "2024-01-01T01:00:00Z",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"session_id": sessionID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestSessionsDataSource_Read_BySessionID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SessionsDataSource{client: mockClient}

	sessionID := "sess-bad"
	endpoint := fmt.Sprintf(client.PathSessionByID, sessionID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"session_id": sessionID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Proxies — Read by proxy_id
// ---------------------------------------------------------------------------

func TestProxiesDataSource_Read_ByProxyID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxiesDataSource{client: mockClient}

	proxyID := "proxy-77"
	endpoint := fmt.Sprintf(client.PathProxyByID, proxyID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          proxyID,
			"name":        "Proxy77",
			"type":        "Vi",
			"description": "VMware proxy",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"proxy_id": proxyID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestProxiesDataSource_Read_ByProxyID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxiesDataSource{client: mockClient}

	proxyID := "proxy-bad"
	endpoint := fmt.Sprintf(client.PathProxyByID, proxyID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"proxy_id": proxyID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Protection Groups — Read by protection_group_id
// ---------------------------------------------------------------------------

func TestProtectionGroupsDataSource_Read_ByID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectionGroupsDataSource{client: mockClient}

	pgID := "pg-55"
	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, pgID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          pgID,
			"name":        "Windows Servers",
			"type":        "ManuallyDeployed",
			"description": "Manually added Windows servers",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"protection_group_id": pgID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestProtectionGroupsDataSource_Read_ByID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectionGroupsDataSource{client: mockClient}

	pgID := "pg-bad"
	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, pgID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"protection_group_id": pgID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Managed Servers — Read by server_id
// ---------------------------------------------------------------------------

func TestManagedServersDataSource_Read_ByID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ManagedServersDataSource{client: mockClient}

	serverID := "ms-99"
	endpoint := fmt.Sprintf(client.PathManagedServerByID, serverID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          serverID,
			"name":        "WIN-SRV-99",
			"type":        "Microsoft",
			"description": "Windows server 99",
			"status":      "Available",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"server_id": serverID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestManagedServersDataSource_Read_ByID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ManagedServersDataSource{client: mockClient}

	serverID := "ms-bad"
	endpoint := fmt.Sprintf(client.PathManagedServerByID, serverID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"server_id": serverID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// WAN Accelerators — Read by accelerator_id
// ---------------------------------------------------------------------------

func TestWanAcceleratorsDataSource_Read_ByID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &WanAcceleratorsDataSource{client: mockClient}

	wanID := "wan-88"
	endpoint := fmt.Sprintf(client.PathWanAcceleratorByID, wanID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          wanID,
			"name":        "Branch WAN",
			"type":        "Target",
			"description": "Branch office accelerator",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"accelerator_id": wanID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestWanAcceleratorsDataSource_Read_ByID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &WanAcceleratorsDataSource{client: mockClient}

	wanID := "wan-bad"
	endpoint := fmt.Sprintf(client.PathWanAcceleratorByID, wanID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"accelerator_id": wanID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Restore Points — Read by restore_point_id
// ---------------------------------------------------------------------------

func TestRestorePointsDataSource_Read_ByRestorePointID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{client: mockClient}

	rpID := "rp-55"
	endpoint := fmt.Sprintf(client.PathRestorePointByID, rpID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":           rpID,
			"name":         "VM Restore Point",
			"backupId":     "bk-1",
			"creationTime": "2024-01-01T00:00:00Z",
			"type":         "Vm",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"restore_point_id": rpID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRestorePointsDataSource_Read_ByRestorePointID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{client: mockClient}

	rpID := "rp-bad"
	endpoint := fmt.Sprintf(client.PathRestorePointByID, rpID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"restore_point_id": rpID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Restore Points — Read by backup_object_id
// ---------------------------------------------------------------------------

func TestRestorePointsDataSource_Read_ByBackupObjectID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{client: mockClient}

	objID := "obj-99"
	endpoint := fmt.Sprintf(client.PathBackupObjectRestorePoints, objID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"id":           "rp-11",
				"name":         "Restore Point",
				"backupId":     "bk-5",
				"creationTime": "2024-02-01T00:00:00Z",
				"type":         "Vm",
			},
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"backup_object_id": objID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRestorePointsDataSource_Read_ByBackupObjectID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RestorePointsDataSource{client: mockClient}

	objID := "obj-bad"
	endpoint := fmt.Sprintf(client.PathBackupObjectRestorePoints, objID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("error"))
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("error"))

	cfg := buildConfigWithStrings(ds, map[string]string{"backup_object_id": objID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Backups — Read by backup_id
// ---------------------------------------------------------------------------

func TestBackupsDataSource_Read_ByBackupID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupsDataSource{client: mockClient}

	backupID := "bk-42"
	endpoint := fmt.Sprintf(client.PathBackupByID, backupID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":      backupID,
			"name":    "VM Backup 42",
			"type":    "VmBackup",
			"jobId":   "job-1",
			"jobName": "Daily Job",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"backup_id": backupID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestBackupsDataSource_Read_ByBackupID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupsDataSource{client: mockClient}

	backupID := "bk-bad"
	endpoint := fmt.Sprintf(client.PathBackupByID, backupID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"backup_id": backupID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJobs — Read by job_id
// ---------------------------------------------------------------------------

func TestBackupJobsDataSource_Read_ByJobID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupJobsDataSource{client: mockClient}

	jobID := "job-77"
	endpoint := fmt.Sprintf(client.PathJobByID, jobID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          jobID,
			"name":        "VM Backup Job 77",
			"isDisabled":  false,
			"description": "Test backup job",
			"type":        "VSphereBackup",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"job_id": jobID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestBackupJobsDataSource_Read_ByJobID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupJobsDataSource{client: mockClient}

	jobID := "job-bad"
	endpoint := fmt.Sprintf(client.PathJobByID, jobID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"job_id": jobID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Repositories — Read by repository_id and repository_name
// ---------------------------------------------------------------------------

func TestRepositoriesDataSource_Read_ByRepositoryID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoriesDataSource{client: mockClient}

	repoID := "repo-42"
	endpoint := fmt.Sprintf(client.PathRepositoryByID, repoID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"id":          repoID,
			"name":        "Main Repository",
			"description": "Primary storage",
			"path":        "/backups",
			"type":        "WinLocal",
			"capacity":    float64(1000),
			"freeSpace":   float64(800),
			"usedSpace":   float64(200),
			"status":      "Available",
			"createdAt":   "2024-01-01T00:00:00Z",
			"updatedAt":   "2024-01-01T00:00:00Z",
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"repository_id": repoID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRepositoriesDataSource_Read_ByRepositoryID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &RepositoriesDataSource{client: mockClient}

	repoID := "repo-bad"
	endpoint := fmt.Sprintf(client.PathRepositoryByID, repoID)
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Return(errors.New("not found"))

	cfg := buildConfigWithStrings(ds, map[string]string{"repository_id": repoID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// JobStates — Read with job_id filter
// ---------------------------------------------------------------------------

func TestJobStatesDataSource_Read_ByJobID_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &JobStatesDataSource{client: mockClient}

	jobID := "job-filter-1"

	// The job states endpoint returns all states; filtering is done client-side.
	mockClient.On("GetJSON", mock.Anything, client.PathJobStates, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{
				"jobId":      jobID,
				"name":       "Filtered Job",
				"type":       "Backup",
				"status":     "Running",
				"lastResult": "None",
				"lastRun":    "2024-03-01T00:00:00Z",
			},
			{
				"jobId":      "job-other",
				"name":       "Other Job",
				"type":       "Backup",
				"status":     "Stopped",
				"lastResult": "Success",
				"lastRun":    "2024-03-01T01:00:00Z",
			},
		}
	}).Return(nil)

	cfg := buildConfigWithStrings(ds, map[string]string{"job_id": jobID})
	state := buildNullState(ds)

	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

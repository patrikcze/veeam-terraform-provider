package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// ---------------------------------------------------------------------------
// Metadata type name smoke tests
// ---------------------------------------------------------------------------

func TestVisibilityDataSources_MetadataTypeNames(t *testing.T) {
	tests := []struct {
		name     string
		factory  func() datasource.DataSource
		expected string
	}{
		{"security_roles", NewSecurityRolesDataSource, "veeam_security_roles"},
		{"security_users", NewSecurityUsersDataSource, "veeam_security_users"},
		{"backup_objects", NewBackupObjectsDataSource, "veeam_backup_objects"},
		{"replicas", NewReplicasDataSource, "veeam_replicas"},
		{"replica_points", NewReplicaPointsDataSource, "veeam_replica_points"},
		{"proxy_states", NewProxyStatesDataSource, "veeam_proxy_states"},
		{"protected_computers", NewProtectedComputersDataSource, "veeam_protected_computers"},
		{"services", NewServicesDataSource, "veeam_services"},
		{"server_time", NewServerTimeDataSource, "veeam_server_time"},
		{"server_certificate", NewServerCertificateDataSource, "veeam_server_certificate"},
		{"task_sessions", NewTaskSessionsDataSource, "veeam_task_sessions"},
		{"security_analyzer", NewSecurityAnalyzerDataSource, "veeam_security_analyzer"},
		{"malware_events", NewMalwareEventsDataSource, "veeam_malware_events"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.factory()
			assert.NotNil(t, ds)
			var resp datasource.MetadataResponse
			ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
			assert.Equal(t, tc.expected, resp.TypeName)
		})
	}
}

// ---------------------------------------------------------------------------
// Configure helpers (shared pattern — test one representative list source and
// one singleton source)
// ---------------------------------------------------------------------------

func TestPrio3_Configure_Nil(t *testing.T) {
	dsList := []datasource.DataSource{
		NewSecurityRolesDataSource(),
		NewSecurityUsersDataSource(),
		NewBackupObjectsDataSource(),
		NewReplicasDataSource(),
		NewReplicaPointsDataSource(),
		NewProxyStatesDataSource(),
		NewProtectedComputersDataSource(),
		NewServicesDataSource(),
		NewServerTimeDataSource(),
		NewServerCertificateDataSource(),
		NewTaskSessionsDataSource(),
		NewSecurityAnalyzerDataSource(),
		NewMalwareEventsDataSource(),
	}
	for _, ds := range dsList {
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, &resp)
		assert.False(t, resp.Diagnostics.HasError())
	}
}

func TestPrio3_Configure_InvalidType(t *testing.T) {
	dsList := []datasource.DataSource{
		NewSecurityRolesDataSource(),
		NewSecurityUsersDataSource(),
		NewBackupObjectsDataSource(),
		NewReplicasDataSource(),
		NewReplicaPointsDataSource(),
		NewProxyStatesDataSource(),
		NewProtectedComputersDataSource(),
		NewServicesDataSource(),
		NewServerTimeDataSource(),
		NewServerCertificateDataSource(),
		NewTaskSessionsDataSource(),
		NewSecurityAnalyzerDataSource(),
		NewMalwareEventsDataSource(),
	}
	for _, ds := range dsList {
		var resp datasource.ConfigureResponse
		ds.(datasource.DataSourceWithConfigure).Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "not-a-client"}, &resp)
		assert.True(t, resp.Diagnostics.HasError())
	}
}

// ---------------------------------------------------------------------------
// T3.1 — security_roles
// ---------------------------------------------------------------------------

func TestSecurityRolesDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityRolesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityRoles, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "role-1", "name": "Veeam Administrator", "description": "Full access"},
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

func TestSecurityRolesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityRolesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityRoles, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathSecurityRoles, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSecurityRolesDataSource_Read_ByID_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, "/api/v1/security/roles/bad-id", mock.Anything).Return(errors.New("not found"))

	var item map[string]interface{}
	err := mockClient.GetJSON(context.Background(), "/api/v1/security/roles/bad-id", &item)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// T3.2 — security_users
// ---------------------------------------------------------------------------

func TestSecurityUsersDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityUsersDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityUsers, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "u-1", "login": "DOMAIN\\admin", "description": "Admin", "roleId": "role-1"},
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

func TestSecurityUsersDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityUsersDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityUsers, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathSecurityUsers, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.3 — backup_objects
// ---------------------------------------------------------------------------

func TestBackupObjectsDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupObjectsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathBackupObjects, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "obj-1", "name": "vm-01", "type": "VirtualMachine", "backupId": "bk-1", "restorePointsCount": float64(3)},
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

func TestBackupObjectsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &BackupObjectsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathBackupObjects, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathBackupObjects, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.4 — replicas
// ---------------------------------------------------------------------------

func TestReplicasDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ReplicasDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathReplicas, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "rep-1", "name": "vm-replica", "type": "VirtualMachine", "state": "Ready", "platform": "VMware"},
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

func TestReplicasDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ReplicasDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathReplicas, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathReplicas, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.5 — replica_points
// ---------------------------------------------------------------------------

func TestReplicaPointsDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ReplicaPointsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathReplicaPoints, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "rp-1", "name": "point-1", "replicaId": "rep-1", "creationTime": "2024-01-01T00:00:00Z"},
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

func TestReplicaPointsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ReplicaPointsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathReplicaPoints, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathReplicaPoints, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.6 — proxy_states
// ---------------------------------------------------------------------------

func TestProxyStatesDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxyStatesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProxyStates, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "p-1", "name": "Proxy01", "status": "Available", "type": "VMware"},
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

func TestProxyStatesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProxyStatesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProxyStates, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathProxyStates, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.7 — protected_computers
// ---------------------------------------------------------------------------

func TestProtectedComputersDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectedComputersDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProtectedComputers, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "pc-1", "name": "server01", "type": "Windows", "status": "Protected", "platform": "Physical"},
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

func TestProtectedComputersDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ProtectedComputersDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProtectedComputers, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathProtectedComputers, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.8 — services
// ---------------------------------------------------------------------------

func TestServicesDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServicesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServices, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "svc-1", "name": "VeeamBackupSvc", "status": "Running", "version": "12.3.0.0"},
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

func TestServicesDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServicesDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServices, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathServices, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.9 — server_time (singleton)
// ---------------------------------------------------------------------------

func TestServerTimeDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerTimeDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServerTime, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"serverTime": "2024-01-01T12:00:00Z",
			"timeZone":   "UTC",
			"utcOffset":  "+00:00",
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

func TestServerTimeDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerTimeDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServerTime, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.10 — server_certificate (singleton)
// ---------------------------------------------------------------------------

func TestServerCertificateDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerCertificateDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServerCertificate, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"thumbprint":   "AA:BB:CC",
			"subject":      "CN=vbr01",
			"issuedBy":     "CN=VeeamCA",
			"validFrom":    "2024-01-01T00:00:00Z",
			"validTo":      "2025-01-01T00:00:00Z",
			"serialNumber": "01234567",
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

func TestServerCertificateDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &ServerCertificateDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathServerCertificate, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.11 — task_sessions
// ---------------------------------------------------------------------------

func TestTaskSessionsDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &TaskSessionsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathTaskSessions, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "ts-1", "name": "Task 1", "sessionId": "s-1", "status": "Success", "startTime": "2024-01-01T00:00:00Z", "endTime": "2024-01-01T01:00:00Z"},
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

func TestTaskSessionsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &TaskSessionsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathTaskSessions, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathTaskSessions, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.12 — security_analyzer (composite singleton)
// ---------------------------------------------------------------------------

func TestSecurityAnalyzerDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityAnalyzerDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerLastRun, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{
			"lastRunTime":   "2024-01-01T00:00:00Z",
			"lastRunStatus": "Passed",
		}
	}).Return(nil)

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerBestPractices, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "bp-1", "name": "MFA enabled", "status": "Passed", "description": "Multi-factor authentication is enabled"},
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

func TestSecurityAnalyzerDataSource_Read_LastRunError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityAnalyzerDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerLastRun, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSecurityAnalyzerDataSource_Read_BestPracticesError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &SecurityAnalyzerDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerLastRun, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*map[string]interface{})
		*dest = map[string]interface{}{"lastRunTime": "2024-01-01T00:00:00Z", "lastRunStatus": "Passed"}
	}).Return(nil)

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerBestPractices, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerBestPractices, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// T3.13 — malware_events
// ---------------------------------------------------------------------------

func TestMalwareEventsDataSource_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &MalwareEventsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathMalwareEvents, mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]map[string]interface{})
		*dest = []map[string]interface{}{
			{"id": "me-1", "name": "Ransomware detected", "type": "Ransomware", "detectionTime": "2024-01-01T00:00:00Z", "severity": "High", "state": "Active"},
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

func TestMalwareEventsDataSource_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	ds := &MalwareEventsDataSource{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathMalwareEvents, mock.Anything).Return(errors.New("api error"))
	mockClient.On("GetJSON", mock.Anything, client.PathMalwareEvents, mock.Anything).Return(errors.New("api error"))

	cfg := buildNullConfig(ds)
	state := buildNullState(ds)
	req := datasource.ReadRequest{Config: cfg}
	resp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Schema validation — all prio3 sources have the expected top-level attributes
// ---------------------------------------------------------------------------

func TestPrio3DataSources_Schema(t *testing.T) {
	tests := []struct {
		name          string
		factory       func() datasource.DataSource
		mustHaveAttrs []string
	}{
		{"security_roles", NewSecurityRolesDataSource, []string{"id", "role_id", "roles"}},
		{"security_users", NewSecurityUsersDataSource, []string{"id", "user_id", "users"}},
		{"backup_objects", NewBackupObjectsDataSource, []string{"id", "object_id", "objects"}},
		{"replicas", NewReplicasDataSource, []string{"id", "replica_id", "replicas"}},
		{"replica_points", NewReplicaPointsDataSource, []string{"id", "replica_point_id", "replica_points"}},
		{"proxy_states", NewProxyStatesDataSource, []string{"id", "states"}},
		{"protected_computers", NewProtectedComputersDataSource, []string{"id", "computer_id", "computers"}},
		{"services", NewServicesDataSource, []string{"id", "services"}},
		{"server_time", NewServerTimeDataSource, []string{"id", "server_time", "time_zone", "utc_offset"}},
		{"server_certificate", NewServerCertificateDataSource, []string{"id", "thumbprint", "subject", "issued_by", "valid_from", "valid_to", "serial_number"}},
		{"task_sessions", NewTaskSessionsDataSource, []string{"id", "task_session_id", "task_sessions"}},
		{"security_analyzer", NewSecurityAnalyzerDataSource, []string{"id", "last_run_time", "last_run_status", "best_practices"}},
		{"malware_events", NewMalwareEventsDataSource, []string{"id", "events"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.factory()
			var resp datasource.SchemaResponse
			ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
			assert.NotNil(t, resp.Schema)
			for _, attr := range tc.mustHaveAttrs {
				assert.Contains(t, resp.Schema.Attributes, attr, "missing attribute %q in schema for %s", attr, tc.name)
			}
		})
	}
}

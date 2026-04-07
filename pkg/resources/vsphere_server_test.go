package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// Ensure MockVeeamClient (defined in backup_job_test.go) satisfies client.APIClient.

// ---------------------------------------------------------------------------
// buildSpec
// ---------------------------------------------------------------------------

func TestVSphereServerBuildSpec_DefaultPort(t *testing.T) {
	r := &VSphereServer{}
	data := &VSphereServerModel{
		Name:          types.StringValue("vcenter.example.com"),
		Description:   types.StringValue("Production vCenter"),
		CredentialsID: types.StringValue("cred-abc"),
		Port:          types.Int64Null(),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, models.ManagedServerTypeViHost, spec.Type, "type must always be ViHost")
	assert.Equal(t, "vcenter.example.com", spec.Name)
	assert.Equal(t, "cred-abc", spec.CredentialsID)
	assert.Equal(t, 0, spec.Port, "null port should leave Port at zero (API uses default 443)")
}

func TestVSphereServerBuildSpec_WithPortAndThumbprint(t *testing.T) {
	r := &VSphereServer{}
	data := &VSphereServerModel{
		Name:                  types.StringValue("esxi01.example.com"),
		Description:           types.StringValue("Standalone ESXi"),
		CredentialsID:         types.StringValue("cred-xyz"),
		Port:                  types.Int64Value(443),
		CertificateThumbprint: types.StringValue("AA:BB:CC:DD"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, models.ManagedServerTypeViHost, spec.Type)
	assert.Equal(t, 443, spec.Port)
	assert.Equal(t, "AA:BB:CC:DD", spec.CertificateThumbprint)
}

func TestVSphereServerBuildSpec_TypeAlwaysViHost(t *testing.T) {
	// Ensure type is hardcoded regardless of data contents — key invariant.
	r := &VSphereServer{}
	data := &VSphereServerModel{
		Name:          types.StringValue("host.example.com"),
		CredentialsID: types.StringValue("cred-1"),
		Port:          types.Int64Null(),
	}
	spec := r.buildSpec(data)
	assert.Equal(t, models.ManagedServerTypeViHost, spec.Type)
	assert.Equal(t, models.ManagedServerTypeViHost, spec.ManagedServerSpec.Type)
}

// ---------------------------------------------------------------------------
// syncFromAPI
// ---------------------------------------------------------------------------

func TestVSphereServerSyncFromAPI_PopulatesFields(t *testing.T) {
	r := &VSphereServer{}
	data := &VSphereServerModel{
		ID:            types.StringValue("srv-1"),
		CredentialsID: types.StringValue("cred-1"),
		Port:          types.Int64Null(),
	}
	api := &models.ManagedServerModel{
		Name:        "vcenter.example.com",
		Description: "Production vCenter",
		Status:      models.ManagedServerStatusAvailable,
	}

	r.syncFromAPI(data, api)

	assert.Equal(t, "vcenter.example.com", data.Name.ValueString())
	assert.Equal(t, "Production vCenter", data.Description.ValueString())
	assert.Equal(t, string(models.ManagedServerStatusAvailable), data.Status.ValueString())
}

func TestVSphereServerSyncFromAPI_DoesNotOverwriteWithEmpty(t *testing.T) {
	r := &VSphereServer{}
	data := &VSphereServerModel{
		Name:        types.StringValue("vcenter.example.com"),
		Description: types.StringValue("existing"),
		Status:      types.StringValue("Available"),
	}
	api := &models.ManagedServerModel{} // all empty

	r.syncFromAPI(data, api)

	// Existing values must be preserved when API returns empty strings.
	assert.Equal(t, "vcenter.example.com", data.Name.ValueString())
	assert.Equal(t, "existing", data.Description.ValueString())
	assert.Equal(t, "Available", data.Status.ValueString())
}

// ---------------------------------------------------------------------------
// isViHostAsyncResult
// ---------------------------------------------------------------------------

func TestIsViHostAsyncResult(t *testing.T) {
	tests := []struct {
		name   string
		result map[string]interface{}
		want   bool
	}{
		{"empty type → async", map[string]interface{}{"id": "sess-1"}, true},
		{"session type → async", map[string]interface{}{"id": "sess-2", "type": "Session"}, true},
		{"session type lowercase", map[string]interface{}{"id": "sess-3", "type": "session"}, true},
		{"vihost type → not async", map[string]interface{}{"id": "srv-1", "type": "ViHost"}, false},
		{"vihost type lowercase", map[string]interface{}{"id": "srv-2", "type": "vihost"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isViHostAsyncResult(tc.result)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// isViHostNotFound
// ---------------------------------------------------------------------------

func TestIsViHostNotFound(t *testing.T) {
	assert.False(t, isViHostNotFound(nil))
	assert.True(t, isViHostNotFound(errors.New("http 404 not found")))
	assert.True(t, isViHostNotFound(errors.New("notfound")))
	assert.True(t, isViHostNotFound(&models.APIError{ErrorCode: "NotFound"}))
	assert.False(t, isViHostNotFound(errors.New("http 500 internal server error")))
}

// ---------------------------------------------------------------------------
// Metadata / Schema
// ---------------------------------------------------------------------------

func TestVSphereServerMetadata(t *testing.T) {
	r := NewVSphereServer()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_vsphere_server", resp.TypeName)
}

func TestVSphereServerSchemaFields(t *testing.T) {
	r := NewVSphereServer()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"name", "credentials_id"}
	optional := []string{"description", "port", "certificate_thumbprint"}
	computed := []string{"id", "status"}

	for _, f := range required {
		a, ok := attrs[f]
		require.True(t, ok, "missing attribute: %s", f)
		assert.False(t, a.IsOptional(), "%s should not be optional", f)
	}
	for _, f := range optional {
		a, ok := attrs[f]
		require.True(t, ok, "missing attribute: %s", f)
		assert.True(t, a.IsOptional(), "%s should be optional", f)
	}
	for _, f := range computed {
		a, ok := attrs[f]
		require.True(t, ok, "missing attribute: %s", f)
		assert.True(t, a.IsComputed(), "%s should be computed", f)
	}

	// Confirm there is NO "type" field — the vSphere resource always uses ViHost.
	_, hasType := attrs["type"]
	assert.False(t, hasType, "veeam_vsphere_server must not expose a 'type' attribute")
}

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func TestVSphereServerConfigure_NilProviderData(t *testing.T) {
	r := &VSphereServer{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestVSphereServerConfigure_WrongType(t *testing.T) {
	r := &VSphereServer{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: "wrong"}, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func TestVSphereServerImportState(t *testing.T) {
	r := NewVSphereServer()
	result := importStateWithID(t, r.(resourceWithImportState), "srv-vsphere-001")
	require.False(t, result.Diagnostics.HasError())

	var data VSphereServerModel
	diags := result.State.Get(context.Background(), &data)
	require.False(t, diags.HasError())
	assert.Equal(t, "srv-vsphere-001", data.ID.ValueString())
}

// ---------------------------------------------------------------------------
// Create — error path
// ---------------------------------------------------------------------------

func TestVSphereServerCreate_APIError(t *testing.T) {
	mc := new(MockVeeamClient)
	mc.On("PostJSON", mock.Anything, "/api/v1/backupInfrastructure/managedServers",
		mock.Anything, mock.Anything).Return(errors.New("503 Service Unavailable"))

	r := &VSphereServer{client: mc}
	plan := buildNullResourcePlan(r)
	resp := &resource.CreateResponse{State: buildNullResourceState(r)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)
	assert.True(t, resp.Diagnostics.HasError())
	mc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Read — error path
// ---------------------------------------------------------------------------

func TestVSphereServerRead_APIError(t *testing.T) {
	mc := new(MockVeeamClient)
	mc.On("GetJSON", mock.Anything, mock.AnythingOfType("string"),
		mock.Anything).Return(errors.New("http 404 not found"))

	r := &VSphereServer{client: mc}
	state := buildNullResourceState(r)
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)
	assert.True(t, resp.Diagnostics.HasError())
	mc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Update — error path
// ---------------------------------------------------------------------------

func TestVSphereServerUpdate_APIError(t *testing.T) {
	mc := new(MockVeeamClient)
	mc.On("PutJSON", mock.Anything, mock.AnythingOfType("string"),
		mock.Anything, mock.Anything).Return(errors.New("forbidden"))

	r := &VSphereServer{client: mc}
	plan := buildNullResourcePlan(r)
	state := buildNullResourceState(r)
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, resp)
	assert.True(t, resp.Diagnostics.HasError())
	mc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete — error path
// ---------------------------------------------------------------------------

func TestVSphereServerDelete_APIError(t *testing.T) {
	mc := new(MockVeeamClient)
	mc.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("forbidden"))

	r := &VSphereServer{client: mc}
	state := buildNullResourceState(r)
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	assert.True(t, resp.Diagnostics.HasError())
	mc.AssertExpectations(t)
}

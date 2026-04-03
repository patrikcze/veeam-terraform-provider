package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// ---------------------------------------------------------------------------
// Metadata / Schema
// ---------------------------------------------------------------------------

func TestMountServer_Metadata(t *testing.T) {
	r := NewMountServer()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_mount_server", resp.TypeName)
}

func TestMountServer_Schema(t *testing.T) {
	r := &MountServer{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "managed_server_id")
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "credentials_id")
}

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func TestMountServer_Configure_Nil(t *testing.T) {
	r := &MountServer{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestMountServer_Configure_InvalidType(t *testing.T) {
	r := &MountServer{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestMountServer_Configure_Valid(t *testing.T) {
	r := &MountServer{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// buildSpec / syncModelFromAPI
// ---------------------------------------------------------------------------

func TestMountServer_BuildSpec(t *testing.T) {
	r := &MountServer{}
	data := &MountServerModel{
		Name:            types.StringValue("mount-srv-01"),
		Description:     types.StringValue("Primary mount server"),
		ManagedServerID: types.StringValue("ms-uuid-1"),
		Type:            types.StringValue("WinServer"),
		CredentialsID:   types.StringValue("cred-uuid-1"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "mount-srv-01", spec.Name)
	assert.Equal(t, "Primary mount server", spec.Description)
	assert.Equal(t, "ms-uuid-1", spec.ManagedServerID)
	assert.Equal(t, "WinServer", spec.Type)
	assert.Equal(t, "cred-uuid-1", spec.CredentialsID)
}

func TestMountServer_BuildSpec_OptionalFieldsOmitted(t *testing.T) {
	r := &MountServer{}
	data := &MountServerModel{
		Name:            types.StringValue("mount-minimal"),
		ManagedServerID: types.StringValue("ms-uuid-2"),
		Type:            types.StringValue("LinuxServer"),
		Description:     types.StringNull(),
		CredentialsID:   types.StringNull(),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "mount-minimal", spec.Name)
	assert.Equal(t, "", spec.Description)
	assert.Equal(t, "", spec.CredentialsID)
}

func TestMountServer_SyncModelFromAPI(t *testing.T) {
	r := &MountServer{}
	data := &MountServerModel{}
	api := &models.MountServerModel{
		ID:              "ms-1",
		Name:            "mount-srv-01",
		Description:     "desc",
		ManagedServerID: "ms-uuid-1",
		Type:            "WinServer",
		CredentialsID:   "cred-uuid-1",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "mount-srv-01", data.Name.ValueString())
	assert.Equal(t, "desc", data.Description.ValueString())
	assert.Equal(t, "ms-uuid-1", data.ManagedServerID.ValueString())
	assert.Equal(t, "WinServer", data.Type.ValueString())
	assert.Equal(t, "cred-uuid-1", data.CredentialsID.ValueString())
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestMountServer_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathMountServers, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.MountServerModel)
			result.ID = "ms-1"
			result.Name = "mount-srv-01"
			result.ManagedServerID = "ms-uuid-1"
			result.Type = "WinServer"
		}).Return(nil)

	var result models.MountServerModel
	err := mockClient.PostJSON(context.Background(), client.PathMountServers, &models.MountServerSpec{
		Name:            "mount-srv-01",
		ManagedServerID: "ms-uuid-1",
		Type:            "WinServer",
	}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "ms-1", result.ID)
	mockClient.AssertExpectations(t)
}

func TestMountServer_Create_PostFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathMountServers, mock.Anything, mock.Anything).
		Return(fmt.Errorf("API error: 500 internal server error"))

	var result models.MountServerModel
	err := mockClient.PostJSON(context.Background(), client.PathMountServers, &models.MountServerSpec{}, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func TestMountServer_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "ms-42"
	endpoint := fmt.Sprintf(client.PathMountServerByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.MountServerModel)
			result.ID = id
			result.Name = "mount-srv-42"
			result.ManagedServerID = "ms-uuid-1"
			result.Type = "WinServer"
		}).Return(nil)

	var result models.MountServerModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "mount-srv-42", result.Name)
	mockClient.AssertExpectations(t)
}

func TestMountServer_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "ms-missing"
	endpoint := fmt.Sprintf(client.PathMountServerByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("API request failed with HTTP 404: not found"))

	var result models.MountServerModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestMountServer_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "ms-1"
	endpoint := fmt.Sprintf(client.PathMountServerByID, id)

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).
		Return(nil)

	err := mockClient.PutJSON(context.Background(), endpoint, &models.MountServerSpec{
		Name:            "updated-mount",
		ManagedServerID: "ms-uuid-1",
		Type:            "WinServer",
	}, nil)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete (no-op)
// ---------------------------------------------------------------------------

func TestMountServer_Delete_NoOp(t *testing.T) {
	r := &MountServer{}
	// Delete is a no-op — it must not call any client methods or add diagnostics.
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func TestMountServer_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &MountServer{}
}

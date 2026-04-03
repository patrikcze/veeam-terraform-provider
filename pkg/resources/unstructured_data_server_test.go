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

func TestUnstructuredDataServer_Metadata(t *testing.T) {
	r := NewUnstructuredDataServer()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_unstructured_data_server", resp.TypeName)
}

func TestUnstructuredDataServer_Schema(t *testing.T) {
	r := &UnstructuredDataServer{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "host_name")
	assert.Contains(t, resp.Schema.Attributes, "credentials_id")
	assert.Contains(t, resp.Schema.Attributes, "access_credentials_id")
}

func TestUnstructuredDataServer_Configure_Nil(t *testing.T) {
	r := &UnstructuredDataServer{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestUnstructuredDataServer_Configure_InvalidType(t *testing.T) {
	r := &UnstructuredDataServer{}
	req := resource.ConfigureRequest{ProviderData: true}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestUnstructuredDataServer_BuildSpec(t *testing.T) {
	r := &UnstructuredDataServer{}
	data := &UnstructuredDataServerModel{
		Name:                types.StringValue("nas-01"),
		Description:         types.StringValue("Primary NAS"),
		Type:                types.StringValue("CifsShare"),
		HostName:            types.StringValue("nas.example.com"),
		CredentialsID:       types.StringValue("cred-1"),
		AccessCredentialsID: types.StringValue("acred-1"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "nas-01", spec.Name)
	assert.Equal(t, "Primary NAS", spec.Description)
	assert.Equal(t, "CifsShare", spec.Type)
	assert.Equal(t, "nas.example.com", spec.HostName)
	assert.Equal(t, "cred-1", spec.CredentialsID)
	assert.Equal(t, "acred-1", spec.AccessCredentialsID)
}

func TestUnstructuredDataServer_BuildSpec_OptionalOmitted(t *testing.T) {
	r := &UnstructuredDataServer{}
	data := &UnstructuredDataServerModel{
		Name:                types.StringValue("nas-minimal"),
		Type:                types.StringValue("NfsShare"),
		HostName:            types.StringValue("nfs.example.com"),
		Description:         types.StringNull(),
		CredentialsID:       types.StringNull(),
		AccessCredentialsID: types.StringNull(),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "", spec.Description)
	assert.Equal(t, "", spec.CredentialsID)
	assert.Equal(t, "", spec.AccessCredentialsID)
}

func TestUnstructuredDataServer_SyncModelFromAPI(t *testing.T) {
	r := &UnstructuredDataServer{}
	data := &UnstructuredDataServerModel{}
	api := &models.UnstructuredDataServerModel{
		ID:                  "uds-1",
		Name:                "nas-01",
		Description:         "Primary NAS",
		Type:                "CifsShare",
		HostName:            "nas.example.com",
		CredentialsID:       "cred-1",
		AccessCredentialsID: "acred-1",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "nas-01", data.Name.ValueString())
	assert.Equal(t, "CifsShare", data.Type.ValueString())
	assert.Equal(t, "nas.example.com", data.HostName.ValueString())
	assert.Equal(t, "cred-1", data.CredentialsID.ValueString())
	assert.Equal(t, "acred-1", data.AccessCredentialsID.ValueString())
}

func TestUnstructuredDataServer_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathUnstructuredDataServers, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.UnstructuredDataServerModel)
			result.ID = "uds-1"
			result.Name = "nas-01"
			result.Type = "CifsShare"
			result.HostName = "nas.example.com"
		}).Return(nil)

	var result models.UnstructuredDataServerModel
	err := mockClient.PostJSON(context.Background(), client.PathUnstructuredDataServers,
		&models.UnstructuredDataServerSpec{Name: "nas-01", Type: "CifsShare", HostName: "nas.example.com"}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "uds-1", result.ID)
	mockClient.AssertExpectations(t)
}

func TestUnstructuredDataServer_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "uds-42"
	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.UnstructuredDataServerModel)
			result.ID = id
			result.Name = "nas-42"
		}).Return(nil)

	var result models.UnstructuredDataServerModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	mockClient.AssertExpectations(t)
}

func TestUnstructuredDataServer_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, "missing")

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("HTTP 500: internal server error"))

	var result models.UnstructuredDataServerModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestUnstructuredDataServer_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, "uds-1")

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).Return(nil)

	err := mockClient.PutJSON(context.Background(), endpoint,
		&models.UnstructuredDataServerSpec{Name: "nas-updated", Type: "CifsShare", HostName: "nas.example.com"}, nil)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestUnstructuredDataServer_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, "uds-1")

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestUnstructuredDataServer_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &UnstructuredDataServer{}
}

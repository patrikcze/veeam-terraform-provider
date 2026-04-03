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

func TestGlobalVMExclusion_Metadata(t *testing.T) {
	r := NewGlobalVMExclusion()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_global_vm_exclusion", resp.TypeName)
}

func TestGlobalVMExclusion_Schema(t *testing.T) {
	r := &GlobalVMExclusion{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "type")
	assert.Contains(t, resp.Schema.Attributes, "host_name")
	assert.Contains(t, resp.Schema.Attributes, "object_id")
}

func TestGlobalVMExclusion_Configure_Nil(t *testing.T) {
	r := &GlobalVMExclusion{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestGlobalVMExclusion_Configure_InvalidType(t *testing.T) {
	r := &GlobalVMExclusion{}
	req := resource.ConfigureRequest{ProviderData: 42}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestGlobalVMExclusion_BuildSpec(t *testing.T) {
	r := &GlobalVMExclusion{}
	data := &GlobalVMExclusionModel{
		Name:        types.StringValue("vm-to-exclude"),
		Type:        types.StringValue("VirtualMachine"),
		HostName:    types.StringValue("vcenter.example.com"),
		ObjectID:    types.StringValue("vm-42"),
		Description: types.StringValue("test vm"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "vm-to-exclude", spec.Name)
	assert.Equal(t, "VirtualMachine", spec.Type)
	assert.Equal(t, "vcenter.example.com", spec.HostName)
	assert.Equal(t, "vm-42", spec.ObjectID)
	assert.Equal(t, "test vm", spec.Description)
}

func TestGlobalVMExclusion_BuildSpec_OptionalOmitted(t *testing.T) {
	r := &GlobalVMExclusion{}
	data := &GlobalVMExclusionModel{
		Name:        types.StringValue("minimal-vm"),
		Type:        types.StringValue("Folder"),
		HostName:    types.StringNull(),
		ObjectID:    types.StringNull(),
		Description: types.StringNull(),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "", spec.HostName)
	assert.Equal(t, "", spec.ObjectID)
	assert.Equal(t, "", spec.Description)
}

func TestGlobalVMExclusion_SyncModelFromAPI(t *testing.T) {
	r := &GlobalVMExclusion{}
	data := &GlobalVMExclusionModel{}
	api := &models.GlobalVMExclusionModel{
		ID:          "excl-1",
		Name:        "vm-to-exclude",
		Type:        "VirtualMachine",
		HostName:    "vcenter.example.com",
		ObjectID:    "vm-42",
		Description: "test vm",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "vm-to-exclude", data.Name.ValueString())
	assert.Equal(t, "VirtualMachine", data.Type.ValueString())
	assert.Equal(t, "vcenter.example.com", data.HostName.ValueString())
	assert.Equal(t, "vm-42", data.ObjectID.ValueString())
}

func TestGlobalVMExclusion_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathGlobalVMExclusions, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.GlobalVMExclusionModel)
			result.ID = "excl-1"
			result.Name = "vm-to-exclude"
			result.Type = "VirtualMachine"
		}).Return(nil)

	var result models.GlobalVMExclusionModel
	err := mockClient.PostJSON(context.Background(), client.PathGlobalVMExclusions,
		&models.GlobalVMExclusionSpec{Name: "vm-to-exclude", Type: "VirtualMachine"}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "excl-1", result.ID)
	mockClient.AssertExpectations(t)
}

func TestGlobalVMExclusion_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "excl-42"
	endpoint := fmt.Sprintf(client.PathGlobalVMExclusionByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.GlobalVMExclusionModel)
			result.ID = id
			result.Name = "my-vm"
			result.Type = "VirtualMachine"
		}).Return(nil)

	var result models.GlobalVMExclusionModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	mockClient.AssertExpectations(t)
}

func TestGlobalVMExclusion_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathGlobalVMExclusionByID, "missing")

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("HTTP 404: not found"))

	var result models.GlobalVMExclusionModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGlobalVMExclusion_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathGlobalVMExclusionByID, "excl-1")

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestGlobalVMExclusion_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathGlobalVMExclusionByID, "excl-1")

	mockClient.On("DeleteJSON", mock.Anything, endpoint).
		Return(fmt.Errorf("HTTP 500: internal server error"))

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGlobalVMExclusion_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &GlobalVMExclusion{}
}

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

func TestEntraIDTenant_Metadata(t *testing.T) {
	r := NewEntraIDTenant()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_entra_id_tenant", resp.TypeName)
}

func TestEntraIDTenant_Schema(t *testing.T) {
	r := &EntraIDTenant{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "tenant_id")
	assert.Contains(t, resp.Schema.Attributes, "credentials_id")
}

func TestEntraIDTenant_Configure_Nil(t *testing.T) {
	r := &EntraIDTenant{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestEntraIDTenant_Configure_InvalidType(t *testing.T) {
	r := &EntraIDTenant{}
	req := resource.ConfigureRequest{ProviderData: 123}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestEntraIDTenant_BuildSpec(t *testing.T) {
	r := &EntraIDTenant{}
	data := &EntraIDTenantModel{
		Name:          types.StringValue("My Tenant"),
		Description:   types.StringValue("Primary Azure tenant"),
		TenantID:      types.StringValue("aaaa-bbbb-cccc"),
		CredentialsID: types.StringValue("cred-uuid-1"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "My Tenant", spec.Name)
	assert.Equal(t, "Primary Azure tenant", spec.Description)
	assert.Equal(t, "aaaa-bbbb-cccc", spec.TenantID)
	assert.Equal(t, "cred-uuid-1", spec.CredentialsID)
}

func TestEntraIDTenant_SyncModelFromAPI(t *testing.T) {
	r := &EntraIDTenant{}
	data := &EntraIDTenantModel{}
	api := &models.EntraIDTenantModel{
		ID:            "tenant-1",
		Name:          "My Tenant",
		Description:   "desc",
		TenantID:      "aaaa-bbbb-cccc",
		CredentialsID: "cred-uuid-1",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "My Tenant", data.Name.ValueString())
	assert.Equal(t, "aaaa-bbbb-cccc", data.TenantID.ValueString())
	assert.Equal(t, "cred-uuid-1", data.CredentialsID.ValueString())
}

func TestEntraIDTenant_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathEntraIDTenants, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.EntraIDTenantModel)
			result.ID = "tenant-1"
			result.Name = "My Tenant"
			result.TenantID = "aaaa-bbbb-cccc"
			result.CredentialsID = "cred-uuid-1"
		}).Return(nil)

	var result models.EntraIDTenantModel
	err := mockClient.PostJSON(context.Background(), client.PathEntraIDTenants,
		&models.EntraIDTenantSpec{Name: "My Tenant", TenantID: "aaaa-bbbb-cccc", CredentialsID: "cred-uuid-1"}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "tenant-1", result.ID)
	mockClient.AssertExpectations(t)
}

func TestEntraIDTenant_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "tenant-42"
	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.EntraIDTenantModel)
			result.ID = id
			result.Name = "My Tenant"
		}).Return(nil)

	var result models.EntraIDTenantModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	mockClient.AssertExpectations(t)
}

func TestEntraIDTenant_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, "missing")

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("HTTP 404: not found"))

	var result models.EntraIDTenantModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestEntraIDTenant_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, "tenant-1")

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).Return(nil)

	err := mockClient.PutJSON(context.Background(), endpoint,
		&models.EntraIDTenantSpec{Name: "Updated Tenant", TenantID: "aaaa-bbbb-cccc", CredentialsID: "cred-2"}, nil)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestEntraIDTenant_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, "tenant-1")

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestEntraIDTenant_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &EntraIDTenant{}
}

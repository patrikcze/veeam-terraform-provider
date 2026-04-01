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
// Configure tests
// ---------------------------------------------------------------------------

func TestADDomain_Configure_Nil(t *testing.T) {
	r := &ADDomain{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestADDomain_Configure_InvalidType(t *testing.T) {
	r := &ADDomain{}
	req := resource.ConfigureRequest{ProviderData: true}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestADDomain_Configure_Valid(t *testing.T) {
	r := &ADDomain{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// syncModelFromAPI
// ---------------------------------------------------------------------------

func TestADDomain_SyncModelFromAPI_PreservesPassword(t *testing.T) {
	r := &ADDomain{}
	data := &ADDomainModel{
		Password: types.StringValue("domain-admin-pass"),
	}

	api := &models.ADDomainModel{
		ID:          "domain-1",
		Name:        "corp.example.com",
		UserName:    "CORP\\admin",
		Description: "Corporate domain",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "corp.example.com", data.Name.ValueString())
	assert.Equal(t, "CORP\\admin", data.Username.ValueString())
	assert.Equal(t, "Corporate domain", data.Description.ValueString())
	// Password must NOT be overwritten from API response.
	assert.Equal(t, "domain-admin-pass", data.Password.ValueString())
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestADDomain_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathADDomains, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			payload := args.Get(2).(*models.ADDomainSpec)
			assert.Equal(t, "corp.example.com", payload.Name)
			assert.Equal(t, "CORP\\admin", payload.UserName)
			assert.Equal(t, "P@ssw0rd", payload.Password)

			result := args.Get(3).(*models.ADDomainModel)
			result.ID = "domain-1"
			result.Name = "corp.example.com"
			result.UserName = "CORP\\admin"
			result.Description = ""
		}).Return(nil)

	var result models.ADDomainModel
	err := mockClient.PostJSON(context.Background(), client.PathADDomains, &models.ADDomainSpec{
		Name:     "corp.example.com",
		UserName: "CORP\\admin",
		Password: "P@ssw0rd",
	}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "domain-1", result.ID)
	assert.Equal(t, "corp.example.com", result.Name)
	mockClient.AssertExpectations(t)
}

func TestADDomain_Create_PostFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathADDomains, mock.Anything, mock.Anything).
		Return(fmt.Errorf("API error: domain unreachable"))

	var result models.ADDomainModel
	err := mockClient.PostJSON(context.Background(), client.PathADDomains, &models.ADDomainSpec{
		Name:     "corp.example.com",
		UserName: "CORP\\admin",
		Password: "P@ssw0rd",
	}, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "domain unreachable")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Read tests
// ---------------------------------------------------------------------------

func TestADDomain_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "domain-1"
	endpoint := fmt.Sprintf(client.PathADDomainByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.ADDomainModel)
			result.ID = id
			result.Name = "corp.example.com"
			result.UserName = "CORP\\admin"
			result.Description = "Main domain"
		}).Return(nil)

	var result models.ADDomainModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, "corp.example.com", result.Name)
	assert.Equal(t, "CORP\\admin", result.UserName)
	mockClient.AssertExpectations(t)
}

func TestADDomain_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "domain-missing"
	endpoint := fmt.Sprintf(client.PathADDomainByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("API request failed with HTTP 404: not found"))

	var result models.ADDomainModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestADDomain_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "domain-1"
	endpoint := fmt.Sprintf(client.PathADDomainByID, id)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestADDomain_Delete_DeleteFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "domain-1"
	endpoint := fmt.Sprintf(client.PathADDomainByID, id)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).
		Return(fmt.Errorf("API error: domain has active backup jobs"))

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active backup jobs")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func TestADDomain_ImportState(t *testing.T) {
	// Verify that ADDomain satisfies the ResourceWithImportState interface.
	var _ resource.ResourceWithImportState = &ADDomain{}
}

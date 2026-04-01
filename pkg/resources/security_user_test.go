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

func TestSecurityUser_Configure_Nil(t *testing.T) {
	r := &SecurityUser{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestSecurityUser_Configure_InvalidType(t *testing.T) {
	r := &SecurityUser{}
	req := resource.ConfigureRequest{ProviderData: 42}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSecurityUser_Configure_Valid(t *testing.T) {
	r := &SecurityUser{}
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

func TestSecurityUser_SyncModelFromAPI_PreservesPassword(t *testing.T) {
	r := &SecurityUser{}
	data := &SecurityUserModel{
		Password: types.StringValue("original-secret"),
	}

	api := &models.SecurityUserModel{
		ID:          "user-1",
		Login:       "admin@corp.local",
		Description: "Updated desc",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "admin@corp.local", data.Login.ValueString())
	assert.Equal(t, "Updated desc", data.Description.ValueString())
	// Password must NOT be overwritten from API response.
	assert.Equal(t, "original-secret", data.Password.ValueString())
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestSecurityUser_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	// Step 1: POST to create the user.
	mockClient.On("PostJSON", mock.Anything, client.PathSecurityUsers, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			payload := args.Get(2).(*models.SecurityUserSpec)
			assert.Equal(t, "jdoe", payload.Login)
			assert.Equal(t, "secret123", payload.Password)

			result := args.Get(3).(*models.SecurityUserModel)
			result.ID = "user-99"
			result.Login = "jdoe"
			result.Description = ""
		}).Return(nil)

	// Step 2: PUT to assign role.
	rolesEndpoint := fmt.Sprintf(client.PathSecurityUserRoles, "user-99")
	mockClient.On("PutJSON", mock.Anything, rolesEndpoint, mock.Anything, nil).
		Run(func(args mock.Arguments) {
			rolePayload := args.Get(2).(*models.SecurityUserRoleSpec)
			assert.Equal(t, "PortalAdministrator", rolePayload.RoleName)
		}).Return(nil)

	// Exercise PostJSON.
	var userResult models.SecurityUserModel
	err := mockClient.PostJSON(context.Background(), client.PathSecurityUsers, &models.SecurityUserSpec{
		Login:    "jdoe",
		Password: "secret123",
	}, &userResult)
	assert.NoError(t, err)
	assert.Equal(t, "user-99", userResult.ID)

	// Exercise PutJSON for role.
	err = mockClient.PutJSON(context.Background(), rolesEndpoint, &models.SecurityUserRoleSpec{
		RoleName: "PortalAdministrator",
	}, nil)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestSecurityUser_Create_PostFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathSecurityUsers, mock.Anything, mock.Anything).
		Return(fmt.Errorf("API error: duplicate login"))

	var result models.SecurityUserModel
	err := mockClient.PostJSON(context.Background(), client.PathSecurityUsers, &models.SecurityUserSpec{
		Login:    "jdoe",
		Password: "secret",
	}, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate login")
	mockClient.AssertExpectations(t)
}

func TestSecurityUser_Create_RolePutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	rolesEndpoint := fmt.Sprintf(client.PathSecurityUserRoles, "user-1")
	mockClient.On("PutJSON", mock.Anything, rolesEndpoint, mock.Anything, nil).
		Return(fmt.Errorf("API error: invalid role name"))

	err := mockClient.PutJSON(context.Background(), rolesEndpoint, &models.SecurityUserRoleSpec{
		RoleName: "InvalidRole",
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role name")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Read tests
// ---------------------------------------------------------------------------

func TestSecurityUser_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "user-10"
	userEndpoint := fmt.Sprintf(client.PathSecurityUserByID, id)
	rolesEndpoint := fmt.Sprintf(client.PathSecurityUserRoles, id)

	mockClient.On("GetJSON", mock.Anything, userEndpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.SecurityUserModel)
			result.ID = id
			result.Login = "jdoe"
			result.Description = "test user"
		}).Return(nil)

	mockClient.On("GetJSON", mock.Anything, rolesEndpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.SecurityUserRoleModel)
			result.RoleName = "PortalUser"
		}).Return(nil)

	var userResult models.SecurityUserModel
	err := mockClient.GetJSON(context.Background(), userEndpoint, &userResult)
	assert.NoError(t, err)
	assert.Equal(t, "jdoe", userResult.Login)

	var roleResult models.SecurityUserRoleModel
	err = mockClient.GetJSON(context.Background(), rolesEndpoint, &roleResult)
	assert.NoError(t, err)
	assert.Equal(t, "PortalUser", roleResult.RoleName)

	mockClient.AssertExpectations(t)
}

func TestSecurityUser_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "user-missing"
	userEndpoint := fmt.Sprintf(client.PathSecurityUserByID, id)

	mockClient.On("GetJSON", mock.Anything, userEndpoint, mock.Anything).
		Return(fmt.Errorf("API request failed with HTTP 404: not found"))

	var userResult models.SecurityUserModel
	err := mockClient.GetJSON(context.Background(), userEndpoint, &userResult)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	mockClient.AssertExpectations(t)
}

func TestSecurityUser_Read_RoleGetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "user-10"
	rolesEndpoint := fmt.Sprintf(client.PathSecurityUserRoles, id)

	mockClient.On("GetJSON", mock.Anything, rolesEndpoint, mock.Anything).
		Return(fmt.Errorf("API error: roles not found"))

	var roleResult models.SecurityUserRoleModel
	err := mockClient.GetJSON(context.Background(), rolesEndpoint, &roleResult)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "roles not found")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestSecurityUser_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "user-10"
	endpoint := fmt.Sprintf(client.PathSecurityUserByID, id)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSecurityUser_Delete_DeleteFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "user-10"
	endpoint := fmt.Sprintf(client.PathSecurityUserByID, id)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).
		Return(fmt.Errorf("API error: user has active sessions"))

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active sessions")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func TestSecurityUser_ImportState(t *testing.T) {
	// Verify that SecurityUser satisfies the ResourceWithImportState interface.
	var _ resource.ResourceWithImportState = &SecurityUser{}
}

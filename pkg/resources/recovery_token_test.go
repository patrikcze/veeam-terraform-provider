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

func TestRecoveryToken_Metadata(t *testing.T) {
	r := NewRecoveryToken()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_recovery_token", resp.TypeName)
}

func TestRecoveryToken_Schema(t *testing.T) {
	r := &RecoveryToken{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "managed_server_id")
	assert.Contains(t, resp.Schema.Attributes, "token_value")
}

func TestRecoveryToken_Schema_TokenValueSensitive(t *testing.T) {
	r := &RecoveryToken{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	attr, ok := resp.Schema.Attributes["token_value"]
	assert.True(t, ok)
	strAttr, ok := attr.(interface{ IsSensitive() bool })
	if ok {
		assert.True(t, strAttr.IsSensitive())
	}
}

func TestRecoveryToken_Configure_Nil(t *testing.T) {
	r := &RecoveryToken{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestRecoveryToken_Configure_InvalidType(t *testing.T) {
	r := &RecoveryToken{}
	req := resource.ConfigureRequest{ProviderData: "bad-type"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestRecoveryToken_BuildSpec(t *testing.T) {
	r := &RecoveryToken{}
	data := &RecoveryTokenModel{
		Name:            types.StringValue("agent-token"),
		Description:     types.StringValue("For agent-01"),
		ManagedServerID: types.StringValue("srv-uuid-1"),
		TokenValue:      types.StringValue("supersecret"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "agent-token", spec.Name)
	assert.Equal(t, "For agent-01", spec.Description)
	assert.Equal(t, "srv-uuid-1", spec.ManagedServerID)
}

func TestRecoveryToken_SyncModelFromAPI(t *testing.T) {
	r := &RecoveryToken{}
	data := &RecoveryTokenModel{
		TokenValue: types.StringValue("preserved-secret"),
	}
	api := &models.RecoveryTokenModel{
		ID:              "tok-1",
		Name:            "agent-token",
		Description:     "updated desc",
		ManagedServerID: "srv-uuid-1",
		// TokenValue intentionally empty — simulates API not returning it on read
		TokenValue: "",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "agent-token", data.Name.ValueString())
	assert.Equal(t, "updated desc", data.Description.ValueString())
	assert.Equal(t, "srv-uuid-1", data.ManagedServerID.ValueString())
	// token_value must NOT be overwritten
	assert.Equal(t, "preserved-secret", data.TokenValue.ValueString())
}

func TestRecoveryToken_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathRecoveryTokens, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.RecoveryTokenModel)
			result.ID = "tok-1"
			result.Name = "agent-token"
			result.ManagedServerID = "srv-uuid-1"
			result.TokenValue = "my-secret-token"
		}).Return(nil)

	var result models.RecoveryTokenModel
	err := mockClient.PostJSON(context.Background(), client.PathRecoveryTokens,
		&models.RecoveryTokenSpec{Name: "agent-token", ManagedServerID: "srv-uuid-1"}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "tok-1", result.ID)
	assert.Equal(t, "my-secret-token", result.TokenValue)
	mockClient.AssertExpectations(t)
}

func TestRecoveryToken_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	id := "tok-42"
	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.RecoveryTokenModel)
			result.ID = id
			result.Name = "agent-token"
			result.ManagedServerID = "srv-uuid-1"
			// TokenValue is empty — API does not return it on reads
		}).Return(nil)

	var result models.RecoveryTokenModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "", result.TokenValue)
	mockClient.AssertExpectations(t)
}

func TestRecoveryToken_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, "missing")

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("HTTP 404: not found"))

	var result models.RecoveryTokenModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestRecoveryToken_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, "tok-1")

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).Return(nil)

	err := mockClient.PutJSON(context.Background(), endpoint,
		&models.RecoveryTokenSpec{Name: "updated-token", ManagedServerID: "srv-uuid-1"}, nil)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRecoveryToken_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, "tok-1")

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRecoveryToken_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &RecoveryToken{}
}

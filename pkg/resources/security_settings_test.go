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
)

// ---------------------------------------------------------------------------
// SecuritySettings — Configure tests
// ---------------------------------------------------------------------------

func TestSecuritySettings_Configure_Nil(t *testing.T) {
	r := &SecuritySettings{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestSecuritySettings_Configure_InvalidType(t *testing.T) {
	r := &SecuritySettings{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSecuritySettings_Configure_Valid(t *testing.T) {
	r := &SecuritySettings{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// SecuritySettings — Create / putSecuritySettings
// ---------------------------------------------------------------------------

func TestSecuritySettings_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &SecuritySettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecuritySettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"requireSsl":                false,
			"requireMfa":                false,
			"blockFirstLogin":           false,
			"loginAttemptLimit":         float64(5),
			"inactivityTimeoutMin":      float64(30),
			"passwordExpirationDays":    float64(90),
			"passwordExpirationEnabled": false,
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathSecuritySettings, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, true, payload["requireSsl"])
		assert.Equal(t, true, payload["requireMfa"])
		assert.Equal(t, true, payload["blockFirstLogin"])
		assert.Equal(t, 3, payload["loginAttemptLimit"])
		assert.Equal(t, 15, payload["inactivityTimeoutMin"])
		assert.Equal(t, 60, payload["passwordExpirationDays"])
		assert.Equal(t, true, payload["passwordExpirationEnabled"])
	}).Return(nil)

	data := &SecuritySettingsModel{
		RequireSSL:                types.BoolValue(true),
		RequireMFA:                types.BoolValue(true),
		BlockFirstLogin:           types.BoolValue(true),
		LoginAttemptLimit:         types.Int64Value(3),
		InactivityTimeoutMin:      types.Int64Value(15),
		PasswordExpirationDays:    types.Int64Value(60),
		PasswordExpirationEnabled: types.BoolValue(true),
	}

	err := r.putSecuritySettings(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSecuritySettings_Create_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &SecuritySettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecuritySettings, mock.Anything).
		Return(fmt.Errorf("connection refused"))

	err := r.putSecuritySettings(context.Background(), &SecuritySettingsModel{
		RequireSSL: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading current security settings")
	mockClient.AssertExpectations(t)
}

func TestSecuritySettings_Create_PutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &SecuritySettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecuritySettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathSecuritySettings, mock.Anything, nil).
		Return(fmt.Errorf("server error"))

	err := r.putSecuritySettings(context.Background(), &SecuritySettingsModel{
		RequireSSL: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// SecuritySettings — Read / syncSecuritySettingsFromAPI
// ---------------------------------------------------------------------------

func TestSecuritySettings_Read_Success(t *testing.T) {
	raw := map[string]interface{}{
		"requireSsl":                true,
		"requireMfa":                false,
		"blockFirstLogin":           true,
		"loginAttemptLimit":         float64(3),
		"inactivityTimeoutMin":      float64(20),
		"passwordExpirationDays":    float64(45),
		"passwordExpirationEnabled": true,
	}

	data := &SecuritySettingsModel{}
	syncSecuritySettingsFromAPI(data, raw)

	assert.Equal(t, true, data.RequireSSL.ValueBool())
	assert.Equal(t, false, data.RequireMFA.ValueBool())
	assert.Equal(t, true, data.BlockFirstLogin.ValueBool())
	assert.Equal(t, int64(3), data.LoginAttemptLimit.ValueInt64())
	assert.Equal(t, int64(20), data.InactivityTimeoutMin.ValueInt64())
	assert.Equal(t, int64(45), data.PasswordExpirationDays.ValueInt64())
	assert.Equal(t, true, data.PasswordExpirationEnabled.ValueBool())
}

func TestSecuritySettings_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathSecuritySettings, mock.Anything).
		Return(fmt.Errorf("unauthorized"))

	var raw map[string]interface{}
	err := mockClient.GetJSON(context.Background(), client.PathSecuritySettings, &raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// SecuritySettings — Update
// ---------------------------------------------------------------------------

func TestSecuritySettings_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &SecuritySettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathSecuritySettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"requireSsl": true,
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathSecuritySettings, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, false, payload["requireSsl"])
		assert.Equal(t, 10, payload["loginAttemptLimit"])
	}).Return(nil)

	err := r.putSecuritySettings(context.Background(), &SecuritySettingsModel{
		RequireSSL:        types.BoolValue(false),
		LoginAttemptLimit: types.Int64Value(10),
	})
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// SecuritySettings — Delete (no-op)
// ---------------------------------------------------------------------------

func TestSecuritySettings_Delete_Success(t *testing.T) {
	r := &SecuritySettings{}
	r.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
}

// ---------------------------------------------------------------------------
// SecuritySettings — ImportState
// ---------------------------------------------------------------------------

func TestSecuritySettings_ImportState(t *testing.T) {
	importStateWithID(t, &SecuritySettings{}, "security-settings")
}

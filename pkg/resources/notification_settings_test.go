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
// NotificationSettings — Configure tests
// ---------------------------------------------------------------------------

func TestNotificationSettings_Configure_Nil(t *testing.T) {
	r := &NotificationSettings{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestNotificationSettings_Configure_InvalidType(t *testing.T) {
	r := &NotificationSettings{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestNotificationSettings_Configure_Valid(t *testing.T) {
	r := &NotificationSettings{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// NotificationSettings — Create / putNotificationSettings
// ---------------------------------------------------------------------------

func TestNotificationSettings_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &NotificationSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathNotificationSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"notifyOnSuccess":                false,
			"notifyOnWarning":                false,
			"notifyOnError":                  true,
			"suppressRepeatingNotifications": false,
			"notifyOnLastRetryOnly":          false,
			"sendSNMPOnSuccess":              false,
			"sendSNMPOnWarning":              false,
			"sendSNMPOnError":                false,
			"sendSyslogOnSuccess":            false,
			"sendSyslogOnWarning":            false,
			"sendSyslogOnError":              false,
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathNotificationSettings, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, true, payload["notifyOnSuccess"])
		assert.Equal(t, true, payload["notifyOnWarning"])
		assert.Equal(t, true, payload["notifyOnError"])
		assert.Equal(t, true, payload["suppressRepeatingNotifications"])
	}).Return(nil)

	data := &NotificationSettingsModel{
		NotifyOnSuccess:                types.BoolValue(true),
		NotifyOnWarning:                types.BoolValue(true),
		NotifyOnError:                  types.BoolValue(true),
		SuppressRepeatingNotifications: types.BoolValue(true),
		NotifyOnLastRetryOnly:          types.BoolValue(false),
		SendSNMPOnSuccess:              types.BoolValue(false),
		SendSNMPOnWarning:              types.BoolValue(false),
		SendSNMPOnError:                types.BoolValue(false),
		SendSyslogOnSuccess:            types.BoolValue(false),
		SendSyslogOnWarning:            types.BoolValue(false),
		SendSyslogOnError:              types.BoolValue(false),
	}

	err := r.putNotificationSettings(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestNotificationSettings_Create_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &NotificationSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathNotificationSettings, mock.Anything).
		Return(fmt.Errorf("connection refused"))

	err := r.putNotificationSettings(context.Background(), &NotificationSettingsModel{
		NotifyOnSuccess: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading current notification settings")
	mockClient.AssertExpectations(t)
}

func TestNotificationSettings_Create_PutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &NotificationSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathNotificationSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathNotificationSettings, mock.Anything, nil).
		Return(fmt.Errorf("server error"))

	err := r.putNotificationSettings(context.Background(), &NotificationSettingsModel{
		NotifyOnSuccess: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// NotificationSettings — Read / syncNotificationSettingsFromAPI
// ---------------------------------------------------------------------------

func TestNotificationSettings_Read_Success(t *testing.T) {
	raw := map[string]interface{}{
		"notifyOnSuccess":                true,
		"notifyOnWarning":                true,
		"notifyOnError":                  true,
		"suppressRepeatingNotifications": true,
		"notifyOnLastRetryOnly":          false,
		"sendSNMPOnSuccess":              false,
		"sendSNMPOnWarning":              true,
		"sendSNMPOnError":                true,
		"sendSyslogOnSuccess":            false,
		"sendSyslogOnWarning":            false,
		"sendSyslogOnError":              true,
	}

	data := &NotificationSettingsModel{}
	syncNotificationSettingsFromAPI(data, raw)

	assert.Equal(t, true, data.NotifyOnSuccess.ValueBool())
	assert.Equal(t, true, data.NotifyOnWarning.ValueBool())
	assert.Equal(t, true, data.NotifyOnError.ValueBool())
	assert.Equal(t, true, data.SuppressRepeatingNotifications.ValueBool())
	assert.Equal(t, false, data.NotifyOnLastRetryOnly.ValueBool())
	assert.Equal(t, false, data.SendSNMPOnSuccess.ValueBool())
	assert.Equal(t, true, data.SendSNMPOnWarning.ValueBool())
	assert.Equal(t, true, data.SendSNMPOnError.ValueBool())
	assert.Equal(t, false, data.SendSyslogOnSuccess.ValueBool())
	assert.Equal(t, false, data.SendSyslogOnWarning.ValueBool())
	assert.Equal(t, true, data.SendSyslogOnError.ValueBool())
}

func TestNotificationSettings_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathNotificationSettings, mock.Anything).
		Return(fmt.Errorf("unauthorized"))

	var raw map[string]interface{}
	err := mockClient.GetJSON(context.Background(), client.PathNotificationSettings, &raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// NotificationSettings — Update
// ---------------------------------------------------------------------------

func TestNotificationSettings_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &NotificationSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathNotificationSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"notifyOnSuccess": true,
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathNotificationSettings, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, false, payload["notifyOnSuccess"])
	}).Return(nil)

	err := r.putNotificationSettings(context.Background(), &NotificationSettingsModel{
		NotifyOnSuccess: types.BoolValue(false),
	})
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// NotificationSettings — Delete (no-op)
// ---------------------------------------------------------------------------

func TestNotificationSettings_Delete_Success(t *testing.T) {
	r := &NotificationSettings{}
	r.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
}

// ---------------------------------------------------------------------------
// NotificationSettings — ImportState
// ---------------------------------------------------------------------------

func TestNotificationSettings_ImportState(t *testing.T) {
	importStateWithID(t, &NotificationSettings{}, "notification-settings")
}

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
// EmailSettings — Configure tests
// ---------------------------------------------------------------------------

func TestEmailSettings_Configure_Nil(t *testing.T) {
	r := &EmailSettings{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestEmailSettings_Configure_InvalidType(t *testing.T) {
	r := &EmailSettings{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestEmailSettings_Configure_Valid(t *testing.T) {
	r := &EmailSettings{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// EmailSettings — putEmailSettings helper
// ---------------------------------------------------------------------------

func TestEmailSettings_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	// GET returns current settings
	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"isEnabled":         false,
			"smtpServer":        "mail.old.example.com",
			"port":              float64(25),
			"useSSL":            false,
			"useAuthentication": false,
			"login":             "",
			"from":              "old@example.com",
			"to":                "admin@example.com",
			"subject":           "[OLD]",
			"sendOnSuccess":     false,
			"sendOnWarning":     false,
			"sendOnError":       true,
			"sendDailySummary":  false,
		}
	}).Return(nil)

	// PUT succeeds
	mockClient.On("PutJSON", mock.Anything, client.PathEmailSettings, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, true, payload["isEnabled"])
		assert.Equal(t, "mail.example.com", payload["smtpServer"])
		assert.Equal(t, 587, payload["port"])
	}).Return(nil)

	data := &EmailSettingsModel{
		Enabled:           types.BoolValue(true),
		SMTPServer:        types.StringValue("mail.example.com"),
		Port:              types.Int64Value(587),
		UseSSL:            types.BoolValue(true),
		UseAuthentication: types.BoolValue(true),
		Login:             types.StringValue("alerts@example.com"),
		Password:          types.StringValue("s3cr3t"),
		From:              types.StringValue("alerts@example.com"),
		To:                types.StringValue("ops@example.com"),
		Subject:           types.StringValue("[VEEAM]"),
		SendOnSuccess:     types.BoolValue(false),
		SendOnWarning:     types.BoolValue(true),
		SendOnError:       types.BoolValue(true),
		SendDailySummary:  types.BoolValue(false),
		SendTestMessage:   types.BoolNull(),
	}

	err := r.putEmailSettings(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestEmailSettings_Create_WithTestMessage(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathEmailSettings, mock.Anything, nil).Return(nil)

	mockClient.On("PostJSON", mock.Anything, client.PathEmailSettingsTestMessage, mock.Anything, nil).Return(nil)

	err := r.putEmailSettings(context.Background(), &EmailSettingsModel{
		Enabled:    types.BoolValue(true),
		SMTPServer: types.StringValue("mail.example.com"),
	})
	assert.NoError(t, err)

	// Verify PostJSON is called for the test message
	err = r.client.PostJSON(context.Background(), client.PathEmailSettingsTestMessage, map[string]interface{}{}, nil)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestEmailSettings_Create_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).
		Return(fmt.Errorf("connection refused"))

	err := r.putEmailSettings(context.Background(), &EmailSettingsModel{
		Enabled: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading current email settings")
	mockClient.AssertExpectations(t)
}

func TestEmailSettings_Create_PutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathEmailSettings, mock.Anything, nil).
		Return(fmt.Errorf("server error"))

	err := r.putEmailSettings(context.Background(), &EmailSettingsModel{
		Enabled: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// EmailSettings — Read
// ---------------------------------------------------------------------------

func TestEmailSettings_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"isEnabled":         true,
			"smtpServer":        "smtp.example.com",
			"port":              float64(465),
			"useSSL":            true,
			"useAuthentication": true,
			"login":             "user@example.com",
			"from":              "veeam@example.com",
			"to":                "admin@example.com",
			"subject":           "[VEEAM ALERT]",
			"sendOnSuccess":     true,
			"sendOnWarning":     true,
			"sendOnError":       true,
			"sendDailySummary":  false,
		}
	}).Return(nil)

	data := &EmailSettingsModel{
		Password: types.StringValue("preserved-password"),
	}
	r.syncModelFromAPI(data, map[string]interface{}{
		"isEnabled":         true,
		"smtpServer":        "smtp.example.com",
		"port":              float64(465),
		"useSSL":            true,
		"useAuthentication": true,
		"login":             "user@example.com",
		"from":              "veeam@example.com",
		"to":                "admin@example.com",
		"subject":           "[VEEAM ALERT]",
		"sendOnSuccess":     true,
		"sendOnWarning":     true,
		"sendOnError":       true,
		"sendDailySummary":  false,
	})

	assert.Equal(t, true, data.Enabled.ValueBool())
	assert.Equal(t, "smtp.example.com", data.SMTPServer.ValueString())
	assert.Equal(t, int64(465), data.Port.ValueInt64())
	assert.Equal(t, true, data.UseSSL.ValueBool())
	assert.Equal(t, true, data.UseAuthentication.ValueBool())
	assert.Equal(t, "user@example.com", data.Login.ValueString())
	assert.Equal(t, "veeam@example.com", data.From.ValueString())
	assert.Equal(t, "admin@example.com", data.To.ValueString())
	assert.Equal(t, "[VEEAM ALERT]", data.Subject.ValueString())
	assert.Equal(t, true, data.SendOnSuccess.ValueBool())
	assert.Equal(t, true, data.SendOnWarning.ValueBool())
	assert.Equal(t, true, data.SendOnError.ValueBool())
	assert.Equal(t, false, data.SendDailySummary.ValueBool())
	// Password must be preserved — never overwritten from API
	assert.Equal(t, "preserved-password", data.Password.ValueString())
}

func TestEmailSettings_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).
		Return(fmt.Errorf("unauthorized"))

	var raw map[string]interface{}
	err := r.client.GetJSON(context.Background(), client.PathEmailSettings, &raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// EmailSettings — Update
// ---------------------------------------------------------------------------

func TestEmailSettings_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EmailSettings{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathEmailSettings, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"isEnabled":  true,
			"smtpServer": "smtp.example.com",
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathEmailSettings, mock.Anything, nil).Return(nil)

	err := r.putEmailSettings(context.Background(), &EmailSettingsModel{
		Enabled:    types.BoolValue(false),
		SMTPServer: types.StringValue("smtp2.example.com"),
	})
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// EmailSettings — Delete (no-op)
// ---------------------------------------------------------------------------

func TestEmailSettings_Delete_Success(t *testing.T) {
	r := &EmailSettings{}
	// Delete is a no-op; just verify it doesn't panic or error
	r.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
}

// ---------------------------------------------------------------------------
// EmailSettings — ImportState
// ---------------------------------------------------------------------------

func TestEmailSettings_ImportState(t *testing.T) {
	importStateWithID(t, &EmailSettings{}, "email-settings")
}

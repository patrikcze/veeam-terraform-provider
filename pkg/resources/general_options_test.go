package resources

// ---------------------------------------------------------------------------
// Unit tests for the veeam_general_options singleton resource.
//
// All tests use MockVeeamClient (defined in backup_job_test.go) and the
// tfsdk helpers from resources_crud_test.go.  No live Veeam server is needed.
//
// Run with:
//   go test ./pkg/resources/ -run TestGeneralOptions -v
// ---------------------------------------------------------------------------

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func TestGeneralOptions_Configure_Nil(t *testing.T) {
	testConfigureNil(t, &GeneralOptions{})
}

func TestGeneralOptions_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &GeneralOptions{})
}

func TestGeneralOptions_Configure_Valid(t *testing.T) {
	testConfigureValid(t, &GeneralOptions{})
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestGeneralOptions_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	// GET returns a payload with all nested sections already present.
	mockClient.On("GetJSON", mock.Anything, client.PathGeneralOptions, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"storageLatencyControl": map[string]interface{}{
					"isEnabled":      false,
					"latencyLimitMs": float64(20),
				},
				"emailNotifications": map[string]interface{}{
					"isEnabled":  false,
					"smtpServer": "smtp.example.com",
					"port":       float64(25),
					"from":       "veeam@example.com",
					"to":         "admin@example.com",
					"subject":    "[Veeam] %JobResult%",
				},
				"snmpNotifications": map[string]interface{}{
					"isEnabled": false,
				},
				"syslogNotifications": map[string]interface{}{
					"isEnabled": false,
					"dnsName":   "syslog.example.com",
					"port":      float64(514),
				},
			}
		}).Return(nil)

	// PUT must succeed.
	mockClient.On("PutJSON", mock.Anything, client.PathGeneralOptions, mock.Anything, nil).
		Return(nil)

	req := resource.CreateRequest{
		Plan: buildNullResourcePlan(NewGeneralOptions()),
	}
	resp := &resource.CreateResponse{
		State: buildNullResourceState(NewGeneralOptions()),
	}

	r.Create(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)

	var state GeneralOptionsModel
	require.False(t, resp.State.Get(context.Background(), &state).HasError())
	assert.Equal(t, generalOptionsID, state.ID.ValueString())

	mockClient.AssertExpectations(t)
}

func TestGeneralOptions_Create_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathGeneralOptions, mock.Anything).
		Return(errors.New("connection refused"))

	req := resource.CreateRequest{
		Plan: buildNullResourcePlan(NewGeneralOptions()),
	}
	resp := &resource.CreateResponse{
		State: buildNullResourceState(NewGeneralOptions()),
	}

	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestGeneralOptions_Create_PutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathGeneralOptions, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{}
		}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathGeneralOptions, mock.Anything, nil).
		Return(errors.New("server error 500"))

	req := resource.CreateRequest{
		Plan: buildNullResourcePlan(NewGeneralOptions()),
	}
	resp := &resource.CreateResponse{
		State: buildNullResourceState(NewGeneralOptions()),
	}

	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func TestGeneralOptions_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathGeneralOptions, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"storageLatencyControl": map[string]interface{}{
					"isEnabled":      true,
					"latencyLimitMs": float64(20),
				},
				"emailNotifications": map[string]interface{}{
					"isEnabled":  true,
					"smtpServer": "smtp.example.com",
					"port":       float64(587),
					"from":       "noreply@example.com",
					"to":         "ops@example.com",
					"subject":    "[Veeam] %JobResult%",
				},
				"snmpNotifications": map[string]interface{}{
					"isEnabled": false,
				},
				"syslogNotifications": map[string]interface{}{
					"isEnabled": true,
					"dnsName":   "syslog.internal",
					"port":      float64(514),
				},
			}
		}).Return(nil)

	req := resource.ReadRequest{
		State: buildNullResourceState(NewGeneralOptions()),
	}
	resp := &resource.ReadResponse{
		State: buildNullResourceState(NewGeneralOptions()),
	}

	r.Read(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)

	var state GeneralOptionsModel
	require.False(t, resp.State.Get(context.Background(), &state).HasError())

	assert.Equal(t, generalOptionsID, state.ID.ValueString())
	assert.True(t, state.StorageLatencyControlEnabled.ValueBool())
	assert.Equal(t, int64(20), state.StorageLatencyLimitMs.ValueInt64())
	assert.True(t, state.EmailNotificationsEnabled.ValueBool())
	assert.Equal(t, "smtp.example.com", state.EmailSMTPServer.ValueString())
	assert.Equal(t, int64(587), state.EmailSMTPPort.ValueInt64())
	assert.Equal(t, "noreply@example.com", state.EmailFrom.ValueString())
	assert.Equal(t, "ops@example.com", state.EmailTo.ValueString())
	assert.Equal(t, "[Veeam] %JobResult%", state.EmailSubject.ValueString())
	assert.False(t, state.SNMPNotificationsEnabled.ValueBool())
	assert.True(t, state.SyslogNotificationsEnabled.ValueBool())
	assert.Equal(t, "syslog.internal", state.SyslogServer.ValueString())
	assert.Equal(t, int64(514), state.SyslogPort.ValueInt64())

	mockClient.AssertExpectations(t)
}

func TestGeneralOptions_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathGeneralOptions, mock.Anything).
		Return(errors.New("timeout"))

	req := resource.ReadRequest{
		State: buildNullResourceState(NewGeneralOptions()),
	}
	resp := &resource.ReadResponse{
		State: buildNullResourceState(NewGeneralOptions()),
	}

	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestGeneralOptions_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathGeneralOptions, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"storageLatencyControl": map[string]interface{}{
					"isEnabled":      false,
					"latencyLimitMs": float64(20),
				},
				"emailNotifications": map[string]interface{}{
					"isEnabled": false,
				},
				"snmpNotifications": map[string]interface{}{
					"isEnabled": false,
				},
				"syslogNotifications": map[string]interface{}{
					"isEnabled": false,
				},
			}
		}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathGeneralOptions, mock.Anything, nil).
		Return(nil)

	req := resource.UpdateRequest{
		Plan:  buildNullResourcePlan(NewGeneralOptions()),
		State: buildNullResourceState(NewGeneralOptions()),
	}
	resp := &resource.UpdateResponse{
		State: buildNullResourceState(NewGeneralOptions()),
	}

	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestGeneralOptions_Delete_Success(t *testing.T) {
	// Delete is a state-only operation; the mock should receive no API calls.
	mockClient := new(MockVeeamClient)
	r := &GeneralOptions{client: mockClient}

	req := resource.DeleteRequest{
		State: buildNullResourceState(NewGeneralOptions()),
	}
	resp := &resource.DeleteResponse{}

	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	// Verify no API calls were made.
	mockClient.AssertNotCalled(t, "GetJSON")
	mockClient.AssertNotCalled(t, "PutJSON")
	mockClient.AssertNotCalled(t, "DeleteJSON")
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func TestGeneralOptions_ImportState(t *testing.T) {
	r := &GeneralOptions{}
	resp := importStateWithID(t, r, generalOptionsID)
	assert.False(t, resp.Diagnostics.HasError())

	var state GeneralOptionsModel
	require.False(t, resp.State.Get(context.Background(), &state).HasError())
	assert.Equal(t, generalOptionsID, state.ID.ValueString())
}

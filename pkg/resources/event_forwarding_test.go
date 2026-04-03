package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

func TestEventForwarding_Metadata(t *testing.T) {
	r := NewEventForwarding()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_event_forwarding", resp.TypeName)
}

func TestEventForwarding_Schema(t *testing.T) {
	r := &EventForwarding{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "snmp_enabled")
	assert.Contains(t, resp.Schema.Attributes, "snmp_host")
	assert.Contains(t, resp.Schema.Attributes, "syslog_enabled")
	assert.Contains(t, resp.Schema.Attributes, "syslog_host")
	assert.Contains(t, resp.Schema.Attributes, "syslog_protocol")
}

func TestEventForwarding_Configure_Nil(t *testing.T) {
	r := &EventForwarding{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestEventForwarding_Configure_InvalidType(t *testing.T) {
	r := &EventForwarding{}
	req := resource.ConfigureRequest{ProviderData: 42}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestEventForwarding_SyncFromPayload_Full(t *testing.T) {
	raw := map[string]interface{}{
		"snmp": map[string]interface{}{
			"isEnabled": true,
			"host":      "snmp.example.com",
			"port":      float64(162),
			"community": "public",
		},
		"syslog": map[string]interface{}{
			"isEnabled": true,
			"host":      "syslog.example.com",
			"port":      float64(514),
			"protocol":  "UDP",
		},
	}

	data := &EventForwardingModel{}
	syncEventForwardingFromPayload(raw, data)

	assert.True(t, data.SNMPEnabled.ValueBool())
	assert.Equal(t, "snmp.example.com", data.SNMPHost.ValueString())
	assert.Equal(t, int64(162), data.SNMPPort.ValueInt64())
	assert.Equal(t, "public", data.SNMPCommunity.ValueString())
	assert.True(t, data.SyslogEnabled.ValueBool())
	assert.Equal(t, "syslog.example.com", data.SyslogHost.ValueString())
	assert.Equal(t, int64(514), data.SyslogPort.ValueInt64())
	assert.Equal(t, "UDP", data.SyslogProtocol.ValueString())
}

func TestEventForwarding_SyncFromPayload_Empty(t *testing.T) {
	raw := map[string]interface{}{}
	data := &EventForwardingModel{}
	syncEventForwardingFromPayload(raw, data)

	assert.False(t, data.SNMPEnabled.ValueBool())
	assert.False(t, data.SyslogEnabled.ValueBool())
}

func TestEventForwarding_Apply_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathEventForwarding, mock.Anything).
		Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathEventForwarding, mock.Anything, nil).
		Return(nil)

	r := &EventForwarding{client: mockClient}
	data := &EventForwardingModel{
		SNMPEnabled:    types.BoolValue(true),
		SNMPHost:       types.StringValue("snmp.example.com"),
		SNMPPort:       types.Int64Value(162),
		SNMPCommunity:  types.StringValue("public"),
		SyslogEnabled:  types.BoolNull(),
		SyslogHost:     types.StringNull(),
		SyslogPort:     types.Int64Null(),
		SyslogProtocol: types.StringNull(),
	}

	err := r.applyEventForwarding(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestEventForwarding_Delete_NoOp(t *testing.T) {
	r := &EventForwarding{}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestEventForwarding_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &EventForwarding{}
}

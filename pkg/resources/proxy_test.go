package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestProxy_BuildSpec(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{
		Type:                  types.StringValue("ViProxy"),
		HostID:                types.StringValue("host-123"),
		Description:           types.StringValue("Proxy test"),
		TransportMode:         types.StringValue("Auto"),
		FailoverToNetwork:     types.BoolValue(true),
		HostToProxyEncryption: types.BoolValue(true),
		MaxTaskCount:          types.Int64Value(2),
	}

	spec := resource.buildSpec(data)

	vi, ok := spec.(*models.ViProxySpec)
	assert.True(t, ok, "expected *models.ViProxySpec")
	assert.Equal(t, models.ProxyTypeViProxy, vi.Type)
	assert.Equal(t, "host-123", vi.Server.HostID)
	assert.Equal(t, models.TransportModeAuto, vi.Server.TransportMode)
	assert.True(t, vi.Server.FailoverToNetwork)
	assert.True(t, vi.Server.HostToProxyEncryption)
	assert.Equal(t, 2, vi.Server.MaxTaskCount)
}

func TestProxy_SyncFromAPI_PreservesPlanOnEmptyAPIValues(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{
		Type:        types.StringValue("ViProxy"),
		Description: types.StringValue("Planned description"),
		Name:        types.StringValue("Planned name"),
	}

	api := &models.ProxyModel{}
	resource.syncFromAPI(data, api)

	assert.Equal(t, "ViProxy", data.Type.ValueString())
	assert.Equal(t, "Planned description", data.Description.ValueString())
	assert.Equal(t, "Planned name", data.Name.ValueString())
}

func TestProxy_SyncFromAPI_UsesAPIValuesWhenPresent(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{}

	api := &models.ProxyModel{
		Name:        "proxy-1",
		Description: "API description",
		Type:        models.ProxyTypeViProxy,
	}
	resource.syncFromAPI(data, api)

	assert.Equal(t, "proxy-1", data.Name.ValueString())
	assert.Equal(t, "API description", data.Description.ValueString())
	assert.Equal(t, "ViProxy", data.Type.ValueString())
}

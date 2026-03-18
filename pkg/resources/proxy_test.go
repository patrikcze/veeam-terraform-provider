package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestProxy_BuildSpec_ViProxy(t *testing.T) {
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

func TestProxy_BuildSpec_HvProxy(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{
		Type:         types.StringValue("HvProxy"),
		HostID:       types.StringValue("hv-host-1"),
		Description:  types.StringValue("Hyper-V proxy"),
		MaxTaskCount: types.Int64Value(4),
	}

	spec := resource.buildSpec(data)

	hv, ok := spec.(*models.HvProxySpec)
	assert.True(t, ok, "expected *models.HvProxySpec")
	assert.Equal(t, models.ProxyTypeHvProxy, hv.Type)
	assert.Equal(t, "hv-host-1", hv.Server.HostID)
	assert.Equal(t, 4, hv.Server.MaxTaskCount)
}

func TestProxy_BuildSpec_GeneralPurposeProxy(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{
		Type:         types.StringValue("GeneralPurposeProxy"),
		HostID:       types.StringValue("gp-host-1"),
		MaxTaskCount: types.Int64Value(2),
	}

	spec := resource.buildSpec(data)

	gp, ok := spec.(*models.GeneralPurposeProxySpec)
	assert.True(t, ok, "expected *models.GeneralPurposeProxySpec")
	assert.Equal(t, models.ProxyTypeGeneralPurposeProxy, gp.Type)
	assert.Equal(t, "gp-host-1", gp.Server.HostID)
	assert.Equal(t, 2, gp.Server.MaxTaskCount)
}

func TestProxy_SyncFromAPI_PreservesPlanOnEmptyAPIValues(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{
		Type:        types.StringValue("ViProxy"),
		Description: types.StringValue("Planned description"),
		Name:        types.StringValue("Planned name"),
	}

	api := &models.ViProxyModel{}
	resource.syncFromAPI(data, api)

	assert.Equal(t, "ViProxy", data.Type.ValueString())
	assert.Equal(t, "Planned description", data.Description.ValueString())
	assert.Equal(t, "Planned name", data.Name.ValueString())
}

func TestProxy_SyncFromAPI_UsesAPIValuesWhenPresent(t *testing.T) {
	resource := &Proxy{}
	data := &ProxyModel{}

	api := &models.ViProxyModel{
		ProxyModel: models.ProxyModel{
			Name:        "proxy-1",
			Description: "API description",
			Type:        models.ProxyTypeViProxy,
		},
		Server: &models.ProxyServerSettings{
			HostID:        "host-456",
			MaxTaskCount:  3,
			TransportMode: models.TransportModeNetwork,
		},
	}
	resource.syncFromAPI(data, api)

	assert.Equal(t, "proxy-1", data.Name.ValueString())
	assert.Equal(t, "API description", data.Description.ValueString())
	assert.Equal(t, "ViProxy", data.Type.ValueString())
	assert.Equal(t, "host-456", data.HostID.ValueString())
	assert.Equal(t, int64(3), data.MaxTaskCount.ValueInt64())
	assert.Equal(t, "Network", data.TransportMode.ValueString())
}

package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

var (
	_ datasource.DataSource              = &ProxiesDataSource{}
	_ datasource.DataSourceWithConfigure = &ProxiesDataSource{}
)

type ProxiesDataSource struct{ client client.APIClient }

type ProxiesDataSourceModel struct {
	ID      types.String     `tfsdk:"id"`
	ProxyID types.String     `tfsdk:"proxy_id"`
	Proxies []ProxyDataModel `tfsdk:"proxies"`
}

type ProxyDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func NewProxiesDataSource() datasource.DataSource { return &ProxiesDataSource{} }
func (d *ProxiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxies"
}
func (d *ProxiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true}, "proxy_id": schema.StringAttribute{Optional: true},
		"proxies": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true}, "type": schema.StringAttribute{Computed: true}, "description": schema.StringAttribute{Computed: true}}}},
	}}
}
func (d *ProxiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider.")
		return
	}
	d.client = c
}
func (d *ProxiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProxiesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mapOne := func(item map[string]interface{}) ProxyDataModel {
		return ProxyDataModel{ID: types.StringValue(getStringValue(item, "id")), Name: types.StringValue(getStringValue(item, "name")), Type: types.StringValue(getStringValue(item, "type")), Description: types.StringValue(getStringValue(item, "description"))}
	}
	if !data.ProxyID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathProxyByID, data.ProxyID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read proxy", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Proxies = []ProxyDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("proxy", data.ProxyID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}
	items, err := fetchList(ctx, d.client.GetJSON, client.PathProxies)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list proxies", fmt.Sprintf("API error: %s", err))
		return
	}
	out := make([]ProxyDataModel, len(items))
	for i, item := range items {
		out[i] = mapOne(item)
	}
	data.ID = types.StringValue("proxies")
	data.Proxies = out
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

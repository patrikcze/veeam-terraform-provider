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
	_ datasource.DataSource              = &ProxyStatesDataSource{}
	_ datasource.DataSourceWithConfigure = &ProxyStatesDataSource{}
)

type ProxyStatesDataSource struct {
	client client.APIClient
}

type ProxyStatesDataSourceModel struct {
	ID     types.String          `tfsdk:"id"`
	States []ProxyStateDataModel `tfsdk:"states"`
}

type ProxyStateDataModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	Type   types.String `tfsdk:"type"`
}

func NewProxyStatesDataSource() datasource.DataSource { return &ProxyStatesDataSource{} }

func (d *ProxyStatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_states"
}

func (d *ProxyStatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"states": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true},
			"name":   schema.StringAttribute{Computed: true},
			"status": schema.StringAttribute{Computed: true},
			"type":   schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *ProxyStatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProxyStatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProxyStatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathProxyStates)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list proxy states", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]ProxyStateDataModel, len(items))
	for i, item := range items {
		mapped[i] = ProxyStateDataModel{
			ID:     types.StringValue(getStringValue(item, "id")),
			Name:   types.StringValue(getStringValue(item, "name")),
			Status: types.StringValue(getStringValue(item, "status")),
			Type:   types.StringValue(getStringValue(item, "type")),
		}
	}

	data.ID = types.StringValue("proxy_states")
	data.States = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

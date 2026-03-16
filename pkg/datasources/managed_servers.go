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
	_ datasource.DataSource              = &ManagedServersDataSource{}
	_ datasource.DataSourceWithConfigure = &ManagedServersDataSource{}
)

type ManagedServersDataSource struct{ client client.APIClient }

type ManagedServersDataSourceModel struct {
	ID       types.String             `tfsdk:"id"`
	ServerID types.String             `tfsdk:"server_id"`
	Servers  []ManagedServerDataModel `tfsdk:"servers"`
}

type ManagedServerDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
}

func NewManagedServersDataSource() datasource.DataSource { return &ManagedServersDataSource{} }
func (d *ManagedServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_servers"
}
func (d *ManagedServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true}, "server_id": schema.StringAttribute{Optional: true},
		"servers": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true}, "type": schema.StringAttribute{Computed: true}, "description": schema.StringAttribute{Computed: true}, "status": schema.StringAttribute{Computed: true}}}},
	}}
}
func (d *ManagedServersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ManagedServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ManagedServersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mapOne := func(item map[string]interface{}) ManagedServerDataModel {
		return ManagedServerDataModel{ID: types.StringValue(getStringValue(item, "id")), Name: types.StringValue(getStringValue(item, "name")), Type: types.StringValue(getStringValue(item, "type")), Description: types.StringValue(getStringValue(item, "description")), Status: types.StringValue(getStringValue(item, "status"))}
	}
	if !data.ServerID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathManagedServerByID, data.ServerID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read managed server", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Servers = []ManagedServerDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("managed_server", data.ServerID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}
	items, err := fetchList(ctx, d.client.GetJSON, client.PathManagedServers)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list managed servers", fmt.Sprintf("API error: %s", err))
		return
	}
	out := make([]ManagedServerDataModel, len(items))
	for i, item := range items {
		out[i] = mapOne(item)
	}
	data.ID = types.StringValue("managed_servers")
	data.Servers = out
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

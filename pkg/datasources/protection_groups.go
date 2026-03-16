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
	_ datasource.DataSource              = &ProtectionGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProtectionGroupsDataSource{}
)

type ProtectionGroupsDataSource struct {
	client client.APIClient
}

type ProtectionGroupsDataSourceModel struct {
	ID                types.String               `tfsdk:"id"`
	ProtectionGroupID types.String               `tfsdk:"protection_group_id"`
	ProtectionGroups  []ProtectionGroupDataModel `tfsdk:"protection_groups"`
}

type ProtectionGroupDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func NewProtectionGroupsDataSource() datasource.DataSource { return &ProtectionGroupsDataSource{} }

func (d *ProtectionGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_groups"
}

func (d *ProtectionGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                  schema.StringAttribute{Computed: true},
		"protection_group_id": schema.StringAttribute{Optional: true},
		"protection_groups": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *ProtectionGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInstance, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider.")
		return
	}
	d.client = clientInstance
}

func (d *ProtectionGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtectionGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) ProtectionGroupDataModel {
		return ProtectionGroupDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	if !data.ProtectionGroupID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathProtectionGroupByID, data.ProtectionGroupID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read protection group", fmt.Sprintf("API error: %s", err))
			return
		}
		data.ProtectionGroups = []ProtectionGroupDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("protection_group", data.ProtectionGroupID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathProtectionGroups)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list protection groups", fmt.Sprintf("API error: %s", err))
		return
	}
	mapped := make([]ProtectionGroupDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("protection_groups")
	data.ProtectionGroups = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

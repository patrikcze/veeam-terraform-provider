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
	_ datasource.DataSource              = &WanAcceleratorsDataSource{}
	_ datasource.DataSourceWithConfigure = &WanAcceleratorsDataSource{}
)

type WanAcceleratorsDataSource struct {
	client client.APIClient
}

type WanAcceleratorsDataSourceModel struct {
	ID            types.String              `tfsdk:"id"`
	AcceleratorID types.String              `tfsdk:"accelerator_id"`
	Accelerators  []WanAcceleratorDataModel `tfsdk:"accelerators"`
}

type WanAcceleratorDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func NewWanAcceleratorsDataSource() datasource.DataSource { return &WanAcceleratorsDataSource{} }

func (d *WanAcceleratorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan_accelerators"
}

func (d *WanAcceleratorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true},
		"accelerator_id": schema.StringAttribute{Optional: true},
		"accelerators": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *WanAcceleratorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WanAcceleratorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WanAcceleratorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) WanAcceleratorDataModel {
		return WanAcceleratorDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Type:        types.StringValue(getStringValue(item, "type")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	if !data.AcceleratorID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathWanAcceleratorByID, data.AcceleratorID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read WAN accelerator", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Accelerators = []WanAcceleratorDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("wan_accelerator", data.AcceleratorID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathWanAccelerators)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list WAN accelerators", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]WanAcceleratorDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("wan_accelerators")
	data.Accelerators = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

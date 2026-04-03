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
	_ datasource.DataSource              = &ProtectedComputersDataSource{}
	_ datasource.DataSourceWithConfigure = &ProtectedComputersDataSource{}
)

type ProtectedComputersDataSource struct {
	client client.APIClient
}

type ProtectedComputersDataSourceModel struct {
	ID         types.String                 `tfsdk:"id"`
	ComputerID types.String                 `tfsdk:"computer_id"`
	Computers  []ProtectedComputerDataModel `tfsdk:"computers"`
}

type ProtectedComputerDataModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Status   types.String `tfsdk:"status"`
	Platform types.String `tfsdk:"platform"`
}

func NewProtectedComputersDataSource() datasource.DataSource { return &ProtectedComputersDataSource{} }

func (d *ProtectedComputersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protected_computers"
}

func (d *ProtectedComputersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"computer_id": schema.StringAttribute{Optional: true},
		"computers": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true},
			"name":     schema.StringAttribute{Computed: true},
			"type":     schema.StringAttribute{Computed: true},
			"status":   schema.StringAttribute{Computed: true},
			"platform": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *ProtectedComputersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProtectedComputersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtectedComputersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) ProtectedComputerDataModel {
		return ProtectedComputerDataModel{
			ID:       types.StringValue(getStringValue(item, "id")),
			Name:     types.StringValue(getStringValue(item, "name")),
			Type:     types.StringValue(getStringValue(item, "type")),
			Status:   types.StringValue(getStringValue(item, "status")),
			Platform: types.StringValue(getStringValue(item, "platform")),
		}
	}

	if !data.ComputerID.IsNull() {
		// The API has no single-computer-by-ID endpoint under protectedComputers; fetch the list and filter.
		items, err := fetchList(ctx, d.client.GetJSON, client.PathProtectedComputers)
		if err != nil {
			resp.Diagnostics.AddError("Failed to list protected computers", fmt.Sprintf("API error: %s", err))
			return
		}
		for _, item := range items {
			if getStringValue(item, "id") == data.ComputerID.ValueString() {
				data.Computers = []ProtectedComputerDataModel{mapOne(item)}
				data.ID = types.StringValue(normalizeDataSourceID("protected_computer", data.ComputerID.ValueString()))
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}
		data.Computers = []ProtectedComputerDataModel{}
		data.ID = types.StringValue(normalizeDataSourceID("protected_computer", data.ComputerID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathProtectedComputers)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list protected computers", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]ProtectedComputerDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("protected_computers")
	data.Computers = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

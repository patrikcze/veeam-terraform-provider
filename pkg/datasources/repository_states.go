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
	_ datasource.DataSource              = &RepositoryStatesDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoryStatesDataSource{}
)

type RepositoryStatesDataSource struct {
	client client.APIClient
}

type RepositoryStatesDataSourceModel struct {
	ID     types.String               `tfsdk:"id"`
	States []RepositoryStateDataModel `tfsdk:"states"`
}

type RepositoryStateDataModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Capacity  types.Int64  `tfsdk:"capacity"`
	FreeSpace types.Int64  `tfsdk:"free_space"`
	UsedSpace types.Int64  `tfsdk:"used_space"`
}

func NewRepositoryStatesDataSource() datasource.DataSource { return &RepositoryStatesDataSource{} }

func (d *RepositoryStatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_states"
}

func (d *RepositoryStatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"states": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"name":       schema.StringAttribute{Computed: true},
			"type":       schema.StringAttribute{Computed: true},
			"status":     schema.StringAttribute{Computed: true},
			"capacity":   schema.Int64Attribute{Computed: true},
			"free_space": schema.Int64Attribute{Computed: true},
			"used_space": schema.Int64Attribute{Computed: true},
		}}},
	}}
}

func (d *RepositoryStatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RepositoryStatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepositoryStatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathRepositoryState)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list repository states", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]RepositoryStateDataModel, len(items))
	for i, item := range items {
		status := "Offline"
		if getBoolValue(item, "isOnline") {
			status = "Online"
		}
		mapped[i] = RepositoryStateDataModel{
			ID:        types.StringValue(getStringValue(item, "id")),
			Name:      types.StringValue(getStringValue(item, "name")),
			Type:      types.StringValue(getStringValue(item, "type")),
			Status:    types.StringValue(status),
			Capacity:  types.Int64Value(getInt64Value(item, "capacityGB")),
			FreeSpace: types.Int64Value(getInt64Value(item, "freeGB")),
			UsedSpace: types.Int64Value(getInt64Value(item, "usedSpaceGB")),
		}
	}

	data.ID = types.StringValue("repository_states")
	data.States = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

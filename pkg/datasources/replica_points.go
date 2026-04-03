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
	_ datasource.DataSource              = &ReplicaPointsDataSource{}
	_ datasource.DataSourceWithConfigure = &ReplicaPointsDataSource{}
)

type ReplicaPointsDataSource struct {
	client client.APIClient
}

type ReplicaPointsDataSourceModel struct {
	ID             types.String            `tfsdk:"id"`
	ReplicaPointID types.String            `tfsdk:"replica_point_id"`
	ReplicaPoints  []ReplicaPointDataModel `tfsdk:"replica_points"`
}

type ReplicaPointDataModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ReplicaID    types.String `tfsdk:"replica_id"`
	CreationTime types.String `tfsdk:"creation_time"`
}

func NewReplicaPointsDataSource() datasource.DataSource { return &ReplicaPointsDataSource{} }

func (d *ReplicaPointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_replica_points"
}

func (d *ReplicaPointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true},
		"replica_point_id": schema.StringAttribute{Optional: true},
		"replica_points": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true},
			"name":          schema.StringAttribute{Computed: true},
			"replica_id":    schema.StringAttribute{Computed: true},
			"creation_time": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *ReplicaPointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ReplicaPointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ReplicaPointsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) ReplicaPointDataModel {
		return ReplicaPointDataModel{
			ID:           types.StringValue(getStringValue(item, "id")),
			Name:         types.StringValue(getStringValue(item, "name")),
			ReplicaID:    types.StringValue(getStringValue(item, "replicaId")),
			CreationTime: types.StringValue(getStringValue(item, "creationTime")),
		}
	}

	if !data.ReplicaPointID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathReplicaPointByID, data.ReplicaPointID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read replica point", fmt.Sprintf("API error: %s", err))
			return
		}
		data.ReplicaPoints = []ReplicaPointDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("replica_point", data.ReplicaPointID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathReplicaPoints)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list replica points", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]ReplicaPointDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("replica_points")
	data.ReplicaPoints = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

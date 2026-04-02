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
	_ datasource.DataSource              = &ReplicasDataSource{}
	_ datasource.DataSourceWithConfigure = &ReplicasDataSource{}
)

type ReplicasDataSource struct {
	client client.APIClient
}

type ReplicasDataSourceModel struct {
	ID        types.String       `tfsdk:"id"`
	ReplicaID types.String       `tfsdk:"replica_id"`
	Replicas  []ReplicaDataModel `tfsdk:"replicas"`
}

type ReplicaDataModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	State    types.String `tfsdk:"state"`
	Platform types.String `tfsdk:"platform"`
}

func NewReplicasDataSource() datasource.DataSource { return &ReplicasDataSource{} }

func (d *ReplicasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_replicas"
}

func (d *ReplicasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true},
		"replica_id": schema.StringAttribute{Optional: true},
		"replicas": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true},
			"name":     schema.StringAttribute{Computed: true},
			"type":     schema.StringAttribute{Computed: true},
			"state":    schema.StringAttribute{Computed: true},
			"platform": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *ReplicasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ReplicasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ReplicasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) ReplicaDataModel {
		return ReplicaDataModel{
			ID:       types.StringValue(getStringValue(item, "id")),
			Name:     types.StringValue(getStringValue(item, "name")),
			Type:     types.StringValue(getStringValue(item, "type")),
			State:    types.StringValue(getStringValue(item, "state")),
			Platform: types.StringValue(getStringValue(item, "platform")),
		}
	}

	if !data.ReplicaID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathReplicaByID, data.ReplicaID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read replica", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Replicas = []ReplicaDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("replica", data.ReplicaID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathReplicas)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list replicas", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]ReplicaDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("replicas")
	data.Replicas = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

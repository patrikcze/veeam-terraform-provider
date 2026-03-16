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
	_ datasource.DataSource              = &RestorePointsDataSource{}
	_ datasource.DataSourceWithConfigure = &RestorePointsDataSource{}
)

type RestorePointsDataSource struct {
	client client.APIClient
}

type RestorePointsDataSourceModel struct {
	ID             types.String            `tfsdk:"id"`
	RestorePointID types.String            `tfsdk:"restore_point_id"`
	BackupObjectID types.String            `tfsdk:"backup_object_id"`
	RestorePoints  []RestorePointDataModel `tfsdk:"restore_points"`
}

type RestorePointDataModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	BackupID     types.String `tfsdk:"backup_id"`
	CreationTime types.String `tfsdk:"creation_time"`
	Type         types.String `tfsdk:"type"`
}

func NewRestorePointsDataSource() datasource.DataSource { return &RestorePointsDataSource{} }

func (d *RestorePointsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_restore_points"
}

func (d *RestorePointsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true},
		"restore_point_id": schema.StringAttribute{Optional: true},
		"backup_object_id": schema.StringAttribute{Optional: true},
		"restore_points": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true},
			"name":          schema.StringAttribute{Computed: true},
			"backup_id":     schema.StringAttribute{Computed: true},
			"creation_time": schema.StringAttribute{Computed: true},
			"type":          schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *RestorePointsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RestorePointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RestorePointsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) RestorePointDataModel {
		return RestorePointDataModel{
			ID:           types.StringValue(getStringValue(item, "id")),
			Name:         types.StringValue(getStringValue(item, "name")),
			BackupID:     types.StringValue(getStringValue(item, "backupId")),
			CreationTime: types.StringValue(getStringValue(item, "creationTime")),
			Type:         types.StringValue(getStringValue(item, "type")),
		}
	}

	if !data.RestorePointID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathRestorePointByID, data.RestorePointID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read restore point", fmt.Sprintf("API error: %s", err))
			return
		}
		data.RestorePoints = []RestorePointDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("restore_point", data.RestorePointID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	endpoint := client.PathRestorePoints
	if !data.BackupObjectID.IsNull() {
		endpoint = fmt.Sprintf(client.PathBackupObjectRestorePoints, data.BackupObjectID.ValueString())
	}

	items, err := fetchList(ctx, d.client.GetJSON, endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list restore points", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]RestorePointDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.RestorePoints = mapped
	if !data.BackupObjectID.IsNull() {
		data.ID = types.StringValue(normalizeDataSourceID("restore_points_object", data.BackupObjectID.ValueString()))
	} else {
		data.ID = types.StringValue("restore_points")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

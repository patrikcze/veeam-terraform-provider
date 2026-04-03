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
	_ datasource.DataSource              = &BackupObjectsDataSource{}
	_ datasource.DataSourceWithConfigure = &BackupObjectsDataSource{}
)

type BackupObjectsDataSource struct {
	client client.APIClient
}

type BackupObjectsDataSourceModel struct {
	ID       types.String            `tfsdk:"id"`
	ObjectID types.String            `tfsdk:"object_id"`
	Objects  []BackupObjectDataModel `tfsdk:"objects"`
}

type BackupObjectDataModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	BackupID          types.String `tfsdk:"backup_id"`
	RestorePointCount types.Int64  `tfsdk:"restore_point_count"`
}

func NewBackupObjectsDataSource() datasource.DataSource { return &BackupObjectsDataSource{} }

func (d *BackupObjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_objects"
}

func (d *BackupObjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":        schema.StringAttribute{Computed: true},
		"object_id": schema.StringAttribute{Optional: true},
		"objects": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true},
			"name":                schema.StringAttribute{Computed: true},
			"type":                schema.StringAttribute{Computed: true},
			"backup_id":           schema.StringAttribute{Computed: true},
			"restore_point_count": schema.Int64Attribute{Computed: true},
		}}},
	}}
}

func (d *BackupObjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BackupObjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackupObjectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) BackupObjectDataModel {
		return BackupObjectDataModel{
			ID:                types.StringValue(getStringValue(item, "id")),
			Name:              types.StringValue(getStringValue(item, "name")),
			Type:              types.StringValue(getStringValue(item, "type")),
			BackupID:          types.StringValue(getStringValue(item, "backupId")),
			RestorePointCount: types.Int64Value(getInt64Value(item, "restorePointsCount")),
		}
	}

	if !data.ObjectID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathBackupObjectByID, data.ObjectID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read backup object", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Objects = []BackupObjectDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("backup_object", data.ObjectID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathBackupObjects)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list backup objects", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]BackupObjectDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("backup_objects")
	data.Objects = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ datasource.DataSource              = &BackupsDataSource{}
	_ datasource.DataSourceWithConfigure = &BackupsDataSource{}
)

type BackupsDataSource struct {
	client client.APIClient
}

type BackupsDataSourceModel struct {
	ID           types.String      `tfsdk:"id"`
	BackupID     types.String      `tfsdk:"backup_id"`
	IncludeFiles types.Bool        `tfsdk:"include_files"`
	Backups      []BackupDataModel `tfsdk:"backups"`
}

type BackupDataModel struct {
	ID      types.String          `tfsdk:"id"`
	Name    types.String          `tfsdk:"name"`
	Type    types.String          `tfsdk:"type"`
	JobID   types.String          `tfsdk:"job_id"`
	JobName types.String          `tfsdk:"job_name"`
	Files   []BackupFileDataModel `tfsdk:"files"`
}

type BackupFileDataModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Size types.Int64  `tfsdk:"size"`
}

func NewBackupsDataSource() datasource.DataSource { return &BackupsDataSource{} }

func (d *BackupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backups"
}

func (d *BackupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true},
			"backup_id":     schema.StringAttribute{Optional: true},
			"include_files": schema.BoolAttribute{Optional: true},
			"backups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":       schema.StringAttribute{Computed: true},
					"name":     schema.StringAttribute{Computed: true},
					"type":     schema.StringAttribute{Computed: true},
					"job_id":   schema.StringAttribute{Computed: true},
					"job_name": schema.StringAttribute{Computed: true},
					"files": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
							"id":   schema.StringAttribute{Computed: true},
							"name": schema.StringAttribute{Computed: true},
							"type": schema.StringAttribute{Computed: true},
							"size": schema.Int64Attribute{Computed: true},
						}},
					},
				}},
			},
		},
	}
}

func (d *BackupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BackupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapBackup := func(item map[string]interface{}) BackupDataModel {
		backup := BackupDataModel{
			ID:      types.StringValue(getStringValue(item, "id")),
			Name:    types.StringValue(getStringValue(item, "name")),
			Type:    types.StringValue(getStringValue(item, "type")),
			JobID:   types.StringValue(getStringValue(item, "jobId")),
			JobName: types.StringValue(getStringValue(item, "jobName")),
			Files:   []BackupFileDataModel{},
		}

		if !data.IncludeFiles.IsNull() && data.IncludeFiles.ValueBool() {
			files, err := fetchList(ctx, d.client.GetJSON, fmt.Sprintf(client.PathBackupFiles, getStringValue(item, "id")))
			if err == nil {
				mappedFiles := make([]BackupFileDataModel, len(files))
				for i, file := range files {
					mappedFiles[i] = BackupFileDataModel{
						ID:   types.StringValue(getStringValue(file, "id")),
						Name: types.StringValue(getStringValue(file, "name")),
						Type: types.StringValue(getStringValue(file, "type")),
						Size: types.Int64Value(getInt64Value(file, "size")),
					}
				}
				backup.Files = mappedFiles
			}
		}

		return backup
	}

	if !data.BackupID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathBackupByID, data.BackupID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read backup", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Backups = []BackupDataModel{mapBackup(item)}
		data.ID = types.StringValue(normalizeDataSourceID("backup", data.BackupID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathBackups)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list backups", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]BackupDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapBackup(item)
	}

	data.ID = types.StringValue("backups")
	data.Backups = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

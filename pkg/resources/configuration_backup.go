package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

var (
	_ resource.Resource                = &ConfigurationBackup{}
	_ resource.ResourceWithConfigure   = &ConfigurationBackup{}
	_ resource.ResourceWithImportState = &ConfigurationBackup{}
)

type ConfigurationBackup struct {
	client client.APIClient
}

type ConfigurationBackupModel struct {
	ID                   types.String `tfsdk:"id"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	RepositoryID         types.String `tfsdk:"repository_id"`
	RestorePointsToKeep  types.Int64  `tfsdk:"restore_points_to_keep"`
	EncryptionEnabled    types.Bool   `tfsdk:"encryption_enabled"`
	EncryptionPasswordID types.String `tfsdk:"encryption_password_id"`
	TriggerOnApply       types.Bool   `tfsdk:"trigger_on_apply"`
	LastSessionID        types.String `tfsdk:"last_session_id"`
	LastSessionState     types.String `tfsdk:"last_session_state"`
	LastSessionResult    types.String `tfsdk:"last_session_result"`
}

func (r *ConfigurationBackup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration_backup"
}

func (r *ConfigurationBackup) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam configuration backup settings and can trigger a configuration backup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled":                schema.BoolAttribute{Required: true},
			"repository_id":          schema.StringAttribute{Optional: true},
			"restore_points_to_keep": schema.Int64Attribute{Optional: true},
			"encryption_enabled":     schema.BoolAttribute{Optional: true},
			"encryption_password_id": schema.StringAttribute{Optional: true},
			"trigger_on_apply":       schema.BoolAttribute{Optional: true},
			"last_session_id":        schema.StringAttribute{Computed: true},
			"last_session_state":     schema.StringAttribute{Computed: true},
			"last_session_result":    schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ConfigurationBackup) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider, got unexpected type.")
		return
	}
	r.client = c
}

func (r *ConfigurationBackup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure backup settings", fmt.Sprintf("API error: %s", err))
		return
	}

	if !data.TriggerOnApply.IsNull() && data.TriggerOnApply.ValueBool() {
		if err := r.triggerBackup(ctx, &data); err != nil {
			resp.Diagnostics.AddError("Failed to trigger configuration backup", fmt.Sprintf("API error: %s", err))
			return
		}
	}

	data.ID = types.StringValue("config-backup")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ConfigurationBackup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ConfigurationBackupModel
	if err := r.client.GetJSON(ctx, client.PathConfigurationBackup, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read configuration backup", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("config-backup")
	data.Enabled = types.BoolValue(result.Enabled)
	if result.RepositoryID != "" {
		data.RepositoryID = types.StringValue(result.RepositoryID)
	}
	data.RestorePointsToKeep = types.Int64Value(int64(result.RestorePointsToKeep))
	data.EncryptionEnabled = types.BoolValue(result.EncryptionEnabled)
	if result.EncryptionPasswordID != "" {
		data.EncryptionPasswordID = types.StringValue(result.EncryptionPasswordID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ConfigurationBackup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update configuration backup", fmt.Sprintf("API error: %s", err))
		return
	}

	if !data.TriggerOnApply.IsNull() && data.TriggerOnApply.ValueBool() {
		if err := r.triggerBackup(ctx, &data); err != nil {
			resp.Diagnostics.AddError("Failed to trigger configuration backup", fmt.Sprintf("API error: %s", err))
			return
		}
	}

	data.ID = types.StringValue("config-backup")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ConfigurationBackup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	disable := models.ConfigurationBackupSpec{Enabled: false}
	if err := r.client.PutJSON(ctx, client.PathConfigurationBackup, &disable, nil); err != nil {
		resp.Diagnostics.AddError("Failed to disable configuration backup", fmt.Sprintf("API error: %s", err))
	}
}

func (r *ConfigurationBackup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewConfigurationBackup() resource.Resource {
	return &ConfigurationBackup{}
}

func (r *ConfigurationBackup) putConfig(ctx context.Context, data *ConfigurationBackupModel) error {
	spec := &models.ConfigurationBackupSpec{Enabled: data.Enabled.ValueBool()}
	if !data.RepositoryID.IsNull() {
		spec.RepositoryID = data.RepositoryID.ValueString()
	}
	if !data.RestorePointsToKeep.IsNull() && !data.RestorePointsToKeep.IsUnknown() {
		spec.RestorePointsToKeep = int(data.RestorePointsToKeep.ValueInt64())
	}
	if !data.EncryptionEnabled.IsNull() {
		spec.EncryptionEnabled = data.EncryptionEnabled.ValueBool()
	}
	if !data.EncryptionPasswordID.IsNull() {
		spec.EncryptionPasswordID = data.EncryptionPasswordID.ValueString()
	}

	return r.client.PutJSON(ctx, client.PathConfigurationBackup, spec, nil)
}

func (r *ConfigurationBackup) triggerBackup(ctx context.Context, data *ConfigurationBackupModel) error {
	var session models.ConfigurationBackupSessionModel
	if err := r.client.PostJSON(ctx, client.PathConfigurationBackup, &models.ConfigurationBackupSpec{}, &session); err != nil {
		return err
	}

	if session.ID != "" {
		data.LastSessionID = types.StringValue(session.ID)
	}
	if session.State != "" {
		data.LastSessionState = types.StringValue(session.State)
	}
	if session.Result != "" {
		data.LastSessionResult = types.StringValue(session.Result)
	}

	return nil
}

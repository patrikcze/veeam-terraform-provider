package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &ProtectionGroup{}
	_ resource.ResourceWithConfigure   = &ProtectionGroup{}
	_ resource.ResourceWithImportState = &ProtectionGroup{}
)

// ProtectionGroup implements the veeam_protection_group resource.
type ProtectionGroup struct {
	client client.APIClient
}

// ProtectionGroupComputerModel is a single computer entry in the group.
type ProtectionGroupComputerModel struct {
	HostName       types.String `tfsdk:"hostname"`
	ConnectionType types.String `tfsdk:"connection_type"`
	CredentialsID  types.String `tfsdk:"credentials_id"`
}

// ProtectionGroupOptionsModel is the deployment/options block for the protection group.
type ProtectionGroupOptionsModel struct {
	DistributionServerID      types.String `tfsdk:"distribution_server_id"`
	DistributionRepositoryID  types.String `tfsdk:"distribution_repository_id"`
	InstallBackupAgent        types.Bool   `tfsdk:"install_backup_agent"`
	InstallCBTDriver          types.Bool   `tfsdk:"install_cbt_driver"`
	InstallApplicationPlugins types.Bool   `tfsdk:"install_application_plugins"`
	ApplicationPlugins        types.List   `tfsdk:"application_plugins"`
	UpdateAutomatically       types.Bool   `tfsdk:"update_automatically"`
	RebootIfRequired          types.Bool   `tfsdk:"reboot_if_required"`
}

// ProtectionGroupModel is the Terraform state model.
type ProtectionGroupModel struct {
	ID          types.String                   `tfsdk:"id"`
	Name        types.String                   `tfsdk:"name"`
	Description types.String                   `tfsdk:"description"`
	Type        types.String                   `tfsdk:"type"`
	IsDisabled  types.Bool                     `tfsdk:"is_disabled"`
	Computers   []ProtectionGroupComputerModel `tfsdk:"computers"`
	Options     []ProtectionGroupOptionsModel  `tfsdk:"options"`
}

func (r *ProtectionGroup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_group"
}

func (r *ProtectionGroup) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam agent protection group (IndividualComputers).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Protection group identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Protection group name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Protection group type: `IndividualComputers`, `CloudMachines`, etc.",
				Required:            true,
			},
			"is_disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the protection group is disabled.",
				Optional:            true,
				Computed:            true,
			},
			"computers": schema.ListNestedAttribute{
				MarkdownDescription: "List of computers in the protection group.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"hostname": schema.StringAttribute{
							MarkdownDescription: "FQDN or IP address of the computer.",
							Required:            true,
						},
						"connection_type": schema.StringAttribute{
							MarkdownDescription: "Connection type: `PermanentCredentials`, `SingleUseCredentials`, or `Certificate`.",
							Required:            true,
						},
						"credentials_id": schema.StringAttribute{
							MarkdownDescription: "Credential ID for the computer. Required with `PermanentCredentials`.",
							Optional:            true,
						},
					},
				},
			},
			"options": schema.ListNestedAttribute{
				MarkdownDescription: "Optional deployment options for backup agents in this protection group.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"distribution_server_id":      schema.StringAttribute{Optional: true, Computed: true},
						"distribution_repository_id":  schema.StringAttribute{Optional: true, Computed: true},
						"install_backup_agent":        schema.BoolAttribute{Optional: true, Computed: true},
						"install_cbt_driver":          schema.BoolAttribute{Optional: true, Computed: true},
						"install_application_plugins": schema.BoolAttribute{Optional: true, Computed: true},
						"application_plugins":         schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
						"update_automatically":        schema.BoolAttribute{Optional: true, Computed: true},
						"reboot_if_required":          schema.BoolAttribute{Optional: true, Computed: true},
					},
				},
			},
		},
	}
}

func (r *ProtectionGroup) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected client.APIClient from provider, got unexpected type.",
		)
		return
	}
	r.client = c
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *ProtectionGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtectionGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateProtectionGroupPlan(&data); err != nil {
		resp.Diagnostics.AddError("Invalid protection group configuration", err.Error())
		return
	}

	payload := r.buildCreateSpec(&data)

	var result models.IndividualComputersProtectionGroupModel
	if err := r.client.PostJSON(ctx, client.PathProtectionGroups, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create protection group",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncFromAPI(&data, &result)

	if !data.IsDisabled.IsNull() && data.IsDisabled.ValueBool() {
		disableEndpoint := fmt.Sprintf(client.PathProtectionGroupDisable, data.ID.ValueString())
		if err := r.client.PostJSON(ctx, disableEndpoint, nil, nil); err != nil {
			resp.Diagnostics.AddError(
				"Failed to disable protection group",
				fmt.Sprintf("Protection group was created but disable request failed: %s", err),
			)
			return
		}

		if err := r.readProtectionGroup(ctx, &data); err != nil {
			resp.Diagnostics.AddError(
				"Failed to refresh protection group",
				fmt.Sprintf("Protection group was created but refresh after disable failed: %s", err),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ProtectionGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtectionGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readProtectionGroup(ctx, &data); err != nil {
		if isProtectionGroupNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Failed to read protection group",
			fmt.Sprintf("API error for group %s: %s", data.ID.ValueString(), err),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ProtectionGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProtectionGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProtectionGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateProtectionGroupPlan(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid protection group configuration", err.Error())
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		plan.ID = state.ID
	}

	payload := r.buildUpdateModel(&plan)

	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, plan.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update protection group",
			fmt.Sprintf("API error for group %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	if !plan.IsDisabled.IsNull() && !state.IsDisabled.IsNull() && plan.IsDisabled.ValueBool() != state.IsDisabled.ValueBool() {
		stateEndpoint := client.PathProtectionGroupEnable
		action := "enable"
		if plan.IsDisabled.ValueBool() {
			stateEndpoint = client.PathProtectionGroupDisable
			action = "disable"
		}

		if err := r.client.PostJSON(ctx, fmt.Sprintf(stateEndpoint, plan.ID.ValueString()), nil, nil); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Failed to %s protection group", action),
				fmt.Sprintf("Protection group %s was updated but %s action failed: %s", plan.ID.ValueString(), action, err),
			)
			return
		}
	}

	if err := r.readProtectionGroup(ctx, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Failed to refresh protection group",
			fmt.Sprintf("Protection group %s was updated but refresh failed: %s", plan.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProtectionGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProtectionGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete protection group",
			fmt.Sprintf("API error for group %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *ProtectionGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewProtectionGroup returns a new veeam_protection_group resource instance.
func NewProtectionGroup() resource.Resource {
	return &ProtectionGroup{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *ProtectionGroup) buildCreateSpec(data *ProtectionGroupModel) interface{} {
	base := models.ProtectionGroupSpec{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        models.EProtectionGroupType(data.Type.ValueString()),
	}

	return &models.IndividualComputersProtectionGroupSpec{
		ProtectionGroupSpec: base,
		Computers:           buildProtectionGroupComputers(data.Computers),
		Options:             buildProtectionGroupOptions(data.Options),
	}
}

func (r *ProtectionGroup) buildUpdateModel(data *ProtectionGroupModel) interface{} {
	return &models.IndividualComputersProtectionGroupModel{
		ProtectionGroupModel: models.ProtectionGroupModel{
			ID:          data.ID.ValueString(),
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			Type:        models.EProtectionGroupType(data.Type.ValueString()),
			IsDisabled:  !data.IsDisabled.IsNull() && data.IsDisabled.ValueBool(),
		},
		Computers: buildProtectionGroupComputers(data.Computers),
		Options:   buildProtectionGroupOptions(data.Options),
	}
}

func (r *ProtectionGroup) syncFromAPI(data *ProtectionGroupModel, api *models.IndividualComputersProtectionGroupModel) {
	if api.Name != "" {
		data.Name = types.StringValue(api.Name)
	}
	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}
	if string(api.Type) != "" {
		data.Type = types.StringValue(string(api.Type))
	}
	data.IsDisabled = types.BoolValue(api.IsDisabled)

	if len(api.Computers) > 0 {
		computers := make([]ProtectionGroupComputerModel, 0, len(api.Computers))
		for _, item := range api.Computers {
			computer := ProtectionGroupComputerModel{
				HostName:       types.StringValue(item.HostName),
				ConnectionType: types.StringValue(string(item.ConnectionType)),
				CredentialsID:  types.StringNull(),
			}
			if item.CredentialsID != "" {
				computer.CredentialsID = types.StringValue(item.CredentialsID)
			}
			computers = append(computers, computer)
		}
		data.Computers = computers
	}

	if api.Options != nil {
		options := ProtectionGroupOptionsModel{
			DistributionServerID:      types.StringNull(),
			DistributionRepositoryID:  types.StringNull(),
			InstallBackupAgent:        types.BoolValue(api.Options.InstallBackupAgent),
			InstallCBTDriver:          types.BoolValue(api.Options.InstallCBTDriver),
			InstallApplicationPlugins: types.BoolValue(api.Options.InstallApplicationPlugins),
			UpdateAutomatically:       types.BoolValue(api.Options.UpdateAutomatically),
			RebootIfRequired:          types.BoolValue(api.Options.RebootIfRequired),
			ApplicationPlugins:        types.ListNull(types.StringType),
		}

		if api.Options.DistributionServerID != "" {
			options.DistributionServerID = types.StringValue(api.Options.DistributionServerID)
		}
		if api.Options.DistributionRepositoryID != "" {
			options.DistributionRepositoryID = types.StringValue(api.Options.DistributionRepositoryID)
		}

		if len(api.Options.ApplicationPlugins) > 0 {
			values := make([]attr.Value, 0, len(api.Options.ApplicationPlugins))
			for _, plugin := range api.Options.ApplicationPlugins {
				values = append(values, types.StringValue(plugin))
			}
			options.ApplicationPlugins = types.ListValueMust(types.StringType, values)
		}

		data.Options = []ProtectionGroupOptionsModel{options}
	}
}

func (r *ProtectionGroup) readProtectionGroup(ctx context.Context, data *ProtectionGroupModel) error {
	var result models.IndividualComputersProtectionGroupModel
	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		return err
	}

	r.syncFromAPI(data, &result)
	return nil
}

func validateProtectionGroupPlan(data *ProtectionGroupModel) error {
	if data == nil {
		return fmt.Errorf("protection group plan is empty")
	}

	if !strings.EqualFold(data.Type.ValueString(), string(models.ProtectionGroupTypeIndividualComputers)) {
		return fmt.Errorf("only protection group type %q is currently supported by this resource", models.ProtectionGroupTypeIndividualComputers)
	}

	if len(data.Computers) == 0 {
		return fmt.Errorf("at least one computers block is required for type %q", models.ProtectionGroupTypeIndividualComputers)
	}

	for index, item := range data.Computers {
		if item.HostName.IsNull() || item.HostName.IsUnknown() || strings.TrimSpace(item.HostName.ValueString()) == "" {
			return fmt.Errorf("computers[%d].hostname must be set", index)
		}

		connectionType := strings.TrimSpace(item.ConnectionType.ValueString())
		switch connectionType {
		case string(models.IndividualComputerConnectionTypePermanentCredentials):
			if item.CredentialsID.IsNull() || item.CredentialsID.IsUnknown() || strings.TrimSpace(item.CredentialsID.ValueString()) == "" {
				return fmt.Errorf("computers[%d].credentials_id is required when connection_type is %q", index, connectionType)
			}
		case string(models.IndividualComputerConnectionTypeCertificate):
			if !item.CredentialsID.IsNull() && strings.TrimSpace(item.CredentialsID.ValueString()) != "" {
				return fmt.Errorf("computers[%d].credentials_id must be omitted when connection_type is %q", index, connectionType)
			}
		case string(models.IndividualComputerConnectionTypeSingleUseCredentials):
			return fmt.Errorf("computers[%d].connection_type %q is not implemented yet in Terraform schema; use %q", index, connectionType, models.IndividualComputerConnectionTypePermanentCredentials)
		default:
			return fmt.Errorf("computers[%d].connection_type must be one of %q, %q, %q", index, models.IndividualComputerConnectionTypePermanentCredentials, models.IndividualComputerConnectionTypeSingleUseCredentials, models.IndividualComputerConnectionTypeCertificate)
		}
	}

	if len(data.Options) > 1 {
		return fmt.Errorf("only one options block is allowed")
	}

	if len(data.Options) == 1 {
		option := data.Options[0]
		if !option.InstallBackupAgent.IsNull() && option.InstallBackupAgent.ValueBool() {
			hasDistributionServer := !option.DistributionServerID.IsNull() && !option.DistributionServerID.IsUnknown() && strings.TrimSpace(option.DistributionServerID.ValueString()) != ""
			hasDistributionRepo := !option.DistributionRepositoryID.IsNull() && !option.DistributionRepositoryID.IsUnknown() && strings.TrimSpace(option.DistributionRepositoryID.ValueString()) != ""
			if !hasDistributionServer && !hasDistributionRepo {
				return fmt.Errorf("options.distribution_server_id or options.distribution_repository_id must be set when options.install_backup_agent is true")
			}
		}
	}

	return nil
}

func buildProtectionGroupComputers(computers []ProtectionGroupComputerModel) []models.ProtectionGroupComputer {
	out := make([]models.ProtectionGroupComputer, 0, len(computers))
	for _, item := range computers {
		entry := models.ProtectionGroupComputer{
			HostName:       item.HostName.ValueString(),
			ConnectionType: models.EIndividualComputerConnectionType(item.ConnectionType.ValueString()),
		}
		if !item.CredentialsID.IsNull() && !item.CredentialsID.IsUnknown() {
			entry.CredentialsID = item.CredentialsID.ValueString()
		}
		out = append(out, entry)
	}
	return out
}

func buildProtectionGroupOptions(options []ProtectionGroupOptionsModel) *models.ProtectionGroupOptions {
	if len(options) == 0 {
		return nil
	}

	option := options[0]
	built := &models.ProtectionGroupOptions{}

	if !option.DistributionServerID.IsNull() && !option.DistributionServerID.IsUnknown() {
		built.DistributionServerID = option.DistributionServerID.ValueString()
	}
	if !option.DistributionRepositoryID.IsNull() && !option.DistributionRepositoryID.IsUnknown() {
		built.DistributionRepositoryID = option.DistributionRepositoryID.ValueString()
	}
	if !option.InstallBackupAgent.IsNull() {
		built.InstallBackupAgent = option.InstallBackupAgent.ValueBool()
	}
	if !option.InstallCBTDriver.IsNull() {
		built.InstallCBTDriver = option.InstallCBTDriver.ValueBool()
	}
	if !option.InstallApplicationPlugins.IsNull() {
		built.InstallApplicationPlugins = option.InstallApplicationPlugins.ValueBool()
	}
	if !option.UpdateAutomatically.IsNull() {
		built.UpdateAutomatically = option.UpdateAutomatically.ValueBool()
	}
	if !option.RebootIfRequired.IsNull() {
		built.RebootIfRequired = option.RebootIfRequired.ValueBool()
	}

	if !option.ApplicationPlugins.IsNull() && !option.ApplicationPlugins.IsUnknown() {
		plugins := make([]string, 0, len(option.ApplicationPlugins.Elements()))
		for _, value := range option.ApplicationPlugins.Elements() {
			stringValue, ok := value.(types.String)
			if !ok || stringValue.IsNull() || stringValue.IsUnknown() {
				continue
			}
			plugins = append(plugins, stringValue.ValueString())
		}
		built.ApplicationPlugins = plugins
	}

	return built
}

func isProtectionGroupNotFound(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		if strings.EqualFold(apiErr.ErrorCode, "NotFound") {
			return true
		}
	}

	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "http 404") || strings.Contains(errText, "notfound")
}

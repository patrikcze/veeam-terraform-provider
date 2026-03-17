package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

// ProtectionGroupCloudAccountModel is the cloud account selector for CloudMachines groups.
type ProtectionGroupCloudAccountModel struct {
	AccountType    types.String `tfsdk:"account_type"`
	CredentialsID  types.String `tfsdk:"credentials_id"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
	RegionType     types.String `tfsdk:"region_type"`
	RegionID       types.String `tfsdk:"region_id"`
	AssignIAMRole  types.Bool   `tfsdk:"assign_iam_role"`
}

// ProtectionGroupCloudMachineModel is a cloud object selector for CloudMachines groups.
type ProtectionGroupCloudMachineModel struct {
	Type     types.String `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	ObjectID types.String `tfsdk:"object_id"`
	Value    types.String `tfsdk:"value"`
}

// ProtectionGroupModel is the Terraform state model.
type ProtectionGroupModel struct {
	ID            types.String                       `tfsdk:"id"`
	Name          types.String                       `tfsdk:"name"`
	Description   types.String                       `tfsdk:"description"`
	Type          types.String                       `tfsdk:"type"`
	IsDisabled    types.Bool                         `tfsdk:"is_disabled"`
	Computers     []ProtectionGroupComputerModel     `tfsdk:"computers"`
	CloudAccount  []ProtectionGroupCloudAccountModel `tfsdk:"cloud_account"`
	CloudMachines []ProtectionGroupCloudMachineModel `tfsdk:"cloud_machines"`
	Options       []ProtectionGroupOptionsModel      `tfsdk:"options"`
}

func (r *ProtectionGroup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_group"
}

func (r *ProtectionGroup) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam agent protection group (IndividualComputers or CloudMachines).",
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
				Optional:            true,
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
			"cloud_account": schema.ListNestedAttribute{
				MarkdownDescription: "Cloud account settings for type `CloudMachines`.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"account_type":    schema.StringAttribute{Required: true},
						"credentials_id":  schema.StringAttribute{Optional: true},
						"subscription_id": schema.StringAttribute{Optional: true},
						"region_type":     schema.StringAttribute{Optional: true},
						"region_id":       schema.StringAttribute{Optional: true},
						"assign_iam_role": schema.BoolAttribute{Optional: true},
					},
				},
			},
			"cloud_machines": schema.ListNestedAttribute{
				MarkdownDescription: "Cloud object selectors for type `CloudMachines`.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":      schema.StringAttribute{Required: true},
						"name":      schema.StringAttribute{Optional: true},
						"object_id": schema.StringAttribute{Optional: true},
						"value":     schema.StringAttribute{Optional: true},
					},
				},
			},
			"options": schema.ListNestedAttribute{
				MarkdownDescription: "Optional deployment options for backup agents in this protection group.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"distribution_server_id":      schema.StringAttribute{Optional: true},
						"distribution_repository_id":  schema.StringAttribute{Optional: true},
						"install_backup_agent":        schema.BoolAttribute{Optional: true},
						"install_cbt_driver":          schema.BoolAttribute{Optional: true},
						"install_application_plugins": schema.BoolAttribute{Optional: true},
						"application_plugins":         schema.ListAttribute{Optional: true, ElementType: types.StringType},
						"update_automatically":        schema.BoolAttribute{Optional: true},
						"reboot_if_required":          schema.BoolAttribute{Optional: true},
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

	var createResult map[string]interface{}
	if err := r.client.PostJSON(ctx, client.PathProtectionGroups, r.buildCreateSpec(&data), &createResult); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create protection group",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	if isAsyncProtectionGroupOperationResult(createResult) {
		sessionID := getStringValue(createResult, "id")
		if sessionID == "" {
			resp.Diagnostics.AddError(
				"Failed to create protection group",
				"API response did not include async session ID.",
			)
			return
		}

		if err := r.client.WaitForTask(ctx, sessionID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to create protection group",
				fmt.Sprintf("Async protection group creation task %s failed: %s", sessionID, err),
			)
			return
		}
	}

	resolvedID, err := r.findProtectionGroupIDByName(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve created protection group",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(resolvedID)

	if err := r.readProtectionGroup(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to refresh protection group",
			fmt.Sprintf("Protection group was created but refresh failed: %s", err),
		)
		return
	}

	if len(data.Options) == 0 {
		data.Options = nil
	}

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
	var updateResult map[string]interface{}
	if err := r.client.PutJSON(ctx, endpoint, payload, &updateResult); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update protection group",
			fmt.Sprintf("API error for group %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	if isAsyncProtectionGroupOperationResult(updateResult) {
		sessionID := getStringValue(updateResult, "id")
		if sessionID == "" {
			resp.Diagnostics.AddError(
				"Failed to update protection group",
				"API response did not include async session ID for update operation.",
			)
			return
		}

		if err := r.client.WaitForTask(ctx, sessionID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update protection group",
				fmt.Sprintf("Async protection group update task %s failed: %s", sessionID, err),
			)
			return
		}
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

	if err := r.waitForProtectionGroupDeleted(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Failed to confirm protection group deletion",
			fmt.Sprintf("Protection group %s delete request was accepted but resource still appears present: %s", data.ID.ValueString(), err),
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

	if strings.EqualFold(data.Type.ValueString(), string(models.ProtectionGroupTypeCloudMachines)) {
		return &models.CloudMachinesProtectionGroupSpec{
			ProtectionGroupSpec: base,
			CloudAccount:        buildProtectionGroupCloudAccount(data.CloudAccount),
			CloudMachines:       buildProtectionGroupCloudMachines(data.CloudMachines),
			Options:             buildProtectionGroupOptions(data.Options),
		}
	}

	return &models.IndividualComputersProtectionGroupSpec{
		ProtectionGroupSpec: base,
		Computers:           buildProtectionGroupComputers(data.Computers),
		Options:             buildProtectionGroupOptions(data.Options),
	}
}

func (r *ProtectionGroup) buildUpdateModel(data *ProtectionGroupModel) interface{} {
	if strings.EqualFold(data.Type.ValueString(), string(models.ProtectionGroupTypeCloudMachines)) {
		return &models.CloudMachinesProtectionGroupModel{
			ProtectionGroupModel: models.ProtectionGroupModel{
				ID:          data.ID.ValueString(),
				Name:        data.Name.ValueString(),
				Description: data.Description.ValueString(),
				Type:        models.EProtectionGroupType(data.Type.ValueString()),
				IsDisabled:  !data.IsDisabled.IsNull() && data.IsDisabled.ValueBool(),
			},
			CloudAccount:  buildProtectionGroupCloudAccount(data.CloudAccount),
			CloudMachines: buildProtectionGroupCloudMachines(data.CloudMachines),
			Options:       buildProtectionGroupOptions(data.Options),
		}
	}

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

func (r *ProtectionGroup) syncFromAPIIndividual(data *ProtectionGroupModel, api *models.IndividualComputersProtectionGroupModel) {
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

	data.CloudAccount = nil
	data.CloudMachines = nil

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

	// Keep options null in Terraform state when user did not configure the
	// options block. VBR may auto-populate defaults, but writing them into a
	// non-computed optional field causes apply consistency errors.
	if len(data.Options) > 0 && api.Options != nil {
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

func (r *ProtectionGroup) syncFromAPICloud(data *ProtectionGroupModel, api *models.CloudMachinesProtectionGroupModel) {
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
	data.Computers = nil

	if api.CloudAccount != nil {
		account := ProtectionGroupCloudAccountModel{
			AccountType:    types.StringValue(string(api.CloudAccount.AccountType)),
			CredentialsID:  types.StringNull(),
			SubscriptionID: types.StringNull(),
			RegionType:     types.StringNull(),
			RegionID:       types.StringNull(),
			AssignIAMRole:  types.BoolNull(),
		}
		if api.CloudAccount.CredentialsID != "" {
			account.CredentialsID = types.StringValue(api.CloudAccount.CredentialsID)
		}
		if api.CloudAccount.SubscriptionID != "" {
			account.SubscriptionID = types.StringValue(api.CloudAccount.SubscriptionID)
		}
		if api.CloudAccount.RegionType != "" {
			account.RegionType = types.StringValue(api.CloudAccount.RegionType)
		}
		if api.CloudAccount.RegionID != "" {
			account.RegionID = types.StringValue(api.CloudAccount.RegionID)
		}
		if api.CloudAccount.AssignIAMRole {
			account.AssignIAMRole = types.BoolValue(true)
		}
		data.CloudAccount = []ProtectionGroupCloudAccountModel{account}
	}

	if len(api.CloudMachines) > 0 {
		items := make([]ProtectionGroupCloudMachineModel, 0, len(api.CloudMachines))
		for _, entry := range api.CloudMachines {
			item := ProtectionGroupCloudMachineModel{
				Type:     types.StringValue(string(entry.Type)),
				Name:     types.StringNull(),
				ObjectID: types.StringNull(),
				Value:    types.StringNull(),
			}
			if entry.Name != "" {
				item.Name = types.StringValue(entry.Name)
			}
			if entry.ObjectID != "" {
				item.ObjectID = types.StringValue(entry.ObjectID)
			}
			if entry.Value != "" {
				item.Value = types.StringValue(entry.Value)
			}
			items = append(items, item)
		}
		data.CloudMachines = items
	}

	if len(data.Options) > 0 && api.Options != nil {
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
	typeValue := strings.TrimSpace(data.Type.ValueString())
	if data.Type.IsNull() || data.Type.IsUnknown() || typeValue == "" {
		var probe map[string]interface{}
		endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
		if err := r.client.GetJSON(ctx, endpoint, &probe); err != nil {
			return err
		}
		typeValue = getStringValue(probe, "type")
	}

	if strings.EqualFold(typeValue, string(models.ProtectionGroupTypeCloudMachines)) {
		var result models.CloudMachinesProtectionGroupModel
		endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
		if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
			return err
		}

		r.syncFromAPICloud(data, &result)
		return nil
	}

	var result models.IndividualComputersProtectionGroupModel
	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		return err
	}

	r.syncFromAPIIndividual(data, &result)
	return nil
}

func (r *ProtectionGroup) findProtectionGroupIDByName(ctx context.Context, data *ProtectionGroupModel) (string, error) {
	var payload map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathProtectionGroups, &payload); err != nil {
		return "", fmt.Errorf("failed to list protection groups after create: %w", err)
	}

	rawData, ok := payload["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected protection groups list response shape: missing data array")
	}

	for _, item := range rawData {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entryName := getStringValue(entry, "name")
		if !strings.EqualFold(entryName, data.Name.ValueString()) {
			continue
		}

		entryType := getStringValue(entry, "type")
		if entryType != "" && !strings.EqualFold(entryType, data.Type.ValueString()) {
			continue
		}

		id := getStringValue(entry, "id")
		if id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf("protection group %q was created but could not be located in protection group list", data.Name.ValueString())
}

func validateProtectionGroupPlan(data *ProtectionGroupModel) error {
	if data == nil {
		return fmt.Errorf("protection group plan is empty")
	}

	if data.Type.IsNull() || data.Type.IsUnknown() || strings.TrimSpace(data.Type.ValueString()) == "" {
		return fmt.Errorf("type must be set")
	}

	typeValue := strings.TrimSpace(data.Type.ValueString())
	if !strings.EqualFold(typeValue, string(models.ProtectionGroupTypeIndividualComputers)) &&
		!strings.EqualFold(typeValue, string(models.ProtectionGroupTypeCloudMachines)) {
		return fmt.Errorf("type must be one of %q or %q", models.ProtectionGroupTypeIndividualComputers, models.ProtectionGroupTypeCloudMachines)
	}

	if strings.EqualFold(typeValue, string(models.ProtectionGroupTypeIndividualComputers)) {
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
	}

	if strings.EqualFold(typeValue, string(models.ProtectionGroupTypeCloudMachines)) {
		if len(data.CloudAccount) != 1 {
			return fmt.Errorf("exactly one cloud_account block is required for type %q", models.ProtectionGroupTypeCloudMachines)
		}

		account := data.CloudAccount[0]
		if account.AccountType.IsNull() || account.AccountType.IsUnknown() || strings.TrimSpace(account.AccountType.ValueString()) == "" {
			return fmt.Errorf("cloud_account[0].account_type must be set")
		}
		if account.RegionType.IsNull() || account.RegionType.IsUnknown() || strings.TrimSpace(account.RegionType.ValueString()) == "" {
			return fmt.Errorf("cloud_account[0].region_type must be set")
		}
		if account.RegionID.IsNull() || account.RegionID.IsUnknown() || strings.TrimSpace(account.RegionID.ValueString()) == "" {
			return fmt.Errorf("cloud_account[0].region_id must be set")
		}

		accountType := strings.TrimSpace(account.AccountType.ValueString())
		switch accountType {
		case string(models.ProtectionGroupCloudAccountTypeAWS):
			if account.CredentialsID.IsNull() || account.CredentialsID.IsUnknown() || strings.TrimSpace(account.CredentialsID.ValueString()) == "" {
				return fmt.Errorf("cloud_account[0].credentials_id is required when account_type is %q", models.ProtectionGroupCloudAccountTypeAWS)
			}
		case string(models.ProtectionGroupCloudAccountTypeAzure):
			if account.SubscriptionID.IsNull() || account.SubscriptionID.IsUnknown() || strings.TrimSpace(account.SubscriptionID.ValueString()) == "" {
				return fmt.Errorf("cloud_account[0].subscription_id is required when account_type is %q", models.ProtectionGroupCloudAccountTypeAzure)
			}
		default:
			return fmt.Errorf("cloud_account[0].account_type must be one of %q or %q", models.ProtectionGroupCloudAccountTypeAWS, models.ProtectionGroupCloudAccountTypeAzure)
		}

		if len(data.CloudMachines) == 0 {
			return fmt.Errorf("at least one cloud_machines block is required for type %q", models.ProtectionGroupTypeCloudMachines)
		}

		for index, item := range data.CloudMachines {
			if item.Type.IsNull() || item.Type.IsUnknown() || strings.TrimSpace(item.Type.ValueString()) == "" {
				return fmt.Errorf("cloud_machines[%d].type must be set", index)
			}

			machineType := strings.TrimSpace(item.Type.ValueString())
			switch machineType {
			case string(models.CloudMachinesObjectTypeMachine), string(models.CloudMachinesObjectTypeRegion):
				if item.ObjectID.IsNull() || item.ObjectID.IsUnknown() || strings.TrimSpace(item.ObjectID.ValueString()) == "" {
					return fmt.Errorf("cloud_machines[%d].object_id must be set when type is %q", index, machineType)
				}
			case string(models.CloudMachinesObjectTypeTag):
				if item.Name.IsNull() || item.Name.IsUnknown() || strings.TrimSpace(item.Name.ValueString()) == "" {
					return fmt.Errorf("cloud_machines[%d].name must be set when type is %q", index, machineType)
				}
				if item.Value.IsNull() || item.Value.IsUnknown() || strings.TrimSpace(item.Value.ValueString()) == "" {
					return fmt.Errorf("cloud_machines[%d].value must be set when type is %q", index, machineType)
				}
			default:
				return fmt.Errorf("cloud_machines[%d].type must be one of %q, %q, %q", index, models.CloudMachinesObjectTypeMachine, models.CloudMachinesObjectTypeRegion, models.CloudMachinesObjectTypeTag)
			}
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

func buildProtectionGroupCloudAccount(accounts []ProtectionGroupCloudAccountModel) *models.CloudMachinesAccount {
	if len(accounts) == 0 {
		return nil
	}

	item := accounts[0]
	account := &models.CloudMachinesAccount{
		AccountType: models.EProtectionGroupCloudAccountType(item.AccountType.ValueString()),
	}

	if !item.CredentialsID.IsNull() && !item.CredentialsID.IsUnknown() {
		account.CredentialsID = item.CredentialsID.ValueString()
	}
	if !item.SubscriptionID.IsNull() && !item.SubscriptionID.IsUnknown() {
		account.SubscriptionID = item.SubscriptionID.ValueString()
	}
	if !item.RegionType.IsNull() && !item.RegionType.IsUnknown() {
		account.RegionType = item.RegionType.ValueString()
	}
	if !item.RegionID.IsNull() && !item.RegionID.IsUnknown() {
		account.RegionID = item.RegionID.ValueString()
	}
	if !item.AssignIAMRole.IsNull() && !item.AssignIAMRole.IsUnknown() {
		account.AssignIAMRole = item.AssignIAMRole.ValueBool()
	}

	return account
}

func buildProtectionGroupCloudMachines(items []ProtectionGroupCloudMachineModel) []models.CloudMachineObject {
	out := make([]models.CloudMachineObject, 0, len(items))
	for _, item := range items {
		entry := models.CloudMachineObject{
			Type: models.ECloudMachinesObjectType(item.Type.ValueString()),
		}
		if !item.Name.IsNull() && !item.Name.IsUnknown() {
			entry.Name = item.Name.ValueString()
		}
		if !item.ObjectID.IsNull() && !item.ObjectID.IsUnknown() {
			entry.ObjectID = item.ObjectID.ValueString()
		}
		if !item.Value.IsNull() && !item.Value.IsUnknown() {
			entry.Value = item.Value.ValueString()
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

func isAsyncProtectionGroupOperationResult(result map[string]interface{}) bool {
	if len(result) == 0 {
		return false
	}

	if _, ok := result["state"]; ok {
		return true
	}

	resultType := strings.ToLower(getStringValue(result, "type"))
	if strings.Contains(resultType, "session") || strings.Contains(resultType, "infrastructure") {
		return true
	}

	if resultType == "" && getStringValue(result, "id") != "" {
		return true
	}

	return false
}

func (r *ProtectionGroup) waitForProtectionGroupDeleted(ctx context.Context, protectionGroupID string) error {
	const pollInterval = 3 * time.Second
	const timeout = 2 * time.Minute

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, protectionGroupID)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		var result models.IndividualComputersProtectionGroupModel
		err := r.client.GetJSON(pollCtx, endpoint, &result)
		if err != nil {
			if isProtectionGroupNotFound(err) {
				return nil
			}
			return err
		}

		select {
		case <-pollCtx.Done():
			return fmt.Errorf("timed out after %s", timeout)
		case <-ticker.C:
		}
	}
}

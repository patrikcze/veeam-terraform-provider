package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &ScaleOutRepository{}
	_ resource.ResourceWithConfigure   = &ScaleOutRepository{}
	_ resource.ResourceWithImportState = &ScaleOutRepository{}
)

// ScaleOutRepository implements the veeam_scale_out_repository resource.
type ScaleOutRepository struct {
	client client.APIClient
}

// ScaleOutRepositoryModel is the Terraform state model for veeam_scale_out_repository.
type ScaleOutRepositoryModel struct {
	ID                   types.String         `tfsdk:"id"`
	Name                 types.String         `tfsdk:"name"`
	Description          types.String         `tfsdk:"description"`
	PerformanceExtentIDs types.List           `tfsdk:"performance_extent_ids"`
	CapacityTierEnabled  types.Bool           `tfsdk:"capacity_tier_enabled"`
	PlacementPolicy      *SOBRPlacementPolicy `tfsdk:"placement_policy"`
}

// SOBRPlacementPolicy maps to PlacementPolicyModel.
type SOBRPlacementPolicy struct {
	// Type is the placement strategy for distributing data across SOBR extents.
	// Allowed values: DataLocality, Performance.
	Type types.String `tfsdk:"type"`
}

func (r *ScaleOutRepository) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scale_out_repository"
}

func (r *ScaleOutRepository) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam scale-out backup repository (SOBR).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Scale-out repository identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SOBR name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"performance_extent_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Ordered list of repository IDs to include as performance extents. At least one is required.",
			},
			"capacity_tier_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable the capacity tier (requires object storage configured in Veeam).",
			},
			"placement_policy": schema.SingleNestedAttribute{
				MarkdownDescription: "Data placement policy that controls how Veeam distributes " +
					"backup data across the SOBR performance extents. " +
					"When omitted, the server applies its default placement strategy.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Placement strategy. " +
							"Allowed values: `DataLocality` (keeps backup chains on the same extent " +
							"as previous restore points to improve deduplication), " +
							"`Performance` (distributes chains across all extents for maximum throughput).",
						Required: true,
					},
				},
			},
		},
	}
}

func (r *ScaleOutRepository) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *ScaleOutRepository) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScaleOutRepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := r.buildSpec(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.PostJSON(ctx, client.PathScaleOutRepositories, payload, &result); err != nil {
		resp.Diagnostics.AddError("Failed to create scale-out repository", fmt.Sprintf("API error: %s", err))
		return
	}

	resultID := getStringValue(result, "id")
	resultType := getStringValue(result, "type")

	// POST returns a SessionModel (async) — wait for task and resolve SOBR ID by name.
	if resultType == "" {
		if resultID == "" {
			resp.Diagnostics.AddError("Failed to create scale-out repository",
				"API response did not include a type or async session ID.")
			return
		}
		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError("Failed to create scale-out repository",
				fmt.Sprintf("Async task %s failed: %s", resultID, err))
			return
		}
		resolvedID, err := r.findSOBRIDByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve created scale-out repository", err.Error())
			return
		}
		data.ID = types.StringValue(resolvedID)
	} else {
		data.ID = types.StringValue(resultID)
	}

	// Read back to populate computed fields.
	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		var created models.ScaleOutRepositoryModel
		endpoint := fmt.Sprintf(client.PathScaleOutRepositoryByID, data.ID.ValueString())
		if err := r.client.GetJSON(ctx, endpoint, &created); err == nil {
			resp.Diagnostics.Append(r.syncFromAPI(ctx, &data, &created)...)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ScaleOutRepository) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScaleOutRepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ScaleOutRepositoryModel
	endpoint := fmt.Sprintf(client.PathScaleOutRepositoryByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		if isRepositoryNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read scale-out repository",
			fmt.Sprintf("API error for SOBR %s: %s", data.ID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(r.syncFromAPI(ctx, &data, &result)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ScaleOutRepository) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScaleOutRepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := r.buildSpec(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathScaleOutRepositoryByID, data.ID.ValueString())
	var result map[string]interface{}
	if err := r.client.PutJSON(ctx, endpoint, payload, &result); err != nil {
		resp.Diagnostics.AddError("Failed to update scale-out repository",
			fmt.Sprintf("API error for SOBR %s: %s", data.ID.ValueString(), err))
		return
	}

	// Handle async response from PUT.
	resultID := getStringValue(result, "id")
	resultType := getStringValue(result, "type")
	if resultType == "" && resultID != "" {
		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError("Failed to update scale-out repository",
				fmt.Sprintf("Async task %s failed: %s", resultID, err))
			return
		}
	}

	// Read back to pick up any server-side computed changes after the PUT.
	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		var updated models.ScaleOutRepositoryModel
		if err := r.client.GetJSON(ctx, endpoint, &updated); err == nil {
			resp.Diagnostics.Append(r.syncFromAPI(ctx, &data, &updated)...)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ScaleOutRepository) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScaleOutRepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathScaleOutRepositoryByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError("Failed to delete scale-out repository",
			fmt.Sprintf("API error for SOBR %s: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *ScaleOutRepository) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewScaleOutRepository returns a new veeam_scale_out_repository resource instance.
func NewScaleOutRepository() resource.Resource {
	return &ScaleOutRepository{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *ScaleOutRepository) buildSpec(ctx context.Context, data *ScaleOutRepositoryModel) (*models.ScaleOutRepositorySpec, diag.Diagnostics) {
	var diags diag.Diagnostics

	var extentIDs []string
	diags.Append(data.PerformanceExtentIDs.ElementsAs(ctx, &extentIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	extents := make([]models.PerformanceExtentSpec, len(extentIDs))
	for i, id := range extentIDs {
		extents[i] = models.PerformanceExtentSpec{ID: id}
	}

	spec := &models.ScaleOutRepositorySpec{
		Name: data.Name.ValueString(),
		PerformanceTier: models.PerformanceTierSpec{
			PerformanceExtents: extents,
		},
	}
	if !data.Description.IsNull() {
		spec.Description = data.Description.ValueString()
	}
	if !data.CapacityTierEnabled.IsNull() && data.CapacityTierEnabled.ValueBool() {
		spec.CapacityTier = &models.CapacityTierSpec{
			IsEnabled: true,
		}
	}
	if data.PlacementPolicy != nil && !data.PlacementPolicy.Type.IsNull() &&
		data.PlacementPolicy.Type.ValueString() != "" {
		spec.PlacementPolicy = &models.PlacementPolicyModel{
			Type: models.EPlacementPolicyType(data.PlacementPolicy.Type.ValueString()),
		}
	}
	return spec, diags
}

func (r *ScaleOutRepository) syncFromAPI(ctx context.Context, data *ScaleOutRepositoryModel, api *models.ScaleOutRepositoryModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)

	// Sync capacity tier.
	if api.CapacityTier != nil {
		data.CapacityTierEnabled = types.BoolValue(api.CapacityTier.IsEnabled)
	} else {
		data.CapacityTierEnabled = types.BoolValue(false)
	}

	// Sync performance extent IDs.
	if api.PerformanceTier != nil {
		ids := make([]string, len(api.PerformanceTier.PerformanceExtents))
		for i, e := range api.PerformanceTier.PerformanceExtents {
			ids[i] = e.ID
		}
		list, d := types.ListValueFrom(ctx, types.StringType, ids)
		diags.Append(d...)
		data.PerformanceExtentIDs = list
	} else if data.PerformanceExtentIDs.IsNull() {
		data.PerformanceExtentIDs, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	// Sync placement policy.
	if api.PlacementPolicy != nil && api.PlacementPolicy.Type != "" {
		data.PlacementPolicy = &SOBRPlacementPolicy{
			Type: types.StringValue(string(api.PlacementPolicy.Type)),
		}
	}

	return diags
}

func (r *ScaleOutRepository) findSOBRIDByName(ctx context.Context, name string) (string, error) {
	var payload map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathScaleOutRepositories, &payload); err != nil {
		return "", fmt.Errorf("failed to list scale-out repositories after create: %w", err)
	}

	rawData, ok := payload["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected scale-out repositories list response: missing data array")
	}

	for _, item := range rawData {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if getStringValue(entry, "name") == name {
			id := getStringValue(entry, "id")
			if id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("scale-out repository %q was created but could not be located in list", name)
}

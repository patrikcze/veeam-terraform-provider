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
	_ resource.Resource                = &ScaleOutRepository{}
	_ resource.ResourceWithConfigure   = &ScaleOutRepository{}
	_ resource.ResourceWithImportState = &ScaleOutRepository{}
)

type ScaleOutRepository struct {
	client client.APIClient
}

type ScaleOutRepositoryModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	CapacityTierEnabled    types.Bool   `tfsdk:"capacity_tier_enabled"`
	MaintenanceModeEnabled types.Bool   `tfsdk:"maintenance_mode_enabled"`
	SealedModeEnabled      types.Bool   `tfsdk:"sealed_mode_enabled"`
}

func (r *ScaleOutRepository) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scale_out_repository"
}

func (r *ScaleOutRepository) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam scale-out backup repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":                     schema.StringAttribute{Required: true},
			"description":              schema.StringAttribute{Optional: true, Computed: true},
			"capacity_tier_enabled":    schema.BoolAttribute{Optional: true, Computed: true},
			"maintenance_mode_enabled": schema.BoolAttribute{Optional: true, Computed: true},
			"sealed_mode_enabled":      schema.BoolAttribute{Optional: true, Computed: true},
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

func (r *ScaleOutRepository) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScaleOutRepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &models.ScaleOutRepositorySpec{
		Name: data.Name.ValueString(),
	}
	if !data.Description.IsNull() {
		payload.Description = data.Description.ValueString()
	}
	if !data.CapacityTierEnabled.IsNull() {
		payload.CapacityTier = data.CapacityTierEnabled.ValueBool()
	}

	var result models.ScaleOutRepositoryModel
	if err := r.client.PostJSON(ctx, client.PathScaleOutRepositories, payload, &result); err != nil {
		resp.Diagnostics.AddError("Failed to create scale-out repository", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	r.applyModesCreate(ctx, &data, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	r.syncFromAPI(&data, &result)
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
		resp.Diagnostics.AddError("Failed to read scale-out repository", fmt.Sprintf("API error for scale-out repository %s: %s", data.ID.ValueString(), err))
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ScaleOutRepository) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScaleOutRepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &models.ScaleOutRepositorySpec{
		Name: data.Name.ValueString(),
	}
	if !data.Description.IsNull() {
		payload.Description = data.Description.ValueString()
	}
	if !data.CapacityTierEnabled.IsNull() {
		payload.CapacityTier = data.CapacityTierEnabled.ValueBool()
	}

	endpoint := fmt.Sprintf(client.PathScaleOutRepositoryByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update scale-out repository", fmt.Sprintf("API error for scale-out repository %s: %s", data.ID.ValueString(), err))
		return
	}

	r.applyModes(ctx, &data, resp)
	if resp.Diagnostics.HasError() {
		return
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
		resp.Diagnostics.AddError("Failed to delete scale-out repository", fmt.Sprintf("API error for scale-out repository %s: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *ScaleOutRepository) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewScaleOutRepository() resource.Resource {
	return &ScaleOutRepository{}
}

func (r *ScaleOutRepository) applyModes(ctx context.Context, data *ScaleOutRepositoryModel, resp *resource.UpdateResponse) {
	if data.ID.IsNull() || data.ID.IsUnknown() {
		return
	}
	id := data.ID.ValueString()

	if !data.SealedModeEnabled.IsNull() {
		var endpoint string
		if data.SealedModeEnabled.ValueBool() {
			endpoint = fmt.Sprintf(client.PathScaleOutEnableSealed, id)
		} else {
			endpoint = fmt.Sprintf(client.PathScaleOutDisableSealed, id)
		}
		if err := r.client.PostJSON(ctx, endpoint, &models.ScaleOutModeSpec{}, nil); err != nil {
			resp.Diagnostics.AddError("Failed to apply sealed mode", fmt.Sprintf("API error for scale-out repository %s: %s", id, err))
			return
		}
	}

	if !data.MaintenanceModeEnabled.IsNull() {
		var endpoint string
		if data.MaintenanceModeEnabled.ValueBool() {
			endpoint = fmt.Sprintf(client.PathScaleOutEnableMaint, id)
		} else {
			endpoint = fmt.Sprintf(client.PathScaleOutDisableMaint, id)
		}
		if err := r.client.PostJSON(ctx, endpoint, &models.ScaleOutModeSpec{}, nil); err != nil {
			resp.Diagnostics.AddError("Failed to apply maintenance mode", fmt.Sprintf("API error for scale-out repository %s: %s", id, err))
			return
		}
	}
}

func (r *ScaleOutRepository) applyModesCreate(ctx context.Context, data *ScaleOutRepositoryModel, resp *resource.CreateResponse) {
	if data.ID.IsNull() || data.ID.IsUnknown() {
		return
	}
	id := data.ID.ValueString()

	if !data.SealedModeEnabled.IsNull() {
		var endpoint string
		if data.SealedModeEnabled.ValueBool() {
			endpoint = fmt.Sprintf(client.PathScaleOutEnableSealed, id)
		} else {
			endpoint = fmt.Sprintf(client.PathScaleOutDisableSealed, id)
		}
		if err := r.client.PostJSON(ctx, endpoint, &models.ScaleOutModeSpec{}, nil); err != nil {
			resp.Diagnostics.AddError("Failed to apply sealed mode", fmt.Sprintf("API error for scale-out repository %s: %s", id, err))
			return
		}
	}

	if !data.MaintenanceModeEnabled.IsNull() {
		var endpoint string
		if data.MaintenanceModeEnabled.ValueBool() {
			endpoint = fmt.Sprintf(client.PathScaleOutEnableMaint, id)
		} else {
			endpoint = fmt.Sprintf(client.PathScaleOutDisableMaint, id)
		}
		if err := r.client.PostJSON(ctx, endpoint, &models.ScaleOutModeSpec{}, nil); err != nil {
			resp.Diagnostics.AddError("Failed to apply maintenance mode", fmt.Sprintf("API error for scale-out repository %s: %s", id, err))
			return
		}
	}
}

func (r *ScaleOutRepository) syncFromAPI(data *ScaleOutRepositoryModel, api *models.ScaleOutRepositoryModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.SealedModeEnabled = types.BoolValue(api.IsSealedModeEnabled)
	data.MaintenanceModeEnabled = types.BoolValue(api.IsMaintenanceModeEnabled)
}

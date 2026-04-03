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

// Compile-time interface checks.
var (
	_ resource.Resource                = &RecoveryToken{}
	_ resource.ResourceWithConfigure   = &RecoveryToken{}
	_ resource.ResourceWithImportState = &RecoveryToken{}
)

// RecoveryToken implements the veeam_recovery_token resource.
type RecoveryToken struct {
	client client.APIClient
}

// RecoveryTokenModel is the Terraform state model for veeam_recovery_token.
type RecoveryTokenModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ManagedServerID types.String `tfsdk:"managed_server_id"`
	TokenValue      types.String `tfsdk:"token_value"`
}

// NewRecoveryToken returns a new veeam_recovery_token resource instance.
func NewRecoveryToken() resource.Resource {
	return &RecoveryToken{}
}

func (r *RecoveryToken) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recovery_token"
}

func (r *RecoveryToken) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an agent recovery token in Veeam Backup & Replication " +
			"(`/api/v1/agents/recoveryTokens`). The `token_value` is only available " +
			"at creation time and is never returned by subsequent API reads — it is " +
			"preserved in Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Recovery token identifier (assigned by the server).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the recovery token.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description of the recovery token.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"managed_server_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the managed server (host) this token is issued for. Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The actual recovery token string. Only available immediately after creation — not returned by subsequent API reads.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RecoveryToken) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RecoveryToken) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RecoveryTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.RecoveryTokenModel
	if err := r.client.PostJSON(ctx, client.PathRecoveryTokens, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create recovery token",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	// Sync non-sensitive fields from the API response.
	r.syncModelFromAPI(&data, &result)
	// Capture token_value from the creation response — it will not appear on Read.
	if result.TokenValue != "" {
		data.TokenValue = types.StringValue(result.TokenValue)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *RecoveryToken) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RecoveryTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.RecoveryTokenModel
	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read recovery token",
			fmt.Sprintf("API error for recovery token %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Sync non-sensitive fields. token_value is preserved from state — the API
	// does not return it after the initial creation response.
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *RecoveryToken) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RecoveryTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve token_value from prior state — it is not sent on update.
	var state RecoveryTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.TokenValue = state.TokenValue

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update recovery token",
			fmt.Sprintf("API error for recovery token %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *RecoveryToken) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RecoveryTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathRecoveryTokenByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete recovery token",
			fmt.Sprintf("API error for recovery token %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *RecoveryToken) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *RecoveryToken) buildSpec(data *RecoveryTokenModel) *models.RecoveryTokenSpec {
	spec := &models.RecoveryTokenSpec{
		Name:            data.Name.ValueString(),
		ManagedServerID: data.ManagedServerID.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		spec.Description = data.Description.ValueString()
	}
	return spec
}

// syncModelFromAPI updates non-sensitive Terraform state fields from an API response.
// token_value is intentionally not updated here — it is never returned after creation.
func (r *RecoveryToken) syncModelFromAPI(data *RecoveryTokenModel, api *models.RecoveryTokenModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.ManagedServerID = types.StringValue(api.ManagedServerID)
}

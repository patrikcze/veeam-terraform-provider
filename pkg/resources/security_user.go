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
	_ resource.Resource                = &SecurityUser{}
	_ resource.ResourceWithConfigure   = &SecurityUser{}
	_ resource.ResourceWithImportState = &SecurityUser{}
)

// SecurityUser implements the veeam_security_user resource.
// Update is not supported — login and role changes require destroy+recreate.
type SecurityUser struct {
	client client.APIClient
}

// SecurityUserModel is the Terraform state model for veeam_security_user.
type SecurityUserModel struct {
	ID          types.String `tfsdk:"id"`
	Login       types.String `tfsdk:"login"`
	Password    types.String `tfsdk:"password"`
	Description types.String `tfsdk:"description"`
	Role        types.String `tfsdk:"role"`
}

func (r *SecurityUser) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_user"
}

func (r *SecurityUser) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam security user account. " +
			"Changing `login` or `role` forces a destroy and recreate of the resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Security user identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				MarkdownDescription: "Login name for the user (e.g., `DOMAIN\\username` or `username`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the user. This value is write-only and is never read back from the API.",
				Required:            true,
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description for the user.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role assigned to the user. Supported values: `PortalAdministrator`, `PortalUser`, " +
					"`PortalReadOnlyUser`, `RestoreOperator`. Changing this forces a destroy and recreate.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SecurityUser) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityUser) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: create the user account.
	userPayload := &models.SecurityUserSpec{
		Login:       data.Login.ValueString(),
		Password:    data.Password.ValueString(),
		Description: data.Description.ValueString(),
	}

	var userResult models.SecurityUserModel
	if err := r.client.PostJSON(ctx, client.PathSecurityUsers, userPayload, &userResult); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create security user",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(userResult.ID)

	// Step 2: assign the role.
	rolePayload := &models.SecurityUserRoleSpec{
		RoleName: data.Role.ValueString(),
	}
	rolesEndpoint := fmt.Sprintf(client.PathSecurityUserRoles, userResult.ID)
	if err := r.client.PutJSON(ctx, rolesEndpoint, rolePayload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to assign role to security user",
			fmt.Sprintf("API error assigning role to user %s: %s", userResult.ID, err),
		)
		return
	}

	// Step 3: read back to populate computed fields.
	r.syncModelFromAPI(&data, &userResult)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *SecurityUser) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read user details.
	var userResult models.SecurityUserModel
	endpoint := fmt.Sprintf(client.PathSecurityUserByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &userResult); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read security user",
			fmt.Sprintf("API error for security user %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Read role.
	var roleResult models.SecurityUserRoleModel
	rolesEndpoint := fmt.Sprintf(client.PathSecurityUserRoles, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, rolesEndpoint, &roleResult); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read security user role",
			fmt.Sprintf("API error reading role for user %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &userResult)
	data.Role = types.StringValue(roleResult.RoleName)
	// Password is never returned by the API — keep current state value.
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update is intentionally not implemented. The RequiresReplace plan modifiers on
// login and role ensure any change to those fields forces a destroy+recreate.
// This method must still exist to satisfy the resource.Resource interface.
func (r *SecurityUser) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// login and role have RequiresReplace modifiers, so Update is only reached
	// if the password or description changed. Those fields cannot be updated
	// through the API, so we preserve the plan state as-is.
	var data SecurityUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *SecurityUser) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathSecurityUserByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete security user",
			fmt.Sprintf("API error for security user %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *SecurityUser) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewSecurityUser returns a new veeam_security_user resource instance.
func NewSecurityUser() resource.Resource {
	return &SecurityUser{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// syncModelFromAPI updates Terraform state fields from an API response.
// Password is never overwritten — it is not returned by the API.
func (r *SecurityUser) syncModelFromAPI(data *SecurityUserModel, api *models.SecurityUserModel) {
	data.Login = types.StringValue(api.Login)
	data.Description = types.StringValue(api.Description)
}

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
	_ resource.Resource                = &EntraIDTenant{}
	_ resource.ResourceWithConfigure   = &EntraIDTenant{}
	_ resource.ResourceWithImportState = &EntraIDTenant{}
)

// EntraIDTenant implements the veeam_entra_id_tenant resource.
type EntraIDTenant struct {
	client client.APIClient
}

// EntraIDTenantModel is the Terraform state model for veeam_entra_id_tenant.
type EntraIDTenantModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	TenantID      types.String `tfsdk:"tenant_id"`
	CredentialsID types.String `tfsdk:"credentials_id"`
}

// NewEntraIDTenant returns a new veeam_entra_id_tenant resource instance.
func NewEntraIDTenant() resource.Resource {
	return &EntraIDTenant{}
}

func (r *EntraIDTenant) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entra_id_tenant"
}

func (r *EntraIDTenant) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Microsoft Entra ID (Azure AD) tenant in the Veeam inventory " +
			"(`/api/v1/inventory/entraId/tenants`). The `tenant_id` is immutable — " +
			"changing it forces a destroy and recreate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Entra ID tenant resource identifier (assigned by the server).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for this tenant entry.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description of the tenant entry.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Azure AD / Entra ID tenant GUID (e.g. `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`). Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credentials_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the OAuth2 application credential used to authenticate to this tenant.",
			},
		},
	}
}

func (r *EntraIDTenant) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EntraIDTenant) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntraIDTenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.EntraIDTenantModel
	if err := r.client.PostJSON(ctx, client.PathEntraIDTenants, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Entra ID tenant",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EntraIDTenant) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntraIDTenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.EntraIDTenantModel
	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read Entra ID tenant",
			fmt.Sprintf("API error for Entra ID tenant %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EntraIDTenant) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntraIDTenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update Entra ID tenant",
			fmt.Sprintf("API error for Entra ID tenant %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EntraIDTenant) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntraIDTenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathEntraIDTenantByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete Entra ID tenant",
			fmt.Sprintf("API error for Entra ID tenant %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *EntraIDTenant) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *EntraIDTenant) buildSpec(data *EntraIDTenantModel) *models.EntraIDTenantSpec {
	spec := &models.EntraIDTenantSpec{
		Name:          data.Name.ValueString(),
		TenantID:      data.TenantID.ValueString(),
		CredentialsID: data.CredentialsID.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		spec.Description = data.Description.ValueString()
	}
	return spec
}

func (r *EntraIDTenant) syncModelFromAPI(data *EntraIDTenantModel, api *models.EntraIDTenantModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.TenantID = types.StringValue(api.TenantID)
	data.CredentialsID = types.StringValue(api.CredentialsID)
}

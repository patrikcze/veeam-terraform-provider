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
	_ resource.Resource                = &ADDomain{}
	_ resource.ResourceWithConfigure   = &ADDomain{}
	_ resource.ResourceWithImportState = &ADDomain{}
)

// ADDomain implements the veeam_ad_domain resource.
// Update is not supported — name and username changes require destroy+recreate.
type ADDomain struct {
	client client.APIClient
}

// ADDomainModel is the Terraform state model for veeam_ad_domain.
type ADDomainModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Description types.String `tfsdk:"description"`
}

func (r *ADDomain) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ad_domain"
}

func (r *ADDomain) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Active Directory domain registration in Veeam Backup & Replication. " +
			"Changing `name` or `username` forces a destroy and recreate of the resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "AD domain identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Fully qualified domain name (e.g., `corp.example.com`). Changing this forces a destroy and recreate.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Domain administrator account (e.g., `DOMAIN\\admin`). Changing this forces a destroy and recreate.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the domain administrator account. This value is write-only and is never read back from the API.",
				Required:            true,
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description for the AD domain.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ADDomain) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ADDomain) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ADDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &models.ADDomainSpec{
		Name:        data.Name.ValueString(),
		UserName:    data.Username.ValueString(),
		Password:    data.Password.ValueString(),
		Description: data.Description.ValueString(),
	}

	var result models.ADDomainModel
	if err := r.client.PostJSON(ctx, client.PathADDomains, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create AD domain",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ADDomain) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ADDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ADDomainModel
	endpoint := fmt.Sprintf(client.PathADDomainByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read AD domain",
			fmt.Sprintf("API error for AD domain %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	// Password is never returned by the API — keep current state value.
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update is intentionally not implemented. The RequiresReplace plan modifiers on
// name and username ensure any change forces a destroy+recreate. This method
// must still exist to satisfy the resource.Resource interface.
func (r *ADDomain) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Only description or password could reach Update. Neither can be modified
	// through the API, so we preserve the plan state as-is.
	var data ADDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ADDomain) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ADDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathADDomainByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete AD domain",
			fmt.Sprintf("API error for AD domain %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *ADDomain) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewADDomain returns a new veeam_ad_domain resource instance.
func NewADDomain() resource.Resource {
	return &ADDomain{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// syncModelFromAPI updates Terraform state fields from an API response.
// Password is never overwritten — it is not returned by the API.
func (r *ADDomain) syncModelFromAPI(data *ADDomainModel, api *models.ADDomainModel) {
	data.Name = types.StringValue(api.Name)
	data.Username = types.StringValue(api.UserName)
	data.Description = types.StringValue(api.Description)
}

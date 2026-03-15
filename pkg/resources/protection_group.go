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
	HostName      types.String `tfsdk:"hostname"`
	CredentialsID types.String `tfsdk:"credentials_id"`
}

// ProtectionGroupModel is the Terraform state model.
type ProtectionGroupModel struct {
	ID          types.String                       `tfsdk:"id"`
	Name        types.String                       `tfsdk:"name"`
	Description types.String                       `tfsdk:"description"`
	Type        types.String                       `tfsdk:"type"`
	Computers   []ProtectionGroupComputerModel     `tfsdk:"computers"`
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
			"computers": schema.ListNestedAttribute{
				MarkdownDescription: "List of computers in the protection group.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"hostname": schema.StringAttribute{
							MarkdownDescription: "FQDN or IP address of the computer.",
							Required:            true,
						},
						"credentials_id": schema.StringAttribute{
							MarkdownDescription: "Credential ID for the computer.",
							Required:            true,
						},
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

	payload := r.buildSpec(&data)

	var result models.ProtectionGroupModel
	if err := r.client.PostJSON(ctx, client.PathProtectionGroups, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create protection group",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ProtectionGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtectionGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ProtectionGroupModel
	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read protection group",
			fmt.Sprintf("API error for group %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ProtectionGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProtectionGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathProtectionGroupByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update protection group",
			fmt.Sprintf("API error for group %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
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

func (r *ProtectionGroup) buildSpec(data *ProtectionGroupModel) interface{} {
	base := models.ProtectionGroupSpec{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        models.EProtectionGroupType(data.Type.ValueString()),
	}

	// IndividualComputers type
	computers := make([]models.ProtectionGroupComputer, 0, len(data.Computers))
	for _, c := range data.Computers {
		computers = append(computers, models.ProtectionGroupComputer{
			HostName:      c.HostName.ValueString(),
			CredentialsID: c.CredentialsID.ValueString(),
		})
	}

	return &models.IndividualComputersProtectionGroupSpec{
		ProtectionGroupSpec: base,
		Computers:           computers,
	}
}

func (r *ProtectionGroup) syncFromAPI(data *ProtectionGroupModel, api *models.ProtectionGroupModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(string(api.Type))
}

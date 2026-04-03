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
	_ resource.Resource                = &GlobalVMExclusion{}
	_ resource.ResourceWithConfigure   = &GlobalVMExclusion{}
	_ resource.ResourceWithImportState = &GlobalVMExclusion{}
)

// GlobalVMExclusion implements the veeam_global_vm_exclusion resource.
// Update is not supported — all meaningful fields carry RequiresReplace, so any
// change triggers a destroy+recreate. The Update method must still exist to
// satisfy the resource.Resource interface.
type GlobalVMExclusion struct {
	client client.APIClient
}

// GlobalVMExclusionModel is the Terraform state model for veeam_global_vm_exclusion.
type GlobalVMExclusionModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	HostName    types.String `tfsdk:"host_name"`
	ObjectID    types.String `tfsdk:"object_id"`
	Description types.String `tfsdk:"description"`
}

// NewGlobalVMExclusion returns a new veeam_global_vm_exclusion resource instance.
func NewGlobalVMExclusion() resource.Resource {
	return &GlobalVMExclusion{}
}

func (r *GlobalVMExclusion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_vm_exclusion"
}

func (r *GlobalVMExclusion) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global VM exclusion entry in Veeam Backup & Replication " +
			"(`/api/v1/globalExclusions/vm`). All identifying fields carry `RequiresReplace`, " +
			"so any change results in a destroy and recreate of the resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Exclusion entry identifier (assigned by the server).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the excluded object. Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Type of the excluded object. Allowed values: `VirtualMachine`, `Folder`, " +
					"`Datacenter`, `Cluster`, `Host`, `Tag`, `VirtualDisk`. Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Hostname of the vCenter or ESXi server that owns the object. Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "vSphere MoRef identifier of the object (e.g. `vm-42`). Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description of the exclusion entry.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GlobalVMExclusion) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GlobalVMExclusion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GlobalVMExclusionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.GlobalVMExclusionModel
	if err := r.client.PostJSON(ctx, client.PathGlobalVMExclusions, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create global VM exclusion",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *GlobalVMExclusion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GlobalVMExclusionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.GlobalVMExclusionModel
	endpoint := fmt.Sprintf(client.PathGlobalVMExclusionByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read global VM exclusion",
			fmt.Sprintf("API error for global VM exclusion %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update is a pass-through. Because all mutable fields carry RequiresReplace,
// Terraform will never call Update — it will always destroy and recreate instead.
func (r *GlobalVMExclusion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GlobalVMExclusionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *GlobalVMExclusion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GlobalVMExclusionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathGlobalVMExclusionByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete global VM exclusion",
			fmt.Sprintf("API error for global VM exclusion %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *GlobalVMExclusion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *GlobalVMExclusion) buildSpec(data *GlobalVMExclusionModel) *models.GlobalVMExclusionSpec {
	spec := &models.GlobalVMExclusionSpec{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueString(),
	}
	if !data.HostName.IsNull() && !data.HostName.IsUnknown() {
		spec.HostName = data.HostName.ValueString()
	}
	if !data.ObjectID.IsNull() && !data.ObjectID.IsUnknown() {
		spec.ObjectID = data.ObjectID.ValueString()
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		spec.Description = data.Description.ValueString()
	}
	return spec
}

func (r *GlobalVMExclusion) syncModelFromAPI(data *GlobalVMExclusionModel, api *models.GlobalVMExclusionModel) {
	data.Name = types.StringValue(api.Name)
	data.Type = types.StringValue(api.Type)
	data.HostName = types.StringValue(api.HostName)
	data.ObjectID = types.StringValue(api.ObjectID)
	data.Description = types.StringValue(api.Description)
}

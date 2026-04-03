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
	_ resource.Resource                = &MountServer{}
	_ resource.ResourceWithConfigure   = &MountServer{}
	_ resource.ResourceWithImportState = &MountServer{}
)

// MountServer implements the veeam_mount_server resource.
// There is no DELETE endpoint for mount servers — their lifecycle is tied to the
// managed server they are assigned to. Delete is therefore a no-op that removes
// the resource from Terraform state only.
type MountServer struct {
	client client.APIClient
}

// MountServerModel is the Terraform state model for veeam_mount_server.
type MountServerModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	ManagedServerID types.String `tfsdk:"managed_server_id"`
	Type            types.String `tfsdk:"type"`
	CredentialsID   types.String `tfsdk:"credentials_id"`
}

// NewMountServer returns a new veeam_mount_server resource instance.
func NewMountServer() resource.Resource {
	return &MountServer{}
}

func (r *MountServer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mount_server"
}

func (r *MountServer) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a mount server in the Veeam backup infrastructure " +
			"(`/api/v1/backupInfrastructure/mountServers`). Mount servers do not have a " +
			"dedicated delete endpoint — their lifecycle is bound to the managed server " +
			"they belong to. Deleting this resource removes it from Terraform state only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Mount server identifier (assigned by the server).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the mount server.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description of the mount server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"managed_server_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the managed server (host) that owns this mount server. Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Mount server type (e.g. `WinServer`, `LinuxServer`). Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credentials_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the credential used to connect to the mount server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MountServer) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MountServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MountServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.MountServerModel
	if err := r.client.PostJSON(ctx, client.PathMountServers, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create mount server",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *MountServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MountServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.MountServerModel
	endpoint := fmt.Sprintf(client.PathMountServerByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read mount server",
			fmt.Sprintf("API error for mount server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *MountServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MountServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathMountServerByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update mount server",
			fmt.Sprintf("API error for mount server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete is a no-op. The Veeam API does not provide a delete endpoint for mount
// servers — Terraform simply removes the resource from state.
func (r *MountServer) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentional no-op: mount servers cannot be deleted via the API.
	// Terraform will remove the resource from state automatically.
}

func (r *MountServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *MountServer) buildSpec(data *MountServerModel) *models.MountServerSpec {
	spec := &models.MountServerSpec{
		Name:            data.Name.ValueString(),
		ManagedServerID: data.ManagedServerID.ValueString(),
		Type:            data.Type.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		spec.Description = data.Description.ValueString()
	}
	if !data.CredentialsID.IsNull() && !data.CredentialsID.IsUnknown() {
		spec.CredentialsID = data.CredentialsID.ValueString()
	}
	return spec
}

func (r *MountServer) syncModelFromAPI(data *MountServerModel, api *models.MountServerModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.ManagedServerID = types.StringValue(api.ManagedServerID)
	data.Type = types.StringValue(api.Type)
	data.CredentialsID = types.StringValue(api.CredentialsID)
}

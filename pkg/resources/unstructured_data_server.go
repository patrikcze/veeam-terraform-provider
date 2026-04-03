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
	_ resource.Resource                = &UnstructuredDataServer{}
	_ resource.ResourceWithConfigure   = &UnstructuredDataServer{}
	_ resource.ResourceWithImportState = &UnstructuredDataServer{}
)

// UnstructuredDataServer implements the veeam_unstructured_data_server resource.
type UnstructuredDataServer struct {
	client client.APIClient
}

// UnstructuredDataServerModel is the Terraform state model for veeam_unstructured_data_server.
type UnstructuredDataServerModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Type                types.String `tfsdk:"type"`
	HostName            types.String `tfsdk:"host_name"`
	CredentialsID       types.String `tfsdk:"credentials_id"`
	AccessCredentialsID types.String `tfsdk:"access_credentials_id"`
}

// NewUnstructuredDataServer returns a new veeam_unstructured_data_server resource instance.
func NewUnstructuredDataServer() resource.Resource {
	return &UnstructuredDataServer{}
}

func (r *UnstructuredDataServer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unstructured_data_server"
}

func (r *UnstructuredDataServer) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an unstructured data server (NAS/file share backup source) in the Veeam inventory " +
			"(`/api/v1/inventory/unstructuredDataServers`). The `type` field is immutable — " +
			"changing it forces a destroy and recreate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unstructured data server identifier (assigned by the server).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the unstructured data server.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description of the unstructured data server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Server type. Allowed values: `CifsShare`, `NfsShare`, `FileServer`. Changing this forces a destroy and recreate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "FQDN or IP address of the NAS device or file server.",
			},
			"credentials_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the credential used to connect to the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_credentials_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the credential used for share-level access (CIFS/NFS authentication).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UnstructuredDataServer) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UnstructuredDataServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UnstructuredDataServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.UnstructuredDataServerModel
	if err := r.client.PostJSON(ctx, client.PathUnstructuredDataServers, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create unstructured data server",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *UnstructuredDataServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UnstructuredDataServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.UnstructuredDataServerModel
	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read unstructured data server",
			fmt.Sprintf("API error for unstructured data server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *UnstructuredDataServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UnstructuredDataServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update unstructured data server",
			fmt.Sprintf("API error for unstructured data server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *UnstructuredDataServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UnstructuredDataServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathUnstructuredDataServerByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete unstructured data server",
			fmt.Sprintf("API error for unstructured data server %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *UnstructuredDataServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *UnstructuredDataServer) buildSpec(data *UnstructuredDataServerModel) *models.UnstructuredDataServerSpec {
	spec := &models.UnstructuredDataServerSpec{
		Name:     data.Name.ValueString(),
		Type:     data.Type.ValueString(),
		HostName: data.HostName.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		spec.Description = data.Description.ValueString()
	}
	if !data.CredentialsID.IsNull() && !data.CredentialsID.IsUnknown() {
		spec.CredentialsID = data.CredentialsID.ValueString()
	}
	if !data.AccessCredentialsID.IsNull() && !data.AccessCredentialsID.IsUnknown() {
		spec.AccessCredentialsID = data.AccessCredentialsID.ValueString()
	}
	return spec
}

func (r *UnstructuredDataServer) syncModelFromAPI(data *UnstructuredDataServerModel, api *models.UnstructuredDataServerModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(api.Type)
	data.HostName = types.StringValue(api.HostName)
	data.CredentialsID = types.StringValue(api.CredentialsID)
	data.AccessCredentialsID = types.StringValue(api.AccessCredentialsID)
}

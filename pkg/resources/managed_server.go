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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &ManagedServer{}
	_ resource.ResourceWithConfigure   = &ManagedServer{}
	_ resource.ResourceWithImportState = &ManagedServer{}
)

// ManagedServer implements the veeam_managed_server resource.
type ManagedServer struct {
	client client.APIClient
}

// ManagedServerModel is the Terraform state model.
type ManagedServerModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Type                  types.String `tfsdk:"type"`
	CredentialsID         types.String `tfsdk:"credentials_id"`
	Port                  types.Int64  `tfsdk:"port"`
	CertificateThumbprint types.String `tfsdk:"certificate_thumbprint"`
	SSHFingerprint        types.String `tfsdk:"ssh_fingerprint"`
	Status                types.String `tfsdk:"status"`
}

func (r *ManagedServer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_server"
}

func (r *ManagedServer) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam managed server (ViHost, WindowsHost, LinuxHost).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "FQDN or IP address of the managed server.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Server type: `ViHost`, `WindowsHost`, or `LinuxHost`.",
				Required:            true,
			},
			"credentials_id": schema.StringAttribute{
				MarkdownDescription: "ID of the saved credential used to connect.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Connection port (e.g. 443 for ViHost).",
				Optional:            true,
			},
			"certificate_thumbprint": schema.StringAttribute{
				MarkdownDescription: "TLS certificate thumbprint (ViHost only).",
				Optional:            true,
			},
			"ssh_fingerprint": schema.StringAttribute{
				MarkdownDescription: "SSH host key fingerprint (LinuxHost only).",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Server availability status (read-only).",
				Computed:            true,
			},
		},
	}
}

func (r *ManagedServer) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
// CRUD — managed server create/delete are async (202 Accepted)
// ---------------------------------------------------------------------------

func (r *ManagedServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ManagedServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.ManagedServerModel
	if err := r.client.PostJSON(ctx, client.PathManagedServers, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create managed server",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	// If the API returned an ID, the operation completed synchronously.
	// Some Veeam operations return 202 — in that case the client should
	// handle async polling transparently, but we log for visibility.
	if result.ID != "" {
		data.ID = types.StringValue(result.ID)
	}

	tflog.Info(ctx, "Created managed server", map[string]interface{}{"id": result.ID})
	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ManagedServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ManagedServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ManagedServerModel
	endpoint := fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read managed server",
			fmt.Sprintf("API error for server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ManagedServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ManagedServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update managed server",
			fmt.Sprintf("API error for server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ManagedServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ManagedServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete managed server",
			fmt.Sprintf("API error for server %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *ManagedServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewManagedServer returns a new veeam_managed_server resource instance.
func NewManagedServer() resource.Resource {
	return &ManagedServer{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *ManagedServer) buildSpec(data *ManagedServerModel) interface{} {
	serverType := models.EManagedServerType(data.Type.ValueString())

	base := models.ManagedServerSpec{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        serverType,
	}

	switch serverType {
	case models.ManagedServerTypeViHost:
		spec := models.ViHostSpec{
			ManagedServerSpec: base,
			CredentialsID:     data.CredentialsID.ValueString(),
		}
		if !data.Port.IsNull() && !data.Port.IsUnknown() {
			spec.Port = int(data.Port.ValueInt64())
		}
		if !data.CertificateThumbprint.IsNull() {
			spec.CertificateThumbprint = data.CertificateThumbprint.ValueString()
		}
		return &spec

	case models.ManagedServerTypeLinuxHost:
		spec := models.LinuxHostSpec{
			ManagedServerSpec:      base,
			CredentialsStorageType: models.CredentialsStorageTypeSaved,
			CredentialsID:          data.CredentialsID.ValueString(),
			SSHFingerprint:         data.SSHFingerprint.ValueString(),
		}
		return &spec

	default: // WindowsHost and others
		spec := models.WindowsHostSpec{
			ManagedServerSpec:      base,
			CredentialsStorageType: models.CredentialsStorageTypeSaved,
			CredentialsID:          data.CredentialsID.ValueString(),
		}
		return &spec
	}
}

func (r *ManagedServer) syncFromAPI(data *ManagedServerModel, api *models.ManagedServerModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(string(api.Type))
	if api.Status != "" {
		data.Status = types.StringValue(string(api.Status))
	}
}

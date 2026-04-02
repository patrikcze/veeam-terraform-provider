package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &KMSServer{}
	_ resource.ResourceWithConfigure   = &KMSServer{}
	_ resource.ResourceWithImportState = &KMSServer{}
)

// KMSServer implements the veeam_kms_server resource.
type KMSServer struct {
	client client.APIClient
}

// KMSServerModel is the Terraform state model for veeam_kms_server.
type KMSServerModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Hostname              types.String `tfsdk:"hostname"`
	Port                  types.Int64  `tfsdk:"port"`
	CertificateThumbprint types.String `tfsdk:"certificate_thumbprint"`
}

func (r *KMSServer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms_server"
}

func (r *KMSServer) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a KMS (Key Management Service) server registration in Veeam Backup & Replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "KMS server identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name of the KMS server.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description of the KMS server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address of the KMS server.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number of the KMS server (default: 9998).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"certificate_thumbprint": schema.StringAttribute{
				MarkdownDescription: "TLS certificate thumbprint of the KMS server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *KMSServer) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KMSServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KMSServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.KMSServerModel
	if err := r.client.PostJSON(ctx, client.PathKMSServers, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create KMS server",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *KMSServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KMSServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.KMSServerModel
	endpoint := fmt.Sprintf(client.PathKMSServerByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read KMS server",
			fmt.Sprintf("API error for KMS server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *KMSServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KMSServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathKMSServerByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update KMS server",
			fmt.Sprintf("API error for KMS server %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *KMSServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KMSServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathKMSServerByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete KMS server",
			fmt.Sprintf("API error for KMS server %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *KMSServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewKMSServer returns a new veeam_kms_server resource instance.
func NewKMSServer() resource.Resource {
	return &KMSServer{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildSpec converts Terraform state to an API request body.
func (r *KMSServer) buildSpec(data *KMSServerModel) *models.KMSServerSpec {
	spec := &models.KMSServerSpec{
		Name:     data.Name.ValueString(),
		HostName: data.Hostname.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		spec.Description = data.Description.ValueString()
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		spec.Port = data.Port.ValueInt64()
	}
	if !data.CertificateThumbprint.IsNull() && !data.CertificateThumbprint.IsUnknown() {
		spec.CertificateThumbprint = data.CertificateThumbprint.ValueString()
	}
	return spec
}

// syncModelFromAPI updates Terraform state fields from an API response.
func (r *KMSServer) syncModelFromAPI(data *KMSServerModel, api *models.KMSServerModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Hostname = types.StringValue(api.HostName)
	data.Port = types.Int64Value(api.Port)
	data.CertificateThumbprint = types.StringValue(api.CertificateThumbprint)
}

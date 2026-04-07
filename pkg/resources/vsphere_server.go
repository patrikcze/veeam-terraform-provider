package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &VSphereServer{}
	_ resource.ResourceWithConfigure   = &VSphereServer{}
	_ resource.ResourceWithImportState = &VSphereServer{}
)

// VSphereServer implements the veeam_vsphere_server resource.
// It registers a vCenter Server or standalone ESXi host in VBR's backup
// infrastructure (type = "ViHost" in the managed servers API).
type VSphereServer struct {
	client client.APIClient
}

// VSphereServerModel is the Terraform state model for veeam_vsphere_server.
type VSphereServerModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	CredentialsID         types.String `tfsdk:"credentials_id"`
	Port                  types.Int64  `tfsdk:"port"`
	CertificateThumbprint types.String `tfsdk:"certificate_thumbprint"`
	Status                types.String `tfsdk:"status"`
}

// NewVSphereServer returns a new veeam_vsphere_server resource instance.
func NewVSphereServer() resource.Resource {
	return &VSphereServer{}
}

func (r *VSphereServer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vsphere_server"
}

func (r *VSphereServer) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers a VMware vCenter Server or standalone ESXi host in Veeam Backup & " +
			"Replication's Virtual Infrastructure (backup infrastructure type `ViHost`). " +
			"Once registered, VBR can discover VMs and datastores for backup jobs.\n\n" +
			"This is a dedicated resource for VMware vSphere — equivalent to using " +
			"`veeam_managed_server` with `type = \"ViHost\"` but with a vSphere-focused schema.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier assigned by VBR.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "FQDN or IP address of the vCenter Server or ESXi host.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credentials_id": schema.StringAttribute{
				MarkdownDescription: "ID of the saved credential used to connect to vCenter/ESXi.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "HTTPS port for the vSphere API (default: 443).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"certificate_thumbprint": schema.StringAttribute{
				MarkdownDescription: "SHA-1 thumbprint of the vCenter/ESXi TLS certificate. " +
					"If omitted, VBR will fetch and trust the certificate on first connection.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Connection status as reported by VBR (read-only).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VSphereServer) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider.")
		return
	}
	r.client = c
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *VSphereServer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VSphereServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result map[string]interface{}
	if err := r.client.PostJSON(ctx, client.PathManagedServers, payload, &result); err != nil {
		resp.Diagnostics.AddError("Failed to create vSphere server", fmt.Sprintf("API error: %s", err))
		return
	}

	resultID := getStringValue(result, "id")
	if resultID == "" {
		resp.Diagnostics.AddError("Failed to create vSphere server", "API response did not include an ID.")
		return
	}

	// If the result looks like a session/task, poll for completion then resolve the ID.
	if isViHostAsyncResult(result) {
		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError("Failed to create vSphere server",
				fmt.Sprintf("Async task %s failed: %s", resultID, err))
			return
		}

		resolvedID, err := r.findServerID(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve created vSphere server", err.Error())
			return
		}
		data.ID = types.StringValue(resolvedID)
	} else {
		data.ID = types.StringValue(resultID)
	}

	// Read back to populate computed fields (status, description, port, thumbprint).
	if id := data.ID.ValueString(); id != "" {
		var created models.ManagedServerModel
		if err := r.client.GetJSON(ctx, fmt.Sprintf(client.PathManagedServerByID, id), &created); err == nil {
			r.syncFromAPI(&data, &created)
		}
	}

	tflog.Info(ctx, "Created vSphere server", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *VSphereServer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VSphereServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ManagedServerModel
	if err := r.client.GetJSON(ctx, fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString()), &result); err != nil {
		resp.Diagnostics.AddError("Failed to read vSphere server",
			fmt.Sprintf("API error for %s: %s", data.ID.ValueString(), err))
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *VSphereServer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VSphereServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)
	if err := r.client.PutJSON(ctx, fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString()), payload, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update vSphere server",
			fmt.Sprintf("API error for %s: %s", data.ID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *VSphereServer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VSphereServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteJSON(ctx, fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Failed to delete vSphere server",
			fmt.Sprintf("API error for %s: %s", data.ID.ValueString(), err))
		return
	}

	if err := r.waitForDeleted(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to confirm vSphere server deletion",
			fmt.Sprintf("Server %s delete accepted but still present: %s", data.ID.ValueString(), err))
	}
}

func (r *VSphereServer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *VSphereServer) buildSpec(data *VSphereServerModel) *models.ViHostSpec {
	spec := models.ViHostSpec{
		ManagedServerSpec: models.ManagedServerSpec{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			Type:        models.ManagedServerTypeViHost,
		},
		CredentialsID: data.CredentialsID.ValueString(),
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		spec.Port = int(data.Port.ValueInt64())
	}
	if !data.CertificateThumbprint.IsNull() && !data.CertificateThumbprint.IsUnknown() {
		spec.CertificateThumbprint = data.CertificateThumbprint.ValueString()
	}
	return &spec
}

func (r *VSphereServer) syncFromAPI(data *VSphereServerModel, api *models.ManagedServerModel) {
	if api.Name != "" {
		data.Name = types.StringValue(api.Name)
	}
	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}
	if string(api.Status) != "" {
		data.Status = types.StringValue(string(api.Status))
	}
}

// isViHostAsyncResult returns true when the POST response is an async session
// (i.e. not a direct ViHost object).
func isViHostAsyncResult(result map[string]interface{}) bool {
	t := strings.ToLower(getStringValue(result, "type"))
	if t == "" {
		return true
	}
	vihost := strings.ToLower(string(models.ManagedServerTypeViHost))
	return t == "session" || t != vihost
}

// findServerID lists all managed servers and returns the ID of the ViHost
// whose name matches. Used after an async create.
func (r *VSphereServer) findServerID(ctx context.Context, name string) (string, error) {
	var payload map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathManagedServers, &payload); err != nil {
		return "", fmt.Errorf("failed to list managed servers after create: %w", err)
	}

	rawData, _ := payload["data"].([]interface{})
	for _, item := range rawData {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if strings.EqualFold(getStringValue(entry, "name"), name) &&
			strings.EqualFold(getStringValue(entry, "type"), string(models.ManagedServerTypeViHost)) {
			if id := getStringValue(entry, "id"); id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("vSphere server %q was created but could not be located in the managed server list", name)
}

func (r *VSphereServer) waitForDeleted(ctx context.Context, serverID string) error {
	const pollInterval = 3 * time.Second
	const timeout = 2 * time.Minute

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	endpoint := fmt.Sprintf(client.PathManagedServerByID, serverID)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		var result models.ManagedServerModel
		err := r.client.GetJSON(pollCtx, endpoint, &result)
		if err != nil {
			if isViHostNotFound(err) {
				return nil
			}
			return err
		}
		select {
		case <-pollCtx.Done():
			return fmt.Errorf("timed out after %s waiting for vSphere server %s to be deleted", timeout, serverID)
		case <-ticker.C:
		}
	}
}

func isViHostNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *models.APIError
	if errors.As(err, &apiErr) && strings.EqualFold(apiErr.ErrorCode, "NotFound") {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "http 404") || strings.Contains(lower, "notfound")
}

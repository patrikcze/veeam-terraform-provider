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

	// Use a create-only copy so we can resolve/adjust payload fields without
	// changing configured attribute values in Terraform state.
	createData := data

	if shouldResolveLinuxFingerprint(&createData) {
		fingerprint, err := r.resolveLinuxSSHFingerprint(ctx, &createData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve Linux SSH fingerprint",
				fmt.Sprintf("API error: %s", err),
			)
			return
		}
		createData.SSHFingerprint = types.StringValue(fingerprint)
	}

	payload := r.buildSpec(&createData)

	result, err := r.createManagedServer(ctx, &createData, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create managed server",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	resultID := getStringValue(result, "id")
	if resultID == "" {
		resp.Diagnostics.AddError(
			"Failed to create managed server",
			"API response did not include managed server ID or async session ID.",
		)
		return
	}

	if isAsyncManagedServerCreateResult(result) {
		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to create managed server",
				fmt.Sprintf("Async managed server creation task %s failed: %s", resultID, err),
			)
			return
		}

		resolvedID, err := r.findManagedServerID(ctx, &createData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve created managed server",
				err.Error(),
			)
			return
		}
		data.ID = types.StringValue(resolvedID)
	} else {
		data.ID = types.StringValue(resultID)
	}

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		var created models.ManagedServerModel
		endpoint := fmt.Sprintf(client.PathManagedServerByID, data.ID.ValueString())
		if err := r.client.GetJSON(ctx, endpoint, &created); err == nil {
			r.syncFromAPI(&data, &created)
		}
	}

	tflog.Info(ctx, "Created managed server", map[string]interface{}{"id": data.ID.ValueString()})
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

	if err := r.waitForManagedServerDeleted(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Failed to confirm managed server deletion",
			fmt.Sprintf("Managed server %s delete request was accepted but resource still appears present: %s", data.ID.ValueString(), err),
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
	if api.Name != "" {
		data.Name = types.StringValue(api.Name)
	}
	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}
	if string(api.Type) != "" {
		data.Type = types.StringValue(string(api.Type))
	}
	if api.Status != "" {
		data.Status = types.StringValue(string(api.Status))
	}
}

func (r *ManagedServer) createManagedServer(ctx context.Context, data *ManagedServerModel, payload interface{}) (map[string]interface{}, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var result map[string]interface{}
		err := r.client.PostJSON(ctx, client.PathManagedServers, payload, &result)
		if err == nil {
			return result, nil
		}

		if shouldRetryManagedServerCreate(data, err) && attempt < maxAttempts {
			fingerprint, resolveErr := r.resolveLinuxSSHFingerprint(ctx, data)
			if resolveErr == nil && fingerprint != "" {
				data.SSHFingerprint = types.StringValue(fingerprint)
				payload = r.buildSpec(data)
			}
		}

		if !shouldRetryManagedServerCreate(data, err) || attempt == maxAttempts {
			return nil, err
		}

		tflog.Warn(ctx, "Managed server create failed during Linux credential validation, retrying", map[string]interface{}{
			"attempt": attempt,
			"name":    data.Name.ValueString(),
		})

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	return nil, fmt.Errorf("managed server create retry exhausted")
}

func shouldRetryManagedServerCreate(data *ManagedServerModel, err error) bool {
	if data == nil || data.Type.IsNull() || !strings.EqualFold(data.Type.ValueString(), string(models.ManagedServerTypeLinuxHost)) {
		return false
	}

	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "failed to validate the specified linux credentials")
}

func shouldResolveLinuxFingerprint(data *ManagedServerModel) bool {
	if data == nil || data.Type.IsNull() {
		return false
	}

	if !strings.EqualFold(data.Type.ValueString(), string(models.ManagedServerTypeLinuxHost)) {
		return false
	}

	if data.SSHFingerprint.IsNull() || data.SSHFingerprint.IsUnknown() {
		return true
	}

	fingerprint := strings.TrimSpace(data.SSHFingerprint.ValueString())
	if fingerprint == "" {
		return true
	}

	if strings.HasPrefix(strings.ToUpper(fingerprint), "SHA256:") {
		return true
	}

	return false
}

func (r *ManagedServer) resolveLinuxSSHFingerprint(ctx context.Context, data *ManagedServerModel) (string, error) {
	request := map[string]interface{}{
		"serverName":             data.Name.ValueString(),
		"credentialsStorageType": string(models.CredentialsStorageTypePermanent),
		"credentialsId":          data.CredentialsID.ValueString(),
		"type":                   string(models.ManagedServerTypeLinuxHost),
	}

	var response map[string]interface{}
	if err := r.client.PostJSON(ctx, client.PathConnectionCertificate, request, &response); err != nil {
		return "", fmt.Errorf("failed to retrieve SSH fingerprint: %w", err)
	}

	fingerprint := strings.TrimSpace(getStringValue(response, "fingerprint"))
	if fingerprint == "" {
		return "", fmt.Errorf("connection fingerprint response did not include fingerprint value")
	}

	tflog.Info(ctx, "Resolved Linux SSH fingerprint from Veeam API", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	return fingerprint, nil
}

func isAsyncManagedServerCreateResult(result map[string]interface{}) bool {
	resultType := getStringValue(result, "type")
	if resultType == "" {
		return true
	}

	if strings.EqualFold(resultType, "session") {
		return true
	}

	return !strings.EqualFold(resultType, string(models.ManagedServerTypeViHost)) &&
		!strings.EqualFold(resultType, string(models.ManagedServerTypeWindowsHost)) &&
		!strings.EqualFold(resultType, string(models.ManagedServerTypeLinuxHost))
}

func (r *ManagedServer) findManagedServerID(ctx context.Context, data *ManagedServerModel) (string, error) {
	var payload map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathManagedServers, &payload); err != nil {
		return "", fmt.Errorf("failed to list managed servers after create: %w", err)
	}

	rawData, ok := payload["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected managed servers list response shape: missing data array")
	}

	for _, item := range rawData {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entryName := getStringValue(entry, "name")
		if entryName == "" || !strings.EqualFold(entryName, data.Name.ValueString()) {
			continue
		}

		if !data.Type.IsNull() {
			entryType := getStringValue(entry, "type")
			if entryType != "" && !strings.EqualFold(entryType, data.Type.ValueString()) {
				continue
			}
		}

		id := getStringValue(entry, "id")
		if id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf("managed server %q was created but could not be located in managed server list", data.Name.ValueString())
}

func (r *ManagedServer) waitForManagedServerDeleted(ctx context.Context, serverID string) error {
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
			if isManagedServerNotFound(err) {
				return nil
			}
			return err
		}

		select {
		case <-pollCtx.Done():
			return fmt.Errorf("timed out after %s", timeout)
		case <-ticker.C:
		}
	}
}

func isManagedServerNotFound(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		if strings.EqualFold(apiErr.ErrorCode, "NotFound") {
			return true
		}
	}

	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "http 404") || strings.Contains(errText, "notfound")
}

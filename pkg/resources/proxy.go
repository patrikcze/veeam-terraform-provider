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
	_ resource.Resource                = &Proxy{}
	_ resource.ResourceWithConfigure   = &Proxy{}
	_ resource.ResourceWithImportState = &Proxy{}
)

// Proxy implements the veeam_proxy resource.
type Proxy struct {
	client client.APIClient
}

// ProxyModel is the Terraform state model.
type ProxyModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Type                  types.String `tfsdk:"type"`
	HostID                types.String `tfsdk:"host_id"`
	TransportMode         types.String `tfsdk:"transport_mode"`
	FailoverToNetwork     types.Bool   `tfsdk:"failover_to_network"`
	HostToProxyEncryption types.Bool   `tfsdk:"host_to_proxy_encryption"`
	MaxTaskCount          types.Int64  `tfsdk:"max_task_count"`
}

func (r *Proxy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy"
}

func (r *Proxy) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam backup proxy (ViProxy, HvProxy, or GeneralPurposeProxy).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Proxy identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Proxy name (read-only, derived from host).",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Proxy type: `ViProxy`, `HvProxy`, or `GeneralPurposeProxy`.",
				Required:            true,
			},
			"host_id": schema.StringAttribute{
				MarkdownDescription: "ID of the managed server used as the proxy.",
				Required:            true,
			},
			"transport_mode": schema.StringAttribute{
				MarkdownDescription: "Data transport mode (`ViProxy` only): `Auto`, `DirectAccess`, `VirtualAppliance`, or `Network`.",
				Optional:            true,
				Computed:            true,
			},
			"failover_to_network": schema.BoolAttribute{
				MarkdownDescription: "Failover to network transport if primary mode fails (`ViProxy` only).",
				Optional:            true,
				Computed:            true,
			},
			"host_to_proxy_encryption": schema.BoolAttribute{
				MarkdownDescription: "Encrypt data between host and proxy (`ViProxy` only).",
				Optional:            true,
				Computed:            true,
			},
			"max_task_count": schema.Int64Attribute{
				MarkdownDescription: "Maximum concurrent tasks.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *Proxy) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *Proxy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result map[string]interface{}
	if err := r.client.PostJSON(ctx, client.PathProxies, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create proxy",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	resultID := getStringValue(result, "id")
	resultType := getStringValue(result, "type")

	// POST returns a SessionModel (async) — wait for task and resolve proxy ID.
	if resultType == "" {
		if resultID == "" {
			resp.Diagnostics.AddError(
				"Failed to create proxy",
				"API response did not include proxy type or async session ID.",
			)
			return
		}

		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to create proxy",
				fmt.Sprintf("Async proxy creation task %s failed: %s", resultID, err),
			)
			return
		}

		resolvedID, err := r.findProxyID(ctx, &data)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve created proxy",
				err.Error(),
			)
			return
		}
		data.ID = types.StringValue(resolvedID)
	} else {
		data.ID = types.StringValue(resultID)
	}

	// Read created proxy to sync computed fields.
	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		var created models.ViProxyModel
		endpoint := fmt.Sprintf(client.PathProxyByID, data.ID.ValueString())
		if err := r.client.GetJSON(ctx, endpoint, &created); err == nil {
			r.syncFromAPI(&data, &created)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Proxy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.ViProxyModel
	endpoint := fmt.Sprintf(client.PathProxyByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read proxy",
			fmt.Sprintf("API error for proxy %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Proxy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathProxyByID, data.ID.ValueString())
	var result map[string]interface{}
	if err := r.client.PutJSON(ctx, endpoint, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update proxy",
			fmt.Sprintf("API error for proxy %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Handle async response from PUT.
	resultID := getStringValue(result, "id")
	resultType := getStringValue(result, "type")
	if resultType == "" && resultID != "" {
		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update proxy",
				fmt.Sprintf("Async task %s failed: %s", resultID, err),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Proxy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathProxyByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete proxy",
			fmt.Sprintf("API error for proxy %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *Proxy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewProxy returns a new veeam_proxy resource instance.
func NewProxy() resource.Resource {
	return &Proxy{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildSpec builds the correct proxy spec based on the proxy type.
func (r *Proxy) buildSpec(data *ProxyModel) interface{} {
	proxyType := models.EProxyType(data.Type.ValueString())
	base := models.ProxySpec{
		Description: data.Description.ValueString(),
		Type:        proxyType,
	}

	switch proxyType {
	case models.ProxyTypeHvProxy:
		server := &models.HvProxyServerSettings{
			HostID: data.HostID.ValueString(),
		}
		if !data.MaxTaskCount.IsNull() && !data.MaxTaskCount.IsUnknown() {
			server.MaxTaskCount = int(data.MaxTaskCount.ValueInt64())
		}
		return &models.HvProxySpec{
			ProxySpec: base,
			Server:    server,
		}

	case models.ProxyTypeGeneralPurposeProxy:
		server := &models.GeneralPurposeProxyServerSettings{
			HostID: data.HostID.ValueString(),
		}
		if !data.MaxTaskCount.IsNull() && !data.MaxTaskCount.IsUnknown() {
			server.MaxTaskCount = int(data.MaxTaskCount.ValueInt64())
		}
		return &models.GeneralPurposeProxySpec{
			ProxySpec: base,
			Server:    server,
		}

	default: // ViProxy
		server := &models.ProxyServerSettings{
			HostID: data.HostID.ValueString(),
		}
		if !data.TransportMode.IsNull() {
			server.TransportMode = models.EBackupProxyTransportMode(data.TransportMode.ValueString())
		}
		if !data.FailoverToNetwork.IsNull() {
			server.FailoverToNetwork = data.FailoverToNetwork.ValueBool()
		}
		if !data.HostToProxyEncryption.IsNull() {
			server.HostToProxyEncryption = data.HostToProxyEncryption.ValueBool()
		}
		if !data.MaxTaskCount.IsNull() && !data.MaxTaskCount.IsUnknown() {
			server.MaxTaskCount = int(data.MaxTaskCount.ValueInt64())
		}
		return &models.ViProxySpec{
			ProxySpec: base,
			Server:    server,
		}
	}
}

// syncFromAPI merges API response fields into the Terraform state.
// Uses ViProxyModel which covers all proxy types (extra fields just remain empty for non-ViProxy).
func (r *Proxy) syncFromAPI(data *ProxyModel, api *models.ViProxyModel) {
	if api.Name != "" {
		data.Name = types.StringValue(api.Name)
	}
	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}
	if string(api.Type) != "" {
		data.Type = types.StringValue(string(api.Type))
	}
	if api.Server != nil {
		if api.Server.HostID != "" {
			data.HostID = types.StringValue(api.Server.HostID)
		}
		if api.Server.MaxTaskCount > 0 {
			data.MaxTaskCount = types.Int64Value(int64(api.Server.MaxTaskCount))
		}
		if string(api.Server.TransportMode) != "" {
			data.TransportMode = types.StringValue(string(api.Server.TransportMode))
		}
	}
}

func (r *Proxy) findProxyID(ctx context.Context, data *ProxyModel) (string, error) {
	var payload map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathProxies, &payload); err != nil {
		return "", fmt.Errorf("failed to list proxies after create: %w", err)
	}

	rawData, ok := payload["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected proxies list response shape: missing data array")
	}

	for _, item := range rawData {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entryType := getStringValue(entry, "type")
		if !data.Type.IsNull() && entryType != "" && entryType != data.Type.ValueString() {
			continue
		}

		if !data.Description.IsNull() {
			entryDescription := getStringValue(entry, "description")
			if entryDescription != data.Description.ValueString() {
				continue
			}
		}

		if !data.HostID.IsNull() {
			server, ok := entry["server"].(map[string]interface{})
			if ok {
				entryHostID := getStringValue(server, "hostId")
				if entryHostID != "" && entryHostID != data.HostID.ValueString() {
					continue
				}
			}
		}

		id := getStringValue(entry, "id")
		if id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf("proxy could not be located in proxy list after create")
}

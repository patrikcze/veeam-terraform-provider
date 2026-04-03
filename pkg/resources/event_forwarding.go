package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &EventForwarding{}
	_ resource.ResourceWithConfigure   = &EventForwarding{}
	_ resource.ResourceWithImportState = &EventForwarding{}
)

// eventForwardingID is the fixed singleton ID used in Terraform state.
const eventForwardingID = "event-forwarding"

// EventForwarding manages the Veeam event forwarding settings singleton
// (GET/PUT /api/v1/generalOptions/eventForwarding).
type EventForwarding struct {
	client client.APIClient
}

// EventForwardingModel is the Terraform state model for veeam_event_forwarding.
type EventForwardingModel struct {
	ID             types.String `tfsdk:"id"`
	SNMPEnabled    types.Bool   `tfsdk:"snmp_enabled"`
	SNMPHost       types.String `tfsdk:"snmp_host"`
	SNMPPort       types.Int64  `tfsdk:"snmp_port"`
	SNMPCommunity  types.String `tfsdk:"snmp_community"`
	SyslogEnabled  types.Bool   `tfsdk:"syslog_enabled"`
	SyslogHost     types.String `tfsdk:"syslog_host"`
	SyslogPort     types.Int64  `tfsdk:"syslog_port"`
	SyslogProtocol types.String `tfsdk:"syslog_protocol"`
}

// NewEventForwarding returns a new veeam_event_forwarding resource instance.
func NewEventForwarding() resource.Resource {
	return &EventForwarding{}
}

func (r *EventForwarding) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_forwarding"
}

func (r *EventForwarding) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useStateString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	useStateBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	useStateInt := []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Veeam event forwarding configuration singleton " +
			"(`/api/v1/generalOptions/eventForwarding`). This is a singleton resource — only " +
			"one instance may exist per provider configuration. Deleting the resource only " +
			"removes it from Terraform state; the server configuration is not reset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `\"event-forwarding\"`. Fixed singleton identifier.",
				PlanModifiers:       useStateString,
			},
			"snmp_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether SNMP trap forwarding is enabled.",
				PlanModifiers:       useStateBool,
			},
			"snmp_host": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hostname or IP address of the SNMP trap receiver.",
				PlanModifiers:       useStateString,
			},
			"snmp_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UDP port of the SNMP trap receiver (default: 162).",
				PlanModifiers:       useStateInt,
			},
			"snmp_community": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "SNMP community string for trap authentication.",
				PlanModifiers:       useStateString,
			},
			"syslog_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether syslog event forwarding is enabled.",
				PlanModifiers:       useStateBool,
			},
			"syslog_host": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hostname or IP address of the syslog server.",
				PlanModifiers:       useStateString,
			},
			"syslog_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Port of the syslog server.",
				PlanModifiers:       useStateInt,
			},
			"syslog_protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Transport protocol for syslog messages. Allowed values: `UDP`, `TCP`.",
				PlanModifiers:       useStateString,
			},
		},
	}
}

func (r *EventForwarding) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create issues a GET → merge → PUT and records state with the fixed singleton ID.
func (r *EventForwarding) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EventForwardingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyEventForwarding(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure event forwarding", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(eventForwardingID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EventForwarding) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EventForwardingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathEventForwarding, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read event forwarding", fmt.Sprintf("API error: %s", err))
		return
	}

	syncEventForwardingFromPayload(raw, &data)
	data.ID = types.StringValue(eventForwardingID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update applies plan changes via GET → merge → PUT.
func (r *EventForwarding) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EventForwardingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyEventForwarding(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update event forwarding", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(eventForwardingID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete removes the resource from Terraform state only. The server-side
// configuration cannot be deleted via the API.
func (r *EventForwarding) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentional no-op: singletons cannot be deleted server-side.
}

func (r *EventForwarding) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

// applyEventForwarding GETs the current payload, merges plan values in, and PUTs the result.
func (r *EventForwarding) applyEventForwarding(ctx context.Context, data *EventForwardingModel) error {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathEventForwarding, &raw); err != nil {
		return fmt.Errorf("reading current event forwarding settings: %w", err)
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}

	snmp := ensureNestedConfigMap(raw, "snmp")
	if !data.SNMPEnabled.IsNull() && !data.SNMPEnabled.IsUnknown() {
		setBoolValue(snmp, data.SNMPEnabled.ValueBool(), "isEnabled")
	}
	if !data.SNMPHost.IsNull() && !data.SNMPHost.IsUnknown() {
		setStringValue(snmp, data.SNMPHost.ValueString(), "host")
	}
	if !data.SNMPPort.IsNull() && !data.SNMPPort.IsUnknown() {
		setIntValue(snmp, int(data.SNMPPort.ValueInt64()), "port")
	}
	if !data.SNMPCommunity.IsNull() && !data.SNMPCommunity.IsUnknown() {
		setStringValue(snmp, data.SNMPCommunity.ValueString(), "community")
	}

	syslog := ensureNestedConfigMap(raw, "syslog")
	if !data.SyslogEnabled.IsNull() && !data.SyslogEnabled.IsUnknown() {
		setBoolValue(syslog, data.SyslogEnabled.ValueBool(), "isEnabled")
	}
	if !data.SyslogHost.IsNull() && !data.SyslogHost.IsUnknown() {
		setStringValue(syslog, data.SyslogHost.ValueString(), "host")
	}
	if !data.SyslogPort.IsNull() && !data.SyslogPort.IsUnknown() {
		setIntValue(syslog, int(data.SyslogPort.ValueInt64()), "port")
	}
	if !data.SyslogProtocol.IsNull() && !data.SyslogProtocol.IsUnknown() {
		setStringValue(syslog, data.SyslogProtocol.ValueString(), "protocol")
	}

	return r.client.PutJSON(ctx, client.PathEventForwarding, raw, nil)
}

// syncEventForwardingFromPayload populates the model from the API response map.
func syncEventForwardingFromPayload(raw map[string]interface{}, data *EventForwardingModel) {
	snmp := getNestedConfigMap(raw, "snmp")
	data.SNMPEnabled = types.BoolValue(getConfigBoolValue(snmp, "isEnabled"))
	if v := getConfigStringValue(snmp, "host"); v != "" {
		data.SNMPHost = types.StringValue(v)
	}
	if v := getConfigIntValue(snmp, "port"); v != 0 {
		data.SNMPPort = types.Int64Value(int64(v))
	}
	if v := getConfigStringValue(snmp, "community"); v != "" {
		data.SNMPCommunity = types.StringValue(v)
	}

	syslog := getNestedConfigMap(raw, "syslog")
	data.SyslogEnabled = types.BoolValue(getConfigBoolValue(syslog, "isEnabled"))
	if v := getConfigStringValue(syslog, "host"); v != "" {
		data.SyslogHost = types.StringValue(v)
	}
	if v := getConfigIntValue(syslog, "port"); v != 0 {
		data.SyslogPort = types.Int64Value(int64(v))
	}
	if v := getConfigStringValue(syslog, "protocol"); v != "" {
		data.SyslogProtocol = types.StringValue(v)
	}
}

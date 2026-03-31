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

// Compile-time interface assertions.
var (
	_ resource.Resource                = &GeneralOptions{}
	_ resource.ResourceWithConfigure   = &GeneralOptions{}
	_ resource.ResourceWithImportState = &GeneralOptions{}
)

// generalOptionsID is the fixed singleton ID used in Terraform state.
const generalOptionsID = "general-options"

// GeneralOptions manages the Veeam server-level general options singleton
// (GET/PUT /api/v1/generalOptions). Because the server object always exists
// and cannot be deleted, Create/Update both issue a PUT and Delete simply
// removes the resource from Terraform state.
type GeneralOptions struct {
	client client.APIClient
}

// GeneralOptionsModel is the Terraform state model for veeam_general_options.
type GeneralOptionsModel struct {
	ID types.String `tfsdk:"id"`

	// Storage Latency Control
	StorageLatencyControlEnabled types.Bool  `tfsdk:"storage_latency_control_enabled"`
	StorageLatencyLimitMs        types.Int64 `tfsdk:"storage_latency_limit_ms"`

	// Email Notifications
	EmailNotificationsEnabled types.Bool   `tfsdk:"email_notifications_enabled"`
	EmailSMTPServer           types.String `tfsdk:"email_smtp_server"`
	EmailSMTPPort             types.Int64  `tfsdk:"email_smtp_port"`
	EmailFrom                 types.String `tfsdk:"email_from"`
	EmailTo                   types.String `tfsdk:"email_to"`
	EmailSubject              types.String `tfsdk:"email_subject"`

	// SNMP Notifications
	SNMPNotificationsEnabled types.Bool `tfsdk:"snmp_notifications_enabled"`

	// Syslog / Event Forwarding
	SyslogNotificationsEnabled types.Bool   `tfsdk:"syslog_notifications_enabled"`
	SyslogServer               types.String `tfsdk:"syslog_server"`
	SyslogPort                 types.Int64  `tfsdk:"syslog_port"`
}

// NewGeneralOptions is the provider factory function.
func NewGeneralOptions() resource.Resource {
	return &GeneralOptions{}
}

func (r *GeneralOptions) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_general_options"
}

func (r *GeneralOptions) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useStateString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	useStateBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	useStateInt := []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam Backup & Replication server-level general options " +
			"(`/api/v1/generalOptions`). This is a singleton resource — only one instance " +
			"may exist per provider configuration. Deleting the resource only removes it " +
			"from Terraform state; the server configuration is not reset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       useStateString,
				MarkdownDescription: "Always `\"general-options\"`. Fixed singleton identifier.",
			},

			// Storage Latency Control
			"storage_latency_control_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateBool,
				MarkdownDescription: "Whether storage latency control is enabled (`storageLatencyControl.isEnabled`).",
			},
			"storage_latency_limit_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateInt,
				MarkdownDescription: "Latency threshold in milliseconds above which backup activity is throttled (`storageLatencyControl.latencyLimitMs`).",
			},

			// Email Notifications
			"email_notifications_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateBool,
				MarkdownDescription: "Whether email notifications are enabled (`emailNotifications.isEnabled`).",
			},
			"email_smtp_server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateString,
				MarkdownDescription: "SMTP server hostname or IP address (`emailNotifications.smtpServer`).",
			},
			"email_smtp_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateInt,
				MarkdownDescription: "SMTP server port (`emailNotifications.port`).",
			},
			"email_from": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateString,
				MarkdownDescription: "Sender email address (`emailNotifications.from`).",
			},
			"email_to": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateString,
				MarkdownDescription: "Recipient email address (`emailNotifications.to`).",
			},
			"email_subject": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateString,
				MarkdownDescription: "Email subject template (`emailNotifications.subject`).",
			},

			// SNMP Notifications
			"snmp_notifications_enabled": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: useStateBool,
				MarkdownDescription: "Whether SNMP notifications are enabled (`snmpNotifications.isEnabled`). " +
					"Requires SNMP to be configured in the Veeam console before enabling.",
			},

			// Syslog / Event Forwarding
			"syslog_notifications_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateBool,
				MarkdownDescription: "Whether syslog event forwarding is enabled (`syslogNotifications.isEnabled`).",
			},
			"syslog_server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateString,
				MarkdownDescription: "Syslog server hostname or IP address (`syslogNotifications.dnsName`).",
			},
			"syslog_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateInt,
				MarkdownDescription: "Syslog server UDP/TCP port (`syslogNotifications.port`).",
			},
		},
	}
}

func (r *GeneralOptions) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create issues a GET → merge → PUT and then records state.
func (r *GeneralOptions) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GeneralOptionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := r.loadGeneralOptionsPayload(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read general options", fmt.Sprintf("API error: %s", err))
		return
	}

	applyModelToPayload(&data, payload)

	if err := r.client.PutJSON(ctx, client.PathGeneralOptions, payload, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update general options", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(generalOptionsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read fetches the current server state and syncs it to Terraform state.
func (r *GeneralOptions) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GeneralOptionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathGeneralOptions, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read general options", fmt.Sprintf("API error: %s", err))
		return
	}

	payload := unwrapGeneralOptionsPayload(raw)
	data.ID = types.StringValue(generalOptionsID)
	syncGeneralOptionsFromPayload(payload, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update applies plan changes via GET → merge → PUT.
func (r *GeneralOptions) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GeneralOptionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := r.loadGeneralOptionsPayload(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read general options", fmt.Sprintf("API error: %s", err))
		return
	}

	applyModelToPayload(&data, payload)

	if err := r.client.PutJSON(ctx, client.PathGeneralOptions, payload, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update general options", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(generalOptionsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete removes the resource from Terraform state only. The server-side
// configuration object cannot be deleted via the API.
func (r *GeneralOptions) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentional no-op: singletons cannot be deleted server-side.
	// Terraform will remove the resource from state automatically.
}

// ImportState supports `terraform import veeam_general_options.name general-options`.
func (r *GeneralOptions) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

// loadGeneralOptionsPayload issues a GET and returns the unwrapped payload map.
func (r *GeneralOptions) loadGeneralOptionsPayload(ctx context.Context) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathGeneralOptions, &raw); err != nil {
		return nil, err
	}
	payload := unwrapGeneralOptionsPayload(raw)
	if payload == nil {
		return map[string]interface{}{}, nil
	}
	return payload, nil
}

// unwrapGeneralOptionsPayload unwraps a potential `{"data": {...}}` envelope.
func unwrapGeneralOptionsPayload(raw map[string]interface{}) map[string]interface{} {
	if raw == nil {
		return map[string]interface{}{}
	}
	if data, ok := raw["data"]; ok {
		if nested, ok := data.(map[string]interface{}); ok {
			return nested
		}
	}
	return raw
}

// applyModelToPayload copies non-null plan values into the live payload map,
// respecting existing key names already present on the server (via setBoolValue
// / setStringValue / setIntValue which prefer existing keys).
func applyModelToPayload(data *GeneralOptionsModel, payload map[string]interface{}) {
	// Storage Latency Control
	slc := ensureNestedConfigMap(payload, "storageLatencyControl")
	if !data.StorageLatencyControlEnabled.IsNull() && !data.StorageLatencyControlEnabled.IsUnknown() {
		setBoolValue(slc, data.StorageLatencyControlEnabled.ValueBool(), "isEnabled")
	}
	if !data.StorageLatencyLimitMs.IsNull() && !data.StorageLatencyLimitMs.IsUnknown() {
		setIntValue(slc, int(data.StorageLatencyLimitMs.ValueInt64()), "latencyLimitMs")
	}

	// Email Notifications
	email := ensureNestedConfigMap(payload, "emailNotifications")
	if !data.EmailNotificationsEnabled.IsNull() && !data.EmailNotificationsEnabled.IsUnknown() {
		setBoolValue(email, data.EmailNotificationsEnabled.ValueBool(), "isEnabled")
	}
	if !data.EmailSMTPServer.IsNull() && !data.EmailSMTPServer.IsUnknown() {
		setStringValue(email, data.EmailSMTPServer.ValueString(), "smtpServer")
	}
	if !data.EmailSMTPPort.IsNull() && !data.EmailSMTPPort.IsUnknown() {
		setIntValue(email, int(data.EmailSMTPPort.ValueInt64()), "port")
	}
	if !data.EmailFrom.IsNull() && !data.EmailFrom.IsUnknown() {
		setStringValue(email, data.EmailFrom.ValueString(), "from")
	}
	if !data.EmailTo.IsNull() && !data.EmailTo.IsUnknown() {
		setStringValue(email, data.EmailTo.ValueString(), "to")
	}
	if !data.EmailSubject.IsNull() && !data.EmailSubject.IsUnknown() {
		setStringValue(email, data.EmailSubject.ValueString(), "subject")
	}

	// SNMP Notifications
	snmp := ensureNestedConfigMap(payload, "snmpNotifications")
	if !data.SNMPNotificationsEnabled.IsNull() && !data.SNMPNotificationsEnabled.IsUnknown() {
		setBoolValue(snmp, data.SNMPNotificationsEnabled.ValueBool(), "isEnabled")
	}

	// Syslog Notifications
	syslog := ensureNestedConfigMap(payload, "syslogNotifications")
	if !data.SyslogNotificationsEnabled.IsNull() && !data.SyslogNotificationsEnabled.IsUnknown() {
		setBoolValue(syslog, data.SyslogNotificationsEnabled.ValueBool(), "isEnabled")
	}
	if !data.SyslogServer.IsNull() && !data.SyslogServer.IsUnknown() {
		setStringValue(syslog, data.SyslogServer.ValueString(), "dnsName")
	}
	if !data.SyslogPort.IsNull() && !data.SyslogPort.IsUnknown() {
		setIntValue(syslog, int(data.SyslogPort.ValueInt64()), "port")
	}
}

// syncGeneralOptionsFromPayload populates the model from the API response map.
// String and int fields are only set when the API returns a non-empty / non-zero
// value. Bool fields are always populated (false is a meaningful value).
func syncGeneralOptionsFromPayload(payload map[string]interface{}, data *GeneralOptionsModel) {
	// Storage Latency Control
	slc := getNestedConfigMap(payload, "storageLatencyControl")
	data.StorageLatencyControlEnabled = types.BoolValue(getConfigBoolValue(slc, "isEnabled"))
	if v := getConfigIntValue(slc, "latencyLimitMs"); v != 0 {
		data.StorageLatencyLimitMs = types.Int64Value(int64(v))
	}

	// Email Notifications
	email := getNestedConfigMap(payload, "emailNotifications")
	data.EmailNotificationsEnabled = types.BoolValue(getConfigBoolValue(email, "isEnabled"))
	if v := getConfigStringValue(email, "smtpServer"); v != "" {
		data.EmailSMTPServer = types.StringValue(v)
	}
	if v := getConfigIntValue(email, "port"); v != 0 {
		data.EmailSMTPPort = types.Int64Value(int64(v))
	}
	if v := getConfigStringValue(email, "from"); v != "" {
		data.EmailFrom = types.StringValue(v)
	}
	if v := getConfigStringValue(email, "to"); v != "" {
		data.EmailTo = types.StringValue(v)
	}
	if v := getConfigStringValue(email, "subject"); v != "" {
		data.EmailSubject = types.StringValue(v)
	}

	// SNMP Notifications
	snmp := getNestedConfigMap(payload, "snmpNotifications")
	data.SNMPNotificationsEnabled = types.BoolValue(getConfigBoolValue(snmp, "isEnabled"))

	// Syslog Notifications
	syslog := getNestedConfigMap(payload, "syslogNotifications")
	data.SyslogNotificationsEnabled = types.BoolValue(getConfigBoolValue(syslog, "isEnabled"))
	if v := getConfigStringValue(syslog, "dnsName"); v != "" {
		data.SyslogServer = types.StringValue(v)
	}
	if v := getConfigIntValue(syslog, "port"); v != 0 {
		data.SyslogPort = types.Int64Value(int64(v))
	}
}

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

var (
	_ resource.Resource                = &NotificationSettings{}
	_ resource.ResourceWithConfigure   = &NotificationSettings{}
	_ resource.ResourceWithImportState = &NotificationSettings{}
)

// NotificationSettings manages the Veeam global notification settings singleton.
type NotificationSettings struct {
	client client.APIClient
}

// NotificationSettingsModel is the Terraform state model for veeam_notification_settings.
type NotificationSettingsModel struct {
	ID                             types.String `tfsdk:"id"`
	NotifyOnSuccess                types.Bool   `tfsdk:"notify_on_success"`
	NotifyOnWarning                types.Bool   `tfsdk:"notify_on_warning"`
	NotifyOnError                  types.Bool   `tfsdk:"notify_on_error"`
	SuppressRepeatingNotifications types.Bool   `tfsdk:"suppress_repeating_notifications"`
	NotifyOnLastRetryOnly          types.Bool   `tfsdk:"notify_on_last_retry_only"`
	SendSNMPOnSuccess              types.Bool   `tfsdk:"send_snmp_on_success"`
	SendSNMPOnWarning              types.Bool   `tfsdk:"send_snmp_on_warning"`
	SendSNMPOnError                types.Bool   `tfsdk:"send_snmp_on_error"`
	SendSyslogOnSuccess            types.Bool   `tfsdk:"send_syslog_on_success"`
	SendSyslogOnWarning            types.Bool   `tfsdk:"send_syslog_on_warning"`
	SendSyslogOnError              types.Bool   `tfsdk:"send_syslog_on_error"`
}

// NewNotificationSettings returns a new NotificationSettings resource.
func NewNotificationSettings() resource.Resource {
	return &NotificationSettings{}
}

func (r *NotificationSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_settings"
}

func (r *NotificationSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useStateForUnknownBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam Backup & Replication global notification settings. " +
			"This is a singleton resource — only one instance may exist. " +
			"Deleting the resource removes it from Terraform state only; it does not reset the server configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Fixed resource identifier. Always `notification-settings`.",
			},
			"notify_on_success": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send notifications when a job succeeds. Maps to API field `notifyOnSuccess`.",
			},
			"notify_on_warning": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send notifications when a job finishes with warnings. Maps to API field `notifyOnWarning`.",
			},
			"notify_on_error": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send notifications when a job fails. Maps to API field `notifyOnError`.",
			},
			"suppress_repeating_notifications": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Suppress duplicate notifications for repeating failures. Maps to API field `suppressRepeatingNotifications`.",
			},
			"notify_on_last_retry_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Only send a notification on the last retry attempt. Maps to API field `notifyOnLastRetryOnly`.",
			},
			"send_snmp_on_success": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send SNMP trap on job success. Maps to API field `sendSNMPOnSuccess`.",
			},
			"send_snmp_on_warning": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send SNMP trap on job warning. Maps to API field `sendSNMPOnWarning`.",
			},
			"send_snmp_on_error": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send SNMP trap on job error. Maps to API field `sendSNMPOnError`.",
			},
			"send_syslog_on_success": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send syslog message on job success. Maps to API field `sendSyslogOnSuccess`.",
			},
			"send_syslog_on_warning": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send syslog message on job warning. Maps to API field `sendSyslogOnWarning`.",
			},
			"send_syslog_on_error": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send syslog message on job error. Maps to API field `sendSyslogOnError`.",
			},
		},
	}
}

func (r *NotificationSettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putNotificationSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure notification settings", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("notification-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *NotificationSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathNotificationSettings, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read notification settings", fmt.Sprintf("API error: %s", err))
		return
	}

	syncNotificationSettingsFromAPI(&data, raw)
	data.ID = types.StringValue("notification-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *NotificationSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putNotificationSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update notification settings", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("notification-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *NotificationSettings) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Notification settings is a singleton — deletion is a no-op (removes from Terraform state only).
}

func (r *NotificationSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// putNotificationSettings GETs the current server config, merges plan fields, and PUTs the result.
func (r *NotificationSettings) putNotificationSettings(ctx context.Context, data *NotificationSettingsModel) error {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathNotificationSettings, &raw); err != nil {
		return fmt.Errorf("reading current notification settings: %w", err)
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}

	if !data.NotifyOnSuccess.IsNull() && !data.NotifyOnSuccess.IsUnknown() {
		setBoolValue(raw, data.NotifyOnSuccess.ValueBool(), "notifyOnSuccess")
	}
	if !data.NotifyOnWarning.IsNull() && !data.NotifyOnWarning.IsUnknown() {
		setBoolValue(raw, data.NotifyOnWarning.ValueBool(), "notifyOnWarning")
	}
	if !data.NotifyOnError.IsNull() && !data.NotifyOnError.IsUnknown() {
		setBoolValue(raw, data.NotifyOnError.ValueBool(), "notifyOnError")
	}
	if !data.SuppressRepeatingNotifications.IsNull() && !data.SuppressRepeatingNotifications.IsUnknown() {
		setBoolValue(raw, data.SuppressRepeatingNotifications.ValueBool(), "suppressRepeatingNotifications")
	}
	if !data.NotifyOnLastRetryOnly.IsNull() && !data.NotifyOnLastRetryOnly.IsUnknown() {
		setBoolValue(raw, data.NotifyOnLastRetryOnly.ValueBool(), "notifyOnLastRetryOnly")
	}
	if !data.SendSNMPOnSuccess.IsNull() && !data.SendSNMPOnSuccess.IsUnknown() {
		setBoolValue(raw, data.SendSNMPOnSuccess.ValueBool(), "sendSNMPOnSuccess")
	}
	if !data.SendSNMPOnWarning.IsNull() && !data.SendSNMPOnWarning.IsUnknown() {
		setBoolValue(raw, data.SendSNMPOnWarning.ValueBool(), "sendSNMPOnWarning")
	}
	if !data.SendSNMPOnError.IsNull() && !data.SendSNMPOnError.IsUnknown() {
		setBoolValue(raw, data.SendSNMPOnError.ValueBool(), "sendSNMPOnError")
	}
	if !data.SendSyslogOnSuccess.IsNull() && !data.SendSyslogOnSuccess.IsUnknown() {
		setBoolValue(raw, data.SendSyslogOnSuccess.ValueBool(), "sendSyslogOnSuccess")
	}
	if !data.SendSyslogOnWarning.IsNull() && !data.SendSyslogOnWarning.IsUnknown() {
		setBoolValue(raw, data.SendSyslogOnWarning.ValueBool(), "sendSyslogOnWarning")
	}
	if !data.SendSyslogOnError.IsNull() && !data.SendSyslogOnError.IsUnknown() {
		setBoolValue(raw, data.SendSyslogOnError.ValueBool(), "sendSyslogOnError")
	}

	return r.client.PutJSON(ctx, client.PathNotificationSettings, raw, nil)
}

// syncNotificationSettingsFromAPI maps API response fields into the Terraform model.
func syncNotificationSettingsFromAPI(data *NotificationSettingsModel, raw map[string]interface{}) {
	data.NotifyOnSuccess = types.BoolValue(getConfigBoolValue(raw, "notifyOnSuccess"))
	data.NotifyOnWarning = types.BoolValue(getConfigBoolValue(raw, "notifyOnWarning"))
	data.NotifyOnError = types.BoolValue(getConfigBoolValue(raw, "notifyOnError"))
	data.SuppressRepeatingNotifications = types.BoolValue(getConfigBoolValue(raw, "suppressRepeatingNotifications"))
	data.NotifyOnLastRetryOnly = types.BoolValue(getConfigBoolValue(raw, "notifyOnLastRetryOnly"))
	data.SendSNMPOnSuccess = types.BoolValue(getConfigBoolValue(raw, "sendSNMPOnSuccess"))
	data.SendSNMPOnWarning = types.BoolValue(getConfigBoolValue(raw, "sendSNMPOnWarning"))
	data.SendSNMPOnError = types.BoolValue(getConfigBoolValue(raw, "sendSNMPOnError"))
	data.SendSyslogOnSuccess = types.BoolValue(getConfigBoolValue(raw, "sendSyslogOnSuccess"))
	data.SendSyslogOnWarning = types.BoolValue(getConfigBoolValue(raw, "sendSyslogOnWarning"))
	data.SendSyslogOnError = types.BoolValue(getConfigBoolValue(raw, "sendSyslogOnError"))
}

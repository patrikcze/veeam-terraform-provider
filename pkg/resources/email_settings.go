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

var (
	_ resource.Resource                = &EmailSettings{}
	_ resource.ResourceWithConfigure   = &EmailSettings{}
	_ resource.ResourceWithImportState = &EmailSettings{}
)

// EmailSettings manages the Veeam email notification settings singleton.
type EmailSettings struct {
	client client.APIClient
}

// EmailSettingsModel is the Terraform state model for veeam_email_settings.
type EmailSettingsModel struct {
	ID                types.String `tfsdk:"id"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SMTPServer        types.String `tfsdk:"smtp_server"`
	Port              types.Int64  `tfsdk:"port"`
	UseSSL            types.Bool   `tfsdk:"use_ssl"`
	UseAuthentication types.Bool   `tfsdk:"use_authentication"`
	Login             types.String `tfsdk:"login"`
	Password          types.String `tfsdk:"password"`
	From              types.String `tfsdk:"from"`
	To                types.String `tfsdk:"to"`
	Subject           types.String `tfsdk:"subject"`
	SendOnSuccess     types.Bool   `tfsdk:"send_on_success"`
	SendOnWarning     types.Bool   `tfsdk:"send_on_warning"`
	SendOnError       types.Bool   `tfsdk:"send_on_error"`
	SendDailySummary  types.Bool   `tfsdk:"send_daily_summary"`
	SendTestMessage   types.Bool   `tfsdk:"send_test_message"`
}

// NewEmailSettings returns a new EmailSettings resource.
func NewEmailSettings() resource.Resource {
	return &EmailSettings{}
}

func (r *EmailSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_settings"
}

func (r *EmailSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useStateForUnknownBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	useStateForUnknownString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	useStateForUnknownInt64 := []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam Backup & Replication email notification settings. " +
			"This is a singleton resource — only one instance may exist. " +
			"Deleting the resource removes it from Terraform state only; it does not reset the server configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Fixed resource identifier. Always `email-settings`.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Whether email notifications are enabled. Maps to API field `isEnabled`.",
			},
			"smtp_server": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownString,
				MarkdownDescription: "SMTP server hostname or IP address.",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownInt64,
				MarkdownDescription: "SMTP server port number.",
			},
			"use_ssl": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Whether to use SSL/TLS for the SMTP connection. Maps to API field `useSSL`.",
			},
			"use_authentication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Whether SMTP authentication is required.",
			},
			"login": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownString,
				MarkdownDescription: "SMTP authentication login name.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "SMTP authentication password. Write-only; never read back from the API.",
			},
			"from": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownString,
				MarkdownDescription: "Sender email address.",
			},
			"to": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownString,
				MarkdownDescription: "Recipient email address.",
			},
			"subject": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownString,
				MarkdownDescription: "Email subject template.",
			},
			"send_on_success": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send notifications when a job succeeds.",
			},
			"send_on_warning": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send notifications when a job finishes with warnings.",
			},
			"send_on_error": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send notifications when a job fails.",
			},
			"send_daily_summary": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Send a daily summary email.",
			},
			"send_test_message": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "When true, sends a test email after each Create or Update apply. " +
					"This is an action trigger, not stored state.",
			},
		},
	}
}

func (r *EmailSettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EmailSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EmailSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putEmailSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure email settings", fmt.Sprintf("API error: %s", err))
		return
	}

	if !data.SendTestMessage.IsNull() && !data.SendTestMessage.IsUnknown() && data.SendTestMessage.ValueBool() {
		if err := r.client.PostJSON(ctx, client.PathEmailSettingsTestMessage, map[string]interface{}{}, nil); err != nil {
			resp.Diagnostics.AddError("Failed to send test email", fmt.Sprintf("API error: %s", err))
			return
		}
	}

	data.ID = types.StringValue("email-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EmailSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EmailSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathEmailSettings, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read email settings", fmt.Sprintf("API error: %s", err))
		return
	}

	r.syncModelFromAPI(&data, raw)
	data.ID = types.StringValue("email-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EmailSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EmailSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve password from prior state (write-only field — not returned by API)
	var state EmailSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Password.IsNull() || data.Password.IsUnknown() {
		data.Password = state.Password
	}

	if err := r.putEmailSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update email settings", fmt.Sprintf("API error: %s", err))
		return
	}

	if !data.SendTestMessage.IsNull() && !data.SendTestMessage.IsUnknown() && data.SendTestMessage.ValueBool() {
		if err := r.client.PostJSON(ctx, client.PathEmailSettingsTestMessage, map[string]interface{}{}, nil); err != nil {
			resp.Diagnostics.AddError("Failed to send test email", fmt.Sprintf("API error: %s", err))
			return
		}
	}

	data.ID = types.StringValue("email-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EmailSettings) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Email settings is a singleton — deletion is a no-op (removes from Terraform state only).
}

func (r *EmailSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// putEmailSettings GETs the current server config, merges plan fields, and PUTs the result.
func (r *EmailSettings) putEmailSettings(ctx context.Context, data *EmailSettingsModel) error {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathEmailSettings, &raw); err != nil {
		return fmt.Errorf("reading current email settings: %w", err)
	}

	if raw == nil {
		raw = map[string]interface{}{}
	}

	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		setBoolValue(raw, data.Enabled.ValueBool(), "isEnabled")
	}
	if !data.SMTPServer.IsNull() && !data.SMTPServer.IsUnknown() {
		setStringValue(raw, data.SMTPServer.ValueString(), "smtpServer")
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		setIntValue(raw, int(data.Port.ValueInt64()), "port")
	}
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		setBoolValue(raw, data.UseSSL.ValueBool(), "useSSL")
	}
	if !data.UseAuthentication.IsNull() && !data.UseAuthentication.IsUnknown() {
		setBoolValue(raw, data.UseAuthentication.ValueBool(), "useAuthentication")
	}
	if !data.Login.IsNull() && !data.Login.IsUnknown() {
		setStringValue(raw, data.Login.ValueString(), "login")
	}
	if !data.Password.IsNull() && !data.Password.IsUnknown() {
		setStringValue(raw, data.Password.ValueString(), "password")
	}
	if !data.From.IsNull() && !data.From.IsUnknown() {
		setStringValue(raw, data.From.ValueString(), "from")
	}
	if !data.To.IsNull() && !data.To.IsUnknown() {
		setStringValue(raw, data.To.ValueString(), "to")
	}
	if !data.Subject.IsNull() && !data.Subject.IsUnknown() {
		setStringValue(raw, data.Subject.ValueString(), "subject")
	}
	if !data.SendOnSuccess.IsNull() && !data.SendOnSuccess.IsUnknown() {
		setBoolValue(raw, data.SendOnSuccess.ValueBool(), "sendOnSuccess")
	}
	if !data.SendOnWarning.IsNull() && !data.SendOnWarning.IsUnknown() {
		setBoolValue(raw, data.SendOnWarning.ValueBool(), "sendOnWarning")
	}
	if !data.SendOnError.IsNull() && !data.SendOnError.IsUnknown() {
		setBoolValue(raw, data.SendOnError.ValueBool(), "sendOnError")
	}
	if !data.SendDailySummary.IsNull() && !data.SendDailySummary.IsUnknown() {
		setBoolValue(raw, data.SendDailySummary.ValueBool(), "sendDailySummary")
	}

	return r.client.PutJSON(ctx, client.PathEmailSettings, raw, nil)
}

// syncModelFromAPI merges API response fields into the Terraform model.
// password and send_test_message are write-only actions and are never overwritten.
func (r *EmailSettings) syncModelFromAPI(data *EmailSettingsModel, raw map[string]interface{}) {
	data.Enabled = types.BoolValue(getConfigBoolValue(raw, "isEnabled"))
	data.SMTPServer = types.StringValue(getConfigStringValue(raw, "smtpServer"))
	data.Port = types.Int64Value(int64(getConfigIntValue(raw, "port")))
	data.UseSSL = types.BoolValue(getConfigBoolValue(raw, "useSSL"))
	data.UseAuthentication = types.BoolValue(getConfigBoolValue(raw, "useAuthentication"))
	data.Login = types.StringValue(getConfigStringValue(raw, "login"))
	data.From = types.StringValue(getConfigStringValue(raw, "from"))
	data.To = types.StringValue(getConfigStringValue(raw, "to"))
	data.Subject = types.StringValue(getConfigStringValue(raw, "subject"))
	data.SendOnSuccess = types.BoolValue(getConfigBoolValue(raw, "sendOnSuccess"))
	data.SendOnWarning = types.BoolValue(getConfigBoolValue(raw, "sendOnWarning"))
	data.SendOnError = types.BoolValue(getConfigBoolValue(raw, "sendOnError"))
	data.SendDailySummary = types.BoolValue(getConfigBoolValue(raw, "sendDailySummary"))
	// password is write-only — never read from API response
	// send_test_message is an action trigger — not stored as state
}

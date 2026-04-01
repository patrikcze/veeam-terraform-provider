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
	_ resource.Resource                = &SecuritySettings{}
	_ resource.ResourceWithConfigure   = &SecuritySettings{}
	_ resource.ResourceWithImportState = &SecuritySettings{}
)

// SecuritySettings manages the Veeam server security settings singleton.
type SecuritySettings struct {
	client client.APIClient
}

// SecuritySettingsModel is the Terraform state model for veeam_security_settings.
type SecuritySettingsModel struct {
	ID                        types.String `tfsdk:"id"`
	RequireSSL                types.Bool   `tfsdk:"require_ssl"`
	RequireMFA                types.Bool   `tfsdk:"require_mfa"`
	BlockFirstLogin           types.Bool   `tfsdk:"block_first_login"`
	LoginAttemptLimit         types.Int64  `tfsdk:"login_attempt_limit"`
	InactivityTimeoutMin      types.Int64  `tfsdk:"inactivity_timeout_min"`
	PasswordExpirationDays    types.Int64  `tfsdk:"password_expiration_days"`
	PasswordExpirationEnabled types.Bool   `tfsdk:"password_expiration_enabled"`
}

// NewSecuritySettings returns a new SecuritySettings resource.
func NewSecuritySettings() resource.Resource {
	return &SecuritySettings{}
}

func (r *SecuritySettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_settings"
}

func (r *SecuritySettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useStateForUnknownBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	useStateForUnknownInt64 := []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam Backup & Replication server security settings. " +
			"This is a singleton resource — only one instance may exist. " +
			"Deleting the resource removes it from Terraform state only; it does not reset the server configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Fixed resource identifier. Always `security-settings`.",
			},
			"require_ssl": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Require SSL/TLS for all API connections. Maps to API field `requireSsl`.",
			},
			"require_mfa": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Require multi-factor authentication for all users. Maps to API field `requireMfa`.",
			},
			"block_first_login": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Block user accounts on first failed login attempt. Maps to API field `blockFirstLogin`.",
			},
			"login_attempt_limit": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownInt64,
				MarkdownDescription: "Maximum number of failed login attempts before an account is locked. Maps to API field `loginAttemptLimit`.",
			},
			"inactivity_timeout_min": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownInt64,
				MarkdownDescription: "Session inactivity timeout in minutes. Maps to API field `inactivityTimeoutMin`.",
			},
			"password_expiration_days": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownInt64,
				MarkdownDescription: "Number of days before passwords expire. Maps to API field `passwordExpirationDays`.",
			},
			"password_expiration_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       useStateForUnknownBool,
				MarkdownDescription: "Whether password expiration is enforced. Maps to API field `passwordExpirationEnabled`.",
			},
		},
	}
}

func (r *SecuritySettings) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecuritySettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecuritySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putSecuritySettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure security settings", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("security-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *SecuritySettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecuritySettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathSecuritySettings, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read security settings", fmt.Sprintf("API error: %s", err))
		return
	}

	syncSecuritySettingsFromAPI(&data, raw)
	data.ID = types.StringValue("security-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *SecuritySettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecuritySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putSecuritySettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update security settings", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("security-settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *SecuritySettings) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Security settings is a singleton — deletion is a no-op (removes from Terraform state only).
}

func (r *SecuritySettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// putSecuritySettings GETs the current server config, merges plan fields, and PUTs the result.
func (r *SecuritySettings) putSecuritySettings(ctx context.Context, data *SecuritySettingsModel) error {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathSecuritySettings, &raw); err != nil {
		return fmt.Errorf("reading current security settings: %w", err)
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}

	if !data.RequireSSL.IsNull() && !data.RequireSSL.IsUnknown() {
		setBoolValue(raw, data.RequireSSL.ValueBool(), "requireSsl")
	}
	if !data.RequireMFA.IsNull() && !data.RequireMFA.IsUnknown() {
		setBoolValue(raw, data.RequireMFA.ValueBool(), "requireMfa")
	}
	if !data.BlockFirstLogin.IsNull() && !data.BlockFirstLogin.IsUnknown() {
		setBoolValue(raw, data.BlockFirstLogin.ValueBool(), "blockFirstLogin")
	}
	if !data.LoginAttemptLimit.IsNull() && !data.LoginAttemptLimit.IsUnknown() {
		setIntValue(raw, int(data.LoginAttemptLimit.ValueInt64()), "loginAttemptLimit")
	}
	if !data.InactivityTimeoutMin.IsNull() && !data.InactivityTimeoutMin.IsUnknown() {
		setIntValue(raw, int(data.InactivityTimeoutMin.ValueInt64()), "inactivityTimeoutMin")
	}
	if !data.PasswordExpirationDays.IsNull() && !data.PasswordExpirationDays.IsUnknown() {
		setIntValue(raw, int(data.PasswordExpirationDays.ValueInt64()), "passwordExpirationDays")
	}
	if !data.PasswordExpirationEnabled.IsNull() && !data.PasswordExpirationEnabled.IsUnknown() {
		setBoolValue(raw, data.PasswordExpirationEnabled.ValueBool(), "passwordExpirationEnabled")
	}

	return r.client.PutJSON(ctx, client.PathSecuritySettings, raw, nil)
}

// syncSecuritySettingsFromAPI maps API response fields into the Terraform model.
func syncSecuritySettingsFromAPI(data *SecuritySettingsModel, raw map[string]interface{}) {
	data.RequireSSL = types.BoolValue(getConfigBoolValue(raw, "requireSsl"))
	data.RequireMFA = types.BoolValue(getConfigBoolValue(raw, "requireMfa"))
	data.BlockFirstLogin = types.BoolValue(getConfigBoolValue(raw, "blockFirstLogin"))
	data.LoginAttemptLimit = types.Int64Value(int64(getConfigIntValue(raw, "loginAttemptLimit")))
	data.InactivityTimeoutMin = types.Int64Value(int64(getConfigIntValue(raw, "inactivityTimeoutMin")))
	data.PasswordExpirationDays = types.Int64Value(int64(getConfigIntValue(raw, "passwordExpirationDays")))
	data.PasswordExpirationEnabled = types.BoolValue(getConfigBoolValue(raw, "passwordExpirationEnabled"))
}

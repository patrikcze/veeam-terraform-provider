package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

var (
	_ resource.Resource                = &ConfigurationBackup{}
	_ resource.ResourceWithConfigure   = &ConfigurationBackup{}
	_ resource.ResourceWithImportState = &ConfigurationBackup{}
)

type ConfigurationBackup struct {
	client client.APIClient
}

type ConfigurationBackupModel struct {
	ID                   types.String `tfsdk:"id"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	RepositoryID         types.String `tfsdk:"repository_id"`
	RestorePointsToKeep  types.Int64  `tfsdk:"restore_points_to_keep"`
	EncryptionEnabled    types.Bool   `tfsdk:"encryption_enabled"`
	EncryptionPasswordID types.String `tfsdk:"encryption_password_id"`
	TriggerOnApply       types.Bool   `tfsdk:"trigger_on_apply"`
	LastSessionID        types.String `tfsdk:"last_session_id"`
	LastSessionState     types.String `tfsdk:"last_session_state"`
	LastSessionResult    types.String `tfsdk:"last_session_result"`
}

func (r *ConfigurationBackup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration_backup"
}

func (r *ConfigurationBackup) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam configuration backup settings and can trigger a configuration backup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled":                schema.BoolAttribute{Required: true},
			"repository_id":          schema.StringAttribute{Optional: true},
			"restore_points_to_keep": schema.Int64Attribute{Optional: true},
			"encryption_enabled":     schema.BoolAttribute{Optional: true},
			"encryption_password_id": schema.StringAttribute{Optional: true},
			"trigger_on_apply":       schema.BoolAttribute{Optional: true},
			"last_session_id":        schema.StringAttribute{Computed: true},
			"last_session_state":     schema.StringAttribute{Computed: true},
			"last_session_result":    schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ConfigurationBackup) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider, got unexpected type.")
		return
	}
	r.client = c
}

func (r *ConfigurationBackup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure backup settings", fmt.Sprintf("API error: %s", err))
		return
	}

	if !data.TriggerOnApply.IsNull() && data.TriggerOnApply.ValueBool() {
		if err := r.triggerBackup(ctx, &data); err != nil {
			resp.Diagnostics.AddError("Failed to trigger configuration backup", fmt.Sprintf("API error: %s", err))
			return
		}
	}

	ensureKnownLastSessionFields(&data)

	data.ID = types.StringValue("config-backup")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ConfigurationBackup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathConfigurationBackup, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read configuration backup", fmt.Sprintf("API error: %s", err))
		return
	}
	payload := unwrapConfigBackupPayload(result)

	data.ID = types.StringValue("config-backup")
	data.Enabled = types.BoolValue(getConfigBoolValue(payload, "isEnabled", "enabled"))

	repositoryID := getConfigStringValue(payload, "backupRepositoryId", "repositoryId")
	if repositoryID != "" {
		data.RepositoryID = types.StringValue(repositoryID)
	}

	data.RestorePointsToKeep = types.Int64Value(int64(getConfigIntValue(payload, "restorePointsToKeep")))

	encryption := getNestedConfigMap(payload, "encryption")
	if len(encryption) > 0 {
		data.EncryptionEnabled = types.BoolValue(getConfigBoolValue(encryption, "isEnabled"))
		passwordID := getConfigStringValue(encryption, "passwordId", "encryptionPasswordId")
		if passwordID != "" {
			data.EncryptionPasswordID = types.StringValue(passwordID)
		}
	} else {
		data.EncryptionEnabled = types.BoolValue(getConfigBoolValue(payload, "encryptionEnabled"))
		passwordID := getConfigStringValue(payload, "encryptionPasswordId")
		if passwordID != "" {
			data.EncryptionPasswordID = types.StringValue(passwordID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ConfigurationBackup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update configuration backup", fmt.Sprintf("API error: %s", err))
		return
	}

	if !data.TriggerOnApply.IsNull() && data.TriggerOnApply.ValueBool() {
		if err := r.triggerBackup(ctx, &data); err != nil {
			resp.Diagnostics.AddError("Failed to trigger configuration backup", fmt.Sprintf("API error: %s", err))
			return
		}
	}

	ensureKnownLastSessionFields(&data)

	data.ID = types.StringValue("config-backup")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ConfigurationBackup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigurationBackupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := r.loadConfigurationBackupPayload(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to disable configuration backup", fmt.Sprintf("API error: %s", err))
		return
	}

	setBoolValue(payload, false, "enabled", "isEnabled")
	encryption := ensureNestedConfigMap(payload, "encryption")
	setBoolValue(encryption, false, "isEnabled", "enabled")

	if err := r.putConfigurationPayload(ctx, payload); err != nil {
		resp.Diagnostics.AddError("Failed to disable configuration backup", fmt.Sprintf("API error: %s", err))
	}
}

func (r *ConfigurationBackup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewConfigurationBackup() resource.Resource {
	return &ConfigurationBackup{}
}

func (r *ConfigurationBackup) putConfig(ctx context.Context, data *ConfigurationBackupModel) error {
	payload, err := r.loadConfigurationBackupPayload(ctx)
	if err != nil {
		return err
	}

	setBoolValue(payload, data.Enabled.ValueBool(), "enabled", "isEnabled")

	if !data.RepositoryID.IsNull() && !data.RepositoryID.IsUnknown() {
		setStringValue(payload, data.RepositoryID.ValueString(), "backupRepositoryId", "repositoryId")
	}

	if !data.RestorePointsToKeep.IsNull() && !data.RestorePointsToKeep.IsUnknown() {
		setIntValue(payload, int(data.RestorePointsToKeep.ValueInt64()), "restorePointsToKeep")
	}

	encryption := ensureNestedConfigMap(payload, "encryption")
	if !data.EncryptionEnabled.IsNull() && !data.EncryptionEnabled.IsUnknown() {
		setBoolValue(encryption, data.EncryptionEnabled.ValueBool(), "isEnabled", "enabled")
	}

	if !data.EncryptionPasswordID.IsNull() && !data.EncryptionPasswordID.IsUnknown() {
		setStringValue(encryption, data.EncryptionPasswordID.ValueString(), "passwordId", "encryptionPasswordId")
	}

	return r.putConfigurationPayload(ctx, payload)
}

func (r *ConfigurationBackup) putConfigurationPayload(ctx context.Context, payload map[string]interface{}) error {
	if err := r.client.PutJSON(ctx, client.PathConfigurationBackup, payload, nil); err != nil {
		if isSNMPGeneralOptionsError(err) {
			notifications := ensureNestedConfigMap(payload, "notifications")
			setBoolValue(notifications, false, "SNMPEnabled", "snmpEnabled", "enabled")
			return r.client.PutJSON(ctx, client.PathConfigurationBackup, payload, nil)
		}
		return err
	}

	return nil
}

func (r *ConfigurationBackup) triggerBackup(ctx context.Context, data *ConfigurationBackupModel) error {
	var session models.ConfigurationBackupSessionModel
	if err := r.client.PostJSON(ctx, client.PathConfigurationBackupStart, map[string]interface{}{}, &session); err != nil {
		return err
	}

	if session.ID != "" {
		data.LastSessionID = types.StringValue(session.ID)
	}
	if session.State != "" {
		data.LastSessionState = types.StringValue(session.State)
	}
	if session.Result != "" {
		data.LastSessionResult = types.StringValue(session.Result)
	}

	return nil
}

func (r *ConfigurationBackup) loadConfigurationBackupPayload(ctx context.Context) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathConfigurationBackup, &raw); err != nil {
		return nil, err
	}

	payload := unwrapConfigBackupPayload(raw)
	if payload == nil {
		return map[string]interface{}{}, nil
	}

	return payload, nil
}

func unwrapConfigBackupPayload(raw map[string]interface{}) map[string]interface{} {
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

func setBoolValue(payload map[string]interface{}, value bool, keys ...string) {
	for _, key := range keys {
		if _, exists := payload[key]; exists {
			payload[key] = value
			return
		}
	}

	if len(keys) > 0 {
		payload[keys[0]] = value
	}
}

func setStringValue(payload map[string]interface{}, value string, keys ...string) {
	for _, key := range keys {
		if _, exists := payload[key]; exists {
			payload[key] = value
			return
		}
	}

	if len(keys) > 0 {
		payload[keys[0]] = value
	}
}

func setIntValue(payload map[string]interface{}, value int, keys ...string) {
	for _, key := range keys {
		if _, exists := payload[key]; exists {
			payload[key] = value
			return
		}
	}

	if len(keys) > 0 {
		payload[keys[0]] = value
	}
}

func getConfigStringValue(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			if asString, ok := value.(string); ok {
				return asString
			}
		}
	}

	return ""
}

func getConfigBoolValue(payload map[string]interface{}, keys ...string) bool {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			if asBool, ok := value.(bool); ok {
				return asBool
			}
		}
	}

	return false
}

func getConfigIntValue(payload map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch typed := value.(type) {
			case int:
				return typed
			case int64:
				return int(typed)
			case float64:
				return int(typed)
			}
		}
	}

	return 0
}

func getNestedConfigMap(payload map[string]interface{}, key string) map[string]interface{} {
	if value, ok := payload[key]; ok {
		if nested, ok := value.(map[string]interface{}); ok {
			return nested
		}
	}

	return map[string]interface{}{}
}

func ensureNestedConfigMap(payload map[string]interface{}, key string) map[string]interface{} {
	if existing := getNestedConfigMap(payload, key); len(existing) > 0 {
		return existing
	}

	nested := map[string]interface{}{}
	payload[key] = nested
	return nested
}

func ensureKnownLastSessionFields(data *ConfigurationBackupModel) {
	if data.LastSessionID.IsUnknown() || data.LastSessionID.IsNull() {
		data.LastSessionID = types.StringNull()
	}
	if data.LastSessionState.IsUnknown() || data.LastSessionState.IsNull() {
		data.LastSessionState = types.StringNull()
	}
	if data.LastSessionResult.IsUnknown() || data.LastSessionResult.IsNull() {
		data.LastSessionResult = types.StringNull()
	}
}

func isSNMPGeneralOptionsError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "specify snmp settings in general options")
}

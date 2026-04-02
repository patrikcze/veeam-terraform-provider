package resources

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = &TrafficRules{}
	_ resource.ResourceWithConfigure   = &TrafficRules{}
	_ resource.ResourceWithImportState = &TrafficRules{}
)

// TrafficRules manages the Veeam network traffic throttling rules singleton.
type TrafficRules struct {
	client client.APIClient
}

// TrafficRulesModel is the Terraform state model for veeam_traffic_rules.
type TrafficRulesModel struct {
	ID                types.String `tfsdk:"id"`
	ThrottlingEnabled types.Bool   `tfsdk:"throttling_enabled"`
	ThrottlingRules   types.String `tfsdk:"throttling_rules"`
}

// NewTrafficRules returns a new TrafficRules resource.
func NewTrafficRules() resource.Resource {
	return &TrafficRules{}
}

func (r *TrafficRules) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_rules"
}

func (r *TrafficRules) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Veeam Backup & Replication network traffic throttling rules. " +
			"This is a singleton resource — only one instance may exist. " +
			"Deleting the resource removes it from Terraform state only; it does not reset the server configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Fixed resource identifier. Always `traffic-rules`.",
			},
			"throttling_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Whether network traffic throttling is enabled. Maps to API field `throttlingEnabled`.",
			},
			"throttling_rules": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "JSON-encoded array of throttling rule objects from the API `rules` field. " +
					"Pass a raw JSON array string (e.g., `\"[]\"`). " +
					"The structure of each rule object matches the Veeam REST API specification.",
			},
		},
	}
}

func (r *TrafficRules) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TrafficRules) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TrafficRulesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putTrafficRules(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure traffic rules", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("traffic-rules")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *TrafficRules) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TrafficRulesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathTrafficRules, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read traffic rules", fmt.Sprintf("API error: %s", err))
		return
	}

	if err := syncTrafficRulesFromAPI(&data, raw); err != nil {
		resp.Diagnostics.AddError("Failed to parse traffic rules response", fmt.Sprintf("JSON marshal error: %s", err))
		return
	}

	data.ID = types.StringValue("traffic-rules")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *TrafficRules) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TrafficRulesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.putTrafficRules(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update traffic rules", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue("traffic-rules")
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *TrafficRules) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Traffic rules is a singleton — deletion is a no-op (removes from Terraform state only).
}

func (r *TrafficRules) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// putTrafficRules GETs the current server config, merges plan fields, and PUTs the result.
func (r *TrafficRules) putTrafficRules(ctx context.Context, data *TrafficRulesModel) error {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathTrafficRules, &raw); err != nil {
		return fmt.Errorf("reading current traffic rules: %w", err)
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}

	if !data.ThrottlingEnabled.IsNull() && !data.ThrottlingEnabled.IsUnknown() {
		setBoolValue(raw, data.ThrottlingEnabled.ValueBool(), "throttlingEnabled")
	}

	if !data.ThrottlingRules.IsNull() && !data.ThrottlingRules.IsUnknown() {
		var rules []interface{}
		if err := json.Unmarshal([]byte(data.ThrottlingRules.ValueString()), &rules); err != nil {
			return fmt.Errorf("parsing throttling_rules JSON: %w", err)
		}
		raw["rules"] = rules
	}

	return r.client.PutJSON(ctx, client.PathTrafficRules, raw, nil)
}

// syncTrafficRulesFromAPI maps API response fields into the Terraform model.
// The rules array is serialized to a JSON string for storage.
func syncTrafficRulesFromAPI(data *TrafficRulesModel, raw map[string]interface{}) error {
	data.ThrottlingEnabled = types.BoolValue(getConfigBoolValue(raw, "throttlingEnabled"))

	if rulesRaw, ok := raw["rules"]; ok {
		b, err := json.Marshal(rulesRaw)
		if err != nil {
			return fmt.Errorf("serializing rules to JSON: %w", err)
		}
		data.ThrottlingRules = types.StringValue(string(b))
	} else {
		data.ThrottlingRules = types.StringValue("[]")
	}

	return nil
}

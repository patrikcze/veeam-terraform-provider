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
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &SecurityAnalyzerSchedule{}
	_ resource.ResourceWithConfigure   = &SecurityAnalyzerSchedule{}
	_ resource.ResourceWithImportState = &SecurityAnalyzerSchedule{}
)

// securityAnalyzerScheduleID is the fixed singleton ID used in Terraform state.
const securityAnalyzerScheduleID = "security-analyzer-schedule"

// SecurityAnalyzerSchedule manages the Veeam security analyzer schedule singleton
// (GET/PUT /api/v1/securityAnalyzer/schedule). Because the object always exists
// and cannot be deleted, Create/Update both issue a PUT and Delete simply removes
// the resource from Terraform state.
type SecurityAnalyzerSchedule struct {
	client client.APIClient
}

// SecurityAnalyzerScheduleModel is the Terraform state model for veeam_security_analyzer_schedule.
type SecurityAnalyzerScheduleModel struct {
	ID                types.String `tfsdk:"id"`
	RunAutomatically  types.Bool   `tfsdk:"run_automatically"`
	DailyEnabled      types.Bool   `tfsdk:"daily_enabled"`
	DailyLocalTime    types.String `tfsdk:"daily_local_time"`
	MonthlyEnabled    types.Bool   `tfsdk:"monthly_enabled"`
	MonthlyDayOfMonth types.Int64  `tfsdk:"monthly_day_of_month"`
}

// NewSecurityAnalyzerSchedule returns a new veeam_security_analyzer_schedule resource instance.
func NewSecurityAnalyzerSchedule() resource.Resource {
	return &SecurityAnalyzerSchedule{}
}

func (r *SecurityAnalyzerSchedule) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_analyzer_schedule"
}

func (r *SecurityAnalyzerSchedule) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Veeam security analyzer scan schedule singleton " +
			"(`/api/v1/securityAnalyzer/schedule`). This is a singleton resource — only one " +
			"instance may exist per provider configuration. Deleting the resource only removes it " +
			"from Terraform state; the server-side schedule configuration is not reset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `\"security-analyzer-schedule\"`. Fixed singleton identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"run_automatically": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the security analyzer runs on schedule automatically (`runAutomatically`).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"daily_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the daily scan schedule is enabled (`daily.isEnabled`).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"daily_local_time": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Local time for the daily scan in `HH:MM` format (`daily.localTime`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"monthly_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the monthly scan schedule is enabled (`monthly.isEnabled`).",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"monthly_day_of_month": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Day of the month (1–31) on which the monthly scan runs (`monthly.dayOfMonth`).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SecurityAnalyzerSchedule) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SecurityAnalyzerSchedule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityAnalyzerScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applySchedule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure security analyzer schedule", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(securityAnalyzerScheduleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *SecurityAnalyzerSchedule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityAnalyzerScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.SecurityAnalyzerScheduleModel
	if err := r.client.GetJSON(ctx, client.PathSecurityAnalyzerSchedule, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read security analyzer schedule", fmt.Sprintf("API error: %s", err))
		return
	}

	syncSecurityAnalyzerScheduleFromAPI(&data, &result)
	data.ID = types.StringValue(securityAnalyzerScheduleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update applies plan changes via GET → merge → PUT.
func (r *SecurityAnalyzerSchedule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityAnalyzerScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applySchedule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update security analyzer schedule", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(securityAnalyzerScheduleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete removes the resource from Terraform state only. The server-side
// configuration object cannot be deleted via the API.
func (r *SecurityAnalyzerSchedule) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentional no-op: singletons cannot be deleted server-side.
}

func (r *SecurityAnalyzerSchedule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

// applySchedule GETs the current schedule, merges plan values in, and PUTs the result.
func (r *SecurityAnalyzerSchedule) applySchedule(ctx context.Context, data *SecurityAnalyzerScheduleModel) error {
	var current models.SecurityAnalyzerScheduleModel
	if err := r.client.GetJSON(ctx, client.PathSecurityAnalyzerSchedule, &current); err != nil {
		return fmt.Errorf("reading current security analyzer schedule: %w", err)
	}

	if !data.RunAutomatically.IsNull() && !data.RunAutomatically.IsUnknown() {
		current.RunAutomatically = data.RunAutomatically.ValueBool()
	}

	if !data.DailyEnabled.IsNull() && !data.DailyEnabled.IsUnknown() ||
		!data.DailyLocalTime.IsNull() && !data.DailyLocalTime.IsUnknown() {
		if current.Daily == nil {
			current.Daily = &models.SecurityAnalyzerDailySchedule{}
		}
		if !data.DailyEnabled.IsNull() && !data.DailyEnabled.IsUnknown() {
			current.Daily.IsEnabled = data.DailyEnabled.ValueBool()
		}
		if !data.DailyLocalTime.IsNull() && !data.DailyLocalTime.IsUnknown() {
			current.Daily.LocalTime = data.DailyLocalTime.ValueString()
		}
	}

	if !data.MonthlyEnabled.IsNull() && !data.MonthlyEnabled.IsUnknown() ||
		!data.MonthlyDayOfMonth.IsNull() && !data.MonthlyDayOfMonth.IsUnknown() {
		if current.Monthly == nil {
			current.Monthly = &models.SecurityAnalyzerMonthlySchedule{}
		}
		if !data.MonthlyEnabled.IsNull() && !data.MonthlyEnabled.IsUnknown() {
			current.Monthly.IsEnabled = data.MonthlyEnabled.ValueBool()
		}
		if !data.MonthlyDayOfMonth.IsNull() && !data.MonthlyDayOfMonth.IsUnknown() {
			current.Monthly.DayOfMonth = int(data.MonthlyDayOfMonth.ValueInt64())
		}
	}

	return r.client.PutJSON(ctx, client.PathSecurityAnalyzerSchedule, &current, nil)
}

// syncSecurityAnalyzerScheduleFromAPI populates the model from the API response.
func syncSecurityAnalyzerScheduleFromAPI(data *SecurityAnalyzerScheduleModel, api *models.SecurityAnalyzerScheduleModel) {
	data.RunAutomatically = types.BoolValue(api.RunAutomatically)

	if api.Daily != nil {
		data.DailyEnabled = types.BoolValue(api.Daily.IsEnabled)
		if api.Daily.LocalTime != "" {
			data.DailyLocalTime = types.StringValue(api.Daily.LocalTime)
		}
	} else {
		data.DailyEnabled = types.BoolValue(false)
	}

	if api.Monthly != nil {
		data.MonthlyEnabled = types.BoolValue(api.Monthly.IsEnabled)
		if api.Monthly.DayOfMonth != 0 {
			data.MonthlyDayOfMonth = types.Int64Value(int64(api.Monthly.DayOfMonth))
		}
	} else {
		data.MonthlyEnabled = types.BoolValue(false)
	}
}

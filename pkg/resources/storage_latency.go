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
	_ resource.Resource                = &StorageLatency{}
	_ resource.ResourceWithConfigure   = &StorageLatency{}
	_ resource.ResourceWithImportState = &StorageLatency{}
)

// storageLatencyID is the fixed singleton ID used in Terraform state.
const storageLatencyID = "storage-latency"

// StorageLatency manages the Veeam storage latency control singleton
// (GET/PUT /api/v1/generalOptions/storageLatency).
type StorageLatency struct {
	client client.APIClient
}

// StorageLatencyModel is the Terraform state model for veeam_storage_latency.
type StorageLatencyModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	LatencyLimitMs      types.Int64  `tfsdk:"latency_limit_ms"`
	ThrottlingIOEnabled types.Bool   `tfsdk:"throttling_io_enabled"`
	ThrottlingIOLimit   types.Int64  `tfsdk:"throttling_io_limit"`
	StopJobsEnabled     types.Bool   `tfsdk:"stop_jobs_enabled"`
	StopJobsLimitMs     types.Int64  `tfsdk:"stop_jobs_limit_ms"`
}

// NewStorageLatency returns a new veeam_storage_latency resource instance.
func NewStorageLatency() resource.Resource {
	return &StorageLatency{}
}

func (r *StorageLatency) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_latency"
}

func (r *StorageLatency) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useStateString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	useStateBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	useStateInt := []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Veeam storage latency control singleton " +
			"(`/api/v1/generalOptions/storageLatency`). This is a singleton resource — only one " +
			"instance may exist per provider configuration. Deleting the resource only removes it " +
			"from Terraform state; the server configuration is not reset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `\"storage-latency\"`. Fixed singleton identifier.",
				PlanModifiers:       useStateString,
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether global storage latency control is enabled.",
				PlanModifiers:       useStateBool,
			},
			"latency_limit_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Global latency threshold in milliseconds. Jobs are throttled when the datastore latency exceeds this value.",
				PlanModifiers:       useStateInt,
			},
			"throttling_io_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether IOPS throttling is enabled when the latency limit is exceeded.",
				PlanModifiers:       useStateBool,
			},
			"throttling_io_limit": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maximum IOPS allowed for backup operations when throttling is active.",
				PlanModifiers:       useStateInt,
			},
			"stop_jobs_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether backup jobs should be stopped when the stop-jobs latency threshold is exceeded.",
				PlanModifiers:       useStateBool,
			},
			"stop_jobs_limit_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Latency threshold in milliseconds above which backup jobs are stopped.",
				PlanModifiers:       useStateInt,
			},
		},
	}
}

func (r *StorageLatency) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *StorageLatency) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StorageLatencyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyStorageLatency(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to configure storage latency", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(storageLatencyID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageLatency) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageLatencyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathStorageLatency, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read storage latency", fmt.Sprintf("API error: %s", err))
		return
	}

	syncStorageLatencyFromPayload(raw, &data)
	data.ID = types.StringValue(storageLatencyID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update applies plan changes via GET → merge → PUT.
func (r *StorageLatency) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StorageLatencyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyStorageLatency(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to update storage latency", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(storageLatencyID)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete removes the resource from Terraform state only. The server-side
// configuration cannot be deleted via the API.
func (r *StorageLatency) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentional no-op: singletons cannot be deleted server-side.
}

func (r *StorageLatency) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

// applyStorageLatency GETs the current payload, merges plan values in, and PUTs the result.
func (r *StorageLatency) applyStorageLatency(ctx context.Context, data *StorageLatencyModel) error {
	var raw map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathStorageLatency, &raw); err != nil {
		return fmt.Errorf("reading current storage latency settings: %w", err)
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}

	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		setBoolValue(raw, data.Enabled.ValueBool(), "isEnabled")
	}
	if !data.LatencyLimitMs.IsNull() && !data.LatencyLimitMs.IsUnknown() {
		setIntValue(raw, int(data.LatencyLimitMs.ValueInt64()), "latencyLimitMs")
	}

	throttle := ensureNestedConfigMap(raw, "throttlingIo")
	if !data.ThrottlingIOEnabled.IsNull() && !data.ThrottlingIOEnabled.IsUnknown() {
		setBoolValue(throttle, data.ThrottlingIOEnabled.ValueBool(), "isEnabled")
	}
	if !data.ThrottlingIOLimit.IsNull() && !data.ThrottlingIOLimit.IsUnknown() {
		setIntValue(throttle, int(data.ThrottlingIOLimit.ValueInt64()), "iopsLimit")
	}

	stopJobs := ensureNestedConfigMap(raw, "stopJobs")
	if !data.StopJobsEnabled.IsNull() && !data.StopJobsEnabled.IsUnknown() {
		setBoolValue(stopJobs, data.StopJobsEnabled.ValueBool(), "isEnabled")
	}
	if !data.StopJobsLimitMs.IsNull() && !data.StopJobsLimitMs.IsUnknown() {
		setIntValue(stopJobs, int(data.StopJobsLimitMs.ValueInt64()), "latencyLimitMs")
	}

	return r.client.PutJSON(ctx, client.PathStorageLatency, raw, nil)
}

// syncStorageLatencyFromPayload populates the model from the API response map.
func syncStorageLatencyFromPayload(raw map[string]interface{}, data *StorageLatencyModel) {
	data.Enabled = types.BoolValue(getConfigBoolValue(raw, "isEnabled"))
	if v := getConfigIntValue(raw, "latencyLimitMs"); v != 0 {
		data.LatencyLimitMs = types.Int64Value(int64(v))
	}

	throttle := getNestedConfigMap(raw, "throttlingIo")
	data.ThrottlingIOEnabled = types.BoolValue(getConfigBoolValue(throttle, "isEnabled"))
	if v := getConfigIntValue(throttle, "iopsLimit"); v != 0 {
		data.ThrottlingIOLimit = types.Int64Value(int64(v))
	}

	stopJobs := getNestedConfigMap(raw, "stopJobs")
	data.StopJobsEnabled = types.BoolValue(getConfigBoolValue(stopJobs, "isEnabled"))
	if v := getConfigIntValue(stopJobs, "latencyLimitMs"); v != 0 {
		data.StopJobsLimitMs = types.Int64Value(int64(v))
	}
}

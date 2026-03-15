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
	_ resource.Resource                = &BackupJob{}
	_ resource.ResourceWithConfigure   = &BackupJob{}
	_ resource.ResourceWithImportState = &BackupJob{}
)

// BackupJob implements the veeam_backup_job resource.
type BackupJob struct {
	client client.APIClient
}

// BackupJobModel is the Terraform state model for veeam_backup_job.
type BackupJobModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	IsHighPriority     types.Bool   `tfsdk:"is_high_priority"`
	RepositoryID       types.String `tfsdk:"repository_id"`
	ProxyAutoSelect    types.Bool   `tfsdk:"proxy_auto_select"`
	RetentionType      types.String `tfsdk:"retention_type"`
	RetentionQuantity  types.Int64  `tfsdk:"retention_quantity"`
	ScheduleEnabled    types.Bool   `tfsdk:"schedule_enabled"`
	ScheduleTime       types.String `tfsdk:"schedule_time"`
	ScheduleKind       types.String `tfsdk:"schedule_kind"`
	RetryEnabled       types.Bool   `tfsdk:"retry_enabled"`
	RetryCount         types.Int64  `tfsdk:"retry_count"`
	RetryAwaitMinutes  types.Int64  `tfsdk:"retry_await_minutes"`
	IsDisabled         types.Bool   `tfsdk:"is_disabled"`
}

func (r *BackupJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_job"
}

func (r *BackupJob) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam backup job.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Job identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Job name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Job type: `Backup`, `BackupCopy`, `VSphereReplica`, etc.",
				Required:            true,
			},
			"is_high_priority": schema.BoolAttribute{
				MarkdownDescription: "Process this job before lower-priority jobs.",
				Optional:            true,
			},
			// --- Storage ---
			"repository_id": schema.StringAttribute{
				MarkdownDescription: "Target backup repository ID.",
				Required:            true,
			},
			"proxy_auto_select": schema.BoolAttribute{
				MarkdownDescription: "Automatically select backup proxy.",
				Optional:            true,
			},
			"retention_type": schema.StringAttribute{
				MarkdownDescription: "Retention type: `RestorePoints` or `Days`.",
				Optional:            true,
			},
			"retention_quantity": schema.Int64Attribute{
				MarkdownDescription: "Number of restore points or days to keep.",
				Optional:            true,
			},
			// --- Schedule ---
			"schedule_enabled": schema.BoolAttribute{
				MarkdownDescription: "Run the job automatically on schedule.",
				Optional:            true,
			},
			"schedule_time": schema.StringAttribute{
				MarkdownDescription: "Daily schedule time (e.g. `22:00`).",
				Optional:            true,
			},
			"schedule_kind": schema.StringAttribute{
				MarkdownDescription: "Daily schedule kind: `Everyday`, `Weekdays`, or `SelectedDays`.",
				Optional:            true,
			},
			"retry_enabled": schema.BoolAttribute{
				MarkdownDescription: "Retry on failure.",
				Optional:            true,
			},
			"retry_count": schema.Int64Attribute{
				MarkdownDescription: "Number of retry attempts.",
				Optional:            true,
			},
			"retry_await_minutes": schema.Int64Attribute{
				MarkdownDescription: "Minutes to wait between retries.",
				Optional:            true,
			},
			"is_disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the job is disabled (read-only).",
				Computed:            true,
			},
		},
	}
}

func (r *BackupJob) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BackupJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.BackupJobModel
	if err := r.client.PostJSON(ctx, client.PathJobs, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create backup job",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *BackupJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.BackupJobModel
	endpoint := fmt.Sprintf(client.PathJobByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read backup job",
			fmt.Sprintf("API error for job %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *BackupJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathJobByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update backup job",
			fmt.Sprintf("API error for job %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *BackupJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathJobByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete backup job",
			fmt.Sprintf("API error for job %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *BackupJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewBackupJob returns a new veeam_backup_job resource instance.
func NewBackupJob() resource.Resource {
	return &BackupJob{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *BackupJob) buildSpec(data *BackupJobModel) *models.BackupJobSpec {
	spec := &models.BackupJobSpec{
		JobSpec: models.JobSpec{
			Name: data.Name.ValueString(),
			Type: models.EJobType(data.Type.ValueString()),
		},
		Description: data.Description.ValueString(),
	}

	if !data.IsHighPriority.IsNull() {
		spec.IsHighPriority = data.IsHighPriority.ValueBool()
	}

	// Storage
	spec.Storage = &models.BackupJobStorage{
		BackupRepositoryID: data.RepositoryID.ValueString(),
	}
	if !data.ProxyAutoSelect.IsNull() {
		spec.Storage.BackupProxies = &models.BackupProxiesSettings{
			AutoSelectEnabled: data.ProxyAutoSelect.ValueBool(),
		}
	}
	if !data.RetentionType.IsNull() {
		qty := 14 // safe default
		if !data.RetentionQuantity.IsNull() && !data.RetentionQuantity.IsUnknown() {
			qty = int(data.RetentionQuantity.ValueInt64())
		}
		spec.Storage.RetentionPolicy = &models.RetentionPolicySettings{
			Type:     models.ERetentionPolicyType(data.RetentionType.ValueString()),
			Quantity: qty,
		}
	}

	// Schedule
	if !data.ScheduleEnabled.IsNull() {
		spec.Schedule = &models.BackupSchedule{
			RunAutomatically: data.ScheduleEnabled.ValueBool(),
		}
		if !data.ScheduleTime.IsNull() {
			spec.Schedule.Daily = &models.ScheduleDaily{
				IsEnabled: true,
				LocalTime: data.ScheduleTime.ValueString(),
			}
			if !data.ScheduleKind.IsNull() {
				spec.Schedule.Daily.DailyKind = models.EDailyKinds(data.ScheduleKind.ValueString())
			}
		}
		if !data.RetryEnabled.IsNull() && data.RetryEnabled.ValueBool() {
			retryCount := 3
			awaitMin := 10
			if !data.RetryCount.IsNull() {
				retryCount = int(data.RetryCount.ValueInt64())
			}
			if !data.RetryAwaitMinutes.IsNull() {
				awaitMin = int(data.RetryAwaitMinutes.ValueInt64())
			}
			spec.Schedule.Retry = &models.ScheduleRetry{
				IsEnabled:    true,
				RetryCount:   retryCount,
				AwaitMinutes: awaitMin,
			}
		}
	}

	return spec
}

func (r *BackupJob) syncFromAPI(data *BackupJobModel, api *models.BackupJobModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(string(api.Type))
	data.IsDisabled = types.BoolValue(api.IsDisabled)
}

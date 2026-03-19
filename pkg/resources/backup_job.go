// Package resources implements Terraform resource types for the Veeam provider.
package resources

// ---------------------------------------------------------------------------
// veeam_backup_job — Terraform resource for Veeam Backup & Replication jobs
//
// Supported job types (controlled by the required "type" attribute):
//
//   VSphereBackup      Backup VMware vSphere VMs and containers.
//                      Requires: virtual_machines block with at least one include.
//
//   HyperVBackup       Backup Microsoft Hyper-V VMs and containers.
//                      Requires: virtual_machines block with at least one include.
//
//   WindowsAgentBackup Backup Windows machines managed via Veeam Agent.
//                      Requires: agent_computers block with at least one computer.
//
//   LinuxAgentBackup   Backup Linux machines managed via Veeam Agent.
//                      Requires: agent_computers block with at least one computer.
//
// Other job types (BackupCopy, VSphereReplica, FileBackup, etc.) are read-only
// via the veeam_backup_jobs data source. Full Terraform management of those
// types requires dedicated resource types due to their distinct API schema.
//
// CRUD behaviour — all operations are SYNCHRONOUS (no async polling required):
//   Create → POST /api/v1/jobs          (201 Created + JobModel body)
//   Read   → GET  /api/v1/jobs/{id}     (200 OK + JobModel body)
//   Update → PUT  /api/v1/jobs/{id}     (200 OK + JobModel body; full model required)
//   Delete → DELETE /api/v1/jobs/{id}   (204 No Content)
//
// Import: terraform import veeam_backup_job.name <job-uuid>
// ---------------------------------------------------------------------------

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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

// BackupJob implements the veeam_backup_job Terraform resource.
type BackupJob struct {
	client client.APIClient
}

// ---------------------------------------------------------------------------
// Terraform State Model Types
//
// These structs mirror the Terraform schema and are used for state serialisation.
// They are SEPARATE from the API models in internal/models — the buildSpec /
// buildAgentSpec / syncFromAPI helpers translate between them.
// ---------------------------------------------------------------------------

// BackupJobModel is the root Terraform state model for veeam_backup_job.
type BackupJobModel struct {
	// Core identity
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Type           types.String `tfsdk:"type"`
	IsHighPriority types.Bool   `tfsdk:"is_high_priority"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`

	// VM scope — required for VSphereBackup and HyperVBackup job types.
	VirtualMachines *VMBackupScope `tfsdk:"virtual_machines"`

	// Agent scope — required for WindowsAgentBackup and LinuxAgentBackup job types.
	AgentComputers []AgentComputerEntry `tfsdk:"agent_computers"`
	// AgentBackupMode selects the agent backup scope (EntireComputer, Volumes, FileLevel).
	AgentBackupMode types.String `tfsdk:"agent_backup_mode"`

	// Storage settings (optional; recommended for all job types).
	Storage *JobStorageSettings `tfsdk:"storage"`

	// Guest processing (optional; VSphereBackup / HyperVBackup only).
	GuestProcessing *JobGuestProcessing `tfsdk:"guest_processing"`

	// Schedule settings (optional).
	Schedule *JobScheduleSettings `tfsdk:"schedule"`
}

// VMBackupScope defines which VMs or containers the job protects.
type VMBackupScope struct {
	// Includes is the list of VMware / Hyper-V objects to back up. At least one required.
	Includes []VMIncludeEntry `tfsdk:"includes"`
	// ExcludeTemplates excludes all VM templates from the backup when true.
	ExcludeTemplates types.Bool `tfsdk:"exclude_templates"`
}

// VMIncludeEntry is a single VMware or Hyper-V inventory object to include.
type VMIncludeEntry struct {
	// Platform identifies the hypervisor platform: "VSphere" or "HyperV".
	Platform types.String `tfsdk:"platform"`
	// Type is the vSphere/Hyper-V object type (VirtualMachine, Folder, Datacenter, Cluster…).
	Type types.String `tfsdk:"type"`
	// HostName is the vCenter Server or ESXi hostname that owns this object.
	HostName types.String `tfsdk:"host_name"`
	// Name is the display name of the inventory object.
	Name types.String `tfsdk:"name"`
	// ObjectID is the vSphere MoRef ID (required for all objects except vCenter/ESXi hosts).
	ObjectID types.String `tfsdk:"object_id"`
}

// AgentComputerEntry is an agent-managed computer or protection group.
type AgentComputerEntry struct {
	// ID is the unique identifier of the agent-managed object.
	ID types.String `tfsdk:"id"`
	// Name is the display name of the computer or protection group.
	Name types.String `tfsdk:"name"`
	// Type is the object class (ProtectionGroup, WindowsComputer, LinuxComputer…).
	Type types.String `tfsdk:"type"`
	// ProtectionGroupID is the ID of the protection group that contains this object.
	ProtectionGroupID types.String `tfsdk:"protection_group_id"`
}

// JobStorageSettings maps to the BackupJobStorageModel / AgentBackupJobStorageModel.
type JobStorageSettings struct {
	// RepositoryID is the UUID of the target backup repository.
	RepositoryID types.String `tfsdk:"repository_id"`
	// ProxyAutoSelect enables automatic proxy selection (default: true).
	ProxyAutoSelect types.Bool `tfsdk:"proxy_auto_select"`
	// RetentionType selects whether retention is measured in RestorePoints or Days.
	RetentionType types.String `tfsdk:"retention_type"`
	// RetentionQuantity is the number of restore points or days to retain.
	RetentionQuantity types.Int64 `tfsdk:"retention_quantity"`
}

// JobGuestProcessing maps to BackupJobGuestProcessingModel.
type JobGuestProcessing struct {
	// AppAwareEnabled activates application-aware processing.
	AppAwareEnabled types.Bool `tfsdk:"app_aware_enabled"`
	// FSIndexingEnabled activates guest OS file indexing for search.
	FSIndexingEnabled types.Bool `tfsdk:"fs_indexing_enabled"`
	// InteractionProxyAutoSelect auto-selects the guest interaction proxy.
	InteractionProxyAutoSelect types.Bool `tfsdk:"interaction_proxy_auto_select"`
}

// JobScheduleSettings maps to BackupScheduleModel.
type JobScheduleSettings struct {
	// RunAutomatically activates automated scheduling.
	RunAutomatically types.Bool `tfsdk:"run_automatically"`

	// --- Daily schedule ---
	DailyEnabled   types.Bool   `tfsdk:"daily_enabled"`
	DailyLocalTime types.String `tfsdk:"daily_local_time"`
	// DailyKind selects which days: Everyday, Weekdays, or SelectedDays.
	DailyKind types.String `tfsdk:"daily_kind"`

	// --- Monthly schedule ---
	MonthlyEnabled    types.Bool   `tfsdk:"monthly_enabled"`
	MonthlyLocalTime  types.String `tfsdk:"monthly_local_time"`
	MonthlyDayOfMonth types.Int64  `tfsdk:"monthly_day_of_month"`

	// --- Periodic (interval) schedule ---
	PeriodicallyEnabled types.Bool `tfsdk:"periodically_enabled"`
	// PeriodicallyKind is the time unit: Hours, Minutes, Seconds, Days.
	PeriodicallyKind      types.String `tfsdk:"periodically_kind"`
	PeriodicallyFrequency types.Int64  `tfsdk:"periodically_frequency"`

	// --- After-job chaining ---
	AfterJobEnabled types.Bool `tfsdk:"after_job_enabled"`
	// AfterJobName is the DISPLAY NAME of the preceding job (not a UUID — per API spec).
	AfterJobName types.String `tfsdk:"after_job_name"`

	// --- Retry on failure ---
	RetryEnabled      types.Bool  `tfsdk:"retry_enabled"`
	RetryCount        types.Int64 `tfsdk:"retry_count"`
	RetryAwaitMinutes types.Int64 `tfsdk:"retry_await_minutes"`
}

// ---------------------------------------------------------------------------
// Metadata / Schema
// ---------------------------------------------------------------------------

func (r *BackupJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_job"
}

func (r *BackupJob) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Veeam Backup & Replication job.

Supported job types:
- **VSphereBackup** — VMware vSphere backup (requires ` + "`virtual_machines`" + ` block)
- **HyperVBackup** — Microsoft Hyper-V backup (requires ` + "`virtual_machines`" + ` block)
- **WindowsAgentBackup** — Veeam Agent for Windows (requires ` + "`agent_computers`" + ` block)
- **LinuxAgentBackup** — Veeam Agent for Linux (requires ` + "`agent_computers`" + ` block)`,

		Attributes: map[string]schema.Attribute{
			// -----------------------------------------------------------------
			// Core identity
			// -----------------------------------------------------------------
			"id": schema.StringAttribute{
				MarkdownDescription: "Job identifier assigned by the server (UUID).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique job name as it appears in the Veeam console.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readable description. Required by the Veeam API.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Job type discriminator. Allowed values: " +
					"`VSphereBackup`, `HyperVBackup`, `WindowsAgentBackup`, `LinuxAgentBackup`.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					// Changing job type requires destroy + recreate.
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_high_priority": schema.BoolAttribute{
				MarkdownDescription: "If `true`, the resource scheduler prioritises this job " +
					"over other similar jobs and allocates resources to it first.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"is_disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the job is disabled (read-only; use the Veeam " +
					"console to disable/enable jobs, or use the enable/disable API endpoints).",
				Computed: true,
			},

			// -----------------------------------------------------------------
			// Agent-specific flat fields
			// -----------------------------------------------------------------
			"agent_backup_mode": schema.StringAttribute{
				MarkdownDescription: "Agent backup scope. Required for `WindowsAgentBackup` " +
					"and `LinuxAgentBackup` job types. Allowed values: " +
					"`EntireComputer`, `Volumes`, `FileLevel`.",
				Optional: true,
				Computed: true,
			},

			// -----------------------------------------------------------------
			// VM scope (required for VSphereBackup / HyperVBackup)
			// -----------------------------------------------------------------
			"virtual_machines": schema.SingleNestedAttribute{
				MarkdownDescription: "Defines which VMs or containers the job protects. " +
					"Required for `VSphereBackup` and `HyperVBackup` job types.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"includes": schema.ListNestedAttribute{
						MarkdownDescription: "List of VMware vSphere or Hyper-V objects to " +
							"include in the backup. At least one entry is required.",
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"platform": schema.StringAttribute{
									MarkdownDescription: "Hypervisor platform. " +
										"Allowed values: `VSphere`, `HyperV`.",
									Required: true,
								},
								"type": schema.StringAttribute{
									MarkdownDescription: "Object type within the platform. " +
										"vSphere examples: `VirtualMachine`, `Folder`, " +
										"`Datacenter`, `Cluster`, `Host`, `ResourcePool`, " +
										"`VirtualApp`, `Tag`. " +
										"Leave empty to let the API infer the type.",
									Optional: true,
									Computed: true,
								},
								"host_name": schema.StringAttribute{
									MarkdownDescription: "FQDN or IP address of the vCenter " +
										"Server or ESXi host that owns this object.",
									Optional: true,
									Computed: true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Display name of the inventory object " +
										"(VM name, folder name, datacenter name, etc.).",
									Required: true,
								},
								"object_id": schema.StringAttribute{
									MarkdownDescription: "vSphere MoRef ID (e.g. `vm-101`, " +
										"`domain-c12`). Required for all objects except " +
										"vCenter Servers and standalone ESXi hosts.",
									Optional: true,
									Computed: true,
								},
							},
						},
					},
					"exclude_templates": schema.BoolAttribute{
						MarkdownDescription: "If `true`, all VM templates are excluded from " +
							"the backup scope.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
				},
			},

			// -----------------------------------------------------------------
			// Agent computer scope (required for WindowsAgentBackup / LinuxAgentBackup)
			// -----------------------------------------------------------------
			"agent_computers": schema.ListNestedAttribute{
				MarkdownDescription: "List of agent-managed computers or protection groups " +
					"to include in the backup. Required for `WindowsAgentBackup` and " +
					"`LinuxAgentBackup` job types. Obtain object IDs from the " +
					"`veeam_protected_computers` data source or the Veeam console.",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "UUID of the agent-managed object.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Display name of the computer or protection group.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Object class. Common values: " +
								"`ProtectionGroup`, `WindowsComputer`, `LinuxComputer`, " +
								"`WindowsCluster`, `Domain`, `OrganizationUnit`.",
							Required: true,
						},
						"protection_group_id": schema.StringAttribute{
							MarkdownDescription: "UUID of the protection group that contains " +
								"this object. Obtain from the `veeam_protection_group` resource " +
								"or data source.",
							Required: true,
						},
					},
				},
			},

			// -----------------------------------------------------------------
			// Storage settings
			// -----------------------------------------------------------------
			"storage": schema.SingleNestedAttribute{
				MarkdownDescription: "Backup storage configuration. When omitted, Veeam " +
					"applies server defaults. **Strongly recommended** to set explicitly.",
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"repository_id": schema.StringAttribute{
						MarkdownDescription: "UUID of the target backup repository. " +
							"Obtain from the `veeam_backup_repositories` data source.",
						Optional: true,
						Computed: true,
					},
					"proxy_auto_select": schema.BoolAttribute{
						MarkdownDescription: "If `true` (default), Veeam automatically " +
							"selects the most suitable backup proxy.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"retention_type": schema.StringAttribute{
						MarkdownDescription: "Retention policy type. " +
							"Allowed values: `RestorePoints`, `Days`.",
						Optional: true,
						Computed: true,
					},
					"retention_quantity": schema.Int64Attribute{
						MarkdownDescription: "Number of restore points or days to retain " +
							"(must be greater than zero).",
						Optional: true,
						Computed: true,
					},
				},
			},

			// -----------------------------------------------------------------
			// Guest processing
			// -----------------------------------------------------------------
			"guest_processing": schema.SingleNestedAttribute{
				MarkdownDescription: "Application-aware processing and guest OS file indexing. " +
					"Applies to `VSphereBackup` and `HyperVBackup` job types only.",
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"app_aware_enabled": schema.BoolAttribute{
						MarkdownDescription: "If `true`, application-aware processing is " +
							"enabled. Requires VMware Tools / Hyper-V Integration Services " +
							"to be installed in each guest VM.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"fs_indexing_enabled": schema.BoolAttribute{
						MarkdownDescription: "If `true`, guest OS file indexing is enabled " +
							"to allow file-level search inside backup archives.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"interaction_proxy_auto_select": schema.BoolAttribute{
						MarkdownDescription: "If `true` (default), Veeam automatically " +
							"selects the guest interaction proxy.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
				},
			},

			// -----------------------------------------------------------------
			// Schedule
			// -----------------------------------------------------------------
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "Job scheduling configuration. When omitted, the job " +
					"must be started manually.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"run_automatically": schema.BoolAttribute{
						MarkdownDescription: "If `true`, the job runs on the configured " +
							"schedule automatically.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					// Daily
					"daily_enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable daily schedule.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"daily_local_time": schema.StringAttribute{
						MarkdownDescription: "Daily start time in `HH:MM` format (server local time).",
						Optional:            true,
						Computed:            true,
					},
					"daily_kind": schema.StringAttribute{
						MarkdownDescription: "Which days to run. " +
							"Allowed values: `Everyday`, `Weekdays`, `SelectedDays`.",
						Optional: true,
						Computed: true,
					},
					// Monthly
					"monthly_enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable monthly schedule.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"monthly_local_time": schema.StringAttribute{
						MarkdownDescription: "Monthly start time in `HH:MM` format.",
						Optional:            true,
						Computed:            true,
					},
					"monthly_day_of_month": schema.Int64Attribute{
						MarkdownDescription: "Day of the month (1–28) on which the job runs.",
						Optional:            true,
						Computed:            true,
					},
					// Periodic (interval)
					"periodically_enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable periodic (interval-based) schedule.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"periodically_kind": schema.StringAttribute{
						MarkdownDescription: "Time unit for the interval. " +
							"Allowed values: `Hours`, `Minutes`, `Seconds`, `Days`.",
						Optional: true,
						Computed: true,
					},
					"periodically_frequency": schema.Int64Attribute{
						MarkdownDescription: "Number of time units between runs.",
						Optional:            true,
						Computed:            true,
					},
					// After-job chaining
					"after_job_enabled": schema.BoolAttribute{
						MarkdownDescription: "If `true`, this job starts automatically after " +
							"another job completes.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"after_job_name": schema.StringAttribute{
						MarkdownDescription: "Display **name** of the preceding job. " +
							"The Veeam API v1.3 identifies chained jobs by name, not UUID.",
						Optional: true,
						Computed: true,
					},
					// Retry
					"retry_enabled": schema.BoolAttribute{
						MarkdownDescription: "If `true`, the job is retried automatically " +
							"when it fails.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"retry_count": schema.Int64Attribute{
						MarkdownDescription: "Number of retry attempts (must be > 0).",
						Optional:            true,
						Computed:            true,
					},
					"retry_await_minutes": schema.Int64Attribute{
						MarkdownDescription: "Wait time between retries in minutes (must be > 0).",
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Configure
// ---------------------------------------------------------------------------

func (r *BackupJob) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected client.APIClient from provider configuration.",
		)
		return
	}
	r.client = c
}

// ---------------------------------------------------------------------------
// CRUD — Create
// ---------------------------------------------------------------------------

func (r *BackupJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobType := models.EJobType(data.Type.ValueString())

	switch jobType {
	case models.JobTypeVSphereBackup, models.JobTypeHyperVBackup:
		if data.VirtualMachines == nil || len(data.VirtualMachines.Includes) == 0 {
			resp.Diagnostics.AddError(
				"Missing required virtual_machines block",
				fmt.Sprintf("Job type '%s' requires a virtual_machines block with at least one include entry.", jobType),
			)
			return
		}

		spec := r.buildVMJobSpec(&data)
		var result models.BackupJobModel
		if err := r.client.PostJSON(ctx, client.PathJobs, spec, &result); err != nil {
			resp.Diagnostics.AddError("Failed to create backup job",
				fmt.Sprintf("POST %s: %s", client.PathJobs, err))
			return
		}
		data.ID = types.StringValue(result.ID)
		r.syncVMJobFromAPI(&data, &result)

	case models.JobTypeWindowsAgentBackup, models.JobTypeLinuxAgentBackup:
		if len(data.AgentComputers) == 0 {
			resp.Diagnostics.AddError(
				"Missing required agent_computers",
				fmt.Sprintf("Job type '%s' requires at least one entry in agent_computers.", jobType),
			)
			return
		}
		if data.AgentBackupMode.IsNull() || data.AgentBackupMode.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing required agent_backup_mode",
				fmt.Sprintf("Job type '%s' requires agent_backup_mode to be set.", jobType),
			)
			return
		}

		spec := r.buildAgentJobSpec(&data)
		var result models.BackupJobModel
		if err := r.client.PostJSON(ctx, client.PathJobs, spec, &result); err != nil {
			resp.Diagnostics.AddError("Failed to create agent backup job",
				fmt.Sprintf("POST %s: %s", client.PathJobs, err))
			return
		}
		data.ID = types.StringValue(result.ID)
		r.syncAgentJobFromAPI(&data, &result)

	default:
		resp.Diagnostics.AddError(
			"Unsupported job type",
			fmt.Sprintf("Job type '%s' is not supported by this resource. Use the Veeam console "+
				"to manage %s jobs, or check if a dedicated Terraform resource exists.", jobType, jobType),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// ---------------------------------------------------------------------------
// CRUD — Read
// ---------------------------------------------------------------------------

func (r *BackupJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathJobByID, data.ID.ValueString())

	jobType := models.EJobType(data.Type.ValueString())

	switch jobType {
	case models.JobTypeVSphereBackup, models.JobTypeHyperVBackup:
		var result models.BackupJobModel
		if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
			resp.Diagnostics.AddError("Failed to read backup job",
				fmt.Sprintf("GET %s: %s", endpoint, err))
			return
		}
		r.syncVMJobFromAPI(&data, &result)

	case models.JobTypeWindowsAgentBackup, models.JobTypeLinuxAgentBackup:
		var result models.BackupJobModel
		if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
			resp.Diagnostics.AddError("Failed to read agent backup job",
				fmt.Sprintf("GET %s: %s", endpoint, err))
			return
		}
		r.syncAgentJobFromAPI(&data, &result)

	default:
		// For unknown/unsupported types encountered during import, fall back to
		// base model which at minimum syncs id / name / type / isDisabled.
		var result models.BackupJobModel
		if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
			resp.Diagnostics.AddError("Failed to read backup job",
				fmt.Sprintf("GET %s: %s", endpoint, err))
			return
		}
		data.Name = types.StringValue(result.Name)
		data.Type = types.StringValue(string(result.Type))
		data.IsDisabled = types.BoolValue(result.IsDisabled)
		if result.Description != "" {
			data.Description = types.StringValue(result.Description)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// ---------------------------------------------------------------------------
// CRUD — Update
// ---------------------------------------------------------------------------

func (r *BackupJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the ID from current state.
	var state BackupJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	endpoint := fmt.Sprintf(client.PathJobByID, data.ID.ValueString())
	jobType := models.EJobType(data.Type.ValueString())

	// The Veeam API PUT /api/v1/jobs/{id} expects the full JobModel (with id / type /
	// isDisabled) — not just the creation spec.  We therefore build the full model
	// object for the update payload.
	switch jobType {
	case models.JobTypeVSphereBackup, models.JobTypeHyperVBackup:
		if data.VirtualMachines == nil || len(data.VirtualMachines.Includes) == 0 {
			resp.Diagnostics.AddError(
				"Missing required virtual_machines block",
				fmt.Sprintf("Job type '%s' requires a virtual_machines block with at least one include entry.", jobType),
			)
			return
		}
		payload := r.buildVMJobModel(&data, state.IsDisabled.ValueBool())
		var result models.BackupJobModel
		if err := r.client.PutJSON(ctx, endpoint, payload, &result); err != nil {
			resp.Diagnostics.AddError("Failed to update backup job",
				fmt.Sprintf("PUT %s: %s", endpoint, err))
			return
		}
		if result.ID != "" {
			r.syncVMJobFromAPI(&data, &result)
		}

	case models.JobTypeWindowsAgentBackup, models.JobTypeLinuxAgentBackup:
		if len(data.AgentComputers) == 0 {
			resp.Diagnostics.AddError(
				"Missing required agent_computers",
				fmt.Sprintf("Job type '%s' requires at least one entry in agent_computers.", jobType),
			)
			return
		}
		payload := r.buildAgentJobModel(&data, state.IsDisabled.ValueBool())
		var result models.BackupJobModel
		if err := r.client.PutJSON(ctx, endpoint, payload, &result); err != nil {
			resp.Diagnostics.AddError("Failed to update agent backup job",
				fmt.Sprintf("PUT %s: %s", endpoint, err))
			return
		}
		if result.ID != "" {
			r.syncAgentJobFromAPI(&data, &result)
		}

	default:
		resp.Diagnostics.AddError(
			"Unsupported job type for update",
			fmt.Sprintf("Cannot update job type '%s' via this resource.", jobType),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// ---------------------------------------------------------------------------
// CRUD — Delete
// ---------------------------------------------------------------------------

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
			fmt.Sprintf("DELETE %s (job %s): %s", endpoint, data.ID.ValueString(), err),
		)
	}
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func (r *BackupJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// NewBackupJob returns a new veeam_backup_job resource instance.
// ---------------------------------------------------------------------------

func NewBackupJob() resource.Resource {
	return &BackupJob{}
}

// ---------------------------------------------------------------------------
// Build helpers — Terraform state → API request body
// ---------------------------------------------------------------------------

// buildVMJobSpec converts Terraform plan state into a BackupJobSpec for POST (create).
func (r *BackupJob) buildVMJobSpec(data *BackupJobModel) *models.BackupJobSpec {
	spec := &models.BackupJobSpec{
		JobSpec: models.JobSpec{
			Name: data.Name.ValueString(),
			Type: models.EJobType(data.Type.ValueString()),
		},
		Description:     data.Description.ValueString(),
		IsHighPriority:  data.IsHighPriority.ValueBool(),
		VirtualMachines: r.buildVirtualMachinesSpec(data.VirtualMachines),
	}

	if data.Storage != nil {
		spec.Storage = r.buildStorageModel(data.Storage)
	}

	if data.GuestProcessing != nil {
		spec.GuestProcessing = r.buildGuestProcessingModel(data.GuestProcessing)
	}

	if data.Schedule != nil {
		spec.Schedule = r.buildScheduleModel(data.Schedule)
	}

	return spec
}

// buildVMJobModel converts Terraform plan state into a full BackupJobModel for PUT (update).
// The Veeam API PUT endpoint expects the complete JobModel (including id / isDisabled).
func (r *BackupJob) buildVMJobModel(data *BackupJobModel, isDisabled bool) *models.BackupJobModel {
	m := &models.BackupJobModel{
		JobModel: models.JobModel{
			ID:         data.ID.ValueString(),
			Name:       data.Name.ValueString(),
			Type:       models.EJobType(data.Type.ValueString()),
			IsDisabled: isDisabled,
		},
		Description:     data.Description.ValueString(),
		IsHighPriority:  data.IsHighPriority.ValueBool(),
		VirtualMachines: r.buildVirtualMachinesModel(data.VirtualMachines),
	}

	if data.Storage != nil {
		m.Storage = r.buildStorageModel(data.Storage)
	}

	if data.GuestProcessing != nil {
		m.GuestProcessing = r.buildGuestProcessingModel(data.GuestProcessing)
	}

	if data.Schedule != nil {
		m.Schedule = r.buildScheduleModel(data.Schedule)
	}

	return m
}

// buildAgentJobSpec builds a WindowsAgentBackupJobSpec or LinuxAgentBackupJobSpec for POST.
// Both have an identical JSON structure for the fields we manage, so we use a common map.
func (r *BackupJob) buildAgentJobSpec(data *BackupJobModel) map[string]any {
	spec := map[string]any{
		"name":        data.Name.ValueString(),
		"type":        data.Type.ValueString(),
		"description": data.Description.ValueString(),
		"backupMode":  data.AgentBackupMode.ValueString(),
		"computers":   r.buildAgentComputerList(data.AgentComputers),
	}

	if !data.IsHighPriority.IsNull() {
		spec["isHighPriority"] = data.IsHighPriority.ValueBool()
	}

	if data.Storage != nil {
		spec["storage"] = r.buildAgentStorageModel(data.Storage)
	}

	if data.Schedule != nil {
		spec["schedule"] = r.buildScheduleModel(data.Schedule)
	}

	return spec
}

// buildAgentJobModel builds a full model map for PUT (update) of agent backup jobs.
func (r *BackupJob) buildAgentJobModel(data *BackupJobModel, isDisabled bool) map[string]any {
	m := map[string]any{
		"id":          data.ID.ValueString(),
		"name":        data.Name.ValueString(),
		"type":        data.Type.ValueString(),
		"isDisabled":  isDisabled,
		"description": data.Description.ValueString(),
		"backupMode":  data.AgentBackupMode.ValueString(),
		"computers":   r.buildAgentComputerList(data.AgentComputers),
	}

	if !data.IsHighPriority.IsNull() {
		m["isHighPriority"] = data.IsHighPriority.ValueBool()
	}

	if data.Storage != nil {
		m["storage"] = r.buildAgentStorageModel(data.Storage)
	}

	if data.Schedule != nil {
		m["schedule"] = r.buildScheduleModel(data.Schedule)
	}

	return m
}

// ---------------------------------------------------------------------------
// Sub-section builders
// ---------------------------------------------------------------------------

func (r *BackupJob) buildVirtualMachinesSpec(scope *VMBackupScope) *models.BackupJobVirtualMachinesSpec {
	if scope == nil {
		return nil
	}

	includes := make([]models.VmwareObjectSpec, 0, len(scope.Includes))
	for _, entry := range scope.Includes {
		platform := entry.Platform.ValueString()
		if platform == "" {
			platform = string(models.InventoryPlatformVSphere)
		}
		obj := models.VmwareObjectSpec{
			Platform: platform,
			Name:     entry.Name.ValueString(),
			HostName: entry.HostName.ValueString(),
			Type:     models.EVmwareInventoryType(entry.Type.ValueString()),
			ObjectID: entry.ObjectID.ValueString(),
		}
		includes = append(includes, obj)
	}

	spec := &models.BackupJobVirtualMachinesSpec{Includes: includes}

	if !scope.ExcludeTemplates.IsNull() && scope.ExcludeTemplates.ValueBool() {
		spec.Excludes = &models.BackupJobExclusionsSpec{
			Templates: &models.BackupJobExclusionsTemplates{IsEnabled: true},
		}
	}

	return spec
}

func (r *BackupJob) buildVirtualMachinesModel(scope *VMBackupScope) *models.BackupJobVirtualMachinesModel {
	if scope == nil {
		return nil
	}

	includes := make([]models.VmwareObjectSpec, 0, len(scope.Includes))
	for _, entry := range scope.Includes {
		platform := entry.Platform.ValueString()
		if platform == "" {
			platform = string(models.InventoryPlatformVSphere)
		}
		obj := models.VmwareObjectSpec{
			Platform: platform,
			Name:     entry.Name.ValueString(),
			HostName: entry.HostName.ValueString(),
			Type:     models.EVmwareInventoryType(entry.Type.ValueString()),
			ObjectID: entry.ObjectID.ValueString(),
		}
		includes = append(includes, obj)
	}

	m := &models.BackupJobVirtualMachinesModel{Includes: includes}

	if !scope.ExcludeTemplates.IsNull() && scope.ExcludeTemplates.ValueBool() {
		m.Excludes = &models.BackupJobExclusions{
			Templates: &models.BackupJobExclusionsTemplates{IsEnabled: true},
		}
	}

	return m
}

func (r *BackupJob) buildAgentComputerList(computers []AgentComputerEntry) []models.AgentObjectSpec {
	result := make([]models.AgentObjectSpec, 0, len(computers))
	for _, c := range computers {
		result = append(result, models.AgentObjectSpec{
			Platform:          string(models.InventoryPlatformAgent),
			ID:                c.ID.ValueString(),
			Name:              c.Name.ValueString(),
			Type:              models.EAgentInventoryObjectType(c.Type.ValueString()),
			ProtectionGroupID: c.ProtectionGroupID.ValueString(),
		})
	}
	return result
}

func (r *BackupJob) buildStorageModel(s *JobStorageSettings) *models.BackupJobStorageModel {
	if s == nil {
		return nil
	}

	m := &models.BackupJobStorageModel{
		BackupRepositoryID: s.RepositoryID.ValueString(),
		BackupProxies: &models.BackupProxiesSettingsModel{
			AutoSelectEnabled: !s.ProxyAutoSelect.IsNull() && s.ProxyAutoSelect.ValueBool(),
		},
	}

	if !s.RetentionType.IsNull() {
		qty := 14 // safe default — matches Veeam console default
		if !s.RetentionQuantity.IsNull() && !s.RetentionQuantity.IsUnknown() {
			qty = int(s.RetentionQuantity.ValueInt64())
		}
		m.RetentionPolicy = &models.BackupJobRetentionPolicySettings{
			Type:     models.ERetentionPolicyType(s.RetentionType.ValueString()),
			Quantity: qty,
		}
	}

	return m
}

func (r *BackupJob) buildAgentStorageModel(s *JobStorageSettings) *models.AgentBackupJobStorageModel {
	if s == nil {
		return nil
	}

	m := &models.AgentBackupJobStorageModel{
		BackupRepositoryID: s.RepositoryID.ValueString(),
	}

	if !s.RetentionType.IsNull() {
		qty := 14
		if !s.RetentionQuantity.IsNull() && !s.RetentionQuantity.IsUnknown() {
			qty = int(s.RetentionQuantity.ValueInt64())
		}
		m.RetentionPolicy = &models.BackupJobRetentionPolicySettings{
			Type:     models.ERetentionPolicyType(s.RetentionType.ValueString()),
			Quantity: qty,
		}
	}

	return m
}

func (r *BackupJob) buildGuestProcessingModel(gp *JobGuestProcessing) *models.BackupJobGuestProcessingModel {
	if gp == nil {
		return nil
	}

	m := &models.BackupJobGuestProcessingModel{
		AppAwareProcessing: &models.BackupApplicationAwareProcessingModel{
			IsEnabled: gp.AppAwareEnabled.ValueBool(),
		},
		GuestFSIndexing: &models.GuestFileSystemIndexingModel{
			IsEnabled: gp.FSIndexingEnabled.ValueBool(),
		},
	}

	// Include proxy settings only when meaningful to avoid sending an empty object.
	if !gp.InteractionProxyAutoSelect.IsNull() {
		m.GuestInteractionProxies = &models.GuestInteractionProxiesSettingsModel{
			AutoSelectEnabled: gp.InteractionProxyAutoSelect.ValueBool(),
		}
	}

	return m
}

func (r *BackupJob) buildScheduleModel(s *JobScheduleSettings) *models.BackupScheduleModel {
	if s == nil {
		return nil
	}

	m := &models.BackupScheduleModel{
		RunAutomatically: s.RunAutomatically.ValueBool(),
	}

	if !s.DailyEnabled.IsNull() && s.DailyEnabled.ValueBool() {
		m.Daily = &models.ScheduleDailyModel{
			IsEnabled: true,
			LocalTime: s.DailyLocalTime.ValueString(),
			DailyKind: models.EDailyKinds(s.DailyKind.ValueString()),
		}
	}

	if !s.MonthlyEnabled.IsNull() && s.MonthlyEnabled.ValueBool() {
		monthly := &models.ScheduleMonthlyModel{IsEnabled: true}
		if !s.MonthlyLocalTime.IsNull() {
			monthly.LocalTime = s.MonthlyLocalTime.ValueString()
		}
		if !s.MonthlyDayOfMonth.IsNull() {
			monthly.DayOfMonth = int(s.MonthlyDayOfMonth.ValueInt64())
		}
		m.Monthly = monthly
	}

	if !s.PeriodicallyEnabled.IsNull() && s.PeriodicallyEnabled.ValueBool() {
		m.Periodically = &models.SchedulePeriodicallyModel{
			IsEnabled:        true,
			PeriodicallyKind: models.EPeriodicallyKinds(s.PeriodicallyKind.ValueString()),
			Frequency:        int(s.PeriodicallyFrequency.ValueInt64()),
		}
	}

	if !s.AfterJobEnabled.IsNull() && s.AfterJobEnabled.ValueBool() {
		m.AfterThisJob = &models.ScheduleAfterThisJobModel{
			IsEnabled: true,
			// NOTE: The API uses job NAME here, not UUID (per v1.3-rev1 spec).
			JobName: s.AfterJobName.ValueString(),
		}
	}

	if !s.RetryEnabled.IsNull() && s.RetryEnabled.ValueBool() {
		retryCount := 3
		awaitMin := 10
		if !s.RetryCount.IsNull() && !s.RetryCount.IsUnknown() {
			retryCount = int(s.RetryCount.ValueInt64())
		}
		if !s.RetryAwaitMinutes.IsNull() && !s.RetryAwaitMinutes.IsUnknown() {
			awaitMin = int(s.RetryAwaitMinutes.ValueInt64())
		}
		m.Retry = &models.ScheduleRetryModel{
			IsEnabled:    true,
			RetryCount:   retryCount,
			AwaitMinutes: awaitMin,
		}
	}

	return m
}

// ---------------------------------------------------------------------------
// Sync helpers — API response → Terraform state
// ---------------------------------------------------------------------------

// syncVMJobFromAPI merges a BackupJobModel API response into Terraform state.
// Sensitive fields (passwords) and fields not returned by the API are preserved
// from the existing state (callers pass &data which already holds prior values).
func (r *BackupJob) syncVMJobFromAPI(data *BackupJobModel, api *models.BackupJobModel) {
	data.Name = types.StringValue(api.Name)
	data.Type = types.StringValue(string(api.Type))
	data.IsDisabled = types.BoolValue(api.IsDisabled)
	data.IsHighPriority = types.BoolValue(api.IsHighPriority)

	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}

	// Sync virtual_machines from API response when present.
	if api.VirtualMachines != nil {
		scope := &VMBackupScope{}
		includes := make([]VMIncludeEntry, 0, len(api.VirtualMachines.Includes))
		for _, vm := range api.VirtualMachines.Includes {
			includes = append(includes, VMIncludeEntry{
				Platform: types.StringValue(vm.Platform),
				Type:     types.StringValue(string(vm.Type)),
				HostName: types.StringValue(vm.HostName),
				Name:     types.StringValue(vm.Name),
				ObjectID: types.StringValue(vm.ObjectID),
			})
		}
		scope.Includes = includes

		if api.VirtualMachines.Excludes != nil &&
			api.VirtualMachines.Excludes.Templates != nil {
			scope.ExcludeTemplates = types.BoolValue(api.VirtualMachines.Excludes.Templates.IsEnabled)
		} else {
			scope.ExcludeTemplates = types.BoolValue(false)
		}

		data.VirtualMachines = scope
	}

	// Sync storage settings.
	if api.Storage != nil {
		s := &JobStorageSettings{
			RepositoryID:    types.StringValue(api.Storage.BackupRepositoryID),
			ProxyAutoSelect: types.BoolValue(false),
		}
		if api.Storage.BackupProxies != nil {
			s.ProxyAutoSelect = types.BoolValue(api.Storage.BackupProxies.AutoSelectEnabled)
		}
		if api.Storage.RetentionPolicy != nil {
			s.RetentionType = types.StringValue(string(api.Storage.RetentionPolicy.Type))
			s.RetentionQuantity = types.Int64Value(int64(api.Storage.RetentionPolicy.Quantity))
		}
		data.Storage = s
	}

	// Sync guest processing.
	if api.GuestProcessing != nil {
		gp := &JobGuestProcessing{
			AppAwareEnabled:            types.BoolValue(false),
			FSIndexingEnabled:          types.BoolValue(false),
			InteractionProxyAutoSelect: types.BoolValue(true),
		}
		if api.GuestProcessing.AppAwareProcessing != nil {
			gp.AppAwareEnabled = types.BoolValue(api.GuestProcessing.AppAwareProcessing.IsEnabled)
		}
		if api.GuestProcessing.GuestFSIndexing != nil {
			gp.FSIndexingEnabled = types.BoolValue(api.GuestProcessing.GuestFSIndexing.IsEnabled)
		}
		if api.GuestProcessing.GuestInteractionProxies != nil {
			gp.InteractionProxyAutoSelect = types.BoolValue(
				api.GuestProcessing.GuestInteractionProxies.AutoSelectEnabled)
		}
		data.GuestProcessing = gp
	}

	// Sync schedule.
	if api.Schedule != nil {
		data.Schedule = r.syncScheduleFromAPI(data.Schedule, api.Schedule)
	}
}

// syncAgentJobFromAPI merges a BackupJobModel (agent type) API response into state.
// Agent job models use a subset of BackupJobModel fields on the response.
func (r *BackupJob) syncAgentJobFromAPI(data *BackupJobModel, api *models.BackupJobModel) {
	data.Name = types.StringValue(api.Name)
	data.Type = types.StringValue(string(api.Type))
	data.IsDisabled = types.BoolValue(api.IsDisabled)
	data.IsHighPriority = types.BoolValue(api.IsHighPriority)

	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}

	// Agent-specific fields are not in BackupJobModel directly — the API returns
	// them in the top-level body.  The generic map-based approach used in
	// buildAgentJobSpec / buildAgentJobModel means those fields are not parsed
	// back automatically.  We preserve whatever the user configured in the plan
	// for agent_computers / agent_backup_mode, and only update the computed fields.
	// TODO: Switch to a typed WindowsAgentBackupJobModel / LinuxAgentBackupJobModel
	//       response parser when stronger validation of agent job round-trips is needed.
}

// syncScheduleFromAPI updates a JobScheduleSettings from an API BackupScheduleModel.
// Preserves existing state values for fields not present in the API response.
func (r *BackupJob) syncScheduleFromAPI(existing *JobScheduleSettings, api *models.BackupScheduleModel) *JobScheduleSettings {
	s := &JobScheduleSettings{}
	if existing != nil {
		// Start from existing to preserve user-set values not returned by the API.
		*s = *existing
	}

	s.RunAutomatically = types.BoolValue(api.RunAutomatically)

	if api.Daily != nil {
		s.DailyEnabled = types.BoolValue(api.Daily.IsEnabled)
		s.DailyLocalTime = types.StringValue(api.Daily.LocalTime)
		s.DailyKind = types.StringValue(string(api.Daily.DailyKind))
	} else {
		s.DailyEnabled = types.BoolValue(false)
	}

	if api.Monthly != nil {
		s.MonthlyEnabled = types.BoolValue(api.Monthly.IsEnabled)
		s.MonthlyLocalTime = types.StringValue(api.Monthly.LocalTime)
		if api.Monthly.DayOfMonth > 0 {
			s.MonthlyDayOfMonth = types.Int64Value(int64(api.Monthly.DayOfMonth))
		}
	} else {
		s.MonthlyEnabled = types.BoolValue(false)
	}

	if api.Periodically != nil {
		s.PeriodicallyEnabled = types.BoolValue(api.Periodically.IsEnabled)
		s.PeriodicallyKind = types.StringValue(string(api.Periodically.PeriodicallyKind))
		s.PeriodicallyFrequency = types.Int64Value(int64(api.Periodically.Frequency))
	} else {
		s.PeriodicallyEnabled = types.BoolValue(false)
	}

	if api.AfterThisJob != nil {
		s.AfterJobEnabled = types.BoolValue(api.AfterThisJob.IsEnabled)
		s.AfterJobName = types.StringValue(api.AfterThisJob.JobName)
	} else {
		s.AfterJobEnabled = types.BoolValue(false)
	}

	if api.Retry != nil {
		s.RetryEnabled = types.BoolValue(api.Retry.IsEnabled)
		s.RetryCount = types.Int64Value(int64(api.Retry.RetryCount))
		s.RetryAwaitMinutes = types.Int64Value(int64(api.Retry.AwaitMinutes))
	} else {
		s.RetryEnabled = types.BoolValue(false)
	}

	return s
}

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

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	// IncludeUsbDrives includes periodically connected USB drives. WindowsAgentBackup only.
	IncludeUsbDrives types.Bool `tfsdk:"include_usb_drives"`
	// AgentType selects the protected computer type. WindowsAgentBackup only.
	AgentType types.String `tfsdk:"agent_type"`
	// UseSnapshotlessFileLevelBackup creates a crash-consistent backup without a snapshot.
	// LinuxAgentBackup only; applies when agent_backup_mode = "FileLevel".
	UseSnapshotlessFileLevelBackup types.Bool `tfsdk:"use_snapshotless_file_level_backup"`

	// VolumesScope configures which volumes to back up.
	// Only used for agent jobs when agent_backup_mode = "Volumes".
	VolumesScope *AgentVolumesScope `tfsdk:"volumes_scope"`

	// FilesScope configures included/excluded paths for file-level backups.
	// Only used for agent jobs when agent_backup_mode = "FileLevel".
	FilesScope *AgentFilesScope `tfsdk:"files_scope"`

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

// AgentVolumesScope maps to AgentBackupJobVolumesModel.
type AgentVolumesScope struct {
	// AllVolumes backs up all local volumes when true.
	AllVolumes types.Bool `tfsdk:"all_volumes"`
	// VolumeNames lists specific volumes to back up (used when all_volumes = false).
	// Examples: ["C:", "D:", "/data"].
	VolumeNames types.List `tfsdk:"volume_names"`
}

// AgentFilesScope maps to AgentBackupJobFilesModel.
type AgentFilesScope struct {
	// IncludedFolders lists directory paths to include in the backup.
	IncludedFolders types.List `tfsdk:"included_folders"`
	// ExcludedFolders lists directory paths to exclude from the backup.
	ExcludedFolders types.List `tfsdk:"excluded_folders"`
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
	// GFSPolicy configures Grandfather-Father-Son long-term archival retention.
	GFSPolicy *JobGFSPolicy `tfsdk:"gfs_policy"`
}

// JobGFSPolicy maps to GFSPolicySettingsModel.
type JobGFSPolicy struct {
	// IsEnabled activates the GFS long-term retention policy.
	IsEnabled types.Bool `tfsdk:"is_enabled"`
	// WeeklyEnabled activates weekly GFS archival.
	WeeklyEnabled types.Bool `tfsdk:"weekly_enabled"`
	// WeeklyKeepFor is the number of weeks to keep weekly full backups (1–9999).
	WeeklyKeepFor types.Int64 `tfsdk:"weekly_keep_for"`
	// WeeklyDesiredTime is the day of the week on which the weekly full is created.
	WeeklyDesiredTime types.String `tfsdk:"weekly_desired_time"`
	// MonthlyEnabled activates monthly GFS archival.
	MonthlyEnabled types.Bool `tfsdk:"monthly_enabled"`
	// MonthlyKeepFor is the number of months to keep monthly full backups (1–999).
	MonthlyKeepFor types.Int64 `tfsdk:"monthly_keep_for"`
	// MonthlyDesiredTime is the week-of-month on which the monthly full is created.
	MonthlyDesiredTime types.String `tfsdk:"monthly_desired_time"`
	// YearlyEnabled activates yearly GFS archival.
	YearlyEnabled types.Bool `tfsdk:"yearly_enabled"`
	// YearlyKeepFor is the number of years to keep yearly full backups (1–999).
	YearlyKeepFor types.Int64 `tfsdk:"yearly_keep_for"`
	// YearlyDesiredTime is the month in which the yearly full is created.
	YearlyDesiredTime types.String `tfsdk:"yearly_desired_time"`
}

// JobGuestProcessing maps to BackupJobGuestProcessingModel.
type JobGuestProcessing struct {
	// AppAwareEnabled activates application-aware processing.
	AppAwareEnabled types.Bool `tfsdk:"app_aware_enabled"`
	// FSIndexingEnabled activates guest OS file indexing for search.
	FSIndexingEnabled types.Bool `tfsdk:"fs_indexing_enabled"`
	// InteractionProxyAutoSelect auto-selects the guest interaction proxy.
	InteractionProxyAutoSelect types.Bool `tfsdk:"interaction_proxy_auto_select"`
	// GuestCredentials specifies the credentials used for guest OS interaction.
	// Applies to VSphereBackup and HyperVBackup job types only.
	GuestCredentials *JobGuestCredentials `tfsdk:"guest_credentials"`
}

// JobGuestCredentials maps to GuestOsCredentialsModel.
type JobGuestCredentials struct {
	// CredentialsID is the UUID of the credentials record.
	// Obtain from the veeam_credentials data source or veeam_credential resource.
	CredentialsID types.String `tfsdk:"credentials_id"`
}

// JobScheduleSettings maps to BackupScheduleModel.
type JobScheduleSettings struct {
	// RunAutomatically activates automated scheduling.
	RunAutomatically types.Bool `tfsdk:"run_automatically"`

	// --- Daily schedule ---
	DailyEnabled   types.Bool   `tfsdk:"daily_enabled"`
	DailyLocalTime types.String `tfsdk:"daily_local_time"`
	// DailyKind selects which days: Everyday, WeekDays, or SelectedDays.
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
					"and `LinuxAgentBackup` job types. Optional, Computed. " +
					"Allowed values: `EntireComputer`, `Volumes`, `FileLevel`.",
				Optional: true,
				Computed: true,
			},
			"include_usb_drives": schema.BoolAttribute{
				MarkdownDescription: "If `true`, periodically connected USB drives are " +
					"included in the backup. Optional, Computed. " +
					"Applies to `WindowsAgentBackup` job type only.",
				Optional: true,
				Computed: true,
			},
			"agent_type": schema.StringAttribute{
				MarkdownDescription: "Protected computer type for Windows agent jobs. " +
					"Optional, Computed. Applies to `WindowsAgentBackup` job type only. " +
					"Allowed values: `Workstation`, `Server`, `FailoverCluster`.",
				Optional: true,
				Computed: true,
			},
			"use_snapshotless_file_level_backup": schema.BoolAttribute{
				MarkdownDescription: "If `true`, creates a crash-consistent file-level backup " +
					"without a snapshot. Optional, Computed. " +
					"Applies to `LinuxAgentBackup` job type only, " +
					"when `agent_backup_mode` = `FileLevel`.",
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
			// Agent volume scope (WindowsAgentBackup / LinuxAgentBackup with mode=Volumes)
			// -----------------------------------------------------------------
			"volumes_scope": schema.SingleNestedAttribute{
				MarkdownDescription: "Selects which volumes to back up for agent-based jobs. " +
					"Only used when `agent_backup_mode = \"Volumes\"` for " +
					"`WindowsAgentBackup` and `LinuxAgentBackup` job types.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"all_volumes": schema.BoolAttribute{
						MarkdownDescription: "If `true`, all local volumes are included in the backup. " +
							"When `false`, the `volume_names` list determines which volumes are backed up.",
						Required: true,
					},
					"volume_names": schema.ListAttribute{
						ElementType: types.StringType,
						MarkdownDescription: "List of volume mount-points or drive letters to back up. " +
							"Used when `all_volumes = false`. " +
							"Windows examples: `[\"C:\\\\\", \"D:\\\\\"]`. Linux examples: `[\"/\", \"/data\"]`.",
						Optional: true,
						Computed: true,
					},
				},
			},

			// -----------------------------------------------------------------
			// Agent file-level scope (WindowsAgentBackup / LinuxAgentBackup with mode=FileLevel)
			// -----------------------------------------------------------------
			"files_scope": schema.SingleNestedAttribute{
				MarkdownDescription: "Selects which directories to include or exclude for agent file-level backups. " +
					"Only used when `agent_backup_mode = \"FileLevel\"` for " +
					"`WindowsAgentBackup` and `LinuxAgentBackup` job types.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"included_folders": schema.ListAttribute{
						ElementType: types.StringType,
						MarkdownDescription: "List of directory paths to include in the backup. " +
							"Windows example: `[\"C:\\\\Users\", \"D:\\\\Data\"]`. " +
							"Linux example: `[\"/home\", \"/etc\"]`.",
						Optional: true,
						Computed: true,
					},
					"excluded_folders": schema.ListAttribute{
						ElementType: types.StringType,
						MarkdownDescription: "List of directory paths to exclude from the backup. " +
							"Exclusions are applied after inclusions.",
						Optional: true,
						Computed: true,
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
					"gfs_policy": schema.SingleNestedAttribute{
						MarkdownDescription: "Grandfather-Father-Son (GFS) long-term archival retention policy. " +
							"When configured, Veeam preserves selected weekly, monthly, and yearly " +
							"full backups beyond the standard retention window.",
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"is_enabled": schema.BoolAttribute{
								MarkdownDescription: "Master switch that activates the GFS retention policy.",
								Required:            true,
							},
							"weekly_enabled": schema.BoolAttribute{
								MarkdownDescription: "Activate weekly GFS archival. " +
									"Preserves one full backup per week for the configured duration.",
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"weekly_keep_for": schema.Int64Attribute{
								MarkdownDescription: "Number of weeks to retain weekly full backups (1\u20139999). " +
									"Only evaluated when `weekly_enabled = true`.",
								Optional: true,
								Computed: true,
							},
							"weekly_desired_time": schema.StringAttribute{
								MarkdownDescription: "Day of the week on which the weekly full backup is created. " +
									"Allowed values: `Monday`, `Tuesday`, `Wednesday`, `Thursday`, " +
									"`Friday`, `Saturday`, `Sunday`.",
								Optional: true,
								Computed: true,
							},
							"monthly_enabled": schema.BoolAttribute{
								MarkdownDescription: "Activate monthly GFS archival. " +
									"Preserves one full backup per month for the configured duration.",
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"monthly_keep_for": schema.Int64Attribute{
								MarkdownDescription: "Number of months to retain monthly full backups (1\u2013999). " +
									"Only evaluated when `monthly_enabled = true`.",
								Optional: true,
								Computed: true,
							},
							"monthly_desired_time": schema.StringAttribute{
								MarkdownDescription: "Week of the month on which the monthly full backup is created. " +
									"Allowed values: `First`, `Second`, `Third`, `Fourth`, `Last`.",
								Optional: true,
								Computed: true,
							},
							"yearly_enabled": schema.BoolAttribute{
								MarkdownDescription: "Activate yearly GFS archival. " +
									"Preserves one full backup per year for the configured duration.",
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(false),
							},
							"yearly_keep_for": schema.Int64Attribute{
								MarkdownDescription: "Number of years to retain yearly full backups (1\u2013999). " +
									"Only evaluated when `yearly_enabled = true`.",
								Optional: true,
								Computed: true,
							},
							"yearly_desired_time": schema.StringAttribute{
								MarkdownDescription: "Month of the year in which the yearly full backup is created. " +
									"Allowed values: `January`, `February`, `March`, `April`, `May`, " +
									"`June`, `July`, `August`, `September`, `October`, `November`, `December`.",
								Optional: true,
								Computed: true,
							},
						},
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
					"guest_credentials": schema.SingleNestedAttribute{
						MarkdownDescription: "Guest OS credentials used for application-aware processing. " +
							"When omitted, Veeam uses the credentials configured on the managed server. " +
							"Applies to `VSphereBackup` and `HyperVBackup` job types only.",
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"credentials_id": schema.StringAttribute{
								MarkdownDescription: "UUID of the credentials record to use for guest OS interaction. " +
									"Obtain from the `veeam_credentials` data source or `veeam_credential` resource.",
								Required: true,
							},
						},
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
							"Allowed values: `Everyday`, `WeekDays`, `SelectedDays`.",
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
		if data.GuestProcessing != nil {
			resp.Diagnostics.AddError(
				"guest_processing is not supported for agent backup jobs",
				fmt.Sprintf("Job type '%s' does not support guest_processing in this provider. Use VSphereBackup/HyperVBackup for guest processing settings.", jobType),
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
		var result map[string]interface{}
		if err := r.client.PostJSON(ctx, client.PathJobs, spec, &result); err != nil {
			resp.Diagnostics.AddError("Failed to create agent backup job",
				fmt.Sprintf("POST %s: %s", client.PathJobs, err))
			return
		}
		if id, ok := result["id"].(string); ok && id != "" {
			data.ID = types.StringValue(id)
		}
		r.syncAgentJobFromAPIMap(&data, result)

	default:
		resp.Diagnostics.AddError(
			"Unsupported job type",
			fmt.Sprintf("Job type '%s' is not supported by this resource. Use the Veeam console "+
				"to manage %s jobs, or check if a dedicated Terraform resource exists.", jobType, jobType),
		)
		return
	}

	r.normalizeUnknownStateFields(&data)

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
		var result map[string]interface{}
		if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
			resp.Diagnostics.AddError("Failed to read agent backup job",
				fmt.Sprintf("GET %s: %s", endpoint, err))
			return
		}
		r.syncAgentJobFromAPIMap(&data, result)

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

	r.normalizeUnknownStateFields(&data)

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
		if data.GuestProcessing != nil {
			resp.Diagnostics.AddError(
				"guest_processing is not supported for agent backup jobs",
				fmt.Sprintf("Job type '%s' does not support guest_processing in this provider. Use VSphereBackup/HyperVBackup for guest processing settings.", jobType),
			)
			return
		}
		payload := r.buildAgentJobModel(&data, state.IsDisabled.ValueBool())

		// Agent job PUT expects immutable discriminator fields to be preserved.
		// Fetch the current model and carry forward agentType when present.
		var current map[string]any
		if err := r.client.GetJSON(ctx, endpoint, &current); err == nil {
			if agentType, ok := current["agentType"].(string); ok && agentType != "" {
				payload["agentType"] = agentType
			}
		}

		var result map[string]interface{}
		if err := r.client.PutJSON(ctx, endpoint, payload, &result); err != nil {
			resp.Diagnostics.AddError("Failed to update agent backup job",
				fmt.Sprintf("PUT %s: %s", endpoint, err))
			return
		}
		if id, ok := result["id"].(string); ok && id != "" {
			r.syncAgentJobFromAPIMap(&data, result)
		}

	default:
		resp.Diagnostics.AddError(
			"Unsupported job type for update",
			fmt.Sprintf("Cannot update job type '%s' via this resource.", jobType),
		)
		return
	}

	r.normalizeUnknownStateFields(&data)

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

	jobType := models.EJobType(data.Type.ValueString())

	if jobType == models.JobTypeWindowsAgentBackup {
		if !data.IncludeUsbDrives.IsNull() && !data.IncludeUsbDrives.IsUnknown() {
			spec["includeUsbDrives"] = data.IncludeUsbDrives.ValueBool()
		}
		if !data.AgentType.IsNull() && !data.AgentType.IsUnknown() && data.AgentType.ValueString() != "" {
			spec["agentType"] = data.AgentType.ValueString()
		}
	}

	if jobType == models.JobTypeLinuxAgentBackup {
		if !data.UseSnapshotlessFileLevelBackup.IsNull() && !data.UseSnapshotlessFileLevelBackup.IsUnknown() {
			spec["useSnapshotlessFileLevelBackup"] = data.UseSnapshotlessFileLevelBackup.ValueBool()
		}
	}

	if data.Storage != nil {
		spec["storage"] = r.buildAgentStorageModel(data.Storage)
	}

	if data.VolumesScope != nil {
		spec["volumes"] = buildAgentVolumesScopeModel(data.VolumesScope)
	}

	if data.FilesScope != nil {
		spec["files"] = buildAgentFilesScopeModel(data.FilesScope)
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

	jobType := models.EJobType(data.Type.ValueString())

	if jobType == models.JobTypeWindowsAgentBackup {
		if !data.IncludeUsbDrives.IsNull() && !data.IncludeUsbDrives.IsUnknown() {
			m["includeUsbDrives"] = data.IncludeUsbDrives.ValueBool()
		}
		if !data.AgentType.IsNull() && !data.AgentType.IsUnknown() && data.AgentType.ValueString() != "" {
			m["agentType"] = data.AgentType.ValueString()
		}
	}

	if jobType == models.JobTypeLinuxAgentBackup {
		if !data.UseSnapshotlessFileLevelBackup.IsNull() && !data.UseSnapshotlessFileLevelBackup.IsUnknown() {
			m["useSnapshotlessFileLevelBackup"] = data.UseSnapshotlessFileLevelBackup.ValueBool()
		}
	}

	if data.Storage != nil {
		m["storage"] = r.buildAgentStorageModel(data.Storage)
	}

	if data.VolumesScope != nil {
		m["volumes"] = buildAgentVolumesScopeModel(data.VolumesScope)
	}

	if data.FilesScope != nil {
		m["files"] = buildAgentFilesScopeModel(data.FilesScope)
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

	if !s.RetentionType.IsNull() && !s.RetentionType.IsUnknown() {
		qty := 14 // safe default — matches Veeam console default
		if !s.RetentionQuantity.IsNull() && !s.RetentionQuantity.IsUnknown() {
			qty = int(s.RetentionQuantity.ValueInt64())
		}
		m.RetentionPolicy = &models.BackupJobRetentionPolicySettings{
			Type:     models.ERetentionPolicyType(s.RetentionType.ValueString()),
			Quantity: qty,
		}
	}

	m.GFSPolicy = buildGFSPolicyModel(s.GFSPolicy)

	return m
}

func (r *BackupJob) buildAgentStorageModel(s *JobStorageSettings) *models.AgentBackupJobStorageModel {
	if s == nil {
		return nil
	}

	m := &models.AgentBackupJobStorageModel{
		BackupRepositoryID: s.RepositoryID.ValueString(),
	}

	if !s.RetentionType.IsNull() && !s.RetentionType.IsUnknown() {
		qty := 14
		if !s.RetentionQuantity.IsNull() && !s.RetentionQuantity.IsUnknown() {
			qty = int(s.RetentionQuantity.ValueInt64())
		}
		m.RetentionPolicy = &models.BackupJobRetentionPolicySettings{
			Type:     models.ERetentionPolicyType(s.RetentionType.ValueString()),
			Quantity: qty,
		}
	}

	m.GFSPolicy = buildGFSPolicyModel(s.GFSPolicy)

	return m
}

// buildGFSPolicyModel converts a JobGFSPolicy Terraform model into an API GFSPolicySettingsModel.
// Returns nil when gfs is nil so the API field is omitted entirely.
func buildGFSPolicyModel(gfs *JobGFSPolicy) *models.GFSPolicySettingsModel {
	if gfs == nil {
		return nil
	}

	m := &models.GFSPolicySettingsModel{
		IsEnabled: gfs.IsEnabled.ValueBool(),
	}

	if !gfs.WeeklyEnabled.IsNull() && gfs.WeeklyEnabled.ValueBool() {
		w := &models.GFSPolicySettingsWeeklyModel{IsEnabled: true}
		if !gfs.WeeklyKeepFor.IsNull() && !gfs.WeeklyKeepFor.IsUnknown() {
			w.KeepForNumberOfWeeks = int(gfs.WeeklyKeepFor.ValueInt64())
		}
		if !gfs.WeeklyDesiredTime.IsNull() {
			w.DesiredTime = models.EDayOfWeek(gfs.WeeklyDesiredTime.ValueString())
		}
		m.Weekly = w
	}

	if !gfs.MonthlyEnabled.IsNull() && gfs.MonthlyEnabled.ValueBool() {
		mo := &models.GFSPolicySettingsMonthlyModel{IsEnabled: true}
		if !gfs.MonthlyKeepFor.IsNull() && !gfs.MonthlyKeepFor.IsUnknown() {
			mo.KeepForNumberOfMonths = int(gfs.MonthlyKeepFor.ValueInt64())
		}
		if !gfs.MonthlyDesiredTime.IsNull() {
			mo.DesiredTime = models.ESennightOfMonth(gfs.MonthlyDesiredTime.ValueString())
		}
		m.Monthly = mo
	}

	if !gfs.YearlyEnabled.IsNull() && gfs.YearlyEnabled.ValueBool() {
		y := &models.GFSPolicySettingsYearlyModel{IsEnabled: true}
		if !gfs.YearlyKeepFor.IsNull() && !gfs.YearlyKeepFor.IsUnknown() {
			y.KeepForNumberOfYears = int(gfs.YearlyKeepFor.ValueInt64())
		}
		if !gfs.YearlyDesiredTime.IsNull() {
			y.DesiredTime = models.EMonth(gfs.YearlyDesiredTime.ValueString())
		}
		m.Yearly = y
	}

	return m
}

// syncGFSPolicyFromAPI converts an API GFSPolicySettingsModel into a JobGFSPolicy.
// Returns nil when the API returns no GFS policy, leaving the Terraform attribute null.
func syncGFSPolicyFromAPI(api *models.GFSPolicySettingsModel) *JobGFSPolicy {
	if api == nil {
		return nil
	}

	gfs := &JobGFSPolicy{
		IsEnabled:          types.BoolValue(api.IsEnabled),
		WeeklyEnabled:      types.BoolValue(false),
		WeeklyKeepFor:      types.Int64Null(),
		WeeklyDesiredTime:  types.StringNull(),
		MonthlyEnabled:     types.BoolValue(false),
		MonthlyKeepFor:     types.Int64Null(),
		MonthlyDesiredTime: types.StringNull(),
		YearlyEnabled:      types.BoolValue(false),
		YearlyKeepFor:      types.Int64Null(),
		YearlyDesiredTime:  types.StringNull(),
	}

	if api.Weekly != nil {
		gfs.WeeklyEnabled = types.BoolValue(api.Weekly.IsEnabled)
		if api.Weekly.KeepForNumberOfWeeks > 0 {
			gfs.WeeklyKeepFor = types.Int64Value(int64(api.Weekly.KeepForNumberOfWeeks))
		}
		if api.Weekly.DesiredTime != "" {
			gfs.WeeklyDesiredTime = types.StringValue(string(api.Weekly.DesiredTime))
		}
	}

	if api.Monthly != nil {
		gfs.MonthlyEnabled = types.BoolValue(api.Monthly.IsEnabled)
		if api.Monthly.KeepForNumberOfMonths > 0 {
			gfs.MonthlyKeepFor = types.Int64Value(int64(api.Monthly.KeepForNumberOfMonths))
		}
		if api.Monthly.DesiredTime != "" {
			gfs.MonthlyDesiredTime = types.StringValue(string(api.Monthly.DesiredTime))
		}
	}

	if api.Yearly != nil {
		gfs.YearlyEnabled = types.BoolValue(api.Yearly.IsEnabled)
		if api.Yearly.KeepForNumberOfYears > 0 {
			gfs.YearlyKeepFor = types.Int64Value(int64(api.Yearly.KeepForNumberOfYears))
		}
		if api.Yearly.DesiredTime != "" {
			gfs.YearlyDesiredTime = types.StringValue(string(api.Yearly.DesiredTime))
		}
	}

	return gfs
}

// buildAgentVolumesScopeModel converts AgentVolumesScope into the API volumes model.
func buildAgentVolumesScopeModel(vs *AgentVolumesScope) *models.AgentBackupJobVolumesModel {
	if vs == nil {
		return nil
	}
	m := &models.AgentBackupJobVolumesModel{
		AllVolumes: vs.AllVolumes.ValueBool(),
	}
	if !vs.VolumeNames.IsNull() && !vs.VolumeNames.IsUnknown() {
		var names []string
		// Ignore conversion error — an empty list is valid.
		_ = vs.VolumeNames.ElementsAs(context.Background(), &names, false)
		m.VolumeNames = names
	}
	return m
}

// buildAgentFilesScopeModel converts AgentFilesScope into the API files model.
func buildAgentFilesScopeModel(fs *AgentFilesScope) *models.AgentBackupJobFilesModel {
	if fs == nil {
		return nil
	}
	m := &models.AgentBackupJobFilesModel{}
	if !fs.IncludedFolders.IsNull() && !fs.IncludedFolders.IsUnknown() {
		var folders []string
		_ = fs.IncludedFolders.ElementsAs(context.Background(), &folders, false)
		m.IncludedFolders = folders
	}
	if !fs.ExcludedFolders.IsNull() && !fs.ExcludedFolders.IsUnknown() {
		var folders []string
		_ = fs.ExcludedFolders.ElementsAs(context.Background(), &folders, false)
		m.ExcludedFolders = folders
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

	if gp.GuestCredentials != nil && !gp.GuestCredentials.CredentialsID.IsNull() &&
		gp.GuestCredentials.CredentialsID.ValueString() != "" {
		m.GuestCredentials = &models.GuestOsCredentialsModel{
			CredentialsID: gp.GuestCredentials.CredentialsID.ValueString(),
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
		s.GFSPolicy = syncGFSPolicyFromAPI(api.Storage.GFSPolicy)
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
		if api.GuestProcessing.GuestCredentials != nil {
			gp.GuestCredentials = &JobGuestCredentials{
				CredentialsID: types.StringValue(api.GuestProcessing.GuestCredentials.CredentialsID),
			}
		}
		data.GuestProcessing = gp
	}

	// Sync schedule.
	if api.Schedule != nil {
		data.Schedule = r.syncScheduleFromAPI(data.Schedule, api.Schedule)
	}
}

// syncAgentJobFromAPIMap merges an agent job API response (raw map) into Terraform state.
// Agent jobs are decoded into map[string]interface{} because BackupJobModel does not carry
// the agent-specific fields (backupMode, computers, includeUsbDrives, agentType, etc.).
func (r *BackupJob) syncAgentJobFromAPIMap(data *BackupJobModel, api map[string]interface{}) {
	data.IncludeUsbDrives = types.BoolNull()
	data.AgentType = types.StringNull()
	data.UseSnapshotlessFileLevelBackup = types.BoolNull()

	if v, ok := api["name"].(string); ok && v != "" {
		data.Name = types.StringValue(v)
	}
	if v, ok := api["type"].(string); ok && v != "" {
		data.Type = types.StringValue(v)
	}
	if v, ok := api["isDisabled"].(bool); ok {
		data.IsDisabled = types.BoolValue(v)
	}
	if v, ok := api["isHighPriority"].(bool); ok {
		data.IsHighPriority = types.BoolValue(v)
	}
	if v, ok := api["description"].(string); ok && v != "" {
		data.Description = types.StringValue(v)
	}
	if v, ok := api["backupMode"].(string); ok && v != "" {
		data.AgentBackupMode = types.StringValue(v)
	}

	jobTypeStr := data.Type.ValueString()

	// Windows-agent-specific fields.
	if jobTypeStr == string(models.JobTypeWindowsAgentBackup) {
		if v, ok := api["includeUsbDrives"].(bool); ok {
			data.IncludeUsbDrives = types.BoolValue(v)
		}
		if v, ok := api["agentType"].(string); ok {
			data.AgentType = types.StringValue(v)
		}
	}

	// Linux-agent-specific fields.
	if jobTypeStr == string(models.JobTypeLinuxAgentBackup) {
		if v, ok := api["useSnapshotlessFileLevelBackup"].(bool); ok {
			data.UseSnapshotlessFileLevelBackup = types.BoolValue(v)
		}
	}

	// Sync computers list when present in the response.
	if computersRaw, ok := api["computers"].([]interface{}); ok {
		computers := make([]AgentComputerEntry, 0, len(computersRaw))
		for _, raw := range computersRaw {
			c, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			entry := AgentComputerEntry{
				ID:                types.StringValue(""),
				Name:              types.StringValue(""),
				Type:              types.StringValue(""),
				ProtectionGroupID: types.StringValue(""),
			}
			if v, ok := c["id"].(string); ok {
				entry.ID = types.StringValue(v)
			}
			if v, ok := c["name"].(string); ok {
				entry.Name = types.StringValue(v)
			}
			if v, ok := c["type"].(string); ok {
				entry.Type = types.StringValue(v)
			}
			if v, ok := c["protectionGroupId"].(string); ok {
				entry.ProtectionGroupID = types.StringValue(v)
			}
			computers = append(computers, entry)
		}
		if len(computers) > 0 {
			data.AgentComputers = computers
		}
	}

	// Sync storage when present.
	if storageRaw, ok := api["storage"].(map[string]interface{}); ok {
		s := &JobStorageSettings{
			RepositoryID:    types.StringValue(""),
			ProxyAutoSelect: types.BoolValue(false),
		}
		if v, ok := storageRaw["backupRepositoryId"].(string); ok {
			s.RepositoryID = types.StringValue(v)
		}
		if retRaw, ok := storageRaw["retentionPolicy"].(map[string]interface{}); ok {
			if v, ok := retRaw["type"].(string); ok {
				s.RetentionType = types.StringValue(v)
			}
			if v, ok := retRaw["quantity"].(float64); ok {
				s.RetentionQuantity = types.Int64Value(int64(v))
			}
		}
		data.Storage = s
	}

	// Sync volumes scope when present (agent jobs with backupMode=Volumes).
	if volRaw, ok := api["volumes"].(map[string]interface{}); ok {
		vs := &AgentVolumesScope{
			AllVolumes:  types.BoolValue(false),
			VolumeNames: types.ListValueMust(types.StringType, []attr.Value{}),
		}
		if v, ok := volRaw["allVolumes"].(bool); ok {
			vs.AllVolumes = types.BoolValue(v)
		}
		if namesRaw, ok := volRaw["volumeNames"].([]interface{}); ok {
			attrVals := make([]attr.Value, 0, len(namesRaw))
			for _, n := range namesRaw {
				if s, ok := n.(string); ok {
					attrVals = append(attrVals, types.StringValue(s))
				}
			}
			vs.VolumeNames = types.ListValueMust(types.StringType, attrVals)
		}
		data.VolumesScope = vs
	}

	// Sync files scope when present (agent jobs with backupMode=FileLevel).
	if filesRaw, ok := api["files"].(map[string]interface{}); ok {
		fs := &AgentFilesScope{
			IncludedFolders: types.ListValueMust(types.StringType, []attr.Value{}),
			ExcludedFolders: types.ListValueMust(types.StringType, []attr.Value{}),
		}
		if inclRaw, ok := filesRaw["includedFolders"].([]interface{}); ok {
			attrVals := make([]attr.Value, 0, len(inclRaw))
			for _, n := range inclRaw {
				if s, ok := n.(string); ok {
					attrVals = append(attrVals, types.StringValue(s))
				}
			}
			fs.IncludedFolders = types.ListValueMust(types.StringType, attrVals)
		}
		if exclRaw, ok := filesRaw["excludedFolders"].([]interface{}); ok {
			attrVals := make([]attr.Value, 0, len(exclRaw))
			for _, n := range exclRaw {
				if s, ok := n.(string); ok {
					attrVals = append(attrVals, types.StringValue(s))
				}
			}
			fs.ExcludedFolders = types.ListValueMust(types.StringType, attrVals)
		}
		data.FilesScope = fs
	}

	// Sync schedule when present.
	if schedRaw, ok := api["schedule"].(map[string]interface{}); ok {
		sched := r.syncScheduleFromAPIMap(data.Schedule, schedRaw)
		data.Schedule = sched
	}
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
		s.DailyLocalTime = types.StringNull()
		s.DailyKind = types.StringNull()
	}

	if api.Monthly != nil {
		s.MonthlyEnabled = types.BoolValue(api.Monthly.IsEnabled)
		s.MonthlyLocalTime = types.StringValue(api.Monthly.LocalTime)
		if api.Monthly.DayOfMonth > 0 {
			s.MonthlyDayOfMonth = types.Int64Value(int64(api.Monthly.DayOfMonth))
		} else {
			s.MonthlyDayOfMonth = types.Int64Null()
		}
	} else {
		s.MonthlyEnabled = types.BoolValue(false)
		s.MonthlyLocalTime = types.StringNull()
		s.MonthlyDayOfMonth = types.Int64Null()
	}

	if api.Periodically != nil {
		s.PeriodicallyEnabled = types.BoolValue(api.Periodically.IsEnabled)
		s.PeriodicallyKind = types.StringValue(string(api.Periodically.PeriodicallyKind))
		s.PeriodicallyFrequency = types.Int64Value(int64(api.Periodically.Frequency))
	} else {
		s.PeriodicallyEnabled = types.BoolValue(false)
		s.PeriodicallyKind = types.StringNull()
		s.PeriodicallyFrequency = types.Int64Null()
	}

	if api.AfterThisJob != nil {
		s.AfterJobEnabled = types.BoolValue(api.AfterThisJob.IsEnabled)
		s.AfterJobName = types.StringValue(api.AfterThisJob.JobName)
	} else {
		s.AfterJobEnabled = types.BoolValue(false)
		s.AfterJobName = types.StringNull()
	}

	if api.Retry != nil {
		s.RetryEnabled = types.BoolValue(api.Retry.IsEnabled)
		s.RetryCount = types.Int64Value(int64(api.Retry.RetryCount))
		s.RetryAwaitMinutes = types.Int64Value(int64(api.Retry.AwaitMinutes))
	} else {
		s.RetryEnabled = types.BoolValue(false)
		s.RetryCount = types.Int64Null()
		s.RetryAwaitMinutes = types.Int64Null()
	}

	return s
}

// normalizeUnknownStateFields converts unknown nested/computed fields to known
// nulls before persisting state. Terraform requires all values to be known
// after apply.
func (r *BackupJob) normalizeUnknownStateFields(data *BackupJobModel) {
	if data == nil {
		return
	}

	if data.AgentBackupMode.IsUnknown() {
		data.AgentBackupMode = types.StringNull()
	}
	if data.IncludeUsbDrives.IsUnknown() {
		data.IncludeUsbDrives = types.BoolNull()
	}
	if data.AgentType.IsUnknown() {
		data.AgentType = types.StringNull()
	}
	if data.UseSnapshotlessFileLevelBackup.IsUnknown() {
		data.UseSnapshotlessFileLevelBackup = types.BoolNull()
	}

	r.normalizeUnknownStorageFields(data.Storage)
	r.normalizeUnknownGuestProcessingFields(data.GuestProcessing)
	r.normalizeUnknownScheduleFields(data.Schedule)
}

func (r *BackupJob) normalizeUnknownStorageFields(storage *JobStorageSettings) {
	if storage == nil {
		return
	}

	if storage.RepositoryID.IsUnknown() {
		storage.RepositoryID = types.StringNull()
	}
	if storage.ProxyAutoSelect.IsUnknown() {
		storage.ProxyAutoSelect = types.BoolNull()
	}
	if storage.RetentionType.IsUnknown() {
		storage.RetentionType = types.StringNull()
	}
	if storage.RetentionQuantity.IsUnknown() {
		storage.RetentionQuantity = types.Int64Null()
	}
}

func (r *BackupJob) normalizeUnknownGuestProcessingFields(guestProcessing *JobGuestProcessing) {
	if guestProcessing == nil {
		return
	}

	if guestProcessing.AppAwareEnabled.IsUnknown() {
		guestProcessing.AppAwareEnabled = types.BoolNull()
	}
	if guestProcessing.FSIndexingEnabled.IsUnknown() {
		guestProcessing.FSIndexingEnabled = types.BoolNull()
	}
	if guestProcessing.InteractionProxyAutoSelect.IsUnknown() {
		guestProcessing.InteractionProxyAutoSelect = types.BoolNull()
	}
}

func (r *BackupJob) normalizeUnknownScheduleFields(schedule *JobScheduleSettings) {
	if schedule == nil {
		return
	}

	if schedule.RunAutomatically.IsUnknown() {
		schedule.RunAutomatically = types.BoolNull()
	}
	if schedule.DailyEnabled.IsUnknown() {
		schedule.DailyEnabled = types.BoolNull()
	}
	if schedule.DailyLocalTime.IsUnknown() {
		schedule.DailyLocalTime = types.StringNull()
	}
	if schedule.DailyKind.IsUnknown() {
		schedule.DailyKind = types.StringNull()
	}
	if schedule.MonthlyEnabled.IsUnknown() {
		schedule.MonthlyEnabled = types.BoolNull()
	}
	if schedule.MonthlyLocalTime.IsUnknown() {
		schedule.MonthlyLocalTime = types.StringNull()
	}
	if schedule.MonthlyDayOfMonth.IsUnknown() {
		schedule.MonthlyDayOfMonth = types.Int64Null()
	}
	if schedule.PeriodicallyEnabled.IsUnknown() {
		schedule.PeriodicallyEnabled = types.BoolNull()
	}
	if schedule.PeriodicallyKind.IsUnknown() {
		schedule.PeriodicallyKind = types.StringNull()
	}
	if schedule.PeriodicallyFrequency.IsUnknown() {
		schedule.PeriodicallyFrequency = types.Int64Null()
	}
	if schedule.AfterJobEnabled.IsUnknown() {
		schedule.AfterJobEnabled = types.BoolNull()
	}
	if schedule.AfterJobName.IsUnknown() {
		schedule.AfterJobName = types.StringNull()
	}
	if schedule.RetryEnabled.IsUnknown() {
		schedule.RetryEnabled = types.BoolNull()
	}
	if schedule.RetryCount.IsUnknown() {
		schedule.RetryCount = types.Int64Null()
	}
	if schedule.RetryAwaitMinutes.IsUnknown() {
		schedule.RetryAwaitMinutes = types.Int64Null()
	}
}

// syncScheduleFromAPIMap updates a JobScheduleSettings from a raw map API response.
// Used for agent job types where the response is decoded as map[string]interface{}.
func (r *BackupJob) syncScheduleFromAPIMap(existing *JobScheduleSettings, api map[string]interface{}) *JobScheduleSettings {
	s := &JobScheduleSettings{}
	if existing != nil {
		*s = *existing
	}

	if v, ok := api["runAutomatically"].(bool); ok {
		s.RunAutomatically = types.BoolValue(v)
	}

	if dailyRaw, ok := api["daily"].(map[string]interface{}); ok {
		if v, ok := dailyRaw["isEnabled"].(bool); ok {
			s.DailyEnabled = types.BoolValue(v)
		}
		if v, ok := dailyRaw["localTime"].(string); ok {
			s.DailyLocalTime = types.StringValue(v)
		}
		if v, ok := dailyRaw["dailyKind"].(string); ok {
			s.DailyKind = types.StringValue(v)
		}
	} else {
		s.DailyEnabled = types.BoolValue(false)
	}

	if monthlyRaw, ok := api["monthly"].(map[string]interface{}); ok {
		if v, ok := monthlyRaw["isEnabled"].(bool); ok {
			s.MonthlyEnabled = types.BoolValue(v)
		}
		if v, ok := monthlyRaw["localTime"].(string); ok {
			s.MonthlyLocalTime = types.StringValue(v)
		}
		if v, ok := monthlyRaw["dayOfMonth"].(float64); ok && v > 0 {
			s.MonthlyDayOfMonth = types.Int64Value(int64(v))
		}
	} else {
		s.MonthlyEnabled = types.BoolValue(false)
	}

	if periodRaw, ok := api["periodically"].(map[string]interface{}); ok {
		if v, ok := periodRaw["isEnabled"].(bool); ok {
			s.PeriodicallyEnabled = types.BoolValue(v)
		}
		if v, ok := periodRaw["periodicallyKind"].(string); ok {
			s.PeriodicallyKind = types.StringValue(v)
		}
		if v, ok := periodRaw["frequency"].(float64); ok {
			s.PeriodicallyFrequency = types.Int64Value(int64(v))
		}
	} else {
		s.PeriodicallyEnabled = types.BoolValue(false)
	}

	if afterRaw, ok := api["afterThisJob"].(map[string]interface{}); ok {
		if v, ok := afterRaw["isEnabled"].(bool); ok {
			s.AfterJobEnabled = types.BoolValue(v)
		}
		if v, ok := afterRaw["jobName"].(string); ok {
			s.AfterJobName = types.StringValue(v)
		}
	} else {
		s.AfterJobEnabled = types.BoolValue(false)
	}

	if retryRaw, ok := api["retry"].(map[string]interface{}); ok {
		if v, ok := retryRaw["isEnabled"].(bool); ok {
			s.RetryEnabled = types.BoolValue(v)
		}
		if v, ok := retryRaw["retryCount"].(float64); ok {
			s.RetryCount = types.Int64Value(int64(v))
		}
		if v, ok := retryRaw["awaitMinutes"].(float64); ok {
			s.RetryAwaitMinutes = types.Int64Value(int64(v))
		}
	} else {
		s.RetryEnabled = types.BoolValue(false)
	}

	return s
}

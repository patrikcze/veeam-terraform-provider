package models

// ---------------------------------------------------------------------------
// Jobs — V13 REST API: /api/v1/jobs
//
// The /api/v1/jobs endpoint is polymorphic: the "type" field acts as a
// discriminator that selects the concrete sub-schema for each job kind.
//
// Supported job types (EJobType):
//   VSphereBackup                      → BackupJobSpec / BackupJobModel
//   HyperVBackup                       → HyperVBackupJobSpec / HyperVBackupJobModel
//   BackupCopy                         → BackupCopyJobSpec / BackupCopyJobModel
//   WindowsAgentBackup                 → WindowsAgentBackupJobSpec / WindowsAgentBackupJobModel
//   LinuxAgentBackup                   → LinuxAgentBackupJobSpec / LinuxAgentBackupJobModel
//   WindowsAgentBackupServerPolicy     → (uses WindowsAgentBackupJobSpec with server policy dest.)
//   LinuxAgentBackupServerPolicy       → (uses LinuxAgentBackupJobSpec with server policy dest.)
//   VSphereReplica                     → (not yet implemented in Terraform resource)
//
// CRUD behaviour (all synchronous — no async polling needed for job management):
//   POST   /api/v1/jobs          → 201 Created, body: JobModel  (create)
//   GET    /api/v1/jobs/{id}     → 200 OK,      body: JobModel  (read)
//   PUT    /api/v1/jobs/{id}     → 200 OK,      body: JobModel  (update; sends full JobModel)
//   DELETE /api/v1/jobs/{id}     → 204 No Content                (delete)
//
// Async job control endpoints (start/stop/retry/clone) return SessionModel and
// are intentionally NOT managed by the Terraform resource — use them via the
// Veeam console or separate automation.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Base Job Models
// ---------------------------------------------------------------------------

// JobModel is the base response model for all job types.
// Discriminator field: "type" (EJobType).
type JobModel struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       EJobType `json:"type"`
	IsDisabled bool     `json:"isDisabled"`
}

// JobSpec is the minimal request body shared by all job create/update operations.
// Concrete job specs embed this via allOf in the OpenAPI schema.
type JobSpec struct {
	Name string   `json:"name"`
	Type EJobType `json:"type"`
}

// ---------------------------------------------------------------------------
// Inventory Object References
//
// Veeam uses a polymorphic InventoryObjectModel discriminated by "platform":
//   VSphere  → VmwareObjectSpec  (hostName, name, type, objectId)
//   HyperV   → (HyperV variant, reserved for future use)
//   Agent    → AgentObjectSpec   (id, name, type, protectionGroupId)
//
// In Go, we represent these as concrete structs that marshal to the correct
// JSON shape so the server can route them to the right sub-handler.
// ---------------------------------------------------------------------------

// VmwareObjectSpec is a VMware vSphere inventory object used in job includes/excludes.
// Corresponds to API schema VmwareObjectModel with platform="VSphere".
type VmwareObjectSpec struct {
	// Platform must be "VSphere" — required discriminator for the API.
	Platform string `json:"platform"`
	// HostName is the vCenter Server or ESXi hostname that owns this object.
	HostName string `json:"hostName"`
	// Name is the display name of the vSphere object (VM name, folder name, etc.).
	Name string `json:"name"`
	// Type identifies the object class (VirtualMachine, Folder, Datacenter, Cluster…).
	Type EVmwareInventoryType `json:"type,omitempty"`
	// ObjectID is the vSphere MoRef ID (e.g. "vm-101"). Required for all objects
	// except vCenter Servers and standalone ESXi hosts.
	ObjectID string `json:"objectId,omitempty"`
}

// AgentObjectSpec is an agent-managed object used in agent backup job computers lists.
// Corresponds to API schema AgentObjectModel with platform="Agent".
type AgentObjectSpec struct {
	// Platform must be "Agent" — required discriminator for the API.
	Platform string `json:"platform"`
	// ID is the unique identifier of the agent-managed object.
	ID string `json:"id"`
	// Name is the display name of the agent-managed object.
	Name string `json:"name"`
	// Type identifies the object class (ProtectionGroup, WindowsComputer, etc.).
	Type EAgentInventoryObjectType `json:"type"`
	// ProtectionGroupID is the ID of the protection group that contains this object.
	ProtectionGroupID string `json:"protectionGroupId"`
	// ParentObjectID is the ID of the parent container object (optional).
	ParentObjectID string `json:"parentObjectId,omitempty"`
}

// ---------------------------------------------------------------------------
// VSphereBackup / HyperVBackup — Spec (POST) and Model (GET/PUT)
// ---------------------------------------------------------------------------

// BackupJobSpec is the request body for creating a VSphereBackup job.
// API discriminator mapping: type="VSphereBackup" → BackupJobSpec.
// Required fields per API: description, virtualMachines.
type BackupJobSpec struct {
	JobSpec
	// Description is required by the Veeam API (must not be empty string).
	Description string `json:"description"`
	// IsHighPriority requests elevated resource-scheduler priority.
	IsHighPriority bool `json:"isHighPriority,omitempty"`
	// VirtualMachines is required — defines which VMs/containers to back up.
	VirtualMachines *BackupJobVirtualMachinesSpec `json:"virtualMachines"`
	// Storage configures target repository, proxies, retention, and advanced settings.
	Storage *BackupJobStorageModel `json:"storage,omitempty"`
	// GuestProcessing configures application-aware processing and file indexing.
	GuestProcessing *BackupJobGuestProcessingModel `json:"guestProcessing,omitempty"`
	// Schedule configures when the job runs automatically.
	Schedule *BackupScheduleModel `json:"schedule,omitempty"`
}

// BackupJobModel is the full response/update body for a VSphereBackup job.
// Used for GET responses and as the PUT request body.
// API discriminator mapping: type="VSphereBackup" → BackupJobModel.
type BackupJobModel struct {
	JobModel
	Description     string                         `json:"description,omitempty"`
	IsHighPriority  bool                           `json:"isHighPriority,omitempty"`
	VirtualMachines *BackupJobVirtualMachinesModel `json:"virtualMachines,omitempty"`
	Storage         *BackupJobStorageModel         `json:"storage,omitempty"`
	GuestProcessing *BackupJobGuestProcessingModel `json:"guestProcessing,omitempty"`
	Schedule        *BackupScheduleModel           `json:"schedule,omitempty"`
}

// ---------------------------------------------------------------------------
// Virtual Machines — includes / excludes
// ---------------------------------------------------------------------------

// BackupJobVirtualMachinesSpec defines the VM scope for a backup job (request body).
// Required: includes (at least one entry).
type BackupJobVirtualMachinesSpec struct {
	// Includes is the list of VMs or containers to protect. At least one entry required.
	Includes []VmwareObjectSpec `json:"includes"`
	// Excludes optionally removes specific VMs or disks from the backup scope.
	Excludes *BackupJobExclusionsSpec `json:"excludes,omitempty"`
}

// BackupJobVirtualMachinesModel is the VM scope returned in GET/PUT responses.
type BackupJobVirtualMachinesModel struct {
	Includes []VmwareObjectSpec   `json:"includes"`
	Excludes *BackupJobExclusions `json:"excludes,omitempty"`
}

// BackupJobExclusionsSpec defines objects to exclude from backup (request body).
type BackupJobExclusionsSpec struct {
	// VMs lists individual VMs to exclude from the backup scope.
	VMs []VmwareObjectSpec `json:"vms,omitempty"`
	// Disks lists specific VM disks to exclude (with per-disk selection mode).
	Disks []VmwareObjectDiskExclusion `json:"disks,omitempty"`
	// Templates configures template VM exclusion behaviour.
	Templates *BackupJobExclusionsTemplates `json:"templates,omitempty"`
}

// BackupJobExclusions is the response model for job exclusions.
type BackupJobExclusions struct {
	VMs       []VmwareObjectSpec            `json:"vms,omitempty"`
	Disks     []VmwareObjectDiskExclusion   `json:"disks,omitempty"`
	Templates *BackupJobExclusionsTemplates `json:"templates,omitempty"`
}

// BackupJobExclusionsTemplates configures VM template exclusion within a job.
type BackupJobExclusionsTemplates struct {
	// IsEnabled excludes ALL VM templates from the job when true.
	IsEnabled bool `json:"isEnabled,omitempty"`
	// ExcludeFromIncremental excludes templates from incremental backup passes only.
	ExcludeFromIncremental bool `json:"excludeFromIncremental,omitempty"`
}

// VmwareObjectDiskExclusion excludes specific disks from a VM backup.
// Corresponds to API schema VmwareObjectDiskModel.
type VmwareObjectDiskExclusion struct {
	// VMObject identifies the VM whose disks are being excluded.
	VMObject VmwareObjectSpec `json:"vmObject"`
	// DisksToProcess controls whether all, system-only, or selected disks are used.
	DisksToProcess EVmwareDisksTypeToProcess `json:"disksToProcess"`
	// Disks lists specific disk IDs when DisksToProcess = "SelectedDisks".
	Disks []string `json:"disks"`
	// RemoveFromVMConfiguration removes the disk from the VM config if true.
	RemoveFromVMConfiguration bool `json:"removeFromVMConfiguration,omitempty"`
}

// EVmwareDisksTypeToProcess controls how disk exclusions are applied.
type EVmwareDisksTypeToProcess string

const (
	DisksTypeAllDisks      EVmwareDisksTypeToProcess = "AllDisks"
	DisksTypeSystemOnly    EVmwareDisksTypeToProcess = "SystemOnly"
	DisksTypeSelectedDisks EVmwareDisksTypeToProcess = "SelectedDisks"
)

// ---------------------------------------------------------------------------
// Storage Settings (shared between VSphereBackup and HyperVBackup)
// ---------------------------------------------------------------------------

// BackupJobStorageModel configures where and how backups are stored.
// API schema: BackupJobStorageModel — required fields: backupRepositoryId,
// backupProxies, retentionPolicy.
type BackupJobStorageModel struct {
	// BackupRepositoryID is the UUID of the target backup repository.
	BackupRepositoryID string `json:"backupRepositoryId"`
	// BackupProxies configures proxy auto-selection or explicit proxy assignment.
	BackupProxies *BackupProxiesSettingsModel `json:"backupProxies"`
	// RetentionPolicy controls how many restore points or days of backups are kept.
	RetentionPolicy *BackupJobRetentionPolicySettings `json:"retentionPolicy"`
	// GFSPolicy optionally configures Grandfather-Father-Son long-term retention.
	GFSPolicy *GFSPolicySettingsModel `json:"gfsPolicy,omitempty"`
	// AdvancedSettings provides compression, dedup, encryption and notification options.
	AdvancedSettings *BackupJobAdvancedSettingsModel `json:"advancedSettings,omitempty"`
}

// BackupProxiesSettingsModel configures backup proxy selection.
type BackupProxiesSettingsModel struct {
	// AutoSelectEnabled lets Veeam automatically pick the best available proxy.
	AutoSelectEnabled bool `json:"autoSelectEnabled"`
	// ProxyIDs lists specific proxy UUIDs when AutoSelectEnabled is false.
	ProxyIDs []string `json:"proxyIds,omitempty"`
}

// BackupJobRetentionPolicySettings controls how long backups are retained.
type BackupJobRetentionPolicySettings struct {
	// Type determines whether retention is measured in restore points or calendar days.
	Type ERetentionPolicyType `json:"type"`
	// Quantity is the number of restore points or days to keep (must be > 0).
	Quantity int `json:"quantity"`
}

// ---------------------------------------------------------------------------
// GFS (Grandfather-Father-Son) Retention Policy
// ---------------------------------------------------------------------------

// GFSPolicySettingsModel configures long-term archival retention.
type GFSPolicySettingsModel struct {
	// IsEnabled activates the GFS long-term retention policy.
	IsEnabled bool `json:"isEnabled"`
	// Weekly configures weekly full-backup archival.
	Weekly *GFSPolicySettingsWeeklyModel `json:"weekly,omitempty"`
	// Monthly configures monthly full-backup archival.
	Monthly *GFSPolicySettingsMonthlyModel `json:"monthly,omitempty"`
	// Yearly configures yearly full-backup archival.
	Yearly *GFSPolicySettingsYearlyModel `json:"yearly,omitempty"`
}

// GFSPolicySettingsWeeklyModel configures weekly GFS archival.
type GFSPolicySettingsWeeklyModel struct {
	// IsEnabled activates weekly GFS archival.
	IsEnabled bool `json:"isEnabled"`
	// KeepForNumberOfWeeks is the retention period (1–9999 weeks).
	KeepForNumberOfWeeks int `json:"keepForNumberOfWeeks,omitempty"`
	// DesiredTime is the day of week on which the weekly full is created.
	DesiredTime EDayOfWeek `json:"desiredTime,omitempty"`
}

// GFSPolicySettingsMonthlyModel configures monthly GFS archival.
type GFSPolicySettingsMonthlyModel struct {
	// IsEnabled activates monthly GFS archival.
	IsEnabled bool `json:"isEnabled"`
	// KeepForNumberOfMonths is the retention period (1–999 months).
	KeepForNumberOfMonths int `json:"keepForNumberOfMonths,omitempty"`
	// DesiredTime is the week-of-month on which the monthly full is created.
	DesiredTime ESennightOfMonth `json:"desiredTime,omitempty"`
}

// GFSPolicySettingsYearlyModel configures yearly GFS archival.
type GFSPolicySettingsYearlyModel struct {
	// IsEnabled activates yearly GFS archival.
	IsEnabled bool `json:"isEnabled"`
	// KeepForNumberOfYears is the retention period (1–999 years).
	KeepForNumberOfYears int `json:"keepForNumberOfYears,omitempty"`
	// DesiredTime is the month in which the yearly full is created.
	DesiredTime EMonth `json:"desiredTime,omitempty"`
}

// ---------------------------------------------------------------------------
// Advanced Storage Settings
// ---------------------------------------------------------------------------

// BackupJobAdvancedSettingsModel holds advanced backup job configuration.
// Corresponds to API schema BackupJobAdvancedSettingsModel.
type BackupJobAdvancedSettingsModel struct {
	// BackupModeType controls how restore points are created (Incremental, Full, etc.).
	BackupModeType EBackupModeType `json:"backupModeType"`
	// StorageData configures compression, dedup, and encryption.
	StorageData *BackupStorageSettingModel `json:"storageData,omitempty"`
	// Notifications configures SNMP and email alerts for the job.
	Notifications *NotificationSettingsModel `json:"notifications,omitempty"`
	// VSphere provides vSphere-specific settings (CBT, VMware Tools quiescence).
	VSphere *BackupJobAdvancedSettingsVSphereModel `json:"vSphere,omitempty"`
}

// BackupStorageSettingModel configures storage-level data reduction and encryption.
// Corresponds to API schema BackupStorageSettingModel.
type BackupStorageSettingModel struct {
	// InlineDataDedupEnabled deduplicates VM data before writing to the repository.
	InlineDataDedupEnabled bool `json:"inlineDataDedupEnabled,omitempty"`
	// ExcludeSwapFileBlocks skips swap file blocks to reduce backup size.
	ExcludeSwapFileBlocks bool `json:"excludeSwapFileBlocks,omitempty"`
	// ExcludeDeletedFileBlocks skips blocks for deleted files.
	ExcludeDeletedFileBlocks bool `json:"excludeDeletedFileBlocks,omitempty"`
	// CompressionLevel sets the compression algorithm (Auto, None, Optimal, High, Extreme…).
	CompressionLevel ECompressionLevel `json:"compressionLevel,omitempty"`
	// StorageOptimization sets the block size (256KB, 512KB, 1MB, 4MB).
	StorageOptimization EStorageOptimization `json:"storageOptimization,omitempty"`
	// Encryption configures backup file encryption.
	Encryption *BackupStorageEncryptionModel `json:"encryption,omitempty"`
}

// BackupStorageEncryptionModel configures backup encryption settings.
type BackupStorageEncryptionModel struct {
	// IsEnabled activates encryption for backup files.
	IsEnabled bool `json:"isEnabled"`
	// EncryptionPasswordID is the UUID of the encryption password record.
	EncryptionPasswordID string `json:"encryptionPasswordId,omitempty"`
	// KMSServerID is the UUID of a KMS server for key-managed encryption.
	KMSServerID string `json:"kmsServerId,omitempty"`
}

// NotificationSettingsModel configures job notifications.
// Corresponds to API schema NotificationSettingsModel.
type NotificationSettingsModel struct {
	// SendSNMPNotifications enables SNMP traps for job events.
	SendSNMPNotifications bool `json:"sendSNMPNotifications,omitempty"`
	// EmailNotifications configures recipient addresses and trigger conditions.
	EmailNotifications *EmailNotificationSettingsModel `json:"emailNotifications,omitempty"`
}

// EmailNotificationSettingsModel configures email alerts for a job.
// Corresponds to API schema EmailNotificationSettingsModel.
type EmailNotificationSettingsModel struct {
	// IsEnabled activates email notifications for this job.
	IsEnabled bool `json:"isEnabled"`
	// Recipients is a list of destination email addresses.
	Recipients []string `json:"recipients,omitempty"`
	// NotifyOnSuccess sends an email when the job succeeds.
	NotifyOnSuccess bool `json:"notifyOnSuccess,omitempty"`
	// NotifyOnWarning sends an email when the job completes with warnings.
	NotifyOnWarning bool `json:"notifyOnWarning,omitempty"`
	// NotifyOnError sends an email when the job fails.
	NotifyOnError bool `json:"notifyOnError,omitempty"`
}

// BackupJobAdvancedSettingsVSphereModel holds vSphere-specific advanced settings.
type BackupJobAdvancedSettingsVSphereModel struct {
	// EnableVMWareToolsQuiescence uses VMware Tools to quiesce the VM file system.
	EnableVMWareToolsQuiescence bool `json:"enableVMWareToolsQuiescence,omitempty"`
}

// ---------------------------------------------------------------------------
// Guest Processing
// ---------------------------------------------------------------------------

// BackupJobGuestProcessingModel configures application-aware processing settings.
// Corresponds to API schema BackupJobGuestProcessingModel.
// Required fields: appAwareProcessing, guestFSIndexing.
type BackupJobGuestProcessingModel struct {
	// AppAwareProcessing controls application-consistent backup behaviour.
	AppAwareProcessing *BackupApplicationAwareProcessingModel `json:"appAwareProcessing"`
	// GuestFSIndexing controls VM guest OS file indexing for search.
	GuestFSIndexing *GuestFileSystemIndexingModel `json:"guestFSIndexing"`
	// GuestInteractionProxies configures which proxies deploy the runtime process.
	GuestInteractionProxies *GuestInteractionProxiesSettingsModel `json:"guestInteractionProxies,omitempty"`
	// GuestCredentials specifies OS-level credentials used for guest interaction.
	GuestCredentials *GuestOsCredentialsModel `json:"guestCredentials,omitempty"`
}

// GuestOsCredentialsModel identifies the credential record used for guest OS interaction.
// The credential must already exist; create it via the veeam_credential resource.
type GuestOsCredentialsModel struct {
	// CredentialsID is the UUID of the credential record.
	CredentialsID string `json:"credentialsId"`
}

// BackupApplicationAwareProcessingModel controls application-aware processing.
type BackupApplicationAwareProcessingModel struct {
	// IsEnabled activates application-aware processing for the job.
	IsEnabled bool `json:"isEnabled"`
}

// GuestFileSystemIndexingModel controls VM guest OS file indexing.
type GuestFileSystemIndexingModel struct {
	// IsEnabled activates file-level indexing for search inside VMs.
	IsEnabled bool `json:"isEnabled"`
}

// GuestInteractionProxiesSettingsModel configures guest interaction proxy selection.
type GuestInteractionProxiesSettingsModel struct {
	// AutoSelectEnabled lets Veeam pick the best available guest interaction proxy.
	AutoSelectEnabled bool `json:"autoSelectEnabled"`
	// ProxyIDs lists specific Windows server UUIDs to use as guest interaction proxies.
	ProxyIDs []string `json:"proxyIds,omitempty"`
}

// ---------------------------------------------------------------------------
// Schedule — shared by VSphereBackup, HyperVBackup, and Agent backup jobs
// ---------------------------------------------------------------------------

// BackupScheduleModel configures all scheduling options for a job.
// Corresponds to API schema BackupScheduleModel (allOf BackupScheduleModelDaily).
type BackupScheduleModel struct {
	// RunAutomatically enables automated scheduling for the job.
	RunAutomatically bool `json:"runAutomatically"`
	// Daily configures a daily (or weekday/selected-day) schedule.
	Daily *ScheduleDailyModel `json:"daily,omitempty"`
	// Monthly configures a monthly schedule.
	Monthly *ScheduleMonthlyModel `json:"monthly,omitempty"`
	// Periodically configures an interval-based schedule (every N hours/minutes).
	Periodically *SchedulePeriodicallyModel `json:"periodically,omitempty"`
	// AfterThisJob chains this job to run after another named job completes.
	AfterThisJob *ScheduleAfterThisJobModel `json:"afterThisJob,omitempty"`
	// Retry configures automatic retry on failure.
	Retry *ScheduleRetryModel `json:"retry,omitempty"`
	// BackupWindow restricts when the job is allowed to run.
	BackupWindow *ScheduleBackupWindowModel `json:"backupWindow,omitempty"`
}

// ScheduleDailyModel configures daily schedule options.
type ScheduleDailyModel struct {
	// IsEnabled activates the daily schedule.
	IsEnabled bool `json:"isEnabled"`
	// LocalTime is the start time in "HH:MM" format (server local time).
	LocalTime string `json:"localTime,omitempty"`
	// DailyKind selects which days the schedule applies to.
	DailyKind EDailyKinds `json:"dailyKind,omitempty"`
	// Days lists specific days of the week when DailyKind = "SelectedDays".
	Days []string `json:"days,omitempty"`
}

// ScheduleMonthlyModel configures monthly schedule options.
type ScheduleMonthlyModel struct {
	// IsEnabled activates the monthly schedule.
	IsEnabled bool `json:"isEnabled"`
	// LocalTime is the start time in "HH:MM" format.
	LocalTime string `json:"localTime,omitempty"`
	// DayOfMonth is the day number (1–28) when the job should run.
	DayOfMonth int `json:"dayOfMonth,omitempty"`
	// DayNumberInMonth selects a week within the month (First, Second…Last).
	DayNumberInMonth string `json:"dayNumberInMonth,omitempty"`
	// DayOfWeek selects the day of week within the chosen week.
	DayOfWeek string `json:"dayOfWeek,omitempty"`
	// Months lists the months in which the job should run.
	Months []string `json:"months,omitempty"`
	// IsLastDayOfMonth schedules the job on the last day of each selected month.
	IsLastDayOfMonth bool `json:"isLastDayOfMonth,omitempty"`
}

// SchedulePeriodicallyModel configures interval-based scheduling.
type SchedulePeriodicallyModel struct {
	// IsEnabled activates periodic scheduling.
	IsEnabled bool `json:"isEnabled"`
	// PeriodicallyKind is the time unit (Hours, Minutes, Seconds, Days).
	PeriodicallyKind EPeriodicallyKinds `json:"periodicallyKind,omitempty"`
	// Frequency is the number of time units between runs.
	Frequency int `json:"frequency,omitempty"`
}

// ScheduleAfterThisJobModel chains a job to run after another job completes.
// NOTE: The API identifies the preceding job by NAME, not by UUID.
type ScheduleAfterThisJobModel struct {
	// IsEnabled activates job chaining.
	IsEnabled bool `json:"isEnabled"`
	// JobName is the display name of the preceding job (not a UUID).
	JobName string `json:"jobName,omitempty"`
}

// ScheduleRetryModel configures job retry behaviour on failure.
type ScheduleRetryModel struct {
	// IsEnabled activates automatic retry.
	IsEnabled bool `json:"isEnabled"`
	// RetryCount is the number of retry attempts (must be > 0 when enabled).
	RetryCount int `json:"retryCount,omitempty"`
	// AwaitMinutes is the wait time between retries in minutes (must be > 0).
	AwaitMinutes int `json:"awaitMinutes,omitempty"`
}

// ScheduleBackupWindowModel restricts the allowed run window for periodic jobs.
type ScheduleBackupWindowModel struct {
	// IsEnabled activates backup window enforcement.
	IsEnabled bool `json:"isEnabled"`
}

// ---------------------------------------------------------------------------
// Windows/Linux Agent Backup Jobs
//
// Agent backup jobs use "computers" instead of "virtualMachines" and require
// a "backupMode" (EntireComputer, Volumes, or FileLevel).  The storage model
// differs from hypervisor backup jobs — use AgentBackupJobStorageModel.
// ---------------------------------------------------------------------------

// WindowsAgentBackupJobSpec is the request body for creating a Windows agent backup job.
// API discriminator mapping: type="WindowsAgentBackup" → WindowsAgentManagementBackupJobSpec.
// Required: description, computers, backupMode.
type WindowsAgentBackupJobSpec struct {
	JobSpec
	// Description is required by the Veeam API.
	Description string `json:"description"`
	// IsHighPriority requests elevated resource-scheduler priority.
	IsHighPriority bool `json:"isHighPriority,omitempty"`
	// Computers is the list of agent-managed objects (computers or protection groups).
	Computers []AgentObjectSpec `json:"computers"`
	// BackupMode selects the backup scope (EntireComputer, Volumes, or FileLevel).
	BackupMode EAgentBackupJobMode `json:"backupMode"`
	// IncludeUsbDrives includes periodically connected USB drives in the backup.
	IncludeUsbDrives bool `json:"includeUsbDrives,omitempty"`
	// AgentType selects the protected computer type (Workstation, Server, FailoverCluster).
	AgentType string `json:"agentType,omitempty"`
	// Storage configures the backup destination repository.
	Storage *AgentBackupJobStorageModel `json:"storage,omitempty"`
	// Schedule configures when the job runs.
	Schedule *BackupScheduleModel `json:"schedule,omitempty"`
}

// WindowsAgentBackupJobModel is the full response/update body for a Windows agent backup job.
type WindowsAgentBackupJobModel struct {
	JobModel
	Description      string                      `json:"description,omitempty"`
	IsHighPriority   bool                        `json:"isHighPriority,omitempty"`
	Computers        []AgentObjectSpec           `json:"computers,omitempty"`
	BackupMode       EAgentBackupJobMode         `json:"backupMode,omitempty"`
	IncludeUsbDrives bool                        `json:"includeUsbDrives,omitempty"`
	AgentType        string                      `json:"agentType,omitempty"`
	Storage          *AgentBackupJobStorageModel `json:"storage,omitempty"`
	Schedule         *BackupScheduleModel        `json:"schedule,omitempty"`
}

// LinuxAgentBackupJobSpec is the request body for creating a Linux agent backup job.
// API discriminator mapping: type="LinuxAgentBackup" → LinuxAgentManagementBackupJobSpec.
// Required: description, computers, backupMode.
type LinuxAgentBackupJobSpec struct {
	JobSpec
	Description    string              `json:"description"`
	IsHighPriority bool                `json:"isHighPriority,omitempty"`
	Computers      []AgentObjectSpec   `json:"computers"`
	BackupMode     EAgentBackupJobMode `json:"backupMode"`
	// UseSnapshotlessFileLevelBackup creates a crash-consistent backup without a snapshot.
	// Only applicable when BackupMode = "FileLevel".
	UseSnapshotlessFileLevelBackup bool                        `json:"useSnapshotlessFileLevelBackup,omitempty"`
	Storage                        *AgentBackupJobStorageModel `json:"storage,omitempty"`
	Schedule                       *BackupScheduleModel        `json:"schedule,omitempty"`
}

// LinuxAgentBackupJobModel is the full response/update body for a Linux agent backup job.
type LinuxAgentBackupJobModel struct {
	JobModel
	Description                    string                      `json:"description,omitempty"`
	IsHighPriority                 bool                        `json:"isHighPriority,omitempty"`
	Computers                      []AgentObjectSpec           `json:"computers,omitempty"`
	BackupMode                     EAgentBackupJobMode         `json:"backupMode,omitempty"`
	UseSnapshotlessFileLevelBackup bool                        `json:"useSnapshotlessFileLevelBackup,omitempty"`
	Storage                        *AgentBackupJobStorageModel `json:"storage,omitempty"`
	Schedule                       *BackupScheduleModel        `json:"schedule,omitempty"`
}

// AgentBackupJobStorageModel configures storage for agent backup jobs.
// Unlike hypervisor jobs, the repository is optional (agents can write locally).
type AgentBackupJobStorageModel struct {
	// BackupRepositoryID is the UUID of the target Veeam backup repository.
	// Leave empty for local/shared-folder destinations defined via backup policy.
	BackupRepositoryID string `json:"backupRepositoryId,omitempty"`
	// RetentionPolicy controls how many restore points or days of backups are kept.
	RetentionPolicy *BackupJobRetentionPolicySettings `json:"retentionPolicy,omitempty"`
	// GFSPolicy optionally configures Grandfather-Father-Son long-term retention.
	GFSPolicy *GFSPolicySettingsModel `json:"gfsPolicy,omitempty"`
}

// AgentBackupJobVolumesModel configures which volumes to back up for agent jobs.
// Used when BackupMode = "Volumes".
type AgentBackupJobVolumesModel struct {
	// AllVolumes backs up every local volume when true.
	AllVolumes bool `json:"allVolumes"`
	// VolumeNames lists specific volumes to include when AllVolumes is false.
	// Windows: drive letters (e.g. "C:", "D:"). Linux: mount points (e.g. "/", "/data").
	VolumeNames []string `json:"volumeNames,omitempty"`
}

// AgentBackupJobFilesModel configures which directories to include or exclude
// for agent file-level backup jobs (BackupMode = "FileLevel").
type AgentBackupJobFilesModel struct {
	// IncludedFolders lists directory paths to include in the backup.
	IncludedFolders []string `json:"includedFolders,omitempty"`
	// ExcludedFolders lists directory paths to exclude from the backup.
	ExcludedFolders []string `json:"excludedFolders,omitempty"`
}

// ---------------------------------------------------------------------------
// Job State Model (read-only, returned by GET /api/v1/jobs/states)
// ---------------------------------------------------------------------------

// JobStateModel is the current operational state of a job.
// Returned by GET /api/v1/jobs/states and GET /api/v1/jobs/states?jobId={id}.
type JobStateModel struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            EJobType       `json:"type"`
	Description     string         `json:"description"`
	Status          EJobStatus     `json:"status"`
	LastRun         string         `json:"lastRun,omitempty"`
	LastResult      ESessionResult `json:"lastResult"`
	NextRun         string         `json:"nextRun,omitempty"`
	Workload        EJobWorkload   `json:"workload"`
	RepositoryID    string         `json:"repositoryId,omitempty"`
	RepositoryName  string         `json:"repositoryName,omitempty"`
	ObjectsCount    int            `json:"objectsCount"`
	SessionID       string         `json:"sessionId,omitempty"`
	HighPriority    bool           `json:"highPriority"`
	ProgressPercent int            `json:"progressPercent"`
}

// ---------------------------------------------------------------------------
// Encryption Passwords — V13 API: /api/v1/encryptionPasswords
// ---------------------------------------------------------------------------

// EncryptionPasswordModel is the response from GET /api/v1/encryptionPasswords/{id}.
type EncryptionPasswordModel struct {
	ID               string `json:"id"`
	Hint             string `json:"hint"`
	ModificationTime string `json:"modificationTime,omitempty"`
	UniqueID         string `json:"uniqueId,omitempty"`
}

// EncryptionPasswordSpec is the request body for POST/PUT encryption passwords.
type EncryptionPasswordSpec struct {
	Password string `json:"password"` // sensitive — never read back from API
	Hint     string `json:"hint"`
	UniqueID string `json:"uniqueId,omitempty"`
}

// ---------------------------------------------------------------------------
// Job Action Specs (start / stop / retry)
// These are used with the PathJobStart / PathJobStop / PathJobRetry endpoints
// and return SessionModel for async polling.
// ---------------------------------------------------------------------------

// JobStartSpec configures an on-demand job start.
type JobStartSpec struct {
	// PerformActiveFull forces a full backup instead of incremental.
	PerformActiveFull bool `json:"performActiveFull"`
	// StartChainedJobs also starts any jobs chained to this job.
	StartChainedJobs bool `json:"startChainedJobs,omitempty"`
}

// JobStopSpec configures a graceful job stop.
type JobStopSpec struct {
	// GracefulStop produces a restore point for already-processed VMs.
	GracefulStop bool `json:"gracefulStop"`
	// CancelChainedJobs also cancels any chained jobs.
	CancelChainedJobs bool `json:"cancelChainedJobs,omitempty"`
}

// JobRetrySpec configures a failed job retry.
type JobRetrySpec struct {
	// StartChainedJobs also starts any chained jobs after retry.
	StartChainedJobs bool `json:"startChainedJobs,omitempty"`
}

package models

// ---------------------------------------------------------------------------
// Jobs — V13 API: /api/v1/jobs
// Polymorphic: discriminator "type" → Backup | BackupCopy | VSphereReplica | ...
// ---------------------------------------------------------------------------

// JobModel is the base response model for all job types.
type JobModel struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       EJobType `json:"type"`
	IsDisabled bool     `json:"isDisabled"`
}

// BackupJobModel is a vSphere backup job (extends JobModel).
type BackupJobModel struct {
	JobModel
	Description     string                      `json:"description,omitempty"`
	IsHighPriority  bool                        `json:"isHighPriority,omitempty"`
	VirtualMachines *BackupJobVirtualMachines    `json:"virtualMachines,omitempty"`
	Storage         *BackupJobStorage            `json:"storage,omitempty"`
	GuestProcessing *BackupJobGuestProcessing    `json:"guestProcessing,omitempty"`
	Schedule        *BackupSchedule              `json:"schedule,omitempty"`
}

// ---------------------------------------------------------------------------
// Job Spec — used for POST/PUT (create/update)
// ---------------------------------------------------------------------------

// JobSpec is the base request body for job CRUD.
type JobSpec struct {
	Name string   `json:"name"`
	Type EJobType `json:"type"`
}

// BackupJobSpec creates/updates a vSphere backup job.
type BackupJobSpec struct {
	JobSpec
	Description     string                      `json:"description,omitempty"`
	IsHighPriority  bool                        `json:"isHighPriority,omitempty"`
	VirtualMachines *BackupJobVirtualMachinesSpec `json:"virtualMachines,omitempty"`
	Storage         *BackupJobStorage            `json:"storage,omitempty"`
	GuestProcessing *BackupJobGuestProcessing    `json:"guestProcessing,omitempty"`
	Schedule        *BackupSchedule              `json:"schedule,omitempty"`
}

// ---------------------------------------------------------------------------
// Virtual Machines — includes/excludes for backup scope
// ---------------------------------------------------------------------------

// BackupJobVirtualMachines defines which VMs are backed up (response).
type BackupJobVirtualMachines struct {
	Includes []VirtualMachineInclude `json:"includes"`
	Excludes *BackupJobExclusions    `json:"excludes,omitempty"`
}

// BackupJobVirtualMachinesSpec defines which VMs are backed up (request).
type BackupJobVirtualMachinesSpec struct {
	Includes []VirtualMachineInclude    `json:"includes"`
	Excludes *BackupJobExclusionsSpec   `json:"excludes,omitempty"`
}

// VirtualMachineInclude is a VM/container reference in the backup scope.
type VirtualMachineInclude struct {
	InventoryObject *InventoryObjectRef `json:"inventoryObject,omitempty"`
}

// InventoryObjectRef references a vSphere inventory object.
type InventoryObjectRef struct {
	Type     string `json:"type,omitempty"`
	HostName string `json:"hostName,omitempty"`
	Name     string `json:"name,omitempty"`
	ObjectID string `json:"objectId,omitempty"`
}

// BackupJobExclusions defines VM exclusions from the backup (response).
type BackupJobExclusions struct {
	VMs     []InventoryObjectRef `json:"vms,omitempty"`
	Disks   []DiskExclusion      `json:"disks,omitempty"`
	Templates bool               `json:"templates,omitempty"`
}

// BackupJobExclusionsSpec defines VM exclusions from the backup (request).
type BackupJobExclusionsSpec struct {
	VMs     []InventoryObjectRef `json:"vms,omitempty"`
	Disks   []DiskExclusion      `json:"disks,omitempty"`
	Templates bool               `json:"templates,omitempty"`
}

// DiskExclusion excludes specific disks from a VM backup.
type DiskExclusion struct {
	VMObjectID string   `json:"vmObjectId,omitempty"`
	DiskNames  []string `json:"diskNames,omitempty"`
}

// ---------------------------------------------------------------------------
// Storage Settings
// ---------------------------------------------------------------------------

// BackupJobStorage configures where and how backups are stored.
type BackupJobStorage struct {
	BackupRepositoryID string                   `json:"backupRepositoryId"`
	BackupProxies      *BackupProxiesSettings   `json:"backupProxies,omitempty"`
	RetentionPolicy    *RetentionPolicySettings `json:"retentionPolicy,omitempty"`
	GFSPolicy          *GFSPolicySettings       `json:"gfsPolicy,omitempty"`
	AdvancedSettings   *BackupAdvancedSettings  `json:"advancedSettings,omitempty"`
}

// BackupProxiesSettings configures proxy selection for the job.
type BackupProxiesSettings struct {
	AutoSelectEnabled bool     `json:"autoSelectEnabled"`
	ProxyIDs          []string `json:"proxyIds,omitempty"`
}

// RetentionPolicySettings configures how long backups are kept.
type RetentionPolicySettings struct {
	Type     ERetentionPolicyType `json:"type"`
	Quantity int                  `json:"quantity"`
}

// GFSPolicySettings configures Grandfather-Father-Son retention.
type GFSPolicySettings struct {
	IsEnabled bool                    `json:"isEnabled"`
	Weekly    *GFSWeeklySettings      `json:"weekly,omitempty"`
	Monthly   *GFSMonthlySettings     `json:"monthly,omitempty"`
	Yearly    *GFSYearlySettings      `json:"yearly,omitempty"`
}

// GFSWeeklySettings for GFS weekly retention.
type GFSWeeklySettings struct {
	IsEnabled        bool   `json:"isEnabled"`
	KeepForWeeks     int    `json:"keepForNumberOfWeeks,omitempty"`
	DesiredDayOfWeek string `json:"desiredDayOfWeek,omitempty"`
}

// GFSMonthlySettings for GFS monthly retention.
type GFSMonthlySettings struct {
	IsEnabled         bool   `json:"isEnabled"`
	KeepForMonths     int    `json:"keepForNumberOfMonths,omitempty"`
	DesiredDayOfMonth string `json:"desiredDayOfMonth,omitempty"`
}

// GFSYearlySettings for GFS yearly retention.
type GFSYearlySettings struct {
	IsEnabled      bool   `json:"isEnabled"`
	KeepForYears   int    `json:"keepForNumberOfYears,omitempty"`
	DesiredMonth   string `json:"desiredMonth,omitempty"`
}

// BackupAdvancedSettings for backup job advanced configuration.
type BackupAdvancedSettings struct {
	BackupModeType     EBackupModeType              `json:"backupModeType,omitempty"`
	StorageData        *BackupStorageDataSettings   `json:"storageData,omitempty"`
	Notifications      *NotificationSettings        `json:"notifications,omitempty"`
}

// BackupStorageDataSettings for storage-level settings.
type BackupStorageDataSettings struct {
	CompressionLevel       string `json:"compressionLevel,omitempty"`
	StorageOptimization    string `json:"storageOptimization,omitempty"`
	EnableInlineDedup      bool   `json:"enableInlineDedup,omitempty"`
	EncryptionEnabled      bool   `json:"encryptionEnabled,omitempty"`
	EncryptionPasswordID   string `json:"encryptionPasswordId,omitempty"`
}

// NotificationSettings for job notifications.
type NotificationSettings struct {
	SendSNMPNotification bool   `json:"sendSNMPNotification,omitempty"`
	EmailNotification    *EmailNotificationSettings `json:"emailNotification,omitempty"`
}

// EmailNotificationSettings for email alerts.
type EmailNotificationSettings struct {
	IsEnabled    bool   `json:"isEnabled"`
	Recipients   string `json:"recipients,omitempty"`
	NotifyOnSuccess bool `json:"notifyOnSuccess,omitempty"`
	NotifyOnWarning bool `json:"notifyOnWarning,omitempty"`
	NotifyOnError   bool `json:"notifyOnError,omitempty"`
}

// ---------------------------------------------------------------------------
// Schedule Settings
// ---------------------------------------------------------------------------

// BackupSchedule configures when a job runs.
type BackupSchedule struct {
	RunAutomatically bool                    `json:"runAutomatically"`
	Daily            *ScheduleDaily          `json:"daily,omitempty"`
	Monthly          *ScheduleMonthly        `json:"monthly,omitempty"`
	Periodically     *SchedulePeriodically   `json:"periodically,omitempty"`
	AfterThisJob     *ScheduleAfterJob       `json:"afterThisJob,omitempty"`
	Retry            *ScheduleRetry          `json:"retry,omitempty"`
}

// ScheduleDaily for daily schedule.
type ScheduleDaily struct {
	IsEnabled bool        `json:"isEnabled"`
	LocalTime string      `json:"localTime,omitempty"`
	DailyKind EDailyKinds `json:"dailyKind,omitempty"`
	Days      []string    `json:"days,omitempty"`
}

// ScheduleMonthly for monthly schedule.
type ScheduleMonthly struct {
	IsEnabled      bool   `json:"isEnabled"`
	LocalTime      string `json:"localTime,omitempty"`
	DayOfMonth     int    `json:"dayOfMonth,omitempty"`
	DayNumberInMonth string `json:"dayNumberInMonth,omitempty"`
	DayOfWeek      string `json:"dayOfWeek,omitempty"`
	Months         []string `json:"months,omitempty"`
}

// SchedulePeriodically for interval-based schedule.
type SchedulePeriodically struct {
	IsEnabled       bool   `json:"isEnabled"`
	PeriodicallyKind string `json:"periodicallyKind,omitempty"`
	Frequency       int    `json:"frequency,omitempty"`
}

// ScheduleAfterJob runs this job after another job completes.
type ScheduleAfterJob struct {
	IsEnabled bool   `json:"isEnabled"`
	JobID     string `json:"jobId,omitempty"`
}

// ScheduleRetry configures job retry on failure.
type ScheduleRetry struct {
	IsEnabled    bool `json:"isEnabled"`
	RetryCount   int  `json:"retryCount,omitempty"`
	AwaitMinutes int  `json:"awaitMinutes,omitempty"`
}

// ---------------------------------------------------------------------------
// Guest Processing
// ---------------------------------------------------------------------------

// BackupJobGuestProcessing configures application-aware processing.
type BackupJobGuestProcessing struct {
	AppAwareEnabled      bool   `json:"appAwareEnabled,omitempty"`
	GuestInteractionProxyID string `json:"guestInteractionProxyId,omitempty"`
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
	Password string `json:"password"` // sensitive
	Hint     string `json:"hint"`
	UniqueID string `json:"uniqueId,omitempty"`
}

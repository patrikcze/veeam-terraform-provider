package models

// ---------------------------------------------------------------------------
// Pagination — V13 list endpoints wrap results in this structure.
// ---------------------------------------------------------------------------

// PaginationResult is returned in every paginated list response.
type PaginationResult struct {
	Total int `json:"total"`
	Count int `json:"count"`
	Skip  int `json:"skip,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// PaginatedResponse wraps a list of items with pagination metadata.
// Use with json.RawMessage for polymorphic lists, or type-assert after.
type PaginatedResponse struct {
	Data       []any            `json:"data"`
	Pagination PaginationResult `json:"pagination"`
}

// ---------------------------------------------------------------------------
// Credentials Enums
// ---------------------------------------------------------------------------

// ECredentialsType is the discriminator for credential subtypes.
type ECredentialsType string

const (
	CredentialsTypeStandard ECredentialsType = "Standard"
	CredentialsTypeLinux    ECredentialsType = "Linux"
)

// EAuthenticationType determines how Linux credentials authenticate.
type EAuthenticationType string

const (
	AuthenticationTypePassword   EAuthenticationType = "Password"
	AuthenticationTypePrivateKey EAuthenticationType = "PrivateKey"
)

// ---------------------------------------------------------------------------
// Managed Server Enums
// ---------------------------------------------------------------------------

// EManagedServerType is the discriminator for managed server subtypes.
type EManagedServerType string

const (
	ManagedServerTypeWindowsHost         EManagedServerType = "WindowsHost"
	ManagedServerTypeLinuxHost           EManagedServerType = "LinuxHost"
	ManagedServerTypeViHost              EManagedServerType = "ViHost"
	ManagedServerTypeCloudDirector       EManagedServerType = "CloudDirectorHost"
	ManagedServerTypeHvServer            EManagedServerType = "HvServer"
	ManagedServerTypeHvCluster           EManagedServerType = "HvCluster"
	ManagedServerTypeSCVMM               EManagedServerType = "SCVMM"
	ManagedServerTypeSmbV3Cluster        EManagedServerType = "SmbV3Cluster"
	ManagedServerTypeSmbV3StandaloneHost EManagedServerType = "SmbV3StandaloneHost"
	ManagedServerTypeSmbV3Host           EManagedServerType = ManagedServerTypeSmbV3StandaloneHost
)

// EManagedServersStatus represents the availability status.
type EManagedServersStatus string

const (
	ManagedServerStatusAvailable   EManagedServersStatus = "Available"
	ManagedServerStatusUnavailable EManagedServersStatus = "Unavailable"
)

// ECredentialsStorageType determines how credentials are provided.
type ECredentialsStorageType string

const (
	CredentialsStorageTypePermanent ECredentialsStorageType = "Permanent"
	CredentialsStorageTypeSingleUse ECredentialsStorageType = "SingleUse"
	CredentialsStorageTypeSaved     ECredentialsStorageType = CredentialsStorageTypePermanent
)

// ---------------------------------------------------------------------------
// Repository Enums
// ---------------------------------------------------------------------------

// ERepositoryType is the discriminator for repository subtypes.
type ERepositoryType string

const (
	RepositoryTypeWinLocal      ERepositoryType = "WinLocal"
	RepositoryTypeLinuxLocal    ERepositoryType = "LinuxLocal"
	RepositoryTypeSmb           ERepositoryType = "Smb"
	RepositoryTypeNfs           ERepositoryType = "Nfs"
	RepositoryTypeAzureBlob     ERepositoryType = "AzureBlob"
	RepositoryTypeAmazonS3      ERepositoryType = "AmazonS3"
	RepositoryTypeS3Compatible  ERepositoryType = "S3Compatible"
	RepositoryTypeGoogleCloud   ERepositoryType = "GoogleCloud"
	RepositoryTypeLinuxHardened ERepositoryType = "LinuxHardened"
)

// ---------------------------------------------------------------------------
// Proxy Enums
// ---------------------------------------------------------------------------

// EProxyType is the discriminator for proxy subtypes.
type EProxyType string

const (
	ProxyTypeViProxy   EProxyType = "ViProxy"
	ProxyTypeHvProxy   EProxyType = "HvProxy"
	ProxyTypeFileProxy EProxyType = "FileProxy"
)

// EBackupProxyTransportMode determines how data is transferred.
type EBackupProxyTransportMode string

const (
	TransportModeAuto             EBackupProxyTransportMode = "Auto"
	TransportModeDirectAccess     EBackupProxyTransportMode = "DirectAccess"
	TransportModeVirtualAppliance EBackupProxyTransportMode = "VirtualAppliance"
	TransportModeNetwork          EBackupProxyTransportMode = "Network"
)

// ---------------------------------------------------------------------------
// Job Type Enum — V13 discriminator for polymorphic /api/v1/jobs endpoints.
// Keep this list in sync with Veeam REST API v1.3-rev1 EJobType enum.
// ---------------------------------------------------------------------------

// EJobType is the discriminator for job subtypes (API field: "type").
type EJobType string

const (
	JobTypeUnknown                             EJobType = "Unknown"
	JobTypeVSphereBackup                       EJobType = "VSphereBackup"
	JobTypeBackup                              EJobType = JobTypeVSphereBackup // alias for backwards compat
	JobTypeHyperVBackup                        EJobType = "HyperVBackup"
	JobTypeVSphereReplica                      EJobType = "VSphereReplica"
	JobTypeCloudDirectorBackup                 EJobType = "CloudDirectorBackup"
	JobTypeEntraIDTenantBackup                 EJobType = "EntraIDTenantBackup"
	JobTypeEntraIDAuditLogBackup               EJobType = "EntraIDAuditLogBackup"
	JobTypeFileBackupCopy                      EJobType = "FileBackupCopy"
	JobTypeLegacyBackupCopy                    EJobType = "LegacyBackupCopy"
	JobTypeBackupCopy                          EJobType = "BackupCopy"
	JobTypeWindowsAgentBackup                  EJobType = "WindowsAgentBackup"
	JobTypeLinuxAgentBackup                    EJobType = "LinuxAgentBackup"
	JobTypeFileBackup                          EJobType = "FileBackup"
	JobTypeObjectStorageBackup                 EJobType = "ObjectStorageBackup"
	JobTypeEntraIDTenantBackupCopy             EJobType = "EntraIDTenantBackupCopy"
	JobTypeSureBackupContentScan               EJobType = "SureBackupContentScan"
	JobTypeWindowsAgentBackupWorkstationPolicy EJobType = "WindowsAgentBackupWorkstationPolicy"
	JobTypeLinuxAgentBackupWorkstationPolicy   EJobType = "LinuxAgentBackupWorkstationPolicy"
	JobTypeWindowsAgentBackupServerPolicy      EJobType = "WindowsAgentBackupServerPolicy"
	JobTypeLinuxAgentBackupServerPolicy        EJobType = "LinuxAgentBackupServerPolicy"
)

// ---------------------------------------------------------------------------
// Retention Enums
// ---------------------------------------------------------------------------

// ERetentionPolicyType determines how retention is calculated.
type ERetentionPolicyType string

const (
	RetentionPolicyTypeRestorePoints ERetentionPolicyType = "RestorePoints"
	RetentionPolicyTypeDays          ERetentionPolicyType = "Days"
)

// ---------------------------------------------------------------------------
// Backup Mode Enum — updated to match API v1.3-rev1 EBackupModeType.
// ---------------------------------------------------------------------------

// EBackupModeType determines the backup method used to create restore points.
type EBackupModeType string

const (
	// BackupModeFull performs a full backup on every run.
	BackupModeFull EBackupModeType = "Full"
	// BackupModeIncremental stores only changed blocks since last backup.
	BackupModeIncremental EBackupModeType = "Incremental"
	// BackupModeReverseIncremental transforms the most recent restore point into full.
	BackupModeReverseIncremental EBackupModeType = "ReverseIncremental"
	// BackupModeTransform merges old restore points with incrementals (forward forever-incremental).
	BackupModeTransform EBackupModeType = "Transform"
	// BackupModeTransformForeverIncremental (forever-forward incremental with merge).
	BackupModeTransformForeverIncremental EBackupModeType = "TransformForeverIncremental"
)

// ---------------------------------------------------------------------------
// Schedule Enums
// ---------------------------------------------------------------------------

// EDailyKinds determines which days a daily schedule runs.
type EDailyKinds string

const (
	DailyKindsEveryday    EDailyKinds = "Everyday"
	DailyKindsWeekdays    EDailyKinds = "Weekdays"
	DailyKindsSelectedDay EDailyKinds = "SelectedDays"
)

// EDayOfWeek represents a day of the week used in schedules and GFS policies.
type EDayOfWeek string

const (
	DayMonday    EDayOfWeek = "Monday"
	DayTuesday   EDayOfWeek = "Tuesday"
	DayWednesday EDayOfWeek = "Wednesday"
	DayThursday  EDayOfWeek = "Thursday"
	DayFriday    EDayOfWeek = "Friday"
	DaySaturday  EDayOfWeek = "Saturday"
	DaySunday    EDayOfWeek = "Sunday"
)

// EMonth represents a calendar month, used in GFS yearly retention.
type EMonth string

const (
	MonthJanuary   EMonth = "January"
	MonthFebruary  EMonth = "February"
	MonthMarch     EMonth = "March"
	MonthApril     EMonth = "April"
	MonthMay       EMonth = "May"
	MonthJune      EMonth = "June"
	MonthJuly      EMonth = "July"
	MonthAugust    EMonth = "August"
	MonthSeptember EMonth = "September"
	MonthOctober   EMonth = "October"
	MonthNovember  EMonth = "November"
	MonthDecember  EMonth = "December"
)

// ESennightOfMonth represents the week-of-month for monthly GFS retention.
type ESennightOfMonth string

const (
	SennightFirst  ESennightOfMonth = "First"
	SennightSecond ESennightOfMonth = "Second"
	SennightThird  ESennightOfMonth = "Third"
	SennightFourth ESennightOfMonth = "Fourth"
	SennightFifth  ESennightOfMonth = "Fifth"
	SennightLast   ESennightOfMonth = "Last"
)

// EPeriodicallyKinds defines time units for periodic job scheduling.
type EPeriodicallyKinds string

const (
	PeriodicallyHours   EPeriodicallyKinds = "Hours"
	PeriodicallyMinutes EPeriodicallyKinds = "Minutes"
	PeriodicallySeconds EPeriodicallyKinds = "Seconds"
	PeriodicallyDays    EPeriodicallyKinds = "Days"
)

// ---------------------------------------------------------------------------
// Storage Data Enums
// ---------------------------------------------------------------------------

// ECompressionLevel controls the compression applied to backup data.
type ECompressionLevel string

const (
	CompressionAuto          ECompressionLevel = "Auto"
	CompressionNone          ECompressionLevel = "None"
	CompressionDedupFriendly ECompressionLevel = "DedupFriendly"
	CompressionOptimal       ECompressionLevel = "Optimal"
	CompressionHigh          ECompressionLevel = "High"
	CompressionExtreme       ECompressionLevel = "Extreme"
)

// EStorageOptimization controls block size used for backup storage.
type EStorageOptimization string

const (
	StorageOptimization256KB EStorageOptimization = "256KB"
	StorageOptimization512KB EStorageOptimization = "512KB"
	StorageOptimization1MB   EStorageOptimization = "1MB"
	StorageOptimization4MB   EStorageOptimization = "4MB"
)

// ---------------------------------------------------------------------------
// Inventory / VM Enums
// ---------------------------------------------------------------------------

// EInventoryPlatformType is the discriminator for inventory object types.
type EInventoryPlatformType string

const (
	InventoryPlatformVSphere       EInventoryPlatformType = "VSphere"
	InventoryPlatformHyperV        EInventoryPlatformType = "HyperV"
	InventoryPlatformCloudDirector EInventoryPlatformType = "CloudDirector"
	InventoryPlatformAgent         EInventoryPlatformType = "Agent"
)

// EVmwareInventoryType is the type of a VMware vSphere inventory object.
type EVmwareInventoryType string

const (
	VmwareTypeUnknown          EVmwareInventoryType = "Unknown"
	VmwareTypeVirtualMachine   EVmwareInventoryType = "VirtualMachine"
	VmwareTypeVCenterServer    EVmwareInventoryType = "vCenterServer"
	VmwareTypeDatacenter       EVmwareInventoryType = "Datacenter"
	VmwareTypeCluster          EVmwareInventoryType = "Cluster"
	VmwareTypeHost             EVmwareInventoryType = "Host"
	VmwareTypeResourcePool     EVmwareInventoryType = "ResourcePool"
	VmwareTypeFolder           EVmwareInventoryType = "Folder"
	VmwareTypeDatastore        EVmwareInventoryType = "Datastore"
	VmwareTypeDatastoreCluster EVmwareInventoryType = "DatastoreCluster"
	VmwareTypeStoragePolicy    EVmwareInventoryType = "StoragePolicy"
	VmwareTypeTemplate         EVmwareInventoryType = "Template"
	VmwareTypeTag              EVmwareInventoryType = "Tag"
	VmwareTypeCategory         EVmwareInventoryType = "Category"
	VmwareTypeVirtualApp       EVmwareInventoryType = "VirtualApp"
)

// EAgentInventoryObjectType is the type of an agent-managed inventory object.
type EAgentInventoryObjectType string

const (
	AgentTypeProtectionGroup EAgentInventoryObjectType = "ProtectionGroup"
	AgentTypeWindowsComputer EAgentInventoryObjectType = "WindowsComputer"
	AgentTypeLinuxComputer   EAgentInventoryObjectType = "LinuxComputer"
	AgentTypeWindowsCluster  EAgentInventoryObjectType = "WindowsCluster"
	AgentTypeContainer       EAgentInventoryObjectType = "Container"
	AgentTypeGroup           EAgentInventoryObjectType = "Group"
	AgentTypeOrgUnit         EAgentInventoryObjectType = "OrganizationUnit"
	AgentTypeDomain          EAgentInventoryObjectType = "Domain"
)

// EAgentBackupJobMode defines the scope for agent backup jobs.
type EAgentBackupJobMode string

const (
	AgentBackupModeEntireComputer EAgentBackupJobMode = "EntireComputer"
	AgentBackupModeVolumes        EAgentBackupJobMode = "Volumes"
	AgentBackupModeFileLevel      EAgentBackupJobMode = "FileLevel"
)

// EJobAgentType defines the protected computer type for agent policies.
type EJobAgentType string

const (
	JobAgentTypeWorkstation     EJobAgentType = "Workstation"
	JobAgentTypeServer          EJobAgentType = "Server"
	JobAgentTypeFailoverCluster EJobAgentType = "FailoverCluster"
)

// ---------------------------------------------------------------------------
// Job Status Enums (read-only, returned in job state queries)
// ---------------------------------------------------------------------------

// EJobStatus represents the current operational status of a job.
type EJobStatus string

const (
	JobStatusRunning  EJobStatus = "Running"
	JobStatusInactive EJobStatus = "Inactive"
	JobStatusDisabled EJobStatus = "Disabled"
	JobStatusEnabled  EJobStatus = "Enabled"
	JobStatusStopping EJobStatus = "Stopping"
	JobStatusStopped  EJobStatus = "Stopped"
	JobStatusStarting EJobStatus = "Starting"
)

// EJobWorkload categorises the workload type that a job protects.
type EJobWorkload string

const (
	JobWorkloadApplication EJobWorkload = "Application"
	JobWorkloadCloudVM     EJobWorkload = "CloudVm"
	JobWorkloadFile        EJobWorkload = "File"
	JobWorkloadServer      EJobWorkload = "Server"
	JobWorkloadWorkstation EJobWorkload = "Workstation"
	JobWorkloadVM          EJobWorkload = "Vm"
)

// ---------------------------------------------------------------------------
// Protection Group Enums
// ---------------------------------------------------------------------------

// EProtectionGroupType is the discriminator for protection group subtypes.
type EProtectionGroupType string

const (
	ProtectionGroupTypeIndividualComputers EProtectionGroupType = "IndividualComputers"
	ProtectionGroupTypeADObjects           EProtectionGroupType = "ADObjects"
	ProtectionGroupTypeCSVFile             EProtectionGroupType = "CSVFile"
	ProtectionGroupTypePreInstalledAgents  EProtectionGroupType = "PreInstalledAgents"
	ProtectionGroupTypeCloudMachines       EProtectionGroupType = "CloudMachines"
)

// EIndividualComputerConnectionType determines how a computer authenticates in an IndividualComputers protection group.
type EIndividualComputerConnectionType string

const (
	IndividualComputerConnectionTypePermanentCredentials EIndividualComputerConnectionType = "PermanentCredentials"
	IndividualComputerConnectionTypeSingleUseCredentials EIndividualComputerConnectionType = "SingleUseCredentials"
	IndividualComputerConnectionTypeCertificate          EIndividualComputerConnectionType = "Certificate"
)

// EProtectionGroupCloudAccountType determines cloud account subtype for CloudMachines protection groups.
type EProtectionGroupCloudAccountType string

const (
	ProtectionGroupCloudAccountTypeAWS   EProtectionGroupCloudAccountType = "AWS"
	ProtectionGroupCloudAccountTypeAzure EProtectionGroupCloudAccountType = "Azure"
)

// ECloudMachinesObjectType is the discriminator for cloud object selectors.
type ECloudMachinesObjectType string

const (
	CloudMachinesObjectTypeMachine ECloudMachinesObjectType = "Machine"
	CloudMachinesObjectTypeRegion  ECloudMachinesObjectType = "Region"
	CloudMachinesObjectTypeTag     ECloudMachinesObjectType = "Tag"
)

// ---------------------------------------------------------------------------
// Session Enums
// ---------------------------------------------------------------------------

// ESessionState represents the state of an async session.
type ESessionState string

const (
	SessionStateStopped_ ESessionState = "Stopped"
	SessionStateStarting ESessionState = "Starting"
	SessionStateStopping ESessionState = "Stopping"
	SessionStateWorking_ ESessionState = "Working"
	SessionStatePausing  ESessionState = "Pausing"
	SessionStateResuming ESessionState = "Resuming"
	SessionStateIdle     ESessionState = "Idle"
)

// ESessionResult represents the result of a completed session.
type ESessionResult string

const (
	SessionResultSuccess_ ESessionResult = "Success"
	SessionResultWarning_ ESessionResult = "Warning"
	SessionResultFailed_  ESessionResult = "Failed"
	SessionResultNone_    ESessionResult = "None"
)

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
	Data       []interface{}    `json:"data"`
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
// Job Enums
// ---------------------------------------------------------------------------

// EJobType is the discriminator for job subtypes.
type EJobType string

const (
	JobTypeUnknown            EJobType = "Unknown"
	JobTypeVSphereBackup      EJobType = "VSphereBackup"
	JobTypeBackup             EJobType = JobTypeVSphereBackup
	JobTypeHyperVBackup       EJobType = "HyperVBackup"
	JobTypeVSphereReplica     EJobType = "VSphereReplica"
	JobTypeBackupCopy         EJobType = "BackupCopy"
	JobTypeWindowsAgentBackup EJobType = "WindowsAgentBackup"
	JobTypeLinuxAgentBackup   EJobType = "LinuxAgentBackup"
	JobTypeFileBackup         EJobType = "FileBackup"
)

// ERetentionPolicyType determines how retention is calculated.
type ERetentionPolicyType string

const (
	RetentionPolicyTypeRestorePoints ERetentionPolicyType = "RestorePoints"
	RetentionPolicyTypeDays          ERetentionPolicyType = "Days"
)

// EBackupModeType determines the backup mode.
type EBackupModeType string

const (
	BackupModeIncremental        EBackupModeType = "Incremental"
	BackupModeReverseIncremental EBackupModeType = "ReverseIncremental"
)

// EDailyKinds determines which days a daily schedule runs.
type EDailyKinds string

const (
	DailyKindsEveryday    EDailyKinds = "Everyday"
	DailyKindsWeekdays    EDailyKinds = "Weekdays"
	DailyKindsSelectedDay EDailyKinds = "SelectedDays"
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

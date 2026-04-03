// Package client provides the Veeam V13 REST API client.
//
// VERSIONING: All API paths and the API version header are centralized in this file.
// When Veeam releases a new REST API version (e.g., V14), update ONLY this file
// to adapt the entire provider to the new API contract.
package client

// ---------------------------------------------------------------------------
// API Version
// ---------------------------------------------------------------------------

// APIVersionHeader is the HTTP header name required by all V13+ API requests.
const APIVersionHeader = "x-api-version"

// NOTE: APIVersion constant is defined in client.go as "1.3-rev1".
// It is referenced here for documentation but kept in client.go to avoid
// a circular dependency during authentication (client.go uses it directly).

// ---------------------------------------------------------------------------
// Authentication
// ---------------------------------------------------------------------------

// PathOAuth2Token is the OAuth2 token endpoint for password grant and refresh.
// Content-Type: application/x-www-form-urlencoded
const PathOAuth2Token = "/api/oauth2/token"

// ---------------------------------------------------------------------------
// Credentials
// ---------------------------------------------------------------------------

const (
	// PathCredentials is the base path for credential CRUD operations.
	// GET (list), POST (create)
	PathCredentials = "/api/v1/credentials"

	// PathCredentialByID returns the path for a specific credential.
	// GET, PUT, DELETE
	// Usage: fmt.Sprintf(PathCredentialByID, id)
	PathCredentialByID = "/api/v1/credentials/%s"

	// PathCredentialChangePassword changes a credential's password.
	// POST with {"password": "newpass"}
	PathCredentialChangePassword = "/api/v1/credentials/%s/changepassword"
)

// ---------------------------------------------------------------------------
// Backup Infrastructure — Managed Servers
// ---------------------------------------------------------------------------

const (
	PathManagedServers      = "/api/v1/backupInfrastructure/managedServers"
	PathManagedServerByID   = "/api/v1/backupInfrastructure/managedServers/%s"
	PathManagedServerRescan = "/api/v1/backupInfrastructure/managedServers/%s/rescan"
)

// ---------------------------------------------------------------------------
// Backup Infrastructure — Repositories
// ---------------------------------------------------------------------------

const (
	PathRepositories    = "/api/v1/backupInfrastructure/repositories"
	PathRepositoryByID  = "/api/v1/backupInfrastructure/repositories/%s"
	PathRepositoryState = "/api/v1/backupInfrastructure/repositories/states"
)

// ---------------------------------------------------------------------------
// Backup Infrastructure — Proxies
// ---------------------------------------------------------------------------

const (
	PathProxies   = "/api/v1/backupInfrastructure/proxies"
	PathProxyByID = "/api/v1/backupInfrastructure/proxies/%s"
)

// ---------------------------------------------------------------------------
// Backup Infrastructure — Scale-Out Repositories
// ---------------------------------------------------------------------------

const (
	PathScaleOutRepositories   = "/api/v1/backupInfrastructure/scaleOutRepositories"
	PathScaleOutRepositoryByID = "/api/v1/backupInfrastructure/scaleOutRepositories/%s"
	PathScaleOutEnableSealed   = "/api/v1/backupInfrastructure/scaleOutRepositories/%s/enableSealedMode"
	PathScaleOutDisableSealed  = "/api/v1/backupInfrastructure/scaleOutRepositories/%s/disableSealedMode"
	PathScaleOutEnableMaint    = "/api/v1/backupInfrastructure/scaleOutRepositories/%s/enableMaintenanceMode"
	PathScaleOutDisableMaint   = "/api/v1/backupInfrastructure/scaleOutRepositories/%s/disableMaintenanceMode"
)

// ---------------------------------------------------------------------------
// Cloud Credentials
// ---------------------------------------------------------------------------

const (
	PathCloudCredentials             = "/api/v1/cloudCredentials"
	PathCloudCredentialByID          = "/api/v1/cloudCredentials/%s"
	PathCloudCredentialChangeSecret  = "/api/v1/cloudCredentials/%s/changeSecretKey"
	PathCloudCredentialChangeAccount = "/api/v1/cloudCredentials/%s/changeAccount"
	PathCloudCredentialChangeCert    = "/api/v1/cloudCredentials/%s/changeCertificate"
)

// ---------------------------------------------------------------------------
// Jobs
// ---------------------------------------------------------------------------

const (
	PathJobs       = "/api/v1/jobs"
	PathJobByID    = "/api/v1/jobs/%s"
	PathJobStart   = "/api/v1/jobs/%s/start"
	PathJobStop    = "/api/v1/jobs/%s/stop"
	PathJobEnable  = "/api/v1/jobs/%s/enable"
	PathJobDisable = "/api/v1/jobs/%s/disable"
	PathJobRetry   = "/api/v1/jobs/%s/retry"
	PathJobClone   = "/api/v1/jobs/%s/clone"
)

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

const (
	PathSessions    = "/api/v1/sessions"
	PathSessionByID = "/api/v1/sessions/%s"
)

// ---------------------------------------------------------------------------
// Connection
// ---------------------------------------------------------------------------

const (
	PathConnectionCertificate = "/api/v1/connectionCertificate"
)

// ---------------------------------------------------------------------------
// Configuration Backup
// ---------------------------------------------------------------------------

const (
	PathConfigurationBackup      = "/api/v1/configBackup"
	PathConfigurationBackupStart = "/api/v1/configBackup/backup"
)

// ---------------------------------------------------------------------------
// Encryption Passwords
// ---------------------------------------------------------------------------

const (
	PathEncryptionPasswords    = "/api/v1/encryptionPasswords"
	PathEncryptionPasswordByID = "/api/v1/encryptionPasswords/%s"
)

// ---------------------------------------------------------------------------
// Agent Management — Protection Groups
// ---------------------------------------------------------------------------

const (
	PathProtectionGroups       = "/api/v1/agents/protectionGroups"
	PathProtectionGroupByID    = "/api/v1/agents/protectionGroups/%s"
	PathProtectionGroupRescan  = "/api/v1/agents/protectionGroups/%s/rescan"
	PathProtectionGroupEnable  = "/api/v1/agents/protectionGroups/%s/enable"
	PathProtectionGroupDisable = "/api/v1/agents/protectionGroups/%s/disable"
)

// ---------------------------------------------------------------------------
// Agent Management — Protected Computers
// ---------------------------------------------------------------------------

const (
	PathProtectedComputers = "/api/v1/agents/protectedComputers"
)

// ---------------------------------------------------------------------------
// Backups
// ---------------------------------------------------------------------------

const (
	PathBackups     = "/api/v1/backups"
	PathBackupByID  = "/api/v1/backups/%s"
	PathBackupFiles = "/api/v1/backups/%s/backupFiles"
)

// ---------------------------------------------------------------------------
// Restore Points
// ---------------------------------------------------------------------------

const (
	PathRestorePoints             = "/api/v1/restorePoints"
	PathRestorePointByID          = "/api/v1/restorePoints/%s"
	PathBackupObjectRestorePoints = "/api/v1/backupObjects/%s/restorePoints"
)

// ---------------------------------------------------------------------------
// WAN Accelerators
// ---------------------------------------------------------------------------

const (
	PathWanAccelerators    = "/api/v1/backupInfrastructure/wanAccelerators"
	PathWanAcceleratorByID = "/api/v1/backupInfrastructure/wanAccelerators/%s"
)

// ---------------------------------------------------------------------------
// Server / License / Job States
// ---------------------------------------------------------------------------

const (
	PathServerInfo       = "/api/v1/serverInfo"
	PathLicense          = "/api/v1/license"
	PathLicenseSockets   = "/api/v1/license/sockets"
	PathLicenseInstances = "/api/v1/license/instances"
	PathLicenseCapacity  = "/api/v1/license/capacity"
	PathJobStates        = "/api/v1/jobs/states"
)

// ---------------------------------------------------------------------------
// General Options
// ---------------------------------------------------------------------------

const (
	PathGeneralOptions = "/api/v1/generalOptions"
)

// ---------------------------------------------------------------------------
// Email Settings
// ---------------------------------------------------------------------------

const (
	PathEmailSettings            = "/api/v1/generalOptions/emailSettings"
	PathEmailSettingsTestMessage = "/api/v1/generalOptions/emailSettings/testMessage"
)

// ---------------------------------------------------------------------------
// Notification Settings
// ---------------------------------------------------------------------------

const (
	PathNotificationSettings = "/api/v1/generalOptions/notifications"
)

// ---------------------------------------------------------------------------
// Traffic Rules
// ---------------------------------------------------------------------------

const (
	PathTrafficRules = "/api/v1/trafficRules"
)

// ---------------------------------------------------------------------------
// Security Settings
// ---------------------------------------------------------------------------

const (
	PathSecuritySettings = "/api/v1/security/settings"
)

// ---------------------------------------------------------------------------
// KMS Servers
// ---------------------------------------------------------------------------

const (
	PathKMSServers          = "/api/v1/kmsServers"
	PathKMSServerByID       = "/api/v1/kmsServers/%s"
	PathKMSServerChangeCert = "/api/v1/kmsServers/%s/changeCertificate"
)

// ---------------------------------------------------------------------------
// Security Users
// ---------------------------------------------------------------------------

const (
	PathSecurityUsers     = "/api/v1/security/users"
	PathSecurityUserByID  = "/api/v1/security/users/%s"
	PathSecurityUserRoles = "/api/v1/security/users/%s/roles"
)

// ---------------------------------------------------------------------------
// AD Domains
// ---------------------------------------------------------------------------

const (
	PathADDomains    = "/api/v1/adDomains"
	PathADDomainByID = "/api/v1/adDomains/%s"
)

// ---------------------------------------------------------------------------
// Security Roles
// ---------------------------------------------------------------------------

const (
	PathSecurityRoles    = "/api/v1/security/roles"
	PathSecurityRoleByID = "/api/v1/security/roles/%s"
)

// ---------------------------------------------------------------------------
// Backup Objects
// ---------------------------------------------------------------------------

const (
	PathBackupObjects    = "/api/v1/backupObjects"
	PathBackupObjectByID = "/api/v1/backupObjects/%s"
)

// ---------------------------------------------------------------------------
// Replicas
// ---------------------------------------------------------------------------

const (
	PathReplicas    = "/api/v1/replicas"
	PathReplicaByID = "/api/v1/replicas/%s"
)

// ---------------------------------------------------------------------------
// Replica Points
// ---------------------------------------------------------------------------

const (
	PathReplicaPoints    = "/api/v1/replicaPoints"
	PathReplicaPointByID = "/api/v1/replicaPoints/%s"
)

// ---------------------------------------------------------------------------
// Proxy States
// ---------------------------------------------------------------------------

const (
	PathProxyStates = "/api/v1/backupInfrastructure/proxies/states"
)

// ---------------------------------------------------------------------------
// Services
// ---------------------------------------------------------------------------

const (
	PathServices = "/api/v1/services"
)

// ---------------------------------------------------------------------------
// Server Time
// ---------------------------------------------------------------------------

const (
	PathServerTime = "/api/v1/serverTime"
)

// ---------------------------------------------------------------------------
// Server Certificate
// ---------------------------------------------------------------------------

const (
	PathServerCertificate = "/api/v1/serverCertificate"
)

// ---------------------------------------------------------------------------
// Task Sessions
// ---------------------------------------------------------------------------

const (
	PathTaskSessions    = "/api/v1/taskSessions"
	PathTaskSessionByID = "/api/v1/taskSessions/%s"
)

// ---------------------------------------------------------------------------
// Security Analyzer
// ---------------------------------------------------------------------------

const (
	PathSecurityAnalyzerBestPractices = "/api/v1/securityAnalyzer/bestPractices"
	PathSecurityAnalyzerLastRun       = "/api/v1/securityAnalyzer/lastRun"
)

// ---------------------------------------------------------------------------
// Malware Detection Events
// ---------------------------------------------------------------------------

const (
	PathMalwareEvents = "/api/v1/malwareDetection/events"
)

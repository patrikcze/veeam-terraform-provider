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

// NOTE: APIVersion constant is defined in client.go as "1.3-rev0".
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
	PathScaleOutRepositories    = "/api/v1/backupInfrastructure/scaleOutRepositories"
	PathScaleOutRepositoryByID  = "/api/v1/backupInfrastructure/scaleOutRepositories/%s"
)

// ---------------------------------------------------------------------------
// Jobs
// ---------------------------------------------------------------------------

const (
	PathJobs        = "/api/v1/jobs"
	PathJobByID     = "/api/v1/jobs/%s"
	PathJobStart    = "/api/v1/jobs/%s/start"
	PathJobStop     = "/api/v1/jobs/%s/stop"
	PathJobEnable   = "/api/v1/jobs/%s/enable"
	PathJobDisable  = "/api/v1/jobs/%s/disable"
	PathJobRetry    = "/api/v1/jobs/%s/retry"
	PathJobClone    = "/api/v1/jobs/%s/clone"
)

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

const (
	PathSessions    = "/api/v1/sessions"
	PathSessionByID = "/api/v1/sessions/%s"
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
	PathProtectionGroups      = "/api/v1/agents/protectionGroups"
	PathProtectionGroupByID   = "/api/v1/agents/protectionGroups/%s"
	PathProtectionGroupRescan = "/api/v1/agents/protectionGroups/%s/rescan"
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
	PathBackups    = "/api/v1/backups"
	PathBackupByID = "/api/v1/backups/%s"
)

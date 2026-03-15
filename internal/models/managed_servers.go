package models

// ---------------------------------------------------------------------------
// Managed Servers — V13 API: /api/v1/backupInfrastructure/managedServers
// Polymorphic: discriminator "type" → WindowsHost | LinuxHost | ViHost | ...
// ---------------------------------------------------------------------------

// ManagedServerModel is the base response model for managed servers.
type ManagedServerModel struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Type        EManagedServerType    `json:"type"`
	Status      EManagedServersStatus `json:"status,omitempty"`
}

// ViHostModel is a vSphere host managed server (extends ManagedServerModel).
type ViHostModel struct {
	ManagedServerModel
}

// WindowsHostModel is a Windows host managed server.
type WindowsHostModel struct {
	ManagedServerModel
}

// LinuxHostModel is a Linux host managed server.
type LinuxHostModel struct {
	ManagedServerModel
}

// ---------------------------------------------------------------------------
// Managed Server Spec — used for POST/PUT (create/update)
// ---------------------------------------------------------------------------

// ManagedServerSpec is the base request body for managed server CRUD.
type ManagedServerSpec struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Type        EManagedServerType `json:"type"`
}

// ViHostSpec adds vSphere-specific fields.
type ViHostSpec struct {
	ManagedServerSpec
	CredentialsID         string `json:"credentialsId,omitempty"`
	Port                  int    `json:"port,omitempty"`
	CertificateThumbprint string `json:"certificateThumbprint,omitempty"`
}

// WindowsHostSpec adds Windows-specific fields.
type WindowsHostSpec struct {
	ManagedServerSpec
	CredentialsStorageType ECredentialsStorageType `json:"credentialsStorageType,omitempty"`
	CredentialsID          string                  `json:"credentialsId,omitempty"`
	NetworkSettings        *ManagedHostPortsModel  `json:"networkSettings,omitempty"`
}

// LinuxHostSpec adds Linux-specific fields.
type LinuxHostSpec struct {
	ManagedServerSpec
	CredentialsStorageType ECredentialsStorageType `json:"credentialsStorageType"`
	CredentialsID          string                  `json:"credentialsId,omitempty"`
	SingleUseCredentials   *LinuxCredentialsSpec   `json:"singleUseCredentials,omitempty"`
	SSHSettings            *LinuxHostSSHSettings   `json:"sshSettings,omitempty"`
	SSHFingerprint         string                  `json:"sshFingerprint"`
}

// ---------------------------------------------------------------------------
// Nested models for managed servers
// ---------------------------------------------------------------------------

// ManagedHostPortsModel defines network port settings for Windows hosts.
type ManagedHostPortsModel struct {
	PortNumber int `json:"portNumber,omitempty"`
}

// LinuxHostSSHSettings defines SSH connection settings for Linux hosts.
type LinuxHostSSHSettings struct {
	SSHPort int `json:"SSHPort,omitempty"`
}

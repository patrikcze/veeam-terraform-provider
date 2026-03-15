package models

// ---------------------------------------------------------------------------
// Repositories — V13 API: /api/v1/backupInfrastructure/repositories
// Polymorphic: discriminator "type" → WinLocal | LinuxLocal | Nfs | Smb | ...
// ---------------------------------------------------------------------------

// RepositoryModel is the base response model for all repository types.
type RepositoryModel struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	UniqueID    string          `json:"uniqueId,omitempty"`
	Type        ERepositoryType `json:"type"`
}

// WindowsLocalStorageModel is a Windows local repository.
type WindowsLocalStorageModel struct {
	RepositoryModel
	HostID      string                          `json:"hostId,omitempty"`
	Repository  *WindowsLocalRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings            `json:"mountServer,omitempty"`
}

// LinuxLocalStorageModel is a Linux local repository.
type LinuxLocalStorageModel struct {
	RepositoryModel
	HostID      string                        `json:"hostId,omitempty"`
	Repository  *LinuxLocalRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings          `json:"mountServer,omitempty"`
}

// NfsStorageModel is an NFS share repository.
type NfsStorageModel struct {
	RepositoryModel
	Share       *NfsShareSettings          `json:"share,omitempty"`
	Repository  *NetworkRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings       `json:"mountServer,omitempty"`
}

// SmbStorageModel is an SMB share repository.
type SmbStorageModel struct {
	RepositoryModel
	Share       *SmbShareSettings          `json:"share,omitempty"`
	Repository  *NetworkRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings       `json:"mountServer,omitempty"`
}

// ---------------------------------------------------------------------------
// Repository Spec — used for POST/PUT (create/update)
// ---------------------------------------------------------------------------

// RepositorySpec is the base request body for repository CRUD.
type RepositorySpec struct {
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	UniqueID     string          `json:"uniqueId,omitempty"`
	ImportBackup bool            `json:"importBackup,omitempty"`
	ImportIndex  bool            `json:"importIndex,omitempty"`
	Type         ERepositoryType `json:"type"`
}

// WindowsLocalStorageSpec creates/updates a Windows local repository.
type WindowsLocalStorageSpec struct {
	RepositorySpec
	HostID      string                          `json:"hostId,omitempty"`
	Repository  *WindowsLocalRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings            `json:"mountServer,omitempty"`
}

// LinuxLocalStorageSpec creates/updates a Linux local repository.
type LinuxLocalStorageSpec struct {
	RepositorySpec
	HostID      string                        `json:"hostId,omitempty"`
	Repository  *LinuxLocalRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings          `json:"mountServer,omitempty"`
}

// NfsStorageSpec creates/updates an NFS repository.
type NfsStorageSpec struct {
	RepositorySpec
	Share       *NfsShareSettings          `json:"share,omitempty"`
	Repository  *NetworkRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings       `json:"mountServer,omitempty"`
}

// SmbStorageSpec creates/updates an SMB repository.
type SmbStorageSpec struct {
	RepositorySpec
	Share       *SmbShareSettings          `json:"share,omitempty"`
	Repository  *NetworkRepositorySettings `json:"repository,omitempty"`
	MountServer *MountServerSettings       `json:"mountServer,omitempty"`
}

// ---------------------------------------------------------------------------
// Repository Settings — nested configuration models
// ---------------------------------------------------------------------------

// WindowsLocalRepositorySettings configures a Windows local path repository.
type WindowsLocalRepositorySettings struct {
	Path             string                      `json:"path"`
	MaxTaskCount     int                         `json:"maxTaskCount,omitempty"`
	ReadWriteRate    int                         `json:"readWriteRate,omitempty"`
	AdvancedSettings *RepositoryAdvancedSettings `json:"advancedSettings,omitempty"`
}

// LinuxLocalRepositorySettings configures a Linux local path repository.
type LinuxLocalRepositorySettings struct {
	Path             string                      `json:"path"`
	MaxTaskCount     int                         `json:"maxTaskCount,omitempty"`
	ReadWriteRate    int                         `json:"readWriteRate,omitempty"`
	AdvancedSettings *RepositoryAdvancedSettings `json:"advancedSettings,omitempty"`
}

// NetworkRepositorySettings configures NFS/SMB network repositories.
type NetworkRepositorySettings struct {
	Path             string                      `json:"path,omitempty"`
	MaxTaskCount     int                         `json:"maxTaskCount,omitempty"`
	ReadWriteRate    int                         `json:"readWriteRate,omitempty"`
	AdvancedSettings *RepositoryAdvancedSettings `json:"advancedSettings,omitempty"`
}

// RepositoryAdvancedSettings holds advanced repository configuration.
type RepositoryAdvancedSettings struct {
	AlignDataBlocks       bool `json:"alignDataBlocks,omitempty"`
	DecompressBeforeStore bool `json:"decompressBeforeStore,omitempty"`
	RotatedDrives         bool `json:"rotatedDrives,omitempty"`
	PerVMBackup           bool `json:"perVMBackup,omitempty"`
}

// MountServerSettings configures the mount server for a repository.
type MountServerSettings struct {
	MountServerID    string `json:"mountServerId,omitempty"`
	WriteCacheFolder string `json:"writeCacheFolder,omitempty"`
	VPowerNFSEnabled bool   `json:"vPowerNFSEnabled,omitempty"`
}

// NfsShareSettings configures NFS share access.
type NfsShareSettings struct {
	SharePath string `json:"sharePath"`
}

// SmbShareSettings configures SMB share access.
type SmbShareSettings struct {
	SharePath     string `json:"sharePath"`
	CredentialsID string `json:"credentialsId,omitempty"`
}

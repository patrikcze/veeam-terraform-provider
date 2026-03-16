package models

// ConfigurationBackupSpec configures backup server configuration backup settings.
type ConfigurationBackupSpec struct {
	Enabled              bool   `json:"enabled"`
	RepositoryID         string `json:"repositoryId,omitempty"`
	RestorePointsToKeep  int    `json:"restorePointsToKeep,omitempty"`
	EncryptionEnabled    bool   `json:"encryptionEnabled,omitempty"`
	EncryptionPasswordID string `json:"encryptionPasswordId,omitempty"`
	NotificationEnabled  bool   `json:"notificationEnabled,omitempty"`
}

// ConfigurationBackupModel is returned by GET /api/v1/configBackup.
type ConfigurationBackupModel struct {
	Enabled              bool   `json:"enabled"`
	RepositoryID         string `json:"repositoryId,omitempty"`
	RestorePointsToKeep  int    `json:"restorePointsToKeep,omitempty"`
	EncryptionEnabled    bool   `json:"encryptionEnabled,omitempty"`
	EncryptionPasswordID string `json:"encryptionPasswordId,omitempty"`
}

// ConfigurationBackupSessionModel is returned when triggering backup.
type ConfigurationBackupSessionModel struct {
	ID          string `json:"id"`
	SessionType string `json:"sessionType,omitempty"`
	State       string `json:"state,omitempty"`
	Result      string `json:"result,omitempty"`
}

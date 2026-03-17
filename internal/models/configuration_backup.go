package models

// ConfigurationBackupSpec configures backup server configuration backup settings.
type ConfigurationBackupSpec struct {
	IsEnabled            bool                               `json:"isEnabled"`
	BackupRepositoryID   string                             `json:"backupRepositoryId,omitempty"`
	RestorePointsToKeep  int                                `json:"restorePointsToKeep,omitempty"`
	Notifications        *ConfigurationBackupNotifications  `json:"notifications,omitempty"`
	Schedule             *ConfigurationBackupSchedule       `json:"schedule,omitempty"`
	LastSuccessfulBackup *ConfigurationBackupLastSuccessful `json:"lastSuccessfulBackup,omitempty"`
	Encryption           *ConfigurationBackupEncryption     `json:"encryption,omitempty"`
}

// ConfigurationBackupModel is returned by GET /api/v1/configBackup.
type ConfigurationBackupModel struct {
	IsEnabled            bool                               `json:"isEnabled"`
	BackupRepositoryID   string                             `json:"backupRepositoryId,omitempty"`
	RestorePointsToKeep  int                                `json:"restorePointsToKeep,omitempty"`
	Notifications        *ConfigurationBackupNotifications  `json:"notifications,omitempty"`
	Schedule             *ConfigurationBackupSchedule       `json:"schedule,omitempty"`
	LastSuccessfulBackup *ConfigurationBackupLastSuccessful `json:"lastSuccessfulBackup,omitempty"`
	Encryption           *ConfigurationBackupEncryption     `json:"encryption,omitempty"`
}

type ConfigurationBackupEncryption struct {
	IsEnabled  bool   `json:"isEnabled"`
	PasswordID string `json:"passwordId,omitempty"`
}

type ConfigurationBackupNotifications struct {
	SNMPEnabled bool `json:"SNMPEnabled"`
}

type ConfigurationBackupSchedule struct {
	IsEnabled bool `json:"isEnabled"`
}

type ConfigurationBackupLastSuccessful struct {
	LastSuccessfulTime string `json:"lastSuccessfulTime,omitempty"`
	SessionID          string `json:"sessionId,omitempty"`
}

// ConfigurationBackupSessionModel is returned when triggering backup.
type ConfigurationBackupSessionModel struct {
	ID          string `json:"id"`
	SessionType string `json:"sessionType,omitempty"`
	State       string `json:"state,omitempty"`
	Result      string `json:"result,omitempty"`
}

package models

// RestorePointModel contains restore point metadata.
type RestorePointModel struct {
	ID           string `json:"id"`
	Name         string `json:"name,omitempty"`
	BackupID     string `json:"backupId,omitempty"`
	CreationTime string `json:"creationTime,omitempty"`
	Type         string `json:"type,omitempty"`
}

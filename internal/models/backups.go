package models

// BackupModel contains backup metadata returned by backup endpoints.
type BackupModel struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	JobID   string `json:"jobId,omitempty"`
	JobName string `json:"jobName,omitempty"`
}

// BackupFileModel contains backup file metadata.
type BackupFileModel struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Size int64  `json:"size,omitempty"`
}

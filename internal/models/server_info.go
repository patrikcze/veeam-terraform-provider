package models

// ServerInfoModel contains backup server installation and version details.
type ServerInfoModel struct {
	InstallationID string `json:"installationId,omitempty"`
	ServerName     string `json:"serverName,omitempty"`
	BuildNumber    string `json:"buildNumber,omitempty"`
	Version        string `json:"version,omitempty"`
}

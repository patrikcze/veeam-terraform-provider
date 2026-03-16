package models

// CloudCredentialSpec is used for create/update operations on /api/v1/cloudCredentials.
type CloudCredentialSpec struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Type           string `json:"type"`
	AccountName    string `json:"accountName,omitempty"`
	SecretKey      string `json:"secretKey,omitempty"`
	TenantID       string `json:"tenantId,omitempty"`
	ApplicationID  string `json:"applicationId,omitempty"`
	ApplicationKey string `json:"applicationKey,omitempty"`
	ProjectID      string `json:"projectId,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// CloudCredentialModel is returned by cloud credential read/list endpoints.
type CloudCredentialModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	AccountName string `json:"accountName,omitempty"`
	TenantID    string `json:"tenantId,omitempty"`
	ProjectID   string `json:"projectId,omitempty"`
}

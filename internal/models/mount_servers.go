package models

// MountServerSpec is the request body for creating or updating a mount server.
type MountServerSpec struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	ManagedServerID string `json:"managedServerId"`
	Type            string `json:"type"`
	CredentialsID   string `json:"credentialsId,omitempty"`
}

// MountServerModel is the API response body for a mount server.
type MountServerModel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	ManagedServerID string `json:"managedServerId"`
	Type            string `json:"type"`
	CredentialsID   string `json:"credentialsId,omitempty"`
}

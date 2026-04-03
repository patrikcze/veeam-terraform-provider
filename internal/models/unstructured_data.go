package models

// UnstructuredDataServerSpec is the request body for creating or updating an unstructured data server.
type UnstructuredDataServerSpec struct {
	Name                string `json:"name"`
	Description         string `json:"description,omitempty"`
	Type                string `json:"type"`
	HostName            string `json:"hostName"`
	CredentialsID       string `json:"credentialsId,omitempty"`
	AccessCredentialsID string `json:"accessCredentialsId,omitempty"`
}

// UnstructuredDataServerModel is the API response body for an unstructured data server.
type UnstructuredDataServerModel struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description,omitempty"`
	Type                string `json:"type"`
	HostName            string `json:"hostName"`
	CredentialsID       string `json:"credentialsId,omitempty"`
	AccessCredentialsID string `json:"accessCredentialsId,omitempty"`
}

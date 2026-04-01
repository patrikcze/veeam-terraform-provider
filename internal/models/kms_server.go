package models

// KMSServerSpec is the request body for creating or updating a KMS server.
type KMSServerSpec struct {
	Name                  string `json:"name"`
	Description           string `json:"description,omitempty"`
	HostName              string `json:"hostName"`
	Port                  int64  `json:"port,omitempty"`
	CertificateThumbprint string `json:"certificateThumbprint,omitempty"`
}

// KMSServerModel is the API response body for a KMS server.
type KMSServerModel struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	HostName              string `json:"hostName"`
	Port                  int64  `json:"port"`
	CertificateThumbprint string `json:"certificateThumbprint"`
}

package models

// EntraIDTenantSpec is the request body for creating or updating an Entra ID tenant.
type EntraIDTenantSpec struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	TenantID      string `json:"tenantId"`
	CredentialsID string `json:"credentialsId"`
}

// EntraIDTenantModel is the API response body for an Entra ID tenant.
type EntraIDTenantModel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	TenantID      string `json:"tenantId"`
	CredentialsID string `json:"credentialsId"`
}

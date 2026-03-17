package models

// CloudCredentialSpec is used for create/update operations on /api/v1/cloudCredentials.
type CloudCredentialSpec struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Type           string `json:"type"`
	AccessKey      string `json:"accessKey,omitempty"`
	Account        string `json:"account,omitempty"`
	SharedKey      string `json:"sharedKey,omitempty"`
	ConnectionName string `json:"connectionName,omitempty"`
	CreationMode   string `json:"creationMode,omitempty"`
	ExistingAccount *AzureComputeCredentialsExistingAccountSpec `json:"existingAccount,omitempty"`
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
	AccessKey   string `json:"accessKey,omitempty"`
	Account     string `json:"account,omitempty"`
	ConnectionName string `json:"connectionName,omitempty"`
	AccountName string `json:"accountName,omitempty"`
	TenantID    string `json:"tenantId,omitempty"`
	ApplicationID string `json:"applicationId,omitempty"`
	ProjectID   string `json:"projectId,omitempty"`
}

type AzureComputeCredentialsExistingAccountSpec struct {
	Deployment   AzureComputeCloudCredentialsDeploymentModel `json:"deployment"`
	Subscription AzureComputeCloudCredentialsSubscriptionSpec `json:"subscription"`
}

type AzureComputeCloudCredentialsDeploymentModel struct {
	DeploymentType string `json:"deploymentType"`
	Region         string `json:"region,omitempty"`
}

type AzureComputeCloudCredentialsSubscriptionSpec struct {
	TenantID      string `json:"tenantId"`
	ApplicationID string `json:"applicationId"`
	Secret        string `json:"secret,omitempty"`
}

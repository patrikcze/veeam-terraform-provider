package models

// RecoveryTokenSpec is the request body for creating or updating a recovery token.
type RecoveryTokenSpec struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	ManagedServerID string `json:"managedServerId"`
}

// RecoveryTokenModel is the API response body for a recovery token.
// TokenValue is only populated on creation — it is never returned on subsequent reads.
type RecoveryTokenModel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	ManagedServerID string `json:"managedServerId"`
	TokenValue      string `json:"tokenValue,omitempty"`
}

package models

// LicenseModel contains basic installed license details.
type LicenseModel struct {
	Type       string `json:"type,omitempty"`
	Status     string `json:"status,omitempty"`
	LicensedTo string `json:"licensedTo,omitempty"`
	Expiration string `json:"expirationDate,omitempty"`
}

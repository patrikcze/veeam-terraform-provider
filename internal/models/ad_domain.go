package models

// ADDomainSpec is the request body for adding an AD domain.
type ADDomainSpec struct {
	Name        string `json:"name"`
	UserName    string `json:"userName"`
	Password    string `json:"password"`
	Description string `json:"description,omitempty"`
}

// ADDomainModel is the API response body for an AD domain.
type ADDomainModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	UserName    string `json:"userName"`
	Description string `json:"description"`
}

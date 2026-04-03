package models

// GlobalVMExclusionSpec is the request body for creating a global VM exclusion.
type GlobalVMExclusionSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	HostName    string `json:"hostName,omitempty"`
	ObjectID    string `json:"objectId,omitempty"`
	Description string `json:"description,omitempty"`
}

// GlobalVMExclusionModel is the API response body for a global VM exclusion.
type GlobalVMExclusionModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	HostName    string `json:"hostName,omitempty"`
	ObjectID    string `json:"objectId,omitempty"`
	Description string `json:"description,omitempty"`
}

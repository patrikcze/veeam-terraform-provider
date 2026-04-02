package models

// SecurityUserSpec is the request body for creating a security user.
type SecurityUserSpec struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	Description string `json:"description,omitempty"`
}

// SecurityUserModel is the API response body for a security user.
type SecurityUserModel struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	Description string `json:"description"`
}

// SecurityUserRoleSpec is the request body for assigning a role to a user.
type SecurityUserRoleSpec struct {
	RoleName string `json:"roleName"`
}

// SecurityUserRoleModel is the API response body for a user's role.
type SecurityUserRoleModel struct {
	RoleName string `json:"roleName"`
}

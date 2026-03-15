package models

// ---------------------------------------------------------------------------
// Credentials — V13 API: /api/v1/credentials
// Polymorphic: discriminator "type" → Standard | Linux
// ---------------------------------------------------------------------------

// CredentialsModel is the base response model for GET /api/v1/credentials/{id}.
// Common fields across all credential types.
type CredentialsModel struct {
	ID           string           `json:"id"`
	Username     string           `json:"username"`
	Description  string           `json:"description"`
	Type         ECredentialsType `json:"type"`
	CreationTime string           `json:"creationTime,omitempty"`
}

// StandardCredentialsModel extends CredentialsModel for Windows/domain credentials.
type StandardCredentialsModel struct {
	CredentialsModel
	UniqueID string `json:"uniqueId,omitempty"`
}

// LinuxCredentialsModel extends CredentialsModel for Linux SSH credentials.
type LinuxCredentialsModel struct {
	CredentialsModel
	UniqueID           string              `json:"uniqueId,omitempty"`
	SSHPort            int                 `json:"SSHPort,omitempty"`
	ElevateToRoot      bool                `json:"elevateToRoot,omitempty"`
	AddToSudoers       bool                `json:"addToSudoers,omitempty"`
	UseSu              bool                `json:"useSu,omitempty"`
	AuthenticationType EAuthenticationType `json:"authenticationType"`
}

// ---------------------------------------------------------------------------
// Credentials Spec — used for POST/PUT (create/update)
// ---------------------------------------------------------------------------

// CredentialsSpec is the base request body for creating/updating credentials.
type CredentialsSpec struct {
	Username    string           `json:"username"`
	Password    string           `json:"password,omitempty"` // sensitive, json:"-" in logging
	Description string           `json:"description,omitempty"`
	Type        ECredentialsType `json:"type"`
}

// StandardCredentialsSpec extends CredentialsSpec for Standard credentials.
type StandardCredentialsSpec struct {
	CredentialsSpec
	UniqueID string `json:"uniqueId,omitempty"`
}

// LinuxCredentialsSpec extends CredentialsSpec for Linux credentials.
type LinuxCredentialsSpec struct {
	CredentialsSpec
	UniqueID           string              `json:"uniqueId,omitempty"`
	SSHPort            int                 `json:"SSHPort,omitempty"`
	ElevateToRoot      bool                `json:"elevateToRoot,omitempty"`
	AddToSudoers       bool                `json:"addToSudoers,omitempty"`
	UseSu              bool                `json:"useSu,omitempty"`
	PrivateKey         string              `json:"privateKey,omitempty"`  // sensitive
	Passphrase         string              `json:"passphrase,omitempty"` // sensitive
	RootPassword       string              `json:"rootPassword,omitempty"` // sensitive
	AuthenticationType EAuthenticationType `json:"authenticationType"`
}

// CredentialChangePasswordSpec is the body for POST /api/v1/credentials/{id}/changepassword.
type CredentialChangePasswordSpec struct {
	Password string `json:"password"`
}

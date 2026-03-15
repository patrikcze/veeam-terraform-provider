package models

// ---------------------------------------------------------------------------
// Protection Groups — V13 API: /api/v1/agents/protectionGroups
// Polymorphic: discriminator "type" → IndividualComputers | ADObjects | CloudMachines | ...
// ---------------------------------------------------------------------------

// ProtectionGroupModel is the base response model for protection groups.
type ProtectionGroupModel struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Type        EProtectionGroupType `json:"type"`
}

// IndividualComputersProtectionGroupModel manages specific computers.
type IndividualComputersProtectionGroupModel struct {
	ProtectionGroupModel
	Computers []ProtectionGroupComputer `json:"computers,omitempty"`
	Options   *ProtectionGroupOptions   `json:"options,omitempty"`
}

// CloudMachinesProtectionGroupModel manages cloud VM instances.
type CloudMachinesProtectionGroupModel struct {
	ProtectionGroupModel
	CloudAccount  *CloudAccountRef        `json:"cloudAccount,omitempty"`
	CloudMachines []CloudMachineRef       `json:"cloudMachines,omitempty"`
	Options       *ProtectionGroupOptions `json:"options,omitempty"`
}

// ---------------------------------------------------------------------------
// Protection Group Spec — used for POST/PUT (create/update)
// ---------------------------------------------------------------------------

// ProtectionGroupSpec is the base request body for protection group CRUD.
type ProtectionGroupSpec struct {
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Tag         string               `json:"tag,omitempty"`
	Type        EProtectionGroupType `json:"type"`
}

// IndividualComputersProtectionGroupSpec creates/updates individual computer groups.
type IndividualComputersProtectionGroupSpec struct {
	ProtectionGroupSpec
	Computers []ProtectionGroupComputer `json:"computers,omitempty"`
	Options   *ProtectionGroupOptions   `json:"options,omitempty"`
}

// CloudMachinesProtectionGroupSpec creates/updates cloud machine groups.
type CloudMachinesProtectionGroupSpec struct {
	ProtectionGroupSpec
	CloudAccount  *CloudAccountRef        `json:"cloudAccount,omitempty"`
	CloudMachines []CloudMachineRef       `json:"cloudMachines,omitempty"`
	Options       *ProtectionGroupOptions `json:"options,omitempty"`
}

// ---------------------------------------------------------------------------
// Nested models for protection groups
// ---------------------------------------------------------------------------

// ProtectionGroupComputer is a computer in an IndividualComputers group.
type ProtectionGroupComputer struct {
	HostName      string `json:"hostName"`
	CredentialsID string `json:"credentialsId,omitempty"`
}

// ProtectionGroupOptions configures protection group behavior.
type ProtectionGroupOptions struct {
	AutoDiscoveryEnabled bool `json:"autoDiscoveryEnabled,omitempty"`
}

// CloudAccountRef references a cloud account for CloudMachines groups.
type CloudAccountRef struct {
	AccountID string `json:"accountId,omitempty"`
	Type      string `json:"type,omitempty"`
}

// CloudMachineRef references a specific cloud machine.
type CloudMachineRef struct {
	InstanceID string `json:"instanceId,omitempty"`
	RegionID   string `json:"regionId,omitempty"`
}

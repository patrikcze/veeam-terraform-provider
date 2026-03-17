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
	IsDisabled  bool                 `json:"isDisabled"`
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
	CloudAccount  *CloudMachinesAccount   `json:"cloudAccount,omitempty"`
	CloudMachines []CloudMachineObject    `json:"cloudMachines,omitempty"`
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
	CloudAccount  *CloudMachinesAccount   `json:"cloudAccount,omitempty"`
	CloudMachines []CloudMachineObject    `json:"cloudMachines,omitempty"`
	Options       *ProtectionGroupOptions `json:"options,omitempty"`
}

// ---------------------------------------------------------------------------
// Nested models for protection groups
// ---------------------------------------------------------------------------

// ProtectionGroupComputer is a computer in an IndividualComputers group.
type ProtectionGroupComputer struct {
	HostName             string                            `json:"hostName"`
	ConnectionType       EIndividualComputerConnectionType `json:"connectionType"`
	CredentialsID        string                            `json:"credentialsId,omitempty"`
	SingleUseCredentials *LinuxCredentialsSpec             `json:"singleUseCredentials,omitempty"`
}

// ProtectionGroupOptions configures protection group behavior.
type ProtectionGroupOptions struct {
	DistributionServerID      string   `json:"distributionServerId,omitempty"`
	DistributionRepositoryID  string   `json:"distributionRepositoryId,omitempty"`
	InstallBackupAgent        bool     `json:"installBackupAgent"`
	InstallCBTDriver          bool     `json:"installCBTDriver"`
	InstallApplicationPlugins bool     `json:"installApplicationPlugins"`
	ApplicationPlugins        []string `json:"applicationPlugins,omitempty"`
	UpdateAutomatically       bool     `json:"updateAutomatically"`
	RebootIfRequired          bool     `json:"rebootIfRequired"`
}

// CloudMachinesAccount references a cloud account for CloudMachines groups.
// AWS requires credentialsId, regionType, regionId.
// Azure requires subscriptionId, regionType, regionId.
type CloudMachinesAccount struct {
	AccountType    EProtectionGroupCloudAccountType `json:"accountType"`
	CredentialsID  string                           `json:"credentialsId,omitempty"`
	SubscriptionID string                           `json:"subscriptionId,omitempty"`
	RegionType     string                           `json:"regionType,omitempty"`
	RegionID       string                           `json:"regionId,omitempty"`
	AssignIAMRole  bool                             `json:"assignIamRole,omitempty"`
}

// CloudMachineObject references a cloud object selector in a CloudMachines group.
type CloudMachineObject struct {
	Type     ECloudMachinesObjectType `json:"type"`
	Name     string                   `json:"name,omitempty"`
	ObjectID string                   `json:"objectId,omitempty"`
	Value    string                   `json:"value,omitempty"`
}

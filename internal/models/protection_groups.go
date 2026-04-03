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

// ---------------------------------------------------------------------------
// ADObjects protection group — Active Directory-based computer discovery
// ---------------------------------------------------------------------------

// ADObjectsProtectionGroupSpec creates/updates AD-based protection groups.
type ADObjectsProtectionGroupSpec struct {
	ProtectionGroupSpec
	ActiveDirectory *ADObjectsAccount       `json:"activeDirectory,omitempty"`
	ADObjects       []ADObject              `json:"adObjects,omitempty"`
	Options         *ProtectionGroupOptions `json:"options,omitempty"`
}

// ADObjectsProtectionGroupModel is the response model for ADObjects groups.
type ADObjectsProtectionGroupModel struct {
	ProtectionGroupModel
	ActiveDirectory *ADObjectsAccount       `json:"activeDirectory,omitempty"`
	ADObjects       []ADObject              `json:"adObjects,omitempty"`
	Options         *ProtectionGroupOptions `json:"options,omitempty"`
}

// ADObjectsAccount references an AD domain for ADObjects groups.
type ADObjectsAccount struct {
	DomainID      string `json:"domainId"`
	CredentialsID string `json:"credentialsId,omitempty"`
}

// ADObject is an Active Directory container, OU, or computer object.
type ADObject struct {
	Type     EADObjectType `json:"type"`
	Name     string        `json:"name"`
	ObjectID string        `json:"objectId,omitempty"`
}

// EADObjectType is the discriminator for AD object types.
type EADObjectType string

const (
	ADObjectTypeOU        EADObjectType = "OrganizationalUnit"
	ADObjectTypeContainer EADObjectType = "Container"
	ADObjectTypeComputer  EADObjectType = "Computer"
	ADObjectTypeGroup     EADObjectType = "Group"
)

// ---------------------------------------------------------------------------
// CSVFile protection group — CSV file-based computer list import
// ---------------------------------------------------------------------------

// CSVFileProtectionGroupSpec creates/updates CSV file-based protection groups.
type CSVFileProtectionGroupSpec struct {
	ProtectionGroupSpec
	DelimiterType ECSVDelimiterType       `json:"delimiterType,omitempty"`
	CSVFilePath   string                  `json:"csvFilePath"`
	Options       *ProtectionGroupOptions `json:"options,omitempty"`
}

// CSVFileProtectionGroupModel is the response model for CSVFile groups.
type CSVFileProtectionGroupModel struct {
	ProtectionGroupModel
	DelimiterType ECSVDelimiterType       `json:"delimiterType,omitempty"`
	CSVFilePath   string                  `json:"csvFilePath,omitempty"`
	Options       *ProtectionGroupOptions `json:"options,omitempty"`
}

// ECSVDelimiterType controls the delimiter used in the CSV file.
type ECSVDelimiterType string

const (
	CSVDelimiterComma     ECSVDelimiterType = "Comma"
	CSVDelimiterSemicolon ECSVDelimiterType = "Semicolon"
	CSVDelimiterTab       ECSVDelimiterType = "Tab"
	CSVDelimiterSpace     ECSVDelimiterType = "Space"
)

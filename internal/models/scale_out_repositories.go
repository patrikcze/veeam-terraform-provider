package models

// ---------------------------------------------------------------------------
// Scale-Out Repositories — V13 API: /api/v1/backupInfrastructure/scaleOutRepositories
// Spec follows Veeam REST API v1.3-rev1.
// ---------------------------------------------------------------------------

// PerformanceExtentSpec references an existing repository to add as a performance extent.
type PerformanceExtentSpec struct {
	ID string `json:"id"`
}

// PerformanceTierAdvancedSettings controls SOBR performance tier limits.
type PerformanceTierAdvancedSettings struct {
	UsedCapacityLimitEnabled         bool `json:"usedCapacityLimitEnabled,omitempty"`
	UsedCapacityLimit                int  `json:"usedCapacityLimit,omitempty"`
	DataLocalityEnabled              bool `json:"dataLocalityEnabled,omitempty"`
	PerformanceTierOverfillEnabled   bool `json:"performanceTierOverfillEnabled,omitempty"`
	PerformanceTierOverfillThreshold int  `json:"performanceTierOverfillThreshold,omitempty"`
}

// PerformanceTierSpec defines the performance tier of a SOBR on create/update.
type PerformanceTierSpec struct {
	PerformanceExtents []PerformanceExtentSpec          `json:"performanceExtents"`
	AdvancedSettings   *PerformanceTierAdvancedSettings `json:"advancedSettings,omitempty"`
}

// CapacityTierSpec configures the optional object-storage capacity tier.
type CapacityTierSpec struct {
	IsEnabled         bool   `json:"isEnabled"`
	CopyPolicyEnabled bool   `json:"copyPolicyEnabled,omitempty"`
	MovePolicyEnabled bool   `json:"movePolicyEnabled,omitempty"`
	ObjectStorageID   string `json:"objectStorageId,omitempty"`
}

// EPlacementPolicyType is the data placement policy for a SOBR.
// Controls how Veeam distributes backup data across performance extents.
type EPlacementPolicyType string

const (
	// PlacementPolicyDataLocality places backup chains on the same extent
	// as previous restore points to improve deduplication and restore speed.
	PlacementPolicyDataLocality EPlacementPolicyType = "DataLocality"
	// PlacementPolicyPerformance distributes backup chains across all extents
	// to maximise parallel throughput.
	PlacementPolicyPerformance EPlacementPolicyType = "Performance"
)

// PlacementPolicyModel configures how backup data is distributed across SOBR extents.
// Corresponds to API schema PlacementPolicyModel.
type PlacementPolicyModel struct {
	// Type selects the placement strategy.
	// Allowed values: DataLocality, Performance.
	Type EPlacementPolicyType `json:"type"`
}

// ScaleOutRepositorySpec is the request body for POST/PUT scale-out repository endpoints.
type ScaleOutRepositorySpec struct {
	Name            string                `json:"name"`
	Description     string                `json:"description,omitempty"`
	PerformanceTier PerformanceTierSpec   `json:"performanceTier"`
	CapacityTier    *CapacityTierSpec     `json:"capacityTier,omitempty"`
	PlacementPolicy *PlacementPolicyModel `json:"placementPolicy,omitempty"`
}

// ---------------------------------------------------------------------------
// Response models
// ---------------------------------------------------------------------------

// PerformanceExtentModel is a performance extent in a SOBR GET response.
type PerformanceExtentModel struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// PerformanceTierModel is the performance tier section of a SOBR GET response.
type PerformanceTierModel struct {
	PerformanceExtents []PerformanceExtentModel         `json:"performanceExtents,omitempty"`
	AdvancedSettings   *PerformanceTierAdvancedSettings `json:"advancedSettings,omitempty"`
}

// CapacityTierModel is the capacity tier section of a SOBR GET response.
type CapacityTierModel struct {
	IsEnabled         bool `json:"isEnabled"`
	CopyPolicyEnabled bool `json:"copyPolicyEnabled,omitempty"`
	MovePolicyEnabled bool `json:"movePolicyEnabled,omitempty"`
}

// ScaleOutRepositoryModel is returned by scale-out repository GET endpoints.
type ScaleOutRepositoryModel struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description,omitempty"`
	PerformanceTier *PerformanceTierModel `json:"performanceTier,omitempty"`
	CapacityTier    *CapacityTierModel    `json:"capacityTier,omitempty"`
	PlacementPolicy *PlacementPolicyModel `json:"placementPolicy,omitempty"`
}

// ScaleOutModeSpec is used when toggling sealed/maintenance mode for extents.
type ScaleOutModeSpec struct {
	ExtentIDs []string `json:"extentIds,omitempty"`
}

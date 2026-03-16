package models

// ScaleOutRepositorySpec is used for create/update of scale-out repositories.
type ScaleOutRepositorySpec struct {
	Name                         string   `json:"name"`
	Description                  string   `json:"description,omitempty"`
	CapacityTier                 bool     `json:"capacityTierEnabled,omitempty"`
	PerformanceTierRepositoryIDs []string `json:"performanceTierRepositoryIds,omitempty"`
}

// ScaleOutRepositoryModel is returned by scale-out repository endpoints.
type ScaleOutRepositoryModel struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description,omitempty"`
	IsSealedModeEnabled      bool   `json:"isSealedModeEnabled,omitempty"`
	IsMaintenanceModeEnabled bool   `json:"isMaintenanceModeEnabled,omitempty"`
}

// ScaleOutModeSpec is used when toggling sealed/maintenance mode for extents.
type ScaleOutModeSpec struct {
	ExtentIDs []string `json:"extentIds,omitempty"`
}

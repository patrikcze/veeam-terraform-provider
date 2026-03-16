package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestScaleOutRepository_SyncFromAPI(t *testing.T) {
	resource := &ScaleOutRepository{}
	state := &ScaleOutRepositoryModel{}
	api := &models.ScaleOutRepositoryModel{
		ID:                       "sobr-1",
		Name:                     "SOBR Main",
		Description:              "Main scale out repository",
		IsSealedModeEnabled:      true,
		IsMaintenanceModeEnabled: false,
	}

	resource.syncFromAPI(state, api)

	assert.Equal(t, "SOBR Main", state.Name.ValueString())
	assert.Equal(t, "Main scale out repository", state.Description.ValueString())
	assert.True(t, state.SealedModeEnabled.ValueBool())
	assert.False(t, state.MaintenanceModeEnabled.ValueBool())
}

func TestScaleOutRepository_ModelValues(t *testing.T) {
	data := &ScaleOutRepositoryModel{
		Name:                   types.StringValue("sobr-main"),
		Description:            types.StringValue("desc"),
		CapacityTierEnabled:    types.BoolValue(true),
		MaintenanceModeEnabled: types.BoolValue(false),
		SealedModeEnabled:      types.BoolValue(true),
	}

	assert.Equal(t, "sobr-main", data.Name.ValueString())
	assert.Equal(t, "desc", data.Description.ValueString())
	assert.True(t, data.CapacityTierEnabled.ValueBool())
	assert.False(t, data.MaintenanceModeEnabled.ValueBool())
	assert.True(t, data.SealedModeEnabled.ValueBool())
}

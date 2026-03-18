package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestScaleOutRepository_BuildSpec(t *testing.T) {
	r := &ScaleOutRepository{}
	ctx := context.Background()

	extentList, diags := types.ListValueFrom(ctx, types.StringType, []string{"repo-1", "repo-2"})
	require.False(t, diags.HasError())

	data := &ScaleOutRepositoryModel{
		Name:                 types.StringValue("SOBR-Main"),
		Description:          types.StringValue("Test SOBR"),
		PerformanceExtentIDs: extentList,
		CapacityTierEnabled:  types.BoolValue(false),
	}

	spec, diags := r.buildSpec(ctx, data)
	require.False(t, diags.HasError())
	require.NotNil(t, spec)

	assert.Equal(t, "SOBR-Main", spec.Name)
	assert.Equal(t, "Test SOBR", spec.Description)
	assert.Len(t, spec.PerformanceTier.PerformanceExtents, 2)
	assert.Equal(t, "repo-1", spec.PerformanceTier.PerformanceExtents[0].ID)
	assert.Equal(t, "repo-2", spec.PerformanceTier.PerformanceExtents[1].ID)
	assert.Nil(t, spec.CapacityTier)
}

func TestScaleOutRepository_BuildSpec_CapacityTierEnabled(t *testing.T) {
	r := &ScaleOutRepository{}
	ctx := context.Background()

	extentList, _ := types.ListValueFrom(ctx, types.StringType, []string{"repo-1"})

	data := &ScaleOutRepositoryModel{
		Name:                 types.StringValue("SOBR-Cap"),
		PerformanceExtentIDs: extentList,
		CapacityTierEnabled:  types.BoolValue(true),
	}

	spec, diags := r.buildSpec(ctx, data)
	require.False(t, diags.HasError())
	require.NotNil(t, spec.CapacityTier)
	assert.True(t, spec.CapacityTier.IsEnabled)
}

func TestScaleOutRepository_SyncFromAPI(t *testing.T) {
	r := &ScaleOutRepository{}
	ctx := context.Background()

	state := &ScaleOutRepositoryModel{}
	api := &models.ScaleOutRepositoryModel{
		ID:          "sobr-1",
		Name:        "SOBR Main",
		Description: "Main scale out repository",
		PerformanceTier: &models.PerformanceTierModel{
			PerformanceExtents: []models.PerformanceExtentModel{
				{ID: "repo-1", Name: "Repo One"},
				{ID: "repo-2", Name: "Repo Two"},
			},
		},
		CapacityTier: &models.CapacityTierModel{
			IsEnabled: true,
		},
	}

	diags := r.syncFromAPI(ctx, state, api)
	require.False(t, diags.HasError())

	assert.Equal(t, "SOBR Main", state.Name.ValueString())
	assert.Equal(t, "Main scale out repository", state.Description.ValueString())
	assert.True(t, state.CapacityTierEnabled.ValueBool())

	var ids []string
	diags = state.PerformanceExtentIDs.ElementsAs(ctx, &ids, false)
	require.False(t, diags.HasError())
	assert.Equal(t, []string{"repo-1", "repo-2"}, ids)
}

func TestScaleOutRepository_SyncFromAPI_NoCapacityTier(t *testing.T) {
	r := &ScaleOutRepository{}
	ctx := context.Background()

	state := &ScaleOutRepositoryModel{}
	api := &models.ScaleOutRepositoryModel{
		ID:   "sobr-2",
		Name: "SOBR NoCapacity",
		PerformanceTier: &models.PerformanceTierModel{
			PerformanceExtents: []models.PerformanceExtentModel{
				{ID: "repo-3"},
			},
		},
	}

	diags := r.syncFromAPI(ctx, state, api)
	require.False(t, diags.HasError())
	assert.False(t, state.CapacityTierEnabled.ValueBool())
}

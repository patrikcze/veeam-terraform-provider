package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestProtectionGroupBuildCreateSpec(t *testing.T) {
	resource := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		Name:        types.StringValue("Office-Servers"),
		Description: types.StringValue("Agent deployment group"),
		Type:        types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("srv01.example.local"),
				ConnectionType: types.StringValue("PermanentCredentials"),
				CredentialsID:  types.StringValue("cred-123"),
			},
		},
		Options: []ProtectionGroupOptionsModel{
			{
				DistributionServerID:      types.StringValue("server-123"),
				InstallBackupAgent:        types.BoolValue(true),
				InstallCBTDriver:          types.BoolValue(false),
				InstallApplicationPlugins: types.BoolValue(true),
				ApplicationPlugins:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("MSSQL")}),
				UpdateAutomatically:       types.BoolValue(true),
				RebootIfRequired:          types.BoolValue(false),
			},
		},
	}

	spec := resource.buildCreateSpec(data)
	pg, ok := spec.(*models.IndividualComputersProtectionGroupSpec)
	assert.True(t, ok, "expected *IndividualComputersProtectionGroupSpec")
	assert.Equal(t, models.ProtectionGroupTypeIndividualComputers, pg.Type)
	if assert.Len(t, pg.Computers, 1) {
		assert.Equal(t, models.IndividualComputerConnectionTypePermanentCredentials, pg.Computers[0].ConnectionType)
		assert.Equal(t, "cred-123", pg.Computers[0].CredentialsID)
	}
	if assert.NotNil(t, pg.Options) {
		assert.True(t, pg.Options.InstallBackupAgent)
		assert.Equal(t, "server-123", pg.Options.DistributionServerID)
	}
}

func TestValidateProtectionGroupPlanRequiresConnectionType(t *testing.T) {
	data := &ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:      types.StringValue("srv01"),
				CredentialsID: types.StringValue("cred-1"),
			},
		},
	}

	err := validateProtectionGroupPlan(data)
	assert.Error(t, err)
}

func TestValidateProtectionGroupPlanPermanentRequiresCredential(t *testing.T) {
	data := &ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("srv01"),
				ConnectionType: types.StringValue("PermanentCredentials"),
			},
		},
	}

	err := validateProtectionGroupPlan(data)
	assert.Error(t, err)
}

func TestValidateProtectionGroupPlanInstallBackupAgentDependency(t *testing.T) {
	data := &ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("srv01"),
				ConnectionType: types.StringValue("PermanentCredentials"),
				CredentialsID:  types.StringValue("cred-1"),
			},
		},
		Options: []ProtectionGroupOptionsModel{
			{
				InstallBackupAgent: types.BoolValue(true),
			},
		},
	}

	err := validateProtectionGroupPlan(data)
	assert.Error(t, err)
}

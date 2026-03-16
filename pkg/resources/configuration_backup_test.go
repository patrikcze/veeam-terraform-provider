package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestConfigurationBackup_PutConfig(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &ConfigurationBackup{client: mockClient}

	data := &ConfigurationBackupModel{
		Enabled:              types.BoolValue(true),
		RepositoryID:         types.StringValue("repo-1"),
		RestorePointsToKeep:  types.Int64Value(14),
		EncryptionEnabled:    types.BoolValue(true),
		EncryptionPasswordID: types.StringValue("enc-1"),
	}

	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, nil).Run(func(args mock.Arguments) {
		spec := args.Get(2).(*models.ConfigurationBackupSpec)
		assert.True(t, spec.Enabled)
		assert.Equal(t, "repo-1", spec.RepositoryID)
		assert.Equal(t, 14, spec.RestorePointsToKeep)
		assert.True(t, spec.EncryptionEnabled)
		assert.Equal(t, "enc-1", spec.EncryptionPasswordID)
	}).Return(nil)

	err := resource.putConfig(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_TriggerBackup(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &ConfigurationBackup{client: mockClient}
	data := &ConfigurationBackupModel{}

	mockClient.On("PostJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(3).(*models.ConfigurationBackupSessionModel)
		result.ID = "session-1"
		result.State = "Stopped"
		result.Result = "Success"
	}).Return(nil)

	err := resource.triggerBackup(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, "session-1", data.LastSessionID.ValueString())
	assert.Equal(t, "Stopped", data.LastSessionState.ValueString())
	assert.Equal(t, "Success", data.LastSessionResult.ValueString())
	mockClient.AssertExpectations(t)
}

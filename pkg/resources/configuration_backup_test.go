package resources

import (
	"context"
	"fmt"
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

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"isEnabled":           false,
			"backupRepositoryId":  "repo-old",
			"restorePointsToKeep": float64(7),
			"encryption": map[string]interface{}{
				"isEnabled":  false,
				"passwordId": "enc-old",
			},
			"schedule": map[string]interface{}{
				"period": "daily",
			},
			"notifications": map[string]interface{}{
				"SNMPEnabled": false,
			},
			"lastSuccessfulBackup": map[string]interface{}{
				"id": "session-1",
			},
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, true, payload["isEnabled"])
		assert.Equal(t, "repo-1", payload["backupRepositoryId"])
		assert.Equal(t, 14, payload["restorePointsToKeep"])

		encryption, ok := payload["encryption"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, true, encryption["isEnabled"])
		assert.Equal(t, "enc-1", encryption["passwordId"])

		_, hasSchedule := payload["schedule"]
		_, hasNotifications := payload["notifications"]
		_, hasLastSuccessful := payload["lastSuccessfulBackup"]
		assert.True(t, hasSchedule)
		assert.True(t, hasNotifications)
		assert.True(t, hasLastSuccessful)
	}).Return(nil)

	err := resource.putConfig(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_PutConfig_SNMPFallbackRetry(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &ConfigurationBackup{client: mockClient}

	data := &ConfigurationBackupModel{
		Enabled:              types.BoolValue(true),
		RepositoryID:         types.StringValue("repo-1"),
		RestorePointsToKeep:  types.Int64Value(14),
		EncryptionEnabled:    types.BoolValue(true),
		EncryptionPasswordID: types.StringValue("enc-1"),
	}

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"isEnabled":           true,
			"backupRepositoryId":  "repo-old",
			"restorePointsToKeep": float64(7),
			"encryption": map[string]interface{}{
				"isEnabled":  false,
				"passwordId": "enc-old",
			},
			"schedule": map[string]interface{}{
				"isEnabled": true,
			},
			"notifications": map[string]interface{}{
				"SNMPEnabled": true,
			},
			"lastSuccessfulBackup": map[string]interface{}{
				"sessionId": "session-1",
			},
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		notifications := payload["notifications"].(map[string]interface{})
		assert.Equal(t, true, notifications["SNMPEnabled"])
	}).Return(fmt.Errorf("API request failed (HTTP 400): UnknownError: Specify SNMP settings in General Options.")).Once()

	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		notifications := payload["notifications"].(map[string]interface{})
		assert.Equal(t, false, notifications["SNMPEnabled"])
	}).Return(nil).Once()

	err := resource.putConfig(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_TriggerBackup(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &ConfigurationBackup{client: mockClient}
	data := &ConfigurationBackupModel{}

	mockClient.On("PostJSON", mock.Anything, client.PathConfigurationBackupStart, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
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

func TestEnsureKnownLastSessionFields(t *testing.T) {
	data := &ConfigurationBackupModel{
		LastSessionID:     types.StringUnknown(),
		LastSessionState:  types.StringUnknown(),
		LastSessionResult: types.StringUnknown(),
	}

	ensureKnownLastSessionFields(data)

	assert.True(t, data.LastSessionID.IsNull())
	assert.True(t, data.LastSessionState.IsNull())
	assert.True(t, data.LastSessionResult.IsNull())
}

func TestConfigurationBackup_DeletePayloadClearsEncryption(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &ConfigurationBackup{client: mockClient}

	payload := map[string]interface{}{
		"isEnabled": true,
		"notifications": map[string]interface{}{
			"SNMPEnabled": false,
		},
		"encryption": map[string]interface{}{
			"isEnabled":  true,
			"passwordId": "enc-1",
		},
	}

	setBoolValue(payload, false, "enabled", "isEnabled")
	encryption := ensureNestedConfigMap(payload, "encryption")
	setBoolValue(encryption, false, "isEnabled", "enabled")
	setStringValue(encryption, "", "passwordId", "encryptionPasswordId")

	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, nil).Run(func(args mock.Arguments) {
		updated := args.Get(2).(map[string]interface{})
		assert.Equal(t, false, updated["isEnabled"])
		encryption, ok := updated["encryption"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, false, encryption["isEnabled"])
		assert.Equal(t, "", encryption["passwordId"])
	}).Return(nil).Once()

	err := resource.putConfigurationPayload(context.Background(), payload)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

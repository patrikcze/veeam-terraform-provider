package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudCredentialSpec_RoundTrip(t *testing.T) {
	original := CloudCredentialSpec{
		Name:           "aws-main",
		Description:    "AWS account",
		Type:           "Amazon",
		AccountName:    "AKIA_TEST",
		SecretKey:      "secret",
		TenantID:       "tenant",
		ApplicationID:  "app-id",
		ApplicationKey: "app-key",
		ProjectID:      "project-id",
		ServiceAccount: "service-account",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CloudCredentialSpec
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.AccountName, decoded.AccountName)
}

func TestConfigurationBackupModel_RoundTrip(t *testing.T) {
	original := ConfigurationBackupModel{
		Enabled:              true,
		RepositoryID:         "repo-1",
		RestorePointsToKeep:  14,
		EncryptionEnabled:    true,
		EncryptionPasswordID: "enc-1",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ConfigurationBackupModel
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.True(t, decoded.Enabled)
	assert.Equal(t, "repo-1", decoded.RepositoryID)
	assert.Equal(t, 14, decoded.RestorePointsToKeep)
}

func TestScaleOutRepositoryModel_RoundTrip(t *testing.T) {
	original := ScaleOutRepositoryModel{
		ID:                       "sobr-1",
		Name:                     "SOBR Main",
		Description:              "Main SOBR",
		IsSealedModeEnabled:      true,
		IsMaintenanceModeEnabled: false,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ScaleOutRepositoryModel
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "sobr-1", decoded.ID)
	assert.True(t, decoded.IsSealedModeEnabled)
	assert.False(t, decoded.IsMaintenanceModeEnabled)
}

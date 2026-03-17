package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Credentials
// ---------------------------------------------------------------------------

func TestStandardCredentialsModel_RoundTrip(t *testing.T) {
	original := StandardCredentialsModel{
		CredentialsModel: CredentialsModel{
			ID:          "cred-123",
			Username:    "DOMAIN\\admin",
			Description: "Domain admin",
			Type:        CredentialsTypeStandard,
		},
		UniqueID: "unique-abc",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded StandardCredentialsModel
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "cred-123", decoded.ID)
	assert.Equal(t, "DOMAIN\\admin", decoded.Username)
	assert.Equal(t, ECredentialsType("Standard"), decoded.Type)
	assert.Equal(t, "unique-abc", decoded.UniqueID)
}

func TestLinuxCredentialsSpec_RoundTrip(t *testing.T) {
	original := LinuxCredentialsSpec{
		CredentialsSpec: CredentialsSpec{
			Username: "root",
			Password: "secret",
			Type:     CredentialsTypeLinux,
		},
		SSHPort:            22,
		ElevateToRoot:      true,
		AuthenticationType: AuthenticationTypePrivateKey,
		PrivateKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded LinuxCredentialsSpec
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "root", decoded.Username)
	assert.Equal(t, CredentialsTypeLinux, decoded.Type)
	assert.Equal(t, 22, decoded.SSHPort)
	assert.True(t, decoded.ElevateToRoot)
	assert.Equal(t, AuthenticationTypePrivateKey, decoded.AuthenticationType)
}

func TestCredentialsModel_DiscriminatorType(t *testing.T) {
	jsonData := `{"id":"abc","username":"user","description":"test","type":"Standard"}`

	var model CredentialsModel
	require.NoError(t, json.Unmarshal([]byte(jsonData), &model))

	assert.Equal(t, CredentialsTypeStandard, model.Type)
}

// ---------------------------------------------------------------------------
// Repositories
// ---------------------------------------------------------------------------

func TestWindowsLocalStorageSpec_RoundTrip(t *testing.T) {
	original := WindowsLocalStorageSpec{
		RepositorySpec: RepositorySpec{
			Name:        "WinRepo",
			Description: "Windows local repo",
			Type:        RepositoryTypeWinLocal,
		},
		HostID: "host-123",
		Repository: &WindowsLocalRepositorySettings{
			Path:         "C:\\Backups",
			MaxTaskCount: 4,
		},
		MountServer: &MountServersSettings{
			MountServerSettingsType: "Windows",
			Windows: &MountServerSettings{
				MountServerID:    "host-123",
				WriteCacheFolder: "C:\\Backups",
				VPowerNFSEnabled: false,
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded WindowsLocalStorageSpec
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "WinRepo", decoded.Name)
	assert.Equal(t, RepositoryTypeWinLocal, decoded.Type)
	assert.Equal(t, "host-123", decoded.HostID)
	assert.Equal(t, "C:\\Backups", decoded.Repository.Path)
	assert.Equal(t, 4, decoded.Repository.MaxTaskCount)
	require.NotNil(t, decoded.MountServer)
	require.NotNil(t, decoded.MountServer.Windows)
	assert.False(t, decoded.MountServer.Windows.VPowerNFSEnabled)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	mountServer, ok := raw["mountServer"].(map[string]interface{})
	require.True(t, ok)
	windows, ok := mountServer["windows"].(map[string]interface{})
	require.True(t, ok)
	_, exists := windows["vPowerNFSEnabled"]
	assert.True(t, exists, "serialized payload must include vPowerNFSEnabled even when false")
}

func TestLinuxLocalStorageModel_RoundTrip(t *testing.T) {
	original := LinuxLocalStorageModel{
		RepositoryModel: RepositoryModel{
			ID:   "repo-456",
			Name: "LinuxRepo",
			Type: RepositoryTypeLinuxLocal,
		},
		HostID: "linux-host-1",
		Repository: &LinuxLocalRepositorySettings{
			Path:         "/mnt/backups",
			MaxTaskCount: 2,
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded LinuxLocalStorageModel
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "repo-456", decoded.ID)
	assert.Equal(t, RepositoryTypeLinuxLocal, decoded.Type)
	assert.Equal(t, "/mnt/backups", decoded.Repository.Path)
}

// ---------------------------------------------------------------------------
// Managed Servers
// ---------------------------------------------------------------------------

func TestViHostSpec_RoundTrip(t *testing.T) {
	original := ViHostSpec{
		ManagedServerSpec: ManagedServerSpec{
			Name:        "vcenter.example.com",
			Description: "Main vCenter",
			Type:        ManagedServerTypeViHost,
		},
		CredentialsID:         "cred-456",
		Port:                  443,
		CertificateThumbprint: "AA:BB:CC:DD",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ViHostSpec
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "vcenter.example.com", decoded.Name)
	assert.Equal(t, ManagedServerTypeViHost, decoded.Type)
	assert.Equal(t, 443, decoded.Port)
	assert.Equal(t, "AA:BB:CC:DD", decoded.CertificateThumbprint)
}

// ---------------------------------------------------------------------------
// Proxies
// ---------------------------------------------------------------------------

func TestViProxySpec_RoundTrip(t *testing.T) {
	original := ViProxySpec{
		ProxySpec: ProxySpec{
			Description: "Main vSphere proxy",
			Type:        ProxyTypeViProxy,
		},
		Server: &ProxyServerSettings{
			HostID:        "host-789",
			TransportMode: TransportModeAuto,
			MaxTaskCount:  4,
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ViProxySpec
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, ProxyTypeViProxy, decoded.Type)
	assert.Equal(t, "host-789", decoded.Server.HostID)
	assert.Equal(t, TransportModeAuto, decoded.Server.TransportMode)
}

// ---------------------------------------------------------------------------
// Jobs
// ---------------------------------------------------------------------------

func TestBackupJobSpec_RoundTrip(t *testing.T) {
	original := BackupJobSpec{
		JobSpec: JobSpec{
			Name: "Daily-Backup",
			Type: JobTypeBackup,
		},
		Description:    "Daily VM backup",
		IsHighPriority: true,
		Storage: &BackupJobStorage{
			BackupRepositoryID: "repo-123",
			BackupProxies: &BackupProxiesSettings{
				AutoSelectEnabled: true,
			},
			RetentionPolicy: &RetentionPolicySettings{
				Type:     RetentionPolicyTypeRestorePoints,
				Quantity: 14,
			},
		},
		Schedule: &BackupSchedule{
			RunAutomatically: true,
			Daily: &ScheduleDaily{
				IsEnabled: true,
				LocalTime: "22:00",
				DailyKind: DailyKindsWeekdays,
			},
			Retry: &ScheduleRetry{
				IsEnabled:    true,
				RetryCount:   3,
				AwaitMinutes: 10,
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded BackupJobSpec
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "Daily-Backup", decoded.Name)
	assert.Equal(t, JobTypeBackup, decoded.Type)
	assert.True(t, decoded.IsHighPriority)
	assert.Equal(t, "repo-123", decoded.Storage.BackupRepositoryID)
	assert.True(t, decoded.Storage.BackupProxies.AutoSelectEnabled)
	assert.Equal(t, 14, decoded.Storage.RetentionPolicy.Quantity)
	assert.True(t, decoded.Schedule.RunAutomatically)
	assert.Equal(t, "22:00", decoded.Schedule.Daily.LocalTime)
	assert.Equal(t, 3, decoded.Schedule.Retry.RetryCount)
}

func TestBackupJobModel_RoundTrip(t *testing.T) {
	jsonData := `{
		"id": "job-abc",
		"name": "Test-Job",
		"type": "VSphereBackup",
		"isDisabled": false,
		"description": "Test backup job",
		"virtualMachines": {
			"includes": [
				{"inventoryObject": {"type": "VirtualMachine", "name": "vm-01", "objectId": "obj-1"}}
			]
		}
	}`

	var model BackupJobModel
	require.NoError(t, json.Unmarshal([]byte(jsonData), &model))

	assert.Equal(t, "job-abc", model.ID)
	assert.Equal(t, JobTypeBackup, model.Type)
	assert.False(t, model.IsDisabled)
	require.NotNil(t, model.VirtualMachines)
	require.Len(t, model.VirtualMachines.Includes, 1)
	assert.Equal(t, "vm-01", model.VirtualMachines.Includes[0].InventoryObject.Name)
}

// ---------------------------------------------------------------------------
// Protection Groups
// ---------------------------------------------------------------------------

func TestIndividualComputersProtectionGroupSpec_RoundTrip(t *testing.T) {
	original := IndividualComputersProtectionGroupSpec{
		ProtectionGroupSpec: ProtectionGroupSpec{
			Name:        "Office-Servers",
			Description: "Office server group",
			Type:        ProtectionGroupTypeIndividualComputers,
		},
		Computers: []ProtectionGroupComputer{
			{HostName: "server1.example.com", CredentialsID: "cred-1"},
			{HostName: "server2.example.com", CredentialsID: "cred-2"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded IndividualComputersProtectionGroupSpec
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "Office-Servers", decoded.Name)
	assert.Equal(t, ProtectionGroupTypeIndividualComputers, decoded.Type)
	require.Len(t, decoded.Computers, 2)
	assert.Equal(t, "server1.example.com", decoded.Computers[0].HostName)
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func TestFullSessionModel_RoundTrip(t *testing.T) {
	jsonData := `{
		"id": "sess-123",
		"name": "Backup session",
		"jobId": "job-456",
		"sessionType": "BackupJob",
		"creationTime": "2024-01-01T00:00:00Z",
		"state": "Stopped",
		"result": {"result": "Success", "message": "Completed successfully"},
		"usn": 42
	}`

	var model FullSessionModel
	require.NoError(t, json.Unmarshal([]byte(jsonData), &model))

	assert.Equal(t, "sess-123", model.ID)
	assert.Equal(t, ESessionState("Stopped"), model.State)
	require.NotNil(t, model.Result)
	assert.Equal(t, ESessionResult("Success"), model.Result.Result)
	assert.Equal(t, 42, model.USN)
}

// ---------------------------------------------------------------------------
// Encryption Passwords
// ---------------------------------------------------------------------------

func TestEncryptionPasswordSpec_RoundTrip(t *testing.T) {
	original := EncryptionPasswordSpec{
		Password: "supersecret",
		Hint:     "backup encryption",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"password":"supersecret"`)

	var decoded EncryptionPasswordSpec
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "backup encryption", decoded.Hint)
}

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

func TestPaginationResult_Unmarshal(t *testing.T) {
	jsonData := `{"total": 100, "count": 25, "skip": 0, "limit": 25}`

	var p PaginationResult
	require.NoError(t, json.Unmarshal([]byte(jsonData), &p))

	assert.Equal(t, 100, p.Total)
	assert.Equal(t, 25, p.Count)
	assert.Equal(t, 0, p.Skip)
	assert.Equal(t, 25, p.Limit)
}

// ---------------------------------------------------------------------------
// Cloud Credentials
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Configuration Backup
// ---------------------------------------------------------------------------

func TestConfigurationBackupModel_RoundTrip(t *testing.T) {
	original := ConfigurationBackupModel{
		IsEnabled:           true,
		BackupRepositoryID:  "repo-1",
		RestorePointsToKeep: 14,
		Notifications: &ConfigurationBackupNotifications{
			SNMPEnabled: false,
		},
		Schedule: &ConfigurationBackupSchedule{
			IsEnabled: true,
		},
		LastSuccessfulBackup: &ConfigurationBackupLastSuccessful{
			SessionID: "sess-1",
		},
		Encryption: &ConfigurationBackupEncryption{
			IsEnabled:  true,
			PasswordID: "enc-1",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ConfigurationBackupModel
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.True(t, decoded.IsEnabled)
	assert.Equal(t, "repo-1", decoded.BackupRepositoryID)
	assert.Equal(t, 14, decoded.RestorePointsToKeep)
	require.NotNil(t, decoded.Encryption)
	assert.True(t, decoded.Encryption.IsEnabled)
	assert.Equal(t, "enc-1", decoded.Encryption.PasswordID)
}

// ---------------------------------------------------------------------------
// Scale-Out Repositories
// ---------------------------------------------------------------------------

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

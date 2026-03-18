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

func TestRepository_BuildSpec_WinLocal(t *testing.T) {
	resource := &Repository{}
	data := &RepositoryModel{
		Name:                  types.StringValue("WinRepo"),
		Description:           types.StringValue("Windows local repo"),
		Type:                  types.StringValue("WinLocal"),
		HostID:                types.StringValue("host-123"),
		Path:                  types.StringValue("C:\\Backups"),
		MaxTaskCount:          types.Int64Value(4),
		TaskLimitEnabled:      types.BoolValue(true),
		ReadWriteRate:         types.Int64Null(),
		ReadWriteLimitEnabled: types.BoolNull(),
		SharePath:             types.StringNull(),
		CredentialsID:         types.StringNull(),
	}

	spec := resource.buildSpec(data)

	win, ok := spec.(*models.WindowsLocalStorageSpec)
	assert.True(t, ok, "expected *WindowsLocalStorageSpec")
	assert.Equal(t, models.RepositoryTypeWinLocal, win.Type)
	assert.Equal(t, "host-123", win.HostID)
	assert.Equal(t, "C:\\Backups", win.Repository.Path)
	assert.Equal(t, 4, win.Repository.MaxTaskCount)
	if assert.NotNil(t, win.MountServer) {
		assert.Equal(t, "Windows", win.MountServer.MountServerSettingsType)
		if assert.NotNil(t, win.MountServer.Windows) {
			assert.Equal(t, "host-123", win.MountServer.Windows.MountServerID)
			assert.Equal(t, "C:\\Backups", win.MountServer.Windows.WriteCacheFolder)
			assert.False(t, win.MountServer.Windows.VPowerNFSEnabled)
		}
	}
}

func TestRepository_BuildSpec_LinuxLocal(t *testing.T) {
	resource := &Repository{}
	data := &RepositoryModel{
		Name:          types.StringValue("LinuxRepo"),
		Type:          types.StringValue("LinuxLocal"),
		HostID:        types.StringValue("linux-host-1"),
		Path:          types.StringValue("/mnt/backups"),
		MaxTaskCount:  types.Int64Value(2),
		Description:   types.StringNull(),
		SharePath:     types.StringNull(),
		CredentialsID: types.StringNull(),
	}

	spec := resource.buildSpec(data)

	linux, ok := spec.(*models.LinuxLocalStorageSpec)
	assert.True(t, ok, "expected *LinuxLocalStorageSpec")
	assert.Equal(t, models.RepositoryTypeLinuxLocal, linux.Type)
	assert.Equal(t, "/mnt/backups", linux.Repository.Path)
	if assert.NotNil(t, linux.MountServer) {
		assert.Equal(t, "Linux", linux.MountServer.MountServerSettingsType)
		if assert.NotNil(t, linux.MountServer.Linux) {
			assert.Equal(t, "linux-host-1", linux.MountServer.Linux.MountServerID)
			assert.Equal(t, "/mnt/backups", linux.MountServer.Linux.WriteCacheFolder)
			assert.False(t, linux.MountServer.Linux.VPowerNFSEnabled)
		}
	}
}

func TestRepository_BuildSpec_Smb(t *testing.T) {
	resource := &Repository{}
	data := &RepositoryModel{
		Name:          types.StringValue("SmbRepo"),
		Type:          types.StringValue("Smb"),
		SharePath:     types.StringValue("\\\\server\\share"),
		CredentialsID: types.StringValue("cred-456"),
		MaxTaskCount:  types.Int64Value(3),
		Description:   types.StringNull(),
		HostID:        types.StringNull(),
		Path:          types.StringNull(),
	}

	spec := resource.buildSpec(data)

	smb, ok := spec.(*models.SmbStorageSpec)
	assert.True(t, ok, "expected *SmbStorageSpec")
	assert.Equal(t, models.RepositoryTypeSmb, smb.Type)
	assert.Equal(t, "\\\\server\\share", smb.Share.SharePath)
	assert.Equal(t, "cred-456", smb.Share.CredentialsID)
}

func TestRepository_CreatePayload(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathRepositories, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(3).(*models.RepositoryModel)
		result.ID = "repo-123"
		result.Name = "WinRepo"
		result.Type = models.RepositoryTypeWinLocal
	}).Return(nil)

	var result models.RepositoryModel
	err := mockClient.PostJSON(context.Background(), client.PathRepositories, nil, &result)

	assert.NoError(t, err)
	assert.Equal(t, "repo-123", result.ID)
	mockClient.AssertExpectations(t)
}

func TestRepository_SyncFromAPI_PreservesPlanOnEmptyAPIValues(t *testing.T) {
	resource := &Repository{}
	data := &RepositoryModel{
		Name:        types.StringValue("Planned-Repo"),
		Description: types.StringValue("Planned description"),
		Type:        types.StringValue("WinLocal"),
	}

	api := &models.RepositoryModel{}
	resource.syncFromAPI(data, api)

	assert.Equal(t, "Planned-Repo", data.Name.ValueString())
	assert.Equal(t, "Planned description", data.Description.ValueString())
	assert.Equal(t, "WinLocal", data.Type.ValueString())
}

func TestRepository_SyncFromAPI_UsesAPIValuesWhenPresent(t *testing.T) {
	resource := &Repository{}
	data := &RepositoryModel{}

	api := &models.RepositoryModel{
		Name:        "API-Repo",
		Description: "API description",
		Type:        models.RepositoryTypeLinuxLocal,
	}
	resource.syncFromAPI(data, api)

	assert.Equal(t, "API-Repo", data.Name.ValueString())
	assert.Equal(t, "API description", data.Description.ValueString())
	assert.Equal(t, "LinuxLocal", data.Type.ValueString())
}

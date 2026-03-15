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
		Name:          types.StringValue("WinRepo"),
		Description:   types.StringValue("Windows local repo"),
		Type:          types.StringValue("WinLocal"),
		HostID:        types.StringValue("host-123"),
		Path:          types.StringValue("C:\\Backups"),
		MaxTaskCount:  types.Int64Value(4),
		SharePath:     types.StringNull(),
		CredentialsID: types.StringNull(),
	}

	spec := resource.buildSpec(data)

	win, ok := spec.(*models.WindowsLocalStorageSpec)
	assert.True(t, ok, "expected *WindowsLocalStorageSpec")
	assert.Equal(t, models.RepositoryTypeWinLocal, win.Type)
	assert.Equal(t, "host-123", win.HostID)
	assert.Equal(t, "C:\\Backups", win.Repository.Path)
	assert.Equal(t, 4, win.Repository.MaxTaskCount)
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

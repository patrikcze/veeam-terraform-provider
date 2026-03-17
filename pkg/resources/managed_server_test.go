package resources

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestManagedServerBuildSpecLinuxHost(t *testing.T) {
	resource := &ManagedServer{}
	data := &ManagedServerModel{
		Name:           types.StringValue("linux1.example.local"),
		Description:    types.StringValue("Linux SERVER TERRAFORM"),
		Type:           types.StringValue("LinuxHost"),
		CredentialsID:  types.StringValue("cred-123"),
		SSHFingerprint: types.StringValue("SHA256:test-fingerprint"),
	}

	spec := resource.buildSpec(data)
	linux, ok := spec.(*models.LinuxHostSpec)
	assert.True(t, ok, "expected *LinuxHostSpec")
	assert.Equal(t, models.ManagedServerTypeLinuxHost, linux.Type)
	assert.Equal(t, models.CredentialsStorageTypeSaved, linux.CredentialsStorageType)
	assert.Equal(t, "cred-123", linux.CredentialsID)
	assert.Equal(t, "SHA256:test-fingerprint", linux.SSHFingerprint)
}

func TestShouldRetryManagedServerCreateLinuxCredentialValidation(t *testing.T) {
	data := &ManagedServerModel{Type: types.StringValue("LinuxHost")}
	err := fmt.Errorf("API request failed (HTTP 400): UnknownError: Failed to validate the specified Linux credentials.")

	assert.True(t, shouldRetryManagedServerCreate(data, err))
}

func TestShouldRetryManagedServerCreateNonLinux(t *testing.T) {
	data := &ManagedServerModel{Type: types.StringValue("WindowsHost")}
	err := fmt.Errorf("API request failed (HTTP 400): UnknownError: Failed to validate the specified Linux credentials.")

	assert.False(t, shouldRetryManagedServerCreate(data, err))
}

func TestIsAsyncManagedServerCreateResult(t *testing.T) {
	assert.True(t, isAsyncManagedServerCreateResult(map[string]interface{}{"id": "session-1"}))
	assert.True(t, isAsyncManagedServerCreateResult(map[string]interface{}{"id": "session-2", "type": "Session"}))
	assert.False(t, isAsyncManagedServerCreateResult(map[string]interface{}{"id": "server-1", "type": "LinuxHost"}))
}

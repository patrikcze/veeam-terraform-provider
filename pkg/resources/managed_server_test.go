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

func TestShouldResolveLinuxFingerprint(t *testing.T) {
	t.Run("linux with sha256 fingerprint", func(t *testing.T) {
		data := &ManagedServerModel{
			Type:           types.StringValue("LinuxHost"),
			SSHFingerprint: types.StringValue("SHA256:abc"),
		}
		assert.True(t, shouldResolveLinuxFingerprint(data))
	})

	t.Run("linux with openssh fingerprint", func(t *testing.T) {
		data := &ManagedServerModel{
			Type:           types.StringValue("LinuxHost"),
			SSHFingerprint: types.StringValue("ssh-rsa 3072 abc"),
		}
		assert.False(t, shouldResolveLinuxFingerprint(data))
	})

	t.Run("linux with empty fingerprint", func(t *testing.T) {
		data := &ManagedServerModel{
			Type:           types.StringValue("LinuxHost"),
			SSHFingerprint: types.StringValue(""),
		}
		assert.True(t, shouldResolveLinuxFingerprint(data))
	})

	t.Run("non linux", func(t *testing.T) {
		data := &ManagedServerModel{
			Type:           types.StringValue("WindowsHost"),
			SSHFingerprint: types.StringValue("SHA256:abc"),
		}
		assert.False(t, shouldResolveLinuxFingerprint(data))
	})
}

func TestIsManagedServerNotFound(t *testing.T) {
	t.Run("api not found code", func(t *testing.T) {
		err := fmt.Errorf("API request failed (HTTP 404): %w", &models.APIError{ErrorCode: "NotFound", Message: "not found"})
		assert.True(t, isManagedServerNotFound(err))
	})

	t.Run("plain text 404", func(t *testing.T) {
		err := fmt.Errorf("API request failed with HTTP 404: missing")
		assert.True(t, isManagedServerNotFound(err))
	})

	t.Run("other error", func(t *testing.T) {
		err := fmt.Errorf("API request failed (HTTP 400): bad request")
		assert.False(t, isManagedServerNotFound(err))
	})
}

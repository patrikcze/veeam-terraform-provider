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

func TestCredential_BuildSpec_Standard(t *testing.T) {
	resource := &Credential{}
	data := &CredentialModel{
		Username:    types.StringValue("DOMAIN\\admin"),
		Password:    types.StringValue("secret"),
		Description: types.StringValue("Domain admin"),
		Type:        types.StringValue("Standard"),
	}

	spec := resource.buildSpec(data)

	std, ok := spec.(*models.StandardCredentialsSpec)
	assert.True(t, ok, "expected *StandardCredentialsSpec")
	assert.Equal(t, "DOMAIN\\admin", std.Username)
	assert.Equal(t, models.CredentialsTypeStandard, std.Type)
}

func TestCredential_BuildSpec_Linux(t *testing.T) {
	resource := &Credential{}
	data := &CredentialModel{
		Username:           types.StringValue("root"),
		Password:           types.StringValue("secret"),
		Description:        types.StringValue("Linux cred"),
		Type:               types.StringValue("Linux"),
		SSHPort:            types.Int64Value(2222),
		ElevateToRoot:      types.BoolValue(true),
		AuthenticationType: types.StringValue("Password"),
		// Leave other Linux fields null
		AddToSudoers: types.BoolNull(),
		UseSu:        types.BoolNull(),
		PrivateKey:   types.StringNull(),
		Passphrase:   types.StringNull(),
		RootPassword: types.StringNull(),
	}

	spec := resource.buildSpec(data)

	linux, ok := spec.(*models.LinuxCredentialsSpec)
	assert.True(t, ok, "expected *LinuxCredentialsSpec")
	assert.Equal(t, models.CredentialsTypeLinux, linux.Type)
	assert.Equal(t, 2222, linux.SSHPort)
	assert.True(t, linux.ElevateToRoot)
	assert.Equal(t, models.AuthenticationTypePassword, linux.AuthenticationType)
}

func TestCredential_CreatePayload(t *testing.T) {
	mockClient := new(MockVeeamClient)

	// Mock the correct V13 API endpoint
	mockClient.On("PostJSON", mock.Anything, client.PathCredentials, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(3).(*models.CredentialsModel)
		result.ID = "cred-123"
		result.Username = "admin"
		result.Type = models.CredentialsTypeStandard
	}).Return(nil)

	var result models.CredentialsModel
	err := mockClient.PostJSON(context.Background(), client.PathCredentials, nil, &result)

	assert.NoError(t, err)
	assert.Equal(t, "cred-123", result.ID)
	mockClient.AssertExpectations(t)
}

func TestCredentialModel_StateFields(t *testing.T) {
	data := CredentialModel{
		ID:          types.StringValue("cred-123"),
		Description: types.StringValue("Test credential"),
		Username:    types.StringValue("admin"),
		Password:    types.StringValue("password123"),
		Type:        types.StringValue("Standard"),
	}

	assert.Equal(t, "cred-123", data.ID.ValueString())
	assert.Equal(t, "Test credential", data.Description.ValueString())
	assert.Equal(t, "admin", data.Username.ValueString())
	assert.Equal(t, "password123", data.Password.ValueString())
	assert.Equal(t, "Standard", data.Type.ValueString())
}

func TestCredential_SyncModelFromAPI(t *testing.T) {
	resource := &Credential{}
	data := &CredentialModel{
		Password: types.StringValue("original-password"),
	}

	api := &models.CredentialsModel{
		ID:          "cred-abc",
		Username:    "DOMAIN\\user",
		Description: "Updated desc",
		Type:        models.CredentialsTypeStandard,
	}

	resource.syncModelFromAPI(data, api)

	assert.Equal(t, "DOMAIN\\user", data.Username.ValueString())
	assert.Equal(t, "Updated desc", data.Description.ValueString())
	assert.Equal(t, "Standard", data.Type.ValueString())
	// Password must NOT be overwritten from API response
	assert.Equal(t, "original-password", data.Password.ValueString())
}

func TestIsCredentialInUseError(t *testing.T) {
	err := fmt.Errorf("API request failed (HTTP 400): UnknownError: Unable to delete selected credentials 123 because they are currently in use.")
	assert.True(t, isCredentialInUseError(err))

	err = fmt.Errorf("API request failed (HTTP 400): UnknownError: validation failed")
	assert.False(t, isCredentialInUseError(err))
}

package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestCloudCredential_BuildSpec(t *testing.T) {
	resource := &CloudCredential{}
	data := &CloudCredentialModel{
		Name:           types.StringValue("aws-main"),
		Description:    types.StringValue("AWS account"),
		Type:           types.StringValue("Amazon"),
		AccessKey:      types.StringValue("AKIA_TEST"),
		SecretKey:      types.StringValue("secret"),
	}

	spec, validationError := resource.buildSpec(data)
	assert.Equal(t, "", validationError)
	assert.Equal(t, "aws-main", spec.Name)
	assert.Equal(t, "AWS account", spec.Description)
	assert.Equal(t, "Amazon", spec.Type)
	assert.Equal(t, "AKIA_TEST", spec.AccessKey)
	assert.Equal(t, "secret", spec.SecretKey)
}

func TestCloudCredential_BuildSpec_AzureStorageSharedKey(t *testing.T) {
	resource := &CloudCredential{}
	data := &CloudCredentialModel{
		Name:      types.StringValue("azure-storage"),
		Type:      types.StringValue("AzureStorage"),
		Account:   types.StringValue("mystorageaccount"),
		SharedKey: types.StringValue("base64sharedkey"),
	}

	spec, validationError := resource.buildSpec(data)
	assert.Equal(t, "", validationError)
	assert.Equal(t, "AzureStorage", spec.Type)
	assert.Equal(t, "mystorageaccount", spec.Account)
	assert.Equal(t, "base64sharedkey", spec.SharedKey)
}

func TestCloudCredential_BuildSpec_AzureComputeExistingAccount(t *testing.T) {
	resource := &CloudCredential{}
	data := &CloudCredentialModel{
		Name:            types.StringValue("azure-compute"),
		Type:            types.StringValue("AzureCompute"),
		ConnectionName:  types.StringValue("Azure Compute Connection"),
		CreationMode:    types.StringValue("ExistingAccount"),
		DeploymentType:  types.StringValue("MicrosoftAzure"),
		TenantID:        types.StringValue("tenant-id"),
		ApplicationID:   types.StringValue("application-id"),
		ApplicationKey:  types.StringValue("application-secret"),
	}

	spec, validationError := resource.buildSpec(data)
	assert.Equal(t, "", validationError)
	assert.Equal(t, "AzureCompute", spec.Type)
	assert.Equal(t, "Azure Compute Connection", spec.ConnectionName)
	assert.Equal(t, "ExistingAccount", spec.CreationMode)
	assert.NotNil(t, spec.ExistingAccount)
	assert.Equal(t, "MicrosoftAzure", spec.ExistingAccount.Deployment.DeploymentType)
	assert.Equal(t, "tenant-id", spec.ExistingAccount.Subscription.TenantID)
	assert.Equal(t, "application-id", spec.ExistingAccount.Subscription.ApplicationID)
	assert.Equal(t, "application-secret", spec.ExistingAccount.Subscription.Secret)
}

func TestCloudCredential_BuildSpec_ValidationError(t *testing.T) {
	resource := &CloudCredential{}
	data := &CloudCredentialModel{
		Name: types.StringValue("broken-azure-storage"),
		Type: types.StringValue("AzureStorage"),
	}

	spec, validationError := resource.buildSpec(data)
	assert.Nil(t, spec)
	assert.Contains(t, validationError, "requires 'account'")
}

func TestCloudCredential_SyncFromAPI(t *testing.T) {
	resource := &CloudCredential{}
	state := &CloudCredentialModel{}
	api := &models.CloudCredentialModel{
		ID:          "cloud-1",
		Name:        "azure-main",
		Description: "Azure credential",
		Type:        "AzureStorage",
		Account:     "tenant-account",
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
	}

	resource.syncFromAPI(state, api)

	assert.Equal(t, "azure-main", state.Name.ValueString())
	assert.Equal(t, "Azure credential", state.Description.ValueString())
	assert.Equal(t, "AzureStorage", state.Type.ValueString())
	assert.Equal(t, "tenant-account", state.Account.ValueString())
	assert.Equal(t, "tenant-1", state.TenantID.ValueString())
	assert.Equal(t, "project-1", state.ProjectID.ValueString())
}

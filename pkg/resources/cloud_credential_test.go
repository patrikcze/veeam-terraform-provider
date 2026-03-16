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
		AccountName:    types.StringValue("AKIA_TEST"),
		SecretKey:      types.StringValue("secret"),
		TenantID:       types.StringValue("tenant"),
		ApplicationID:  types.StringValue("app-id"),
		ApplicationKey: types.StringValue("app-key"),
		ProjectID:      types.StringValue("project-id"),
		ServiceAccount: types.StringValue("service-account-json"),
	}

	spec := resource.buildSpec(data)
	assert.Equal(t, "aws-main", spec.Name)
	assert.Equal(t, "AWS account", spec.Description)
	assert.Equal(t, "Amazon", spec.Type)
	assert.Equal(t, "AKIA_TEST", spec.AccountName)
	assert.Equal(t, "secret", spec.SecretKey)
	assert.Equal(t, "tenant", spec.TenantID)
	assert.Equal(t, "app-id", spec.ApplicationID)
	assert.Equal(t, "app-key", spec.ApplicationKey)
	assert.Equal(t, "project-id", spec.ProjectID)
	assert.Equal(t, "service-account-json", spec.ServiceAccount)
}

func TestCloudCredential_SyncFromAPI(t *testing.T) {
	resource := &CloudCredential{}
	state := &CloudCredentialModel{}
	api := &models.CloudCredentialModel{
		ID:          "cloud-1",
		Name:        "azure-main",
		Description: "Azure credential",
		Type:        "Azure",
		AccountName: "tenant-account",
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
	}

	resource.syncFromAPI(state, api)

	assert.Equal(t, "azure-main", state.Name.ValueString())
	assert.Equal(t, "Azure credential", state.Description.ValueString())
	assert.Equal(t, "Azure", state.Type.ValueString())
	assert.Equal(t, "tenant-account", state.AccountName.ValueString())
	assert.Equal(t, "tenant-1", state.TenantID.ValueString())
	assert.Equal(t, "project-1", state.ProjectID.ValueString())
}

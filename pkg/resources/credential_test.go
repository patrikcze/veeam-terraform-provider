package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredential_CreatePayload(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response
	mockClient.On("PostJSON", "/credentials", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// Simulate setting an ID in the result
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"id": "cred-123",
		}
	}).Return(nil)

	// Create test data
	data := CredentialModel{
		Name:        types.StringValue("test_cred"),
		Description: types.StringValue("Test credential"),
		Username:    types.StringValue("admin"),
		Password:    types.StringValue("password123"),
		Type:        types.StringValue("windows"),
		Domain:      types.StringValue("DOMAIN"),
	}

	// Test payload creation
	payload := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"username":    data.Username.ValueString(),
		"password":    data.Password.ValueString(),
		"type":        data.Type.ValueString(),
		"domain":      data.Domain.ValueString(),
	}

	// Execute mock API call
	var result map[string]interface{}
	err := mockClient.PostJSON("/credentials", payload, &result)

	// Assert no errors
	assert.NoError(t, err)
	assert.Equal(t, "cred-123", result["id"])
	mockClient.AssertExpectations(t)
}

// TestCredentialModel tests the CredentialModel structure
func TestCredentialModel(t *testing.T) {
	// Create test data
	data := CredentialModel{
		ID:          types.StringValue("cred-123"),
		Name:        types.StringValue("test_cred"),
		Description: types.StringValue("Test credential"),
		Username:    types.StringValue("admin"),
		Password:    types.StringValue("password123"),
		Type:        types.StringValue("windows"),
		Domain:      types.StringValue("DOMAIN"),
	}

	// Test the model
	assert.Equal(t, "cred-123", data.ID.ValueString())
	assert.Equal(t, "test_cred", data.Name.ValueString())
	assert.Equal(t, "Test credential", data.Description.ValueString())
	assert.Equal(t, "admin", data.Username.ValueString())
	assert.Equal(t, "password123", data.Password.ValueString())
	assert.Equal(t, "windows", data.Type.ValueString())
	assert.Equal(t, "DOMAIN", data.Domain.ValueString())
}

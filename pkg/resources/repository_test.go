package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRepository_CreatePayload(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response
	mockClient.On("PostJSON", "/repositories", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// Simulate setting an ID in the result
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"id": "repo-123",
		}
	}).Return(nil)

	// Create test data
	data := RepositoryModel{
		Name:        types.StringValue("test_repo"),
		Description: types.StringValue("Test repository"),
		Path:        types.StringValue("/backup/test"),
		Type:        types.StringValue("local"),
		Capacity:    types.Int64Value(1000),
	}

	// Test payload creation
	payload := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"path":        data.Path.ValueString(),
		"type":        data.Type.ValueString(),
		"capacity":    data.Capacity.ValueInt64(),
	}

	// Execute mock API call
	var result map[string]interface{}
	err := mockClient.PostJSON("/repositories", payload, &result)

	// Assert no errors
	assert.NoError(t, err)
	assert.Equal(t, "repo-123", result["id"])
	mockClient.AssertExpectations(t)
}

// TestRepositoryModel tests the RepositoryModel structure
func TestRepositoryModel(t *testing.T) {
	// Create test data
	data := RepositoryModel{
		ID:          types.StringValue("repo-123"),
		Name:        types.StringValue("test_repo"),
		Description: types.StringValue("Test repository"),
		Path:        types.StringValue("/backup/test"),
		Type:        types.StringValue("local"),
		Capacity:    types.Int64Value(1000),
	}

	// Test the model
	assert.Equal(t, "repo-123", data.ID.ValueString())
	assert.Equal(t, "test_repo", data.Name.ValueString())
	assert.Equal(t, "Test repository", data.Description.ValueString())
	assert.Equal(t, "/backup/test", data.Path.ValueString())
	assert.Equal(t, "local", data.Type.ValueString())
	assert.Equal(t, int64(1000), data.Capacity.ValueInt64())
}

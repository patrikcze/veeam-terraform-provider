package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRepositoriesDataSource_ReadAllRepositories(t *testing.T) {
	// Setup mock client
	mockClient := new(MockVeeamClient)

	// Mock successful API response for all repositories
	mockClient.On("GetJSON", mock.Anything, "/api/v1/repositories", mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*[]map[string]interface{})
		*result = []map[string]interface{}{
			{
				"id":          "repo-1",
				"name":        "Main Repository",
				"description": "Main backup repository",
				"path":        "/backup/main",
				"type":        "local",
				"capacity":    float64(1000),
				"freeSpace":   float64(800),
				"usedSpace":   float64(200),
				"status":      "active",
				"createdAt":   "2023-01-01T00:00:00Z",
				"updatedAt":   "2023-01-01T00:00:00Z",
			},
			{
				"id":          "repo-2",
				"name":        "Archive Repository",
				"description": "Archive backup repository",
				"path":        "/backup/archive",
				"type":        "local",
				"capacity":    float64(2000),
				"freeSpace":   float64(1500),
				"usedSpace":   float64(500),
				"status":      "active",
				"createdAt":   "2023-01-01T00:00:00Z",
				"updatedAt":   "2023-01-01T00:00:00Z",
			},
		}
	}).Return(nil)

	// Execute mock API call
	var response []map[string]interface{}
	err := mockClient.GetJSON(context.Background(), "/api/v1/repositories", &response)

	// Assert no errors
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "repo-1", response[0]["id"])
	assert.Equal(t, "Main Repository", response[0]["name"])
	mockClient.AssertExpectations(t)
}

// TestRepositoriesDataSourceModel tests the RepositoriesDataSourceModel structure
func TestRepositoriesDataSourceModel(t *testing.T) {
	// Create test data
	data := RepositoriesDataSourceModel{
		ID:             types.StringValue("repositories"),
		RepositoryID:   types.StringValue("repo-1"),
		RepositoryName: types.StringValue("Main Repository"),
		Repositories:   []RepositoryDataModel{},
	}

	// Test the model
	assert.Equal(t, "repositories", data.ID.ValueString())
	assert.Equal(t, "repo-1", data.RepositoryID.ValueString())
	assert.Equal(t, "Main Repository", data.RepositoryName.ValueString())
	assert.NotNil(t, data.Repositories)
}

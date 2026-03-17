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

func TestEncryptionPassword_CreatePayload(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &EncryptionPassword{client: mockClient}

	data := &EncryptionPasswordModel{
		Password: types.StringValue("super-secret"),
		Hint:     types.StringValue("backup key"),
	}

	mockClient.On("PostJSON", mock.Anything, client.PathEncryptionPasswords, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		payload := args.Get(2).(*models.EncryptionPasswordSpec)
		assert.Equal(t, "super-secret", payload.Password)
		assert.Equal(t, "backup key", payload.Hint)

		result := args.Get(3).(*models.EncryptionPasswordModel)
		result.ID = "enc-123"
		result.Hint = "backup key"
	}).Return(nil)

	var result models.EncryptionPasswordModel
	err := resource.client.PostJSON(context.Background(), client.PathEncryptionPasswords, &models.EncryptionPasswordSpec{
		Password: data.Password.ValueString(),
		Hint:     data.Hint.ValueString(),
	}, &result)

	assert.NoError(t, err)
	assert.Equal(t, "enc-123", result.ID)
	assert.Equal(t, "backup key", result.Hint)
	mockClient.AssertExpectations(t)
}

func TestEncryptionPassword_ReadKeepsPasswordState(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &EncryptionPassword{client: mockClient}

	data := &EncryptionPasswordModel{
		ID:       types.StringValue("enc-123"),
		Password: types.StringValue("super-secret"),
		Hint:     types.StringValue("old-hint"),
	}

	endpoint := fmt.Sprintf(client.PathEncryptionPasswordByID, data.ID.ValueString())
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*models.EncryptionPasswordModel)
		result.ID = "enc-123"
		result.Hint = "new-hint"
	}).Return(nil)

	var result models.EncryptionPasswordModel
	err := resource.client.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	data.Hint = types.StringValue(result.Hint)
	assert.Equal(t, "new-hint", data.Hint.ValueString())
	assert.Equal(t, "super-secret", data.Password.ValueString())
	mockClient.AssertExpectations(t)
}

func TestEncryptionPassword_UpdateAndDeleteEndpoints(t *testing.T) {
	mockClient := new(MockVeeamClient)
	resource := &EncryptionPassword{client: mockClient}

	id := "enc-123"
	endpoint := fmt.Sprintf(client.PathEncryptionPasswordByID, id)

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(*models.EncryptionPasswordSpec)
		assert.Equal(t, "updated-secret", payload.Password)
		assert.Equal(t, "updated-hint", payload.Hint)
	}).Return(nil)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := resource.client.PutJSON(context.Background(), endpoint, &models.EncryptionPasswordSpec{
		Password: "updated-secret",
		Hint:     "updated-hint",
	}, nil)
	assert.NoError(t, err)

	err = resource.client.DeleteJSON(context.Background(), endpoint)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

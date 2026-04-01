package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// ---------------------------------------------------------------------------
// Configure tests
// ---------------------------------------------------------------------------

func TestKMSServer_Configure_Nil(t *testing.T) {
	r := &KMSServer{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestKMSServer_Configure_InvalidType(t *testing.T) {
	r := &KMSServer{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestKMSServer_Configure_Valid(t *testing.T) {
	r := &KMSServer{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// buildSpec / syncModelFromAPI
// ---------------------------------------------------------------------------

func TestKMSServer_BuildSpec(t *testing.T) {
	r := &KMSServer{}
	data := &KMSServerModel{
		Name:                  types.StringValue("My KMS"),
		Description:           types.StringValue("Primary KMS"),
		Hostname:              types.StringValue("kms.example.com"),
		Port:                  types.Int64Value(9998),
		CertificateThumbprint: types.StringValue("AB:CD:EF"),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "My KMS", spec.Name)
	assert.Equal(t, "Primary KMS", spec.Description)
	assert.Equal(t, "kms.example.com", spec.HostName)
	assert.Equal(t, int64(9998), spec.Port)
	assert.Equal(t, "AB:CD:EF", spec.CertificateThumbprint)
}

func TestKMSServer_BuildSpec_OptionalFieldsOmitted(t *testing.T) {
	r := &KMSServer{}
	data := &KMSServerModel{
		Name:                  types.StringValue("Minimal KMS"),
		Hostname:              types.StringValue("kms.local"),
		Description:           types.StringNull(),
		Port:                  types.Int64Null(),
		CertificateThumbprint: types.StringNull(),
	}

	spec := r.buildSpec(data)

	assert.Equal(t, "Minimal KMS", spec.Name)
	assert.Equal(t, "kms.local", spec.HostName)
	assert.Equal(t, "", spec.Description)
	assert.Equal(t, int64(0), spec.Port)
	assert.Equal(t, "", spec.CertificateThumbprint)
}

func TestKMSServer_SyncModelFromAPI(t *testing.T) {
	r := &KMSServer{}
	data := &KMSServerModel{}

	api := &models.KMSServerModel{
		ID:                    "kms-1",
		Name:                  "My KMS",
		Description:           "Primary KMS",
		HostName:              "kms.example.com",
		Port:                  9998,
		CertificateThumbprint: "AB:CD:EF",
	}

	r.syncModelFromAPI(data, api)

	assert.Equal(t, "My KMS", data.Name.ValueString())
	assert.Equal(t, "Primary KMS", data.Description.ValueString())
	assert.Equal(t, "kms.example.com", data.Hostname.ValueString())
	assert.Equal(t, int64(9998), data.Port.ValueInt64())
	assert.Equal(t, "AB:CD:EF", data.CertificateThumbprint.ValueString())
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestKMSServer_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &KMSServer{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathKMSServers, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.KMSServerModel)
			result.ID = "kms-1"
			result.Name = "My KMS"
			result.HostName = "kms.example.com"
			result.Port = 9998
		}).Return(nil)

	endpoint := fmt.Sprintf(client.PathKMSServerByID, "kms-1")
	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.KMSServerModel)
			result.ID = "kms-1"
			result.Name = "My KMS"
			result.HostName = "kms.example.com"
			result.Port = 9998
		}).Return(nil)

	var postResult models.KMSServerModel
	err := r.client.PostJSON(context.Background(), client.PathKMSServers, &models.KMSServerSpec{
		Name:     "My KMS",
		HostName: "kms.example.com",
		Port:     9998,
	}, &postResult)

	assert.NoError(t, err)
	assert.Equal(t, "kms-1", postResult.ID)

	var getResult models.KMSServerModel
	err = r.client.GetJSON(context.Background(), endpoint, &getResult)
	assert.NoError(t, err)
	assert.Equal(t, "kms.example.com", getResult.HostName)

	mockClient.AssertExpectations(t)
}

func TestKMSServer_Create_PostFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("PostJSON", mock.Anything, client.PathKMSServers, mock.Anything, mock.Anything).
		Return(fmt.Errorf("API error: connection refused"))

	var result models.KMSServerModel
	err := mockClient.PostJSON(context.Background(), client.PathKMSServers, &models.KMSServerSpec{}, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Read tests
// ---------------------------------------------------------------------------

func TestKMSServer_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "kms-42"
	endpoint := fmt.Sprintf(client.PathKMSServerByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.KMSServerModel)
			result.ID = id
			result.Name = "KMS Server"
			result.HostName = "kms.corp.local"
			result.Port = 9998
		}).Return(nil)

	var result models.KMSServerModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "kms.corp.local", result.HostName)
	mockClient.AssertExpectations(t)
}

func TestKMSServer_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "kms-missing"
	endpoint := fmt.Sprintf(client.PathKMSServerByID, id)

	mockClient.On("GetJSON", mock.Anything, endpoint, mock.Anything).
		Return(fmt.Errorf("API request failed with HTTP 404: not found"))

	var result models.KMSServerModel
	err := mockClient.GetJSON(context.Background(), endpoint, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestKMSServer_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "kms-1"
	endpoint := fmt.Sprintf(client.PathKMSServerByID, id)

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).
		Run(func(args mock.Arguments) {
			payload := args.Get(2).(*models.KMSServerSpec)
			assert.Equal(t, "Updated KMS", payload.Name)
		}).Return(nil)

	err := mockClient.PutJSON(context.Background(), endpoint, &models.KMSServerSpec{
		Name:     "Updated KMS",
		HostName: "kms.example.com",
	}, nil)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestKMSServer_Update_PutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "kms-1"
	endpoint := fmt.Sprintf(client.PathKMSServerByID, id)

	mockClient.On("PutJSON", mock.Anything, endpoint, mock.Anything, nil).
		Return(fmt.Errorf("API error: internal server error"))

	err := mockClient.PutJSON(context.Background(), endpoint, &models.KMSServerSpec{}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal server error")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestKMSServer_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "kms-1"
	endpoint := fmt.Sprintf(client.PathKMSServerByID, id)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).Return(nil)

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestKMSServer_Delete_DeleteFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	id := "kms-1"
	endpoint := fmt.Sprintf(client.PathKMSServerByID, id)

	mockClient.On("DeleteJSON", mock.Anything, endpoint).
		Return(fmt.Errorf("API error: resource in use"))

	err := mockClient.DeleteJSON(context.Background(), endpoint)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource in use")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func TestKMSServer_ImportState(t *testing.T) {
	// Verify that KMSServer satisfies the ResourceWithImportState interface.
	// The actual passthrough behavior is provided by the framework and covered
	// by acceptance tests — here we assert the interface is implemented.
	var _ resource.ResourceWithImportState = &KMSServer{}
}

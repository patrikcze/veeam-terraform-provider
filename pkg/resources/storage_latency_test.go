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
)

func TestStorageLatency_Metadata(t *testing.T) {
	r := NewStorageLatency()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_storage_latency", resp.TypeName)
}

func TestStorageLatency_Schema(t *testing.T) {
	r := &StorageLatency{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "enabled")
	assert.Contains(t, resp.Schema.Attributes, "latency_limit_ms")
	assert.Contains(t, resp.Schema.Attributes, "throttling_io_enabled")
	assert.Contains(t, resp.Schema.Attributes, "throttling_io_limit")
	assert.Contains(t, resp.Schema.Attributes, "stop_jobs_enabled")
	assert.Contains(t, resp.Schema.Attributes, "stop_jobs_limit_ms")
}

func TestStorageLatency_Configure_Nil(t *testing.T) {
	r := &StorageLatency{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestStorageLatency_Configure_InvalidType(t *testing.T) {
	r := &StorageLatency{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestStorageLatency_SyncFromPayload_Full(t *testing.T) {
	raw := map[string]interface{}{
		"isEnabled":      true,
		"latencyLimitMs": float64(20),
		"throttlingIo": map[string]interface{}{
			"isEnabled": true,
			"iopsLimit": float64(100),
		},
		"stopJobs": map[string]interface{}{
			"isEnabled":      false,
			"latencyLimitMs": float64(50),
		},
	}

	data := &StorageLatencyModel{}
	syncStorageLatencyFromPayload(raw, data)

	assert.True(t, data.Enabled.ValueBool())
	assert.Equal(t, int64(20), data.LatencyLimitMs.ValueInt64())
	assert.True(t, data.ThrottlingIOEnabled.ValueBool())
	assert.Equal(t, int64(100), data.ThrottlingIOLimit.ValueInt64())
	assert.False(t, data.StopJobsEnabled.ValueBool())
	assert.Equal(t, int64(50), data.StopJobsLimitMs.ValueInt64())
}

func TestStorageLatency_SyncFromPayload_Empty(t *testing.T) {
	raw := map[string]interface{}{}
	data := &StorageLatencyModel{}
	syncStorageLatencyFromPayload(raw, data)

	assert.False(t, data.Enabled.ValueBool())
	assert.False(t, data.ThrottlingIOEnabled.ValueBool())
	assert.False(t, data.StopJobsEnabled.ValueBool())
}

func TestStorageLatency_Apply_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathStorageLatency, mock.Anything).
		Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathStorageLatency, mock.Anything, nil).
		Return(nil)

	r := &StorageLatency{client: mockClient}
	data := &StorageLatencyModel{
		Enabled:             types.BoolValue(true),
		LatencyLimitMs:      types.Int64Value(20),
		ThrottlingIOEnabled: types.BoolNull(),
		ThrottlingIOLimit:   types.Int64Null(),
		StopJobsEnabled:     types.BoolNull(),
		StopJobsLimitMs:     types.Int64Null(),
	}

	err := r.applyStorageLatency(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestStorageLatency_Apply_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathStorageLatency, mock.Anything).
		Return(fmt.Errorf("HTTP 503"))

	r := &StorageLatency{client: mockClient}
	data := &StorageLatencyModel{
		Enabled: types.BoolValue(true),
	}

	err := r.applyStorageLatency(context.Background(), data)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestStorageLatency_Delete_NoOp(t *testing.T) {
	r := &StorageLatency{}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestStorageLatency_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &StorageLatency{}
}

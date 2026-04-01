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

// ---------------------------------------------------------------------------
// TrafficRules — Configure tests
// ---------------------------------------------------------------------------

func TestTrafficRules_Configure_Nil(t *testing.T) {
	r := &TrafficRules{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.client)
}

func TestTrafficRules_Configure_InvalidType(t *testing.T) {
	r := &TrafficRules{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestTrafficRules_Configure_Valid(t *testing.T) {
	r := &TrafficRules{}
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

// ---------------------------------------------------------------------------
// TrafficRules — Create / putTrafficRules
// ---------------------------------------------------------------------------

func TestTrafficRules_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &TrafficRules{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathTrafficRules, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"throttlingEnabled": false,
			"rules":             []interface{}{},
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathTrafficRules, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, true, payload["throttlingEnabled"])
		rules, ok := payload["rules"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, rules, 0)
	}).Return(nil)

	data := &TrafficRulesModel{
		ThrottlingEnabled: types.BoolValue(true),
		ThrottlingRules:   types.StringValue("[]"),
	}

	err := r.putTrafficRules(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestTrafficRules_Create_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &TrafficRules{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathTrafficRules, mock.Anything).
		Return(fmt.Errorf("connection refused"))

	err := r.putTrafficRules(context.Background(), &TrafficRulesModel{
		ThrottlingEnabled: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading current traffic rules")
	mockClient.AssertExpectations(t)
}

func TestTrafficRules_Create_PutFails(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &TrafficRules{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathTrafficRules, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathTrafficRules, mock.Anything, nil).
		Return(fmt.Errorf("server error"))

	err := r.putTrafficRules(context.Background(), &TrafficRulesModel{
		ThrottlingEnabled: types.BoolValue(true),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// TrafficRules — Read / syncTrafficRulesFromAPI
// ---------------------------------------------------------------------------

func TestTrafficRules_Read_Success(t *testing.T) {
	raw := map[string]interface{}{
		"throttlingEnabled": true,
		"rules":             []interface{}{},
	}

	data := &TrafficRulesModel{}
	err := syncTrafficRulesFromAPI(data, raw)
	assert.NoError(t, err)
	assert.Equal(t, true, data.ThrottlingEnabled.ValueBool())
	assert.Equal(t, "[]", data.ThrottlingRules.ValueString())
}

func TestTrafficRules_Read_WithRules(t *testing.T) {
	raw := map[string]interface{}{
		"throttlingEnabled": true,
		"rules": []interface{}{
			map[string]interface{}{
				"name":      "Rule1",
				"sendSpeed": float64(10),
			},
		},
	}

	data := &TrafficRulesModel{}
	err := syncTrafficRulesFromAPI(data, raw)
	assert.NoError(t, err)
	assert.Equal(t, true, data.ThrottlingEnabled.ValueBool())
	assert.Contains(t, data.ThrottlingRules.ValueString(), "Rule1")
}

func TestTrafficRules_Read_MissingRulesKey(t *testing.T) {
	raw := map[string]interface{}{
		"throttlingEnabled": false,
	}

	data := &TrafficRulesModel{}
	err := syncTrafficRulesFromAPI(data, raw)
	assert.NoError(t, err)
	assert.Equal(t, false, data.ThrottlingEnabled.ValueBool())
	assert.Equal(t, "[]", data.ThrottlingRules.ValueString())
}

func TestTrafficRules_Read_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathTrafficRules, mock.Anything).
		Return(fmt.Errorf("forbidden"))

	var raw map[string]interface{}
	err := mockClient.GetJSON(context.Background(), client.PathTrafficRules, &raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// TrafficRules — Update
// ---------------------------------------------------------------------------

func TestTrafficRules_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &TrafficRules{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathTrafficRules, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*map[string]interface{})
		*result = map[string]interface{}{
			"throttlingEnabled": true,
			"rules":             []interface{}{},
		}
	}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathTrafficRules, mock.Anything, nil).Run(func(args mock.Arguments) {
		payload := args.Get(2).(map[string]interface{})
		assert.Equal(t, false, payload["throttlingEnabled"])
	}).Return(nil)

	err := r.putTrafficRules(context.Background(), &TrafficRulesModel{
		ThrottlingEnabled: types.BoolValue(false),
	})
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// TrafficRules — Delete (no-op)
// ---------------------------------------------------------------------------

func TestTrafficRules_Delete_Success(t *testing.T) {
	r := &TrafficRules{}
	r.Delete(context.Background(), resource.DeleteRequest{}, &resource.DeleteResponse{})
}

// ---------------------------------------------------------------------------
// TrafficRules — ImportState
// ---------------------------------------------------------------------------

func TestTrafficRules_ImportState(t *testing.T) {
	importStateWithID(t, &TrafficRules{}, "traffic-rules")
}

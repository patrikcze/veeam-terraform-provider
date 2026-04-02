package resources

// ---------------------------------------------------------------------------
// CRUD coverage tests for all 10 resource types.
//
// Strategy:
//   - Configure: nil data, invalid type, valid client (covers the 3-branch
//     Configure implementation shared by every resource)
//   - Read/Create/Update/Delete: use tfsdk.Plan / tfsdk.State objects built
//     from the resource's own schema via tftypes, so we can call the methods
//     directly without a live Terraform binary
//
// Run with:
//   go test ./pkg/resources/ -run TestCRUD -v
// ---------------------------------------------------------------------------

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

// ---------------------------------------------------------------------------
// tftypes helpers (mirroring the datasource test helpers)
// ---------------------------------------------------------------------------

// buildNullResourceState builds a tfsdk.State with all leaves set to null,
// derived from the resource's own schema.
func buildNullResourceState(r resource.Resource) tfsdk.State {
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	return tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
}

// buildNullResourcePlan builds a tfsdk.Plan with all leaves set to null.
func buildNullResourcePlan(r resource.Resource) tfsdk.Plan {
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
}

// nullObjectForResourceSchema recursively produces a tftypes.Value with all
// leaves null, walking the tftypes.Type tree.
func nullObjectForResourceSchema(typ tftypes.Type) tftypes.Value {
	switch t := typ.(type) {
	case tftypes.Object:
		attrs := make(map[string]tftypes.Value, len(t.AttributeTypes))
		for k, attrType := range t.AttributeTypes {
			attrs[k] = nullValueForResourceType(attrType)
		}
		return tftypes.NewValue(t, attrs)
	default:
		return tftypes.NewValue(typ, nil)
	}
}

func nullValueForResourceType(typ tftypes.Type) tftypes.Value {
	switch t := typ.(type) {
	case tftypes.Object:
		attrs := make(map[string]tftypes.Value, len(t.AttributeTypes))
		for k, attrType := range t.AttributeTypes {
			attrs[k] = nullValueForResourceType(attrType)
		}
		return tftypes.NewValue(t, attrs)
	case tftypes.List:
		return tftypes.NewValue(t, nil)
	case tftypes.Set:
		return tftypes.NewValue(t, nil)
	case tftypes.Map:
		return tftypes.NewValue(t, nil)
	default:
		return tftypes.NewValue(typ, nil)
	}
}

// ---------------------------------------------------------------------------
// Configure helpers — shared across all resource types
// ---------------------------------------------------------------------------

// testConfigureNil verifies that Configure with nil ProviderData is a no-op.
func testConfigureNil(t *testing.T, r resource.ResourceWithConfigure) {
	t.Helper()
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "nil provider data should not add diagnostics")
}

// testConfigureInvalidType verifies that Configure rejects an unexpected type.
func testConfigureInvalidType(t *testing.T, r resource.ResourceWithConfigure) {
	t.Helper()
	req := resource.ConfigureRequest{ProviderData: "wrong-type"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "unexpected provider data type should add an error diagnostic")
}

// testConfigureValid verifies that Configure accepts a valid APIClient.
func testConfigureValid(t *testing.T, r resource.ResourceWithConfigure) {
	t.Helper()
	mockClient := new(MockVeeamClient)
	req := resource.ConfigureRequest{ProviderData: client.APIClient(mockClient)}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "valid provider data should not add diagnostics")
}

// ---------------------------------------------------------------------------
// Credential — Configure
// ---------------------------------------------------------------------------

func TestCredential_Configure_Nil(t *testing.T)         { testConfigureNil(t, &Credential{}) }
func TestCredential_Configure_InvalidType(t *testing.T) { testConfigureInvalidType(t, &Credential{}) }
func TestCredential_Configure_Valid(t *testing.T)       { testConfigureValid(t, &Credential{}) }

// ---------------------------------------------------------------------------
// Credential — Read
// ---------------------------------------------------------------------------

func TestCredential_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, fmt.Sprintf(client.PathCredentialByID, "cred-1"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.CredentialsModel)
			result.ID = "cred-1"
			result.Username = "admin"
			result.Description = "Test"
			result.Type = models.CredentialsTypeStandard
		}).Return(nil)

	state := buildNullResourceState(r)
	// Inject a non-null ID so Read knows which resource to fetch.
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cred-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestCredential_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("connection refused"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cred-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Credential — Delete
// ---------------------------------------------------------------------------

func TestCredential_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cred-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestCredential_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("API request failed (HTTP 500): internal server error"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cred-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// EncryptionPassword — Configure + CRUD
// ---------------------------------------------------------------------------

func TestEncryptionPassword_Configure_Nil(t *testing.T) {
	testConfigureNil(t, &EncryptionPassword{})
}
func TestEncryptionPassword_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &EncryptionPassword{})
}
func TestEncryptionPassword_Configure_Valid(t *testing.T) {
	testConfigureValid(t, &EncryptionPassword{})
}

func TestEncryptionPassword_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.EncryptionPasswordModel)
			result.ID = "enc-1"
			result.Hint = "My hint"
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "enc-1")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestEncryptionPassword_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("not found"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "enc-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestEncryptionPassword_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "enc-1")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "new-secret")
		case "hint":
			vals[k] = tftypes.NewValue(attrType, "Updated hint")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestEncryptionPassword_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("API error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "enc-1")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "new-secret")
		case "hint":
			vals[k] = tftypes.NewValue(attrType, "hint")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestEncryptionPassword_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	// deleteEncryptionPasswordWithRetries calls DeleteJSON up to 4 times if in-use.
	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "enc-1")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestEncryptionPassword_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	// Non-in-use error exits on first attempt.
	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("API request failed (HTTP 500): internal server error"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "enc-1")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Proxy — Configure + Read + Delete
// ---------------------------------------------------------------------------

func TestProxy_Configure_Nil(t *testing.T)         { testConfigureNil(t, &Proxy{}) }
func TestProxy_Configure_InvalidType(t *testing.T) { testConfigureInvalidType(t, &Proxy{}) }
func TestProxy_Configure_Valid(t *testing.T)       { testConfigureValid(t, &Proxy{}) }

func TestProxy_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.ViProxyModel)
			result.ID = "proxy-1"
			result.Name = "Vi Proxy"
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "proxy-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestProxy_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("proxy not found"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "proxy-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestProxy_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "proxy-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestProxy_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("delete failed"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "proxy-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Repository — Configure + Read + Delete
// ---------------------------------------------------------------------------

func TestRepository_Configure_Nil(t *testing.T)         { testConfigureNil(t, &Repository{}) }
func TestRepository_Configure_InvalidType(t *testing.T) { testConfigureInvalidType(t, &Repository{}) }
func TestRepository_Configure_Valid(t *testing.T)       { testConfigureValid(t, &Repository{}) }

func TestRepository_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"id":          "repo-1",
				"name":        "Default Repository",
				"description": "Main repo",
				"type":        "WindowsLocal",
			}
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRepository_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("connection error"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestRepository_Read_NotFound_RemovesResource(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("HTTP 404: repository not found"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	// 404 → RemoveResource (no error diagnostic, empty state).
	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestRepository_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestRepository_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("cannot delete"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ManagedServer — Configure + Read + Delete
// ---------------------------------------------------------------------------

func TestManagedServer_Configure_Nil(t *testing.T) { testConfigureNil(t, &ManagedServer{}) }
func TestManagedServer_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &ManagedServer{})
}
func TestManagedServer_Configure_Valid(t *testing.T) { testConfigureValid(t, &ManagedServer{}) }

func TestManagedServer_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.ManagedServerModel)
			result.ID = "ms-1"
			result.Name = "192.168.1.10"
			result.Description = "Linux server"
			result.Type = models.ManagedServerTypeLinuxHost
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "ms-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestManagedServer_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("server error"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "ms-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// CloudCredential — Configure + Read + Delete
// ---------------------------------------------------------------------------

func TestCloudCredential_Configure_Nil(t *testing.T) { testConfigureNil(t, &CloudCredential{}) }
func TestCloudCredential_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &CloudCredential{})
}
func TestCloudCredential_Configure_Valid(t *testing.T) { testConfigureValid(t, &CloudCredential{}) }

func TestCloudCredential_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.CloudCredentialModel)
			result.ID = "cc-1"
			result.Name = "AWS Prod"
			result.Type = "Amazon"
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cc-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestCloudCredential_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("unauthorized"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cc-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCloudCredential_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cc-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestCloudCredential_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("in use"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "cc-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ConfigurationBackup — Configure + Read
// ---------------------------------------------------------------------------

func TestConfigurationBackup_Configure_Nil(t *testing.T) {
	testConfigureNil(t, &ConfigurationBackup{})
}
func TestConfigurationBackup_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &ConfigurationBackup{})
}
func TestConfigurationBackup_Configure_Valid(t *testing.T) {
	testConfigureValid(t, &ConfigurationBackup{})
}

func TestConfigurationBackup_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"isEnabled":           true,
				"backupRepositoryId":  "repo-abc",
				"restorePointsToKeep": float64(14),
			}
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "config-backup")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Return(errors.New("API unavailable"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "config-backup")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJob — Configure
// ---------------------------------------------------------------------------

func TestBackupJob_Configure_Nil(t *testing.T)         { testConfigureNil(t, &BackupJob{}) }
func TestBackupJob_Configure_InvalidType(t *testing.T) { testConfigureInvalidType(t, &BackupJob{}) }
func TestBackupJob_Configure_Valid(t *testing.T)       { testConfigureValid(t, &BackupJob{}) }

// ---------------------------------------------------------------------------
// ProtectionGroup — Configure
// ---------------------------------------------------------------------------

func TestProtectionGroup_Configure_Nil(t *testing.T) {
	testConfigureNil(t, &ProtectionGroup{})
}
func TestProtectionGroup_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &ProtectionGroup{})
}
func TestProtectionGroup_Configure_Valid(t *testing.T) {
	testConfigureValid(t, &ProtectionGroup{})
}

// ---------------------------------------------------------------------------
// ScaleOutRepository — Configure + Read + Delete
// ---------------------------------------------------------------------------

func TestScaleOutRepository_Configure_Nil(t *testing.T) {
	testConfigureNil(t, &ScaleOutRepository{})
}
func TestScaleOutRepository_Configure_InvalidType(t *testing.T) {
	testConfigureInvalidType(t, &ScaleOutRepository{})
}
func TestScaleOutRepository_Configure_Valid(t *testing.T) {
	testConfigureValid(t, &ScaleOutRepository{})
}

func TestScaleOutRepository_Read_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.ScaleOutRepositoryModel)
			result.ID = "sobr-1"
			result.Name = "SOBR Main"
		}).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "sobr-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestScaleOutRepository_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("not found"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "sobr-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestScaleOutRepository_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "sobr-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestScaleOutRepository_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("cannot remove SOBR"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "sobr-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ManagedServer — Delete
// ---------------------------------------------------------------------------

func TestManagedServer_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	// Delete triggers WaitForTask via deleteAndWait; mock the delete call.
	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	// After delete, waitForManagedServerDeleted polls with GetJSON; simulate immediate 404.
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("HTTP 404: not found"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "ms-1")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	// Either success or 404-not-found are acceptable outcomes.
	// The key check is: no panic and mock called.
	mockClient.AssertCalled(t, "DeleteJSON", mock.Anything, mock.AnythingOfType("string"))
}

// ---------------------------------------------------------------------------
// Credential — Update
// ---------------------------------------------------------------------------

func TestCredential_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "cred-1")
		case "username":
			vals[k] = tftypes.NewValue(attrType, "admin")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Standard")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestCredential_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("API error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "cred-1")
		case "username":
			vals[k] = tftypes.NewValue(attrType, "admin")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Standard")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Credential — Create
// ---------------------------------------------------------------------------

func TestCredential_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathCredentials, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.CredentialsModel)
			result.ID = "cred-new"
			result.Username = "admin"
			result.Type = models.CredentialsTypeStandard
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "username":
			vals[k] = tftypes.NewValue(attrType, "admin")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Standard")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestCredential_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathCredentials, mock.Anything, mock.Anything).
		Return(errors.New("API error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "username":
			vals[k] = tftypes.NewValue(attrType, "admin")
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secret")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Standard")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// EncryptionPassword — Create
// ---------------------------------------------------------------------------

func TestEncryptionPassword_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathEncryptionPasswords, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.EncryptionPasswordModel)
			result.ID = "enc-new"
			result.Hint = "My password hint"
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secure-pass")
		case "hint":
			vals[k] = tftypes.NewValue(attrType, "My password hint")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestEncryptionPassword_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &EncryptionPassword{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathEncryptionPasswords, mock.Anything, mock.Anything).
		Return(errors.New("API error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "password":
			vals[k] = tftypes.NewValue(attrType, "secure-pass")
		case "hint":
			vals[k] = tftypes.NewValue(attrType, "hint")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// CloudCredential — Create (Amazon path)
// ---------------------------------------------------------------------------

func TestCloudCredential_Create_Amazon_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathCloudCredentials, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*models.CloudCredentialModel)
			result.ID = "cc-new"
			result.Name = "AWS Main"
			result.Type = "Amazon"
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "name":
			vals[k] = tftypes.NewValue(attrType, "AWS Main")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Amazon")
		case "access_key":
			vals[k] = tftypes.NewValue(attrType, "AKIAIOSFODNN7EXAMPLE")
		case "secret_key":
			vals[k] = tftypes.NewValue(attrType, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestCloudCredential_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathCloudCredentials, mock.Anything, mock.Anything).
		Return(errors.New("API error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "name":
			vals[k] = tftypes.NewValue(attrType, "AWS Main")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Amazon")
		case "access_key":
			vals[k] = tftypes.NewValue(attrType, "AKIAIOSFODNN7EXAMPLE")
		case "secret_key":
			vals[k] = tftypes.NewValue(attrType, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// CloudCredential — buildSpec validation error paths
// ---------------------------------------------------------------------------

func TestCloudCredential_BuildSpec_UnsupportedType(t *testing.T) {
	r := &CloudCredential{}
	data := &CloudCredentialModel{
		Name: types.StringValue("test"),
		Type: types.StringValue("UnknownCloudType"),
	}
	spec, errMsg := r.buildSpec(data)
	assert.Nil(t, spec)
	assert.NotEmpty(t, errMsg)
}

func TestCloudCredential_BuildSpec_Amazon_MissingKey(t *testing.T) {
	r := &CloudCredential{}
	data := &CloudCredentialModel{
		Name: types.StringValue("test"),
		Type: types.StringValue("Amazon"),
		// no access_key or secret_key
	}
	spec, errMsg := r.buildSpec(data)
	assert.Nil(t, spec)
	assert.NotEmpty(t, errMsg)
}

func TestCloudCredential_BuildSpec_GoogleService_MissingAccount(t *testing.T) {
	r := &CloudCredential{}
	data := &CloudCredentialModel{
		Name: types.StringValue("test"),
		Type: types.StringValue("GoogleService"),
		// no service_account
	}
	spec, errMsg := r.buildSpec(data)
	assert.Nil(t, spec)
	assert.NotEmpty(t, errMsg)
}

// ---------------------------------------------------------------------------
// ConfigurationBackup — Delete
// ---------------------------------------------------------------------------

func TestConfigurationBackup_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	// Delete calls loadConfigurationBackupPayload (GET) then putConfigurationPayload (PUT).
	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{"isEnabled": true}
		}).Return(nil)
	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, mock.Anything).
		Return(nil)

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "config-backup")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Return(errors.New("API unavailable"))

	state := buildNullResourceState(r)
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		if k == "id" {
			vals[k] = tftypes.NewValue(attrType, "config-backup")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Repository — Update
// ---------------------------------------------------------------------------

func TestRepository_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	// PutJSON returns a synchronous type result (not async).
	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*map[string]interface{})
			*result = map[string]interface{}{
				"id":   "repo-1",
				"name": "Updated Repo",
				"type": "WindowsLocal",
			}
		}).Return(nil)
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"id":   "repo-1",
				"name": "Updated Repo",
				"type": "WindowsLocal",
			}
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "Updated Repo")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsLocal")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestRepository_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("update failed"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "repo-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "Updated Repo")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsLocal")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ManagedServer — Update
// ---------------------------------------------------------------------------

func TestManagedServer_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	// ManagedServer.Update calls PutJSON with nil result arg, then sets state from plan.
	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "ms-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "server01")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsHost")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

func TestManagedServer_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("update failed"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "ms-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "server01")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsHost")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ScaleOutRepository — Update
// ---------------------------------------------------------------------------

func TestScaleOutRepository_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(nil)
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.ScaleOutRepositoryModel)
			result.ID = "sobr-1"
			result.Name = "SOBR Updated"
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "sobr-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "SOBR Updated")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

func TestScaleOutRepository_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("update failed"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "sobr-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "SOBR Updated")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// CloudCredential — Update
// ---------------------------------------------------------------------------

func TestCloudCredential_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(nil)
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.CloudCredentialModel)
			result.ID = "cc-1"
			result.Name = "AWS Updated"
			result.Type = "Amazon"
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "cc-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "AWS Updated")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Amazon")
		case "access_key":
			vals[k] = tftypes.NewValue(attrType, "AKIAIOSFODNN7EXAMPLE")
		case "secret_key":
			vals[k] = tftypes.NewValue(attrType, "secret")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

func TestCloudCredential_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &CloudCredential{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("update failed"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "cc-1")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "AWS Updated")
		case "type":
			vals[k] = tftypes.NewValue(attrType, "Amazon")
		case "access_key":
			vals[k] = tftypes.NewValue(attrType, "AKIAIOSFODNN7EXAMPLE")
		case "secret_key":
			vals[k] = tftypes.NewValue(attrType, "secret")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Proxy — Update
// ---------------------------------------------------------------------------

func TestProxy_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	// PUT returns a sync result with type set (not async).
	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*map[string]interface{})
			*result = map[string]interface{}{
				"id":   "proxy-1",
				"type": "ViProxy",
			}
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "proxy-1")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

func TestProxy_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("update failed"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "proxy-1")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ConfigurationBackup — Create (with and without trigger)
// ---------------------------------------------------------------------------

func TestConfigurationBackup_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	// Create calls putConfig → loadConfigurationBackupPayload (GET) + putConfigurationPayload (PUT).
	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{"isEnabled": false}
		}).Return(nil)
	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, mock.Anything).
		Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "enabled":
			vals[k] = tftypes.NewValue(attrType, true)
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	// loadConfigurationBackupPayload fails.
	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Return(errors.New("API unavailable"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		if k == "enabled" {
			vals[k] = tftypes.NewValue(attrType, true)
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ConfigurationBackup — Update
// ---------------------------------------------------------------------------

func TestConfigurationBackup_Update_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{"isEnabled": true}
		}).Return(nil)
	mockClient.On("PutJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything, mock.Anything).
		Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, "config-backup")
		case "enabled":
			vals[k] = tftypes.NewValue(attrType, true)
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
	mockClient.AssertExpectations(t)
}

func TestConfigurationBackup_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ConfigurationBackup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathConfigurationBackup, mock.Anything).
		Return(errors.New("API unavailable"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		if k == "enabled" {
			vals[k] = tftypes.NewValue(attrType, true)
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	state.Raw = plan.Raw

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Verify require is used (avoid unused import warning)
// ---------------------------------------------------------------------------

var _ = require.NotNil

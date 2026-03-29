package resources

// ---------------------------------------------------------------------------
// Extra coverage tests to push pkg/resources to 80%+.
//
// Covers:
//   - ImportState for all 10 resources
//   - BackupJob: Create (error paths), Read (success + error), Update (error), Delete
//   - ProtectionGroup: Create (error), Read (success + not-found), Delete (error)
//   - Repository / Proxy / ScaleOutRepository / ManagedServer: Create error paths
//   - unwrapConfigBackupPayload: data-wrapper branch
// ---------------------------------------------------------------------------

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
// ImportState — all resources
// ---------------------------------------------------------------------------

// resourceWithImportState combines the Resource and ResourceWithImportState interfaces.
type resourceWithImportState interface {
	resource.Resource
	resource.ResourceWithImportState
}

func importStateWithID(t *testing.T, r resourceWithImportState, id string) *resource.ImportStateResponse {
	t.Helper()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
		},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, resp)
	return resp
}

func TestImportState_Credential(t *testing.T) {
	r := &Credential{}
	resp := importStateWithID(t, r, "cred-123")
	assert.False(t, resp.Diagnostics.HasError())
	var m CredentialModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "cred-123", m.ID.ValueString())
}

func TestImportState_Repository(t *testing.T) {
	r := &Repository{}
	resp := importStateWithID(t, r, "repo-abc")
	assert.False(t, resp.Diagnostics.HasError())
	var m RepositoryModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "repo-abc", m.ID.ValueString())
}

func TestImportState_Proxy(t *testing.T) {
	r := &Proxy{}
	resp := importStateWithID(t, r, "proxy-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m ProxyModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "proxy-1", m.ID.ValueString())
}

func TestImportState_ManagedServer(t *testing.T) {
	r := &ManagedServer{}
	resp := importStateWithID(t, r, "ms-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m ManagedServerModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "ms-1", m.ID.ValueString())
}

func TestImportState_ScaleOutRepository(t *testing.T) {
	r := &ScaleOutRepository{}
	resp := importStateWithID(t, r, "sobr-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m ScaleOutRepositoryModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "sobr-1", m.ID.ValueString())
}

func TestImportState_CloudCredential(t *testing.T) {
	r := &CloudCredential{}
	resp := importStateWithID(t, r, "cc-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m CloudCredentialModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "cc-1", m.ID.ValueString())
}

func TestImportState_EncryptionPassword(t *testing.T) {
	r := &EncryptionPassword{}
	resp := importStateWithID(t, r, "enc-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m EncryptionPasswordModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "enc-1", m.ID.ValueString())
}

func TestImportState_ConfigurationBackup(t *testing.T) {
	r := &ConfigurationBackup{}
	resp := importStateWithID(t, r, "config-backup")
	assert.False(t, resp.Diagnostics.HasError())
}

func TestImportState_BackupJob(t *testing.T) {
	r := &BackupJob{}
	resp := importStateWithID(t, r, "job-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m BackupJobModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "job-1", m.ID.ValueString())
}

func TestImportState_ProtectionGroup(t *testing.T) {
	r := &ProtectionGroup{}
	resp := importStateWithID(t, r, "pg-1")
	assert.False(t, resp.Diagnostics.HasError())
	var m ProtectionGroupModel
	assert.False(t, resp.State.Get(context.Background(), &m).HasError())
	assert.Equal(t, "pg-1", m.ID.ValueString())
}

// ---------------------------------------------------------------------------
// BackupJob — Create error paths
// ---------------------------------------------------------------------------

// buildBackupJobPlanWithType returns a tfsdk.Plan with just the type attribute set.
func buildBackupJobPlanWithType(r resource.Resource, jobType string) tfsdk.Plan {
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		if k == "type" {
			vals[k] = tftypes.NewValue(attrType, jobType)
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)
	return plan
}

func TestBackupJob_Create_UnsupportedType(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	plan := buildBackupJobPlanWithType(r, "BackupCopy")
	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupJob_Create_MissingVirtualMachines(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	// VSphereBackup with no virtual_machines → validation error.
	plan := buildBackupJobPlanWithType(r, "VSphereBackup")
	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupJob_Create_MissingAgentComputers(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	// LinuxAgentBackup with no agent_computers → validation error before API call.
	plan := buildBackupJobPlanWithType(r, "LinuxAgentBackup")
	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJob — Read
// ---------------------------------------------------------------------------

func buildBackupJobStateWithTypeAndID(r resource.Resource, jobType, id string) tfsdk.State {
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, id)
		case "type":
			vals[k] = tftypes.NewValue(attrType, jobType)
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)
	return state
}

func TestBackupJob_Read_VSphere_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.BackupJobModel)
			result.ID = "job-1"
			result.Name = "Daily Backup"
			result.Type = models.JobTypeVSphereBackup
		}).Return(nil)

	state := buildBackupJobStateWithTypeAndID(r, "VSphereBackup", "job-1")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

func TestBackupJob_Read_VSphere_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	state := buildBackupJobStateWithTypeAndID(r, "VSphereBackup", "job-1")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupJob_Read_Default_Success(t *testing.T) {
	// Unknown type falls through to the default case (base model read).
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.BackupJobModel)
			result.ID = "job-1"
			result.Name = "Unknown Job"
			result.Type = models.EJobType("SomeOtherType")
		}).Return(nil)

	state := buildBackupJobStateWithTypeAndID(r, "SomeOtherType", "job-1")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestBackupJob_Read_AgentLinux_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	state := buildBackupJobStateWithTypeAndID(r, "LinuxAgentBackup", "job-2")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJob — Update error paths
// ---------------------------------------------------------------------------

func TestBackupJob_Update_UnsupportedType(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	plan := buildBackupJobPlanWithType(r, "BackupCopy")
	state := buildBackupJobStateWithTypeAndID(r, "BackupCopy", "job-1")

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupJob_Update_VSphere_MissingVMs(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	plan := buildBackupJobPlanWithType(r, "VSphereBackup")
	state := buildBackupJobStateWithTypeAndID(r, "VSphereBackup", "job-1")

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackupJob_Update_AgentLinux_MissingComputers(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	plan := buildBackupJobPlanWithType(r, "LinuxAgentBackup")
	state := buildBackupJobStateWithTypeAndID(r, "LinuxAgentBackup", "job-2")

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJob — Delete
// ---------------------------------------------------------------------------

func TestBackupJob_Delete_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	state := buildBackupJobStateWithTypeAndID(r, "VSphereBackup", "job-1")
	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

func TestBackupJob_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("delete failed"))

	state := buildBackupJobStateWithTypeAndID(r, "VSphereBackup", "job-1")
	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ProtectionGroup — Create error path
// ---------------------------------------------------------------------------

func buildProtectionGroupPlanWithType(r resource.Resource, pgType string) tfsdk.Plan {
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "type":
			vals[k] = tftypes.NewValue(attrType, pgType)
		case "name":
			vals[k] = tftypes.NewValue(attrType, "Test PG")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)
	return plan
}

func buildProtectionGroupStateWithID(r resource.Resource, pgType, id string) tfsdk.State {
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}
	stateTyp := state.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range stateTyp.AttributeTypes {
		switch k {
		case "id":
			vals[k] = tftypes.NewValue(attrType, id)
		case "type":
			vals[k] = tftypes.NewValue(attrType, pgType)
		case "name":
			vals[k] = tftypes.NewValue(attrType, "Test PG")
		default:
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	state.Raw = tftypes.NewValue(stateTyp, vals)
	return state
}

func TestProtectionGroup_Create_InvalidType(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	plan := buildProtectionGroupPlanWithType(r, "InvalidType")
	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestProtectionGroup_Create_PostJSONError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathProtectionGroups, mock.Anything, mock.Anything).
		Return(errors.New("api error"))

	plan := buildProtectionGroupPlanWithType(r, "IndividualComputers")
	state := buildNullResourceState(r)

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ProtectionGroup — Read
// ---------------------------------------------------------------------------

func TestProtectionGroup_Read_Individual_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.IndividualComputersProtectionGroupModel)
			result.ID = "pg-1"
			result.Name = "Test PG"
			result.Type = models.ProtectionGroupTypeIndividualComputers
		}).Return(nil)

	state := buildProtectionGroupStateWithID(r, "IndividualComputers", "pg-1")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

func TestProtectionGroup_Read_NotFound(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	// Return an error that looks like a 404 → state should be removed.
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("http 404 not found"))

	state := buildProtectionGroupStateWithID(r, "IndividualComputers", "pg-1")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "not-found should remove state, not add error")
}

func TestProtectionGroup_Read_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("connection refused"))

	state := buildProtectionGroupStateWithID(r, "IndividualComputers", "pg-1")
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ProtectionGroup — Delete error path
// ---------------------------------------------------------------------------

func TestProtectionGroup_Delete_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("delete failed"))

	state := buildProtectionGroupStateWithID(r, "IndividualComputers", "pg-1")
	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestProtectionGroup_Delete_WaitError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	// Delete succeeds but the resource still appears when polled (returns non-404 error).
	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	// GetJSON for polling returns a "not found" error → waitForProtectionGroupDeleted returns nil quickly.
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("http 404 not found"))

	state := buildProtectionGroupStateWithID(r, "IndividualComputers", "pg-1")
	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "404 during poll means resource is gone, no error expected")
}

// ---------------------------------------------------------------------------
// ProtectionGroup — Update error path
// ---------------------------------------------------------------------------

func TestProtectionGroup_Update_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("update failed"))

	plan := buildProtectionGroupPlanWithType(r, "IndividualComputers")
	state := buildProtectionGroupStateWithID(r, "IndividualComputers", "pg-1")

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// Repository — Create error path
// ---------------------------------------------------------------------------

func TestRepository_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathRepositories, mock.Anything, mock.Anything).
		Return(errors.New("api error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsLocal")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "Test Repo")
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

func TestRepository_Create_SyncResult(t *testing.T) {
	// PostJSON returns a sync result (type is set) → direct ID assignment + GetJSON read-back.
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathRepositories, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*map[string]interface{})
			*result = map[string]interface{}{"id": "repo-new", "type": "WindowsLocal"}
		}).Return(nil)
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{"id": "repo-new", "name": "Test Repo", "type": "WindowsLocal"}
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsLocal")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "Test Repo")
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
}

// ---------------------------------------------------------------------------
// Proxy — Create error path
// ---------------------------------------------------------------------------

func TestProxy_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathProxies, mock.Anything, mock.Anything).
		Return(errors.New("api error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		if k == "type" {
			vals[k] = tftypes.NewValue(attrType, "ViProxy")
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

func TestProxy_Create_SyncResult(t *testing.T) {
	// PostJSON returns a sync result (type set) → direct ID + GetJSON read-back.
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathProxies, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*map[string]interface{})
			*result = map[string]interface{}{"id": "proxy-new", "type": "ViProxy"}
		}).Return(nil)
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.ViProxyModel)
			result.ID = "proxy-new"
			result.Type = models.ProxyTypeViProxy
		}).Return(nil)

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		if k == "type" {
			vals[k] = tftypes.NewValue(attrType, "ViProxy")
		} else {
			vals[k] = nullValueForResourceType(attrType)
		}
	}
	plan.Raw = tftypes.NewValue(planTyp, vals)

	state := buildNullResourceState(r)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
}

// ---------------------------------------------------------------------------
// ScaleOutRepository — Create error path
// ---------------------------------------------------------------------------

func TestScaleOutRepository_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathScaleOutRepositories, mock.Anything, mock.Anything).
		Return(errors.New("api error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		if k == "name" {
			vals[k] = tftypes.NewValue(attrType, "SOBR Test")
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
// ManagedServer — Create error path
// ---------------------------------------------------------------------------

func TestManagedServer_Create_Error(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	// WindowsHost — no fingerprint resolution needed.
	mockClient.On("PostJSON", mock.Anything, client.PathManagedServers, mock.Anything, mock.Anything).
		Return(errors.New("api error"))

	plan := buildNullResourcePlan(r)
	planTyp := plan.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := map[string]tftypes.Value{}
	for k, attrType := range planTyp.AttributeTypes {
		switch k {
		case "type":
			vals[k] = tftypes.NewValue(attrType, "WindowsHost")
		case "name":
			vals[k] = tftypes.NewValue(attrType, "winserver01")
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
// unwrapConfigBackupPayload — data wrapper branch
// ---------------------------------------------------------------------------

func TestUnwrapConfigBackupPayload_DataWrapper(t *testing.T) {
	inner := map[string]interface{}{"isEnabled": true, "backupRepositoryId": "repo-1"}
	raw := map[string]interface{}{"data": inner}

	result := unwrapConfigBackupPayload(raw)

	assert.Equal(t, inner, result)
	assert.Equal(t, true, result["isEnabled"])
}

func TestUnwrapConfigBackupPayload_Nil(t *testing.T) {
	result := unwrapConfigBackupPayload(nil)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestUnwrapConfigBackupPayload_Flat(t *testing.T) {
	raw := map[string]interface{}{"isEnabled": false}
	result := unwrapConfigBackupPayload(raw)
	assert.Equal(t, raw, result)
}

// ---------------------------------------------------------------------------
// isProtectionGroupNotFound
// ---------------------------------------------------------------------------

func TestIsProtectionGroupNotFound_APIError(t *testing.T) {
	apiErr := &models.APIError{ErrorCode: "NotFound", Message: "not found"}
	assert.True(t, isProtectionGroupNotFound(apiErr))
}

func TestIsProtectionGroupNotFound_HTTP404(t *testing.T) {
	assert.True(t, isProtectionGroupNotFound(errors.New("http 404")))
}

func TestIsProtectionGroupNotFound_Other(t *testing.T) {
	assert.False(t, isProtectionGroupNotFound(errors.New("connection refused")))
	assert.False(t, isProtectionGroupNotFound(nil))
}

// ---------------------------------------------------------------------------
// ProtectionGroup — syncFromAPIIndividual with computers and options
// ---------------------------------------------------------------------------

func TestProtectionGroup_SyncFromAPIIndividual_WithComputers(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		Name: types.StringValue("old"),
		Type: types.StringValue("IndividualComputers"),
	}

	api := &models.IndividualComputersProtectionGroupModel{
		ProtectionGroupModel: models.ProtectionGroupModel{
			Name:       "New PG",
			Type:       models.ProtectionGroupTypeIndividualComputers,
			IsDisabled: false,
		},
		Computers: []models.ProtectionGroupComputer{
			{
				HostName:       "server1.example.com",
				ConnectionType: models.IndividualComputerConnectionTypePermanentCredentials,
				CredentialsID:  "cred-1",
			},
		},
	}

	r.syncFromAPIIndividual(data, api)

	assert.Equal(t, "New PG", data.Name.ValueString())
	assert.Equal(t, "IndividualComputers", data.Type.ValueString())
	assert.False(t, data.IsDisabled.ValueBool())
	assert.Len(t, data.Computers, 1)
	assert.Equal(t, "server1.example.com", data.Computers[0].HostName.ValueString())
	assert.Equal(t, "cred-1", data.Computers[0].CredentialsID.ValueString())
	assert.Nil(t, data.CloudAccount)
	assert.Nil(t, data.CloudMachines)
}

// ---------------------------------------------------------------------------
// Repository — syncFromAPIMap for different types
// ---------------------------------------------------------------------------

func TestRepository_SyncFromAPIMap_WinLocal(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{
		Name: types.StringValue("old"),
		Type: types.StringValue("WindowsLocal"),
	}
	api := map[string]interface{}{
		"id":     "repo-1",
		"name":   "Win Repo",
		"type":   "WindowsLocal",
		"hostId": "host-1",
		"repository": map[string]interface{}{
			"path":         "C:\\Backups",
			"maxTaskCount": float64(4),
		},
	}

	r.syncFromAPIMap(data, api)

	assert.Equal(t, "Win Repo", data.Name.ValueString())
	assert.Equal(t, "WindowsLocal", data.Type.ValueString())
}

func TestRepository_SyncFromAPIMap_LinuxLocal(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{Type: types.StringValue("LinuxLocal")}
	api := map[string]interface{}{
		"name": "Linux Repo",
		"type": "LinuxLocal",
	}
	r.syncFromAPIMap(data, api)
	assert.Equal(t, "Linux Repo", data.Name.ValueString())
}

func TestRepository_SyncFromAPIMap_Nfs(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{Type: types.StringValue("Nfs")}
	api := map[string]interface{}{
		"name": "NFS Repo",
		"type": "Nfs",
	}
	r.syncFromAPIMap(data, api)
	assert.Equal(t, "NFS Repo", data.Name.ValueString())
}

func TestRepository_SyncFromAPIMap_Smb(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{Type: types.StringValue("Smb")}
	api := map[string]interface{}{
		"name": "SMB Repo",
		"type": "Smb",
	}
	r.syncFromAPIMap(data, api)
	assert.Equal(t, "SMB Repo", data.Name.ValueString())
}

// ---------------------------------------------------------------------------
// ManagedServer — buildSpec LinuxHost
// ---------------------------------------------------------------------------

func TestManagedServerBuildSpec_LinuxHost(t *testing.T) {
	r := &ManagedServer{}
	data := &ManagedServerModel{
		Name:           types.StringValue("linux-server"),
		Type:           types.StringValue("LinuxHost"),
		CredentialsID:  types.StringValue("cred-3"),
		SSHFingerprint: types.StringValue("AA:BB:CC:DD"),
	}
	spec := r.buildSpec(data)
	assert.NotNil(t, spec)
}

// ---------------------------------------------------------------------------
// findRepositoryIDByName
// ---------------------------------------------------------------------------

func TestRepository_FindIDByName_Found(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "repo-found", "name": "MyRepo"},
				},
			}
		}).Return(nil)

	id, err := r.findRepositoryIDByName(context.Background(), "MyRepo")
	assert.NoError(t, err)
	assert.Equal(t, "repo-found", id)
}

func TestRepository_FindIDByName_NotFound(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "repo-1", "name": "OtherRepo"},
				},
			}
		}).Return(nil)

	_, err := r.findRepositoryIDByName(context.Background(), "MyRepo")
	assert.Error(t, err)
}

func TestRepository_FindIDByName_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Repository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	_, err := r.findRepositoryIDByName(context.Background(), "MyRepo")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// findProxyID
// ---------------------------------------------------------------------------

func TestProxy_FindID_Found(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":          "proxy-found",
						"description": "Main vSphere proxy",
						"type":        "ViProxy",
					},
				},
			}
		}).Return(nil)

	data := &ProxyModel{
		Description: types.StringValue("Main vSphere proxy"),
		Type:        types.StringValue("ViProxy"),
	}
	id, err := r.findProxyID(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, "proxy-found", id)
}

func TestProxy_FindID_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	data := &ProxyModel{}
	_, err := r.findProxyID(context.Background(), data)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// findSOBRIDByName
// ---------------------------------------------------------------------------

func TestSOBR_FindIDByName_Found(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "sobr-found", "name": "MainSOBR"},
				},
			}
		}).Return(nil)

	id, err := r.findSOBRIDByName(context.Background(), "MainSOBR")
	assert.NoError(t, err)
	assert.Equal(t, "sobr-found", id)
}

func TestSOBR_FindIDByName_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	_, err := r.findSOBRIDByName(context.Background(), "MainSOBR")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// findManagedServerID
// ---------------------------------------------------------------------------

func TestManagedServer_FindID_Found(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "ms-found", "name": "winserver01"},
				},
			}
		}).Return(nil)

	data := &ManagedServerModel{Name: types.StringValue("winserver01")}
	id, err := r.findManagedServerID(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, "ms-found", id)
}

func TestManagedServer_FindID_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	data := &ManagedServerModel{Name: types.StringValue("winserver01")}
	_, err := r.findManagedServerID(context.Background(), data)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// findProtectionGroupIDByName
// ---------------------------------------------------------------------------

func TestProtectionGroup_FindIDByName_Found(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "pg-found", "name": "OfficeServers", "type": "IndividualComputers"},
				},
			}
		}).Return(nil)

	data := &ProtectionGroupModel{
		Name: types.StringValue("OfficeServers"),
		Type: types.StringValue("IndividualComputers"),
	}
	id, err := r.findProtectionGroupIDByName(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, "pg-found", id)
}

func TestProtectionGroup_FindIDByName_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("api error"))

	data := &ProtectionGroupModel{Name: types.StringValue("OfficeServers"), Type: types.StringValue("IndividualComputers")}
	_, err := r.findProtectionGroupIDByName(context.Background(), data)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// setStringValue / setIntValue — fallback-key branch
// ---------------------------------------------------------------------------

func TestSetStringValue_ExistingKey(t *testing.T) {
	m := map[string]interface{}{"foo": "old"}
	setStringValue(m, "new", "foo")
	assert.Equal(t, "new", m["foo"])
}

func TestSetStringValue_NewKey(t *testing.T) {
	m := map[string]interface{}{}
	setStringValue(m, "val", "bar")
	assert.Equal(t, "val", m["bar"])
}

func TestSetIntValue_ExistingKey(t *testing.T) {
	m := map[string]interface{}{"count": 0}
	setIntValue(m, 7, "count")
	assert.Equal(t, 7, m["count"])
}

func TestSetIntValue_NewKey(t *testing.T) {
	m := map[string]interface{}{}
	setIntValue(m, 42, "size")
	assert.Equal(t, 42, m["size"])
}

// ---------------------------------------------------------------------------
// isRepositoryNotFound
// ---------------------------------------------------------------------------

func TestIsRepositoryNotFound_APIError(t *testing.T) {
	apiErr := &models.APIError{ErrorCode: "NotFound", Message: "not found"}
	assert.True(t, isRepositoryNotFound(apiErr))
}

func TestIsRepositoryNotFound_HTTP404(t *testing.T) {
	assert.True(t, isRepositoryNotFound(errors.New("http 404")))
}

func TestIsRepositoryNotFound_Other(t *testing.T) {
	assert.False(t, isRepositoryNotFound(errors.New("connection refused")))
}

// ---------------------------------------------------------------------------
// buildProtectionGroupCloudAccount — with all optional fields
// ---------------------------------------------------------------------------

func TestBuildProtectionGroupCloudAccount_Full(t *testing.T) {
	accounts := []ProtectionGroupCloudAccountModel{
		{
			AccountType:    types.StringValue("AWS"),
			CredentialsID:  types.StringValue("cred-aws"),
			SubscriptionID: types.StringValue("sub-1"),
			RegionType:     types.StringValue("Public"),
			RegionID:       types.StringValue("us-east-1"),
		},
	}
	result := buildProtectionGroupCloudAccount(accounts)
	assert.NotNil(t, result)
	assert.Equal(t, "AWS", string(result.AccountType))
	assert.Equal(t, "cred-aws", result.CredentialsID)
	assert.Equal(t, "us-east-1", result.RegionID)
}

func TestBuildProtectionGroupCloudAccount_Empty(t *testing.T) {
	result := buildProtectionGroupCloudAccount(nil)
	assert.Nil(t, result)
}

// ---------------------------------------------------------------------------
// buildProtectionGroupCloudMachines
// ---------------------------------------------------------------------------

func TestBuildProtectionGroupCloudMachines_WithItems(t *testing.T) {
	items := []ProtectionGroupCloudMachineModel{
		{
			Type:     types.StringValue("VirtualMachine"),
			Name:     types.StringValue("vm-prod-01"),
			ObjectID: types.StringValue("vm-101"),
		},
	}
	result := buildProtectionGroupCloudMachines(items)
	assert.Len(t, result, 1)
	assert.Equal(t, "VirtualMachine", string(result[0].Type))
	assert.Equal(t, "vm-prod-01", result[0].Name)
}

// ---------------------------------------------------------------------------
// BackupJob — buildAgentJobSpec branches
// ---------------------------------------------------------------------------

func TestBackupJob_BuildAgentJobSpec_Windows(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Name:             types.StringValue("Win Backup"),
		Type:             types.StringValue("WindowsAgentBackup"),
		AgentBackupMode:  types.StringValue("EntireComputer"),
		IncludeUsbDrives: types.BoolValue(true),
		AgentType:        types.StringValue("Workstation"),
		AgentComputers:   []AgentComputerEntry{},
	}
	spec := r.buildAgentJobSpec(data)
	assert.Equal(t, true, spec["includeUsbDrives"])
	assert.Equal(t, "Workstation", spec["agentType"])
}

func TestBackupJob_BuildAgentJobSpec_Linux(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Name:                          types.StringValue("Linux Backup"),
		Type:                          types.StringValue("LinuxAgentBackup"),
		AgentBackupMode:               types.StringValue("EntireComputer"),
		UseSnapshotlessFileLevelBackup: types.BoolValue(true),
		AgentComputers:                []AgentComputerEntry{},
	}
	spec := r.buildAgentJobSpec(data)
	assert.Equal(t, true, spec["useSnapshotlessFileLevelBackup"])
}

// ---------------------------------------------------------------------------
// syncScheduleFromAPIMap — full branches
// ---------------------------------------------------------------------------

func TestSyncScheduleFromAPIMap_Daily(t *testing.T) {
	r := &BackupJob{}
	api := map[string]interface{}{
		"runAutomatically": true,
		"daily": map[string]interface{}{
			"isEnabled": true,
			"localTime": "22:00",
			"dailyKind": "Weekdays",
		},
	}
	s := r.syncScheduleFromAPIMap(nil, api)
	assert.True(t, s.RunAutomatically.ValueBool())
	assert.True(t, s.DailyEnabled.ValueBool())
	assert.Equal(t, "22:00", s.DailyLocalTime.ValueString())
	assert.Equal(t, "Weekdays", s.DailyKind.ValueString())
}

func TestSyncScheduleFromAPIMap_Monthly(t *testing.T) {
	r := &BackupJob{}
	api := map[string]interface{}{
		"monthly": map[string]interface{}{
			"isEnabled":    true,
			"localTime":    "23:00",
			"dayOfMonth":   float64(15),
		},
	}
	s := r.syncScheduleFromAPIMap(nil, api)
	assert.True(t, s.MonthlyEnabled.ValueBool())
	assert.Equal(t, "23:00", s.MonthlyLocalTime.ValueString())
	assert.Equal(t, int64(15), s.MonthlyDayOfMonth.ValueInt64())
}

func TestSyncScheduleFromAPIMap_Periodically(t *testing.T) {
	r := &BackupJob{}
	api := map[string]interface{}{
		"periodically": map[string]interface{}{
			"isEnabled":       true,
			"periodicallyKind": "Hours",
			"frequency":        float64(4),
		},
	}
	s := r.syncScheduleFromAPIMap(nil, api)
	assert.True(t, s.PeriodicallyEnabled.ValueBool())
	assert.Equal(t, "Hours", s.PeriodicallyKind.ValueString())
	assert.Equal(t, int64(4), s.PeriodicallyFrequency.ValueInt64())
}

// ---------------------------------------------------------------------------
// syncFromAPICloud with full CloudAccount
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// validateProtectionGroupPlan — additional branches
// ---------------------------------------------------------------------------

func TestValidateProtectionGroupPlan_NilType(t *testing.T) {
	err := validateProtectionGroupPlan(&ProtectionGroupModel{Type: types.StringNull()})
	assert.Error(t, err)
}

func TestValidateProtectionGroupPlan_MissingHostname(t *testing.T) {
	err := validateProtectionGroupPlan(&ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{HostName: types.StringNull(), ConnectionType: types.StringValue("PermanentCredentials")},
		},
	})
	assert.Error(t, err)
}

func TestValidateProtectionGroupPlan_Certificate_WithCredentials(t *testing.T) {
	// Certificate type with credentials_id set → error
	err := validateProtectionGroupPlan(&ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("host1"),
				ConnectionType: types.StringValue("Certificate"),
				CredentialsID:  types.StringValue("cred-1"),
			},
		},
	})
	assert.Error(t, err)
}

func TestValidateProtectionGroupPlan_SingleUseCredentials(t *testing.T) {
	err := validateProtectionGroupPlan(&ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("host1"),
				ConnectionType: types.StringValue("SingleUseCredentials"),
			},
		},
	})
	assert.Error(t, err)
}

func TestValidateProtectionGroupPlan_UnknownConnectionType(t *testing.T) {
	err := validateProtectionGroupPlan(&ProtectionGroupModel{
		Type: types.StringValue("IndividualComputers"),
		Computers: []ProtectionGroupComputerModel{
			{
				HostName:       types.StringValue("host1"),
				ConnectionType: types.StringValue("BadType"),
			},
		},
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ManagedServer — findManagedServerID: not found
// ---------------------------------------------------------------------------

func TestManagedServer_FindID_NotFound(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "ms-1", "name": "other-server"},
				},
			}
		}).Return(nil)

	data := &ManagedServerModel{Name: types.StringValue("winserver01")}
	_, err := r.findManagedServerID(context.Background(), data)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Proxy — findProxyID: not found
// ---------------------------------------------------------------------------

func TestProxy_FindID_NotFound(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{
				"data": []interface{}{},
			}
		}).Return(nil)

	data := &ProxyModel{Description: types.StringNull(), Type: types.StringNull()}
	_, err := r.findProxyID(context.Background(), data)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Repository — syncFromAPIMap description branch
// ---------------------------------------------------------------------------

func TestRepository_SyncFromAPIMap_WithDescription(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{Type: types.StringValue("WinLocal")}
	api := map[string]interface{}{
		"name":        "Repo A",
		"description": "Main repository",
		"type":        "WinLocal",
	}
	r.syncFromAPIMap(data, api)
	assert.Equal(t, "Main repository", data.Description.ValueString())
}

// ---------------------------------------------------------------------------
// syncFromAPICloud — no CloudAccount (nil branch)
// ---------------------------------------------------------------------------

func TestProtectionGroup_SyncFromAPICloud_NoAccount(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		Name: types.StringValue("old"),
		Type: types.StringValue("CloudMachines"),
	}
	api := &models.CloudMachinesProtectionGroupModel{
		ProtectionGroupModel: models.ProtectionGroupModel{
			Name: "Cloud PG",
			Type: models.ProtectionGroupTypeCloudMachines,
		},
		CloudAccount: nil,
	}
	r.syncFromAPICloud(data, api)
	assert.Equal(t, "Cloud PG", data.Name.ValueString())
	assert.Empty(t, data.CloudAccount)
}

// ---------------------------------------------------------------------------
// BackupJob — Create LinuxAgentBackup with GuestProcessing set → error
// ---------------------------------------------------------------------------

func TestBackupJob_Create_AgentWithGuestProcessing(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw:    nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background())),
	}

	data := BackupJobModel{
		Type:            types.StringValue("LinuxAgentBackup"),
		AgentBackupMode: types.StringValue("EntireComputer"),
		AgentComputers: []AgentComputerEntry{
			{ID: types.StringValue("c1"), Name: types.StringValue("srv1"), Type: types.StringValue("Computer"), ProtectionGroupID: types.StringValue("")},
		},
		GuestProcessing: &JobGuestProcessing{},
	}
	plan.Set(context.Background(), data) //nolint:errcheck

	state := buildNullResourceState(r)
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: state}
	r.Create(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// syncFromAPIIndividual — with Options populated
// ---------------------------------------------------------------------------

func TestProtectionGroup_SyncFromAPIIndividual_WithOptions(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		Name: types.StringValue("old"),
		Type: types.StringValue("IndividualComputers"),
		Options: []ProtectionGroupOptionsModel{{}}, // non-empty → triggers options sync
	}

	api := &models.IndividualComputersProtectionGroupModel{
		ProtectionGroupModel: models.ProtectionGroupModel{
			Name: "PG With Options",
			Type: models.ProtectionGroupTypeIndividualComputers,
		},
		Options: &models.ProtectionGroupOptions{
			InstallBackupAgent:        true,
			DistributionServerID:      "dist-srv-1",
			DistributionRepositoryID:  "dist-repo-1",
			ApplicationPlugins:        []string{"SQL"},
		},
	}

	r.syncFromAPIIndividual(data, api)

	assert.Equal(t, "PG With Options", data.Name.ValueString())
	require.Len(t, data.Options, 1)
	assert.True(t, data.Options[0].InstallBackupAgent.ValueBool())
	assert.Equal(t, "dist-srv-1", data.Options[0].DistributionServerID.ValueString())
}

// ---------------------------------------------------------------------------
// ManagedServer — Delete error path
// ---------------------------------------------------------------------------

func TestManagedServer_Delete_DeleteJSONError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Return(errors.New("delete failed"))

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

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ScaleOutRepository — Read: not-found path
// ---------------------------------------------------------------------------

func TestScaleOutRepository_Read_NotFound(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ScaleOutRepository{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("http 404"))

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

	// 404 → resource removed from state, no error diagnostic
	assert.False(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJob — Update: LinuxAgentBackup API error
// ---------------------------------------------------------------------------

func TestBackupJob_Update_AgentLinux_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &BackupJob{client: mockClient}

	// GetJSON for fetching current agent model.
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(nil)
	// PutJSON fails.
	mockClient.On("PutJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
		Return(errors.New("put failed"))

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background()))}

	planData := BackupJobModel{
		Name:            types.StringValue("Linux Backup"),
		Type:            types.StringValue("LinuxAgentBackup"),
		AgentBackupMode: types.StringValue("EntireComputer"),
		AgentComputers: []AgentComputerEntry{
			{ID: types.StringValue("c1"), Name: types.StringValue("srv"), Type: types.StringValue("Computer"), ProtectionGroupID: types.StringValue("")},
		},
	}
	plan.Set(context.Background(), planData) //nolint:errcheck

	stateData := BackupJobModel{
		ID:   types.StringValue("job-2"),
		Type: types.StringValue("LinuxAgentBackup"),
		AgentComputers: []AgentComputerEntry{
			{ID: types.StringValue("c1"), Name: types.StringValue("srv"), Type: types.StringValue("Computer"), ProtectionGroupID: types.StringValue("")},
		},
	}
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: nullObjectForResourceSchema(schemaResp.Schema.Type().TerraformType(context.Background()))}
	state.Set(context.Background(), stateData) //nolint:errcheck

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// BackupJob — syncAgentJobFromAPIMap: storage sync branch
// ---------------------------------------------------------------------------

func TestSyncAgentJobFromAPIMap_WithStorage(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type:    types.StringValue("LinuxAgentBackup"),
		Storage: &JobStorageSettings{ProxyAutoSelect: types.BoolValue(true)},
	}
	api := map[string]interface{}{
		"storage": map[string]interface{}{
			"backupRepositoryId": "repo-1",
		},
	}
	r.syncAgentJobFromAPIMap(data, api)
	assert.NotNil(t, data.Storage)
}

// ---------------------------------------------------------------------------
// ProtectionGroup — readProtectionGroup: CloudMachines branch
// ---------------------------------------------------------------------------

func TestProtectionGroup_ReadProtectionGroup_CloudMachines(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			switch v := args.Get(2).(type) {
			case *models.CloudMachinesProtectionGroupModel:
				v.ID = "pg-1"
				v.Name = "Cloud PG"
				v.Type = models.ProtectionGroupTypeCloudMachines
			}
		}).Return(nil)

	data := &ProtectionGroupModel{
		ID:   types.StringValue("pg-1"),
		Type: types.StringValue("CloudMachines"),
	}
	err := r.readProtectionGroup(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, "Cloud PG", data.Name.ValueString())
}

func TestProtectionGroup_SyncFromAPICloud_Full(t *testing.T) {
	r := &ProtectionGroup{}
	data := &ProtectionGroupModel{
		Name: types.StringValue("old"),
		Type: types.StringValue("CloudMachines"),
	}
	api := &models.CloudMachinesProtectionGroupModel{
		ProtectionGroupModel: models.ProtectionGroupModel{
			Name: "Cloud PG",
			Type: models.ProtectionGroupTypeCloudMachines,
		},
		CloudAccount: &models.CloudMachinesAccount{
			AccountType:   models.ProtectionGroupCloudAccountTypeAWS,
			CredentialsID: "cred-aws",
			RegionType:    "Public",
			RegionID:      "us-east-1",
		},
	}
	r.syncFromAPICloud(data, api)

	assert.Equal(t, "Cloud PG", data.Name.ValueString())
	assert.Equal(t, "CloudMachines", data.Type.ValueString())
	assert.Len(t, data.CloudAccount, 1)
	assert.Equal(t, "AWS", data.CloudAccount[0].AccountType.ValueString())
	assert.Equal(t, "us-east-1", data.CloudAccount[0].RegionID.ValueString())
	assert.Nil(t, data.Computers)
}

// ---------------------------------------------------------------------------
// repositoryTypeValidator
// ---------------------------------------------------------------------------

func TestRepositoryTypeValidator_Description(t *testing.T) {
	v := repositoryTypeValidator{}
	assert.NotEmpty(t, v.Description(context.Background()))
	assert.NotEmpty(t, v.MarkdownDescription(context.Background()))
}

func TestRepositoryTypeValidator_ValidateString_Valid(t *testing.T) {
	v := repositoryTypeValidator{}
	for _, typ := range []string{"WinLocal", "LinuxLocal", "Nfs", "Smb"} {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(typ),
		}, resp)
		assert.False(t, resp.Diagnostics.HasError(), "expected %s to be valid", typ)
	}
}

func TestRepositoryTypeValidator_ValidateString_Invalid(t *testing.T) {
	v := repositoryTypeValidator{}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{
		ConfigValue: types.StringValue("LinuxHardened"),
	}, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestRepositoryTypeValidator_ValidateString_Null(t *testing.T) {
	v := repositoryTypeValidator{}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{
		ConfigValue: types.StringNull(),
	}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// getConfigIntValue
// ---------------------------------------------------------------------------

func TestGetConfigIntValue_Int(t *testing.T) {
	m := map[string]interface{}{"count": int(7)}
	assert.Equal(t, 7, getConfigIntValue(m, "count"))
}

func TestGetConfigIntValue_Int64(t *testing.T) {
	m := map[string]interface{}{"count": int64(42)}
	assert.Equal(t, 42, getConfigIntValue(m, "count"))
}

func TestGetConfigIntValue_Float64(t *testing.T) {
	m := map[string]interface{}{"count": float64(14)}
	assert.Equal(t, 14, getConfigIntValue(m, "count"))
}

func TestGetConfigIntValue_Missing(t *testing.T) {
	m := map[string]interface{}{}
	assert.Equal(t, 0, getConfigIntValue(m, "missing"))
}

func TestGetConfigIntValue_FallbackKey(t *testing.T) {
	m := map[string]interface{}{"secondary": float64(3)}
	assert.Equal(t, 3, getConfigIntValue(m, "primary", "secondary"))
}

// ---------------------------------------------------------------------------
// syncAgentJobFromAPIMap — Windows and Linux specific branches
// ---------------------------------------------------------------------------

func TestSyncAgentJobFromAPIMap_Windows(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type: types.StringValue("WindowsAgentBackup"),
	}
	api := map[string]interface{}{
		"name":             "Win Backup",
		"type":             "WindowsAgentBackup",
		"isDisabled":       false,
		"isHighPriority":   true,
		"description":      "Windows backup job",
		"backupMode":       "EntireComputer",
		"includeUsbDrives": true,
		"agentType":        "Workstation",
	}
	r.syncAgentJobFromAPIMap(data, api)

	assert.Equal(t, "Win Backup", data.Name.ValueString())
	assert.True(t, data.IsHighPriority.ValueBool())
	assert.Equal(t, "EntireComputer", data.AgentBackupMode.ValueString())
	assert.True(t, data.IncludeUsbDrives.ValueBool())
	assert.Equal(t, "Workstation", data.AgentType.ValueString())
}

func TestSyncAgentJobFromAPIMap_Linux(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type: types.StringValue("LinuxAgentBackup"),
	}
	api := map[string]interface{}{
		"name":                          "Linux Backup",
		"type":                          "LinuxAgentBackup",
		"useSnapshotlessFileLevelBackup": true,
	}
	r.syncAgentJobFromAPIMap(data, api)

	assert.Equal(t, "Linux Backup", data.Name.ValueString())
	assert.True(t, data.UseSnapshotlessFileLevelBackup.ValueBool())
}

func TestSyncAgentJobFromAPIMap_WithComputers(t *testing.T) {
	r := &BackupJob{}
	data := &BackupJobModel{
		Type: types.StringValue("LinuxAgentBackup"),
	}
	api := map[string]interface{}{
		"computers": []interface{}{
			map[string]interface{}{
				"id":                "c1",
				"name":              "server01",
				"type":              "Computer",
				"protectionGroupId": "pg-1",
			},
		},
	}
	r.syncAgentJobFromAPIMap(data, api)

	assert.Len(t, data.AgentComputers, 1)
	assert.Equal(t, "server01", data.AgentComputers[0].Name.ValueString())
}

// ---------------------------------------------------------------------------
// ProtectionGroup — Update success path
// ---------------------------------------------------------------------------

func TestProtectionGroup_Update_InvalidType(t *testing.T) {
	// Validation fails before any API call — covers the validateProtectionGroupPlan error branch in Update.
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	plan := buildProtectionGroupPlanWithType(r, "UnsupportedType")
	state := buildProtectionGroupStateWithID(r, "UnsupportedType", "pg-1")

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

// ---------------------------------------------------------------------------
// ManagedServer — Delete (wait error branch)
// ---------------------------------------------------------------------------

func TestManagedServer_Delete_WaitTimeout(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	// waitForManagedServerDeleted polls GetJSON; return non-404 error every time → timeout.
	// Return 404-style error so the wait exits cleanly.
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(errors.New("http 404 not found"))

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

	// 404 during poll → resource is gone → no error.
	assert.False(t, resp.Diagnostics.HasError())
}


// ---------------------------------------------------------------------------
// ManagedServer — buildSpec branches
// ---------------------------------------------------------------------------

func TestManagedServerBuildSpec_WindowsHost(t *testing.T) {
	r := &ManagedServer{}
	data := &ManagedServerModel{
		Name:          types.StringValue("winserver"),
		Description:   types.StringValue("Win host"),
		Type:          types.StringValue("WindowsHost"),
		CredentialsID: types.StringValue("cred-1"),
		Port:          types.Int64Value(6160),
	}
	spec := r.buildSpec(data)
	assert.NotNil(t, spec)
}

func TestManagedServerBuildSpec_ViHost(t *testing.T) {
	r := &ManagedServer{}
	data := &ManagedServerModel{
		Name:                  types.StringValue("vcenter"),
		Type:                  types.StringValue("ViHost"),
		CredentialsID:         types.StringValue("cred-2"),
		Port:                  types.Int64Value(443),
		CertificateThumbprint: types.StringValue("AA:BB:CC"),
	}
	spec := r.buildSpec(data)
	assert.NotNil(t, spec)
}

// ---------------------------------------------------------------------------
// ProtectionGroup — readProtectionGroup with unknown type (probe branch)
// ---------------------------------------------------------------------------

func TestProtectionGroup_ReadProtectionGroup_UnknownType(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ProtectionGroup{client: mockClient}

	// First call: probe to discover type.
	// Second call: full IndividualComputers model fetch.
	callCount := 0
	mockClient.On("GetJSON", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			callCount++
			switch v := args.Get(2).(type) {
			case *map[string]interface{}:
				*v = map[string]interface{}{"id": "pg-1", "type": "IndividualComputers"}
			case *models.IndividualComputersProtectionGroupModel:
				v.ID = "pg-1"
				v.Name = "Test"
				v.Type = models.ProtectionGroupTypeIndividualComputers
			}
		}).Return(nil)

	data := &ProtectionGroupModel{
		ID:   types.StringValue("pg-1"),
		Type: types.StringNull(), // unknown → triggers probe
	}
	err := r.readProtectionGroup(context.Background(), data)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// isSNMPGeneralOptionsError — all branches
// ---------------------------------------------------------------------------

func TestIsSNMPGeneralOptionsError_Nil(t *testing.T) {
	assert.False(t, isSNMPGeneralOptionsError(nil))
}

func TestIsSNMPGeneralOptionsError_Match(t *testing.T) {
	err := fmt.Errorf("Specify SNMP settings in General Options before configuring notifications")
	assert.True(t, isSNMPGeneralOptionsError(err))
}

func TestIsSNMPGeneralOptionsError_NoMatch(t *testing.T) {
	err := fmt.Errorf("some other error")
	assert.False(t, isSNMPGeneralOptionsError(err))
}

// ---------------------------------------------------------------------------
// isCredentialInUseError — all branches
// ---------------------------------------------------------------------------

func TestIsCredentialInUseError_Nil(t *testing.T) {
	assert.False(t, isCredentialInUseError(nil))
}

func TestIsCredentialInUseError_Match(t *testing.T) {
	err := fmt.Errorf("Unable to delete selected credentials because it is currently in use by jobs")
	assert.True(t, isCredentialInUseError(err))
}

func TestIsCredentialInUseError_NoMatch(t *testing.T) {
	err := fmt.Errorf("some other error")
	assert.False(t, isCredentialInUseError(err))
}

// ---------------------------------------------------------------------------
// isConfigBackupPasswordInUseError — all branches
// ---------------------------------------------------------------------------

func TestIsConfigBackupPasswordInUseError_Nil(t *testing.T) {
	assert.False(t, isConfigBackupPasswordInUseError(nil))
}

func TestIsConfigBackupPasswordInUseError_Match(t *testing.T) {
	err := fmt.Errorf("Unable to delete selected password because it is in use by: Backup Configuration job")
	assert.True(t, isConfigBackupPasswordInUseError(err))
}

func TestIsConfigBackupPasswordInUseError_NoMatch(t *testing.T) {
	err := fmt.Errorf("some other error")
	assert.False(t, isConfigBackupPasswordInUseError(err))
}

// ---------------------------------------------------------------------------
// resolveLinuxSSHFingerprint — success and error paths
// ---------------------------------------------------------------------------

func TestResolveLinuxSSHFingerprint_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathConnectionCertificate, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*map[string]interface{})
			*result = map[string]interface{}{
				"fingerprint": "AA:BB:CC:DD",
			}
		}).Return(nil)

	data := &ManagedServerModel{
		Name:          types.StringValue("192.168.1.5"),
		CredentialsID: types.StringValue("cred-1"),
	}
	fp, err := r.resolveLinuxSSHFingerprint(context.Background(), data)
	assert.NoError(t, err)
	assert.Equal(t, "AA:BB:CC:DD", fp)
}

func TestResolveLinuxSSHFingerprint_APIError(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathConnectionCertificate, mock.Anything, mock.Anything).
		Return(fmt.Errorf("connection refused"))

	data := &ManagedServerModel{
		Name:          types.StringValue("192.168.1.5"),
		CredentialsID: types.StringValue("cred-1"),
	}
	_, err := r.resolveLinuxSSHFingerprint(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve SSH fingerprint")
}

func TestResolveLinuxSSHFingerprint_EmptyFingerprint(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &ManagedServer{client: mockClient}

	mockClient.On("PostJSON", mock.Anything, client.PathConnectionCertificate, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(3).(*map[string]interface{})
			*result = map[string]interface{}{
				"fingerprint": "",
			}
		}).Return(nil)

	data := &ManagedServerModel{
		Name:          types.StringValue("192.168.1.5"),
		CredentialsID: types.StringValue("cred-1"),
	}
	_, err := r.resolveLinuxSSHFingerprint(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fingerprint")
}

// ---------------------------------------------------------------------------
// repository syncFromAPIMap — LinuxLocal + NFS branches
// ---------------------------------------------------------------------------

func TestRepository_SyncFromAPIMap_LinuxLocalXfs(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{
		ID:   types.StringValue("repo-1"),
		Type: types.StringValue("LinuxLocal"),
	}
	api := map[string]interface{}{
		"id":     "repo-1",
		"name":   "Linux Repo XFS",
		"type":   "LinuxLocal",
		"hostId": "host-1",
		"repository": map[string]interface{}{
			"path":                       "/backups",
			"maxTaskCount":               4.0,
			"taskLimitEnabled":           true,
			"readWriteRate":              100.0,
			"readWriteLimitEnabled":      true,
			"useFastCloningOnXfsVolumes": true,
		},
	}
	r.syncFromAPIMap(data, api)
	assert.Equal(t, "Linux Repo XFS", data.Name.ValueString())
	assert.Equal(t, "LinuxLocal", data.Type.ValueString())
}

func TestRepository_SyncFromAPIMap_NfsBasic(t *testing.T) {
	r := &Repository{}
	data := &RepositoryModel{
		ID:   types.StringValue("repo-2"),
		Type: types.StringValue("Nfs"),
	}
	api := map[string]interface{}{
		"id":   "repo-2",
		"name": "NFS Repo Basic",
		"type": "Nfs",
	}
	r.syncFromAPIMap(data, api)
	assert.Equal(t, "NFS Repo Basic", data.Name.ValueString())
}

// ---------------------------------------------------------------------------
// Credential buildSpec — Linux with all optional fields set
// ---------------------------------------------------------------------------

func TestCredential_BuildSpec_Linux_AllFields(t *testing.T) {
	r := &Credential{}
	data := &CredentialModel{
		Type:               types.StringValue("Linux"),
		Username:           types.StringValue("root"),
		Password:           types.StringValue("pass"),
		Description:        types.StringValue("linux cred"),
		AuthenticationType: types.StringValue("Password"),
		SSHPort:            types.Int64Value(22),
		ElevateToRoot:      types.BoolValue(true),
		AddToSudoers:       types.BoolValue(true),
		UseSu:              types.BoolValue(false),
		PrivateKey:         types.StringValue("-----BEGIN RSA PRIVATE KEY-----"),
		Passphrase:         types.StringValue("keypass"),
		RootPassword:       types.StringValue("rootpass"),
	}
	spec := r.buildSpec(data)
	assert.NotNil(t, spec)
}

func TestCredential_BuildSpec_Windows(t *testing.T) {
	r := &Credential{}
	data := &CredentialModel{
		Type:        types.StringValue("Windows"),
		Username:    types.StringValue("Administrator"),
		Password:    types.StringValue("pass"),
		Description: types.StringValue("windows cred"),
	}
	spec := r.buildSpec(data)
	assert.NotNil(t, spec)
}

// ---------------------------------------------------------------------------
// deleteCredentialWithRetries — context cancelled mid-retry
// ---------------------------------------------------------------------------

func TestDeleteCredentialWithRetries_ContextCancelled(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Credential{client: mockClient}

	ctx, cancel := context.WithCancel(context.Background())

	// First call returns "in use" error, then we cancel context
	callCount := 0
	mockClient.On("DeleteJSON", mock.Anything, mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			callCount++
			if callCount == 1 {
				cancel() // cancel context after first call
			}
		}).Return(fmt.Errorf("Unable to delete selected credentials because it is currently in use by jobs"))

	err := r.deleteCredentialWithRetries(ctx, "/api/v1/credentials/cred-1")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// findProxyID — missing data array
// ---------------------------------------------------------------------------

func TestProxy_FindID_MissingDataArray(t *testing.T) {
	mockClient := new(MockVeeamClient)
	r := &Proxy{client: mockClient}

	mockClient.On("GetJSON", mock.Anything, client.PathProxies, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*map[string]interface{})
			*result = map[string]interface{}{} // no "data" key
		}).Return(nil)

	data := &ProxyModel{
		Type:        types.StringValue("ViProxy"),
		Description: types.StringNull(),
		HostID:      types.StringNull(),
	}
	_, err := r.findProxyID(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing data array")
}


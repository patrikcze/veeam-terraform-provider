package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

var (
	_ resource.Resource                = &CloudCredential{}
	_ resource.ResourceWithConfigure   = &CloudCredential{}
	_ resource.ResourceWithImportState = &CloudCredential{}
)

type CloudCredential struct {
	client client.APIClient
}

type CloudCredentialModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Type           types.String `tfsdk:"type"`
	AccountName    types.String `tfsdk:"account_name"`
	SecretKey      types.String `tfsdk:"secret_key"`
	TenantID       types.String `tfsdk:"tenant_id"`
	ApplicationID  types.String `tfsdk:"application_id"`
	ApplicationKey types.String `tfsdk:"application_key"`
	ProjectID      types.String `tfsdk:"project_id"`
	ServiceAccount types.String `tfsdk:"service_account"`
}

func (r *CloudCredential) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_credential"
}

func (r *CloudCredential) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam cloud credential (AWS, Azure, or GCP).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":            schema.StringAttribute{Required: true},
			"description":     schema.StringAttribute{Optional: true, Computed: true},
			"type":            schema.StringAttribute{Required: true},
			"account_name":    schema.StringAttribute{Optional: true},
			"secret_key":      schema.StringAttribute{Optional: true, Sensitive: true},
			"tenant_id":       schema.StringAttribute{Optional: true},
			"application_id":  schema.StringAttribute{Optional: true},
			"application_key": schema.StringAttribute{Optional: true, Sensitive: true},
			"project_id":      schema.StringAttribute{Optional: true},
			"service_account": schema.StringAttribute{Optional: true, Sensitive: true},
		},
	}
}

func (r *CloudCredential) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider, got unexpected type.")
		return
	}
	r.client = c
}

func (r *CloudCredential) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CloudCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.CloudCredentialModel
	if err := r.client.PostJSON(ctx, client.PathCloudCredentials, payload, &result); err != nil {
		resp.Diagnostics.AddError("Failed to create cloud credential", fmt.Sprintf("API error: %s", err))
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *CloudCredential) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CloudCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.CloudCredentialModel
	endpoint := fmt.Sprintf(client.PathCloudCredentialByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read cloud credential", fmt.Sprintf("API error for cloud credential %s: %s", data.ID.ValueString(), err))
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *CloudCredential) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CloudCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)
	endpoint := fmt.Sprintf(client.PathCloudCredentialByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError("Failed to update cloud credential", fmt.Sprintf("API error for cloud credential %s: %s", data.ID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *CloudCredential) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CloudCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathCloudCredentialByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError("Failed to delete cloud credential", fmt.Sprintf("API error for cloud credential %s: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *CloudCredential) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewCloudCredential() resource.Resource {
	return &CloudCredential{}
}

func (r *CloudCredential) buildSpec(data *CloudCredentialModel) *models.CloudCredentialSpec {
	spec := &models.CloudCredentialSpec{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueString(),
	}
	if !data.Description.IsNull() {
		spec.Description = data.Description.ValueString()
	}
	if !data.AccountName.IsNull() {
		spec.AccountName = data.AccountName.ValueString()
	}
	if !data.SecretKey.IsNull() {
		spec.SecretKey = data.SecretKey.ValueString()
	}
	if !data.TenantID.IsNull() {
		spec.TenantID = data.TenantID.ValueString()
	}
	if !data.ApplicationID.IsNull() {
		spec.ApplicationID = data.ApplicationID.ValueString()
	}
	if !data.ApplicationKey.IsNull() {
		spec.ApplicationKey = data.ApplicationKey.ValueString()
	}
	if !data.ProjectID.IsNull() {
		spec.ProjectID = data.ProjectID.ValueString()
	}
	if !data.ServiceAccount.IsNull() {
		spec.ServiceAccount = data.ServiceAccount.ValueString()
	}
	return spec
}

func (r *CloudCredential) syncFromAPI(data *CloudCredentialModel, api *models.CloudCredentialModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(api.Type)
	if api.AccountName != "" {
		data.AccountName = types.StringValue(api.AccountName)
	}
	if api.TenantID != "" {
		data.TenantID = types.StringValue(api.TenantID)
	}
	if api.ProjectID != "" {
		data.ProjectID = types.StringValue(api.ProjectID)
	}
}

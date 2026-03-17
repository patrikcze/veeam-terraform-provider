package resources

import (
	"context"
	"fmt"
	"strings"

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
	AccessKey      types.String `tfsdk:"access_key"`
	Account        types.String `tfsdk:"account"`
	SharedKey      types.String `tfsdk:"shared_key"`
	ConnectionName types.String `tfsdk:"connection_name"`
	CreationMode   types.String `tfsdk:"creation_mode"`
	DeploymentType types.String `tfsdk:"deployment_type"`
	DeploymentRegion types.String `tfsdk:"deployment_region"`
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
			"access_key":      schema.StringAttribute{Optional: true},
			"account":         schema.StringAttribute{Optional: true},
			"shared_key":      schema.StringAttribute{Optional: true, Sensitive: true},
			"connection_name": schema.StringAttribute{Optional: true},
			"creation_mode":   schema.StringAttribute{Optional: true},
			"deployment_type": schema.StringAttribute{Optional: true},
			"deployment_region": schema.StringAttribute{Optional: true},
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

	payload, validationError := r.buildSpec(&data)
	if validationError != "" {
		resp.Diagnostics.AddError("Invalid cloud credential configuration", validationError)
		return
	}

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

	payload, validationError := r.buildSpec(&data)
	if validationError != "" {
		resp.Diagnostics.AddError("Invalid cloud credential configuration", validationError)
		return
	}
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

func (r *CloudCredential) buildSpec(data *CloudCredentialModel) (*models.CloudCredentialSpec, string) {
	credentialType := strings.TrimSpace(data.Type.ValueString())

	spec := &models.CloudCredentialSpec{
		Name: data.Name.ValueString(),
		Type: credentialType,
	}
	if !data.Description.IsNull() {
		spec.Description = data.Description.ValueString()
	}

	switch credentialType {
	case "Amazon":
		accessKey := firstNonEmptyString(
			valueIfKnown(data.AccessKey),
			valueIfKnown(data.AccountName),
		)
		secretKey := valueIfKnown(data.SecretKey)
		if accessKey == "" || secretKey == "" {
			return nil, "Type 'Amazon' requires 'access_key' (or 'account_name') and 'secret_key'."
		}
		spec.AccessKey = accessKey
		spec.SecretKey = secretKey

	case "AzureStorage":
		account := firstNonEmptyString(
			valueIfKnown(data.Account),
			valueIfKnown(data.AccountName),
		)
		sharedKey := firstNonEmptyString(
			valueIfKnown(data.SharedKey),
			valueIfKnown(data.SecretKey),
		)
		if account == "" || sharedKey == "" {
			return nil, "Type 'AzureStorage' requires 'account' (or 'account_name') and 'shared_key' (or 'secret_key')."
		}
		spec.Account = account
		spec.SharedKey = sharedKey

	case "AzureCompute":
		tenantID := valueIfKnown(data.TenantID)
		applicationID := valueIfKnown(data.ApplicationID)
		secret := valueIfKnown(data.ApplicationKey)
		if tenantID == "" || applicationID == "" || secret == "" {
			return nil, "Type 'AzureCompute' requires 'tenant_id', 'application_id', and 'application_key' (secret) for ExistingAccount mode."
		}

		connectionName := firstNonEmptyString(valueIfKnown(data.ConnectionName), valueIfKnown(data.Name))
		creationMode := firstNonEmptyString(valueIfKnown(data.CreationMode), "ExistingAccount")
		if creationMode != "ExistingAccount" {
			return nil, "Currently only 'ExistingAccount' is supported for type 'AzureCompute'. Set 'creation_mode' to 'ExistingAccount'."
		}

		deploymentType := firstNonEmptyString(valueIfKnown(data.DeploymentType), "MicrosoftAzure")
		if deploymentType != "MicrosoftAzure" && deploymentType != "MicrosoftAzureStack" {
			return nil, "'deployment_type' must be either 'MicrosoftAzure' or 'MicrosoftAzureStack'."
		}

		spec.ConnectionName = connectionName
		spec.CreationMode = creationMode
		spec.ExistingAccount = &models.AzureComputeCredentialsExistingAccountSpec{
			Deployment: models.AzureComputeCloudCredentialsDeploymentModel{
				DeploymentType: deploymentType,
				Region:         valueIfKnown(data.DeploymentRegion),
			},
			Subscription: models.AzureComputeCloudCredentialsSubscriptionSpec{
				TenantID:      tenantID,
				ApplicationID: applicationID,
				Secret:        secret,
			},
		}

	case "Google":
		account := firstNonEmptyString(valueIfKnown(data.AccountName), valueIfKnown(data.Account))
		secretKey := valueIfKnown(data.SecretKey)
		if account != "" {
			spec.AccountName = account
		}
		if secretKey != "" {
			spec.SecretKey = secretKey
		}

	case "GoogleService":
		serviceAccount := valueIfKnown(data.ServiceAccount)
		projectID := valueIfKnown(data.ProjectID)
		if serviceAccount == "" {
			return nil, "Type 'GoogleService' requires 'service_account'."
		}
		spec.ServiceAccount = serviceAccount
		if projectID != "" {
			spec.ProjectID = projectID
		}

	default:
		return nil, "Unsupported cloud credential type. Use one of: Amazon, AzureStorage, AzureCompute, Google, GoogleService."
	}

	return spec, ""
}

func (r *CloudCredential) syncFromAPI(data *CloudCredentialModel, api *models.CloudCredentialModel) {
	data.Name = types.StringValue(api.Name)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(api.Type)
	if api.AccessKey != "" {
		data.AccessKey = types.StringValue(api.AccessKey)
	}
	if api.Account != "" {
		data.Account = types.StringValue(api.Account)
	}
	if api.ConnectionName != "" {
		data.ConnectionName = types.StringValue(api.ConnectionName)
	}
	if api.AccountName != "" {
		data.AccountName = types.StringValue(api.AccountName)
	}
	if api.TenantID != "" {
		data.TenantID = types.StringValue(api.TenantID)
	}
	if api.ApplicationID != "" {
		data.ApplicationID = types.StringValue(api.ApplicationID)
	}
	if api.ProjectID != "" {
		data.ProjectID = types.StringValue(api.ProjectID)
	}
}

func valueIfKnown(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

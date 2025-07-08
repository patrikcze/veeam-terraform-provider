package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Credential{}
	_ resource.ResourceWithConfigure   = &Credential{}
	_ resource.ResourceWithImportState = &Credential{}
)

// Credential defines the resource implementation.
type Credential struct {
	client *client.VeeamClient
}

// CredentialModel describes the Terraform resource data model.
type CredentialModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Type        types.String `tfsdk:"type"`
	Domain      types.String `tfsdk:"domain"`
	// Add additional fields as required
}

// Metadata returns the resource type name.
func (r *Credential) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

// Schema defines the schema for the resource.
func (r *Credential) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Veeam Credential resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Credential identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Credential.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Credential.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the Credential.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the Credential.",
				Required:            true,
				Sensitive:           true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the Credential (e.g., 'windows', 'linux', 'standard').",
				Required:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "Domain for the Credential (if applicable).",
				Optional:            true,
			},
			// Additional attributes as needed
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *Credential) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.VeeamClient)
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *Credential) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"username":    data.Username.ValueString(),
		"password":    data.Password.ValueString(),
		"type":        data.Type.ValueString(),
	}

	if !data.Domain.IsNull() {
		payload["domain"] = data.Domain.ValueString()
	}

	var result map[string]interface{}
	err := r.client.PostJSON("/credentials", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating credential",
			fmt.Sprintf("Could not create credential: %s", err),
		)
		return
	}

	// Set the ID from the API response
	if id, ok := result["id"].(string); ok {
		data.ID = types.StringValue(id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Credential) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.GetJSON(fmt.Sprintf("/credentials/%s", data.ID.ValueString()), &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading credential",
			fmt.Sprintf("Could not read credential: %s", err),
		)
		return
	}

	// Update the data model with the API response
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if username, ok := result["username"].(string); ok {
		data.Username = types.StringValue(username)
	}
	// Note: Password is typically not returned from API for security reasons
	// Keep the current password value in state
	if credType, ok := result["type"].(string); ok {
		data.Type = types.StringValue(credType)
	}
	if domain, ok := result["domain"].(string); ok {
		data.Domain = types.StringValue(domain)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Credential) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"username":    data.Username.ValueString(),
		"password":    data.Password.ValueString(),
		"type":        data.Type.ValueString(),
	}

	if !data.Domain.IsNull() {
		payload["domain"] = data.Domain.ValueString()
	}

	err := r.client.PutJSON(fmt.Sprintf("/credentials/%s", data.ID.ValueString()), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating credential",
			fmt.Sprintf("Could not update credential: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Credential) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJSON(fmt.Sprintf("/credentials/%s", data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting credential",
			fmt.Sprintf("Could not delete credential: %s", err),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *Credential) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Set the ID from the import request
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

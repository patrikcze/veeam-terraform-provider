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
	_ resource.Resource                = &Repository{}
	_ resource.ResourceWithConfigure   = &Repository{}
	_ resource.ResourceWithImportState = &Repository{}
)

// Repository defines the resource implementation.
type Repository struct {
	client *client.VeeamClient
}

// RepositoryModel describes the Terraform resource data model.
type RepositoryModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Path        types.String `tfsdk:"path"`
	Type        types.String `tfsdk:"type"`
	Capacity    types.Int64  `tfsdk:"capacity"`
	// Add additional fields as required
}

// Metadata returns the resource type name.
func (r *Repository) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

// Schema defines the schema for the resource.
func (r *Repository) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Veeam Repository resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Repository identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Repository.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Repository.",
				Optional:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Path of the Repository.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the Repository.",
				Required:            true,
			},
			"capacity": schema.Int64Attribute{
				MarkdownDescription: "Capacity of the Repository in bytes.",
				Optional:            true,
			},
			// Additional attributes as needed
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *Repository) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.VeeamClient)
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *Repository) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RepositoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"path":        data.Path.ValueString(),
		"type":        data.Type.ValueString(),
	}

	if !data.Capacity.IsNull() {
		payload["capacity"] = data.Capacity.ValueInt64()
	}

	var result map[string]interface{}
	err := r.client.PostJSON("/repositories", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			fmt.Sprintf("Could not create repository: %s", err),
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
func (r *Repository) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RepositoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.GetJSON(fmt.Sprintf("/repositories/%s", data.ID.ValueString()), &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository",
			fmt.Sprintf("Could not read repository: %s", err),
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
	if path, ok := result["path"].(string); ok {
		data.Path = types.StringValue(path)
	}
	if repoType, ok := result["type"].(string); ok {
		data.Type = types.StringValue(repoType)
	}
	if capacity, ok := result["capacity"].(float64); ok {
		data.Capacity = types.Int64Value(int64(capacity))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Repository) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RepositoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"path":        data.Path.ValueString(),
		"type":        data.Type.ValueString(),
	}

	if !data.Capacity.IsNull() {
		payload["capacity"] = data.Capacity.ValueInt64()
	}

	err := r.client.PutJSON(fmt.Sprintf("/repositories/%s", data.ID.ValueString()), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating repository",
			fmt.Sprintf("Could not update repository: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Repository) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RepositoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJSON(fmt.Sprintf("/repositories/%s", data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting repository",
			fmt.Sprintf("Could not delete repository: %s", err),
		)
		return
	}
}

// NewRepository is a helper function to simplify the provider implementation.
func NewRepository() resource.Resource {
	return &Repository{}
}

// ImportState imports the resource state.
func (r *Repository) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Set the ID from the import request
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

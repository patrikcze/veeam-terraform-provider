package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &RepositoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &RepositoriesDataSource{}
)

// RepositoriesDataSource defines the data source implementation.
type RepositoriesDataSource struct {
	client *client.VeeamClient
}

// RepositoriesDataSourceModel describes the data source data model.
type RepositoriesDataSourceModel struct {
	ID             types.String          `tfsdk:"id"`
	RepositoryID   types.String          `tfsdk:"repository_id"`
	RepositoryName types.String          `tfsdk:"repository_name"`
	Repositories   []RepositoryDataModel `tfsdk:"repositories"`
}

// RepositoryDataModel describes the repository data model.
type RepositoryDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Path        types.String `tfsdk:"path"`
	Type        types.String `tfsdk:"type"`
	Capacity    types.Int64  `tfsdk:"capacity"`
	FreeSpace   types.Int64  `tfsdk:"free_space"`
	UsedSpace   types.Int64  `tfsdk:"used_space"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewRepositoriesDataSource is a helper function to simplify the provider implementation.
func NewRepositoriesDataSource() datasource.DataSource {
	return &RepositoriesDataSource{}
}

// Metadata returns the data source type name.
func (d *RepositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

// Schema defines the schema for the data source.
func (d *RepositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for fetching repositories from Veeam. Can fetch all repositories or a specific repository by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"repository_id": schema.StringAttribute{
				MarkdownDescription: "ID of a specific repository to fetch",
				Optional:            true,
			},
			"repository_name": schema.StringAttribute{
				MarkdownDescription: "Name of a specific repository to fetch",
				Optional:            true,
			},
			"repositories": schema.ListNestedAttribute{
				MarkdownDescription: "List of repositories",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Repository identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the repository",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the repository",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Path of the repository",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the repository",
							Computed:            true,
						},
						"capacity": schema.Int64Attribute{
							MarkdownDescription: "Total capacity of the repository in bytes",
							Computed:            true,
						},
						"free_space": schema.Int64Attribute{
							MarkdownDescription: "Free space in the repository in bytes",
							Computed:            true,
						},
						"used_space": schema.Int64Attribute{
							MarkdownDescription: "Used space in the repository in bytes",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Status of the repository",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Creation timestamp",
							Computed:            true,
						},
						"updated_at": schema.StringAttribute{
							MarkdownDescription: "Last update timestamp",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *RepositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.VeeamClient)
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *RepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepositoriesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if we're looking for a specific repository
	if !data.RepositoryID.IsNull() {
		// Fetch single repository by ID
		var apiResult map[string]interface{}
		err := d.client.GetJSON(ctx, fmt.Sprintf("/api/v1/repositories/%s", data.RepositoryID.ValueString()), &apiResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching repository",
				fmt.Sprintf("Could not fetch repository with ID %s: %s", data.RepositoryID.ValueString(), err),
			)
			return
		}

		// Convert single repository to array
		repositories := []RepositoryDataModel{
			{
				ID:          types.StringValue(getStringValue(apiResult, "id")),
				Name:        types.StringValue(getStringValue(apiResult, "name")),
				Description: types.StringValue(getStringValue(apiResult, "description")),
				Path:        types.StringValue(getStringValue(apiResult, "path")),
				Type:        types.StringValue(getStringValue(apiResult, "type")),
				Capacity:    types.Int64Value(getInt64Value(apiResult, "capacity")),
				FreeSpace:   types.Int64Value(getInt64Value(apiResult, "freeSpace")),
				UsedSpace:   types.Int64Value(getInt64Value(apiResult, "usedSpace")),
				Status:      types.StringValue(getStringValue(apiResult, "status")),
				CreatedAt:   types.StringValue(getStringValue(apiResult, "createdAt")),
				UpdatedAt:   types.StringValue(getStringValue(apiResult, "updatedAt")),
			},
		}

		data.ID = types.StringValue(fmt.Sprintf("repository_%s", data.RepositoryID.ValueString()))
		data.Repositories = repositories
	} else {
		// Fetch all repositories
		var apiResult []map[string]interface{}
		err := d.client.GetJSON(ctx, "/api/v1/repositories", &apiResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching repositories",
				fmt.Sprintf("Could not fetch repositories: %s", err),
			)
			return
		}

		// Filter by name if specified
		if !data.RepositoryName.IsNull() {
			filtered := make([]map[string]interface{}, 0)
			for _, repo := range apiResult {
				if getStringValue(repo, "name") == data.RepositoryName.ValueString() {
					filtered = append(filtered, repo)
				}
			}
			apiResult = filtered
		}

		// Map API response to the data model
		repositories := make([]RepositoryDataModel, len(apiResult))
		for i, repo := range apiResult {
			repositories[i] = RepositoryDataModel{
				ID:          types.StringValue(getStringValue(repo, "id")),
				Name:        types.StringValue(getStringValue(repo, "name")),
				Description: types.StringValue(getStringValue(repo, "description")),
				Path:        types.StringValue(getStringValue(repo, "path")),
				Type:        types.StringValue(getStringValue(repo, "type")),
				Capacity:    types.Int64Value(getInt64Value(repo, "capacity")),
				FreeSpace:   types.Int64Value(getInt64Value(repo, "freeSpace")),
				UsedSpace:   types.Int64Value(getInt64Value(repo, "usedSpace")),
				Status:      types.StringValue(getStringValue(repo, "status")),
				CreatedAt:   types.StringValue(getStringValue(repo, "createdAt")),
				UpdatedAt:   types.StringValue(getStringValue(repo, "updatedAt")),
			}
		}

		// Set the data source identifier
		if !data.RepositoryName.IsNull() {
			data.ID = types.StringValue(fmt.Sprintf("repository_name_%s", data.RepositoryName.ValueString()))
		} else {
			data.ID = types.StringValue("repositories")
		}
		data.Repositories = repositories
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

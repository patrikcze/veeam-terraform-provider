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
	_ datasource.DataSource              = &BackupJobsDataSource{}
	_ datasource.DataSourceWithConfigure = &BackupJobsDataSource{}
)

// BackupJobsDataSource defines the data source implementation.
type BackupJobsDataSource struct {
	client *client.VeeamClient
}

// BackupJobsDataSourceModel describes the data source data model.
type BackupJobsDataSourceModel struct {
	ID         types.String         `tfsdk:"id"`
	JobID      types.String         `tfsdk:"job_id"`
	JobName    types.String         `tfsdk:"job_name"`
	BackupJobs []BackupJobDataModel `tfsdk:"backup_jobs"`
}

// BackupJobDataModel describes the backup job data model.
type BackupJobDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Repository  types.String `tfsdk:"repository"`
	Schedule    types.String `tfsdk:"schedule"`
	JobType     types.String `tfsdk:"job_type"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewBackupJobsDataSource is a helper function to simplify the provider implementation.
func NewBackupJobsDataSource() datasource.DataSource {
	return &BackupJobsDataSource{}
}

// Metadata returns the data source type name.
func (d *BackupJobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_jobs"
}

// Schema defines the schema for the data source.
func (d *BackupJobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for fetching backup jobs from Veeam. Can fetch all jobs or a specific job by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "ID of a specific backup job to fetch",
				Optional:            true,
			},
			"job_name": schema.StringAttribute{
				MarkdownDescription: "Name of a specific backup job to fetch",
				Optional:            true,
			},
			"backup_jobs": schema.ListNestedAttribute{
				MarkdownDescription: "List of backup jobs",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Backup job identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the backup job",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the backup job is enabled",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the backup job",
							Computed:            true,
						},
						"repository": schema.StringAttribute{
							MarkdownDescription: "Repository used by the backup job",
							Computed:            true,
						},
						"schedule": schema.StringAttribute{
							MarkdownDescription: "Schedule of the backup job",
							Computed:            true,
						},
						"job_type": schema.StringAttribute{
							MarkdownDescription: "Type of the backup job",
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
func (d *BackupJobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.VeeamClient)
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *BackupJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackupJobsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if we're looking for a specific job
	if !data.JobID.IsNull() {
		// Fetch single job by ID
		var apiResult map[string]interface{}
		err := d.client.GetJSON(fmt.Sprintf("/api/v1/backupJobs/%s", data.JobID.ValueString()), &apiResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching backup job",
				fmt.Sprintf("Could not fetch backup job with ID %s: %s", data.JobID.ValueString(), err),
			)
			return
		}

		// Convert single job to array
		backupJobs := []BackupJobDataModel{
			{
				ID:          types.StringValue(getStringValue(apiResult, "id")),
				Name:        types.StringValue(getStringValue(apiResult, "name")),
				Enabled:     types.BoolValue(getBoolValue(apiResult, "enabled")),
				Description: types.StringValue(getStringValue(apiResult, "description")),
				Repository:  types.StringValue(getStringValue(apiResult, "repository")),
				Schedule:    types.StringValue(getStringValue(apiResult, "schedule")),
				JobType:     types.StringValue(getStringValue(apiResult, "jobType")),
				CreatedAt:   types.StringValue(getStringValue(apiResult, "createdAt")),
				UpdatedAt:   types.StringValue(getStringValue(apiResult, "updatedAt")),
			},
		}

		data.ID = types.StringValue(fmt.Sprintf("backup_job_%s", data.JobID.ValueString()))
		data.BackupJobs = backupJobs
	} else {
		// Fetch all backup jobs
		var apiResult []map[string]interface{}
		err := d.client.GetJSON("/api/v1/backupJobs", &apiResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching backup jobs",
				fmt.Sprintf("Could not fetch backup jobs: %s", err),
			)
			return
		}

		// Filter by name if specified
		if !data.JobName.IsNull() {
			filtered := make([]map[string]interface{}, 0)
			for _, job := range apiResult {
				if getStringValue(job, "name") == data.JobName.ValueString() {
					filtered = append(filtered, job)
				}
			}
			apiResult = filtered
		}

		// Map API response to the data model
		backupJobs := make([]BackupJobDataModel, len(apiResult))
		for i, job := range apiResult {
			backupJobs[i] = BackupJobDataModel{
				ID:          types.StringValue(getStringValue(job, "id")),
				Name:        types.StringValue(getStringValue(job, "name")),
				Enabled:     types.BoolValue(getBoolValue(job, "enabled")),
				Description: types.StringValue(getStringValue(job, "description")),
				Repository:  types.StringValue(getStringValue(job, "repository")),
				Schedule:    types.StringValue(getStringValue(job, "schedule")),
				JobType:     types.StringValue(getStringValue(job, "jobType")),
				CreatedAt:   types.StringValue(getStringValue(job, "createdAt")),
				UpdatedAt:   types.StringValue(getStringValue(job, "updatedAt")),
			}
		}

		// Set the data source identifier
		if !data.JobName.IsNull() {
			data.ID = types.StringValue(fmt.Sprintf("backup_job_name_%s", data.JobName.ValueString()))
		} else {
			data.ID = types.StringValue("backup_jobs")
		}
		data.BackupJobs = backupJobs
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

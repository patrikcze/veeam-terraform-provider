package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &BackupJob{}
	_ resource.ResourceWithConfigure = &BackupJob{}
)

// BackupJob defines the resource implementation.
type BackupJob struct {
	client *client.VeeamClient
}

// BackupJobModel describes the Terraform resource data model.
type BackupJobModel struct {
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
	// Add additional fields as required
}

// Metadata returns the resource type name.
func (r *BackupJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_job"
}

// Schema defines the schema for the resource.
func (r *BackupJob) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Backup Job.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Backup Job is enabled.",
				Optional:            true,
			},
			// Additional attributes as needed
		},
	}
}

// Configure assigns the provider-configured client to the resource.
func (r *BackupJob) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.VeeamClient)
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *BackupJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupJobModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":    data.Name.ValueString(),
		"enabled": data.Enabled.ValueBool(),
	}

	var result map[string]interface{}
	err := r.client.PostJSON(ctx, "/backupJobs", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating backup job",
			fmt.Sprintf("Could not create backup job: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *BackupJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupJobModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.GetJSON(ctx, fmt.Sprintf("/backupJobs/%s", data.Name.ValueString()), &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backup job",
			fmt.Sprintf("Could not read backup job: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *BackupJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupJobModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":    data.Name.ValueString(),
		"enabled": data.Enabled.ValueBool(),
	}

	err := r.client.PutJSON(ctx, fmt.Sprintf("/backupJobs/%s", data.Name.ValueString()), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating backup job",
			fmt.Sprintf("Could not update backup job: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *BackupJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupJobModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteJSON(ctx, fmt.Sprintf("/backupJobs/%s", data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backup job",
			fmt.Sprintf("Could not delete backup job: %s", err.Error()),
		)
		return
	}
}

// NewBackupJob is a helper function to simplify the provider implementation.
func NewBackupJob() resource.Resource {
	return &BackupJob{}
}

// Launch the CRUD operations and necessary configurations.
func (r *BackupJob) ImportState(_ context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Example import functionality would be added here, if required
}

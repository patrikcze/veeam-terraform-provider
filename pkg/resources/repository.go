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

// Compile-time interface checks.
var (
	_ resource.Resource                = &Repository{}
	_ resource.ResourceWithConfigure   = &Repository{}
	_ resource.ResourceWithImportState = &Repository{}
)

// Repository implements the veeam_repository resource.
type Repository struct {
	client client.APIClient
}

// RepositoryModel is the Terraform state model for veeam_repository.
// Supports WinLocal, LinuxLocal, Nfs, and Smb types.
type RepositoryModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	HostID        types.String `tfsdk:"host_id"`
	Path          types.String `tfsdk:"path"`
	MaxTaskCount  types.Int64  `tfsdk:"max_task_count"`
	SharePath     types.String `tfsdk:"share_path"`
	CredentialsID types.String `tfsdk:"credentials_id"`
}

func (r *Repository) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *Repository) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam backup repository (WinLocal, LinuxLocal, Nfs, Smb).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Repository identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Repository name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Repository type: `WinLocal`, `LinuxLocal`, `Nfs`, or `Smb`.",
				Required:            true,
			},
			"host_id": schema.StringAttribute{
				MarkdownDescription: "Managed server host ID (WinLocal and LinuxLocal types).",
				Optional:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Folder path on the host (WinLocal / LinuxLocal).",
				Optional:            true,
			},
			"max_task_count": schema.Int64Attribute{
				MarkdownDescription: "Maximum concurrent tasks.",
				Optional:            true,
			},
			"share_path": schema.StringAttribute{
				MarkdownDescription: "Network share path (Nfs or Smb types).",
				Optional:            true,
			},
			"credentials_id": schema.StringAttribute{
				MarkdownDescription: "Credential ID for SMB share access.",
				Optional:            true,
			},
		},
	}
}

func (r *Repository) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected client.APIClient from provider, got unexpected type.",
		)
		return
	}
	r.client = c
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *Repository) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result map[string]interface{}
	if err := r.client.PostJSON(ctx, client.PathRepositories, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create repository",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	resultID := getStringValue(result, "id")
	resultType := getStringValue(result, "type")

	// Some Veeam operations return async session objects. In that case,
	// wait for completion and then resolve repository ID by name.
	if resultType == "" {
		if resultID == "" {
			resp.Diagnostics.AddError(
				"Failed to create repository",
				"API response did not include repository type or async session ID.",
			)
			return
		}

		if err := r.client.WaitForTask(ctx, resultID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to create repository",
				fmt.Sprintf("Async repository creation task %s failed: %s", resultID, err),
			)
			return
		}

		resolvedID, err := r.findRepositoryIDByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve created repository",
				err.Error(),
			)
			return
		}
		data.ID = types.StringValue(resolvedID)
	} else {
		data.ID = types.StringValue(resultID)
	}

	// Read the created repository to sync computed fields while preserving plan values.
	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		var created models.RepositoryModel
		endpoint := fmt.Sprintf(client.PathRepositoryByID, data.ID.ValueString())
		if err := r.client.GetJSON(ctx, endpoint, &created); err == nil {
			r.syncFromAPI(&data, &created)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Repository) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.RepositoryModel
	endpoint := fmt.Sprintf(client.PathRepositoryByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read repository",
			fmt.Sprintf("API error for repository %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Repository) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathRepositoryByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update repository",
			fmt.Sprintf("API error for repository %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Repository) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathRepositoryByID, data.ID.ValueString())
	if err := r.client.DeleteJSON(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete repository",
			fmt.Sprintf("API error for repository %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *Repository) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewRepository returns a new veeam_repository resource instance.
func NewRepository() resource.Resource {
	return &Repository{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *Repository) buildSpec(data *RepositoryModel) interface{} {
	repoType := models.ERepositoryType(data.Type.ValueString())

	base := models.RepositorySpec{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        repoType,
	}

	maxTasks := 0
	if !data.MaxTaskCount.IsNull() && !data.MaxTaskCount.IsUnknown() {
		maxTasks = int(data.MaxTaskCount.ValueInt64())
	}

	switch repoType {
	case models.RepositoryTypeWinLocal:
		mountServer := &models.MountServersSettings{
			MountServerSettingsType: "Windows",
			Windows: &models.MountServerSettings{
				MountServerID:    data.HostID.ValueString(),
				WriteCacheFolder: data.Path.ValueString(),
				VPowerNFSEnabled: false,
			},
		}
		return &models.WindowsLocalStorageSpec{
			RepositorySpec: base,
			HostID:         data.HostID.ValueString(),
			Repository: &models.WindowsLocalRepositorySettings{
				Path:         data.Path.ValueString(),
				MaxTaskCount: maxTasks,
			},
			MountServer: mountServer,
		}

	case models.RepositoryTypeLinuxLocal:
		mountServer := &models.MountServersSettings{
			MountServerSettingsType: "Linux",
			Linux: &models.MountServerSettings{
				MountServerID:    data.HostID.ValueString(),
				WriteCacheFolder: data.Path.ValueString(),
				VPowerNFSEnabled: false,
			},
		}
		return &models.LinuxLocalStorageSpec{
			RepositorySpec: base,
			HostID:         data.HostID.ValueString(),
			Repository: &models.LinuxLocalRepositorySettings{
				Path:         data.Path.ValueString(),
				MaxTaskCount: maxTasks,
			},
			MountServer: mountServer,
		}

	case models.RepositoryTypeNfs:
		return &models.NfsStorageSpec{
			RepositorySpec: base,
			Share: &models.NfsShareSettings{
				SharePath: data.SharePath.ValueString(),
			},
			Repository: &models.NetworkRepositorySettings{
				MaxTaskCount: maxTasks,
			},
		}

	case models.RepositoryTypeSmb:
		spec := &models.SmbStorageSpec{
			RepositorySpec: base,
			Share: &models.SmbShareSettings{
				SharePath: data.SharePath.ValueString(),
			},
			Repository: &models.NetworkRepositorySettings{
				MaxTaskCount: maxTasks,
			},
		}
		if !data.CredentialsID.IsNull() {
			spec.Share.CredentialsID = data.CredentialsID.ValueString()
		}
		return spec

	default:
		// Fallback for unknown types — send base spec
		return &base
	}
}

func (r *Repository) syncFromAPI(data *RepositoryModel, api *models.RepositoryModel) {
	if api.Name != "" {
		data.Name = types.StringValue(api.Name)
	}
	if api.Description != "" {
		data.Description = types.StringValue(api.Description)
	}
	if string(api.Type) != "" {
		data.Type = types.StringValue(string(api.Type))
	}
}

func (r *Repository) findRepositoryIDByName(ctx context.Context, name string) (string, error) {
	var payload map[string]interface{}
	if err := r.client.GetJSON(ctx, client.PathRepositories, &payload); err != nil {
		return "", fmt.Errorf("failed to list repositories after create: %w", err)
	}

	rawData, ok := payload["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected repositories list response shape: missing data array")
	}

	for _, item := range rawData {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if getStringValue(entry, "name") == name {
			id := getStringValue(entry, "id")
			if id != "" {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("repository %q was created but could not be located in repository list", name)
}

func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

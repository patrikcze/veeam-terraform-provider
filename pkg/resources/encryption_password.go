package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	_ resource.Resource                = &EncryptionPassword{}
	_ resource.ResourceWithConfigure   = &EncryptionPassword{}
	_ resource.ResourceWithImportState = &EncryptionPassword{}
)

// EncryptionPassword implements the veeam_encryption_password resource.
type EncryptionPassword struct {
	client client.APIClient
}

// EncryptionPasswordModel is the Terraform state model.
type EncryptionPasswordModel struct {
	ID       types.String `tfsdk:"id"`
	Password types.String `tfsdk:"password"`
	Hint     types.String `tfsdk:"hint"`
}

func (r *EncryptionPassword) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encryption_password"
}

func (r *EncryptionPassword) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam encryption password used for backup encryption.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Encryption password identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The encryption password.",
				Required:            true,
				Sensitive:           true,
			},
			"hint": schema.StringAttribute{
				MarkdownDescription: "Hint to help remember the password.",
				Required:            true,
			},
		},
	}
}

func (r *EncryptionPassword) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EncryptionPassword) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EncryptionPasswordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := models.EncryptionPasswordSpec{
		Password: data.Password.ValueString(),
		Hint:     data.Hint.ValueString(),
	}

	var result models.EncryptionPasswordModel
	if err := r.client.PostJSON(ctx, client.PathEncryptionPasswords, &payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create encryption password",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	data.Hint = types.StringValue(result.Hint)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EncryptionPassword) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EncryptionPasswordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.EncryptionPasswordModel
	endpoint := fmt.Sprintf(client.PathEncryptionPasswordByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read encryption password",
			fmt.Sprintf("API error for encryption password %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	data.Hint = types.StringValue(result.Hint)
	// Password is never returned by the API — keep current state value.
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EncryptionPassword) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EncryptionPasswordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := models.EncryptionPasswordSpec{
		Password: data.Password.ValueString(),
		Hint:     data.Hint.ValueString(),
	}

	endpoint := fmt.Sprintf(client.PathEncryptionPasswordByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, &payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update encryption password",
			fmt.Sprintf("API error for encryption password %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *EncryptionPassword) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EncryptionPasswordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathEncryptionPasswordByID, data.ID.ValueString())
	if err := r.deleteEncryptionPasswordWithRetries(ctx, endpoint); err != nil {
		if isConfigBackupPasswordInUseError(err) {
			resp.Diagnostics.AddError(
				"Failed to delete encryption password",
				fmt.Sprintf("API error for encryption password %s: %s. Veeam still reports the password in use by Configuration Backup. Ensure configuration backup encryption is switched to another password (or fully disabled in VBR), then retry destroy.", data.ID.ValueString(), err),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Failed to delete encryption password",
			fmt.Sprintf("API error for encryption password %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *EncryptionPassword) deleteEncryptionPasswordWithRetries(ctx context.Context, endpoint string) error {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		lastErr = r.client.DeleteJSON(ctx, endpoint)
		if lastErr == nil {
			return nil
		}

		if !isConfigBackupPasswordInUseError(lastErr) {
			return lastErr
		}

		if attempt < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	return lastErr
}

func isConfigBackupPasswordInUseError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unable to delete selected password because it is in use by: backup configuration job")
}

func (r *EncryptionPassword) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewEncryptionPassword returns a new veeam_encryption_password resource instance.
func NewEncryptionPassword() resource.Resource {
	return &EncryptionPassword{}
}

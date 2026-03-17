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
	_ resource.Resource                = &Credential{}
	_ resource.ResourceWithConfigure   = &Credential{}
	_ resource.ResourceWithImportState = &Credential{}
)

// Credential implements the veeam_credential resource.
type Credential struct {
	client client.APIClient
}

// CredentialModel is the Terraform state model for veeam_credential.
type CredentialModel struct {
	ID                 types.String `tfsdk:"id"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Description        types.String `tfsdk:"description"`
	Type               types.String `tfsdk:"type"`
	SSHPort            types.Int64  `tfsdk:"ssh_port"`
	ElevateToRoot      types.Bool   `tfsdk:"elevate_to_root"`
	AddToSudoers       types.Bool   `tfsdk:"add_to_sudoers"`
	UseSu              types.Bool   `tfsdk:"use_su"`
	AuthenticationType types.String `tfsdk:"authentication_type"`
	PrivateKey         types.String `tfsdk:"private_key"`
	Passphrase         types.String `tfsdk:"passphrase"`
	RootPassword       types.String `tfsdk:"root_password"`
}

func (r *Credential) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (r *Credential) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Veeam credential (Standard or Linux).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Credential identifier (assigned by the server).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username. For domain accounts use `DOMAIN\\user` format.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the credential.",
				Required:            true,
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Credential type: `Standard` or `Linux`.",
				Required:            true,
			},
			// --- Linux-specific (optional) ---
			"ssh_port": schema.Int64Attribute{
				MarkdownDescription: "SSH port (Linux credentials only, default 22).",
				Optional:            true,
			},
			"elevate_to_root": schema.BoolAttribute{
				MarkdownDescription: "Elevate to root via sudo (Linux only).",
				Optional:            true,
			},
			"add_to_sudoers": schema.BoolAttribute{
				MarkdownDescription: "Automatically add to sudoers (Linux only).",
				Optional:            true,
			},
			"use_su": schema.BoolAttribute{
				MarkdownDescription: "Use su instead of sudo (Linux only).",
				Optional:            true,
			},
			"authentication_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type for Linux: `Password` or `PrivateKey`.",
				Optional:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "SSH private key (Linux only, PrivateKey auth).",
				Optional:            true,
				Sensitive:           true,
			},
			"passphrase": schema.StringAttribute{
				MarkdownDescription: "Private key passphrase (Linux only).",
				Optional:            true,
				Sensitive:           true,
			},
			"root_password": schema.StringAttribute{
				MarkdownDescription: "Root password for su elevation (Linux only).",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *Credential) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *Credential) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	var result models.CredentialsModel
	if err := r.client.PostJSON(ctx, client.PathCredentials, payload, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create credential",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(result.ID)
	r.syncModelFromAPI(&data, &result)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Credential) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result models.CredentialsModel
	endpoint := fmt.Sprintf(client.PathCredentialByID, data.ID.ValueString())
	if err := r.client.GetJSON(ctx, endpoint, &result); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read credential",
			fmt.Sprintf("API error for credential %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	r.syncModelFromAPI(&data, &result)
	// Password is never returned by the API — keep current state value.
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Credential) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildSpec(&data)

	endpoint := fmt.Sprintf(client.PathCredentialByID, data.ID.ValueString())
	if err := r.client.PutJSON(ctx, endpoint, payload, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update credential",
			fmt.Sprintf("API error for credential %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *Credential) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf(client.PathCredentialByID, data.ID.ValueString())
	if err := r.deleteCredentialWithRetries(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete credential",
			fmt.Sprintf("API error for credential %s: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *Credential) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewCredential returns a new veeam_credential resource instance.
func NewCredential() resource.Resource {
	return &Credential{}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildSpec converts Terraform state to an API request body.
func (r *Credential) buildSpec(data *CredentialModel) interface{} {
	credType := models.ECredentialsType(data.Type.ValueString())

	if credType == models.CredentialsTypeLinux {
		spec := models.LinuxCredentialsSpec{
			CredentialsSpec: models.CredentialsSpec{
				Username:    data.Username.ValueString(),
				Password:    data.Password.ValueString(),
				Description: data.Description.ValueString(),
				Type:        models.CredentialsTypeLinux,
			},
			AuthenticationType: models.EAuthenticationType(data.AuthenticationType.ValueString()),
		}
		if !data.SSHPort.IsNull() && !data.SSHPort.IsUnknown() {
			spec.SSHPort = int(data.SSHPort.ValueInt64())
		}
		if !data.ElevateToRoot.IsNull() {
			spec.ElevateToRoot = data.ElevateToRoot.ValueBool()
		}
		if !data.AddToSudoers.IsNull() {
			spec.AddToSudoers = data.AddToSudoers.ValueBool()
		}
		if !data.UseSu.IsNull() {
			spec.UseSu = data.UseSu.ValueBool()
		}
		if !data.PrivateKey.IsNull() {
			spec.PrivateKey = data.PrivateKey.ValueString()
		}
		if !data.Passphrase.IsNull() {
			spec.Passphrase = data.Passphrase.ValueString()
		}
		if !data.RootPassword.IsNull() {
			spec.RootPassword = data.RootPassword.ValueString()
		}
		return &spec
	}

	// Default: Standard
	return &models.StandardCredentialsSpec{
		CredentialsSpec: models.CredentialsSpec{
			Username:    data.Username.ValueString(),
			Password:    data.Password.ValueString(),
			Description: data.Description.ValueString(),
			Type:        models.CredentialsTypeStandard,
		},
	}
}

// syncModelFromAPI updates Terraform state fields from an API response.
// Password and other sensitive fields are NOT returned by the API.
func (r *Credential) syncModelFromAPI(data *CredentialModel, api *models.CredentialsModel) {
	data.Username = types.StringValue(api.Username)
	data.Description = types.StringValue(api.Description)
	data.Type = types.StringValue(string(api.Type))
}

func (r *Credential) deleteCredentialWithRetries(ctx context.Context, endpoint string) error {
	const maxAttempts = 5

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = r.client.DeleteJSON(ctx, endpoint)
		if lastErr == nil {
			return nil
		}

		if !isCredentialInUseError(lastErr) || attempt == maxAttempts-1 {
			return lastErr
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}

	return lastErr
}

func isCredentialInUseError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unable to delete selected credentials") && strings.Contains(message, "currently in use")
}

package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/pkg/datasources"
	"github.com/patrikcze/terraform-provider-veeam/pkg/resources"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &Provider{}

// Provider defines the provider implementation.
type Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ProviderModel describes the provider data model.
type ProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// New creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "veeam"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for managing Veeam Backup & Replication V13 resources via the REST API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Veeam Backup & Replication server hostname or IP address. " +
					"Can also be set via the `VEEAM_HOST` environment variable.",
				Optional: true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "REST API port (default: 9419). " +
					"Can also be set via the `VEEAM_PORT` environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication (e.g., `DOMAIN\\admin`). " +
					"Can also be set via the `VEEAM_USERNAME` environment variable.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication. " +
					"Can also be set via the `VEEAM_PASSWORD` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification (default: false). " +
					"**WARNING:** Do not use in production. " +
					"Can also be set via the `VEEAM_INSECURE` environment variable.",
				Optional: true,
			},
		},
	}
}

// Configure prepares a Veeam API client for data sources and resources.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve host: config > env var
	host := data.Host.ValueString()
	if host == "" {
		host = os.Getenv("VEEAM_HOST")
	}

	// Resolve port: config > env var > default 9419
	port := 9419
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		port = int(data.Port.ValueInt64())
	} else if envPort := os.Getenv("VEEAM_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	// Resolve username: config > env var
	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("VEEAM_USERNAME")
	}

	// Resolve password: config > env var
	password := data.Password.ValueString()
	if password == "" {
		password = os.Getenv("VEEAM_PASSWORD")
	}

	// Validate required fields — fail closed
	if host == "" {
		resp.Diagnostics.AddError(
			"Missing Host Configuration",
			"The provider requires a host. Set the 'host' attribute or VEEAM_HOST environment variable.",
		)
		return
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing Username Configuration",
			"The provider requires a username. Set the 'username' attribute or VEEAM_USERNAME environment variable.",
		)
		return
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Missing Password Configuration",
			"The provider requires a password. Set the 'password' attribute or VEEAM_PASSWORD environment variable.",
		)
		return
	}

	// Resolve insecure: config > env var > default false
	insecure := false
	if !data.Insecure.IsNull() && !data.Insecure.IsUnknown() {
		insecure = data.Insecure.ValueBool()
	} else if os.Getenv("VEEAM_INSECURE") == "true" {
		insecure = true
	}

	if insecure {
		tflog.Warn(ctx, "TLS verification disabled via provider configuration")
	}

	// Initialize the API client
	veeamClient, err := client.NewVeeamClient(ctx, host, port, username, password, insecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Veeam API Client",
			fmt.Sprintf("Failed to authenticate with Veeam server at %s:%d: %s", host, port, err),
		)
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = veeamClient
	resp.ResourceData = veeamClient
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewBackupJob,
		resources.NewCredential,
		resources.NewEncryptionPassword,
		resources.NewManagedServer,
		resources.NewProtectionGroup,
		resources.NewProxy,
		resources.NewRepository,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewBackupJobsDataSource,
		datasources.NewCredentialsDataSource,
		datasources.NewRepositoriesDataSource,
	}
}

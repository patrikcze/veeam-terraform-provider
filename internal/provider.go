package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The Veeam Backup & Replication server hostname or IP address",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication to the Veeam server",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication to the Veeam server",
				Required:            true,
				Sensitive:           true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification",
				Optional:            true,
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

	// Validate configuration
	if data.Host.IsNull() || data.Host.IsUnknown() || data.Host.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Host Configuration",
			"The provider requires a host to be configured. "+
				"Please provide a value for the host attribute.",
		)
		return
	}

	if data.Username.IsNull() || data.Username.IsUnknown() || data.Username.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Username Configuration",
			"The provider requires a username to be configured. "+
				"Please provide a value for the username attribute.",
		)
		return
	}

	if data.Password.IsNull() || data.Password.IsUnknown() || data.Password.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Password Configuration",
			"The provider requires a password to be configured. "+
				"Please provide a value for the password attribute.",
		)
		return
	}

	// Set default value for insecure if not provided
	insecure := false
	if !data.Insecure.IsNull() && !data.Insecure.IsUnknown() {
		insecure = data.Insecure.ValueBool()
	}

	// Initialize the API client
	client, err := NewAPIClient(data.Host.ValueString(), data.Username.ValueString(), data.Password.ValueString(), insecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Veeam API Client",
			fmt.Sprintf("An unexpected error occurred when creating the Veeam API client: %s", err),
		)
		return
	}

	// Make the Veeam client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Resources will be added here
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be added here
	}
}

// APIClient represents the Veeam API client.
type APIClient struct {
	Host     string
	Username string
	Password string
	Insecure bool
	// Add other client fields as needed
}

// NewAPIClient creates a new Veeam API client instance.
func NewAPIClient(host, username, password string, insecure bool) (*APIClient, error) {
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	client := &APIClient{
		Host:     host,
		Username: username,
		Password: password,
		Insecure: insecure,
	}

	// Initialize HTTP client, authenticate, etc.
	// This would typically involve creating an HTTP client and
	// performing initial authentication with the Veeam server

	return client, nil
}

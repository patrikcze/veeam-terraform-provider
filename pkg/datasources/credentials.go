package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

var (
	_ datasource.DataSource              = &CredentialsDataSource{}
	_ datasource.DataSourceWithConfigure = &CredentialsDataSource{}
)

// CredentialsDataSource lists Veeam credentials.
type CredentialsDataSource struct {
	client client.APIClient
}

type CredentialsDataSourceModel struct {
	ID          types.String          `tfsdk:"id"`
	Credentials []CredentialDataModel `tfsdk:"credentials"`
}

type CredentialDataModel struct {
	ID          types.String `tfsdk:"id"`
	Username    types.String `tfsdk:"username"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

func NewCredentialsDataSource() datasource.DataSource {
	return &CredentialsDataSource{}
}

func (d *CredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credentials"
}

func (d *CredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Veeam credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"credentials": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"username":    schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"type":        schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *CredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient.")
		return
	}
	d.client = c
}

func (d *CredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CredentialsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResult []map[string]interface{}
	if err := d.client.GetJSON(ctx, client.PathCredentials, &apiResult); err != nil {
		resp.Diagnostics.AddError(
			"Failed to list credentials",
			fmt.Sprintf("API error: %s", err),
		)
		return
	}

	creds := make([]CredentialDataModel, len(apiResult))
	for i, c := range apiResult {
		creds[i] = CredentialDataModel{
			ID:          types.StringValue(getStringValue(c, "id")),
			Username:    types.StringValue(getStringValue(c, "username")),
			Description: types.StringValue(getStringValue(c, "description")),
			Type:        types.StringValue(getStringValue(c, "type")),
		}
	}

	data.ID = types.StringValue("credentials")
	data.Credentials = creds
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

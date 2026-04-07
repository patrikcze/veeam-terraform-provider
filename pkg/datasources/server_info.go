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
	_ datasource.DataSource              = &ServerInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &ServerInfoDataSource{}
)

type ServerInfoDataSource struct {
	client client.APIClient
}

type ServerInfoDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	InstallationID types.String `tfsdk:"installation_id"`
	ServerName     types.String `tfsdk:"server_name"`
	BuildNumber    types.String `tfsdk:"build_number"`
	Version        types.String `tfsdk:"version"`
}

func NewServerInfoDataSource() datasource.DataSource { return &ServerInfoDataSource{} }

func (d *ServerInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_info"
}

func (d *ServerInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":              schema.StringAttribute{Computed: true},
		"installation_id": schema.StringAttribute{Computed: true},
		"server_name":     schema.StringAttribute{Computed: true},
		"build_number":    schema.StringAttribute{Computed: true},
		"version":         schema.StringAttribute{Computed: true},
	}}
}

func (d *ServerInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider.")
		return
	}
	d.client = c
}

func (d *ServerInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerInfoDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := d.client.GetJSON(ctx, client.PathServerInfo, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read server info", fmt.Sprintf("API error: %s", err))
		return
	}

	payload := unwrapObjectData(result)

	data.ID = types.StringValue("server_info")
	data.InstallationID = types.StringValue(getFirstStringValue(payload, "vbrId", "installationId", "installationID", "installation_id"))
	data.ServerName = types.StringValue(getFirstStringValue(payload, "name", "serverName", "server_name"))
	data.BuildNumber = types.StringValue(getFirstStringValue(payload, "buildVersion", "buildNumber", "build_number", "build"))
	data.Version = types.StringValue(getFirstStringValue(payload, "buildVersion", "version", "productVersion", "apiVersion"))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func unwrapObjectData(data map[string]interface{}) map[string]interface{} {
	if raw, ok := data["data"]; ok {
		if nested, ok := raw.(map[string]interface{}); ok {
			return nested
		}
	}

	return data
}

func getFirstStringValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value := getStringValue(data, key)
		if value != "" {
			return value
		}
	}

	return ""
}

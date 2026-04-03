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
	_ datasource.DataSource              = &ServerTimeDataSource{}
	_ datasource.DataSourceWithConfigure = &ServerTimeDataSource{}
)

type ServerTimeDataSource struct {
	client client.APIClient
}

type ServerTimeDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ServerTime types.String `tfsdk:"server_time"`
	TimeZone   types.String `tfsdk:"time_zone"`
	UTCOffset  types.String `tfsdk:"utc_offset"`
}

func NewServerTimeDataSource() datasource.DataSource { return &ServerTimeDataSource{} }

func (d *ServerTimeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_time"
}

func (d *ServerTimeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"server_time": schema.StringAttribute{Computed: true},
		"time_zone":   schema.StringAttribute{Computed: true},
		"utc_offset":  schema.StringAttribute{Computed: true},
	}}
}

func (d *ServerTimeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerTimeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerTimeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := d.client.GetJSON(ctx, client.PathServerTime, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read server time", fmt.Sprintf("API error: %s", err))
		return
	}

	payload := unwrapObjectData(result)

	data.ID = types.StringValue("server-time")
	data.ServerTime = types.StringValue(getFirstStringValue(payload, "serverTime", "time", "currentTime"))
	data.TimeZone = types.StringValue(getFirstStringValue(payload, "timeZone", "timezone"))
	data.UTCOffset = types.StringValue(getFirstStringValue(payload, "utcOffset", "utcoffset", "offset"))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

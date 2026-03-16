package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

var (
	_ datasource.DataSource              = &LicenseDataSource{}
	_ datasource.DataSourceWithConfigure = &LicenseDataSource{}
)

type LicenseDataSource struct {
	client client.APIClient
}

type LicenseDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	Status             types.String `tfsdk:"status"`
	LicensedTo         types.String `tfsdk:"licensed_to"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	LicensedSockets    types.Int64  `tfsdk:"licensed_sockets"`
	ConsumedSockets    types.Int64  `tfsdk:"consumed_sockets"`
	LicensedInstances  types.Int64  `tfsdk:"licensed_instances"`
	ConsumedInstances  types.Int64  `tfsdk:"consumed_instances"`
	LicensedCapacityTB types.Int64  `tfsdk:"licensed_capacity_tb"`
	ConsumedCapacityTB types.Int64  `tfsdk:"consumed_capacity_tb"`
}

func NewLicenseDataSource() datasource.DataSource { return &LicenseDataSource{} }

func (d *LicenseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (d *LicenseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true},
		"type":                 schema.StringAttribute{Computed: true},
		"status":               schema.StringAttribute{Computed: true},
		"licensed_to":          schema.StringAttribute{Computed: true},
		"expiration_date":      schema.StringAttribute{Computed: true},
		"licensed_sockets":     schema.Int64Attribute{Computed: true},
		"consumed_sockets":     schema.Int64Attribute{Computed: true},
		"licensed_instances":   schema.Int64Attribute{Computed: true},
		"consumed_instances":   schema.Int64Attribute{Computed: true},
		"licensed_capacity_tb": schema.Int64Attribute{Computed: true},
		"consumed_capacity_tb": schema.Int64Attribute{Computed: true},
	}}
}

func (d *LicenseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInstance, ok := req.ProviderData.(client.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected client.APIClient from provider.")
		return
	}
	d.client = clientInstance
}

func (d *LicenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LicenseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var license models.LicenseModel
	if err := d.client.GetJSON(ctx, client.PathLicense, &license); err != nil {
		resp.Diagnostics.AddError("Failed to read license", fmt.Sprintf("API error: %s", err))
		return
	}

	var sockets map[string]interface{}
	_ = d.client.GetJSON(ctx, client.PathLicenseSockets, &sockets)
	var instances map[string]interface{}
	_ = d.client.GetJSON(ctx, client.PathLicenseInstances, &instances)
	var capacity map[string]interface{}
	_ = d.client.GetJSON(ctx, client.PathLicenseCapacity, &capacity)

	data.ID = types.StringValue("license")
	data.Type = types.StringValue(license.Type)
	data.Status = types.StringValue(license.Status)
	data.LicensedTo = types.StringValue(license.LicensedTo)
	data.ExpirationDate = types.StringValue(license.Expiration)
	data.LicensedSockets = types.Int64Value(getInt64Value(sockets, "licensedSocketsNumber"))
	data.ConsumedSockets = types.Int64Value(getInt64Value(sockets, "consumedSocketsNumber"))
	data.LicensedInstances = types.Int64Value(getInt64Value(instances, "licensedInstancesNumber"))
	data.ConsumedInstances = types.Int64Value(getInt64Value(instances, "consumedInstancesNumber"))
	data.LicensedCapacityTB = types.Int64Value(getInt64Value(capacity, "licensedCapacityTb"))
	data.ConsumedCapacityTB = types.Int64Value(getInt64Value(capacity, "consumedCapacityTb"))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

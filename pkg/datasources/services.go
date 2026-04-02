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
	_ datasource.DataSource              = &ServicesDataSource{}
	_ datasource.DataSourceWithConfigure = &ServicesDataSource{}
)

type ServicesDataSource struct {
	client client.APIClient
}

type ServicesDataSourceModel struct {
	ID       types.String       `tfsdk:"id"`
	Services []ServiceDataModel `tfsdk:"services"`
}

type ServiceDataModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Status  types.String `tfsdk:"status"`
	Version types.String `tfsdk:"version"`
}

func NewServicesDataSource() datasource.DataSource { return &ServicesDataSource{} }

func (d *ServicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_services"
}

func (d *ServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"services": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true},
			"name":    schema.StringAttribute{Computed: true},
			"status":  schema.StringAttribute{Computed: true},
			"version": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *ServicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathServices)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list services", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]ServiceDataModel, len(items))
	for i, item := range items {
		mapped[i] = ServiceDataModel{
			ID:      types.StringValue(getStringValue(item, "id")),
			Name:    types.StringValue(getStringValue(item, "name")),
			Status:  types.StringValue(getStringValue(item, "status")),
			Version: types.StringValue(getStringValue(item, "version")),
		}
	}

	data.ID = types.StringValue("services")
	data.Services = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

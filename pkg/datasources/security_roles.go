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
	_ datasource.DataSource              = &SecurityRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &SecurityRolesDataSource{}
)

type SecurityRolesDataSource struct {
	client client.APIClient
}

type SecurityRolesDataSourceModel struct {
	ID     types.String            `tfsdk:"id"`
	RoleID types.String            `tfsdk:"role_id"`
	Roles  []SecurityRoleDataModel `tfsdk:"roles"`
}

type SecurityRoleDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewSecurityRolesDataSource() datasource.DataSource { return &SecurityRolesDataSource{} }

func (d *SecurityRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_roles"
}

func (d *SecurityRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":      schema.StringAttribute{Computed: true},
		"role_id": schema.StringAttribute{Optional: true},
		"roles": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *SecurityRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) SecurityRoleDataModel {
		return SecurityRoleDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	if !data.RoleID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathSecurityRoleByID, data.RoleID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read security role", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Roles = []SecurityRoleDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("security_role", data.RoleID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathSecurityRoles)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list security roles", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]SecurityRoleDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("security_roles")
	data.Roles = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ datasource.DataSource              = &SecurityUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &SecurityUsersDataSource{}
)

type SecurityUsersDataSource struct {
	client client.APIClient
}

type SecurityUsersDataSourceModel struct {
	ID     types.String             `tfsdk:"id"`
	UserID types.String             `tfsdk:"user_id"`
	Users  []SecurityUserDataModel2 `tfsdk:"users"`
}

type SecurityUserDataModel2 struct {
	ID          types.String `tfsdk:"id"`
	Login       types.String `tfsdk:"login"`
	Description types.String `tfsdk:"description"`
	RoleID      types.String `tfsdk:"role_id"`
}

func NewSecurityUsersDataSource() datasource.DataSource { return &SecurityUsersDataSource{} }

func (d *SecurityUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_users"
}

func (d *SecurityUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":      schema.StringAttribute{Computed: true},
		"user_id": schema.StringAttribute{Optional: true},
		"users": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"login":       schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"role_id":     schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *SecurityUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityUsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) SecurityUserDataModel2 {
		// VBR API returns the account identifier as "name" (e.g. "DOMAIN\\user"),
		// not as "login". The role is a nested array; we expose the first role's id.
		return SecurityUserDataModel2{
			ID:          types.StringValue(getStringValue(item, "id")),
			Login:       types.StringValue(getFirstStringValue(item, "name", "login")),
			Description: types.StringValue(getStringValue(item, "description")),
			RoleID:      types.StringValue(firstNestedID(item, "roles")),
		}
	}

	if !data.UserID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathSecurityUserByID, data.UserID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read security user", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Users = []SecurityUserDataModel2{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("security_user", data.UserID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathSecurityUsers)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list security users", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]SecurityUserDataModel2, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("security_users")
	data.Users = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

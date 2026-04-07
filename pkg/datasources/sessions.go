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
	_ datasource.DataSource              = &SessionsDataSource{}
	_ datasource.DataSourceWithConfigure = &SessionsDataSource{}
)

type SessionsDataSource struct{ client client.APIClient }

type SessionsDataSourceModel struct {
	ID        types.String       `tfsdk:"id"`
	SessionID types.String       `tfsdk:"session_id"`
	Sessions  []SessionDataModel `tfsdk:"sessions"`
}

type SessionDataModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	JobID        types.String `tfsdk:"job_id"`
	SessionType  types.String `tfsdk:"session_type"`
	State        types.String `tfsdk:"state"`
	Result       types.String `tfsdk:"result"`
	CreationTime types.String `tfsdk:"creation_time"`
	EndTime      types.String `tfsdk:"end_time"`
}

func NewSessionsDataSource() datasource.DataSource { return &SessionsDataSource{} }
func (d *SessionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sessions"
}
func (d *SessionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true},
		"session_id": schema.StringAttribute{Optional: true},
		"sessions": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true}, "job_id": schema.StringAttribute{Computed: true},
			"session_type": schema.StringAttribute{Computed: true}, "state": schema.StringAttribute{Computed: true}, "result": schema.StringAttribute{Computed: true},
			"creation_time": schema.StringAttribute{Computed: true}, "end_time": schema.StringAttribute{Computed: true},
		}}},
	}}
}
func (d *SessionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SessionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SessionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.SessionID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathSessionByID, data.SessionID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read session", fmt.Sprintf("API error: %s", err))
			return
		}
		data.Sessions = []SessionDataModel{{
			ID:           types.StringValue(getStringValue(item, "id")),
			Name:         types.StringValue(getStringValue(item, "name")),
			JobID:        types.StringValue(getStringValue(item, "jobId")),
			SessionType:  types.StringValue(getStringValue(item, "sessionType")),
			State:        types.StringValue(getStringValue(item, "state")),
			Result:       types.StringValue(sessionResultString(item)),
			CreationTime: types.StringValue(getStringValue(item, "creationTime")),
			EndTime:      types.StringValue(getStringValue(item, "endTime")),
		}}
		data.ID = types.StringValue(normalizeDataSourceID("session", data.SessionID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}
	items, err := fetchList(ctx, d.client.GetJSON, client.PathSessions)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list sessions", fmt.Sprintf("API error: %s", err))
		return
	}
	out := make([]SessionDataModel, len(items))
	for i, item := range items {
		out[i] = SessionDataModel{
			ID:           types.StringValue(getStringValue(item, "id")),
			Name:         types.StringValue(getStringValue(item, "name")),
			JobID:        types.StringValue(getStringValue(item, "jobId")),
			SessionType:  types.StringValue(getStringValue(item, "sessionType")),
			State:        types.StringValue(getStringValue(item, "state")),
			Result:       types.StringValue(sessionResultString(item)),
			CreationTime: types.StringValue(getStringValue(item, "creationTime")),
			EndTime:      types.StringValue(getStringValue(item, "endTime")),
		}
	}
	data.ID = types.StringValue("sessions")
	data.Sessions = out
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// sessionResultString extracts the result string from a session item.
// The VBR API returns result as a nested object {"result":"Success","message":"..."},
// not as a flat string, so we must look inside the nested map first.
func sessionResultString(item map[string]interface{}) string {
	if nested, ok := item["result"].(map[string]interface{}); ok {
		return getStringValue(nested, "result")
	}
	// fallback: some mock/test payloads use a flat string directly
	return getStringValue(item, "result")
}

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
	_ datasource.DataSource              = &TaskSessionsDataSource{}
	_ datasource.DataSourceWithConfigure = &TaskSessionsDataSource{}
)

type TaskSessionsDataSource struct {
	client client.APIClient
}

type TaskSessionsDataSourceModel struct {
	ID            types.String           `tfsdk:"id"`
	TaskSessionID types.String           `tfsdk:"task_session_id"`
	TaskSessions  []TaskSessionDataModel `tfsdk:"task_sessions"`
}

type TaskSessionDataModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	SessionID types.String `tfsdk:"session_id"`
	Status    types.String `tfsdk:"status"`
	StartTime types.String `tfsdk:"start_time"`
	EndTime   types.String `tfsdk:"end_time"`
}

func NewTaskSessionsDataSource() datasource.DataSource { return &TaskSessionsDataSource{} }

func (d *TaskSessionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_sessions"
}

func (d *TaskSessionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":              schema.StringAttribute{Computed: true},
		"task_session_id": schema.StringAttribute{Optional: true},
		"task_sessions": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"name":       schema.StringAttribute{Computed: true},
			"session_id": schema.StringAttribute{Computed: true},
			"status":     schema.StringAttribute{Computed: true},
			"start_time": schema.StringAttribute{Computed: true},
			"end_time":   schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *TaskSessionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TaskSessionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TaskSessionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapOne := func(item map[string]interface{}) TaskSessionDataModel {
		return TaskSessionDataModel{
			ID:        types.StringValue(getStringValue(item, "id")),
			Name:      types.StringValue(getStringValue(item, "name")),
			SessionID: types.StringValue(getStringValue(item, "sessionId")),
			Status:    types.StringValue(getStringValue(item, "status")),
			StartTime: types.StringValue(getStringValue(item, "startTime")),
			EndTime:   types.StringValue(getStringValue(item, "endTime")),
		}
	}

	if !data.TaskSessionID.IsNull() {
		var item map[string]interface{}
		if err := d.client.GetJSON(ctx, fmt.Sprintf(client.PathTaskSessionByID, data.TaskSessionID.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Failed to read task session", fmt.Sprintf("API error: %s", err))
			return
		}
		data.TaskSessions = []TaskSessionDataModel{mapOne(item)}
		data.ID = types.StringValue(normalizeDataSourceID("task_session", data.TaskSessionID.ValueString()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	items, err := fetchList(ctx, d.client.GetJSON, client.PathTaskSessions)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list task sessions", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]TaskSessionDataModel, len(items))
	for i, item := range items {
		mapped[i] = mapOne(item)
	}

	data.ID = types.StringValue("task_sessions")
	data.TaskSessions = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

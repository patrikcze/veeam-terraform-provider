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
	_ datasource.DataSource              = &JobStatesDataSource{}
	_ datasource.DataSourceWithConfigure = &JobStatesDataSource{}
)

type JobStatesDataSource struct{ client client.APIClient }

type JobStatesDataSourceModel struct {
	ID     types.String        `tfsdk:"id"`
	JobID  types.String        `tfsdk:"job_id"`
	States []JobStateDataModel `tfsdk:"states"`
}

type JobStateDataModel struct {
	JobID      types.String `tfsdk:"job_id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Status     types.String `tfsdk:"status"`
	LastResult types.String `tfsdk:"last_result"`
	LastRun    types.String `tfsdk:"last_run"`
}

func NewJobStatesDataSource() datasource.DataSource { return &JobStatesDataSource{} }
func (d *JobStatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_states"
}
func (d *JobStatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true}, "job_id": schema.StringAttribute{Optional: true},
		"states": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"job_id": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true}, "type": schema.StringAttribute{Computed: true}, "status": schema.StringAttribute{Computed: true}, "last_result": schema.StringAttribute{Computed: true}, "last_run": schema.StringAttribute{Computed: true},
		}}},
	}}
}
func (d *JobStatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *JobStatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobStatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := fetchList(ctx, d.client.GetJSON, client.PathJobStates)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list job states", fmt.Sprintf("API error: %s", err))
		return
	}
	out := make([]JobStateDataModel, 0, len(items))
	for _, item := range items {
		jobID := getStringValue(item, "jobId")
		if !data.JobID.IsNull() && data.JobID.ValueString() != jobID {
			continue
		}
		out = append(out, JobStateDataModel{JobID: types.StringValue(jobID), Name: types.StringValue(getStringValue(item, "name")), Type: types.StringValue(getStringValue(item, "type")), Status: types.StringValue(getStringValue(item, "status")), LastResult: types.StringValue(getStringValue(item, "lastResult")), LastRun: types.StringValue(getStringValue(item, "lastRun"))})
	}
	if !data.JobID.IsNull() {
		data.ID = types.StringValue(normalizeDataSourceID("job_states", data.JobID.ValueString()))
	} else {
		data.ID = types.StringValue("job_states")
	}
	data.States = out
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ datasource.DataSource              = &SecurityAnalyzerDataSource{}
	_ datasource.DataSourceWithConfigure = &SecurityAnalyzerDataSource{}
)

type SecurityAnalyzerDataSource struct {
	client client.APIClient
}

type SecurityAnalyzerDataSourceModel struct {
	ID            types.String                    `tfsdk:"id"`
	LastRunTime   types.String                    `tfsdk:"last_run_time"`
	LastRunStatus types.String                    `tfsdk:"last_run_status"`
	BestPractices []SecurityBestPracticeDataModel `tfsdk:"best_practices"`
}

type SecurityBestPracticeDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
}

func NewSecurityAnalyzerDataSource() datasource.DataSource { return &SecurityAnalyzerDataSource{} }

func (d *SecurityAnalyzerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_analyzer"
}

func (d *SecurityAnalyzerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":              schema.StringAttribute{Computed: true},
		"last_run_time":   schema.StringAttribute{Computed: true},
		"last_run_status": schema.StringAttribute{Computed: true},
		"best_practices": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
		}}},
	}}
}

func (d *SecurityAnalyzerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityAnalyzerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityAnalyzerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch last run metadata.
	var lastRun map[string]interface{}
	if err := d.client.GetJSON(ctx, client.PathSecurityAnalyzerLastRun, &lastRun); err != nil {
		resp.Diagnostics.AddError("Failed to read security analyzer last run", fmt.Sprintf("API error: %s", err))
		return
	}
	lastRunPayload := unwrapObjectData(lastRun)

	// Fetch best practices list.
	practices, err := fetchList(ctx, d.client.GetJSON, client.PathSecurityAnalyzerBestPractices)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list security analyzer best practices", fmt.Sprintf("API error: %s", err))
		return
	}

	mapped := make([]SecurityBestPracticeDataModel, len(practices))
	for i, item := range practices {
		mapped[i] = SecurityBestPracticeDataModel{
			ID:          types.StringValue(getStringValue(item, "id")),
			Name:        types.StringValue(getStringValue(item, "name")),
			Status:      types.StringValue(getStringValue(item, "status")),
			Description: types.StringValue(getStringValue(item, "description")),
		}
	}

	data.ID = types.StringValue("security-analyzer")
	data.LastRunTime = types.StringValue(getFirstStringValue(lastRunPayload, "lastRunTime", "runTime", "time"))
	data.LastRunStatus = types.StringValue(getFirstStringValue(lastRunPayload, "lastRunStatus", "status", "result"))
	data.BestPractices = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

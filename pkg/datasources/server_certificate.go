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
	_ datasource.DataSource              = &ServerCertificateDataSource{}
	_ datasource.DataSourceWithConfigure = &ServerCertificateDataSource{}
)

type ServerCertificateDataSource struct {
	client client.APIClient
}

type ServerCertificateDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Thumbprint   types.String `tfsdk:"thumbprint"`
	Subject      types.String `tfsdk:"subject"`
	IssuedBy     types.String `tfsdk:"issued_by"`
	ValidFrom    types.String `tfsdk:"valid_from"`
	ValidTo      types.String `tfsdk:"valid_to"`
	SerialNumber types.String `tfsdk:"serial_number"`
}

func NewServerCertificateDataSource() datasource.DataSource { return &ServerCertificateDataSource{} }

func (d *ServerCertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_certificate"
}

func (d *ServerCertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true},
		"thumbprint":    schema.StringAttribute{Computed: true},
		"subject":       schema.StringAttribute{Computed: true},
		"issued_by":     schema.StringAttribute{Computed: true},
		"valid_from":    schema.StringAttribute{Computed: true},
		"valid_to":      schema.StringAttribute{Computed: true},
		"serial_number": schema.StringAttribute{Computed: true},
	}}
}

func (d *ServerCertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerCertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := d.client.GetJSON(ctx, client.PathServerCertificate, &result); err != nil {
		resp.Diagnostics.AddError("Failed to read server certificate", fmt.Sprintf("API error: %s", err))
		return
	}

	payload := unwrapObjectData(result)

	data.ID = types.StringValue("server-certificate")
	data.Thumbprint = types.StringValue(getFirstStringValue(payload, "thumbprint", "certificateThumbprint"))
	data.Subject = types.StringValue(getFirstStringValue(payload, "subject", "subjectName"))
	data.IssuedBy = types.StringValue(getFirstStringValue(payload, "issuedBy", "issuer"))
	data.ValidFrom = types.StringValue(getFirstStringValue(payload, "validFrom", "notBefore"))
	data.ValidTo = types.StringValue(getFirstStringValue(payload, "validTo", "notAfter"))
	data.SerialNumber = types.StringValue(getFirstStringValue(payload, "serialNumber", "serial"))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

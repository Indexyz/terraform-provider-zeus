// Copyright (c) WANIX Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PortDataSource{}

func NewPortDataSource() datasource.DataSource {
	return &PortDataSource{}
}

type PortDataSource struct {
	client *zeusapi.Client
}

type portDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	AssignID   types.String `tfsdk:"assign_id"`
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	TargetPort types.Int64  `tfsdk:"target_port"`
	Service    types.String `tfsdk:"service"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (d *PortDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port"
}

func (d *PortDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup a Zeus port rule by ID",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Port rule ID",
				Required:            true,
			},
			"assign_id": schema.StringAttribute{Computed: true},
			"host":      schema.StringAttribute{Computed: true},
			"port":      schema.Int64Attribute{Computed: true},
			"target_port": schema.Int64Attribute{
				Computed: true,
			},
			"service": schema.StringAttribute{Computed: true},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *PortDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zeusapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *zeusapi.Client")
		return
	}
	d.client = client
}

func (d *PortDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data portDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	portInfo, err := d.client.GetPortByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read port failed", err.Error())
		return
	}

	data.AssignID = types.StringValue(portInfo.AssignID)
	data.Host = types.StringValue(portInfo.Host)
	data.Port = types.Int64Value(portInfo.Port)
	data.TargetPort = types.Int64Value(portInfo.TargetPort)
	data.Service = types.StringValue(portInfo.Service)
	data.CreatedAt = types.StringValue(portInfo.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

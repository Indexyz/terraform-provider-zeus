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

var _ datasource.DataSource = &PoolDataSource{}

func NewPoolDataSource() datasource.DataSource {
	return &PoolDataSource{}
}

type PoolDataSource struct {
	client *zeusapi.Client
}

type poolDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Region       types.String `tfsdk:"region"`
	FriendlyName types.String `tfsdk:"friendly_name"`
	Begin        types.String `tfsdk:"begin"`
	End          types.String `tfsdk:"end"`
	GatewayIP    types.String `tfsdk:"gateway_ip"`
	State        types.List   `tfsdk:"state"`
	Size         types.Int64  `tfsdk:"size"`
}

func (d *PoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (d *PoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup a Zeus pool by ID",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Pool ID",
				Required:            true,
			},
			"region": schema.StringAttribute{
				Computed: true,
			},
			"friendly_name": schema.StringAttribute{
				Computed: true,
			},
			"begin": schema.StringAttribute{
				Computed: true,
			},
			"end": schema.StringAttribute{
				Computed: true,
			},
			"gateway_ip": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *PoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data poolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := d.client.GetPoolByID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read pool failed", err.Error())
		return
	}

	data.Region = types.StringValue(detail.Region)
	data.FriendlyName = types.StringValue(detail.FriendlyName)
	data.Begin = types.StringValue(detail.Begin)
	data.End = types.StringValue(detail.End)
	data.GatewayIP = types.StringValue(detail.Gateway)
	data.Size = types.Int64Value(int64(len(detail.State)))
	data.State, _ = types.ListValueFrom(ctx, types.Int64Type, detail.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

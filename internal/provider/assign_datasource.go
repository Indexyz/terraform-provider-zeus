// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AssignDataSource{}

func NewAssignDataSource() datasource.DataSource {
	return &AssignDataSource{}
}

type AssignDataSource struct {
	client *zeusapi.Client
}

type assignDataSourceModel struct {
	ID        types.String  `tfsdk:"id"`
	Key       types.String  `tfsdk:"key"`
	Type      types.String  `tfsdk:"type"`
	Data      types.Dynamic `tfsdk:"data"`
	CreatedAt types.String  `tfsdk:"created_at"`
	Leases    types.Map     `tfsdk:"leases"`
}

func (d *AssignDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assign"
}

func (d *AssignDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup assign info by ID",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Assign ID",
				Required:            true,
			},
			"key": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"data": schema.DynamicAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"leases": schema.MapAttribute{
				Computed:    true,
				ElementType: leaseAttrType(),
			},
		},
	}
}

func (d *AssignDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AssignDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data assignDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assign, err := d.client.GetAssign(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read assign failed", err.Error())
		return
	}

	data.Key = types.StringValue(assign.Key)
	data.Type = types.StringValue(assign.Type)
	data.CreatedAt = types.StringValue(assign.CreatedAt)
	dyn, err := dynamicFromInterface(assign.Data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assign data", err.Error())
		return
	}
	data.Data = dyn
	data.Leases = encodeLeases(assign.Leases)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

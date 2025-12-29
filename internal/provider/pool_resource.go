// Copyright (c) WANIX Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PoolResource{}
var _ resource.ResourceWithImportState = &PoolResource{}

func NewPoolResource() resource.Resource {
	return &PoolResource{}
}

type PoolResource struct {
	client *zeusapi.Client
}

type poolModel struct {
	ID           types.String `tfsdk:"id"`
	Start        types.Int64  `tfsdk:"start"`
	Gateway      types.Int64  `tfsdk:"gateway"`
	Size         types.Int64  `tfsdk:"size"`
	Region       types.String `tfsdk:"region"`
	FriendlyName types.String `tfsdk:"friendly_name"`
	Begin        types.String `tfsdk:"begin"`
	End          types.String `tfsdk:"end"`
	GatewayIP    types.String `tfsdk:"gateway_ip"`
	State        types.List   `tfsdk:"state"`
}

func (r *PoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (r *PoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Zeus address pool",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start": schema.Int64Attribute{
				MarkdownDescription: "Start address (integer form)",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"gateway": schema.Int64Attribute{
				MarkdownDescription: "Gateway address (integer form)",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Pool size",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
		},
	}
}

func (r *PoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zeusapi.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *zeusapi.Client")
		return
	}
	r.client = client
}

func (r *PoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan poolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.client.CreatePool(ctx, zeusapi.CreatePoolRequest{
		Start:   plan.Start.ValueInt64(),
		Gateway: plan.Gateway.ValueInt64(),
		Size:    plan.Size.ValueInt64(),
		Region:  plan.Region.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create pool failed", err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.ID)

	if err := r.refresh(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Read pool after create failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state poolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.refresh(ctx, &state); err != nil {
		var apiErr *zeusapi.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read pool failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan poolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state poolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePool(ctx, state.ID.ValueString()); err != nil {
		var apiErr *zeusapi.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return
		}
		resp.Diagnostics.AddError("Delete pool failed", err.Error())
	}
}

func (r *PoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PoolResource) refresh(ctx context.Context, m *poolModel) error {
	detail, err := r.client.GetPoolByID(ctx, m.ID.ValueString())
	if err != nil {
		return err
	}

	m.Region = types.StringValue(detail.Region)
	m.FriendlyName = types.StringValue(detail.FriendlyName)
	m.Begin = types.StringValue(detail.Begin)
	m.End = types.StringValue(detail.End)
	m.GatewayIP = types.StringValue(detail.Gateway)

	if detail.State != nil {
		m.State, _ = types.ListValueFrom(ctx, types.Int64Type, detail.State)
		m.Size = types.Int64Value(int64(len(detail.State)))
	}

	return nil
}

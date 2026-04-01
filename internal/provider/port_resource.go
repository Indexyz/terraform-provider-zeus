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

var _ resource.Resource = &PortResource{}
var _ resource.ResourceWithImportState = &PortResource{}

func NewPortResource() resource.Resource {
	return &PortResource{}
}

type PortResource struct {
	client *zeusapi.Client
}

type portModel struct {
	ID         types.String `tfsdk:"id"`
	AssignID   types.String `tfsdk:"assign_id"`
	ScopeHost  types.String `tfsdk:"scope_host"`
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	TargetPort types.Int64  `tfsdk:"target_port"`
	Service    types.String `tfsdk:"service"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (r *PortResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port"
}

func (r *PortResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Zeus port forwarding rule",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"assign_id": schema.StringAttribute{
				MarkdownDescription: "Assign ID backing the port rule",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_host": schema.StringAttribute{
				MarkdownDescription: "Optional create-time scope host sent as X-Portd-Host. This value is not reconstructed during read or import, so imported resources typically omit it.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Observed Zeus host for the port rule",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Allocated external port",
				Computed:            true,
			},
			"target_port": schema.Int64Attribute{
				MarkdownDescription: "Target service port",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				MarkdownDescription: "Service name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *PortResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan portModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.client.CreatePort(ctx, zeusapi.CreatePortRequest{
		AssignID:   plan.AssignID.ValueString(),
		TargetPort: plan.TargetPort.ValueInt64(),
		Service:    plan.Service.ValueString(),
		Host:       plan.ScopeHost.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create port failed", err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.ID)
	if err := r.refresh(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Read port after create failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portModel
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
		resp.Diagnostics.AddError("Read port failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan portModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state portModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePortByID(ctx, state.ID.ValueString()); err != nil {
		var apiErr *zeusapi.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return
		}
		resp.Diagnostics.AddError("Delete port failed", err.Error())
	}
}

func (r *PortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PortResource) refresh(ctx context.Context, m *portModel) error {
	portInfo, err := r.client.GetPortByID(ctx, m.ID.ValueString())
	if err != nil {
		return err
	}

	m.AssignID = types.StringValue(portInfo.AssignID)
	m.Host = types.StringValue(portInfo.Host)
	m.Port = types.Int64Value(portInfo.Port)
	m.TargetPort = types.Int64Value(portInfo.TargetPort)
	m.Service = types.StringValue(portInfo.Service)
	m.CreatedAt = types.StringValue(portInfo.CreatedAt)

	return nil
}

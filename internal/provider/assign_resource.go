// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AssignResource{}
var _ resource.ResourceWithImportState = &AssignResource{}

func NewAssignResource() resource.Resource {
	return &AssignResource{}
}

type AssignResource struct {
	client *zeusapi.Client
}

type assignModel struct {
	ID        types.String  `tfsdk:"id"`
	Region    types.List    `tfsdk:"region"`
	Host      types.String  `tfsdk:"host"`
	Key       types.String  `tfsdk:"key"`
	Type      types.String  `tfsdk:"type"`
	Data      types.Dynamic `tfsdk:"data"`
	CreatedAt types.String  `tfsdk:"created_at"`
	Leases    types.Map     `tfsdk:"leases"`
}

func (r *AssignResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assign"
}

func (r *AssignResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assignment of addresses across regions",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.ListAttribute{
				MarkdownDescription: "Regions to allocate in",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Host identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Business key for idempotency",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type tag",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.DynamicAttribute{
				MarkdownDescription: "Arbitrary JSON payload",
				Optional:            true,
				PlanModifiers: []planmodifier.Dynamic{
					requiresReplaceDynamic(),
				},
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

func (r *AssignResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AssignResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assignModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []string
	resp.Diagnostics.Append(plan.Region.ElementsAs(ctx, &regions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var data any
	if !plan.Data.IsNull() {
		if plan.Data.IsUnknown() {
			resp.Diagnostics.AddError("Invalid data", "data must be known during apply")
			return
		}
		converted, err := dynamicToJSONCompatible(plan.Data)
		if err != nil {
			resp.Diagnostics.AddError("Invalid data", err.Error())
			return
		}
		data = converted
	}

	createResp, err := r.client.CreateAssign(ctx, zeusapi.AssignCreateRequest{
		Region: regions,
		Host:   plan.Host.ValueString(),
		Key:    plan.Key.ValueString(),
		Type:   plan.Type.ValueString(),
		Data:   data,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create assign failed", err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.ID)
	if err := r.refresh(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Read assign after create failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AssignResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assignModel
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
		resp.Diagnostics.AddError("Read assign failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AssignResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan assignModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AssignResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assignModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAssign(ctx, state.ID.ValueString()); err != nil {
		var apiErr *zeusapi.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return
		}
		resp.Diagnostics.AddError("Delete assign failed", err.Error())
	}
}

func (r *AssignResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AssignResource) refresh(ctx context.Context, m *assignModel) error {
	assign, err := r.client.GetAssign(ctx, m.ID.ValueString())
	if err != nil {
		return err
	}

	m.Key = types.StringValue(assign.Key)
	m.Type = types.StringValue(assign.Type)
	m.CreatedAt = types.StringValue(assign.CreatedAt)

	m.Leases = encodeLeases(assign.Leases)
	return nil
}

func leaseAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"address":  types.StringType,
			"gateway":  types.StringType,
			"lease_id": types.StringType,
			"vlan":     types.Int64Type,
		},
	}
}

func encodeLeases(leases map[string]zeusapi.AddressResult) types.Map {
	if len(leases) == 0 {
		return types.MapNull(leaseAttrType())
	}

	entries := make(map[string]attr.Value, len(leases))
	for region, lease := range leases {
		var vlan attr.Value
		if lease.VLAN == nil {
			vlan = types.Int64Null()
		} else {
			vlan = types.Int64Value(*lease.VLAN)
		}
		entries[region] = types.ObjectValueMust(
			leaseAttrType().AttrTypes,
			map[string]attr.Value{
				"address":  types.StringValue(lease.Address),
				"gateway":  types.StringValue(lease.Gateway),
				"lease_id": types.StringValue(lease.LeaseID),
				"vlan":     vlan,
			},
		)
	}

	return types.MapValueMust(leaseAttrType(), entries)
}

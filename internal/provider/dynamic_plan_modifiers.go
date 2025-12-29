// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type requiresReplaceDynamicModifier struct{}

func (m requiresReplaceDynamicModifier) Description(ctx context.Context) string {
	return "requires resource replacement if the value changes"
}

func (m requiresReplaceDynamicModifier) MarkdownDescription(ctx context.Context) string {
	return "requires resource replacement if the value changes"
}

func (m requiresReplaceDynamicModifier) PlanModifyDynamic(ctx context.Context, req planmodifier.DynamicRequest, resp *planmodifier.DynamicResponse) {
	if req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() {
		return
	}

	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

func requiresReplaceDynamic() planmodifier.Dynamic {
	return requiresReplaceDynamicModifier{}
}

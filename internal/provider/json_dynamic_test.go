// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDynamicToJSONCompatible(t *testing.T) {
	t.Parallel()

	nestedAttrTypes := map[string]attr.Type{
		"owner": types.StringType,
	}
	nestedObj, diags := types.ObjectValue(
		nestedAttrTypes,
		map[string]attr.Value{
			"owner": types.StringValue("terraform"),
		},
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	ports, diags := types.ListValue(
		types.Int64Type,
		[]attr.Value{types.Int64Value(80), types.Int64Value(443)},
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	rootObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"env":     types.StringType,
			"enabled": types.BoolType,
			"ports":   types.ListType{ElemType: types.Int64Type},
			"meta":    types.ObjectType{AttrTypes: nestedAttrTypes},
		},
		map[string]attr.Value{
			"env":     types.StringValue("dev"),
			"enabled": types.BoolValue(true),
			"ports":   ports,
			"meta":    nestedObj,
		},
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	got, err := dynamicToJSONCompatible(types.DynamicValue(rootObj))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{
		"env":     "dev",
		"enabled": true,
		"ports":   []any{int64(80), int64(443)},
		"meta": map[string]any{
			"owner": "terraform",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected diff: %s", diff)
	}
}

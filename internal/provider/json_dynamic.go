// Copyright (c) WANIX Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func dynamicToJSONCompatible(v types.Dynamic) (any, error) {
	if v.IsNull() {
		return nil, nil
	}
	if v.IsUnknown() {
		return nil, fmt.Errorf("value must be known")
	}
	if v.UnderlyingValue() == nil {
		return nil, fmt.Errorf("value must be known")
	}
	return attrValueToJSONCompatible(v.UnderlyingValue())
}

func attrValueToJSONCompatible(v attr.Value) (any, error) {
	switch tv := v.(type) {
	case types.String:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("string value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		return tv.ValueString(), nil
	case types.Bool:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("bool value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		return tv.ValueBool(), nil
	case types.Int64:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("int64 value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		return tv.ValueInt64(), nil
	case types.Float64:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("float64 value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		return tv.ValueFloat64(), nil
	case types.Number:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("number value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		bf := tv.ValueBigFloat()
		if bf == nil {
			return nil, fmt.Errorf("number value must be known")
		}
		f, _ := bf.Float64()
		return f, nil
	case types.Map:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("map value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		out := make(map[string]any, len(tv.Elements()))
		for k, ev := range tv.Elements() {
			cv, err := attrValueToJSONCompatible(ev)
			if err != nil {
				return nil, fmt.Errorf("map[%q]: %w", k, err)
			}
			out[k] = cv
		}
		return out, nil
	case types.Object:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("object value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		out := make(map[string]any, len(tv.Attributes()))
		for k, av := range tv.Attributes() {
			cv, err := attrValueToJSONCompatible(av)
			if err != nil {
				return nil, fmt.Errorf("object.%s: %w", k, err)
			}
			out[k] = cv
		}
		return out, nil
	case types.List:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("list value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		out := make([]any, 0, len(tv.Elements()))
		for i, ev := range tv.Elements() {
			cv, err := attrValueToJSONCompatible(ev)
			if err != nil {
				return nil, fmt.Errorf("list[%d]: %w", i, err)
			}
			out = append(out, cv)
		}
		return out, nil
	case types.Set:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("set value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		out := make([]any, 0, len(tv.Elements()))
		for i, ev := range tv.Elements() {
			cv, err := attrValueToJSONCompatible(ev)
			if err != nil {
				return nil, fmt.Errorf("set[%d]: %w", i, err)
			}
			out = append(out, cv)
		}
		return out, nil
	case types.Tuple:
		if tv.IsUnknown() {
			return nil, fmt.Errorf("tuple value must be known")
		}
		if tv.IsNull() {
			return nil, nil
		}
		out := make([]any, 0, len(tv.Elements()))
		for i, ev := range tv.Elements() {
			cv, err := attrValueToJSONCompatible(ev)
			if err != nil {
				return nil, fmt.Errorf("tuple[%d]: %w", i, err)
			}
			out = append(out, cv)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", v)
	}
}

func dynamicFromInterface(v any) (types.Dynamic, error) {
	if v == nil {
		return types.DynamicNull(), nil
	}

	av, err := interfaceToAttrValue(v)
	if err != nil {
		return types.DynamicNull(), err
	}

	return types.DynamicValue(av), nil
}

func interfaceToAttrValue(v any) (attr.Value, error) {
	switch tv := v.(type) {
	case string:
		return types.StringValue(tv), nil
	case bool:
		return types.BoolValue(tv), nil
	case int:
		return types.Int64Value(int64(tv)), nil
	case int64:
		return types.Int64Value(tv), nil
	case float64:
		return types.Float64Value(tv), nil
	case *big.Float:
		if tv == nil {
			return types.NumberNull(), nil
		}
		return types.NumberValue(tv), nil
	case []any:
		elems := make([]attr.Value, 0, len(tv))
		for i, ev := range tv {
			inner, err := interfaceToAttrValue(ev)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("list[%d]: %w", i, err)
			}
			elems = append(elems, types.DynamicValue(inner))
		}
		return types.ListValueMust(types.DynamicType, elems), nil
	case map[string]any:
		elems := make(map[string]attr.Value, len(tv))
		for k, ev := range tv {
			inner, err := interfaceToAttrValue(ev)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("map[%q]: %w", k, err)
			}
			elems[k] = types.DynamicValue(inner)
		}
		return types.MapValueMust(types.DynamicType, elems), nil
	default:
		return types.DynamicNull(), fmt.Errorf("unsupported json type %T", v)
	}
}

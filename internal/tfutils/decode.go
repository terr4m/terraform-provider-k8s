package tfutils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DecodeDynamic decodes an object into a Terraform dynamic value.
func DecodeDynamic(ctx context.Context, obj any, unknownPaths ...path.Path) (types.Dynamic, diag.Diagnostics) {
	if obj == nil {
		return types.DynamicNull(), nil
	}

	val, diags := decodeScalar(ctx, obj, path.Empty(), unknownPaths)
	if diags.HasError() {
		return types.Dynamic{}, diags
	}

	return types.DynamicValue(val), diags
}

// decodeScalar decodes a scalar value into a Terraform attribute value.
func decodeScalar(ctx context.Context, a any, thisPath path.Path, unknownPaths path.Paths) (attr.Value, diag.Diagnostics) {
	isUnknown := unknownPaths.Contains(thisPath)

	switch v := a.(type) {
	case nil:
		if isUnknown {
			return types.DynamicUnknown(), nil
		}
		return types.DynamicNull(), nil
	case int64:
		if isUnknown {
			return types.NumberUnknown(), nil
		}
		return types.NumberValue(big.NewFloat(float64(v))), nil
	case float64:
		if isUnknown {
			return types.NumberUnknown(), nil
		}
		return types.NumberValue(big.NewFloat(v)), nil
	case bool:
		if isUnknown {
			return types.BoolUnknown(), nil
		}
		return types.BoolValue(v), nil
	case string:
		if isUnknown {
			return types.StringUnknown(), nil
		}
		return types.StringValue(v), nil
	case []any:
		return decodeSequence(ctx, v, thisPath, unknownPaths)
	case map[string]any:
		return decodeMapping(ctx, v, thisPath, unknownPaths)
	default:
		diagnostics := diag.Diagnostics{}
		diagnostics.AddError("Unexpected type.", fmt.Sprintf("unexpected type: %T for value %#v", v, v))
		return nil, diagnostics
	}
}

// decodeMapping decodes a mapping value into a Terraform attribute value.
func decodeMapping(ctx context.Context, m map[string]any, thisPath path.Path, unknownPaths path.Paths) (attr.Value, diag.Diagnostics) {
	if unknownPaths.Contains(thisPath) {
		return types.DynamicUnknown(), nil
	}

	l := len(m)
	vm := make(map[string]attr.Value, l)
	tm := make(map[string]attr.Type, l)

	for k, v := range m {
		p := thisPath.AtName(k)
		vv, diags := decodeScalar(ctx, v, p, unknownPaths)
		if diags.HasError() {
			return nil, diags
		}

		vm[k] = vv
		tm[k] = vv.Type(ctx)
	}

	return types.ObjectValue(tm, vm)
}

// decodeSequence decodes a sequence value into a Terraform attribute value.
func decodeSequence(ctx context.Context, s []any, thisPath path.Path, unknownPaths path.Paths) (attr.Value, diag.Diagnostics) {
	if unknownPaths.Contains(thisPath) {
		return types.DynamicUnknown(), nil
	}

	l := len(s)
	vl := make([]attr.Value, l)
	tl := make([]attr.Type, l)

	for i, v := range s {
		p := thisPath.AtListIndex(i)
		vv, err := decodeScalar(ctx, v, p, unknownPaths)
		if err != nil {
			return nil, err
		}
		vl[i] = vv
		tl[i] = vv.Type(ctx)
	}

	return types.TupleValue(tl, vl)
}

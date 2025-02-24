package tfutils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	high "github.com/pb33f/libopenapi/datamodel/high/base"
)

// DecodeDynamic decodes an object into a dynamic value.
func DecodeDynamic(ctx context.Context, ignore path.Expressions, obj any) (types.Dynamic, error) {
	if obj == nil {
		return types.DynamicNull(), nil
	}

	v, _, err := decodeScalar(ignore, path.Empty(), obj)
	if err != nil {
		return types.Dynamic{}, err
	}

	return dynamicFromTerraform(ctx, v)
}

// DecodeDynamicWithType decodes an object into a dynamic value with type information.
func DecodeDynamicWithType(ctx context.Context, s *high.Schema, ignore path.Expressions, obj any) (types.Dynamic, error) {
	t, err := GetTfTypeFromOpenAPI(s)
	if err != nil {
		return types.Dynamic{}, err
	}

	v, _, err := decodeScalarWithType(ignore, nil, false, path.Empty(), t, obj)
	if err != nil {
		return types.Dynamic{}, err
	}

	return dynamicFromTerraform(ctx, v)
}

// DecodeDynamicWithTypeAndUnknowns decodes an object into a dynamic value with type information and unknowns.
func DecodeDynamicWithTypeAndUnknowns(ctx context.Context, s *high.Schema, ignore path.Expressions, unknown path.Expressions, obj any) (types.Dynamic, error) {
	t, err := GetTfTypeFromOpenAPI(s)
	if err != nil {
		return types.Dynamic{}, err
	}

	v, _, err := decodeScalarWithType(ignore, unknown, true, path.Empty(), t, obj)
	if err != nil {
		return types.Dynamic{}, err
	}

	return dynamicFromTerraform(ctx, v)
}

// decodeScalarWithType decodes a scalar value into a tftypes value.
func decodeScalarWithType(ignore, unknown path.Expressions, withUnknown bool, p path.Path, t tftypes.Type, a any) (tftypes.Value, bool, error) {
	if ignore.Matches(p) {
		return tftypes.Value{}, false, nil
	}

	isUnknown := withUnknown && unknown.Matches(p)

	switch v := a.(type) {
	case nil:
		if isUnknown {
			return tftypes.NewValue(t, tftypes.UnknownValue), true, nil
		}

		return tftypes.NewValue(t, nil), true, nil
	case int64:
		if isUnknown {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), true, nil
		}

		return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(v))), true, nil
	case float64:
		if isUnknown {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), true, nil
		}

		return tftypes.NewValue(tftypes.Number, big.NewFloat(v)), true, nil
	case bool:
		if isUnknown {
			return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), true, nil
		}

		return tftypes.NewValue(tftypes.Bool, v), true, nil
	case string:
		if isUnknown {
			return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), true, nil
		}

		return tftypes.NewValue(tftypes.String, v), true, nil
	case []any:
		tt, ok := t.(tftypes.Tuple)
		if ok {
			return decodeTupleWithType(ignore, unknown, withUnknown, p, tt, v)
		}

		return tftypes.Value{}, false, fmt.Errorf("unexpected type %T", t)
	case map[string]any:
		to, ok := t.(tftypes.Object)
		if ok {
			return decodeObjectWithType(ignore, unknown, withUnknown, p, to, v)
		}

		tm, ok := t.(tftypes.Map)
		if ok {
			return decodeMapWithType(ignore, unknown, withUnknown, p, tm, v)
		}

		if t.Is(tftypes.DynamicPseudoType) {
			return decodeDynamicObjectWithType(ignore, unknown, withUnknown, p, v)
		}

		return tftypes.Value{}, false, fmt.Errorf("unexpected type %T", t)
	default:
		return tftypes.Value{}, false, fmt.Errorf("unexpected type %T", v)
	}
}

// decodeTupleWithType decodes an array value into a tftypes tuple value.
func decodeTupleWithType(ignore, unknown path.Expressions, withUnknown bool, p path.Path, t tftypes.Tuple, s []any) (tftypes.Value, bool, error) {
	l := len(s)
	vl := make([]tftypes.Value, 0, l)
	tl := make([]tftypes.Type, 0, l)

	for i, v := range s {
		nv, ok, err := decodeScalarWithType(ignore, unknown, withUnknown, p.AtTupleIndex(i), t.ElementTypes[0], v)
		if err != nil {
			return tftypes.Value{}, false, err
		}

		if ok {
			vl = append(vl, nv)
			tl = append(tl, nv.Type())
		}
	}

	return tftypes.NewValue(tftypes.Tuple{ElementTypes: tl}, vl), true, nil
}

// decodeMapWithType decodes an object value into a tftypes map value.
func decodeMapWithType(ignore, unknown path.Expressions, withUnknown bool, p path.Path, t tftypes.Map, m map[string]any) (tftypes.Value, bool, error) {
	vm := make(map[string]tftypes.Value, len(m))

	for k, v := range m {
		nv, ok, err := decodeScalarWithType(ignore, unknown, withUnknown, p.AtMapKey(k), t.ElementType, v)
		if err != nil {
			return tftypes.Value{}, false, err
		}

		if ok {
			vm[k] = nv
		}
	}

	return tftypes.NewValue(t, vm), true, nil
}

// decodeDynamicObjectWithType decodes an object value into a dynamic tftypes map value.
func decodeDynamicObjectWithType(ignore, unknown path.Expressions, withUnknown bool, p path.Path, m map[string]any) (tftypes.Value, bool, error) {
	l := len(m)
	vm := make(map[string]tftypes.Value, l)
	tm := make(map[string]tftypes.Type, l)

	for k, v := range m {
		nv, ok, err := decodeScalarWithType(ignore, unknown, withUnknown, p.AtMapKey(k), tftypes.DynamicPseudoType, v)
		if err != nil {
			return tftypes.Value{}, false, err
		}

		if ok {
			vm[k] = nv
			tm[k] = nv.Type()
		}
	}

	return tftypes.NewValue(tftypes.Object{AttributeTypes: tm}, vm), true, nil
}

// decodeObjectWithType decodes an object value into a tftypes object value.
func decodeObjectWithType(ignore, unknown path.Expressions, withUnknown bool, p path.Path, t tftypes.Object, m map[string]any) (tftypes.Value, bool, error) {
	if withUnknown {
		l := len(t.AttributeTypes)
		vm := make(map[string]tftypes.Value, l)
		tm := make(map[string]tftypes.Type, l)

		for k, tv := range t.AttributeTypes {
			v, ok := m[k]
			if !ok {
				v = nil
			}

			nv, ok, err := decodeScalarWithType(ignore, unknown, withUnknown, p.AtName(k), tv, v)
			if err != nil {
				return tftypes.Value{}, false, err
			}

			if ok {
				vm[k] = nv
				tm[k] = nv.Type()
			}
		}

		return tftypes.NewValue(tftypes.Object{AttributeTypes: tm}, vm), true, nil
	}

	l := len(m)
	vm := make(map[string]tftypes.Value, l)
	tm := make(map[string]tftypes.Type, l)

	for k, v := range m {
		nv, ok, err := decodeScalarWithType(ignore, unknown, withUnknown, p.AtName(k), t.AttributeTypes[k], v)
		if err != nil {
			return tftypes.Value{}, false, err
		}

		if ok {
			vm[k] = nv
			tm[k] = nv.Type()
		}
	}

	return tftypes.NewValue(tftypes.Object{AttributeTypes: tm}, vm), true, nil
}

// decodeScalar decodes a scalar value into an tftypes value.
func decodeScalar(ignore path.Expressions, p path.Path, a any) (tftypes.Value, bool, error) {
	if ignore.Matches(p) {
		return tftypes.Value{}, false, nil
	}

	switch v := a.(type) {
	case nil:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil), true, nil
	case int64:
		return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(v))), true, nil
	case float64:
		return tftypes.NewValue(tftypes.Number, big.NewFloat(v)), true, nil
	case bool:
		return tftypes.NewValue(tftypes.Bool, v), true, nil
	case string:
		return tftypes.NewValue(tftypes.String, v), true, nil
	case []any:
		return decodeTuple(ignore, p, v)
	case map[string]any:
		return decodeObject(ignore, p, v)
	default:
		return tftypes.Value{}, false, fmt.Errorf("unexpected type %T", v)
	}
}

// decodeTuple decodes an array into a tuple tftypes value.
func decodeTuple(ignore path.Expressions, p path.Path, s []any) (tftypes.Value, bool, error) {
	l := len(s)
	vl := make([]tftypes.Value, 0, l)
	tl := make([]tftypes.Type, 0, l)

	for i, v := range s {
		nv, ok, err := decodeScalar(ignore, p.AtTupleIndex(i), v)
		if err != nil {
			return tftypes.Value{}, false, err
		}

		if ok {
			vl = append(vl, nv)
			tl = append(tl, nv.Type())
		}
	}

	return tftypes.NewValue(tftypes.Tuple{ElementTypes: tl}, vl), true, nil
}

// decodeObject decodes a map into an object tftypes value.
func decodeObject(ignore path.Expressions, p path.Path, m map[string]any) (tftypes.Value, bool, error) {
	l := len(m)
	vm := make(map[string]tftypes.Value, l)
	tm := make(map[string]tftypes.Type, l)

	for k, v := range m {
		nv, ok, err := decodeScalar(ignore, p.AtName(k), v)
		if err != nil {
			return tftypes.Value{}, false, err
		}

		if ok {
			vm[k] = nv
			tm[k] = nv.Type()
		}
	}

	return tftypes.NewValue(tftypes.Object{AttributeTypes: tm}, vm), true, nil
}

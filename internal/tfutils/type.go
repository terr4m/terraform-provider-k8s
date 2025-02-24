package tfutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// dynamicFromTerraform converts a tftypes value to a dynamic value.
func dynamicFromTerraform(ctx context.Context, v tftypes.Value) (types.Dynamic, error) {
	dv, err := types.DynamicType.ValueFromTerraform(ctx, v)
	if err != nil {
		return types.Dynamic{}, err
	}

	d, ok := dv.(types.Dynamic)
	if !ok {
		return types.Dynamic{}, fmt.Errorf("could not cast value as dynamic")
	}

	return d, nil
}

// setValueType sets the type of an existing tftypes value.
func setValueType(unknown path.Paths, p path.Path, t tftypes.Type, v tftypes.Value) (tftypes.Value, error) {
	if !v.IsKnown() || v.IsNull() || unknown.Contains(p) {
		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	}

	switch tt := t.(type) {
	case tftypes.Tuple:
		return setTupleType(unknown, p, tt, v)
	case tftypes.Map:
		return setMapType(unknown, p, tt, v)
	case tftypes.Object:
		return setObjectType(unknown, p, tt, v)
	default:
		if v.Type().Equal(t) || t.Is(tftypes.DynamicPseudoType) {
			return v, nil
		}

		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	}
}

// setTupleType sets the type of an existing tftypes tuple value.
func setTupleType(unknown path.Paths, p path.Path, t tftypes.Tuple, v tftypes.Value) (tftypes.Value, error) {
	var sv []tftypes.Value
	err := v.As(&sv)
	if err != nil {
		return tftypes.Value{}, err
	}

	l := len(sv)
	ets := make([]tftypes.Type, l)
	evs := make([]tftypes.Value, l)

	for i, e := range sv {
		ev, err := setValueType(unknown, p.AtTupleIndex(i), t.ElementTypes[0], e)
		if err != nil {
			return tftypes.Value{}, err
		}

		evs[i] = ev
		ets[i] = ev.Type()
	}

	return tftypes.NewValue(tftypes.Tuple{ElementTypes: ets}, evs), nil
}

// setMapType sets the type of an existing tftypes map value.
func setMapType(unknown path.Paths, p path.Path, t tftypes.Map, v tftypes.Value) (tftypes.Value, error) {
	var sv map[string]tftypes.Value
	err := v.As(&sv)
	if err != nil {
		return tftypes.Value{}, err
	}

	evs := make(map[string]tftypes.Value, len(sv))

	for k, v := range sv {
		ev, err := setValueType(unknown, p.AtMapKey(k), t.ElementType, v)
		if err != nil {
			return tftypes.Value{}, err
		}

		evs[k] = ev
	}

	return tftypes.NewValue(t, evs), nil
}

// setObjectType sets the type of an existing tftypes object value.
func setObjectType(unknown path.Paths, p path.Path, t tftypes.Object, v tftypes.Value) (tftypes.Value, error) {
	var sv map[string]tftypes.Value
	err := v.As(&sv)
	if err != nil {
		return tftypes.Value{}, err
	}

	l := len(sv)
	ets := make(map[string]tftypes.Type, l)
	evs := make(map[string]tftypes.Value, l)

	for k, v := range sv {
		ev, err := setValueType(unknown, p.AtName(k), t.AttributeTypes[k], v)
		if err != nil {
			return tftypes.Value{}, err
		}

		evs[k] = ev
		ets[k] = ev.Type()
	}

	return tftypes.NewValue(tftypes.Object{AttributeTypes: ets}, evs), nil
}

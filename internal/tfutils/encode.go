package tfutils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EncodeDynamicObject encodes a Terraform dynamic object into a Go object map.
func EncodeDynamicObject(d types.Dynamic) (map[string]any, error) {
	if d.IsNull() {
		return nil, fmt.Errorf("expected object value, got null")
	}

	if d.IsUnknown() {
		return nil, fmt.Errorf("expected known object value, got unknown")
	}

	o, ok := d.UnderlyingValue().(types.Object)
	if !ok {
		return nil, fmt.Errorf("expected object value, got %T", d.UnderlyingValue())
	}

	return encodeMapping(o.Attributes())
}

func encodeScalar(v attr.Value) (any, error) {
	if v == nil {
		return nil, nil
	}

	if v.IsUnknown() {
		return nil, fmt.Errorf("expected known value, got unknown %T", v)
	}

	if v.IsNull() {
		return nil, nil
	}

	switch val := v.(type) {
	case types.Dynamic:
		return encodeScalar(val.UnderlyingValue())
	case types.Bool:
		return val.ValueBool(), nil
	case types.String:
		return val.ValueString(), nil
	case types.Int64:
		return val.ValueInt64(), nil
	case types.Float64:
		return val.ValueFloat64(), nil
	case types.Number:
		f, _ := val.ValueBigFloat().Float64()
		return f, nil
	case types.List:
		return encodeSequence(val.Elements())
	case types.Set:
		return encodeSequence(val.Elements())
	case types.Tuple:
		return encodeSequence(val.Elements())
	case types.Map:
		return encodeMapping(val.Elements())
	case types.Object:
		return encodeMapping(val.Attributes())
	default:
		return nil, fmt.Errorf("unexpected type: %T", val)
	}
}

func encodeMapping(m map[string]attr.Value) (map[string]any, error) {
	result := make(map[string]any, len(m))
	for k, v := range m {
		a, err := encodeScalar(v)
		if err != nil {
			return nil, err
		}

		result[k] = a
	}

	return result, nil
}

func encodeSequence(s []attr.Value) ([]any, error) {
	result := make([]any, len(s))
	for i, v := range s {
		a, err := encodeScalar(v)
		if err != nil {
			return nil, err
		}

		result[i] = a
	}

	return result, nil
}

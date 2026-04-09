package tfutils

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEncodeDynamicObject(t *testing.T) {
	t.Parallel()

	simpleObject, _ := types.ObjectValue(
		map[string]attr.Type{"foo": types.StringType, "bar": types.StringType},
		map[string]attr.Value{"foo": types.StringValue("baz"), "bar": types.StringNull()},
	)

	nestedObject, _ := types.ObjectValue(
		map[string]attr.Type{"metadata": simpleObject.Type(t.Context())},
		map[string]attr.Value{"metadata": simpleObject},
	)

	for _, d := range []struct {
		testName string
		input    types.Dynamic
		want     map[string]any
		errMsg   string
	}{
		{
			testName: "unknown_dynamic",
			input:    types.DynamicUnknown(),
			errMsg:   "expected known object value, got unknown",
		},
		{
			testName: "invalid_dynamic",
			input:    types.DynamicValue(types.StringValue("foo")),
			errMsg:   "expected object value, got basetypes.StringValue",
		},
		{
			testName: "object_with_nulls",
			input:    types.DynamicValue(simpleObject),
			want:     map[string]any{"foo": "baz", "bar": nil},
		},
		{
			testName: "nested_object",
			input:    types.DynamicValue(nestedObject),
			want: map[string]any{
				"metadata": map[string]any{"foo": "baz", "bar": nil},
			},
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			got, err := EncodeDynamicObject(d.input)

			if d.errMsg == "" && err != nil {
				t.Fatalf("EncodeDynamicObject returned unexpected error: %v", err)
			}

			if d.errMsg != "" {
				if err == nil || err.Error() != d.errMsg {
					t.Fatalf("EncodeDynamicObject returned error %v, want %q", err, d.errMsg)
				}
				return
			}

			if !reflect.DeepEqual(d.want, got) {
				t.Fatalf("EncodeDynamicObject returned %#v, want %#v", got, d.want)
			}
		})
	}
}

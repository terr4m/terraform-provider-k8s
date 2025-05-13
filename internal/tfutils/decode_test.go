package tfutils

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDecodeDynamic(t *testing.T) {
	t.Parallel()

	simpleObject, _ := types.ObjectValue(map[string]attr.Type{"foo": types.StringType}, map[string]attr.Value{"foo": types.StringValue("bar")})
	stringTuple, _ := types.TupleValue([]attr.Type{types.StringType, types.StringType}, []attr.Value{types.StringValue("foo"), types.StringValue("bar")})

	for _, d := range []struct {
		testName string
		in       any
		want     types.Dynamic
		errMsg   string
	}{
		{
			testName: "unexpected_type",
			in:       1,
			want:     types.Dynamic{},
			errMsg:   "Unexpected type.",
		},
		{
			testName: "null",
			in:       nil,
			want:     types.DynamicNull(),
		},
		{
			testName: "int64",
			in:       int64(1),
			want:     types.DynamicValue(types.NumberValue(big.NewFloat(float64(1)))),
		},
		{
			testName: "float64",
			in:       float64(1.1),
			want:     types.DynamicValue(types.NumberValue(big.NewFloat(1.1))),
		},
		{
			testName: "bool",
			in:       true,
			want:     types.DynamicValue(types.BoolValue(true)),
		},
		{
			testName: "string",
			in:       "foo",
			want:     types.DynamicValue(types.StringValue("foo")),
		},
		{
			testName: "object_empty",
			in:       map[string]any{},
			want:     types.DynamicNull(),
		},
		{
			testName: "object_simple",
			in:       map[string]any{"foo": "bar"},
			want:     types.DynamicValue(simpleObject),
		},
		{
			testName: "array_empty",
			in:       []any{},
			want:     types.DynamicNull(),
		},
		{
			testName: "array_strings",
			in:       []any{"foo", "bar"},
			want:     types.DynamicValue(stringTuple),
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			got, diags := DecodeDynamic(ctx, d.in)

			if !got.Equal(d.want) {
				t.Errorf("DecodeDynamic returned:\n%v\nwant:\n%v", got, d.want)
			}

			var errMsg string
			if diags.HasError() {
				for i, diag := range diags.Errors() {
					if i == 0 {
						errMsg = diag.Summary()
						continue
					}
					errMsg = fmt.Sprintf("%s: %s", errMsg, diag.Summary())
				}
			}

			if errMsg != d.errMsg {
				t.Errorf("DecodeDynamic returned error message %q, want %q", errMsg, d.errMsg)
			}
		})
	}
}

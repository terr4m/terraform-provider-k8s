package k8sutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	for _, d := range []struct {
		testName string
		in       any
	}{
		{
			testName: "nil",
			in:       nil,
		},
		{
			testName: "bool",
			in:       true,
		},
		{
			testName: "int",
			in:       1,
		},
		{
			testName: "string",
			in:       "foo",
		},
		{
			testName: "map",
			in:       map[string]any{"foo": "bar"},
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			want := &d.in

			got := Ptr(d.in)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf(
					"Ptr returned:\n%v\nwant:\n%v\ndiff:\n%v",
					got,
					want,
					diff,
				)
			}
		})
	}
}

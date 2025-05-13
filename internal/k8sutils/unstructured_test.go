package k8sutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestUnstructuredListToObjects(t *testing.T) {
	t.Parallel()

	for _, d := range []struct {
		testName string
		in       *unstructured.UnstructuredList
		want     []any
	}{
		{
			testName: "nil",
			in:       nil,
			want:     []any{},
		},
		{
			testName: "empty",
			in:       &unstructured.UnstructuredList{},
			want:     []any{},
		},
		{
			testName: "single",
			in:       &unstructured.UnstructuredList{Items: []unstructured.Unstructured{{Object: map[string]any{"foo": "bar"}}}},
			want:     []any{map[string]any{"foo": "bar"}},
		},
		{
			testName: "multiple",
			in:       &unstructured.UnstructuredList{Items: []unstructured.Unstructured{{Object: map[string]any{"foo": "bar"}}, {Object: map[string]any{"foo": "baz"}}}},
			want:     []any{map[string]any{"foo": "bar"}, map[string]any{"foo": "baz"}},
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			got := UnstructuredListToObjects(d.in)

			if diff := cmp.Diff(d.want, got); diff != "" {
				t.Errorf(
					"UnstructuredListToObjects returned:\n%v\nwant:\n%v\ndiff:\n%v",
					got,
					d.want,
					diff,
				)
			}
		})
	}
}

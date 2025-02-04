package k8sutils

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UnstructuredListToObjects converts an unstructured list to a list of objects.
func UnstructuredListToObjects(ul *unstructured.UnstructuredList) []any {
	s := make([]any, len(ul.Items))
	for i, v := range ul.Items {
		s[i] = v.Object
	}

	return s
}

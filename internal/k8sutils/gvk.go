package k8sutils

import "k8s.io/apimachinery/pkg/runtime/schema"

// ParseGVK parses the input API version and kind into a GroupVersionKind.
func ParseGVK(apiVersion, kind string) (*schema.GroupVersionKind, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}

	gvk := gv.WithKind(kind)

	return &gvk, nil
}

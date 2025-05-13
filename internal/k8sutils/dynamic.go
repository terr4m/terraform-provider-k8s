package k8sutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetMapping returns a REST mapping for the given REST mapper and GroupVersionKind.
func GetMapping(m meta.RESTMapper, gvk *schema.GroupVersionKind) (*meta.RESTMapping, error) {
	mapping, err := m.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %w", err)
	}
	return mapping, nil
}

// GetResourceInterface returns a dynamic resource interface for the given REST mapping and namespace.
func GetResourceInterface(c dynamic.Interface, mapping *meta.RESTMapping, requireNamespace bool, namespace string) (dynamic.ResourceInterface, error) {
	res := c.Resource(mapping.Resource)

	if mapping.Scope.Name() != meta.RESTScopeNameNamespace || (!requireNamespace && len(namespace) == 0) {
		return res, nil
	}

	if len(namespace) > 0 {
		return res.Namespace(namespace), nil
	}

	return nil, fmt.Errorf("namespace is required for namespaced resources")
}

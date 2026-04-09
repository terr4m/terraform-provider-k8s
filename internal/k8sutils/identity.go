package k8sutils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ManifestIdentity(manifest map[string]any) (*schema.GroupVersionKind, string, string, error) {
	u := &unstructured.Unstructured{Object: manifest}
	if u.GetAPIVersion() == "" {
		return nil, "", "", fmt.Errorf("manifest.apiVersion is required")
	}
	if u.GetKind() == "" {
		return nil, "", "", fmt.Errorf("manifest.kind is required")
	}
	if u.GetName() == "" {
		return nil, "", "", fmt.Errorf("manifest.metadata.name is required")
	}

	gvk, err := ParseGVK(u.GetAPIVersion(), u.GetKind())
	if err != nil {
		return nil, "", "", err
	}

	return gvk, u.GetNamespace(), u.GetName(), nil
}

func BuildApplyOptions(defaults FieldManager, override *FieldManager) metav1.ApplyOptions {
	options := metav1.ApplyOptions{FieldManager: defaults.Name, Force: defaults.ForceConflicts}
	if override == nil {
		return options
	}

	if override.Name != "" {
		options.FieldManager = override.Name
	}
	options.Force = override.ForceConflicts

	return options
}

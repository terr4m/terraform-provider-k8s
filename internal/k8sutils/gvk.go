package k8sutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

// ParseGVK parses the input API version and kind into a GroupVersionKind.
func ParseGVK(apiVersion, kind string) (*schema.GroupVersionKind, error) {
	if len(apiVersion) == 0 {
		return nil, fmt.Errorf("no API version provided")
	}

	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}

	gvk := gv.WithKind(kind)

	return &gvk, nil
}

// CheckGVKExists checks if the given API version and kind exists in the cluster.
func CheckGVKExists(dc discovery.DiscoveryInterface, apiVersion, kind string) (bool, error) {
	rl, err := dc.ServerResourcesForGroupVersion(apiVersion)
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, v := range rl.APIResources {
		if v.Kind == kind {
			return true, nil
		}
	}

	return false, nil
}

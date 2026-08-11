package k8sutils

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type discoveryClientResult struct {
	resources *metav1.APIResourceList
	err       error
}

type discoveryClientStub struct {
	discovery.DiscoveryInterface

	results map[string]discoveryClientResult
}

func (c *discoveryClientStub) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	res, ok := c.results[groupVersion]
	if !ok {
		return nil, &errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
	}

	return res.resources, res.err
}

type dynamicClientStub struct {
	dynamic.Interface

	results map[string]dynamic.NamespaceableResourceInterface
}

func (c *dynamicClientStub) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	k := strings.ToLower(fmt.Sprintf("%s/%s/%s", resource.Group, resource.Version, resource.Resource))
	return c.results[k]
}

type resourceInterfaceStub struct {
	dynamic.NamespaceableResourceInterface
}

func (r *resourceInterfaceStub) Namespace(_ string) dynamic.ResourceInterface {
	return r
}

type restMapperResult struct {
	mapping *meta.RESTMapping
	err     error
}

type restMapperStub struct {
	meta.ResettableRESTMapper

	results map[string]restMapperResult
}

func (m *restMapperStub) RESTMapping(gk schema.GroupKind, _ ...string) (*meta.RESTMapping, error) {
	k := strings.ToLower(gk.String())
	res, ok := m.results[k]
	if !ok {
		return nil, nil
	}
	return res.mapping, res.err
}

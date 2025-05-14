package k8sutils

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type discoveryClientStub struct {
	discovery.DiscoveryInterface

	serverResourcesForGroupVersionResult struct {
		resources *metav1.APIResourceList
		err       error
	}
}

func (c *discoveryClientStub) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return c.serverResourcesForGroupVersionResult.resources, c.serverResourcesForGroupVersionResult.err
}

type dynamicClientStub struct {
	dynamic.Interface

	resourceResult dynamic.NamespaceableResourceInterface
}

func (c *dynamicClientStub) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return c.resourceResult
}

type resourceInterfaceStub struct {
	dynamic.NamespaceableResourceInterface

	namespaceResult dynamic.ResourceInterface
}

func (r resourceInterfaceStub) Namespace(namespace string) dynamic.ResourceInterface {
	return r.namespaceResult
}

type restMapperStub struct {
	meta.ResettableRESTMapper

	restMappingResult struct {
		mapping *meta.RESTMapping
		err     error
	}
}

func (m *restMapperStub) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	return m.restMappingResult.mapping, m.restMappingResult.err
}

package k8sutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/openapi"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	aggregator "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

// NewClient creates a new K8s client.
func NewClient(config *rest.Config) *Client {
	return &Client{
		config:              config,
		aggregatorClientset: nil,
	}
}

// Client is a generic K8s client.
type Client struct {
	config              *rest.Config
	aggregatorClientset *aggregator.Clientset
	discoveryClient     discovery.DiscoveryInterface
	openapiClient       openapi.Client
	dynamicClient       dynamic.Interface
	restMapper          meta.RESTMapper
}

// AggregatorClientset returns an aggregator K8s clientset.
func (c *Client) AggregatorClientset() (*aggregator.Clientset, error) {
	if c.aggregatorClientset != nil {
		return c.aggregatorClientset, nil
	}
	if c.config != nil {
		ac, err := aggregator.NewForConfig(c.config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure aggregator client: %w", err)
		}
		c.aggregatorClientset = ac
	}
	return c.aggregatorClientset, nil
}

// DiscoveryClient returns a discovery K8s client.
func (c *Client) DiscoveryClient() (discovery.DiscoveryInterface, error) {
	if c.discoveryClient != nil {
		return c.discoveryClient, nil
	}

	if c.config != nil {
		kc, err := discovery.NewDiscoveryClientForConfig(c.config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure discovery client: %w", err)
		}
		c.discoveryClient = kc
	}
	return c.discoveryClient, nil
}

// OpenAPIClient returns an OpenAPI K8s client.
func (c *Client) OpenAPIClient() (openapi.Client, error) {
	if c.openapiClient != nil {
		return c.openapiClient, nil
	}

	dc, err := c.DiscoveryClient()
	if err != nil {
		return nil, err
	}

	c.openapiClient = dc.OpenAPIV3()
	return c.openapiClient, nil
}

// DynamicClient returns a dynamic K8s client.
func (c *Client) DynamicClient() (dynamic.Interface, error) {
	if c.dynamicClient != nil {
		return c.dynamicClient, nil
	}

	if c.config != nil {
		kc, err := dynamic.NewForConfig(c.config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure dynamic client: %w", err)
		}
		c.dynamicClient = kc
	}
	return c.dynamicClient, nil
}

// RESTMapper returns a REST mapper.
func (c *Client) RESTMapper() (meta.RESTMapper, error) {
	if c.restMapper != nil {
		return c.restMapper, nil
	}

	dc, err := c.DiscoveryClient()
	if err != nil {
		return nil, err
	}

	cache := memory.NewMemCacheClient(dc)
	c.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(cache)
	return c.restMapper, nil
}

// CheckGVK checks if the group version kind is supported.
func (c *Client) CheckGVK(apiVersion, kind string) (bool, error) {
	d, err := c.DiscoveryClient()
	if err != nil {
		return false, err
	}

	rl, err := d.ServerResourcesForGroupVersion(apiVersion)
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

// ResourceInterface returns a resource interface for a given api version and kind.
func (c *Client) ResourceInterface(gvk *schema.GroupVersionKind, namespace string, defaultNamespace bool) (dynamic.ResourceInterface, error) {
	mapper, err := c.RESTMapper()
	if err != nil {
		return nil, err
	}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	dyn, err := c.DynamicClient()
	if err != nil {
		return nil, err
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if len(namespace) == 0 && defaultNamespace {
			namespace = "default"
		}

		if len(namespace) > 0 {
			return dyn.Resource(mapping.Resource).Namespace(namespace), nil
		}
	}

	return dyn.Resource(mapping.Resource), nil
}

// GetGVOpenAPISchemaLookup returns the schema lookup for the given GV key.
func (c *Client) GetGVOpenAPISchemaLookup(gvk *schema.GroupVersionKind) (openapi.GroupVersion, error) {
	oc, err := c.OpenAPIClient()
	if err != nil {
		return nil, err
	}

	paths, err := oc.Paths()
	if err != nil {
		return nil, err
	}

	key := "api"
	if len(gvk.Group) > 0 {
		key = fmt.Sprintf("%ss/%s", key, gvk.Group)
	}

	if len(gvk.Version) > 0 {
		key = fmt.Sprintf("%s/%s", key, gvk.Version)
	}

	return paths[key], nil
}

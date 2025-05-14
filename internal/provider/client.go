package provider

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	aggregator "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

// NewK8sProviderClient creates a new K8s provider client.
func NewK8sProviderClient(restConfig *rest.Config) *K8sProviderClient {
	return &K8sProviderClient{
		restConfig: restConfig,
	}
}

// K8sProviderClient is a K8s provider client.
type K8sProviderClient struct {
	restConfig       *rest.Config
	aggregatorClient aggregator.Interface
	discoveryClient  discovery.CachedDiscoveryInterface
	dynamicClient    dynamic.Interface
	restMapper       meta.ResettableRESTMapper
}

// AggregatorClientset returns an aggregator K8s clientset.
func (c *K8sProviderClient) AggregatorClient() (aggregator.Interface, error) {
	if c.restConfig == nil {
		return nil, fmt.Errorf("rest config is required")
	}

	if c.aggregatorClient != nil {
		return c.aggregatorClient, nil
	}

	ac, err := aggregator.NewForConfig(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure aggregator client: %w", err)
	}
	c.aggregatorClient = ac

	return c.aggregatorClient, nil
}

// DiscoveryClient returns a discovery K8s client.
func (c *K8sProviderClient) DiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	if c.restConfig == nil {
		return nil, fmt.Errorf("rest config is required")
	}

	if c.discoveryClient != nil {
		return c.discoveryClient, nil
	}

	dc, err := discovery.NewDiscoveryClientForConfig(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure discovery client: %w", err)
	}

	c.discoveryClient = memory.NewMemCacheClient(dc)

	return c.discoveryClient, nil
}

// DynamicClient returns a dynamic K8s client.
func (c *K8sProviderClient) DynamicClient() (dynamic.Interface, error) {
	if c.restConfig == nil {
		return nil, fmt.Errorf("rest config is required")
	}

	if c.dynamicClient != nil {
		return c.dynamicClient, nil
	}

	kc, err := dynamic.NewForConfig(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure dynamic client: %w", err)
	}
	c.dynamicClient = kc

	return c.dynamicClient, nil
}

// RESTMapper returns a REST mapper.
func (c *K8sProviderClient) RESTMapper() (meta.ResettableRESTMapper, error) {
	if c.restConfig == nil {
		return nil, fmt.Errorf("rest config is required")
	}

	if c.restMapper != nil {
		return c.restMapper, nil
	}

	dc, err := c.DiscoveryClient()
	if err != nil {
		return nil, err
	}

	c.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(dc)

	return c.restMapper, nil
}

package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type discoveryClientStub struct {
	discovery.CachedDiscoveryInterface
}

type dynamicClientStub struct {
	dynamic.Interface
}

type restMapperStub struct {
	meta.ResettableRESTMapper
}

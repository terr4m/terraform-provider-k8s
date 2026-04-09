package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"k8s": providerserver.NewProtocol6WithError(New("test", "test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
}

type discoveryClientStub struct {
	discovery.CachedDiscoveryInterface
}

type dynamicClientStub struct {
	dynamic.Interface
}

type restMapperStub struct {
	meta.ResettableRESTMapper
}

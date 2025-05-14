package k8sutils

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func TestGetMapping(t *testing.T) {
	t.Parallel()

	for _, d := range []struct {
		testName  string
		mockSetup func() *restMapperStub
		gvk       *schema.GroupVersionKind
		errMsg    string
	}{
		{
			testName: "success",
			mockSetup: func() *restMapperStub {
				return &restMapperStub{
					restMappingResult: struct {
						mapping *meta.RESTMapping
						err     error
					}{
						mapping: &meta.RESTMapping{},
						err:     nil,
					},
				}
			},
			gvk: &schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		},
		{
			testName: "failure",
			mockSetup: func() *restMapperStub {
				return &restMapperStub{
					restMappingResult: struct {
						mapping *meta.RESTMapping
						err     error
					}{
						mapping: nil,
						err:     fmt.Errorf("server error"),
					},
				}
			},
			gvk:    &schema.GroupVersionKind{Group: "apps", Version: "v2", Kind: "Deployment"},
			errMsg: "failed to get REST mapping: server error",
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			m := d.mockSetup()
			got, err := GetMapping(m, d.gvk)

			if len(d.errMsg) == 0 && got == nil {
				t.Errorf("GetMapping returned nil, want non-nil")
			}

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != d.errMsg {
				t.Errorf("GetMapping returned error message %q, want %q", errMsg, d.errMsg)
			}
		})
	}
}

func TestGetResourceInterface(t *testing.T) {
	t.Parallel()

	for _, d := range []struct {
		testName         string
		mockSetup        func() dynamic.Interface
		mapping          *meta.RESTMapping
		requireNamespace bool
		namespace        string
		errMsg           string
	}{
		{
			testName: "non_namespaced",
			mockSetup: func() dynamic.Interface {
				return &dynamicClientStub{
					resourceResult: resourceInterfaceStub{},
				}
			},
			mapping: &meta.RESTMapping{
				Scope: meta.RESTScopeRoot,
			},
			requireNamespace: true,
			namespace:        "",
		},
		{
			testName: "namespaced",
			mockSetup: func() dynamic.Interface {
				return &dynamicClientStub{
					resourceResult: resourceInterfaceStub{
						namespaceResult: resourceInterfaceStub{},
					},
				}
			},
			mapping: &meta.RESTMapping{
				Scope: meta.RESTScopeNamespace,
			},
			requireNamespace: true,
			namespace:        "test",
		},
		{
			testName: "namespaced_no_namespace_allowed",
			mockSetup: func() dynamic.Interface {
				return &dynamicClientStub{
					resourceResult: resourceInterfaceStub{
						namespaceResult: resourceInterfaceStub{},
					},
				}
			},
			mapping: &meta.RESTMapping{
				Scope: meta.RESTScopeNamespace,
			},
			requireNamespace: false,
			namespace:        "",
		},
		{
			testName: "namespaced_namespace_missing",
			mockSetup: func() dynamic.Interface {
				return &dynamicClientStub{
					resourceResult: resourceInterfaceStub{
						namespaceResult: resourceInterfaceStub{},
					},
				}
			},
			mapping: &meta.RESTMapping{
				Scope: meta.RESTScopeNamespace,
			},
			requireNamespace: true,
			namespace:        "",
			errMsg:           "namespace is required for namespaced resources",
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			c := d.mockSetup()
			got, err := GetResourceInterface(c, d.mapping, d.requireNamespace, d.namespace)

			if len(d.errMsg) == 0 && got == nil {
				t.Errorf("GetResourceInterface returned nil, want non-nil")
			}

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != d.errMsg {
				t.Errorf("GetResourceInterface returned error message %q, want %q", errMsg, d.errMsg)
			}
		})
	}
}

package k8sutils

import (
	"fmt"
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func TestGetMapping(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		testName string
		mapper   meta.ResettableRESTMapper
		gvk      *schema.GroupVersionKind
		wantErr  *string
	}{
		{
			testName: "success",
			mapper: &restMapperStub{
				results: map[string]restMapperResult{
					"deployment.apps": {
						mapping: &meta.RESTMapping{},
					},
				},
			},
			gvk: &schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		},
		{
			testName: "failure",
			mapper: &restMapperStub{
				results: map[string]restMapperResult{
					"deployment.apps": {
						err: fmt.Errorf("server error"),
					},
				},
			},
			gvk:     &schema.GroupVersionKind{Group: "apps", Version: "v2", Kind: "Deployment"},
			wantErr: new("failed to get rest mapping: server error"),
		},
	} {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			got, err := GetMapping(tt.mapper, tt.gvk)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("GetMapping() returned unexpected error: %v", err)
				}

				if !regexp.MustCompile(regexp.QuoteMeta(*tt.wantErr)).MatchString(err.Error()) {
					t.Errorf("GetMapping() returned error %q, want %q", err.Error(), *tt.wantErr)
				}

				return
			}

			if tt.wantErr != nil {
				t.Errorf("GetMapping() returned no error, want %q", *tt.wantErr)
			}

			if got == nil {
				t.Errorf("GetMapping() returned nil, want non-nil")
			}
		})
	}
}

func TestGetResourceInterface(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		testName         string
		client           dynamic.Interface
		mapping          *meta.RESTMapping
		requireNamespace bool
		namespace        string
		wantErr          *string
	}{
		{
			testName: "non_namespaced",
			client: &dynamicClientStub{
				results: map[string]dynamic.NamespaceableResourceInterface{
					"rbac.authorization.k8s.io/v1/clusterroles": &resourceInterfaceStub{},
				},
			},
			mapping: &meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
				Scope:    meta.RESTScopeRoot,
			},
			requireNamespace: false,
			namespace:        "",
		},
		{
			testName: "namespaced",
			client: &dynamicClientStub{
				results: map[string]dynamic.NamespaceableResourceInterface{
					"apps/v1/deployments": &resourceInterfaceStub{},
				},
			},
			mapping: &meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
				Scope:    meta.RESTScopeNamespace,
			},
			requireNamespace: true,
			namespace:        "test",
		},
		{
			testName: "namespaced_no_namespace_allowed",
			client: &dynamicClientStub{
				results: map[string]dynamic.NamespaceableResourceInterface{
					"apps/v1/deployments": &resourceInterfaceStub{},
				},
			},
			mapping: &meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
				Scope:    meta.RESTScopeNamespace,
			},
			requireNamespace: false,
			namespace:        "",
		},
		{
			testName: "namespaced_namespace_missing",
			client: &dynamicClientStub{
				results: map[string]dynamic.NamespaceableResourceInterface{
					"apps/v1/deployments": &resourceInterfaceStub{},
				},
			},
			mapping: &meta.RESTMapping{
				Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
				Scope:    meta.RESTScopeNamespace,
			},
			requireNamespace: true,
			namespace:        "",
			wantErr:          new("namespace is required for namespaced resources"),
		},
	} {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			got, err := GetResourceInterface(tt.client, tt.mapping, tt.requireNamespace, tt.namespace)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("GetResourceInterface returned unexpected error: %v", err)
				}

				if !regexp.MustCompile(regexp.QuoteMeta(*tt.wantErr)).MatchString(err.Error()) {
					t.Errorf("GetResourceInterface returned error %q, want %q", err.Error(), *tt.wantErr)
				}

				return
			}

			if tt.wantErr != nil {
				t.Errorf("GetResourceInterface returned nil error, want %q", *tt.wantErr)
			}

			if got == nil {
				t.Errorf("GetResourceInterface returned nil, want non-nil")
			}
		})
	}
}

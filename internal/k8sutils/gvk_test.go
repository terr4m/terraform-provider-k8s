package k8sutils

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func TestParseGVK(t *testing.T) {
	t.Parallel()

	for _, d := range []struct {
		testName   string
		apiVersion string
		kind       string
		want       *schema.GroupVersionKind
		errMsg     string
	}{
		{
			testName:   "core_api_only",
			apiVersion: "v1",
			kind:       "",
			want:       &schema.GroupVersionKind{Group: "", Version: "v1"},
		},
		{
			testName:   "core",
			apiVersion: "v1",
			kind:       "ConfigMap",
			want:       &schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		},
		{
			testName:   "standard_api_only",
			apiVersion: "apps/v1",
			kind:       "",
			want:       &schema.GroupVersionKind{Group: "apps", Version: "v1"},
		},
		{
			testName:   "standard",
			apiVersion: "apps/v1",
			kind:       "Deployment",
			want:       &schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		},
		{
			testName:   "no_api_version_kind",
			apiVersion: "",
			kind:       "",
			want:       nil,
			errMsg:     "no API version provided",
		},
		{
			testName:   "no_api_version",
			apiVersion: "",
			kind:       "Deployment",
			want:       nil,
			errMsg:     "no API version provided",
		},
		{
			testName:   "invalid_api_version",
			apiVersion: "a/b/c",
			kind:       "",
			want:       nil,
			errMsg:     "unexpected GroupVersion string: a/b/c",
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			got, err := ParseGVK(d.apiVersion, d.kind)

			if diff := cmp.Diff(d.want, got); diff != "" {
				t.Errorf(
					"ParseGVK returned:\n%v\nwant:\n%v\ndiff:\n%v",
					got,
					d.want,
					diff,
				)
			}

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != d.errMsg {
				t.Errorf("ParseGVK returned error message %q, want %q", errMsg, d.errMsg)
			}
		})
	}
}

func TestCheckGVKExists(t *testing.T) {
	t.Parallel()

	for _, d := range []struct {
		testName   string
		apiVersion string
		kind       string
		mockSetup  func() discovery.DiscoveryInterface
		want       bool
		errMsg     string
	}{
		{
			testName:   "resource_exists",
			apiVersion: "apps/v1",
			kind:       "Deployment",
			mockSetup: func() discovery.DiscoveryInterface {
				return &discoveryClientStub{
					serverResourcesForGroupVersionResult: struct {
						resources *metav1.APIResourceList
						err       error
					}{
						resources: &metav1.APIResourceList{
							GroupVersion: "apps/v1",
							APIResources: []metav1.APIResource{
								{Kind: "Deployment"},
							},
						},
						err: nil,
					},
				}
			},
			want: true,
		},
		{
			testName:   "resource_not_found",
			apiVersion: "apps/v1",
			kind:       "Deploymentx",
			mockSetup: func() discovery.DiscoveryInterface {
				return &discoveryClientStub{
					serverResourcesForGroupVersionResult: struct {
						resources *metav1.APIResourceList
						err       error
					}{
						resources: &metav1.APIResourceList{
							GroupVersion: "apps/v1",
							APIResources: []metav1.APIResource{
								{Kind: "Deployment"},
							},
						},
						err: nil,
					},
				}
			},
			want: false,
		},
		{
			testName:   "group_version_not_found",
			apiVersion: "apps/v2",
			kind:       "Deployment",
			mockSetup: func() discovery.DiscoveryInterface {
				return &discoveryClientStub{
					serverResourcesForGroupVersionResult: struct {
						resources *metav1.APIResourceList
						err       error
					}{
						err: errors.NewNotFound(schema.GroupResource{Group: "apps/v2"}, ""),
					},
				}
			},
			want: false,
		},
		{
			testName:   "error",
			apiVersion: "apps/v1",
			kind:       "Deployment",
			mockSetup: func() discovery.DiscoveryInterface {
				return &discoveryClientStub{
					serverResourcesForGroupVersionResult: struct {
						resources *metav1.APIResourceList
						err       error
					}{
						err: fmt.Errorf("server error"),
					},
				}
			},
			want:   false,
			errMsg: "server error",
		},
	} {
		t.Run(d.testName, func(t *testing.T) {
			t.Parallel()

			dc := d.mockSetup()
			got, err := CheckGVKExists(dc, d.apiVersion, d.kind)

			if got != d.want {
				t.Errorf("CheckGVKExists returned %t, want %t", got, d.want)
			}

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != d.errMsg {
				t.Errorf("CheckGVKExists returned error message %q, want %q", errMsg, d.errMsg)
			}
		})
	}
}

package k8sutils

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func TestParseGVK(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		testName   string
		apiVersion string
		kind       string
		want       *schema.GroupVersionKind
		wantErr    *string
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
			wantErr:    new("no api version provided"),
		},
		{
			testName:   "no_api_version",
			apiVersion: "",
			kind:       "Deployment",
			want:       nil,
			wantErr:    new("no api version provided"),
		},
		{
			testName:   "invalid_api_version",
			apiVersion: "a/b/c",
			kind:       "",
			want:       nil,
			wantErr:    new("unexpected GroupVersion string: a/b/c"),
		},
	} {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			got, err := ParseGVK(tt.apiVersion, tt.kind)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("ParseGVK() returned unexpected error: %v", err)
				}

				if !regexp.MustCompile(regexp.QuoteMeta(*tt.wantErr)).MatchString(err.Error()) {
					t.Errorf("ParseGVK() returned error %q, want %q", err.Error(), *tt.wantErr)
				}

				return
			}

			if tt.wantErr != nil {
				t.Errorf("ParseGVK() returned no error, want %q", *tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseGVK() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckGVKExists(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		testName   string
		apiVersion string
		kind       string
		dc         discovery.DiscoveryInterface
		want       bool
		wantErr    *string
	}{
		{
			testName:   "resource_exists",
			apiVersion: "apps/v1",
			kind:       "Deployment",
			dc: &discoveryClientStub{
				results: map[string]discoveryClientResult{
					"apps/v1": {
						resources: &metav1.APIResourceList{
							GroupVersion: "apps/v1",
							APIResources: []metav1.APIResource{
								{Kind: "Deployment"},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			testName:   "resource_not_found",
			apiVersion: "apps/v1",
			kind:       "Deploymentx",
			dc: &discoveryClientStub{
				results: map[string]discoveryClientResult{
					"apps/v1": {
						resources: &metav1.APIResourceList{
							GroupVersion: "apps/v1",
							APIResources: []metav1.APIResource{
								{Kind: "Deployment"},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			testName:   "group_version_not_found",
			apiVersion: "apps/v2",
			kind:       "Deployment",
			dc: &discoveryClientStub{
				results: map[string]discoveryClientResult{},
			},
			want: false,
		},
		{
			testName:   "error",
			apiVersion: "apps/v1",
			kind:       "Deployment",
			dc: &discoveryClientStub{
				results: map[string]discoveryClientResult{
					"apps/v1": {
						resources: nil,
						err:       fmt.Errorf("server error"),
					},
				},
			},
			want:    false,
			wantErr: new("server error"),
		},
	} {
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()

			got, err := CheckGVKExists(tt.dc, tt.apiVersion, tt.kind)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("CheckGVKExists returned unexpected error: %v", err)
				}

				if !regexp.MustCompile(regexp.QuoteMeta(*tt.wantErr)).MatchString(err.Error()) {
					t.Errorf("CheckGVKExists returned error %q, want %q", err.Error(), *tt.wantErr)
				}

				return
			}

			if tt.wantErr != nil {
				t.Errorf("CheckGVKExists returned no error, want %q", *tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("CheckGVKExists returned %t, want %t", got, tt.want)
			}
		})
	}
}

package k8sutils

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestManifestIdentity(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		manifest      map[string]any
		wantGVK       *schema.GroupVersionKind
		wantNamespace string
		wantName      string
		errMsg        string
	}{
		{name: "missing apiVersion", manifest: map[string]any{"kind": "ConfigMap", "metadata": map[string]any{"name": "example"}}, errMsg: "manifest.apiVersion is required"},
		{name: "missing kind", manifest: map[string]any{"apiVersion": "v1", "metadata": map[string]any{"name": "example"}}, errMsg: "manifest.kind is required"},
		{name: "missing name", manifest: map[string]any{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]any{}}, errMsg: "manifest.metadata.name is required"},
		{name: "valid", manifest: map[string]any{"apiVersion": "apps/v1", "kind": "Deployment", "metadata": map[string]any{"name": "example", "namespace": "default"}}, wantGVK: &schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, wantNamespace: "default", wantName: "example"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotGVK, gotNamespace, gotName, err := ManifestIdentity(tc.manifest)
			if tc.errMsg != "" {
				if err == nil || err.Error() != tc.errMsg {
					t.Fatalf("ManifestIdentity() error = %v, want %q", err, tc.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("ManifestIdentity() unexpected error: %v", err)
			}
			if *gotGVK != *tc.wantGVK || gotNamespace != tc.wantNamespace || gotName != tc.wantName {
				t.Fatalf("ManifestIdentity() = (%v, %q, %q), want (%v, %q, %q)", gotGVK, gotNamespace, gotName, tc.wantGVK, tc.wantNamespace, tc.wantName)
			}
		})
	}
}

func TestBuildApplyOptions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		defaults FieldManager
		override *FieldManager
		want     metav1.ApplyOptions
	}{
		{name: "defaults only", defaults: FieldManager{Name: "provider", ForceConflicts: true}, want: metav1.ApplyOptions{FieldManager: "provider", Force: true}},
		{name: "override name", defaults: FieldManager{Name: "provider"}, override: &FieldManager{Name: "resource", ForceConflicts: true}, want: metav1.ApplyOptions{FieldManager: "resource", Force: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := BuildApplyOptions(tc.defaults, tc.override)
			if got.FieldManager != tc.want.FieldManager || got.Force != tc.want.Force {
				t.Fatalf("BuildApplyOptions() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

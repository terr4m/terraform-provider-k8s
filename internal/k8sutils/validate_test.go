package k8sutils

import "testing"

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		manifest  map[string]any
		wantCount int
	}{
		{name: "nil manifest", manifest: nil, wantCount: 1},
		{name: "missing fields", manifest: map[string]any{"kind": "ConfigMap"}, wantCount: 2},
		{name: "valid manifest", manifest: map[string]any{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]any{"name": "example"}}, wantCount: 0},
		{name: "forbidden fields", manifest: map[string]any{"apiVersion": "v1", "kind": "ConfigMap", "status": map[string]any{}, "metadata": map[string]any{"name": "example", "uid": "abc"}}, wantCount: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := len(ValidateManifest(tc.manifest)); got != tc.wantCount {
				t.Fatalf("ValidateManifest() count = %d, want %d", got, tc.wantCount)
			}
		})
	}
}

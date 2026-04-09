package k8sutils

import "testing"

func TestManifestFingerprints(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		manifest  map[string]any
		live      map[string]any
		wantEqual bool
	}{
		{
			name: "ignore unrelated fields",
			manifest: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "example", "namespace": "default"},
				"data":       map[string]any{"foo": "bar"},
			},
			live: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "example", "namespace": "default", "creationTimestamp": "2026-04-09T00:00:00Z", "uid": "abc123"},
				"data":       map[string]any{"foo": "bar", "baz": "qux"},
				"status":     map[string]any{"ignored": true},
			},
			wantEqual: true,
		},
		{
			name: "detect managed field drift",
			manifest: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "example", "namespace": "default"},
				"data":       map[string]any{"foo": "bar"},
			},
			live: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "example", "namespace": "default"},
				"data":       map[string]any{"foo": "changed"},
			},
			wantEqual: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			desired, err := DesiredManifestFingerprint(tc.manifest)
			if err != nil {
				t.Fatalf("DesiredManifestFingerprint returned error: %v", err)
			}

			observed, err := ObservedManifestFingerprint(tc.manifest, tc.live)
			if err != nil {
				t.Fatalf("ObservedManifestFingerprint returned error: %v", err)
			}

			if got := observed == desired; got != tc.wantEqual {
				t.Fatalf("ObservedManifestFingerprint equality = %t, want %t", got, tc.wantEqual)
			}
		})
	}
}

func TestDesiredManifestFingerprintIgnoresServerManagedFields(t *testing.T) {
	t.Parallel()

	withServerFields := map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":              "example",
			"namespace":         "default",
			"creationTimestamp": "2026-04-09T00:00:00Z",
			"managedFields":     []any{"ignored"},
			"uid":               "abc123",
		},
		"status": map[string]any{"ignored": true},
		"data": map[string]any{
			"foo": "bar",
		},
	}
	withoutServerFields := map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      "example",
			"namespace": "default",
		},
		"data": map[string]any{
			"foo": "bar",
		},
	}

	withHash, err := DesiredManifestFingerprint(withServerFields)
	if err != nil {
		t.Fatalf("DesiredManifestFingerprint returned error: %v", err)
	}

	withoutHash, err := DesiredManifestFingerprint(withoutServerFields)
	if err != nil {
		t.Fatalf("DesiredManifestFingerprint returned error: %v", err)
	}

	if withHash != withoutHash {
		t.Fatalf("DesiredManifestFingerprint with server fields = %q, want %q", withHash, withoutHash)
	}
}

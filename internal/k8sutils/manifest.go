package k8sutils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

var metadataServerManagedFields = []string{"creationTimestamp", "generation", "managedFields", "resourceVersion", "selfLink", "uid"}

// DesiredManifestFingerprint hashes the canonical desired manifest subset.
func DesiredManifestFingerprint(manifest map[string]any) (string, error) {
	canonical := sanitizeManifest(manifest)
	return hashObject(canonical)
}

// ObservedManifestFingerprint hashes the live object projected onto the manifest subset.
func ObservedManifestFingerprint(manifest, live map[string]any) (string, error) {
	canonical := sanitizeManifest(manifest)
	projected, err := projectValue(canonical, live)
	if err != nil {
		return "", err
	}

	return hashObject(projected)
}

func sanitizeManifest(manifest map[string]any) map[string]any {
	clone, _ := deepCopyValue(manifest).(map[string]any)
	if clone == nil {
		return map[string]any{}
	}

	delete(clone, "status")

	meta, ok := clone["metadata"].(map[string]any)
	if ok {
		for _, field := range metadataServerManagedFields {
			delete(meta, field)
		}
	}

	return clone
}

func projectValue(template, actual any) (any, error) {
	switch tv := template.(type) {
	case nil:
		return nil, nil
	case bool, string, float64, int64, int32, int, uint64, uint32, uint:
		return normalizeScalar(actual), nil
	case []any:
		actualSlice, ok := actual.([]any)
		if !ok {
			return make([]any, len(tv)), nil
		}

		result := make([]any, len(tv))
		for i, item := range tv {
			var actualItem any
			if i < len(actualSlice) {
				actualItem = actualSlice[i]
			}

			projected, err := projectValue(item, actualItem)
			if err != nil {
				return nil, err
			}

			result[i] = projected
		}

		return result, nil
	case map[string]any:
		actualMap, ok := actual.(map[string]any)
		if !ok {
			actualMap = map[string]any{}
		}

		result := make(map[string]any, len(tv))
		for key, item := range tv {
			projected, err := projectValue(item, actualMap[key])
			if err != nil {
				return nil, fmt.Errorf("project field %q: %w", key, err)
			}

			result[key] = projected
		}

		return result, nil
	default:
		return normalizeScalar(actual), nil
	}
}

func normalizeScalar(v any) any {
	switch n := v.(type) {
	case json.Number:
		return n.String()
	default:
		return v
	}
}

func hashObject(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal canonical object: %w", err)
	}

	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func deepCopyValue(v any) any {
	switch tv := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(tv))
		for key, value := range tv {
			result[key] = deepCopyValue(value)
		}
		return result
	case []any:
		result := make([]any, len(tv))
		for i, value := range tv {
			result[i] = deepCopyValue(value)
		}
		return result
	default:
		return tv
	}
}

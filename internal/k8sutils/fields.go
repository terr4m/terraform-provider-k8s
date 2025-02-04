package k8sutils

import (
	"fmt"
	"slices"
)

// ServerSideMetadataFields returns a list of fields that are only set server-side.
func ServerSideMetadataFields() []string {
	return []string{"creationTimestamp", "generation", "resourceVersion", "selfLink"}
}

// ServerSideFields returns a list of fields that are only set server-side.
func ServerSideFields() []string {
	return []string{"status"}
}

// RemoveServerSideFields removes server-side fields from the given object.
func RemoveServerSideFields(manager string, manifest, obj map[string]any, labels, annotations bool) error {
	meta, ok := obj["metadata"].(map[string]any)
	if !ok {
		return fmt.Errorf("expected metadata field, got %T", obj["metadata"])
	}

	for _, f := range ServerSideMetadataFields() {
		delete(meta, f)
	}

	if labels {
		lObj, ok := meta["labels"].(map[string]any)
		if ok {
			sourceLabels, _ := getLabelKeys(manifest)
			for k := range lObj {
				if !slices.Contains(sourceLabels, k) {
					delete(lObj, k)
				}
			}
		}
	}

	if annotations {
		aObj, ok := meta["annotations"].(map[string]any)
		if ok {
			sourceAnnotations, _ := getAnnotationKeys(manifest)
			for k := range aObj {
				if !slices.Contains(sourceAnnotations, k) {
					delete(aObj, k)
				}
			}
		}
	}

	mf, ok := meta["managedFields"].([]any)
	if ok {
		mfNew := make([]any, 0, 1)

		for _, v := range mf {
			fm, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("expected fieldManager field, got %T", v)
			}

			if fm["manager"] == manager && fm["operation"] == "Update" {
				delete(fm, "time")
				mfNew = append(mfNew, fm)
			}
		}

		meta["managedFields"] = mfNew
	}

	for _, f := range ServerSideFields() {
		delete(obj, f)
	}

	return nil
}

// getLabelKeys returns the keys of the labels field of the given object.
func getLabelKeys(obj map[string]any) ([]string, bool) {
	meta, ok := obj["metadata"].(map[string]any)
	if !ok {
		return nil, false
	}

	labels, ok := meta["labels"].(map[string]any)
	if !ok {
		return nil, false
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}

	return keys, true
}

// getAnnotationKeys returns the keys of the annotations field of the given object.
func getAnnotationKeys(obj map[string]any) ([]string, bool) {
	meta, ok := obj["metadata"].(map[string]any)
	if !ok {
		return nil, false
	}

	annotations, ok := meta["annotations"].(map[string]any)
	if !ok {
		return nil, false
	}

	keys := make([]string, 0, len(annotations))
	for k := range annotations {
		keys = append(keys, k)
	}

	return keys, true
}

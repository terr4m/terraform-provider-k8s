package k8sutils

import "fmt"

type ValidationError struct {
	Path    string
	Summary string
	Detail  string
}

func ValidateManifest(manifest map[string]any) []ValidationError {
	if manifest == nil {
		return []ValidationError{{Path: "", Summary: "Manifest is required.", Detail: "expected non-null manifest"}}
	}

	errs := make([]ValidationError, 0)
	for _, name := range []string{"apiVersion", "kind", "metadata"} {
		if _, ok := manifest[name]; !ok {
			errs = append(errs, ValidationError{
				Path:    name,
				Summary: "Missing required attribute.",
				Detail:  fmt.Sprintf("expected manifest to have attribute %q", name),
			})
		}
	}

	metadataValue, ok := manifest["metadata"]
	if !ok {
		return errs
	}

	metadata, ok := metadataValue.(map[string]any)
	if !ok {
		return append(errs, ValidationError{Path: "metadata", Summary: "Metadata type not object.", Detail: "expected metadata to be an object"})
	}

	if _, ok := metadata["name"]; !ok {
		errs = append(errs, ValidationError{Path: "metadata.name", Summary: "Missing required attribute.", Detail: "expected metadata.name to be set"})
	}

	if _, ok := manifest["status"]; ok {
		errs = append(errs, ValidationError{Path: "status", Summary: "Forbidden attribute.", Detail: "manifest must not include status; it is managed by the API server"})
	}

	for _, field := range metadataServerManagedFields {
		if _, ok := metadata[field]; ok {
			errs = append(errs, ValidationError{
				Path:    "metadata." + field,
				Summary: "Forbidden attribute.",
				Detail:  fmt.Sprintf("manifest must not include metadata.%s; it is managed by the API server", field),
			})
		}
	}

	return errs
}

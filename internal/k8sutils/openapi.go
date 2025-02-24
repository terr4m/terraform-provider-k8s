package k8sutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/openapi"

	"github.com/pb33f/libopenapi"
	high "github.com/pb33f/libopenapi/datamodel/high/base"
)

// GetOpenAPISchema returns the OpenAPI schema for the given group version and schema ID.
func GetOpenAPISchema(gv openapi.GroupVersion, gvk *schema.GroupVersionKind) (*high.Schema, error) {
	bytes, err := gv.Schema("application/json")
	if err != nil {
		return nil, err
	}

	document, err := libopenapi.NewDocument(bytes)
	if err != nil {
		return nil, err
	}

	docModel, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to build model: %v", errs)
	}

	for sp := docModel.Model.Components.Schemas.First(); sp != nil; sp = sp.Next() {
		spn := sp.Key()
		spv := sp.Value()
		s := spv.Schema()
		if s == nil {
			return nil, fmt.Errorf("schema %q can't be loaded", spn)
		}

		n, ok := s.Extensions.Get("x-kubernetes-group-version-kind")
		if !ok {
			continue
		}

		var sgvk []map[string]string
		err = n.Decode(&sgvk)
		if err != nil {
			return nil, err
		}

		if len(sgvk) == 1 && gvk.Kind == sgvk[0]["kind"] {
			return s, nil
		}
	}

	return nil, fmt.Errorf("schema %q not found", gvk.Kind)
}

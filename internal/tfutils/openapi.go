package tfutils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	high "github.com/pb33f/libopenapi/datamodel/high/base"
)

// GetTfTypeFromOpenAPI converts an OpenAPI schema to a Terraform type.
func GetTfTypeFromOpenAPI(s *high.Schema) (tftypes.Type, error) {
	return getTfTypeFromOpenAPIScalar(s)
}

// getTfTypeFromOpenAPIScalar converts an OpenAPI scalar type to a Terraform type.
func getTfTypeFromOpenAPIScalar(s *high.Schema) (tftypes.Type, error) {
	ts := s.Type
	slices.Sort(ts)
	t := strings.Join(ts, "-")

	switch t {
	case "integer", "number":
		return tftypes.Number, nil
	case "boolean":
		return tftypes.Bool, nil
	case "string":
		return tftypes.String, nil
	case "array":
		return getTfTypeFromOpenAPIArray(s)
	case "object":
		return getTfTypeFromOpenAPIObject(s)
	case "":
		if s.Format == "int-or-string" {
			return tftypes.DynamicPseudoType, nil
		}

		if len(s.OneOf) > 1 {
			return tftypes.DynamicPseudoType, nil
		}

		if len(s.AllOf) == 1 {
			sc := s.AllOf[0].Schema()
			if sc == nil {
				return nil, fmt.Errorf("all of schema can't be loaded")
			}
			return getTfTypeFromOpenAPIScalar(sc)
		}

		fallthrough
	default:
		return nil, fmt.Errorf("unexpected scalar type: %s", t)
	}
}

// getTfTypeFromOpenAPIArray converts an OpenAPI array type to a Terraform type.
func getTfTypeFromOpenAPIArray(s *high.Schema) (tftypes.Type, error) {
	if s.Items.IsA() {
		cs := s.Items.A.Schema()
		if cs == nil {
			return nil, fmt.Errorf("item schema can't be loaded")
		}

		et, err := getTfTypeFromOpenAPIScalar(cs)
		if err != nil {
			return nil, err
		}

		return tftypes.Tuple{ElementTypes: []tftypes.Type{et}}, nil
	}

	return tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.DynamicPseudoType}}, nil
}

// getTfTypeFromOpenAPIObject converts an OpenAPI object type to a Terraform type.
func getTfTypeFromOpenAPIObject(s *high.Schema) (tftypes.Type, error) {
	if s.Properties != nil {
		attrTypes := make(map[string]tftypes.Type, s.Properties.Len())

		for pp := s.Properties.First(); pp != nil; pp = pp.Next() {
			pn := pp.Key()
			pv := pp.Value()
			ps := pv.Schema()
			if ps == nil {
				return nil, fmt.Errorf("property schema %q can't be loaded", pn)
			}

			attrType, err := getTfTypeFromOpenAPIScalar(ps)
			if err != nil {
				return nil, err
			}

			attrTypes[pn] = attrType
		}

		return tftypes.Object{AttributeTypes: attrTypes}, nil
	} else if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
		cs := s.AdditionalProperties.A.Schema()
		if cs == nil {
			return nil, fmt.Errorf("additional property schema can't be loaded")
		}

		et, err := getTfTypeFromOpenAPIScalar(cs)
		if err != nil {
			return nil, err
		}

		return tftypes.Map{ElementType: et}, nil
	}

	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{"#": tftypes.DynamicPseudoType}}, nil
}

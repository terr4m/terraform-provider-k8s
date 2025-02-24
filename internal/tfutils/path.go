package tfutils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// AttributePathToPath converts an attribute path to a path.
func AttributePathToPath(ap *tftypes.AttributePath) (path.Path, error) {
	p := path.Empty()
	for _, s := range ap.Steps() {
		switch v := s.(type) {
		case tftypes.AttributeName:
			p = p.AtName(string(v))
		case tftypes.ElementKeyString:
			p = p.AtMapKey(string(v))
		case tftypes.ElementKeyInt:
			p = p.AtTupleIndex(int(v))
		default:
			return path.Path{}, fmt.Errorf("unexpected step type %T", s)
		}
	}

	return p, nil
}

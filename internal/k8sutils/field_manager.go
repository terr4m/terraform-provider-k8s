package k8sutils

// NewFieldManager creates a new field manager.
func NewFieldManager(name string, forceConflicts bool) *FieldManager {
	return &FieldManager{
		Name:           name,
		ForceConflicts: forceConflicts,
	}
}

// FieldManager holds the field manager configuration.
type FieldManager struct {
	Name           string
	ForceConflicts bool
}

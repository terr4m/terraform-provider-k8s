package k8sutils

// Ptr returns a pointer to the given value.
//
//go:fix inline
func Ptr[T any](d T) *T {
	return new(d)
}

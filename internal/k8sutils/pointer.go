package k8sutils

// Ptr returns a pointer to the given value.
func Ptr[T any](d T) *T {
	return &d
}

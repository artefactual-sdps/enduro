package ref

// New creates a pointer to x.
func New[T any](x T) *T {
	return &x
}

// Deref dereferences the pointer variable x.
func Deref[T any](x *T) T {
	return *x
}

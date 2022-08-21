package ref

// New creates a pointer to x.
func New[T any](x T) *T {
	return &x
}

// Deref dereferences the pointer variable x.
func Deref[T any](x *T) T {
	return *x
}

// Deref dereferences the pointer variable x.
// When the point is nil, return the zero value of T.
func DerefZero[T any](x *T) T {
	if x == nil {
		var z T
		return z
	}
	return *x
}

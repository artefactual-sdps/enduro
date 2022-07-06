package ref

// New creates a pointer to x.
func New[T any](x T) *T {
	return &x
}

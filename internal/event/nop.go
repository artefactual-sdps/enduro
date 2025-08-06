package event

import (
	"context"
	"errors"
)

type nopService[T any] struct{}

var _ Service[any] = (*nopService[any])(nil)

// NewServiceNop returns a new instance of a no-op event service.
func NewServiceNop[T any]() Service[T] {
	return &nopService[T]{}
}

func (*nopService[T]) PublishEvent(ctx context.Context, event T) {}

func (*nopService[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	return nil, errors.New("Subscribe not supported by nop service")
}

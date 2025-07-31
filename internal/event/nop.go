package event

import (
	"context"
	"errors"
)

type nopService[T any] struct{}

func (*nopService[T]) PublishEvent(ctx context.Context, event T) {}

func (*nopService[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	return nil, errors.New("Subscribe not supported by nop service")
}

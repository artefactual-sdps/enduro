package event

import (
	"context"
)

const (
	// EventBufferSize is the buffer size of the channel for each subscription.
	EventBufferSize = 256
)

// Service represents a generic service for managing event dispatch and event
// listeners (aka subscriptions).
type Service[T any] interface {
	// PublishEvent publishes an event to a user's event listeners.
	// If the user is not currently subscribed then this is a no-op.
	PublishEvent(ctx context.Context, event T)

	// Subscribe creates a subscription. Caller must call Subscription.Close() when done
	// with the subscription.
	Subscribe(ctx context.Context) (Subscription[T], error)
}

// Subscription represents a stream of events for a single user.
type Subscription[T any] interface {
	// C returns the event stream for all user's events.
	C() <-chan T

	// Close closes the event stream channel and disconnects from the event service.
	Close() error
}

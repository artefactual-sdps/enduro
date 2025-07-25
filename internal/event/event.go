package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

const (
	// EventBufferSize is the buffer size of the channel for each subscription.
	EventBufferSize = 16
)

// Service represents a generic service for managing event dispatch and event
// listeners (aka subscriptions).
type Service[T any] interface {
	// Publishes an event to a user's event listeners.
	// If the user is not currently subscribed then this is a no-op.
	PublishEvent(ctx context.Context, event T)

	// Creates a subscription. Caller must call Subscription.Close() when done
	// with the subscription.
	Subscribe(ctx context.Context) (Subscription[T], error)
}

// Subscription represents a stream of events for a single user.
type Subscription[T any] interface {
	// Event stream for all user's events.
	C() <-chan T

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}

// EventService represents a service for managing ingest event dispatch and event
// listeners (aka subscriptions).
type EventService = Service[*goaingest.IngestEvent]

// StorageEventService represents a service for managing storage event dispatch and event listeners.
type StorageEventService = Service[*goastorage.StorageEvent]

// EventSubscription represents a stream of ingest events for a single user.
type EventSubscription = Subscription[*goaingest.IngestEvent]

// StorageSubscription represents a stream of storage events for a single user.
type StorageSubscription = Subscription[*goastorage.StorageEvent]

// NopEventService returns an event service that does nothing.
func NopEventService() EventService { return &nopService[*goaingest.IngestEvent]{} }

// NopStorageEventService returns a storage event service that does nothing.
func NopStorageEventService() StorageEventService {
	return &nopService[*goastorage.StorageEvent]{}
}

type nopService[T any] struct{}

func (*nopService[T]) PublishEvent(ctx context.Context, event T) {}

func (*nopService[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	panic("not implemented")
}

package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

const (
	// EventBufferSize is the buffer size of the channel for each subscription.
	EventBufferSize = 16
)

// EventService represents a service for managing event dispatch and event
// listeners (aka subscriptions).
type EventService interface {
	// Publishes an event to a user's event listeners.
	// If the user is not currently subscribed then this is a no-op.
	PublishEvent(ctx context.Context, event *goaingest.MonitorEvent)

	// Creates a subscription. Caller must call Subscription.Close() when done
	// with the subscription.
	Subscribe(ctx context.Context) (Subscription, error)
}

// NopEventService returns an event service that does nothing.
func NopEventService() EventService { return &nopEventService{} }

type nopEventService struct{}

func (*nopEventService) PublishEvent(ctx context.Context, event *goaingest.MonitorEvent) {}

func (*nopEventService) Subscribe(ctx context.Context) (Subscription, error) {
	panic("not implemented")
}

// Subscription represents a stream of events for a single user.
type Subscription interface {
	// Event stream for all user's event.
	C() <-chan *goaingest.MonitorEvent

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}

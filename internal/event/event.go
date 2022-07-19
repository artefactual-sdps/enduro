package event

import (
	"context"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
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
	PublishEvent(event *goapackage.EnduroMonitorEvent)

	// Creates a subscription. Caller must call Subscription.Close() when done
	// with the subscription.
	Subscribe(ctx context.Context) (Subscription, error)
}

// NopEventService returns an event service that does nothing.
func NopEventService() EventService { return &nopEventService{} }

type nopEventService struct{}

func (*nopEventService) PublishEvent(event *goapackage.EnduroMonitorEvent) {}

func (*nopEventService) Subscribe(ctx context.Context) (Subscription, error) {
	panic("not implemented")
}

// Subscription represents a stream of events for a single user.
type Subscription interface {
	// Event stream for all user's event.
	C() <-chan *goapackage.EnduroMonitorEvent

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}

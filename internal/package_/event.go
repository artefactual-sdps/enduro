package package_

import (
	"context"
	"sync"

	"github.com/google/uuid"

	goapackage "github.com/artefactual-labs/enduro/internal/api/gen/package_"
)

const (
	// EventBufferSize is the buffer size of the channel for each subscription.
	EventBufferSize = 16

	EventTypePackageCreated = "package:created"
	EventTypePackageUpdated = "package:updated"
	EventTypePackageDeleted = "package:deleted"
)

// EventService represents a service for managing event dispatch and event
// listeners (aka subscriptions).
type EventService interface {
	// Publishes an event to a user's event listeners.
	// If the user is not currently subscribed then this is a no-op.
	PublishEvent(event *goapackage.EnduroMonitorUpdate)

	// Creates a subscription. Caller must call Subscription.Close() when done
	// with the subscription.
	Subscribe(ctx context.Context) (Subscription, error)
}

// EventService represents a service for managing events in the system.
type EventServiceImpl struct {
	mu   sync.Mutex
	subs map[uuid.UUID]*SubscriptionImpl
}

// NewEventService returns a new instance of EventService.
func NewEventService() *EventServiceImpl {
	return &EventServiceImpl{
		subs: map[uuid.UUID]*SubscriptionImpl{},
	}
}

// PublishEvent publishes event to all of a user's subscriptions.
//
// If user's channel is full then the user is disconnected. This is to prevent
// slow users from blocking progress.
func (s *EventServiceImpl) PublishEvent(event *goapackage.EnduroMonitorUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Publish event to all subscriptions for the user.
	for _, sub := range s.subs {
		select {
		case sub.c <- event:
		default:
			s.unsubscribe(sub)
		}
	}
}

// Subscribe creates a new subscription.
func (s *EventServiceImpl) Subscribe(ctx context.Context) (Subscription, error) {
	sub := &SubscriptionImpl{
		service: s,
		c:       make(chan *goapackage.EnduroMonitorUpdate, EventBufferSize),
		id:      uuid.New(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subs[sub.id] = sub

	return sub, nil
}

// Unsubscribe disconnects sub from the service.
func (s *EventServiceImpl) Unsubscribe(sub *SubscriptionImpl) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.unsubscribe(sub)
}

func (s *EventServiceImpl) unsubscribe(sub *SubscriptionImpl) {
	// Only close the underlying channel once. Otherwise Go will panic.
	sub.once.Do(func() {
		close(sub.c)
	})

	_, ok := s.subs[sub.id]
	if !ok {
		return
	}

	delete(s.subs, sub.id)
}

// NopEventService returns an event service that does nothing.
func NopEventService() EventService { return &nopEventService{} }

type nopEventService struct{}

func (*nopEventService) PublishEvent(event *goapackage.EnduroMonitorUpdate) {}

func (*nopEventService) Subscribe(ctx context.Context) (Subscription, error) {
	panic("not implemented")
}

// Subscription represents a stream of events for a single user.
type Subscription interface {
	// Event stream for all user's event.
	C() <-chan *goapackage.EnduroMonitorUpdate

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}

// SubscriptionImpl represents a stream of user-related events.
type SubscriptionImpl struct {
	service *EventServiceImpl                    // service subscription was created from
	c       chan *goapackage.EnduroMonitorUpdate // channel of events
	once    sync.Once                            // ensures c only closed once
	id      uuid.UUID                            // subscription identifier
}

var _ Subscription = (*SubscriptionImpl)(nil)

// Close disconnects the subscription from the service it was created from.
func (s *SubscriptionImpl) Close() error {
	s.service.Unsubscribe(s)
	return nil
}

// C returns a receive-only channel of user-related events.
func (s *SubscriptionImpl) C() <-chan *goapackage.EnduroMonitorUpdate {
	return s.c
}

package event

import (
	"context"
	"sync"

	"github.com/google/uuid"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
)

// EventServiceInMemImpl represents a service for managing events in the system.
type EventServiceInMemImpl struct {
	mu   sync.Mutex
	subs map[uuid.UUID]*SubscriptionInMemImpl
}

// NewEventServiceInMemImpl returns a new instance of EventService.
func NewEventServiceInMemImpl() *EventServiceInMemImpl {
	return &EventServiceInMemImpl{
		subs: map[uuid.UUID]*SubscriptionInMemImpl{},
	}
}

// PublishEvent publishes event to all of a user's subscriptions.
//
// If user's channel is full then the user is disconnected. This is to prevent
// slow users from blocking progress.
func (s *EventServiceInMemImpl) PublishEvent(event *goapackage.EnduroMonitorEvent) {
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
func (s *EventServiceInMemImpl) Subscribe(ctx context.Context) (Subscription, error) {
	sub := &SubscriptionInMemImpl{
		service: s,
		c:       make(chan *goapackage.EnduroMonitorEvent, EventBufferSize),
		id:      uuid.New(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subs[sub.id] = sub

	return sub, nil
}

// Unsubscribe disconnects sub from the service.
func (s *EventServiceInMemImpl) Unsubscribe(sub *SubscriptionInMemImpl) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.unsubscribe(sub)
}

func (s *EventServiceInMemImpl) unsubscribe(sub *SubscriptionInMemImpl) {
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

// SubscriptionInMemImpl represents a stream of user-related events.
type SubscriptionInMemImpl struct {
	service *EventServiceInMemImpl              // service subscription was created from
	c       chan *goapackage.EnduroMonitorEvent // channel of events
	once    sync.Once                           // ensures c only closed once
	id      uuid.UUID                           // subscription identifier
}

var _ Subscription = (*SubscriptionInMemImpl)(nil)

// Close disconnects the subscription from the service it was created from.
func (s *SubscriptionInMemImpl) Close() error {
	s.service.Unsubscribe(s)
	return nil
}

// C returns a receive-only channel of user-related events.
func (s *SubscriptionInMemImpl) C() <-chan *goapackage.EnduroMonitorEvent {
	return s.c
}

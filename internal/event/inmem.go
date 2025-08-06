package event

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// serviceInMemImpl represents a generic service for managing events in memory.
type serviceInMemImpl[T any] struct {
	mu   sync.Mutex
	subs map[uuid.UUID]*subscriptionInMemImpl[T]
}

var _ Service[any] = (*serviceInMemImpl[any])(nil)

// NewServiceInMem returns a new instance of a generic event service.
func NewServiceInMem[T any]() Service[T] {
	return &serviceInMemImpl[T]{
		subs: map[uuid.UUID]*subscriptionInMemImpl[T]{},
	}
}

// PublishEvent publishes event to all subscriptions.
func (s *serviceInMemImpl[T]) PublishEvent(ctx context.Context, event T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Publish event to all subscriptions for the user.
	for _, sub := range s.subs {
		select {
		case sub.c <- event:
			// Event successfully sent to subscriber.
		default:
			// Skip this event if the subscriber's buffer is full.
			// This prevents disconnecting slow subscribers and allows
			// them to continue receiving future events.
		}
	}
}

// Subscribe creates a new subscription.
func (s *serviceInMemImpl[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	sub := &subscriptionInMemImpl[T]{
		service: s,
		c:       make(chan T, EventBufferSize),
		id:      uuid.New(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subs[sub.id] = sub

	return sub, nil
}

// unsubscribe disconnects sub from the service.
func (s *serviceInMemImpl[T]) unsubscribe(sub *subscriptionInMemImpl[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

// subscriptionInMemImpl represents a stream of events.
type subscriptionInMemImpl[T any] struct {
	// Service the subscription was created from.
	service *serviceInMemImpl[T]
	// Channel of events.
	c chan T
	// Ensures c is only closed once.
	once sync.Once
	// Subscription identifier.
	id uuid.UUID
}

var _ Subscription[any] = (*subscriptionInMemImpl[any])(nil)

// Close disconnects the subscription from the service it was created from.
func (s *subscriptionInMemImpl[T]) Close() error {
	s.service.unsubscribe(s)
	return nil
}

// C returns a receive-only channel of events.
func (s *subscriptionInMemImpl[T]) C() <-chan T {
	return s.c
}

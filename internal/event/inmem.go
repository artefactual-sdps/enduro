package event

import (
	"context"
	"sync"

	"github.com/google/uuid"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// ServiceInMemImpl represents a generic service for managing events in the system.
type ServiceInMemImpl[T any] struct {
	mu   sync.Mutex
	subs map[uuid.UUID]*SubscriptionInMemImpl[T]
}

var (
	_ Service[*goaingest.IngestMonitorEvent]   = (*ServiceInMemImpl[*goaingest.IngestMonitorEvent])(nil)
	_ Service[*goastorage.StorageMonitorEvent] = (*ServiceInMemImpl[*goastorage.StorageMonitorEvent])(nil)
)

// NewServiceInMem returns a new instance of a generic event service.
func NewServiceInMem[T any]() *ServiceInMemImpl[T] {
	return &ServiceInMemImpl[T]{
		subs: map[uuid.UUID]*SubscriptionInMemImpl[T]{},
	}
}

// NewEventServiceInMem returns a new instance of an ingest event service.
func NewEventServiceInMem() EventService {
	return NewServiceInMem[*goaingest.IngestMonitorEvent]()
}

// NewStorageEventServiceInMem returns a new instance of a storage event service.
func NewStorageEventServiceInMem() StorageEventService {
	return NewServiceInMem[*goastorage.StorageMonitorEvent]()
}

// Backward compatibility functions
func NewEventServiceInMemImpl() EventService {
	return NewEventServiceInMem()
}

func NewStorageEventServiceInMemImpl() StorageEventService {
	return NewStorageEventServiceInMem()
}

// PublishEvent publishes event to all subscriptions.
func (s *ServiceInMemImpl[T]) PublishEvent(ctx context.Context, event T) {
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
func (s *ServiceInMemImpl[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	sub := &SubscriptionInMemImpl[T]{
		service: s,
		c:       make(chan T, EventBufferSize),
		id:      uuid.New(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subs[sub.id] = sub

	return sub, nil
}

// Unsubscribe disconnects sub from the service.
func (s *ServiceInMemImpl[T]) Unsubscribe(sub *SubscriptionInMemImpl[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.unsubscribe(sub)
}

func (s *ServiceInMemImpl[T]) unsubscribe(sub *SubscriptionInMemImpl[T]) {
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

// SubscriptionInMemImpl represents a stream of events.
type SubscriptionInMemImpl[T any] struct {
	service *ServiceInMemImpl[T] // service subscription was created from
	c       chan T               // channel of events
	once    sync.Once            // ensures c only closed once
	id      uuid.UUID            // subscription identifier
}

var (
	_ Subscription[*goaingest.IngestMonitorEvent]   = (*SubscriptionInMemImpl[*goaingest.IngestMonitorEvent])(nil)
	_ Subscription[*goastorage.StorageMonitorEvent] = (*SubscriptionInMemImpl[*goastorage.StorageMonitorEvent])(nil)
)

// Close disconnects the subscription from the service it was created from.
func (s *SubscriptionInMemImpl[T]) Close() error {
	s.service.Unsubscribe(s)
	return nil
}

// C returns a receive-only channel of events.
func (s *SubscriptionInMemImpl[T]) C() <-chan T {
	return s.c
}

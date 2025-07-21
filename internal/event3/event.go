package event3

import (
	"context"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"

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

// Type aliases for convenience
type (
	IngestEventService  = Service[*goaingest.MonitorEvent]
	StorageEventService = Service[*goastorage.StorageMonitorEvent]
	IngestSubscription  = Subscription[*goaingest.MonitorEvent]
	StorageSubscription = Subscription[*goastorage.StorageMonitorEvent]
)

// NopIngestEventService returns an ingest event service that does nothing.
func NopIngestEventService() IngestEventService {
	return &nopService[*goaingest.MonitorEvent]{}
}

// NopStorageEventService returns a storage event service that does nothing.
func NopStorageEventService() StorageEventService {
	return &nopService[*goastorage.StorageMonitorEvent]{}
}

type nopService[T any] struct{}

func (*nopService[T]) PublishEvent(ctx context.Context, event T) {}

func (*nopService[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	panic("not implemented")
}

// Constructor functions for different backends

// NewIngestEventServiceInMem returns a new instance of an in-memory ingest event service.
func NewIngestEventServiceInMem() IngestEventService {
	return NewServiceInMem[*goaingest.MonitorEvent]()
}

// NewStorageEventServiceInMem returns a new instance of an in-memory storage event service.
func NewStorageEventServiceInMem() StorageEventService {
	return NewServiceInMem[*goastorage.StorageMonitorEvent]()
}

// NewIngestEventServiceRedis returns a new instance of a Redis ingest event service.
func NewIngestEventServiceRedis(logger logr.Logger, tp trace.TracerProvider, cfg *Config) (IngestEventService, error) {
	return NewServiceRedis(logger, tp, cfg, &IngestEventSerializer{})
}

// NewStorageEventServiceRedis returns a new instance of a Redis storage event service.
func NewStorageEventServiceRedis(
	logger logr.Logger,
	tp trace.TracerProvider,
	cfg *Config,
) (StorageEventService, error) {
	return NewServiceRedis(logger, tp, cfg, &StorageEventSerializer{})
}

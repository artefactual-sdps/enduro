package event

import (
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// Type aliases for convenience.
type (
	IngestEventService  = Service[*goaingest.IngestEvent]
	StorageEventService = Service[*goastorage.StorageEvent]
	IngestSubscription  = Subscription[*goaingest.IngestEvent]
	StorageSubscription = Subscription[*goastorage.StorageEvent]
)

// Compile-time interface compliance checks.
var (
	_ IngestEventService  = (*serviceInMemImpl[*goaingest.IngestEvent])(nil)
	_ StorageEventService = (*serviceInMemImpl[*goastorage.StorageEvent])(nil)
	_ IngestSubscription  = (*subscriptionInMemImpl[*goaingest.IngestEvent])(nil)
	_ StorageSubscription = (*subscriptionInMemImpl[*goastorage.StorageEvent])(nil)
	_ IngestEventService  = (*nopService[*goaingest.IngestEvent])(nil)
	_ StorageEventService = (*nopService[*goastorage.StorageEvent])(nil)
	_ IngestEventService  = (*serviceRedisImpl[*goaingest.IngestEvent])(nil)
	_ StorageEventService = (*serviceRedisImpl[*goastorage.StorageEvent])(nil)
	_ IngestSubscription  = (*subscriptionRedisImpl[*goaingest.IngestEvent])(nil)
	_ StorageSubscription = (*subscriptionRedisImpl[*goastorage.StorageEvent])(nil)
)

// NewIngestEventServiceInMem returns a new instance of an in-memory ingest event service.
func NewIngestEventServiceInMem() IngestEventService {
	return newServiceInMem[*goaingest.IngestEvent]()
}

// NewStorageEventServiceInMem returns a new instance of an in-memory storage event service.
func NewStorageEventServiceInMem() StorageEventService {
	return newServiceInMem[*goastorage.StorageEvent]()
}

// NewIngestEventServiceNop returns an ingest event service that does nothing.
func NewIngestEventServiceNop() IngestEventService {
	return &nopService[*goaingest.IngestEvent]{}
}

// NewStorageEventServiceNop returns a storage event service that does nothing.
func NewStorageEventServiceNop() StorageEventService {
	return &nopService[*goastorage.StorageEvent]{}
}

// NewIngestEventServiceRedis returns a new instance of a Redis ingest event service.
func NewIngestEventServiceRedis(
	logger logr.Logger,
	tp trace.TracerProvider,
	cfg *Config,
) (IngestEventService, error) {
	return newServiceRedis(logger, tp, cfg.RedisAddress, cfg.IngestRedisChannel, &ingestEventSerializer{})
}

// NewStorageEventServiceRedis returns a new instance of a Redis storage event service.
func NewStorageEventServiceRedis(
	logger logr.Logger,
	tp trace.TracerProvider,
	cfg *Config,
) (StorageEventService, error) {
	return newServiceRedis(logger, tp, cfg.RedisAddress, cfg.StorageRedisChannel, &storageEventSerializer{})
}

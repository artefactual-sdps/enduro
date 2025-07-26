package event

import (
	"context"
	"errors"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// NopIngestEventService returns an ingest event service that does nothing.
func NopIngestEventService() IngestEventService {
	return &nopService[*goaingest.IngestEvent]{}
}

// NopStorageEventService returns a storage event service that does nothing.
func NopStorageEventService() StorageEventService {
	return &nopService[*goastorage.StorageEvent]{}
}

type nopService[T any] struct{}

func (*nopService[T]) PublishEvent(ctx context.Context, event T) {}

func (*nopService[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	return nil, errors.New("Subscribe not supported by nop service")
}

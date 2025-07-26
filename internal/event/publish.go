package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// PublishIngestEvent publishes an ingest event with type safety.
func PublishIngestEvent(ctx context.Context, svc IngestEventService, event any) {
	update := &goaingest.IngestEvent{}

	switch v := event.(type) {
	case *goaingest.IngestPingEvent:
		update.IngestValue = v
	case *goaingest.SIPCreatedEvent:
		update.IngestValue = v
	case *goaingest.SIPUpdatedEvent:
		update.IngestValue = v
	case *goaingest.SIPStatusUpdatedEvent:
		update.IngestValue = v
	case *goaingest.SIPWorkflowCreatedEvent:
		update.IngestValue = v
	case *goaingest.SIPWorkflowUpdatedEvent:
		update.IngestValue = v
	case *goaingest.SIPTaskCreatedEvent:
		update.IngestValue = v
	case *goaingest.SIPTaskUpdatedEvent:
		update.IngestValue = v
	default:
		panic("tried to publish unexpected event")
	}

	svc.PublishEvent(ctx, update)
}

// PublishStorageEvent publishes a storage event with type safety.
func PublishStorageEvent(ctx context.Context, svc StorageEventService, event any) {
	update := &goastorage.StorageEvent{}

	switch v := event.(type) {
	case *goastorage.StoragePingEvent:
		update.StorageValue = v
	case *goastorage.LocationCreatedEvent:
		update.StorageValue = v
	case *goastorage.LocationUpdatedEvent:
		update.StorageValue = v
	case *goastorage.AIPCreatedEvent:
		update.StorageValue = v
	case *goastorage.AIPUpdatedEvent:
		update.StorageValue = v
	case *goastorage.AIPWorkflowCreatedEvent:
		update.StorageValue = v
	case *goastorage.AIPWorkflowUpdatedEvent:
		update.StorageValue = v
	case *goastorage.AIPTaskCreatedEvent:
		update.StorageValue = v
	case *goastorage.AIPTaskUpdatedEvent:
		update.StorageValue = v
	default:
		panic("tried to publish unexpected event")
	}

	svc.PublishEvent(ctx, update)
}

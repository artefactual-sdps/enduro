package event3

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// PublishIngestEvent publishes an ingest event with type safety.
func PublishIngestEvent(ctx context.Context, svc IngestEventService, event any) {
	switch v := event.(type) {
	case *goaingest.IngestPingEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPCreatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPStatusUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPWorkflowCreatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPWorkflowUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPTaskCreatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	case *goaingest.SIPTaskUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.IngestEvent{IngestValue: v})
	default:
		panic("invalid ingest event type")
	}
}

// PublishStorageEvent publishes a storage event with type safety.
func PublishStorageEvent(ctx context.Context, svc StorageEventService, event any) {
	switch v := event.(type) {
	case *goastorage.StoragePingEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.LocationCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.LocationUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.AIPCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.AIPUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.AIPWorkflowCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.AIPWorkflowUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.AIPTaskCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	case *goastorage.AIPTaskUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: v})
	default:
		panic("invalid storage event type")
	}
}

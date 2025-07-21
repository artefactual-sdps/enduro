package event3

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// PublishIngestEvent publishes an ingest event with type safety.
func PublishIngestEvent(ctx context.Context, svc IngestEventService, event any) {
	switch v := event.(type) {
	case *goaingest.MonitorPingEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPCreatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPStatusUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPWorkflowCreatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPWorkflowUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPTaskCreatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	case *goaingest.SIPTaskUpdatedEvent:
		svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
	default:
		panic("invalid ingest event type")
	}
}

// PublishStorageEvent publishes a storage event with type safety.
func PublishStorageEvent(ctx context.Context, svc StorageEventService, event any) {
	switch v := event.(type) {
	case *goastorage.StorageMonitorPingEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.LocationCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.LocationUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.AIPCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.AIPUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.WorkflowCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.WorkflowUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.TaskCreatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	case *goastorage.TaskUpdatedEvent:
		svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
	default:
		panic("invalid storage event type")
	}
}

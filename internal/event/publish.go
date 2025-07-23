package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func PublishEvent(ctx context.Context, events EventService, event any) {
	switch v := event.(type) {
	case *goaingest.IngestPingEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPCreatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPUpdatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPStatusUpdatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPWorkflowCreatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPWorkflowUpdatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPTaskCreatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPTaskUpdatedEvent:
		update := &goaingest.IngestMonitorEvent{IngestEvent: v}
		events.PublishEvent(ctx, update)
	default:
		panic("tried to publish unexpected event")
	}
}

func PublishStorageEvent(ctx context.Context, events StorageEventService, event any) {
	switch v := event.(type) {
	case *goastorage.StoragePingEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.LocationCreatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.LocationUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPCreatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPWorkflowCreatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPWorkflowUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPTaskCreatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPTaskUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{StorageEvent: v}
		events.PublishEvent(ctx, update)
	default:
		panic("tried to publish unexpected storage event")
	}
}

package event2

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func PublishEvent(ctx context.Context, events EventService, event any) {
	var e any

	switch v := event.(type) {
	case *goaingest.IngestPingEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPCreatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPUpdatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPStatusUpdatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPWorkflowCreatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPWorkflowUpdatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPTaskCreatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goaingest.SIPTaskUpdatedEvent:
		e = &goaingest.IngestMonitorEvent{IngestEvent: v}
	case *goastorage.StoragePingEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.LocationCreatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.LocationUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.AIPCreatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.AIPUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.AIPWorkflowCreatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.AIPWorkflowUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.AIPTaskCreatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	case *goastorage.AIPTaskUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{StorageEvent: v}
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(ctx, e)
}

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
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPCreatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPUpdatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPStatusUpdatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPWorkflowCreatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPWorkflowUpdatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPTaskCreatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goaingest.SIPTaskUpdatedEvent:
		e = &goaingest.IngestEvent{IngestValue: v}
	case *goastorage.StoragePingEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.LocationCreatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.LocationUpdatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.AIPCreatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.AIPUpdatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.AIPWorkflowCreatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.AIPWorkflowUpdatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.AIPTaskCreatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	case *goastorage.AIPTaskUpdatedEvent:
		e = &goastorage.StorageEvent{StorageValue: v}
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(ctx, e)
}

package event2

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func PublishEvent(ctx context.Context, events EventService, event any) {
	var e any

	switch v := event.(type) {
	case *goaingest.MonitorPingEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPCreatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPUpdatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPStatusUpdatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPWorkflowCreatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPWorkflowUpdatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPTaskCreatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goaingest.SIPTaskUpdatedEvent:
		e = &goaingest.MonitorEvent{Event: v}
	case *goastorage.StorageMonitorPingEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.LocationCreatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.LocationUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.AIPCreatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.AIPUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.WorkflowCreatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.WorkflowUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.TaskCreatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	case *goastorage.TaskUpdatedEvent:
		e = &goastorage.StorageMonitorEvent{Event: v}
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(ctx, e)
}

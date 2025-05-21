package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

func PublishEvent(ctx context.Context, events EventService, event any) {
	update := &goaingest.MonitorEvent{}

	switch v := event.(type) {
	case *goaingest.MonitorPingEvent:
		update.Event = v
	case *goaingest.SIPCreatedEvent:
		update.Event = v
	case *goaingest.SIPUpdatedEvent:
		update.Event = v
	case *goaingest.SIPStatusUpdatedEvent:
		update.Event = v
	case *goaingest.SIPWorkflowCreatedEvent:
		update.Event = v
	case *goaingest.SIPWorkflowUpdatedEvent:
		update.Event = v
	case *goaingest.SIPTaskCreatedEvent:
		update.Event = v
	case *goaingest.SIPTaskUpdatedEvent:
		update.Event = v
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(ctx, update)
}

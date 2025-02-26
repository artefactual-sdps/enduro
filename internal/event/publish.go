package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

func PublishEvent(ctx context.Context, events EventService, event interface{}) {
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
	case *goaingest.SIPLocationUpdatedEvent:
		update.Event = v
	case *goaingest.SIPPreservationActionCreatedEvent:
		update.Event = v
	case *goaingest.SIPPreservationActionUpdatedEvent:
		update.Event = v
	case *goaingest.SIPPreservationTaskCreatedEvent:
		update.Event = v
	case *goaingest.SIPPreservationTaskUpdatedEvent:
		update.Event = v
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(ctx, update)
}

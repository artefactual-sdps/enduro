package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

func PublishEvent(ctx context.Context, events EventService, event any) {
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

	events.PublishEvent(ctx, update)
}

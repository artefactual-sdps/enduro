package event

import (
	"context"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
)

func PublishEvent(ctx context.Context, events EventService, event interface{}) {
	update := &goapackage.MonitorEvent{}

	switch v := event.(type) {
	case *goapackage.MonitorPingEvent:
		update.Event = v
	case *goapackage.PackageCreatedEvent:
		update.Event = v
	case *goapackage.PackageUpdatedEvent:
		update.Event = v
	case *goapackage.PackageStatusUpdatedEvent:
		update.Event = v
	case *goapackage.PackageLocationUpdatedEvent:
		update.Event = v
	case *goapackage.PreservationActionCreatedEvent:
		update.Event = v
	case *goapackage.PreservationActionUpdatedEvent:
		update.Event = v
	case *goapackage.PreservationTaskCreatedEvent:
		update.Event = v
	case *goapackage.PreservationTaskUpdatedEvent:
		update.Event = v
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(update)
}

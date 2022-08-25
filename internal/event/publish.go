package event

import (
	"context"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
)

func PublishEvent(ctx context.Context, events EventService, event interface{}) {
	update := &goapackage.EnduroMonitorEvent{}

	switch v := event.(type) {
	case *goapackage.EnduroMonitorPingEvent:
		update.MonitorPingEvent = v
	case *goapackage.EnduroPackageCreatedEvent:
		update.PackageCreatedEvent = v
	case *goapackage.EnduroPackageUpdatedEvent:
		update.PackageUpdatedEvent = v
	case *goapackage.EnduroPackageStatusUpdatedEvent:
		update.PackageStatusUpdatedEvent = v
	case *goapackage.EnduroPackageLocationUpdatedEvent:
		update.PackageLocationUpdatedEvent = v
	case *goapackage.EnduroPreservationActionCreatedEvent:
		update.PreservationActionCreatedEvent = v
	case *goapackage.EnduroPreservationActionUpdatedEvent:
		update.PreservationActionUpdatedEvent = v
	case *goapackage.EnduroPreservationTaskCreatedEvent:
		update.PreservationTaskCreatedEvent = v
	case *goapackage.EnduroPreservationTaskUpdatedEvent:
		update.PreservationTaskUpdatedEvent = v
	default:
		panic("tried to publish unexpected event")
	}

	events.PublishEvent(update)
}

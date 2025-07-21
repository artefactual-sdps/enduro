package event

import (
	"context"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func PublishEvent(ctx context.Context, events EventService, event any) {
	switch v := event.(type) {
	case *goaingest.MonitorPingEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPCreatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPUpdatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPStatusUpdatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPWorkflowCreatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPWorkflowUpdatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPTaskCreatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goaingest.SIPTaskUpdatedEvent:
		update := &goaingest.MonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	default:
		panic("tried to publish unexpected event")
	}
}

// StorageEventService represents a service for managing storage event dispatch and event listeners.
type StorageEventService interface {
	// Publishes a storage event to a user's event listeners.
	PublishEvent(ctx context.Context, event *goastorage.StorageMonitorEvent)
	
	// Creates a subscription. Caller must call StorageSubscription.Close() when done.
	Subscribe(ctx context.Context) (StorageSubscription, error)
}

// StorageSubscription represents a stream of storage events for a single user.
type StorageSubscription interface {
	// Event stream for all user's storage events.
	C() <-chan *goastorage.StorageMonitorEvent

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}

func PublishStorageEvent(ctx context.Context, events StorageEventService, event any) {
	switch v := event.(type) {
	case *goastorage.StorageMonitorPingEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.LocationCreatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.LocationUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPCreatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.AIPUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.WorkflowCreatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.WorkflowUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.TaskCreatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	case *goastorage.TaskUpdatedEvent:
		update := &goastorage.StorageMonitorEvent{Event: v}
		events.PublishEvent(ctx, update)
	default:
		panic("tried to publish unexpected storage event")
	}
}

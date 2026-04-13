package storage

import (
	"context"
	"encoding/json"
	"fmt"

	storageclient "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	storageserver "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
)

// EventSerializer handles serialization/deserialization of storage events.
type EventSerializer struct{}

var _ event.Serializer[*goastorage.StorageEvent] = (*EventSerializer)(nil)

func (s *EventSerializer) Marshal(event *goastorage.StorageEvent) ([]byte, error) {
	return json.Marshal(storageserver.NewMonitorResponseBody(event))
}

func (s *EventSerializer) Unmarshal(data []byte) (*goastorage.StorageEvent, error) {
	payload := storageclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := storageclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return storageclient.NewMonitorStorageEventOK(&payload), nil
}

// Event is a type constraint for all storage events.
type Event interface {
	*goastorage.StoragePingEvent |
		*goastorage.LocationCreatedEvent |
		*goastorage.AIPCreatedEvent |
		*goastorage.AIPUpdatedEvent |
		*goastorage.AIPStatusUpdatedEvent |
		*goastorage.AIPLocationUpdatedEvent |
		*goastorage.AIPWorkflowCreatedEvent |
		*goastorage.AIPWorkflowUpdatedEvent |
		*goastorage.AIPTaskCreatedEvent |
		*goastorage.AIPTaskUpdatedEvent
}

// PublishEvent publishes a storage event with type safety.
func PublishEvent[E Event](ctx context.Context, svc event.Service[*goastorage.StorageEvent], event E) {
	svc.PublishEvent(ctx, &goastorage.StorageEvent{Value: NewEventValue(event)})
}

func NewEventValue[E Event](event E) goastorage.Value {
	switch e := any(event).(type) {
	case *goastorage.StoragePingEvent:
		return goastorage.NewValueStoragePingEvent(e)
	case *goastorage.LocationCreatedEvent:
		return goastorage.NewValueLocationCreatedEvent(e)
	case *goastorage.AIPCreatedEvent:
		return goastorage.NewValueAipCreatedEvent(e)
	case *goastorage.AIPUpdatedEvent:
		return goastorage.NewValueAipUpdatedEvent(e)
	case *goastorage.AIPStatusUpdatedEvent:
		return goastorage.NewValueAipStatusUpdatedEvent(e)
	case *goastorage.AIPLocationUpdatedEvent:
		return goastorage.NewValueAipLocationUpdatedEvent(e)
	case *goastorage.AIPWorkflowCreatedEvent:
		return goastorage.NewValueAipWorkflowCreatedEvent(e)
	case *goastorage.AIPWorkflowUpdatedEvent:
		return goastorage.NewValueAipWorkflowUpdatedEvent(e)
	case *goastorage.AIPTaskCreatedEvent:
		return goastorage.NewValueAipTaskCreatedEvent(e)
	case *goastorage.AIPTaskUpdatedEvent:
		return goastorage.NewValueAipTaskUpdatedEvent(e)
	default:
		panic(fmt.Sprintf("unsupported storage event type %T", event))
	}
}

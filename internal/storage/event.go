package storage

import (
	"context"
	"encoding/json"

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
		*goastorage.AIPStatusUpdatedEvent |
		*goastorage.AIPLocationUpdatedEvent |
		*goastorage.AIPWorkflowCreatedEvent |
		*goastorage.AIPWorkflowUpdatedEvent |
		*goastorage.AIPTaskCreatedEvent |
		*goastorage.AIPTaskUpdatedEvent |
		*goastorage.AIPDeletionRequestCreatedEvent |
		*goastorage.AIPDeletionRequestUpdatedEvent
}

// PublishEvent publishes a storage event with type safety.
func PublishEvent[E Event](ctx context.Context, svc event.Service[*goastorage.StorageEvent], event E) {
	svc.PublishEvent(ctx, &goastorage.StorageEvent{StorageValue: event})
}

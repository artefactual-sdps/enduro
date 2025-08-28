package storage_test

import (
	"testing"

	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/storage"
)

func TestPublishEventTypes(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewServiceInMem[*goastorage.StorageEvent]()

	// Test all event types can be published.
	storage.PublishEvent(ctx, svc, &goastorage.StoragePingEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.LocationCreatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPCreatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPStatusUpdatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPLocationUpdatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPWorkflowCreatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPWorkflowUpdatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPTaskCreatedEvent{})
	storage.PublishEvent(ctx, svc, &goastorage.AIPTaskUpdatedEvent{})
}

func TestEventSerializer(t *testing.T) {
	t.Parallel()

	serializer := &storage.EventSerializer{}
	originalEvent := &goastorage.StorageEvent{
		StorageValue: &goastorage.StoragePingEvent{Message: ref.New("test")},
	}

	data, err := serializer.Marshal(originalEvent)
	assert.NilError(t, err)
	assert.Assert(t, len(data) > 0)

	deserializedEvent, err := serializer.Unmarshal(data)
	assert.NilError(t, err)
	assert.DeepEqual(t, deserializedEvent, originalEvent)
}

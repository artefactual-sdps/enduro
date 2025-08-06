package ingest_test

import (
	"testing"

	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

func TestPublishEventTypes(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewServiceInMem[*goaingest.IngestEvent]()

	// Test all event types can be published.
	ingest.PublishEvent(ctx, svc, &goaingest.IngestPingEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPCreatedEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPUpdatedEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPStatusUpdatedEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPWorkflowCreatedEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPWorkflowUpdatedEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPTaskCreatedEvent{})
	ingest.PublishEvent(ctx, svc, &goaingest.SIPTaskUpdatedEvent{})
}

func TestEventSerializer(t *testing.T) {
	t.Parallel()

	serializer := &ingest.EventSerializer{}
	originalEvent := &goaingest.IngestEvent{
		IngestValue: &goaingest.IngestPingEvent{Message: ref.New("test")},
	}

	data, err := serializer.Marshal(originalEvent)
	assert.NilError(t, err)
	assert.Assert(t, len(data) > 0)

	deserializedEvent, err := serializer.Unmarshal(data)
	assert.NilError(t, err)
	assert.DeepEqual(t, deserializedEvent, originalEvent)
}

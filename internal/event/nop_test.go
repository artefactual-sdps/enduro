package event_test

import (
	"testing"

	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestIngestEventServiceNopPublishEvent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewIngestEventServiceNop()

	svc.PublishEvent(ctx, &goaingest.IngestEvent{})
}

func TestStorageEventServiceNopPublishEvent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewStorageEventServiceNop()

	svc.PublishEvent(ctx, &goastorage.StorageEvent{})
}

func TestIngestEventServiceNopSubscribe(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewIngestEventServiceNop()

	_, err := svc.Subscribe(ctx)
	assert.ErrorContains(t, err, "Subscribe not supported by nop service")
}

func TestStorageEventServiceNopSubscribe(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewStorageEventServiceNop()

	_, err := svc.Subscribe(ctx)
	assert.ErrorContains(t, err, "Subscribe not supported by nop service")
}

func TestIngestEventServiceNopWithPublishHelper(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewIngestEventServiceNop()

	event.PublishIngestEvent(ctx, svc, &goaingest.IngestPingEvent{})
}

func TestStorageEventServiceNopWithPublishHelper(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewStorageEventServiceNop()

	event.PublishStorageEvent(ctx, svc, &goastorage.StoragePingEvent{})
}

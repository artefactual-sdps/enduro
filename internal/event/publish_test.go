package event_test

import (
	"testing"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestPublishIngestEventTypes(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewIngestEventServiceInMem()

	for _, tt := range []struct {
		name   string
		event  any
		panics bool
	}{
		{"IngestPingEvent", &goaingest.IngestPingEvent{}, false},
		{"SIPCreatedEvent", &goaingest.SIPCreatedEvent{}, false},
		{"SIPUpdatedEvent", &goaingest.SIPUpdatedEvent{}, false},
		{"SIPStatusUpdatedEvent", &goaingest.SIPStatusUpdatedEvent{}, false},
		{"SIPWorkflowCreatedEvent", &goaingest.SIPWorkflowCreatedEvent{}, false},
		{"SIPWorkflowUpdatedEvent", &goaingest.SIPWorkflowUpdatedEvent{}, false},
		{"SIPTaskCreatedEvent", &goaingest.SIPTaskCreatedEvent{}, false},
		{"SIPTaskUpdatedEvent", &goaingest.SIPTaskUpdatedEvent{}, false},
		{"UnexpectedEvent", &goastorage.StoragePingEvent{}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Fatal("Expected panic but none occurred")
					}
				}()
			}

			event.PublishIngestEvent(ctx, svc, tt.event)
		})
	}
}

func TestPublishStorageEventTypes(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := event.NewStorageEventServiceInMem()

	for _, tt := range []struct {
		name   string
		event  any
		panics bool
	}{
		{"StoragePingEvent", &goastorage.StoragePingEvent{}, false},
		{"LocationCreatedEvent", &goastorage.LocationCreatedEvent{}, false},
		{"AIPCreatedEvent", &goastorage.AIPCreatedEvent{}, false},
		{"AIPStatusUpdatedEvent", &goastorage.AIPStatusUpdatedEvent{}, false},
		{"AIPLocationUpdatedEvent", &goastorage.AIPLocationUpdatedEvent{}, false},
		{"AIPWorkflowCreatedEvent", &goastorage.AIPWorkflowCreatedEvent{}, false},
		{"AIPWorkflowUpdatedEvent", &goastorage.AIPWorkflowUpdatedEvent{}, false},
		{"AIPTaskCreatedEvent", &goastorage.AIPTaskCreatedEvent{}, false},
		{"AIPTaskUpdatedEvent", &goastorage.AIPTaskUpdatedEvent{}, false},
		{"UnexpectedEvent", &goastorage.IngestPingEvent{}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Fatal("Expected panic but none occurred")
					}
				}()
			}

			event.PublishStorageEvent(ctx, svc, tt.event)
		})
	}
}

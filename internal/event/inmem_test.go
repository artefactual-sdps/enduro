package event_test

import (
	"testing"

	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestIngestEventService(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		ctx := t.Context()
		s := event.NewIngestEventServiceInMem()

		subA, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		subB, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Publish event to both users.
		s.PublishEvent(ctx, &goaingest.IngestEvent{})

		// Verify both subscribers received the update.
		select {
		case <-subA.C():
		default:
			t.Fatal("expected an event")
		}

		select {
		case <-subB.C():
		default:
			t.Fatal("expected an event")
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		ctx := t.Context()
		s := event.NewIngestEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := sub.Close(); err != nil {
			t.Fatal(err)
		}

		// Publish event after unsubscribe.
		s.PublishEvent(ctx, &goaingest.IngestEvent{})

		// Verify subscriber did not receive the update.
		select {
		case _, ok := <-sub.C():
			if ok {
				t.Fatal("unexpected event")
			}
		default:
			t.Fatal("expected channel to be closed")
		}
	})
}

func TestStorageEventService(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		ctx := t.Context()
		s := event.NewStorageEventServiceInMem()

		subA, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		subB, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Publish event to both users.
		s.PublishEvent(ctx, &goastorage.StorageEvent{})

		// Verify both subscribers received the update.
		select {
		case <-subA.C():
		default:
			t.Fatal("expected an event")
		}

		select {
		case <-subB.C():
		default:
			t.Fatal("expected an event")
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		ctx := t.Context()
		s := event.NewStorageEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := sub.Close(); err != nil {
			t.Fatal(err)
		}

		// Publish event after unsubscribe.
		s.PublishEvent(ctx, &goastorage.StorageEvent{})

		// Verify subscriber did not receive the update.
		select {
		case _, ok := <-sub.C():
			if ok {
				t.Fatal("unexpected event")
			}
		default:
			t.Fatal("expected channel to be closed")
		}
	})
}

func TestPublishHelpers(t *testing.T) {
	t.Run("PublishIngestEvent", func(t *testing.T) {
		ctx := t.Context()
		s := event.NewIngestEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		event.PublishIngestEvent(ctx, s, &goaingest.SIPCreatedEvent{})

		// Verify subscriber received the event.
		select {
		case event := <-sub.C():
			if event.IngestValue == nil {
				t.Fatal("expected event to contain data")
			}
		default:
			t.Fatal("expected an event")
		}
	})

	t.Run("PublishStorageEvent", func(t *testing.T) {
		ctx := t.Context()
		s := event.NewStorageEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		event.PublishStorageEvent(ctx, s, &goastorage.AIPCreatedEvent{})

		// Verify subscriber received the event.
		select {
		case event := <-sub.C():
			if event.StorageValue == nil {
				t.Fatal("expected event to contain data")
			}
		default:
			t.Fatal("expected event")
		}
	})
}

func TestInMemEventBufferOverflow(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := event.NewIngestEventServiceInMem()

	sub, err := s.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	// Fill the buffer beyond its capacity without reading from it.
	// This should not cause the subscription to hang or disconnect.
	for range event.EventBufferSize + 5 {
		event.PublishIngestEvent(ctx, s, &goaingest.IngestPingEvent{})
	}

	// Verify subscriber is still connected by checking that we can
	// read events from the buffer (it should contain EventBufferSize events).
	eventsReceived := 0
loop:
	for range event.EventBufferSize {
		select {
		case <-sub.C():
			eventsReceived++
		default:
			break loop
		}
	}

	if eventsReceived != event.EventBufferSize {
		t.Fatalf("expected %d events in buffer, got %d", event.EventBufferSize, eventsReceived)
	}

	// Verify subscriber can still receive new events after buffer overflow.
	event.PublishIngestEvent(ctx, s, &goaingest.IngestPingEvent{})

	select {
	case <-sub.C():
		// Successfully received new event after overflow.
	default:
		t.Fatal("subscriber should still be able to receive events after buffer overflow")
	}
}

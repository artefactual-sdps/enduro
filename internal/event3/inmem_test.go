package event3_test

import (
	"context"
	"testing"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event3"
)

func TestIngestEventService(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		ctx := context.Background()
		s := event3.NewIngestEventServiceInMem()

		subA, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		subB, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Publish event to both users
		s.PublishEvent(ctx, &goaingest.IngestEvent{})

		// Verify both subscribers received the update.
		select {
		case <-subA.C():
		default:
			t.Fatal("expected event")
		}

		select {
		case <-subB.C():
		default:
			t.Fatal("expected event")
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		ctx := context.Background()
		s := event3.NewIngestEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := sub.Close(); err != nil {
			t.Fatal(err)
		}

		// Publish event after unsubscribe
		s.PublishEvent(ctx, &goaingest.IngestEvent{})

		// Verify subscriber did not receive the update (channel should be closed).
		select {
		case _, ok := <-sub.C():
			if ok {
				t.Fatal("unexpected event")
			}
			// Channel is closed, which is expected
		default:
			t.Fatal("expected channel to be closed")
		}
	})
}

func TestStorageEventService(t *testing.T) {
	t.Run("Subscribe", func(t *testing.T) {
		ctx := context.Background()
		s := event3.NewStorageEventServiceInMem()

		subA, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		subB, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Publish event to both users
		s.PublishEvent(ctx, &goastorage.StorageEvent{})

		// Verify both subscribers received the update.
		select {
		case <-subA.C():
		default:
			t.Fatal("expected event")
		}

		select {
		case <-subB.C():
		default:
			t.Fatal("expected event")
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		ctx := context.Background()
		s := event3.NewStorageEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := sub.Close(); err != nil {
			t.Fatal(err)
		}

		// Publish event after unsubscribe
		s.PublishEvent(ctx, &goastorage.StorageEvent{})

		// Verify subscriber did not receive the update (channel should be closed).
		select {
		case _, ok := <-sub.C():
			if ok {
				t.Fatal("unexpected event")
			}
			// Channel is closed, which is expected
		default:
			t.Fatal("expected channel to be closed")
		}
	})
}

func TestPublishHelpers(t *testing.T) {
	t.Run("PublishIngestEvent", func(t *testing.T) {
		ctx := context.Background()
		s := event3.NewIngestEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Test publishing different event types
		event3.PublishIngestEvent(ctx, s, &goaingest.SIPCreatedEvent{})

		// Verify subscriber received the event
		select {
		case event := <-sub.C():
			if event.IngestValue == nil {
				t.Fatal("expected event to contain data")
			}
		default:
			t.Fatal("expected event")
		}
	})

	t.Run("PublishStorageEvent", func(t *testing.T) {
		ctx := context.Background()
		s := event3.NewStorageEventServiceInMem()

		sub, err := s.Subscribe(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Test publishing different event types
		event3.PublishStorageEvent(ctx, s, &goastorage.AIPCreatedEvent{})

		// Verify subscriber received the event
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

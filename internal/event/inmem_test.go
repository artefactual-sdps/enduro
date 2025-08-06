package event

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestInMemEventService(t *testing.T) {
	t.Parallel()

	t.Run("Subscribe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := NewServiceInMem[string]()

		subA, err := s.Subscribe(ctx)
		assert.NilError(t, err)
		defer subA.Close()

		subB, err := s.Subscribe(ctx)
		assert.NilError(t, err)
		defer subB.Close()

		// Publish event to both subscribers.
		event := "test-event"
		s.PublishEvent(ctx, event)

		// Both should receive the event.
		select {
		case e := <-subA.C():
			assert.Equal(t, e, event)
		default:
			t.Fatal("expected event from subA")
		}

		select {
		case e := <-subB.C():
			assert.Equal(t, e, event)
		default:
			t.Fatal("expected event from subB")
		}
	})

	t.Run("BufferOverflow", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := NewServiceInMem[int]()

		sub, err := s.Subscribe(ctx)
		assert.NilError(t, err)
		defer sub.Close()

		// Publish more events than the buffer can hold.
		for i := range EventBufferSize + 10 {
			s.PublishEvent(ctx, i)
		}

		// Should receive some events, but not all due to buffer overflow.
		received := 0
	loop:
		for {
			select {
			case <-sub.C():
				received++
			default:
				break loop
			}
		}

		assert.Equal(t, received, EventBufferSize)

		// Now that we've read some events, publish more events to verify
		// the subscription continues to work after buffer overflow.
		newEvents := []int{1000, 1001, 1002}
		for _, event := range newEvents {
			s.PublishEvent(ctx, event)
		}

		// Should receive the new events.
		newReceived := 0
	newLoop:
		for {
			select {
			case <-sub.C():
				newReceived++
			default:
				break newLoop
			}
		}

		assert.Equal(t, newReceived, len(newEvents))
	})
}

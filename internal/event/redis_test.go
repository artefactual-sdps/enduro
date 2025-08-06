package event_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-logr/logr"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/event"
)

type TestEvent struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type testEventSerializer struct{}

var _ event.Serializer[*TestEvent] = (*testEventSerializer)(nil)

func (s *testEventSerializer) Marshal(e *TestEvent) ([]byte, error) {
	return json.Marshal(e)
}

func (s *testEventSerializer) Unmarshal(data []byte) (*TestEvent, error) {
	var e TestEvent
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func TestRedisServicePublish(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "test-events"

	// Set up Redis client to receive messages.
	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	sub := c.Subscribe(ctx, channel)
	t.Cleanup(func() {
		sub.Close()
	})

	// Call Receive to force the connection to wait a response from
	// Redis so the subscription is active immediately.
	_, err := sub.Receive(ctx)
	assert.NilError(t, err)

	input := make(chan *redis.Message)

	go func() {
		ch := sub.Channel()
		for message := range ch {
			input <- message
			break
		}
	}()

	// Create Redis event service.
	svc, err := event.NewServiceRedis(
		logr.Discard(),
		noop.NewTracerProvider(),
		"redis://"+s.Addr(),
		channel,
		&testEventSerializer{},
	)
	assert.NilError(t, err)

	// Publish test event.
	svc.PublishEvent(ctx, &TestEvent{ID: "test-123", Message: "hello world"})

	// Verify the message was published correctly.
	select {
	case ret := <-input:
		assert.Equal(t, ret.Payload, `{"id":"test-123","message":"hello world"}`)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for published message")
	}
}

func TestRedisServiceSubscribe(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "test-events"

	// Create Redis event service.
	svc, err := event.NewServiceRedis(
		logr.Discard(),
		noop.NewTracerProvider(),
		"redis://"+s.Addr(),
		channel,
		&testEventSerializer{},
	)
	assert.NilError(t, err)

	// Subscribe to events.
	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	// Publish message directly to Redis.
	c := redis.NewClient(&redis.Options{Addr: s.Addr()})

	testEvent := &TestEvent{ID: "test-456", Message: "test message"}
	blob, err := json.Marshal(testEvent)
	assert.NilError(t, err)

	err = c.Publish(ctx, channel, blob).Err()
	assert.NilError(t, err)

	// Verify subscription receives the event.
	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, testEvent)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for subscribed message")
	}
}

func TestRedisServiceSubscribeInvalidMessage(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "test-events"

	// Create Redis event service.
	svc, err := event.NewServiceRedis(
		logr.Discard(),
		noop.NewTracerProvider(),
		"redis://"+s.Addr(),
		channel,
		&testEventSerializer{},
	)
	assert.NilError(t, err)

	// Subscribe to events.
	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	// Publish invalid event to Redis.
	c := redis.NewClient(&redis.Options{Addr: s.Addr()})
	err = c.Publish(ctx, channel, "invalid event").Err()
	assert.NilError(t, err)

	// Publish valid event after invalid one.
	testEvent := &TestEvent{ID: "test-789", Message: "valid message"}
	blob, err := json.Marshal(testEvent)
	assert.NilError(t, err)

	err = c.Publish(ctx, channel, blob).Err()
	assert.NilError(t, err)

	// Should receive the valid message.
	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, testEvent)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for valid message after invalid one")
	}
}

func TestRedisServiceBufferOverflow(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "test-events"

	// Create Redis event service.
	svc, err := event.NewServiceRedis(
		logr.Discard(),
		noop.NewTracerProvider(),
		"redis://"+s.Addr(),
		channel,
		&testEventSerializer{},
	)
	assert.NilError(t, err)

	// Subscribe to events.
	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	// Fill the buffer beyond its capacity without reading from it.
	// This should not cause the subscription to hang or disconnect.
	for range event.EventBufferSize + 10 {
		testEvent := &TestEvent{
			ID:      "overflow-test",
			Message: "buffer overflow test",
		}
		svc.PublishEvent(ctx, testEvent)
	}

	// Give some time for events to be processed.
	time.Sleep(100 * time.Millisecond)

	// Verify subscriber is still connected by reading available events from the buffer.
	// Due to Redis async nature, we may not get exactly EventBufferSize events,
	// but we should get a substantial number if the subscriber remained connected.
	eventsReceived := 0
loop:
	for eventsReceived < event.EventBufferSize {
		select {
		case <-sub.C():
			eventsReceived++
		case <-time.After(2 * time.Second):
			break loop
		default:
			break loop
		}
	}

	// We should have received a significant number of events (at least half the buffer)
	// to prove the subscriber remained connected during overflow.
	minExpected := event.EventBufferSize / 2
	if eventsReceived < minExpected {
		t.Fatalf("expected at least %d events in buffer, got %d", minExpected, eventsReceived)
	}

	// Verify subscriber can still receive new events after buffer overflow.
	newEvent := &TestEvent{
		ID:      "post-overflow",
		Message: "post overflow test",
	}
	svc.PublishEvent(ctx, newEvent)

	// Wait for the new event.
	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, newEvent)
	case <-time.After(5 * time.Second):
		t.Fatal("subscriber should still be able to receive events after buffer overflow")
	}
}

func TestRedisServiceSubscriptionClosure(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "test-events"

	// Create Redis event service.
	svc, err := event.NewServiceRedis(
		logr.Discard(),
		noop.NewTracerProvider(),
		"redis://"+s.Addr(),
		channel,
		&testEventSerializer{},
	)
	assert.NilError(t, err)

	// Subscribe to events.
	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)

	// Verify subscription is active.
	testEvent := &TestEvent{
		ID:      "closure-test",
		Message: "before close",
	}
	svc.PublishEvent(ctx, testEvent)

	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, testEvent)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message before close")
	}

	// Close subscription.
	err = sub.Close()
	assert.NilError(t, err)

	// Verify we don't receive any more events after closing.
	svc.PublishEvent(ctx, &TestEvent{ID: "after-close", Message: "should not receive"})

	// Wait a bit to ensure the message would have been received.
	time.Sleep(100 * time.Millisecond)

	// Try to read from the channel with a timeout.
	select {
	case <-sub.C():
		t.Fatal("should not receive events after closing subscription")
	case <-time.After(1 * time.Second):
		// This is expected - no events should be received after closing.
	}
}

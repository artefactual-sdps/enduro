package event_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-logr/logr/testr"
	"github.com/redis/go-redis/v9"
	"go.artefactual.dev/tools/ref"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/gen/http/package_/server"
	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/event"
)

// Confirm that PublishEvent is streaming events via Redis.
func TestEventServiceRedisPublish(t *testing.T) {
	t.Parallel()

	const channel = "enduro-events"

	ctx := context.Background()
	s := miniredis.RunT(t)

	c := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
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

	cfg := event.Config{
		RedisAddress: "redis://" + s.Addr(),
		RedisChannel: channel,
	}
	svc, err := event.NewEventServiceRedis(testr.New(t), noop.NewTracerProvider(), &cfg)
	assert.NilError(t, err)

	svc.PublishEvent(ctx, &goapackage.MonitorEvent{
		Event: &goapackage.MonitorPingEvent{
			Message: ref.New("hello"),
		},
	})

	select {
	case ret := <-input:
		assert.DeepEqual(t, ret.Payload, `{"event":{"Type":"monitor_ping_event","Value":"{\"Message\":\"hello\"}"}}`)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout!")
	}
}

// Confirm that Subscribe is capturing events received via Redis.
func TestEventServiceRedisSubscribe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := miniredis.RunT(t)

	cfg := event.Config{
		RedisAddress: "redis://" + s.Addr(),
		RedisChannel: "enduro-events",
	}
	svc, err := event.NewEventServiceRedis(testr.New(t), noop.NewTracerProvider(), &cfg)
	assert.NilError(t, err)

	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	c := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	ev := goapackage.MonitorEvent{
		Event: &goapackage.MonitorPingEvent{
			Message: ref.New("hello"),
		},
	}
	blob, err := json.Marshal(server.NewMonitorResponseBody(&ev))
	assert.NilError(t, err)

	err = c.Publish(ctx, cfg.RedisChannel, blob).Err()
	assert.NilError(t, err)

	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, &ev)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout!")
	}
}

package event_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-logr/logr/testr"
	"github.com/redis/go-redis/v9"
	"go.artefactual.dev/tools/ref"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/server"
	storageserver "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestIngestEventServiceRedisPublish(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "enduro-ingest-events"

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
		RedisAddress:       "redis://" + s.Addr(),
		IngestRedisChannel: channel,
	}
	svc, err := event.NewIngestEventServiceRedis(testr.New(t), noop.NewTracerProvider(), &cfg)
	assert.NilError(t, err)

	event.PublishIngestEvent(ctx, svc, &goaingest.IngestPingEvent{
		Message: ref.New("hello"),
	})

	select {
	case ret := <-input:
		assert.DeepEqual(
			t,
			ret.Payload,
			`{"ingest_value":{"Type":"ingest_ping_event","Value":"{\"Message\":\"hello\"}"}}`,
		)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout!")
	}
}

func TestIngestEventServiceRedisSubscribe(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)

	cfg := event.Config{
		RedisAddress:       "redis://" + s.Addr(),
		IngestRedisChannel: "enduro-ingest-events",
	}
	svc, err := event.NewIngestEventServiceRedis(testr.New(t), noop.NewTracerProvider(), &cfg)
	assert.NilError(t, err)

	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	c := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	ev := goaingest.IngestEvent{
		IngestValue: &goaingest.IngestPingEvent{
			Message: ref.New("hello"),
		},
	}
	blob, err := json.Marshal(server.NewMonitorResponseBody(&ev))
	assert.NilError(t, err)

	err = c.Publish(ctx, cfg.IngestRedisChannel, blob).Err()
	assert.NilError(t, err)

	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, &ev)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout!")
	}
}

func TestStorageEventServiceRedisPublish(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)
	channel := "enduro-storage-events"

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
		RedisAddress:        "redis://" + s.Addr(),
		StorageRedisChannel: channel,
	}
	svc, err := event.NewStorageEventServiceRedis(testr.New(t), noop.NewTracerProvider(), &cfg)
	assert.NilError(t, err)

	event.PublishStorageEvent(ctx, svc, &goastorage.StoragePingEvent{
		Message: ref.New("hello"),
	})

	select {
	case ret := <-input:
		assert.DeepEqual(
			t,
			ret.Payload,
			`{"storage_value":{"Type":"storage_ping_event","Value":"{\"Message\":\"hello\"}"}}`,
		)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout!")
	}
}

func TestStorageEventServiceRedisSubscribe(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	s := miniredis.RunT(t)

	cfg := event.Config{
		RedisAddress:        "redis://" + s.Addr(),
		StorageRedisChannel: "enduro-storage-events",
	}
	svc, err := event.NewStorageEventServiceRedis(testr.New(t), noop.NewTracerProvider(), &cfg)
	assert.NilError(t, err)

	sub, err := svc.Subscribe(ctx)
	assert.NilError(t, err)
	t.Cleanup(func() {
		sub.Close()
	})

	c := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	ev := goastorage.StorageEvent{
		StorageValue: &goastorage.StoragePingEvent{
			Message: ref.New("hello"),
		},
	}

	blob, err := json.Marshal(storageserver.NewMonitorResponseBody(&ev))
	assert.NilError(t, err)

	err = c.Publish(ctx, cfg.StorageRedisChannel, blob).Err()
	assert.NilError(t, err)

	select {
	case ret := <-sub.C():
		assert.DeepEqual(t, ret, &ev)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout!")
	}
}

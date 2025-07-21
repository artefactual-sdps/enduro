package event3

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"

	"github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/client"
	"github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/server"
	storageclient "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	storageserver "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// EventSerializer handles serialization/deserialization of events for Redis
type EventSerializer[T any] interface {
	Marshal(event T) ([]byte, error)
	Unmarshal(data []byte) (T, error)
}

// IngestEventSerializer handles ingest events
type IngestEventSerializer struct{}

func (s *IngestEventSerializer) Marshal(event *goaingest.IngestMonitorEvent) ([]byte, error) {
	return json.Marshal(server.NewMonitorResponseBody(event))
}

func (s *IngestEventSerializer) Unmarshal(data []byte) (*goaingest.IngestMonitorEvent, error) {
	payload := client.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := client.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return client.NewMonitorIngestMonitorEventOK(&payload), nil
}

// StorageEventSerializer handles storage events
type StorageEventSerializer struct{}

func (s *StorageEventSerializer) Marshal(event *goastorage.StorageMonitorEvent) ([]byte, error) {
	return json.Marshal(storageserver.NewMonitorResponseBody(event))
}

func (s *StorageEventSerializer) Unmarshal(data []byte) (*goastorage.StorageMonitorEvent, error) {
	payload := storageclient.MonitorResponseBody{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if err := storageclient.ValidateMonitorResponseBody(&payload); err != nil {
		return nil, err
	}
	return storageclient.NewMonitorStorageMonitorEventOK(&payload), nil
}

// ServiceRedisImpl represents a generic Redis-based service for managing events.
type ServiceRedisImpl[T any] struct {
	logger     logr.Logger
	client     redis.UniversalClient
	cfg        *Config
	serializer EventSerializer[T]
}

var (
	_ Service[*goaingest.IngestMonitorEvent]   = (*ServiceRedisImpl[*goaingest.IngestMonitorEvent])(nil)
	_ Service[*goastorage.StorageMonitorEvent] = (*ServiceRedisImpl[*goastorage.StorageMonitorEvent])(nil)
)

// NewServiceRedis returns a new instance of a generic Redis event service.
func NewServiceRedis[T any](
	logger logr.Logger,
	tp trace.TracerProvider,
	cfg *Config,
	serializer EventSerializer[T],
) (Service[T], error) {
	opts, err := redis.ParseURL(cfg.RedisAddress)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	if err := redisotel.InstrumentTracing(
		client,
		redisotel.WithTracerProvider(tp),
		redisotel.WithDBStatement(false),
	); err != nil {
		return nil, fmt.Errorf("instrument redis client tracing: %v", err)
	}

	return &ServiceRedisImpl[T]{
		logger:     logger,
		client:     client,
		cfg:        cfg,
		serializer: serializer,
	}, nil
}

func (s *ServiceRedisImpl[T]) PublishEvent(ctx context.Context, event T) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	blob, err := s.serializer.Marshal(event)
	if err != nil {
		s.logger.Error(err, "Error encoding event.")
		return
	}

	if err := s.client.Publish(ctx, s.cfg.RedisChannel, blob).Err(); err != nil {
		s.logger.Error(err, "Error publishing event.")
	}
}

func (s *ServiceRedisImpl[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	sub := NewSubscriptionRedis(ctx, s.logger, s.client, s.cfg.RedisChannel, s.serializer)
	return sub, nil
}

// SubscriptionRedisImpl represents a stream of events.
type SubscriptionRedisImpl[T any] struct {
	logger     logr.Logger
	pubsub     *redis.PubSub
	c          chan T
	stopCh     chan struct{}
	serializer EventSerializer[T]
}

var (
	_ Subscription[*goaingest.IngestMonitorEvent]   = (*SubscriptionRedisImpl[*goaingest.IngestMonitorEvent])(nil)
	_ Subscription[*goastorage.StorageMonitorEvent] = (*SubscriptionRedisImpl[*goastorage.StorageMonitorEvent])(nil)
)

func NewSubscriptionRedis[T any](
	ctx context.Context,
	logger logr.Logger,
	c redis.UniversalClient,
	channel string,
	serializer EventSerializer[T],
) Subscription[T] {
	pubsub := c.Subscribe(ctx, channel)
	_, _ = pubsub.Receive(ctx)
	sub := SubscriptionRedisImpl[T]{
		logger:     logger,
		pubsub:     pubsub,
		c:          make(chan T, EventBufferSize),
		stopCh:     make(chan struct{}),
		serializer: serializer,
	}
	go sub.loop()
	return &sub
}

func (s *SubscriptionRedisImpl[T]) loop() {
	ch := s.pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			event, err := s.serializer.Unmarshal([]byte(msg.Payload))
			if err != nil {
				s.logger.Error(err, "Error decoding event.")
				continue
			}
			s.c <- event
		case <-s.stopCh:
			return
		}
	}
}

// Close disconnects the subscription from the service it was created from.
func (s *SubscriptionRedisImpl[T]) Close() error {
	s.stopCh <- struct{}{}
	return s.pubsub.Close()
}

// C returns a receive-only channel of events.
func (s *SubscriptionRedisImpl[T]) C() <-chan T {
	return s.c
}

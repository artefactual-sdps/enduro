package event

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// ServiceRedisImpl represents a generic Redis-based service for managing events.
type ServiceRedisImpl[T any] struct {
	logger     logr.Logger
	client     redis.UniversalClient
	channel    string
	serializer EventSerializer[T]
}

var (
	_ Service[*goaingest.IngestEvent]   = (*ServiceRedisImpl[*goaingest.IngestEvent])(nil)
	_ Service[*goastorage.StorageEvent] = (*ServiceRedisImpl[*goastorage.StorageEvent])(nil)
)

// NewServiceRedis returns a new instance of a generic Redis event service.
func NewServiceRedis[T any](
	logger logr.Logger,
	tp trace.TracerProvider,
	address string,
	channel string,
	serializer EventSerializer[T],
) (Service[T], error) {
	opts, err := redis.ParseURL(address)
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
		channel:    channel,
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

	if err := s.client.Publish(ctx, s.channel, blob).Err(); err != nil {
		s.logger.Error(err, "Error publishing event.", "channel", s.channel)
	}
}

func (s *ServiceRedisImpl[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	sub := NewSubscriptionRedis(ctx, s.logger, s.client, s.channel, s.serializer)
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
	_ Subscription[*goaingest.IngestEvent]   = (*SubscriptionRedisImpl[*goaingest.IngestEvent])(nil)
	_ Subscription[*goastorage.StorageEvent] = (*SubscriptionRedisImpl[*goastorage.StorageEvent])(nil)
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
		case msg, ok := <-ch:
			if !ok || msg == nil {
				continue
			}
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

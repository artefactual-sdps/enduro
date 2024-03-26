package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"

	"github.com/artefactual-sdps/enduro/internal/api/gen/http/package_/client"
	"github.com/artefactual-sdps/enduro/internal/api/gen/http/package_/server"
	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
)

type EventServiceRedisImpl struct {
	logger logr.Logger
	client redis.UniversalClient
	cfg    *Config
}

var _ EventService = (*EventServiceRedisImpl)(nil)

func NewEventServiceRedis(logger logr.Logger, tp trace.TracerProvider, cfg *Config) (EventService, error) {
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

	return &EventServiceRedisImpl{
		logger: logger,
		client: client,
		cfg:    cfg,
	}, nil
}

func (s *EventServiceRedisImpl) PublishEvent(ctx context.Context, event *goapackage.MonitorEvent) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	blob, err := json.Marshal(server.NewMonitorResponseBody(event))
	if err != nil {
		s.logger.Error(err, "Error encoding monitor event.")
	}

	if err := s.client.Publish(ctx, s.cfg.RedisChannel, blob).Err(); err != nil {
		s.logger.Error(err, "Error publishing monitor event.")
	}
}

func (s *EventServiceRedisImpl) Subscribe(ctx context.Context) (Subscription, error) {
	sub := NewSubscriptionRedis(ctx, s.logger, s.client, s.cfg.RedisChannel)

	return sub, nil
}

// SubscriptionRedisImpl represents a stream of user-related events.
type SubscriptionRedisImpl struct {
	logger logr.Logger
	pubsub *redis.PubSub
	c      chan *goapackage.MonitorEvent // channel of events
	stopCh chan struct{}
}

var _ Subscription = (*SubscriptionRedisImpl)(nil)

func NewSubscriptionRedis(
	ctx context.Context,
	logger logr.Logger,
	c redis.UniversalClient,
	channel string,
) Subscription {
	pubsub := c.Subscribe(ctx, channel)
	// Call Receive to force the connection to wait a response from
	// Redis so the subscription is active immediately.
	_, _ = pubsub.Receive(ctx)
	sub := SubscriptionRedisImpl{
		logger: logger,
		pubsub: pubsub,
		c:      make(chan *goapackage.MonitorEvent, EventBufferSize),
		stopCh: make(chan struct{}),
	}
	go sub.loop()
	return &sub
}

func (s *SubscriptionRedisImpl) loop() {
	ch := s.pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			payload := client.MonitorResponseBody{}
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				s.logger.Error(err, "Error decoding monitor event.")
				continue
			}
			if err := client.ValidateMonitorResponseBody(&payload); err != nil {
				s.logger.Error(err, "Monitor event is invalid.")
				continue
			}
			s.c <- client.NewMonitorEventOK(&payload)
		case <-s.stopCh:
			return
		}
	}
}

// Close disconnects the subscription from the service it was created from.
func (s *SubscriptionRedisImpl) Close() error {
	s.stopCh <- struct{}{}

	return s.pubsub.Close()
}

// C returns a receive-only channel of user-related events.
func (s *SubscriptionRedisImpl) C() <-chan *goapackage.MonitorEvent {
	return s.c
}

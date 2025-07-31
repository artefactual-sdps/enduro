package event

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

// serviceRedisImpl represents a generic Redis-based service for managing events.
type serviceRedisImpl[T any] struct {
	logger     logr.Logger
	client     redis.UniversalClient
	channel    string
	serializer eventSerializer[T]
}

// newServiceRedis returns a new instance of a generic Redis event service.
func newServiceRedis[T any](
	logger logr.Logger,
	tp trace.TracerProvider,
	address string,
	channel string,
	serializer eventSerializer[T],
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

	return &serviceRedisImpl[T]{
		logger:     logger,
		client:     client,
		channel:    channel,
		serializer: serializer,
	}, nil
}

func (s *serviceRedisImpl[T]) PublishEvent(ctx context.Context, event T) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	blob, err := s.serializer.marshal(event)
	if err != nil {
		s.logger.Error(err, "Error encoding event.")
		return
	}

	if err := s.client.Publish(ctx, s.channel, blob).Err(); err != nil {
		s.logger.Error(err, "Error publishing event.", "channel", s.channel)
	}
}

func (s *serviceRedisImpl[T]) Subscribe(ctx context.Context) (Subscription[T], error) {
	sub := newSubscriptionRedis(ctx, s.logger, s.client, s.channel, s.serializer)
	return sub, nil
}

// subscriptionRedisImpl represents a stream of events.
type subscriptionRedisImpl[T any] struct {
	logger     logr.Logger
	pubsub     *redis.PubSub
	c          chan T
	stopCh     chan struct{}
	serializer eventSerializer[T]
}

func newSubscriptionRedis[T any](
	ctx context.Context,
	logger logr.Logger,
	c redis.UniversalClient,
	channel string,
	serializer eventSerializer[T],
) Subscription[T] {
	pubsub := c.Subscribe(ctx, channel)
	_, _ = pubsub.Receive(ctx)
	sub := subscriptionRedisImpl[T]{
		logger:     logger,
		pubsub:     pubsub,
		c:          make(chan T, EventBufferSize),
		stopCh:     make(chan struct{}, 1),
		serializer: serializer,
	}
	go sub.loop(ctx)
	return &sub
}

func (s *subscriptionRedisImpl[T]) loop(ctx context.Context) {
	ch := s.pubsub.Channel()

	for {
		select {
		case msg, ok := <-ch:
			if !ok || msg == nil {
				continue
			}
			event, err := s.serializer.unmarshal([]byte(msg.Payload))
			if err != nil {
				s.logger.Error(err, "Error decoding event.")
				continue
			}
			// Non-blocking send to avoid deadlock. If buffer is full, skip the event.
			// Also check for stop signal to ensure responsive shutdown.
			select {
			case s.c <- event:
				// Event successfully sent to subscriber.
			case <-s.stopCh:
				return
			default:
				// Skip this event if the subscriber's buffer is full.
				// This prevents blocking the loop and allows the subscription
				// to continue receiving future events and respond to Close().
			}
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Close disconnects the subscription from the service it was created from.
func (s *subscriptionRedisImpl[T]) Close() error {
	s.stopCh <- struct{}{}
	return s.pubsub.Close()
}

// C returns a receive-only channel of events.
func (s *subscriptionRedisImpl[T]) C() <-chan T {
	return s.c
}

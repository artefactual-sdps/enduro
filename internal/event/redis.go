package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
)

type EventServiceRedisImpl struct {
	client redis.UniversalClient
	cfg    *Config
}

var _ EventService = (*EventServiceRedisImpl)(nil)

func NewEventServiceRedis(cfg *Config) (EventService, error) {
	opts, err := redis.ParseURL(cfg.RedisAddress)
	if err != nil {
		return nil, err
	}
	return &EventServiceRedisImpl{
		client: redis.NewClient(opts),
		cfg:    cfg,
	}, nil
}

func (s *EventServiceRedisImpl) PublishEvent(event *goapackage.EnduroMonitorEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	blob, _ := json.Marshal(event)
	_ = s.client.Publish(ctx, s.cfg.RedisChannel, blob).Err()
}

func (s *EventServiceRedisImpl) Subscribe(ctx context.Context) (Subscription, error) {
	sub := NewSubscriptionRedis(s.client, s.cfg.RedisChannel)

	return sub, nil
}

// SubscriptionRedisImpl represents a stream of user-related events.
type SubscriptionRedisImpl struct {
	pubsub *redis.PubSub
	c      chan *goapackage.EnduroMonitorEvent // channel of events
	stopCh chan struct{}
}

var _ Subscription = (*SubscriptionRedisImpl)(nil)

func NewSubscriptionRedis(c redis.UniversalClient, channel string) Subscription {
	ctx := context.Background()
	pubsub := c.Subscribe(ctx, channel)
	// Call Receive to force the connection to wait a response from
	// Redis so the subscription is active immediately.
	_, _ = pubsub.Receive(ctx)
	sub := SubscriptionRedisImpl{
		pubsub: pubsub,
		c:      make(chan *goapackage.EnduroMonitorEvent, EventBufferSize),
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
			event := goapackage.EnduroMonitorEvent{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err == nil {
				s.c <- &event
			}
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
func (s *SubscriptionRedisImpl) C() <-chan *goapackage.EnduroMonitorEvent {
	return s.c
}

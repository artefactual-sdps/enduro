package auth

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

// TicketStore persists expirable tickets.
type TicketStore interface {
	// SetEx persists a key/value pair with a timeout.
	SetEx(ctx context.Context, key string, value any, ttl time.Duration) error
	// GetDel checks whether a key exists in the store and scans the value.
	// It returns ErrKeyNotFound if the key was not found or expired.
	GetDel(ctx context.Context, key string, value any) error
	// Close the client.
	Close() error
}

var ErrKeyNotFound = errors.New("key not found")

// RedisStore is an implementation of TicketStore based on Redis.
type RedisStore struct {
	client redis.UniversalClient

	// Keys will be prefixed to allow for sharing the same list with other apps.
	prefix string
}

var _ TicketStore = (*RedisStore)(nil)

const (
	// Components used to build keys, e.g. prefix:ticket:key.
	keySeparator  = ":"
	keyClassifier = "ticket"
)

func NewRedisStore(ctx context.Context, tp trace.TracerProvider, cfg *RedisConfig) (*RedisStore, error) {
	opts, err := redis.ParseURL(cfg.Address)
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

	return &RedisStore{
		client: client,
		prefix: strings.TrimSuffix(cfg.Prefix, keySeparator),
	}, nil
}

// key generates the final key to be stored including the configured prefix.
func (s *RedisStore) key(key string) string {
	return strings.Join([]string{s.prefix, keyClassifier, key}, keySeparator)
}

func (s *RedisStore) SetEx(ctx context.Context, key string, value any, ttl time.Duration) error {
	return s.client.SetEx(ctx, s.key(key), value, ttl).Err()
}

func (s *RedisStore) GetDel(ctx context.Context, key string, value any) error {
	cmd := s.client.GetDel(ctx, s.key(key))
	if err := cmd.Err(); err == redis.Nil {
		return ErrKeyNotFound
	} else if err != nil {
		return err
	}
	if value != nil {
		if err := cmd.Scan(value); err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

type InMemStore struct {
	keys map[string]*InMemKey
	mu   sync.Mutex
}

type InMemKey struct {
	value     any
	createdAt time.Time
	ttl       time.Duration
}

var _ TicketStore = (*InMemStore)(nil)

func NewInMemStore() *InMemStore {
	return &InMemStore{
		keys: map[string]*InMemKey{},
	}
}

func (s *InMemStore) SetEx(ctx context.Context, key string, value any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[key] = &InMemKey{
		value:     value,
		createdAt: time.Now(),
		ttl:       ttl,
	}

	return nil
}

func (s *InMemStore) GetDel(ctx context.Context, key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	match, ok := s.keys[key]
	if !ok {
		return ErrKeyNotFound
	}

	if age := time.Until(match.createdAt).Abs(); age > match.ttl {
		delete(s.keys, key)
		return ErrKeyNotFound
	}

	// Set the value using reflection.
	if value != nil {
		valPtr := reflect.ValueOf(value)
		if valPtr.Kind() != reflect.Ptr {
			return fmt.Errorf("value argument must be a pointer")
		}
		val := valPtr.Elem()
		memVal := reflect.ValueOf(match.value)
		// If stored value is a pointer, dereference it.
		if memVal.Kind() == reflect.Ptr {
			memVal = memVal.Elem()
		}
		if memVal.Type() != val.Type() {
			return fmt.Errorf("type mismatch: store value is %s, argument is %s", memVal.Type(), val.Type())
		}
		val.Set(memVal)
	}

	delete(s.keys, key)

	return nil
}

func (s *InMemStore) Close() error {
	return nil
}

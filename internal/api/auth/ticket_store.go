package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// TicketStore persists expirable tickets.
type TicketStore interface {
	// SetEX persists a key with a timeout.
	SetEX(ctx context.Context, key string, ttl time.Duration) error
	// GetDel checks whether a key exists in the store. It returns a non-nil
	// error if the key was not found or expired.
	GetDel(ctx context.Context, key string) error
	// Close the client.
	Close() error
}

// RedisStore is an implementation of TicketStore based on Redis.
type RedisStore struct {
	client redis.UniversalClient
}

var _ TicketStore = (*RedisStore)(nil)

func NewRedisStore(ctx context.Context, cfg *RedisConfig) (*RedisStore, error) {
	opts, err := redis.ParseURL(cfg.Address)
	if err != nil {
		return nil, err
	}
	return &RedisStore{client: redis.NewClient(opts).WithContext(ctx)}, nil
}

func (s *RedisStore) SetEX(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.SetEX(ctx, key, "", ttl).Err()
}

func (s *RedisStore) GetDel(ctx context.Context, key string) error {
	return s.client.GetDel(ctx, key).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

var ErrKeyNotFound = errors.New("key not found")

type InMemStore struct {
	keys map[string]*InMemKey
	mu   sync.Mutex
}

type InMemKey struct {
	createdAt time.Time
	ttl       time.Duration
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		keys: map[string]*InMemKey{},
	}
}

func (s *InMemStore) SetEX(ctx context.Context, key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[key] = &InMemKey{
		createdAt: time.Now(),
		ttl:       ttl,
	}

	return nil
}

func (s *InMemStore) GetDel(ctx context.Context, key string) error {
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

	delete(s.keys, key)

	return nil
}

func (s *InMemStore) Close() error {
	return nil
}

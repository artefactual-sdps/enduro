package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TicketStore persists expirable tickets.
type TicketStore interface {
	// SetEx persists a key with a timeout.
	SetEx(ctx context.Context, key string, ttl time.Duration) error
	// GetDel checks whether a key exists in the store. It returns
	// ErrKeyNotFound if the key was not found or expired.
	GetDel(ctx context.Context, key string) error
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

func NewRedisStore(ctx context.Context, cfg *RedisConfig) (*RedisStore, error) {
	opts, err := redis.ParseURL(cfg.Address)
	if err != nil {
		return nil, err
	}

	return &RedisStore{
		client: redis.NewClient(opts),
		prefix: strings.TrimSuffix(cfg.Prefix, keySeparator),
	}, nil
}

// key generates the final key to be stored including the configured prefix.
func (s *RedisStore) key(key string) string {
	return strings.Join([]string{s.prefix, keyClassifier, key}, keySeparator)
}

func (s *RedisStore) SetEx(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.SetEx(ctx, s.key(key), "", ttl).Err()
}

func (s *RedisStore) GetDel(ctx context.Context, key string) error {
	if err := s.client.GetDel(ctx, s.key(key)).Err(); err == redis.Nil {
		return ErrKeyNotFound
	} else if err != nil {
		return err
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
	createdAt time.Time
	ttl       time.Duration
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		keys: map[string]*InMemKey{},
	}
}

func (s *InMemStore) SetEx(ctx context.Context, key string, ttl time.Duration) error {
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

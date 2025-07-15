package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

func TestRedisStore(t *testing.T) {
	t.Parallel()

	storeKey := "key"
	tp := noop.NewTracerProvider()

	t.Run("Fails when parsing invalid URL", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		_, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "scheme://unknown",
		})
		assert.Error(t, err, "redis: invalid URL scheme: scheme")
	})

	t.Run("Fails when server is unreachable", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://127.0.0.1:12345",
		})
		assert.NilError(t, err)

		err = s.GetDel(ctx, storeKey, nil)
		assert.Error(t, err, "dial tcp 127.0.0.1:12345: connect: connection refused")
	})

	t.Run("Stores the ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		store, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "prefix",
		})
		assert.NilError(t, err)

		value := "value"
		err = store.SetEx(ctx, storeKey, value, time.Second)
		assert.NilError(t, err)

		// It should find the item.
		cmd := redisClient.Get(ctx, "prefix:ticket:"+storeKey)
		assert.NilError(t, cmd.Err())
		assert.Equal(t, cmd.Val(), value)

		// It should error as keys can only be used once.
		redisServer.FastForward(time.Minute)
		cmd = redisClient.Get(ctx, "prefix:ticket:"+storeKey)
		assert.ErrorIs(t, cmd.Err(), redis.Nil)
	})

	t.Run("Handles prefix config with trailing separator", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		store, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "prefix:",
		})
		assert.NilError(t, err)

		value := "value"
		err = store.SetEx(ctx, storeKey, value, time.Second)
		assert.NilError(t, err)

		// It should find the item.
		cmd := redisClient.Get(ctx, "prefix:ticket:"+storeKey)
		assert.NilError(t, cmd.Err())
		assert.Equal(t, cmd.Val(), value)
	})

	t.Run("Checks the ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		value := "value"
		err := redisClient.SetEx(ctx, "prefix:ticket:"+storeKey, value, time.Minute).Err()
		assert.NilError(t, err)

		store, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "prefix",
		})
		assert.NilError(t, err)

		// It scans the value and returns nil error as the key is not expired.
		var scannedValue string
		assert.NilError(t, store.GetDel(ctx, storeKey, &scannedValue))
		assert.Equal(t, scannedValue, value)

		// It returns an error as the key was removed in the previous operation.
		assert.ErrorIs(t, store.GetDel(ctx, storeKey, nil), auth.ErrKeyNotFound)
	})

	t.Run("Fails checking an expired ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		err := redisClient.SetEx(ctx, "prefix:ticket:"+storeKey, nil, time.Second*5).Err()
		assert.NilError(t, err)

		store, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "prefix",
		})
		assert.NilError(t, err)

		redisServer.FastForward(time.Minute)

		assert.ErrorIs(t, store.GetDel(ctx, storeKey, nil), auth.ErrKeyNotFound)
	})

	t.Run("Doesn't scan a nil value", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		err := redisClient.SetEx(ctx, "prefix:ticket:"+storeKey, nil, time.Second*5).Err()
		assert.NilError(t, err)

		store, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "prefix",
		})
		assert.NilError(t, err)

		assert.NilError(t, store.GetDel(ctx, storeKey, nil))
	})

	t.Run("Closes the client", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)

		store, err := auth.NewRedisStore(ctx, tp, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "prefix",
		})
		assert.NilError(t, err)

		store.Close() // Close the client.
		assert.Error(t, store.SetEx(ctx, "key", nil, time.Second), "redis: client is closed")
	})
}

func TestInMemStore(t *testing.T) {
	t.Parallel()

	t.Run("Stores and checks the ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		// It stores the ticket.
		value := "value"
		err := store.SetEx(ctx, "ticket", value, time.Second)
		assert.NilError(t, err)

		// It scans the value and returns non-nil indicating that the ticket was found.
		var scannedValue string
		assert.NilError(t, store.GetDel(ctx, "ticket", &scannedValue))
		assert.Equal(t, scannedValue, value)

		// It returns error, confirming that the element was removed.
		assert.Error(t, store.GetDel(ctx, "ticket", nil), "key not found")
	})

	t.Run("Fails checking an expired ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		err := store.SetEx(ctx, "ticket", nil, time.Nanosecond)
		assert.NilError(t, err)

		// ttl was one billionth of a second, should be expired already.
		assert.Error(t, store.GetDel(ctx, "ticket", nil), "key not found")
	})

	t.Run("Fails checking an unknown ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		assert.Error(t, store.GetDel(ctx, "ticket", nil), "key not found")
	})

	t.Run("Fails if value argument is not a pointer", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		err := store.SetEx(ctx, "ticket", "value", time.Second)
		assert.NilError(t, err)

		var notPtr string
		err = store.GetDel(ctx, "ticket", notPtr)
		assert.ErrorContains(t, err, "value argument must be a pointer")
	})

	t.Run("Fails if value types do not match", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		err := store.SetEx(ctx, "ticket", "value", time.Second)
		assert.NilError(t, err)

		var wrongType int
		err = store.GetDel(ctx, "ticket", &wrongType)
		assert.ErrorContains(t, err, "type mismatch")
	})

	t.Run("Works if stored value is a pointer", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		val := "value"
		err := store.SetEx(ctx, "ticket", &val, time.Second)
		assert.NilError(t, err)

		var scannedValue string
		err = store.GetDel(ctx, "ticket", &scannedValue)
		assert.NilError(t, err)
		assert.Equal(t, scannedValue, val)
	})

	t.Run("Works with a nil value argument", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		err := store.SetEx(ctx, "ticket", "value", time.Second)
		assert.NilError(t, err)

		err = store.GetDel(ctx, "ticket", nil)
		assert.NilError(t, err)
	})
}

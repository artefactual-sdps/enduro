package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

func TestRedisStore(t *testing.T) {
	t.Parallel()

	storeKey := "key"

	t.Run("Fails when parsing invalid URL", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		_, err := auth.NewRedisStore(ctx, &auth.RedisConfig{
			Address: "scheme://unknown",
		})
		assert.Error(t, err, "redis: invalid URL scheme: scheme")
	})

	t.Run("Fails when server is unreachable", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s, err := auth.NewRedisStore(ctx, &auth.RedisConfig{
			Address: "redis://127.0.0.1:12345",
		})
		assert.NilError(t, err)

		err = s.GetDel(ctx, storeKey)
		assert.Error(t, err, "dial tcp 127.0.0.1:12345: connect: connection refused")
	})

	t.Run("Stores the ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		store, err := auth.NewRedisStore(ctx, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "foo",
		})
		assert.NilError(t, err)

		err = store.SetEX(ctx, storeKey, time.Second)
		assert.NilError(t, err)

		// It should find the item.
		cmd := redisClient.Get(ctx, storeKey)
		assert.NilError(t, cmd.Err())
		assert.Equal(t, cmd.Val(), "")

		// It should error when key is expired.
		redisServer.FastForward(time.Minute)
		cmd = redisClient.Get(ctx, storeKey)
		assert.Error(t, cmd.Err(), "redis: nil")
	})

	t.Run("Checks the ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		err := redisClient.SetEX(ctx, storeKey, "", time.Minute).Err()
		assert.NilError(t, err)

		store, err := auth.NewRedisStore(ctx, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "foo",
		})
		assert.NilError(t, err)

		// It returns nil error as the key is not expired.
		assert.NilError(t, store.GetDel(ctx, storeKey))

		// It returns an error as the key was removed in the previous operation.
		assert.Error(t, store.GetDel(ctx, storeKey), "redis: nil")
	})

	t.Run("Fails checking an expired ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		err := redisClient.SetEX(ctx, storeKey, "", time.Second*5).Err()
		assert.NilError(t, err)

		store, err := auth.NewRedisStore(ctx, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "foo",
		})
		assert.NilError(t, err)

		redisServer.FastForward(time.Minute)

		assert.Error(t, store.GetDel(ctx, storeKey), "redis: nil")
	})

	t.Run("Closes the client", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		redisServer := miniredis.RunT(t)

		store, err := auth.NewRedisStore(ctx, &auth.RedisConfig{
			Address: "redis://" + redisServer.Addr(),
			Prefix:  "foo",
		})
		assert.NilError(t, err)

		store.Close() // Close the client.
		assert.Error(t, store.SetEX(ctx, "x", time.Second), "redis: client is closed")
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
		err := store.SetEX(ctx, "ticket", time.Second)
		assert.NilError(t, err)

		// It returns non-nil indicating that the ticket was found
		assert.NilError(t, store.GetDel(ctx, "ticket"))

		// It returns error, confirming that the element was removed.
		assert.Error(t, store.GetDel(ctx, "ticket"), "key not found")
	})

	t.Run("Fails checking an expired ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		err := store.SetEX(ctx, "ticket", time.Nanosecond)
		assert.NilError(t, err)

		// ttl was one billionth of a second, should be expired already.
		assert.Error(t, store.GetDel(ctx, "ticket"), "key not found")
	})

	t.Run("Fails checking an unknown ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := auth.NewInMemStore()
		defer store.Close()

		assert.Error(t, store.GetDel(ctx, "ticket"), "key not found")
	})
}

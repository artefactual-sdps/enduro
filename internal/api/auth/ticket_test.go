package auth_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"testing/iotest"

	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	"github.com/artefactual-sdps/enduro/internal/api/auth/fake"
)

func TestTicketProviderNop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := auth.NewTicketProvider(ctx, nil, nil)

	ticket, err := provider.Request(ctx)
	assert.NilError(t, err)
	assert.Equal(t, ticket, "")

	err = provider.Check(ctx, ticket)
	assert.NilError(t, err)

	err = provider.Close()
	assert.NilError(t, err)
}

func TestTicketProviderRequest(t *testing.T) {
	t.Parallel()

	prefix := "prefix"

	t.Run("Generates a ticket on request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			SetEX(gomock.Any(), "prefix:session:Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk", auth.TicketTTL).
			Return(nil)

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, prefix, rander)

		ticket, err := provider.Request(ctx)
		assert.NilError(t, err)
		assert.Equal(t, ticket, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk")
	})

	t.Run("Fails when the source of randomness errors", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		rander := iotest.ErrReader(errors.New("rand source error"))
		provider := auth.NewTicketProvider(ctx, store, prefix, rander)

		ticket, err := provider.Request(ctx)
		assert.Error(t, err, "error creating ticket: rand source error")
		assert.Assert(t, ticket == "")
	})

	t.Run("Fails when the store operation fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			SetEX(gomock.Any(), "prefix:session:Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk", auth.TicketTTL).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, prefix, rander)

		ticket, err := provider.Request(ctx)
		assert.Error(t, err, "error storing ticket: fake error")
		assert.Assert(t, ticket == "")
	})
}

func TestTicketProviderCheck(t *testing.T) {
	t.Parallel()

	prefix := "prefix"

	t.Run("Checks the existence of a ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), "prefix:session:Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk").
			Return(nil)

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, prefix, rander)

		err := provider.Check(ctx, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk")
		assert.NilError(t, err)
	})

	t.Run("Fails when the ticket does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), "prefix:session:Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk").
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, prefix, rander)

		err := provider.Check(ctx, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk")
		assert.Error(t, err, "error retrieving ticket: fake error")
	})
}

func TestTicketProviderClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := fake.NewMockTicketStore(ctrl)

	store.EXPECT().Close().Return(nil)

	provider := auth.NewTicketProvider(ctx, store, "prefix", nil)
	assert.NilError(t, provider.Close())
}

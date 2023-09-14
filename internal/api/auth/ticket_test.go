package auth_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"testing/iotest"

	"go.uber.org/mock/gomock"
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

	rander := rand.New(rand.NewSource(1)) //#nosec
	ticket := "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"

	t.Run("Generates a ticket on request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			SetEx(gomock.Any(), ticket, auth.TicketTTL).
			Return(nil)

		provider := auth.NewTicketProvider(ctx, store, rander)

		ticket, err := provider.Request(ctx)
		assert.NilError(t, err)
		assert.Equal(t, ticket, ticket)
	})

	t.Run("Fails when the source of randomness errors", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		rander := iotest.ErrReader(errors.New("rand source error"))
		provider := auth.NewTicketProvider(ctx, store, rander)

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
			SetEx(gomock.Any(), ticket, auth.TicketTTL).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		ticket, err := provider.Request(ctx)
		assert.Error(t, err, "error storing ticket: fake error")
		assert.Assert(t, ticket == "")
	})
}

func TestTicketProviderCheck(t *testing.T) {
	t.Parallel()

	ticket := "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"

	t.Run("Checks the existence of a ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), ticket).
			Return(nil)

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		err := provider.Check(ctx, ticket)
		assert.NilError(t, err)
	})

	t.Run("Fails when the ticket does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), ticket).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		err := provider.Check(ctx, ticket)
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

	provider := auth.NewTicketProvider(ctx, store, nil)
	assert.NilError(t, provider.Close())
}

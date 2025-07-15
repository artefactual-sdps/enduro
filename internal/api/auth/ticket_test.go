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

	ticket, err := provider.Request(ctx, nil)
	assert.NilError(t, err)
	assert.Equal(t, ticket, "")

	err = provider.Check(ctx, &ticket, nil)
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
		value := "value"

		store.EXPECT().
			SetEx(gomock.Any(), ticket, value, auth.TicketTTL).
			Return(nil)

		provider := auth.NewTicketProvider(ctx, store, rander)

		re, err := provider.Request(ctx, value)
		assert.NilError(t, err)
		assert.Equal(t, re, ticket)
	})

	t.Run("Fails when the source of randomness errors", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		rander := iotest.ErrReader(errors.New("rand source error"))
		provider := auth.NewTicketProvider(ctx, store, rander)

		re, err := provider.Request(ctx, nil)
		assert.Error(t, err, "error creating ticket: rand source error")
		assert.Equal(t, re, "")
	})

	t.Run("Fails when the store operation fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			SetEx(gomock.Any(), ticket, nil, auth.TicketTTL).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		re, err := provider.Request(ctx, nil)
		assert.Error(t, err, "error storing ticket: fake error")
		assert.Equal(t, re, "")
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

		value := "value"
		var scannedValue string
		store.EXPECT().
			GetDel(gomock.Any(), ticket, &scannedValue).
			DoAndReturn(func(ctx context.Context, key string, val any) error {
				*(val.(*string)) = value
				return nil
			})

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		err := provider.Check(ctx, &ticket, &scannedValue)
		assert.NilError(t, err)
		assert.Equal(t, scannedValue, value)
	})

	t.Run("Fails when the ticket does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), ticket, nil).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		err := provider.Check(ctx, &ticket, nil)
		assert.Error(t, err, "error retrieving ticket: fake error")
	})

	t.Run("Fails when the ticket is not sent and the store exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, &auth.InMemStore{}, rander)

		err := provider.Check(ctx, nil, nil)
		assert.Error(t, err, "missing ticket to retrieve")
	})

	t.Run("Always checks without store", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, nil, rander)

		err := provider.Check(ctx, nil, nil)
		assert.NilError(t, err)
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

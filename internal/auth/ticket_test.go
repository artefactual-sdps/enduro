package auth_test

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"testing"
	"testing/iotest"

	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/auth"
	"github.com/artefactual-sdps/enduro/internal/auth/fake"
)

func TestTicketProviderNop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := auth.NewTicketProvider(ctx, nil, nil)
	grant := auth.NewTicketGrant(auth.TicketPurposeIngestSIPDownload, "resource")

	ticket, err := provider.Request(ctx, grant)
	assert.NilError(t, err)
	assert.Equal(t, ticket, "")

	err = provider.Check(ctx, &ticket, grant.Purpose, grant.ResourceID)
	assert.NilError(t, err)

	err = provider.Close()
	assert.NilError(t, err)
}

func TestTicketProviderRequest(t *testing.T) {
	t.Parallel()

	rander := rand.New(rand.NewSource(1)) //#nosec
	ticket := "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"
	grant := auth.NewTicketGrant(auth.TicketPurposeIngestSIPDownload, "resource")

	t.Run("Generates a ticket on request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)
		store.EXPECT().
			SetEx(gomock.Any(), ticket, grant, auth.TicketTTL).
			Return(nil)

		provider := auth.NewTicketProvider(ctx, store, rander)

		re, err := provider.Request(ctx, grant)
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

		re, err := provider.Request(ctx, grant)
		assert.Error(t, err, "error creating ticket: rand source error")
		assert.Equal(t, re, "")
	})

	t.Run("Fails when the source of randomness is too short", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := fake.NewMockTicketStore(gomock.NewController(t))
		rander := bytes.NewReader(make([]byte, 16))
		provider := auth.NewTicketProvider(ctx, store, rander)

		re, err := provider.Request(ctx, grant)
		assert.Error(t, err, "error creating ticket: unexpected EOF")
		assert.Equal(t, re, "")
	})

	t.Run("Fails when the store operation fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			SetEx(gomock.Any(), ticket, grant, auth.TicketTTL).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		re, err := provider.Request(ctx, grant)
		assert.Error(t, err, "error storing ticket: fake error")
		assert.Equal(t, re, "")
	})
}

func TestTicketProviderCheck(t *testing.T) {
	t.Parallel()

	ticket := "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"
	grant := auth.NewTicketGrant(auth.TicketPurposeIngestSIPDownload, "resource")

	t.Run("Checks the existence of a ticket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), ticket, gomock.AssignableToTypeOf(&auth.TicketGrant{})).
			DoAndReturn(func(ctx context.Context, key string, val any) error {
				*(val.(*auth.TicketGrant)) = grant
				return nil
			})

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		err := provider.Check(ctx, &ticket, grant.Purpose, grant.ResourceID)
		assert.NilError(t, err)
	})

	t.Run("Fails when the ticket does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := fake.NewMockTicketStore(ctrl)

		store.EXPECT().
			GetDel(gomock.Any(), ticket, gomock.AssignableToTypeOf(&auth.TicketGrant{})).
			Return(errors.New("fake error"))

		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, store, rander)

		err := provider.Check(ctx, &ticket, grant.Purpose, grant.ResourceID)
		assert.Error(t, err, "error retrieving ticket: fake error")
	})

	t.Run("Fails when the ticket is not sent and the store exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, &auth.InMemStore{}, rander)

		err := provider.Check(ctx, nil, grant.Purpose, grant.ResourceID)
		assert.Error(t, err, "missing ticket to retrieve")
	})

	t.Run("Always checks without store", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		rander := rand.New(rand.NewSource(1)) //#nosec
		provider := auth.NewTicketProvider(ctx, nil, rander)

		err := provider.Check(ctx, nil, grant.Purpose, grant.ResourceID)
		assert.NilError(t, err)
	})

	for _, tt := range []struct {
		name       string
		purpose    auth.TicketPurpose
		resourceID string
		wantErr    string
	}{
		{
			name:       "Rejects the wrong purpose",
			purpose:    auth.TicketPurposeStorageAIPDownload,
			resourceID: grant.ResourceID,
			wantErr:    "ticket purpose mismatch",
		},
		{
			name:       "Rejects the wrong resource",
			purpose:    grant.Purpose,
			resourceID: "other-resource",
			wantErr:    "ticket resource mismatch",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			provider := auth.NewTicketProvider(ctx, auth.NewInMemStore(), nil)
			ticket, err := provider.Request(ctx, grant)
			assert.NilError(t, err)

			err = provider.Check(ctx, &ticket, tt.purpose, tt.resourceID)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
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

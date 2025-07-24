package storage_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	auth_fake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/storage/persistence/fake"
)

func TestDownloadAipRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	td := fs.NewDir(t, "enduro-service-test")
	missingAIPUUID := uuid.New()

	// Write a test blob to the bucket.
	writeTestBlob(ctx, t, "file://"+td.Path(), aipID.String())

	for _, tt := range []struct {
		name    string
		payload *goastorage.DownloadAipRequestPayload
		mock    func(context.Context, *auth_fake.MockTicketStore, *persistence_fake.MockStorage)
		wantErr string
		wantRes *goastorage.DownloadAipRequestResult
	}{
		{
			name:    "Fails to request a AIP download (invalid UUID)",
			payload: &goastorage.DownloadAipRequestPayload{UUID: "invalid-uuid"},
			wantErr: "cannot perform operation",
		},
		{
			name:    "Fails to request a AIP download (AIP not found)",
			payload: &goastorage.DownloadAipRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, &goastorage.AIPNotFound{UUID: aipID, Message: "AIP not found"})
			},
			wantErr: "AIP not found.",
		},
		{
			name:    "Fails to request a AIP download (persistence error)",
			payload: &goastorage.DownloadAipRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, goastorage.MakeNotAvailable(errors.New("persistence error")))
			},
			wantErr: "persistence error",
		},
		{
			name:    "Fails to request a AIP download (invalid status)",
			payload: &goastorage.DownloadAipRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(&goastorage.AIP{UUID: aipID, Status: enums.AIPStatusDeleted.String()}, nil)
			},
			wantErr: "AIP is not available for download",
		},
		{
			name:    "Fails to request a AIP download (AIP file not found)",
			payload: &goastorage.DownloadAipRequestPayload{UUID: missingAIPUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, missingAIPUUID).
					Return(
						&goastorage.AIP{
							UUID:         missingAIPUUID,
							Status:       enums.AIPStatusStored.String(),
							ObjectKey:    missingAIPUUID,
							LocationUUID: &locationID,
						},
						nil,
					)
				psvc.
					EXPECT().
					ReadLocation(ctx, locationID).
					Return(
						&goastorage.Location{
							UUID: locationID,
							Config: &goastorage.URLConfig{
								URL: "file://" + td.Path(),
							},
						},
						nil,
					)
			},
			wantErr: "AIP not found.",
		},
		{
			name:    "Fails to request a AIP download (fails to create ticket)",
			payload: &goastorage.DownloadAipRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:         aipID,
							Status:       enums.AIPStatusStored.String(),
							ObjectKey:    aipID,
							LocationUUID: &locationID,
						},
						nil,
					)
				psvc.
					EXPECT().
					ReadLocation(ctx, locationID).
					Return(
						&goastorage.Location{
							UUID: locationID,
							Config: &goastorage.URLConfig{
								URL: "file://" + td.Path(),
							},
						},
						nil,
					)
				ts.EXPECT().
					SetEx(ctx, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk", nil, time.Second*5).
					Return(errors.New("ticket error"))
			},
			wantErr: "ticket request failed",
		},
		{
			name:    "Requests a AIP download",
			payload: &goastorage.DownloadAipRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:         aipID,
							Status:       enums.AIPStatusStored.String(),
							ObjectKey:    aipID,
							LocationUUID: &locationID,
						},
						nil,
					)
				psvc.
					EXPECT().
					ReadLocation(ctx, locationID).
					Return(
						&goastorage.Location{
							UUID: locationID,
							Config: &goastorage.URLConfig{
								URL: "file://" + td.Path(),
							},
						},
						nil,
					)
				ts.EXPECT().SetEx(ctx, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk", nil, time.Second*5).Return(nil)
			},
			wantRes: &goastorage.DownloadAipRequestResult{
				Ticket: ref.New("Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rander := rand.New(rand.NewSource(1)) // #nosec
			ticketStoreMock := auth_fake.NewMockTicketStore(gomock.NewController(t))
			ticketProvider := auth.NewTicketProvider(ctx, ticketStoreMock, rander)

			var attrs setUpAttrs
			attrs.ticketProvider = ticketProvider
			svc := setUpService(t, &attrs)

			if tt.mock != nil {
				tt.mock(ctx, ticketStoreMock, attrs.persistenceMock)
			}

			res, err := svc.DownloadAipRequest(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.wantRes)
		})
	}
}

func TestDownloadAip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	td := fs.NewDir(t, "enduro-service-test")
	missingAIPUUID := uuid.New()
	content := []byte("Testing 1-2-3!")

	// Write a test blob to the bucket.
	writeTestBlob(ctx, t, "file://"+td.Path(), aipID.String())

	for _, tt := range []struct {
		name     string
		payload  *goastorage.DownloadAipPayload
		mock     func(context.Context, *auth_fake.MockTicketStore, *persistence_fake.MockStorage)
		wantErr  string
		wantRes  *goastorage.DownloadAipResult
		wantBody []byte
	}{
		{
			name:    "Fails to download a AIP (missing ticket)",
			payload: &goastorage.DownloadAipPayload{},
			wantErr: "Unauthorized",
		},
		{
			name:    "Fails to download a AIP (invalid ticket)",
			payload: &goastorage.DownloadAipPayload{Ticket: ref.New("invalid-ticket")},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "invalid-ticket", nil).Return(auth.ErrKeyNotFound)
			},
			wantErr: "Unauthorized",
		},
		{
			name: "Fails to download a AIP (invalid UUID)",
			payload: &goastorage.DownloadAipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   "invalid-uuid",
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "valid-ticket", nil).Return(nil)
			},
			wantErr: "cannot perform operation",
		},
		{
			name: "Fails to download a AIP (AIP not found)",
			payload: &goastorage.DownloadAipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   aipID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "valid-ticket", nil).Return(nil)
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, &goastorage.AIPNotFound{UUID: aipID, Message: "AIP not found"})
			},
			wantErr: "AIP not found.",
		},
		{
			name: "Fails to download a AIP (persistence error)",
			payload: &goastorage.DownloadAipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   aipID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "valid-ticket", nil).Return(nil)
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, goastorage.MakeNotAvailable(errors.New("persistence error")))
			},
			wantErr: "persistence error",
		},
		{
			name: "Fails to download a AIP (invalid status)",
			payload: &goastorage.DownloadAipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   aipID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "valid-ticket", nil).Return(nil)
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(&goastorage.AIP{UUID: aipID, Status: enums.AIPStatusDeleted.String()}, nil)
			},
			wantErr: "AIP is not available for download",
		},
		{
			name: "Fails to download a AIP (AIP file not found)",
			payload: &goastorage.DownloadAipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   missingAIPUUID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "valid-ticket", nil).Return(nil)
				psvc.EXPECT().
					ReadAIP(ctx, missingAIPUUID).
					Return(
						&goastorage.AIP{
							UUID:         missingAIPUUID,
							Status:       enums.AIPStatusStored.String(),
							ObjectKey:    missingAIPUUID,
							LocationUUID: &locationID,
						},
						nil,
					)
				psvc.
					EXPECT().
					ReadLocation(ctx, locationID).
					Return(
						&goastorage.Location{
							UUID: locationID,
							Config: &goastorage.URLConfig{
								URL: "file://" + td.Path(),
							},
						},
						nil,
					)
			},
			wantErr: "AIP not found.",
		},
		{
			name: "Downloads a AIP",
			payload: &goastorage.DownloadAipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   aipID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				ts.EXPECT().GetDel(ctx, "valid-ticket", nil).Return(nil)
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							Name:         "AIP.zip",
							UUID:         aipID,
							Status:       enums.AIPStatusStored.String(),
							ObjectKey:    aipID,
							LocationUUID: &locationID,
						},
						nil,
					)
				psvc.
					EXPECT().
					ReadLocation(ctx, locationID).
					Return(
						&goastorage.Location{
							UUID: locationID,
							Config: &goastorage.URLConfig{
								URL: "file://" + td.Path(),
							},
						},
						nil,
					)
			},
			wantBody: content,
			wantRes: &goastorage.DownloadAipResult{
				ContentDisposition: fmt.Sprintf("attachment; filename=%q", fmt.Sprintf("AIP-%s.7z", aipID)),
				ContentType:        "text/plain; charset=utf-8",
				ContentLength:      int64(len(content)),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rander := rand.New(rand.NewSource(1)) // #nosec
			ticketStoreMock := auth_fake.NewMockTicketStore(gomock.NewController(t))
			ticketProvider := auth.NewTicketProvider(ctx, ticketStoreMock, rander)

			var attrs setUpAttrs
			attrs.ticketProvider = ticketProvider
			svc := setUpService(t, &attrs)

			if tt.mock != nil {
				tt.mock(ctx, ticketStoreMock, attrs.persistenceMock)
			}

			res, body, err := svc.DownloadAip(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			t.Cleanup(func() {
				err := body.Close()
				assert.NilError(t, err)
			})

			assert.DeepEqual(t, res, tt.wantRes)
			cont, err := io.ReadAll(body)
			assert.NilError(t, err)
			assert.DeepEqual(t, cont, tt.wantBody)
		})
	}
}

package ingest_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"go.uber.org/mock/gomock"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	auth_fake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

const (
	key         = "failed.zip"
	contentType = "application/zip"
)

var (
	sipUUID = uuid.New()
	content = []byte("zipcontent")
)

func setupBucket(t *testing.T) *blob.Bucket {
	bucket := memblob.OpenBucket(nil)
	t.Cleanup(func() {
		if err := bucket.Close(); err != nil {
			t.Fatalf("close bucket: %v", err)
		}
	})

	w, err := bucket.NewWriter(t.Context(), key, &blob.WriterOptions{ContentType: contentType})
	assert.NilError(t, err)
	_, err = w.Write(content)
	assert.NilError(t, err)
	assert.NilError(t, w.Close())

	return bucket
}

func TestDownloadSipRequest(t *testing.T) {
	t.Parallel()

	bucket := setupBucket(t)

	for _, tt := range []struct {
		name    string
		payload *goaingest.DownloadSipRequestPayload
		mock    func(context.Context, *auth_fake.MockTicketStore, *persistence_fake.MockService)
		wantErr string
		wantRes *goaingest.DownloadSipRequestResult
	}{
		{
			name:    "Fails to request a SIP download (invalid UUID)",
			payload: &goaingest.DownloadSipRequestPayload{UUID: "invalid-uuid"},
			wantErr: "invalid UUID",
		},
		{
			name:    "Fails to request a SIP download (SIP not found)",
			payload: &goaingest.DownloadSipRequestPayload{UUID: sipUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				psvc.EXPECT().ReadSIP(ctx, sipUUID).Return(nil, persistence.ErrNotFound)
			},
			wantErr: "SIP not found.",
		},
		{
			name:    "Fails to request a SIP download (persistence error)",
			payload: &goaingest.DownloadSipRequestPayload{UUID: sipUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				psvc.EXPECT().ReadSIP(ctx, sipUUID).Return(nil, persistence.ErrInternal)
			},
			wantErr: "error reading SIP",
		},
		{
			name:    "Fails to request a SIP download (missing failed values)",
			payload: &goaingest.DownloadSipRequestPayload{UUID: sipUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				psvc.EXPECT().ReadSIP(ctx, sipUUID).Return(&datatypes.SIP{UUID: sipUUID}, nil)
			},
			wantErr: "SIP has no failed values",
		},
		{
			name:    "Fails to request a SIP download (SIP file not found)",
			payload: &goaingest.DownloadSipRequestPayload{UUID: sipUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				psvc.EXPECT().
					ReadSIP(ctx, sipUUID).
					Return(
						&datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: "missing"},
						nil,
					)
			},
			wantErr: "SIP not found.",
		},
		{
			name:    "Fails to request a SIP download (fails to create ticket)",
			payload: &goaingest.DownloadSipRequestPayload{UUID: sipUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				psvc.EXPECT().
					ReadSIP(ctx, sipUUID).
					Return(
						&datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: key},
						nil,
					)
				ts.EXPECT().
					SetEx(ctx, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk", time.Second*5).
					Return(errors.New("ticket error"))
			},
			wantErr: "ticket request failed",
		},
		{
			name:    "Requests a SIP download",
			payload: &goaingest.DownloadSipRequestPayload{UUID: sipUUID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				psvc.EXPECT().
					ReadSIP(ctx, sipUUID).
					Return(
						&datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: key},
						nil,
					)
				ts.EXPECT().SetEx(ctx, "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk", time.Second*5).Return(nil)
			},
			wantRes: &goaingest.DownloadSipRequestResult{
				Ticket: ref.New("Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ticketStoreMock := auth_fake.NewMockTicketStore(gomock.NewController(t))
			psvcMock := persistence_fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(ctx, ticketStoreMock, psvcMock)
			}

			rander := rand.New(rand.NewSource(1)) // #nosec
			ticketProvider := auth.NewTicketProvider(ctx, ticketStoreMock, rander)
			svc := ingest.NewService(logr.Discard(), nil, nil, nil, psvcMock, nil, ticketProvider, "", bucket, 0, nil)

			res, err := svc.Goa().DownloadSipRequest(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.wantRes)
		})
	}
}

func TestDownloadSip(t *testing.T) {
	t.Parallel()

	bucket := setupBucket(t)

	for _, tt := range []struct {
		name     string
		payload  *goaingest.DownloadSipPayload
		mock     func(context.Context, *auth_fake.MockTicketStore, *persistence_fake.MockService)
		wantErr  string
		wantRes  *goaingest.DownloadSipResult
		wantBody []byte
	}{
		{
			name:    "Fails to download a SIP (missing ticket)",
			payload: &goaingest.DownloadSipPayload{},
			wantErr: "Unauthorized",
		},
		{
			name:    "Fails to download a SIP (invalid ticket)",
			payload: &goaingest.DownloadSipPayload{Ticket: ref.New("invalid-ticket")},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "invalid-ticket").Return(auth.ErrKeyNotFound)
			},
			wantErr: "Unauthorized",
		},
		{
			name: "Fails to download a SIP (invalid UUID)",
			payload: &goaingest.DownloadSipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   "invalid-uuid",
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "valid-ticket").Return(nil)
			},
			wantErr: "invalid UUID",
		},
		{
			name: "Fails to download a SIP (SIP not found)",
			payload: &goaingest.DownloadSipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   sipUUID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "valid-ticket").Return(nil)
				psvc.EXPECT().ReadSIP(ctx, sipUUID).Return(nil, persistence.ErrNotFound)
			},
			wantErr: "SIP not found.",
		},
		{
			name: "Fails to download a SIP (persistence error)",
			payload: &goaingest.DownloadSipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   sipUUID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "valid-ticket").Return(nil)
				psvc.EXPECT().ReadSIP(ctx, sipUUID).Return(nil, persistence.ErrInternal)
			},
			wantErr: "error reading SIP",
		},
		{
			name: "Fails to download a SIP (missing failed values)",
			payload: &goaingest.DownloadSipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   sipUUID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "valid-ticket").Return(nil)
				psvc.EXPECT().ReadSIP(ctx, sipUUID).Return(&datatypes.SIP{UUID: sipUUID}, nil)
			},
			wantErr: "SIP has no failed values",
		},
		{
			name: "Fails to download a SIP (SIP file not found)",
			payload: &goaingest.DownloadSipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   sipUUID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "valid-ticket").Return(nil)
				psvc.EXPECT().
					ReadSIP(ctx, sipUUID).
					Return(
						&datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: "missing"},
						nil,
					)
			},
			wantErr: "SIP not found.",
		},
		{
			name: "Downloads a SIP",
			payload: &goaingest.DownloadSipPayload{
				Ticket: ref.New("valid-ticket"),
				UUID:   sipUUID.String(),
			},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockService) {
				ts.EXPECT().GetDel(ctx, "valid-ticket").Return(nil)
				psvc.EXPECT().
					ReadSIP(ctx, sipUUID).
					Return(
						&datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: key},
						nil,
					)
			},
			wantBody: content,
			wantRes: &goaingest.DownloadSipResult{
				ContentDisposition: fmt.Sprintf("attachment; filename=%q", key),
				ContentType:        contentType,
				ContentLength:      int64(len(content)),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ticketStoreMock := auth_fake.NewMockTicketStore(gomock.NewController(t))
			psvcMock := persistence_fake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(ctx, ticketStoreMock, psvcMock)
			}

			ticketProvider := auth.NewTicketProvider(ctx, ticketStoreMock, nil)
			svc := ingest.NewService(logr.Discard(), nil, nil, nil, psvcMock, nil, ticketProvider, "", bucket, 0, nil)

			res, body, err := svc.Goa().DownloadSip(ctx, tt.payload)
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

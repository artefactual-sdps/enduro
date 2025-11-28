package storage_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	auth_fake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/storage/persistence/fake"
)

func TestAipDeletionReportRequest(t *testing.T) {
	t.Parallel()

	aipID := uuid.New()
	invalidID := uuid.New()
	reportKey := fmt.Sprintf("reports/aip_deletion_report_%s", aipID)

	for _, tt := range []struct {
		name    string
		payload *goastorage.AipDeletionReportRequestPayload
		mock    func(context.Context, *auth_fake.MockTicketStore, *persistence_fake.MockStorage)
		wantErr string
		wantRes *goastorage.AipDeletionReportRequestResult
	}{
		{
			name:    "Fails to request a deletion report download (invalid UUID)",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: "invalid-uuid"},
			wantErr: "invalid UUID",
		},
		{
			name:    "Fails to request a deletion report download (AIP not found)",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: aipID.String()},
			mock: func(
				ctx context.Context,
				ts *auth_fake.MockTicketStore,
				psvc *persistence_fake.MockStorage,
			) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, goastorage.MakeNotFound(errors.New("AIP not found")))
			},
			wantErr: "AIP not found",
		},
		{
			name:    "Fails to request a deletion report download (persistence error)",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: aipID.String()},
			mock: func(
				ctx context.Context,
				ts *auth_fake.MockTicketStore,
				psvc *persistence_fake.MockStorage,
			) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, goastorage.MakeNotAvailable(errors.New("persistence error")))
			},
			wantErr: "persistence error",
		},
		{
			name:    "Fails to request a deletion report download (invalid status)",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: aipID.String()},
			mock: func(
				ctx context.Context,
				ts *auth_fake.MockTicketStore,
				psvc *persistence_fake.MockStorage,
			) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:              aipID,
							Status:            enums.DeletionRequestStatusPending.String(),
							DeletionReportKey: ref.New(reportKey),
						},
						nil,
					)
			},
			wantErr: "deletion report is not available for download",
		},
		{
			name:    "Fails to request a deletion report download (report file not found)",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: invalidID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, invalidID).
					Return(
						&goastorage.AIP{
							UUID:              invalidID,
							Status:            enums.AIPStatusDeleted.String(),
							DeletionReportKey: ref.New(fmt.Sprintf("reports/aip_deletion_report_%s", invalidID)),
						},
						nil,
					)
			},
			wantErr: "deletion report not found",
		},
		{
			name:    "Fails to request a AIP download (fails to create ticket)",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:              aipID,
							Status:            enums.AIPStatusDeleted.String(),
							DeletionReportKey: ref.New(reportKey),
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
			name:    "Requests a deletion report download",
			payload: &goastorage.AipDeletionReportRequestPayload{UUID: aipID.String()},
			mock: func(ctx context.Context, ts *auth_fake.MockTicketStore, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(mockutil.Context(), aipID).
					Return(
						&goastorage.AIP{
							UUID:              aipID,
							Status:            enums.AIPStatusDeleted.String(),
							DeletionReportKey: ref.New(reportKey),
						},
						nil,
					)
				ts.EXPECT().SetEx(
					ctx,
					"Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk",
					nil,
					time.Second*5,
				).Return(nil)
			},
			wantRes: &goastorage.AipDeletionReportRequestResult{
				Ticket: ref.New("Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			rander := rand.New(rand.NewSource(1)) // #nosec
			ticketStoreMock := auth_fake.NewMockTicketStore(gomock.NewController(t))
			ticketProvider := auth.NewTicketProvider(ctx, ticketStoreMock, rander)
			td := t.TempDir()

			var attrs setUpAttrs
			attrs.ticketProvider = ticketProvider
			attrs.config = &storage.Config{
				TaskQueue: "global",
				Internal: storage.LocationConfig{
					URL: "file://" + td,
				},
			}
			svc := setUpService(t, &attrs)

			// Write a test blob to the bucket.
			writeTestBlob(ctx, t, fmt.Sprintf("file://%s", td), reportKey)

			if tt.mock != nil {
				tt.mock(ctx, ticketStoreMock, attrs.persistenceMock)
			}

			res, err := svc.AipDeletionReportRequest(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.wantRes)
		})
	}
}

func TestAipDeletionReport(t *testing.T) {
	t.Parallel()

	aipID := uuid.New()
	reportKey := fmt.Sprintf("reports/aip_deletion_report_%s", aipID)
	ticket := "valid-ticket-123"

	for _, tc := range []struct {
		name               string
		payload            *goastorage.AipDeletionReportPayload
		mockStorageSvc     func(context.Context, *persistence_fake.MockStorage)
		mockTicketProvider func(context.Context, *auth_fake.MockTicketProvider)
		want               *goastorage.AipDeletionReportResult
		wantErr            string
	}{
		{
			name: "Errors if ticket is invalid",
			payload: &goastorage.AipDeletionReportPayload{
				UUID:   aipID.String(),
				Ticket: &ticket,
			},
			mockTicketProvider: func(ctx context.Context, tp *auth_fake.MockTicketProvider) {
				tp.EXPECT().
					Check(ctx, &ticket, nil).
					Return(fmt.Errorf("error retrieving ticket: invalid ticket"))
			},
			wantErr: "Unauthorized",
		},
		{
			name: "Errors if UUID is invalid",
			payload: &goastorage.AipDeletionReportPayload{
				UUID:   "invalid-uuid",
				Ticket: &ticket,
			},
			mockTicketProvider: func(ctx context.Context, tp *auth_fake.MockTicketProvider) {
				tp.EXPECT().
					Check(ctx, &ticket, nil).
					Return(nil)
			},
			wantErr: "invalid UUID",
		},
		{
			name: "Errors if AIP is not found",
			payload: &goastorage.AipDeletionReportPayload{
				UUID:   aipID.String(),
				Ticket: &ticket,
			},
			mockStorageSvc: func(ctx context.Context, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(nil, goastorage.MakeNotFound(errors.New("AIP not found")))
			},
			mockTicketProvider: func(ctx context.Context, tp *auth_fake.MockTicketProvider) {
				tp.EXPECT().
					Check(ctx, &ticket, nil).
					Return(nil)
			},
			wantErr: "AIP not found",
		},
		{
			name: "Errors if AIP status is invalid",
			payload: &goastorage.AipDeletionReportPayload{
				UUID:   aipID.String(),
				Ticket: &ticket,
			},
			mockStorageSvc: func(ctx context.Context, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:   aipID,
							Status: enums.AIPStatusStored.String(),
						},
						nil,
					)
			},
			mockTicketProvider: func(ctx context.Context, tp *auth_fake.MockTicketProvider) {
				tp.EXPECT().
					Check(ctx, &ticket, nil).
					Return(nil)
			},
			wantErr: "deletion report is not available for download",
		},
		{
			name: "Errors if deletion report file is not found",
			payload: &goastorage.AipDeletionReportPayload{
				UUID:   aipID.String(),
				Ticket: &ticket,
			},
			mockStorageSvc: func(ctx context.Context, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:              aipID,
							Status:            enums.AIPStatusDeleted.String(),
							DeletionReportKey: ref.New("reports/missing_deletion_report.pdf"),
						},
						nil,
					)
			},
			mockTicketProvider: func(ctx context.Context, tp *auth_fake.MockTicketProvider) {
				tp.EXPECT().
					Check(ctx, &ticket, nil).
					Return(nil)
			},
			wantErr: "deletion report not found",
		},
		{
			name: "Downloads a deletion report",
			payload: &goastorage.AipDeletionReportPayload{
				UUID:   aipID.String(),
				Ticket: &ticket,
			},
			mockStorageSvc: func(ctx context.Context, psvc *persistence_fake.MockStorage) {
				psvc.EXPECT().
					ReadAIP(ctx, aipID).
					Return(
						&goastorage.AIP{
							UUID:              aipID,
							Status:            enums.AIPStatusDeleted.String(),
							DeletionReportKey: ref.New(reportKey),
						},
						nil,
					)
			},
			mockTicketProvider: func(ctx context.Context, tp *auth_fake.MockTicketProvider) {
				tp.EXPECT().
					Check(ctx, &ticket, nil).
					Return(nil)
			},
			want: &goastorage.AipDeletionReportResult{
				ContentType:        "text/plain; charset=utf-8",
				ContentLength:      14,
				ContentDisposition: fmt.Sprintf("attachment; filename=\"aip_deletion_report_%s.pdf\"", aipID),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ticketProvider := auth_fake.NewMockTicketProvider(gomock.NewController(t))
			td := t.TempDir()

			var attrs setUpAttrs
			attrs.ticketProvider = ticketProvider
			attrs.config = &storage.Config{
				TaskQueue: "global",
				Internal: storage.LocationConfig{
					URL: "file://" + td,
				},
			}
			svc := setUpService(t, &attrs)

			// Write a test blob to the bucket.
			writeTestBlob(ctx, t, fmt.Sprintf("file://%s", td), reportKey)

			if tc.mockStorageSvc != nil {
				tc.mockStorageSvc(ctx, attrs.persistenceMock)
			}
			if tc.mockTicketProvider != nil {
				tc.mockTicketProvider(ctx, ticketProvider)
			}

			res, _, err := svc.AipDeletionReport(ctx, tc.payload)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, res, tc.want)
		})
	}
}

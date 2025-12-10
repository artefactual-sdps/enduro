package ingest_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	"go.uber.org/mock/gomock"
	"gocloud.dev/blob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

func testSvc(t *testing.T, internalBucket *blob.Bucket, uploadMaxSize int64) (
	ingest.Service,
	*persistence_fake.MockService,
	*temporalsdk_mocks.Client,
) {
	t.Helper()

	psvc := persistence_fake.NewMockService(gomock.NewController(t))
	temporalClient := new(temporalsdk_mocks.Client)
	taskQueue := "test"
	ingestsvc := ingest.NewService(ingest.ServiceParams{
		Logger:             logr.Discard(),
		DB:                 &sql.DB{},
		TemporalClient:     temporalClient,
		EventService:       event.NewServiceNop[*goaingest.IngestEvent](),
		PersistenceService: psvc,
		TokenVerifier:      &auth.NoopTokenVerifier{},
		TicketProvider:     auth.NewTicketProvider(t.Context(), nil, nil),
		TaskQueue:          taskQueue,
		InternalStorage:    internalBucket,
		UploadMaxSize:      uploadMaxSize,
		Rander:             rand.New(rand.NewSource(1)), // #nosec: G404
		SIPSource:          &sipsource.BucketSource{},
		AuditLogger: auditlog.NewFromConfig(auditlog.Config{
			Filepath: filepath.Join(t.TempDir(), "audit.log"),
		}),
	})

	return ingestsvc, psvc, temporalClient
}

func TestCreateSIP(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		sip     datatypes.SIP
		mock    func(*persistence_fake.MockService, datatypes.SIP) *persistence_fake.MockService
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "creates a SIP",
			sip: datatypes.SIP{
				Name:   "test",
				Status: enums.SIPStatusQueued,
			},
			mock: func(svc *persistence_fake.MockService, s datatypes.SIP) *persistence_fake.MockService {
				svc.EXPECT().
					CreateSIP(mockutil.Context(), &s).
					DoAndReturn(
						func(ctx context.Context, s *datatypes.SIP) error {
							s.ID = 1
							s.CreatedAt = time.Date(2024, 3, 14, 15, 57, 25, 0, time.UTC)
							return nil
						},
					)
				return svc
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.sip)
			}

			sip := tt.sip
			err := ingestsvc.CreateSIP(context.Background(), &sip)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

type bufCloser struct {
	*bytes.Buffer
}

func (b *bufCloser) Close() error {
	return nil
}

func TestCreateSIP_AuditLog(t *testing.T) {
	t.Parallel()

	sipID := uuid.MustParse("e8d32bd5-faa4-4ce1-bb50-55d9c28b306d")
	psvc := persistence_fake.NewMockService(gomock.NewController(t))
	temporalClient := new(temporalsdk_mocks.Client)

	buf := &bufCloser{new(bytes.Buffer)}
	auditLogger := auditlog.New(buf, slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{})))

	ingestsvc := ingest.NewService(ingest.ServiceParams{
		Logger:             logr.Discard(),
		DB:                 &sql.DB{},
		TemporalClient:     temporalClient,
		EventService:       event.NewServiceNop[*goaingest.IngestEvent](),
		PersistenceService: psvc,
		TokenVerifier:      &auth.NoopTokenVerifier{},
		TicketProvider:     auth.NewTicketProvider(t.Context(), nil, nil),
		TaskQueue:          "test",
		Rander:             rand.New(rand.NewSource(1)), // #nosec: G404
		SIPSource:          &sipsource.BucketSource{},
		AuditLogger:        auditLogger,
	})

	s := datatypes.SIP{
		UUID: sipID,
		Uploader: &datatypes.User{
			Email: "test@example.com",
		},
	}

	psvc.EXPECT().
		CreateSIP(mockutil.Context(), &s).
		DoAndReturn(
			func(ctx context.Context, s *datatypes.SIP) error {
				s.ID = 1
				s.CreatedAt = time.Date(2024, 3, 14, 15, 57, 25, 0, time.UTC)
				return nil
			},
		)

	err := ingestsvc.CreateSIP(context.Background(), &s)
	assert.NilError(t, err)

	want := `"level":"INFO","msg":"SIP ingest started","type":"SIP.ingest","resourceID":"e8d32bd5-faa4-4ce1-bb50-55d9c28b306d","user":"test@example.com"`
	got := buf.String()

	assert.Assert(t,
		strings.Contains(got, want),
		fmt.Sprintf("expected: %s, got: %s", want, got),
	)
}

func TestUpdateSIP(t *testing.T) {
	t.Parallel()

	sip := &datatypes.SIP{
		ID:        1,
		UUID:      uuid.MustParse("e8d32bd5-faa4-4ce1-bb50-55d9c28b306d"),
		Name:      "sip-name",
		Status:    enums.SIPStatusQueued,
		CreatedAt: time.Date(2024, 3, 14, 15, 57, 25, 0, time.UTC),
	}
	updater := func(s *datatypes.SIP) (*datatypes.SIP, error) { return s, nil }

	for _, tt := range []struct {
		name    string
		mock    func(*persistence_fake.MockService, uuid.UUID, persistence.SIPUpdater) *persistence_fake.MockService
		want    *datatypes.SIP
		wantErr string
	}{
		{
			name: "Updates a SIP",
			mock: func(
				svc *persistence_fake.MockService,
				id uuid.UUID,
				updater persistence.SIPUpdater,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateSIP(
						mockutil.Context(),
						sip.UUID,
						mockutil.Func(
							"should update SIP",
							func(upd persistence.SIPUpdater) error {
								_, err := upd(&datatypes.SIP{})
								return err
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id uuid.UUID,
							upd persistence.SIPUpdater,
						) (*datatypes.SIP, error) {
							sip, err := upd(sip)
							return sip, err
						},
					)
				return svc
			},
			want: sip,
		},
		{
			name: "Fails to update a SIP",
			mock: func(
				svc *persistence_fake.MockService,
				id uuid.UUID,
				updater persistence.SIPUpdater,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateSIP(
						mockutil.Context(),
						sip.UUID,
						mockutil.Func(
							"should update SIP",
							func(upd persistence.SIPUpdater) error {
								_, err := upd(&datatypes.SIP{})
								return err
							},
						),
					).
					Return(nil, fmt.Errorf("persistence error"))
				return svc
			},
			wantErr: "ingest: update SIP: persistence error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			tt.mock(perSvc, sip.UUID, updater)

			s, err := ingestsvc.UpdateSIP(context.Background(), sip.UUID, updater)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, s, tt.want)
		})
	}
}

func TestUpdateBatch(t *testing.T) {
	t.Parallel()

	batch := &datatypes.Batch{
		ID:         1,
		UUID:       uuid.MustParse("e8d32bd5-faa4-4ce1-bb50-55d9c28b306d"),
		Identifier: "batch-identifier",
		SIPSCount:  5,
		Status:     enums.BatchStatusQueued,
		CreatedAt:  time.Date(2024, 3, 14, 15, 57, 25, 0, time.UTC),
	}
	updater := func(b *datatypes.Batch) (*datatypes.Batch, error) { return b, nil }

	for _, tt := range []struct {
		name    string
		mock    func(*persistence_fake.MockService, uuid.UUID, persistence.BatchUpdater) *persistence_fake.MockService
		want    *datatypes.Batch
		wantErr string
	}{
		{
			name: "Updates a batch",
			mock: func(
				svc *persistence_fake.MockService,
				id uuid.UUID,
				updater persistence.BatchUpdater,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateBatch(
						mockutil.Context(),
						batch.UUID,
						mockutil.Func(
							"should update batch",
							func(upd persistence.BatchUpdater) error {
								_, err := upd(&datatypes.Batch{})
								return err
							},
						),
					).
					DoAndReturn(
						func(
							ctx context.Context,
							id uuid.UUID,
							upd persistence.BatchUpdater,
						) (*datatypes.Batch, error) {
							batch, err := upd(batch)
							return batch, err
						},
					)
				return svc
			},
			want: batch,
		},
		{
			name: "Fails to update a batch",
			mock: func(
				svc *persistence_fake.MockService,
				id uuid.UUID,
				updater persistence.BatchUpdater,
			) *persistence_fake.MockService {
				svc.EXPECT().
					UpdateBatch(
						mockutil.Context(),
						batch.UUID,
						mockutil.Func(
							"should update batch",
							func(upd persistence.BatchUpdater) error {
								_, err := upd(&datatypes.Batch{})
								return err
							},
						),
					).
					Return(nil, fmt.Errorf("persistence error"))
				return svc
			},
			wantErr: "ingest: update Batch: persistence error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ingestsvc, perSvc, _ := testSvc(t, nil, 0)
			tt.mock(perSvc, batch.UUID, updater)

			b, err := ingestsvc.UpdateBatch(context.Background(), batch.UUID, updater)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, b, tt.want)
		})
	}
}

func TestSetStatus(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	status := enums.SIPStatusProcessing

	ingestsvc, perSvc, _ := testSvc(t, nil, 0)
	perSvc.EXPECT().
		UpdateSIP(
			mockutil.Context(),
			sipUUID,
			mockutil.Func(
				"should update SIP status",
				func(upd persistence.SIPUpdater) error {
					updated, err := upd(&datatypes.SIP{})
					if err != nil {
						return err
					}
					assert.Equal(t, updated.Status, status)
					return nil
				},
			),
		).
		Return(&datatypes.SIP{UUID: sipUUID, Status: status}, nil)

	assert.NilError(t, ingestsvc.SetStatus(context.Background(), sipUUID, status))
}

func TestSetStatusInProgress(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	startedAt := time.Now().UTC()

	ingestsvc, perSvc, _ := testSvc(t, nil, 0)
	perSvc.EXPECT().
		UpdateSIP(
			mockutil.Context(),
			sipUUID,
			mockutil.Func(
				"should update SIP started at",
				func(upd persistence.SIPUpdater) error {
					updated, err := upd(&datatypes.SIP{})
					if err != nil {
						return err
					}
					assert.Equal(t, updated.Status, enums.SIPStatusProcessing)
					assert.Assert(t, updated.StartedAt.Valid)
					assert.Equal(t, updated.StartedAt.Time, startedAt)
					return nil
				},
			),
		).
		Return(&datatypes.SIP{
			UUID:      sipUUID,
			Status:    enums.SIPStatusProcessing,
			StartedAt: sql.NullTime{Time: startedAt, Valid: true},
		}, nil)

	assert.NilError(t, ingestsvc.SetStatusInProgress(context.Background(), sipUUID, startedAt))
}

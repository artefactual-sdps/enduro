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

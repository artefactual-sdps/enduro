package ingest_test

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	"go.uber.org/mock/gomock"
	"gocloud.dev/blob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func testSvc(t *testing.T, b *blob.Bucket, s int64) (
	ingest.Service,
	*persistence_fake.MockService,
	*temporalsdk_mocks.Client,
) {
	t.Helper()

	psvc := persistence_fake.NewMockService(gomock.NewController(t))
	tc := new(temporalsdk_mocks.Client)
	ingestsvc := ingest.NewService(
		logr.Discard(),
		&sql.DB{},
		tc,
		event.NopEventService(),
		psvc,
		&auth.NoopTokenVerifier{},
		&auth.TicketProvider{},
		"test",
		b,
		s,
		rand.New(rand.NewSource(1)), // #nosec: G404
	)

	return ingestsvc, psvc, tc
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

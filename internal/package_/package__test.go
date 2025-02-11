package package__test

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/artefactual-sdps/enduro/internal/package_"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func testSvc(t *testing.T, b *blob.Bucket, s int64) (package_.Service, *persistence_fake.MockService) {
	t.Helper()

	psvc := persistence_fake.NewMockService(gomock.NewController(t))
	pkgSvc := package_.NewService(
		logr.Discard(),
		&sql.DB{},
		new(temporalsdk_mocks.Client),
		event.NopEventService(),
		psvc,
		&auth.NoopTokenVerifier{},
		&auth.TicketProvider{},
		"test",
		b,
		s,
	)

	return pkgSvc, psvc
}

func TestCreatePackage(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		sip     datatypes.SIP
		mock    func(*persistence_fake.MockService, datatypes.SIP) *persistence_fake.MockService
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "creates a package",
			sip: datatypes.SIP{
				Name:       "test",
				WorkflowID: "4258090a-e27b-4fd9-a76b-28deb3d16813",
				RunID:      "8f3a5756-6bc5-4d82-846d-59442dd6ad8f",
				Status:     enums.SIPStatusQueued,
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
		{
			name: "errors creating a package with a missing RunID",
			sip: datatypes.SIP{
				Name:       "test",
				WorkflowID: "4258090a-e27b-4fd9-a76b-28deb3d16813",
				Status:     enums.SIPStatusQueued,
			},
			mock: func(svc *persistence_fake.MockService, s datatypes.SIP) *persistence_fake.MockService {
				svc.EXPECT().
					CreateSIP(mockutil.Context(), &s).
					DoAndReturn(
						func(ctx context.Context, s *datatypes.SIP) error {
							return fmt.Errorf("invalid data error: field \"RunID\" is required")
						},
					)
				return svc
			},
			wantErr: "package: create: invalid data error: field \"RunID\" is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pkgSvc, perSvc := testSvc(t, nil, 0)
			if tt.mock != nil {
				tt.mock(perSvc, tt.sip)
			}

			sip := tt.sip
			err := pkgSvc.Create(context.Background(), &sip)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

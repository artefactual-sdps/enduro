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
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/package_"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func testSvc(t *testing.T) (package_.Service, *persistence_fake.MockService) {
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
	)

	return pkgSvc, psvc
}

func TestCreatePackage(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		pkg     datatypes.Package
		mock    func(*persistence_fake.MockService, datatypes.Package) *persistence_fake.MockService
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "creates a package",
			pkg: datatypes.Package{
				Name:       "test",
				WorkflowID: "4258090a-e27b-4fd9-a76b-28deb3d16813",
				RunID:      "8f3a5756-6bc5-4d82-846d-59442dd6ad8f",
				Status:     enums.NewPackageStatus("new"),
			},
			mock: func(svc *persistence_fake.MockService, p datatypes.Package) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePackage(mockutil.Context(), &p).
					DoAndReturn(
						func(ctx context.Context, p *datatypes.Package) error {
							p.ID = 1
							p.CreatedAt = time.Date(2024, 3, 14, 15, 57, 25, 0, time.UTC)
							return nil
						},
					)
				return svc
			},
		},
		{
			name: "errors creating a package with a missing RunID",
			pkg: datatypes.Package{
				Name:       "test",
				WorkflowID: "4258090a-e27b-4fd9-a76b-28deb3d16813",
				Status:     enums.NewPackageStatus("new"),
			},
			mock: func(svc *persistence_fake.MockService, p datatypes.Package) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePackage(mockutil.Context(), &p).
					DoAndReturn(
						func(ctx context.Context, p *datatypes.Package) error {
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

			pkgSvc, perSvc := testSvc(t)
			if tt.mock != nil {
				tt.mock(perSvc, tt.pkg)
			}

			pkg := tt.pkg
			err := pkgSvc.Create(context.Background(), &pkg)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}

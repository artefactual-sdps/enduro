package package__test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
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

func TestPackage(t *testing.T) {
	t.Run("Create a package", func(t *testing.T) {
		testPkg := datatypes.Package{
			Name:       "test",
			WorkflowID: "4258090a-e27b-4fd9-a76b-28deb3d16813",
			RunID:      "8f3a5756-6bc5-4d82-846d-59442dd6ad8f",
			AIPID:      "99698bb5-2eb0-4cf5-aebf-0da2efe7ce94",
			LocationID: uuid.NullUUID{
				UUID:  uuid.MustParse("a0075fc6-13bb-4ed4-b485-5d363fc7d048"),
				Valid: true,
			},
			Status: enums.NewPackageStatus("new"),
		}

		pkgSvc, perSvc := testSvc(t)
		perSvc.EXPECT().CreatePackage(mockutil.Context(), &testPkg).DoAndReturn(
			func(ctx context.Context, p *datatypes.Package) (*datatypes.Package, error) {
				p.ID = 1
				return p, nil
			},
		)

		err := pkgSvc.Create(context.Background(), &testPkg)
		assert.NilError(t, err)
	})
}

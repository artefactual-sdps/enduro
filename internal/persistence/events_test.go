package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	mockclient "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

var (
	CreatedAt = time.Unix(1694213364, 0) // 2023-09-08T22:49:24+00:00
	StartedAt = time.Unix(1694213435, 0) // 2023-09-08T22:50:35+00:00
)

func TestCreatePackage(t *testing.T) {
	ctx := context.Background()
	aipID := uuid.NullUUID{
		UUID:  uuid.MustParse("57e9d085-5716-43d2-bad9-bba3c9a74bd8"),
		Valid: true,
	}

	evsvc := event.NewEventServiceInMemImpl()
	sub, err := evsvc.Subscribe(ctx)
	assert.NilError(t, err)

	msvc := mockclient.NewMockService(gomock.NewController(t))
	msvc.
		EXPECT().
		CreatePackage(mockutil.Context(),
			&datatypes.Package{
				Name:       "Fake package",
				WorkflowID: "workflow-1",
				RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
				AIPID:      aipID,
				Status:     enums.PackageStatusInProgress,
				StartedAt:  sql.NullTime{Time: StartedAt, Valid: true},
			},
		).
		DoAndReturn(func(ctx context.Context, p *datatypes.Package) error {
			p.ID = 1
			p.CreatedAt = CreatedAt

			return nil
		})

	svc := persistence.WithEvents(evsvc, msvc)
	pkg := datatypes.Package{
		Name:       "Fake package",
		WorkflowID: "workflow-1",
		RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
		AIPID:      aipID,
		Status:     enums.PackageStatusInProgress,
		StartedAt:  sql.NullTime{Time: StartedAt, Valid: true},
	}

	err = svc.CreatePackage(ctx, &pkg)

	assert.NilError(t, err)
	assert.DeepEqual(t, pkg, datatypes.Package{
		ID:         1,
		Name:       "Fake package",
		WorkflowID: "workflow-1",
		RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
		AIPID:      aipID,
		Status:     enums.PackageStatusInProgress,
		CreatedAt:  CreatedAt,
		StartedAt:  sql.NullTime{Time: StartedAt, Valid: true},
	})

	// Verify subscriber received the event.
	select {
	case ev := <-sub.C():
		assert.Assert(t, ev.Event != nil)
	default:
		t.Fatal("expected event")
	}
}

func TestUpdatePackage(t *testing.T) {
	ctx := context.Background()
	aipID := uuid.NullUUID{
		UUID:  uuid.MustParse("57e9d085-5716-43d2-bad9-bba3c9a74bd8"),
		Valid: true,
	}
	completed := time.Now()

	evsvc := event.NewEventServiceInMemImpl()
	sub, err := evsvc.Subscribe(ctx)
	assert.NilError(t, err)

	msvc := mockclient.NewMockService(gomock.NewController(t))
	msvc.
		EXPECT().
		UpdatePackage(mockutil.Context(), uint(1), mockutil.Func(
			"updates package",
			func(updater persistence.PackageUpdater) error {
				_, err := updater(&datatypes.Package{})
				return err
			}),
		).
		Return(&datatypes.Package{
			ID:          1,
			Name:        "Fake package",
			WorkflowID:  "workflow-1",
			RunID:       "d1fec389-d50f-423f-843f-a510584cc02c",
			AIPID:       aipID,
			Status:      enums.PackageStatusDone,
			CreatedAt:   CreatedAt,
			StartedAt:   sql.NullTime{Time: StartedAt, Valid: true},
			CompletedAt: sql.NullTime{Time: completed, Valid: true},
		}, nil)

	svc := persistence.WithEvents(evsvc, msvc)
	got, err := svc.UpdatePackage(ctx, uint(1), func(pkg *datatypes.Package) (*datatypes.Package, error) {
		pkg.Status = enums.PackageStatusDone
		pkg.CompletedAt = sql.NullTime{Time: completed, Valid: true}
		return pkg, nil
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, got, &datatypes.Package{
		ID:          1,
		Name:        "Fake package",
		WorkflowID:  "workflow-1",
		RunID:       "d1fec389-d50f-423f-843f-a510584cc02c",
		AIPID:       aipID,
		Status:      enums.PackageStatusDone,
		CreatedAt:   CreatedAt,
		StartedAt:   sql.NullTime{Time: StartedAt, Valid: true},
		CompletedAt: sql.NullTime{Time: completed, Valid: true},
	})

	// Verify subscriber received the event.
	select {
	case ev := <-sub.C():
		assert.Assert(t, ev.Event != nil)
	default:
		t.Fatal("expected event")
	}
}

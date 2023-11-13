package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go.artefactual.dev/tools/mockutil"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	mockclient "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

var (
	CreatedAt = time.Unix(1694213364, 0) // 2023-09-08T22:49:24+00:00
	StartedAt = time.Unix(1694213435, 0) // 2023-09-08T22:50:35+00:00
)

func TestCreatePackage(t *testing.T) {
	ctx := context.Background()

	evsvc := event.NewEventServiceInMemImpl()
	sub, err := evsvc.Subscribe(ctx)
	assert.NilError(t, err)

	msvc := mockclient.NewMockService(gomock.NewController(t))
	msvc.
		EXPECT().
		CreatePackage(mockutil.Context(),
			&package_.Package{
				Name:       "Fake package",
				WorkflowID: "workflow-1",
				RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
				AIPID:      "57e9d085-5716-43d2-bad9-bba3c9a74bd8",
				Status:     package_.StatusInProgress,
				StartedAt:  sql.NullTime{Time: StartedAt, Valid: true},
			},
		).
		Return(&package_.Package{
			ID:         1,
			Name:       "Fake package",
			WorkflowID: "workflow-1",
			RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
			AIPID:      "57e9d085-5716-43d2-bad9-bba3c9a74bd8",
			Status:     package_.StatusInProgress,
			CreatedAt:  CreatedAt,
			StartedAt:  sql.NullTime{Time: StartedAt, Valid: true},
		}, nil)

	svc := persistence.WithEvents(evsvc, msvc)
	got, err := svc.CreatePackage(ctx, &package_.Package{
		Name:       "Fake package",
		WorkflowID: "workflow-1",
		RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
		AIPID:      "57e9d085-5716-43d2-bad9-bba3c9a74bd8",
		Status:     package_.StatusInProgress,
		StartedAt:  sql.NullTime{Time: StartedAt, Valid: true},
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, got, &package_.Package{
		ID:         1,
		Name:       "Fake package",
		WorkflowID: "workflow-1",
		RunID:      "d1fec389-d50f-423f-843f-a510584cc02c",
		AIPID:      "57e9d085-5716-43d2-bad9-bba3c9a74bd8",
		Status:     package_.StatusInProgress,
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
				_, err := updater(&package_.Package{})
				return err
			}),
		).
		Return(&package_.Package{
			ID:          1,
			Name:        "Fake package",
			WorkflowID:  "workflow-1",
			RunID:       "d1fec389-d50f-423f-843f-a510584cc02c",
			AIPID:       "57e9d085-5716-43d2-bad9-bba3c9a74bd8",
			Status:      package_.StatusDone,
			CreatedAt:   CreatedAt,
			StartedAt:   sql.NullTime{Time: StartedAt, Valid: true},
			CompletedAt: sql.NullTime{Time: completed, Valid: true},
		}, nil)

	svc := persistence.WithEvents(evsvc, msvc)
	got, err := svc.UpdatePackage(ctx, uint(1), func(pkg *package_.Package) (*package_.Package, error) {
		pkg.Status = package_.StatusDone
		pkg.CompletedAt = sql.NullTime{Time: completed, Valid: true}
		return pkg, nil
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, got, &package_.Package{
		ID:          1,
		Name:        "Fake package",
		WorkflowID:  "workflow-1",
		RunID:       "d1fec389-d50f-423f-843f-a510584cc02c",
		AIPID:       "57e9d085-5716-43d2-bad9-bba3c9a74bd8",
		Status:      package_.StatusDone,
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

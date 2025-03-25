package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func initialDataForWorkflowTests(t *testing.T, ctx context.Context, entc *db.Client) {
	t.Helper()

	entc.AIP.Create().
		SetName("AIP").
		SetAipID(aipID).
		SetObjectKey(objectKey).
		SetStatus(enums.AIPStatusStored).
		ExecX(ctx)
}

func TestCreateWorkflow(t *testing.T) {
	t.Parallel()

	type test struct {
		name         string
		workflow     *types.Workflow
		wantWorkflow *db.Workflow
		wantErr      string
	}

	workflowUUID := uuid.New()
	startedAt := time.Now().Add(-time.Minute)
	completedAt := time.Now()

	for _, tt := range []test{
		{
			name: "Creates a Workflow",
			workflow: &types.Workflow{
				UUID:        workflowUUID,
				TemporalID:  "temporal-id",
				Type:        enums.WorkflowTypeMoveAip,
				Status:      enums.WorkflowStatusInProgress,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				AIPUUID:     aipID,
			},
			wantWorkflow: &db.Workflow{
				ID:          1,
				UUID:        workflowUUID,
				TemporalID:  "temporal-id",
				Type:        enums.WorkflowTypeMoveAip,
				Status:      enums.WorkflowStatusInProgress,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				AipID:       1,
			},
		},
		{
			name: "Fails to create a Workflow without AIP UUID",
			workflow: &types.Workflow{
				UUID:       workflowUUID,
				TemporalID: "temporal-id",
				Type:       enums.WorkflowTypeMoveAip,
				Status:     enums.WorkflowStatusInProgress,
			},
			wantErr: "create workflow: db: aip not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForWorkflowTests(t, ctx, entc)

			err := c.CreateWorkflow(ctx, tt.workflow)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			dbw := entc.Workflow.GetX(ctx, tt.workflow.DBID)
			assert.DeepEqual(
				t,
				dbw,
				tt.wantWorkflow,
				cmpopts.IgnoreFields(db.Workflow{}, "config", "Edges", "selectValues"),
			)
		})
	}
}

func TestUpdateWorkflow(t *testing.T) {
	t.Parallel()

	type test struct {
		name         string
		updater      persistence.WorkflowUpdater
		dbID         int
		wantWorkflow *types.Workflow
		wantErr      string
	}

	workflowUUID := uuid.New()
	startedAt := time.Now().Add(-time.Minute)
	completedAt := time.Now()

	for _, tt := range []test{
		{
			name: "Updates a Workflow",
			updater: func(w *types.Workflow) (*types.Workflow, error) {
				w.UUID = workflowUUID
				w.TemporalID = "Updated temporal-id"
				w.Type = enums.WorkflowTypeMoveAip
				w.Status = enums.WorkflowStatusDone
				w.StartedAt = startedAt
				w.CompletedAt = completedAt
				return w, nil
			},
			wantWorkflow: &types.Workflow{
				DBID:        1,
				UUID:        workflowUUID,
				TemporalID:  "Updated temporal-id",
				Type:        enums.WorkflowTypeMoveAip,
				Status:      enums.WorkflowStatusDone,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
			},
		},
		{
			name:    "Fails to update a Workflow (not found)",
			updater: func(w *types.Workflow) (*types.Workflow, error) { return w, nil },
			dbID:    1234,
			wantErr: "update workflow: db: workflow not found",
		},
		{
			name:    "Fails to update a Workflow (updater error)",
			updater: func(t *types.Workflow) (*types.Workflow, error) { return nil, errors.New("updater error") },
			wantErr: "update workflow: updater error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForWorkflowTests(t, ctx, entc)

			if tt.dbID == 0 {
				dbt := entc.Workflow.Create().
					SetUUID(uuid.New()).
					SetTemporalID("Previous temporal-id").
					SetType(enums.WorkflowTypeUnspecified).
					SetStatus(enums.WorkflowStatusInProgress).
					SetAipID(1).
					SaveX(ctx)

				tt.dbID = dbt.ID
			}

			workflow, err := c.UpdateWorkflow(ctx, tt.dbID, tt.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, workflow, tt.wantWorkflow)
		})
	}
}

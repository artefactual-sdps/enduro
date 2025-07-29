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

var wUUID = uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")

func initialDataForTaskTests(t *testing.T, ctx context.Context, entc *db.Client) {
	t.Helper()

	initialDataForWorkflowTests(t, ctx, entc)

	entc.Workflow.Create().
		SetUUID(wUUID).
		SetTemporalID("temporal-id").
		SetType(enums.WorkflowTypeMoveAip).
		SetStatus(enums.WorkflowStatusInProgress).
		SetAipID(1).
		ExecX(ctx)
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	type test struct {
		name     string
		task     *types.Task
		wantTask *db.Task
		wantErr  string
	}

	taskUUID := uuid.New()
	startedAt := time.Now().Add(-time.Minute)
	completedAt := time.Now()

	for _, tt := range []test{
		{
			name: "Creates a Task",
			task: &types.Task{
				UUID:         taskUUID,
				Name:         "task",
				Status:       enums.TaskStatusInProgress,
				StartedAt:    startedAt,
				CompletedAt:  completedAt,
				WorkflowDBID: 1,
			},
			wantTask: &db.Task{
				ID:          1,
				UUID:        taskUUID,
				Name:        "task",
				Status:      enums.TaskStatusInProgress,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				WorkflowID:  1,
			},
		},
		{
			name: "Fails to create a Task without Workflow ID",
			task: &types.Task{
				UUID:   taskUUID,
				Name:   "task",
				Status: enums.TaskStatusInProgress,
			},
			wantErr: "create task: db: workflow not found",
		},
		{
			name: "Fails to create a Task with an unknown Workflow ID",
			task: &types.Task{
				UUID:         taskUUID,
				Name:         "task",
				Status:       enums.TaskStatusInProgress,
				WorkflowDBID: 1234,
			},
			wantErr: "create task: db: workflow not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForTaskTests(t, ctx, entc)

			err := c.CreateTask(ctx, tt.task)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			dbt := entc.Task.GetX(ctx, tt.task.DBID)
			assert.DeepEqual(t, dbt, tt.wantTask, cmpopts.IgnoreFields(db.Task{}, "config", "Edges", "selectValues"))
		})
	}
}

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	type test struct {
		name     string
		updater  persistence.TaskUpdater
		dbID     int
		wantTask *types.Task
		wantErr  string
	}

	taskUUID := uuid.New()
	startedAt := time.Now().Add(-time.Minute)
	completedAt := time.Now()

	for _, tt := range []test{
		{
			name: "Updates a Task",
			updater: func(t *types.Task) (*types.Task, error) {
				t.UUID = taskUUID
				t.Name = "Updated name"
				t.Status = enums.TaskStatusDone
				t.StartedAt = startedAt
				t.CompletedAt = completedAt
				t.Note = "Updated note"
				return t, nil
			},
			wantTask: &types.Task{
				DBID:         1,
				UUID:         taskUUID,
				Name:         "Updated name",
				Status:       enums.TaskStatusDone,
				StartedAt:    startedAt,
				CompletedAt:  completedAt,
				Note:         "Updated note",
				WorkflowDBID: 1,
				WorkflowUUID: wUUID,
			},
		},
		{
			name:    "Fails to update a Task (not found)",
			updater: func(t *types.Task) (*types.Task, error) { return t, nil },
			dbID:    1234,
			wantErr: "update task: db: task not found",
		},
		{
			name:    "Fails to update a Task (updater error)",
			updater: func(t *types.Task) (*types.Task, error) { return nil, errors.New("updater error") },
			wantErr: "update task: updater error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			initialDataForTaskTests(t, ctx, entc)

			if tt.dbID == 0 {
				dbt := entc.Task.Create().
					SetUUID(uuid.New()).
					SetName("Previous name").
					SetStatus(enums.TaskStatusInProgress).
					SetNote("Previous note").
					SetWorkflowID(1).
					SaveX(ctx)

				tt.dbID = dbt.ID
			}

			task, err := c.UpdateTask(ctx, tt.dbID, tt.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, task, tt.wantTask)
		})
	}
}

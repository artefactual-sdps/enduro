package workflows

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestStorageMoveWorkflow(t *testing.T) {
	s := temporalsdk_testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	aipUUID := uuid.New()
	locationUUID := uuid.MustParse("e7452225-53d6-46f3-9f90-d0f2ee18b7cd")
	workflowDBID := 1
	workflow := &types.Workflow{
		UUID:       uuid.New(),
		TemporalID: "default-test-workflow-id",
		Type:       enums.WorkflowTypeMoveAip,
		Status:     enums.WorkflowStatusInProgress,
		StartedAt:  time.Now(),
		AIPUUID:    aipUUID,
	}
	copyTaskDBID := 1
	copyTask := &types.Task{
		UUID:         uuid.New(),
		Name:         "Copy AIP",
		Status:       enums.TaskStatusInProgress,
		StartedAt:    time.Now(),
		Note:         "Copying AIP to target location",
		WorkflowDBID: workflowDBID,
	}
	deleteTaskDBID := 2
	deleteTask := &types.Task{
		UUID:         uuid.New(),
		Name:         "Delete AIP",
		Status:       enums.TaskStatusInProgress,
		StartedAt:    time.Now(),
		Note:         "Deleting AIP from source location",
		WorkflowDBID: workflowDBID,
	}

	// Mock services and their expected calls
	// TODO: Move these to local activities tests and mock activities here.
	ctrl := gomock.NewController(t)
	storagesvc := fake.NewMockService(ctrl)
	storagesvc.EXPECT().DeleteAip(mockutil.Context(), aipUUID)
	storagesvc.EXPECT().UpdateAipLocationID(mockutil.Context(), aipUUID, locationUUID)
	storagesvc.EXPECT().UpdateAipStatus(mockutil.Context(), aipUUID, enums.AIPStatusMoving)
	storagesvc.EXPECT().UpdateAipStatus(mockutil.Context(), aipUUID, enums.AIPStatusStored)
	storagesvc.EXPECT().CreateWorkflow(
		mockutil.Context(),
		mockutil.Eq(
			workflow,
			cmpopts.IgnoreFields(types.Workflow{}, "UUID"),
			mockutil.EquateNearlySameTime(),
		),
	).DoAndReturn(func(ctx context.Context, w *types.Workflow) error {
		w.DBID = workflowDBID
		return nil
	})
	storagesvc.EXPECT().UpdateWorkflow(
		mockutil.Context(),
		workflowDBID,
		mockutil.Func(
			"Should update workflow fields",
			func(updater persistence.WorkflowUpdater) error {
				w, err := updater(&types.Workflow{})
				assert.NilError(t, err)
				assert.DeepEqual(t, w.Status, enums.WorkflowStatusDone)
				assert.DeepEqual(t, w.CompletedAt, time.Now(), mockutil.EquateNearlySameTime())
				return nil
			},
		),
	)
	storagesvc.EXPECT().CreateTask(
		mockutil.Context(),
		mockutil.Eq(
			copyTask,
			cmpopts.IgnoreFields(types.Task{}, "UUID"),
			mockutil.EquateNearlySameTime(),
		),
	).DoAndReturn(func(ctx context.Context, t *types.Task) error {
		t.DBID = copyTaskDBID
		return nil
	})
	storagesvc.EXPECT().UpdateTask(
		mockutil.Context(),
		copyTaskDBID,
		mockutil.Func(
			"Should update task fields",
			func(updater persistence.TaskUpdater) error {
				task, err := updater(&types.Task{})
				assert.NilError(t, err)
				assert.DeepEqual(t, task.Status, enums.TaskStatusDone)
				assert.DeepEqual(t, task.CompletedAt, time.Now(), mockutil.EquateNearlySameTime())
				assert.DeepEqual(t, task.Note, "AIP copied to target location")
				return nil
			},
		),
	)
	storagesvc.EXPECT().CreateTask(
		mockutil.Context(),
		mockutil.Eq(
			deleteTask,
			cmpopts.IgnoreFields(types.Task{}, "UUID"),
			mockutil.EquateNearlySameTime(),
		),
	).DoAndReturn(func(ctx context.Context, t *types.Task) error {
		t.DBID = deleteTaskDBID
		return nil
	})
	storagesvc.EXPECT().UpdateTask(
		mockutil.Context(),
		deleteTaskDBID,
		mockutil.Func(
			"Should update task fields",
			func(updater persistence.TaskUpdater) error {
				task, err := updater(&types.Task{})
				assert.NilError(t, err)
				assert.DeepEqual(t, task.Status, enums.TaskStatusDone)
				assert.DeepEqual(t, task.CompletedAt, time.Now(), mockutil.EquateNearlySameTime())
				assert.DeepEqual(t, task.Note, "AIP deleted from source location")
				return nil
			},
		),
	)

	// Worker activities
	env.RegisterActivityWithOptions(
		activities.NewCopyToPermanentLocationActivity(storagesvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: storage.CopyToPermanentLocationActivityName},
	)
	env.OnActivity(storage.CopyToPermanentLocationActivityName, mock.Anything, mock.Anything).Return(nil, nil)

	env.ExecuteWorkflow(
		NewStorageMoveWorkflow(storagesvc).Execute,
		storage.StorageMoveWorkflowRequest{
			AIPID:      aipUUID,
			LocationID: locationUUID,
			TaskQueue:  "global",
		},
	)

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowResult(nil)
	require.NoError(t, err)
}

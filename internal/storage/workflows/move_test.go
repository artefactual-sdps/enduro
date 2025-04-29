package workflows

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
)

func TestStorageMoveWorkflow(t *testing.T) {
	s := temporalsdk_testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	ctrl := gomock.NewController(t)
	storagesvc := fake.NewMockService(ctrl)

	env.RegisterActivityWithOptions(
		activities.NewCopyToPermanentLocationActivity(storagesvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: storage.CopyToPermanentLocationActivityName},
	)

	req := storage.StorageMoveWorkflowRequest{
		AIPID:      uuid.New(),
		LocationID: uuid.New(),
		TaskQueue:  "global",
	}
	workflowDBID := 1
	copyTaskDBID := 1
	deleteTaskDBID := 2

	env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  req.AIPID,
			Status: enums.AIPStatusProcessing,
		},
	).Return(nil)

	env.OnActivity(
		storage.CreateWorkflowLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CreateWorkflowLocalActivityParams{
			AIPID:      req.AIPID,
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeMoveAip,
		},
	).Return(workflowDBID, nil)

	env.OnActivity(
		storage.CreateTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CreateTaskLocalActivityParams{
			WorkflowDBID: workflowDBID,
			Status:       enums.TaskStatusInProgress,
			Name:         "Copy AIP",
			Note:         "Copying AIP to target location",
		},
	).Return(copyTaskDBID, nil)

	env.OnActivity(
		storage.CopyToPermanentLocationActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.CopyToPermanentLocationActivityParams{
			AIPID:      req.AIPID,
			LocationID: req.LocationID,
		},
	).Return(nil, nil)

	env.OnActivity(
		storage.CompleteTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CompleteTaskLocalActivityParams{
			DBID:   copyTaskDBID,
			Status: enums.TaskStatusDone,
			Note:   "AIP copied to target location",
		},
	).Return(nil)

	env.OnActivity(
		storage.CreateTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CreateTaskLocalActivityParams{
			WorkflowDBID: workflowDBID,
			Status:       enums.TaskStatusInProgress,
			Name:         "Delete AIP",
			Note:         "Deleting AIP from source location",
		},
	).Return(deleteTaskDBID, nil)

	env.OnActivity(
		storage.DeleteFromMinIOLocationLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.DeleteFromMinIOLocationLocalActivityParams{
			AIPID: req.AIPID,
		},
	).Return(nil)

	env.OnActivity(
		storage.CompleteTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CompleteTaskLocalActivityParams{
			DBID:   deleteTaskDBID,
			Status: enums.TaskStatusDone,
			Note:   "AIP deleted from source location",
		},
	).Return(nil)

	env.OnActivity(
		storage.UpdateAIPLocationLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPLocationLocalActivityParams{
			AIPID:      req.AIPID,
			LocationID: req.LocationID,
		},
	).Return(nil)

	env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  req.AIPID,
			Status: enums.AIPStatusStored,
		},
	).Return(nil)

	env.OnActivity(
		storage.CompleteWorkflowLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CompleteWorkflowLocalActivityParams{
			DBID:   workflowDBID,
			Status: enums.WorkflowStatusDone,
		},
	).Return(nil)

	env.ExecuteWorkflow(NewStorageMoveWorkflow(storagesvc).Execute, req)

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowResult(nil)
	require.NoError(t, err)
}

package workflows

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.artefactual.dev/tools/ref"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestStorageDeleteWorkflow(t *testing.T) {
	s := temporalsdk_testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	ctrl := gomock.NewController(t)
	storagesvc := fake.NewMockService(ctrl)

	env.RegisterActivityWithOptions(
		activities.NewDeleteFromAMSSLocationActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: storage.DeleteFromAMSSLocationActivityName},
	)

	req := storage.StorageDeleteWorkflowRequest{
		AIPID:     uuid.New(),
		Reason:    "Reason",
		UserEmail: "requester@example.com",
		UserSub:   "subject",
		UserISS:   "issuer",
		TaskQueue: "global",
	}
	reviewSignal := storage.DeletionReviewedSignal{
		Approved:  true,
		UserEmail: "reviewer@example.com",
		UserISS:   "issuer",
		UserSub:   "subject-2",
	}
	aip := &goastorage.AIP{
		UUID:       req.AIPID,
		LocationID: ref.New(uuid.New()),
		Status:     enums.AIPStatusStored.String(),
	}
	workflowDBID := 1
	reviewTaskDBID := 1
	reviewTaskNote := fmt.Sprintf("An AIP deletion has been requested by %s. Reason:\n\n%s", req.UserEmail, req.Reason)
	deletionRequestDBID := 1
	deleteTaskDBID := 2
	locationInfo := &storage.ReadLocationInfoLocalActivityResult{
		Source: enums.LocationSourceAmss,
		Config: types.LocationConfig{Value: &types.AMSSConfig{
			URL:      "http://127.0.0.1:62081",
			Username: "test",
			APIKey:   "secret",
		}},
	}

	env.OnActivity(
		storage.ReadAIPLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		aip.UUID,
	).Return(aip, nil)

	env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  aip.UUID,
			Status: enums.AIPStatusProcessing,
		},
	).Return(nil)

	env.OnActivity(
		storage.CreateWorkflowLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CreateWorkflowLocalActivityParams{
			AIPID:      aip.UUID,
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeDeleteAip,
		},
	).Return(workflowDBID, nil)

	env.OnActivity(
		storage.CreateTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CreateTaskLocalActivityParams{
			WorkflowDBID: workflowDBID,
			Status:       enums.TaskStatusPending,
			Name:         "Review AIP deletion request",
			Note:         fmt.Sprintf("%s\n\nAwaiting user review.", reviewTaskNote),
		},
	).Return(reviewTaskDBID, nil)

	env.OnActivity(
		storage.CreateDeletionRequestLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CreateDeletionRequestLocalActivityParams{
			Requester:    req.UserEmail,
			RequesterISS: req.UserISS,
			RequesterSub: req.UserSub,
			Reason:       req.Reason,
			WorkflowDBID: workflowDBID,
			AIPUUID:      req.AIPID,
		},
	).Return(deletionRequestDBID, nil)

	env.OnActivity(
		storage.UpdateWorkflowStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateWorkflowStatusLocalActivityParams{
			DBID:   workflowDBID,
			Status: enums.WorkflowStatusPending,
		},
	).Return(nil)

	env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  aip.UUID,
			Status: enums.AIPStatusPending,
		},
	).Return(nil)

	env.RegisterDelayedCallback(
		func() {
			env.SignalWorkflow(storage.DeletionReviewedSignalName, reviewSignal)
		},
		0,
	)

	env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  aip.UUID,
			Status: enums.AIPStatusProcessing,
		},
	).Return(nil)

	env.OnActivity(
		storage.UpdateWorkflowStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateWorkflowStatusLocalActivityParams{
			DBID:   workflowDBID,
			Status: enums.WorkflowStatusInProgress,
		},
	).Return(nil)

	env.OnActivity(
		storage.ReviewDeletionRequestLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		deletionRequestDBID,
		reviewSignal,
	).Return(nil)

	env.OnActivity(
		storage.CompleteTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CompleteTaskLocalActivityParams{
			DBID:   reviewTaskDBID,
			Status: enums.TaskStatusDone,
			Note:   fmt.Sprintf("%s\n\nAIP deletion request approved by %s.", reviewTaskNote, reviewSignal.UserEmail),
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
			Note:         "Deleting AIP",
		},
	).Return(deleteTaskDBID, nil)

	env.OnActivity(
		storage.ReadLocationInfoLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		*aip.LocationID,
	).Return(locationInfo, nil)

	env.OnActivity(
		storage.DeleteFromAMSSLocationActivityName,
		mock.AnythingOfType("*context.timerCtx"),
		&activities.DeleteFromAMSSLocationActivityParams{
			Config:  *locationInfo.Config.Value.(*types.AMSSConfig),
			AIPUUID: aip.UUID,
		},
	).Return(&activities.DeleteFromAMSSLocationActivityResult{Deleted: true}, nil)

	env.OnActivity(
		storage.CompleteTaskLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.CompleteTaskLocalActivityParams{
			DBID:   deleteTaskDBID,
			Status: enums.TaskStatusDone,
			Note:   fmt.Sprintf("AIP deleted from %s source location", strings.ToUpper(locationInfo.Source.String())),
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

	env.OnActivity(
		storage.UpdateAIPStatusLocalActivity,
		mock.AnythingOfType("*context.valueCtx"),
		storagesvc,
		&storage.UpdateAIPStatusLocalActivityParams{
			AIPID:  aip.UUID,
			Status: enums.AIPStatusDeleted,
		},
	).Return(nil)

	env.ExecuteWorkflow(NewStorageDeleteWorkflow(storagesvc).Execute, req)

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowResult(nil)
	require.NoError(t, err)
}

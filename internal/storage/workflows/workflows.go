package workflows

import (
	"time"

	"github.com/google/uuid"
	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

func localActivityOptions(ctx temporalsdk_workflow.Context) temporalsdk_workflow.Context {
	return temporalsdk_workflow.WithLocalActivityOptions(ctx, temporalsdk_workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})
}

func createWorkflow(
	ctx temporalsdk_workflow.Context,
	storagesvc storage.Service,
	aipID uuid.UUID,
	t enums.WorkflowType,
) (int, error) {
	var workflowDBID int
	activityOpts := localActivityOptions(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.CreateWorkflowLocalActivity,
		storagesvc,
		&storage.CreateWorkflowLocalActivityParams{
			AIPID:      aipID,
			TemporalID: temporalsdk_workflow.GetInfo(ctx).WorkflowExecution.ID,
			Type:       t,
		},
	).Get(activityOpts, &workflowDBID)
	if err != nil {
		return 0, err
	}

	return workflowDBID, nil
}

func completeWorkflow(
	ctx temporalsdk_workflow.Context,
	storagesvc storage.Service,
	dbID int,
	status enums.WorkflowStatus,
) error {
	activityOpts := localActivityOptions(ctx)
	return temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.CompleteWorkflowLocalActivity,
		storagesvc,
		&storage.CompleteWorkflowLocalActivityParams{
			DBID:   dbID,
			Status: status,
		},
	).Get(activityOpts, nil)
}

func createTask(
	ctx temporalsdk_workflow.Context,
	storagesvc storage.Service,
	dbID int,
	name, note string,
) (int, error) {
	var taskDBID int
	activityOpts := localActivityOptions(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.CreateTaskLocalActivity,
		storagesvc,
		&storage.CreateTaskLocalActivityParams{
			WorkflowDBID: dbID,
			Name:         name,
			Note:         note,
		},
	).Get(activityOpts, &taskDBID)
	if err != nil {
		return 0, err
	}

	return taskDBID, nil
}

func completeTask(
	ctx temporalsdk_workflow.Context,
	storagesvc storage.Service,
	dbID int,
	status enums.TaskStatus,
	note string,
) error {
	activityOpts := localActivityOptions(ctx)
	return temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.CompleteTaskLocalActivity,
		storagesvc,
		&storage.CompleteTaskLocalActivityParams{
			DBID:   dbID,
			Status: status,
			Note:   note,
		},
	).Get(activityOpts, nil)
}

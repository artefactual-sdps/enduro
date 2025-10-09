package workflow

import (
	"errors"
	"fmt"

	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/google/uuid"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func (w *ProcessingWorkflow) deleteOriginalSIP(ctx temporalsdk_workflow.Context, state *workflowState) error {
	// If retention period is negative, do nothing.
	if state.req.RetentionPeriod < 0 {
		return nil
	}

	// Create a "delete original SIP" task.
	id, err := w.createTask(
		ctx,
		&datatypes.Task{
			Name:         "Delete original SIP",
			Note:         fmt.Sprintf("The original SIP will be deleted in %s", state.req.RetentionPeriod.String()),
			Status:       enums.TaskStatusInProgress,
			WorkflowUUID: state.workflowUUID,
		},
	)
	if err != nil {
		return fmt.Errorf("create delete original SIP task: %v", err)
	}

	// Set the default (successful) delete original SIP task completion values.
	task := datatypes.Task{
		ID:     id,
		Status: enums.TaskStatusDone,
		Note:   "SIP successfully deleted",
	}

	// Set a timer for the retention period.
	if err := temporalsdk_workflow.Sleep(ctx, state.req.RetentionPeriod); err != nil {
		return fmt.Errorf("retention period timer failed: %v", err)
	}

	// Delete the original SIP based on its origin.
	activityOpts := withActivityOptsForRequest(ctx)
	if state.req.WatcherName != "" {
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.DeleteOriginalActivityName,
			state.req.WatcherName,
			state.req.Key,
		).Get(activityOpts, nil)
	} else if state.req.SIPSourceID != uuid.Nil {
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.DeleteOriginalFromSIPSourceActivityName,
			&bucketdelete.Params{Key: state.req.Key},
		).Get(activityOpts, nil)
	} else {
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.DeleteOriginalFromInternalBucketActivityName,
			&bucketdelete.Params{Key: state.req.Key},
		).Get(activityOpts, nil)
	}

	// Update task completion values on error.
	if err != nil {
		task.SystemError(
			"Original SIP deletion has failed.",
			"An error has occurred while attempting to delete the original SIP.",
		)
	}

	// Complete the delete original SIP task.
	if e := w.completeTask(ctx, task); e != nil {
		err = errors.Join(err, fmt.Errorf("complete delete original SIP task: %v", e))
	}

	return err
}

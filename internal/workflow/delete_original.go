package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/google/uuid"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

// deleteOriginalSIP deletes the original SIP from its source (watched
// location, SIP source or internal bucket) after the request's retention
// period, following a successful ingest.
func (w *ProcessingWorkflow) deleteOriginalSIP(ctx temporalsdk_workflow.Context, state *workflowState) error {
	return w.deleteOriginalSIPAfter(ctx, state, state.req.RetentionPeriod)
}

// deleteOriginalFailedSIP deletes the original SIP from its source after the
// request's failed retention period. It does nothing if the failed retention
// period is not set or is negative, or if the SIP is not in a watched location
// or SIP source. The caller must ensure a copy of the failed SIP has been
// stored before calling this method.
func (w *ProcessingWorkflow) deleteOriginalFailedSIP(ctx temporalsdk_workflow.Context, state *workflowState) error {
	period := state.req.FailedRetentionPeriod
	if period == nil || *period < 0 {
		return nil
	}

	// SIPs uploaded to the internal bucket are already moved to the failed
	// bucket and deleted from the internal bucket.
	if state.req.WatcherName == "" && state.req.SIPSourceID == uuid.Nil {
		return nil
	}

	return w.deleteOriginalSIPAfter(ctx, state, *period)
}

func (w *ProcessingWorkflow) deleteOriginalSIPAfter(
	ctx temporalsdk_workflow.Context,
	state *workflowState,
	retentionPeriod time.Duration,
) error {
	// If retention period is negative, do nothing.
	if retentionPeriod < 0 {
		return nil
	}

	// Create a "delete original SIP" task.
	id, err := w.createTask(
		ctx,
		&datatypes.Task{
			Name:         "Delete original SIP",
			Note:         fmt.Sprintf("The original SIP will be deleted in %s", retentionPeriod.String()),
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
	if err := temporalsdk_workflow.Sleep(ctx, retentionPeriod); err != nil {
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

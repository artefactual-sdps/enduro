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
// duration.
//
// If the SIP ingest was successful, the retention duration is determined by the
// `retentionPeriod` value; an unsuccessful ingest uses the
// `deleteFailedSIPAfter` value.
//
// A positive retention duration deletes the SIP after the specified time. A
// zero duration deletes the SIP immediately. If the duration is nil or
// negative, deleteOriginalSIP returns immediately and the SIP is not deleted.
func (w *ProcessingWorkflow) deleteOriginalSIP(ctx temporalsdk_workflow.Context, state *workflowState) error {
	switch state.status {
	case enums.WorkflowStatusDone:
		return w.deleteSuccessfulSIP(ctx, state)
	case enums.WorkflowStatusFailed, enums.WorkflowStatusError, enums.WorkflowStatusCanceled:
		return w.deleteFailedSIP(ctx, state)
	default:
		return nil
	}
}

// deleteSuccessfulSIP deletes the original SIP from its source (watched
// location, SIP source or internal bucket) after the request's retention
// period, following a successful ingest.
func (w *ProcessingWorkflow) deleteSuccessfulSIP(ctx temporalsdk_workflow.Context, state *workflowState) error {
	p := state.req.RetentionPeriod
	if p == nil || *p < 0 {
		return nil
	}

	return w.deleteOriginalSIPAfter(ctx, state, *p)
}

// deleteFailedSIP deletes the original SIP from its source (watched location or
// SIP source) after the `deleteFailedSIPAfter` period.
func (w *ProcessingWorkflow) deleteFailedSIP(ctx temporalsdk_workflow.Context, state *workflowState) error {
	p := state.req.DeleteFailedSIPAfter
	if p == nil || *p < 0 {
		return nil
	}

	// SIPs uploaded to the internal bucket are already moved to the failed
	// bucket and deleted from the internal bucket.
	if state.req.WatcherName == "" && state.req.SIPSourceID == uuid.Nil {
		return nil
	}

	return w.deleteOriginalSIPAfter(ctx, state, *p)
}

// deleteOriginalSIPAfter deletes the original SIP from its source (watched
// location, SIP source or internal bucket) after the specified retention
// period.
//
// If the retention period is zero, the SIP is deleted immediately. If the
// retention period is negative, deleteOriginalSIPAfter returns immediately
// and the SIP is not deleted.
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

	// Set a timer if retention period is greater than zero.
	if retentionPeriod > 0 {
		if err := temporalsdk_workflow.Sleep(ctx, retentionPeriod); err != nil {
			return fmt.Errorf("retention period timer failed: %v", err)
		}
	}

	// Delete the original SIP based on its origin.
	activityOpts := withActivityOptsForRequest(ctx)
	if state.req.WatcherName != "" {
		// Delete the original SIP from the watched location.
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.DeleteOriginalActivityName,
			state.req.WatcherName,
			state.req.Key,
		).Get(activityOpts, nil)
	} else if state.req.SIPSourceID != uuid.Nil {
		// Delete the original SIP from the SIP source.
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.DeleteOriginalFromSIPSourceActivityName,
			&bucketdelete.Params{Key: state.req.Key},
		).Get(activityOpts, nil)
	} else {
		// Delete the original SIP from the internal bucket.
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

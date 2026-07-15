package workflow

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/google/uuid"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func (w *ProcessingWorkflow) downloadSIP(sessCtx temporalsdk_workflow.Context, state *workflowState) error {
	// Create a "copy SIP" task.
	id, err := w.createTask(
		sessCtx,
		&datatypes.Task{
			Name:         "Copy SIP to workspace",
			Status:       enums.TaskStatusInProgress,
			WorkflowUUID: state.workflowUUID,
		},
	)
	if err != nil {
		return fmt.Errorf("create copy SIP task: %v", err)
	}

	// Set the default (successful) copy SIP task completion values.
	task := datatypes.Task{
		ID:     id,
		Status: enums.TaskStatusDone,
		Note:   "SIP successfully copied",
	}

	var destinationPath string
	if cfg := w.cfg.ChildWorkflows.ByType(enums.ChildWorkflowTypePreprocessing); cfg != nil {
		destinationPath = cfg.SharedPath
	}

	if state.req.WatcherName != "" {
		err = w.watcherDownload(sessCtx, state, destinationPath)
	} else {
		err = w.bucketDownload(sessCtx, state, destinationPath)
	}
	if err != nil {
		task.SystemError(
			"SIP copy has failed.",
			"An error has occurred while attempting to copy the SIP to the local workspace. Please try again, or ask a system administrator to investigate.",
		)
		state.status = enums.WorkflowStatusError
	} else {
		state.initialPath = state.sip.path
	}

	// Complete the copy SIP task.
	if e := w.completeTask(sessCtx, task); e != nil {
		return errors.Join(
			err,
			fmt.Errorf("complete copy SIP task: %v", e),
		)
	}

	if err != nil {
		return fmt.Errorf("download SIP: %v", err)
	}

	return nil
}

// Watcher request, use watcher bucket. The download activity will create a
// temporary directory to download the file to.
func (w *ProcessingWorkflow) watcherDownload(
	ctx temporalsdk_workflow.Context,
	state *workflowState,
	dest string,
) error {
	var res activities.DownloadActivityResult
	opts := withActivityOptsForLongLivedRequest(ctx)

	err := temporalsdk_workflow.ExecuteActivity(
		opts,
		activities.DownloadActivityName,
		&activities.DownloadActivityParams{
			Key:             state.req.Key,
			WatcherName:     state.req.WatcherName,
			DestinationPath: dest,
		},
	).Get(opts, &res)
	if err != nil {
		return fmt.Errorf("watcher download: %v", err)
	}

	state.sip.path = res.Path

	return nil
}

func (w *ProcessingWorkflow) bucketDownload(ctx temporalsdk_workflow.Context, state *workflowState, dest string) error {
	// If the destination path is set, the bucketdownload activity
	// will download directly there. We will create an extra directory
	// with the SIP UUID to avoid collisions and normalize the deletion
	// of the original SIP from the parent directory.
	if dest != "" {
		dest = filepath.Join(dest, state.sip.uuid.String())
	}

	var activityName string
	if state.req.SIPSourceID != uuid.Nil {
		// SIP source request, use SIP source bucket.
		// TODO: At some point there may be multiple SIP sources, so we
		// should use the source ID to determine which bucket to use.
		activityName = activities.DownloadFromSIPSourceActivityName
	} else {
		// API upload request, use internal bucket.
		activityName = activities.DownloadFromInternalBucketActivityName
	}

	var re bucketdownload.Result
	opts := withActivityOptsForLongLivedRequest(ctx)
	err := temporalsdk_workflow.ExecuteActivity(
		opts,
		activityName,
		&bucketdownload.Params{
			DirPath: dest,
			Key:     state.req.Key,
		},
	).Get(opts, &re)
	if err != nil {
		return fmt.Errorf("bucket download: %v", err)
	}

	state.sip.path = re.FilePath

	return nil
}

// Package workflow contains a Temporal workflow for ingesting and preserving
// SIPs using Archivematica or A3M.
package workflow

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketcopy"
	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalsdk_log "go.temporal.io/sdk/log"
	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/watcher"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
)

type ProcessingWorkflow struct {
	logger    temporalsdk_log.Logger
	cfg       config.Configuration
	rng       io.Reader
	ingestsvc ingest.Service
	wsvc      watcher.Service
}

func NewProcessingWorkflow(
	cfg config.Configuration,
	rng io.Reader,
	ingestsvc ingest.Service,
	wsvc watcher.Service,
) *ProcessingWorkflow {
	return &ProcessingWorkflow{
		cfg:       cfg,
		rng:       rng,
		ingestsvc: ingestsvc,
		wsvc:      wsvc,
	}
}

func (w *ProcessingWorkflow) cleanup(ctx temporalsdk_workflow.Context, state *workflowState) {
	w.logger.Debug("Cleaning up workflow state")

	// Set workflow status to "error" unless it completed successfully or failed
	// due to invalid contents.
	if state.status != enums.WorkflowStatusDone && state.status != enums.WorkflowStatusFailed {
		state.status = enums.WorkflowStatusError
	}

	// Set SIP status.
	switch state.status {
	case enums.WorkflowStatusDone:
		state.sip.status = enums.SIPStatusIngested
	case enums.WorkflowStatusFailed:
		state.sip.status = enums.SIPStatusFailed
	default:
		// Mark SIP as an error because something went wrong.
		state.sip.status = enums.SIPStatusError
	}

	// Determine if it failed as a SIP or as a PIP
	var failedAs enums.SIPFailedAs
	if state.sip.failed_key != "" {
		failedAs = enums.SIPFailedAsSIP
		if state.sip.transformed {
			failedAs = enums.SIPFailedAsPIP
		}
	}

	// Use a disconnected context so it also runs after cancellation.
	dctx, _ := temporalsdk_workflow.NewDisconnectedContext(ctx)
	activityOpts := withLocalActivityOpts(dctx)

	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		updateSIPLocalActivity,
		w.ingestsvc,
		&updateSIPLocalActivityParams{
			UUID:        state.sip.uuid,
			Name:        state.sip.name,
			AIPUUID:     state.aip.id,
			CompletedAt: temporalsdk_workflow.Now(dctx).UTC(),
			Status:      state.sip.status,
			FailedAs:    failedAs,
			FailedKey:   state.sip.failed_key,
		},
	).Get(activityOpts, nil)
	if err != nil {
		w.logger.Error("cleanup: error updating SIP status", "error", err.Error())
	}

	err = temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		completeWorkflowLocalActivity,
		w.ingestsvc,
		&completeWorkflowLocalActivityParams{
			WorkflowID:  state.workflowID,
			Status:      state.status,
			CompletedAt: temporalsdk_workflow.Now(dctx).UTC(),
		},
	).Get(activityOpts, nil)
	if err != nil {
		w.logger.Error("cleanup: error completing workflow", "error", err.Error())
	}
}

func (w *ProcessingWorkflow) sessionCleanup(ctx temporalsdk_workflow.Context, state *workflowState) {
	if state.status != enums.WorkflowStatusDone {
		if err := w.sendFailedToInternalBucket(ctx, state); err != nil {
			w.logger.Error(
				"session cleanup: error sending failed SIP/PIP to internal bucket",
				"error", err.Error(),
			)
		}
	}

	ctx = temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
		StartToCloseTimeout: time.Second,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	err := temporalsdk_workflow.ExecuteActivity(
		ctx,
		removepaths.Name,
		removepaths.Params{Paths: state.tempDirs},
	).Get(ctx, nil)
	if err != nil {
		w.logger.Error(
			"session cleanup: error(s) removing temporary directories",
			"errors", err.Error(),
		)
	}

	temporalsdk_workflow.CompleteSession(ctx)
}

// ProcessingWorkflow orchestrates all the activities related to the processing
// of a SIP in Archivematica, including is retrieval, creation of transfer,
// etc...
//
// Retrying this workflow would result in a new Archivematica transfer. We  do
// not have a retry policy in place. The user could trigger a new instance via
// the API.
func (w *ProcessingWorkflow) Execute(ctx temporalsdk_workflow.Context, req *ingest.ProcessingWorkflowRequest) error {
	w.logger = temporalsdk_workflow.GetLogger(ctx)

	// Create the initial workflow state.
	state := newWorkflowState(req)

	// Persist the SIP as early as possible if the request comes from a watcher.
	if state.req.WatcherName != "" {
		activityOpts := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(
			activityOpts,
			createSIPLocalActivity,
			w.ingestsvc,
			&createSIPLocalActivityParams{
				UUID:   state.sip.uuid,
				Name:   state.sip.name,
				Status: state.sip.status,
			},
		).Get(activityOpts, nil)
		if err != nil {
			return fmt.Errorf("error persisting SIP: %v", err)
		}
	}

	// Ensure that the status of the SIP and the workflow is always updated when
	// this function returns.
	defer w.cleanup(ctx, state)

	// Activities running within a session.
	{
		var sessErr error
		maxAttempts := 5

		activityOpts := temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
			TaskQueue:           w.cfg.Preservation.TaskQueue,
		})
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			sessCtx, err := temporalsdk_workflow.CreateSession(activityOpts, &temporalsdk_workflow.SessionOptions{
				CreationTimeout:  forever,
				ExecutionTimeout: forever,
			})
			if err != nil {
				return fmt.Errorf("error creating session: %v", err)
			}

			sessErr = w.SessionHandler(sessCtx, attempt, state)

			// We want to retry the session if it has been canceled as a result
			// of losing the worker but not otherwise. This scenario seems to be
			// identifiable when we have an error but the root context has not
			// been canceled.
			if sessErr != nil &&
				(errors.Is(sessErr, temporalsdk_workflow.ErrSessionFailed) || temporalsdk_temporal.IsCanceledError(sessErr)) {
				// Root context canceled, hence workflow canceled.
				if ctx.Err() == temporalsdk_workflow.ErrCanceled {
					return nil
				}

				w.logger.Error(
					"Session failed, will retry shortly (10s)...",
					"err", ctx.Err(),
					"attemptFailed", attempt,
					"attemptsLeft", maxAttempts-attempt,
				)

				_ = temporalsdk_workflow.Sleep(ctx, time.Second*10)

				continue
			}
			break
		}

		if sessErr != nil {
			return sessErr
		}
	}

	// Schedule deletion of the original in the watched data source.
	{
		if state.status == enums.WorkflowStatusDone {
			if req.RetentionPeriod != nil {
				err := temporalsdk_workflow.NewTimer(ctx, *req.RetentionPeriod).Get(ctx, nil)
				if err != nil {
					w.logger.Warn("Retention policy timer failed", "err", err.Error())
				} else {
					activityOpts := withActivityOptsForRequest(ctx)
					_ = temporalsdk_workflow.ExecuteActivity(
						activityOpts,
						activities.DeleteOriginalActivityName,
						req.WatcherName,
						req.Key,
					).Get(activityOpts, nil)
				}
			} else if req.CompletedDir != "" {
				activityOpts := withActivityOptsForLocalAction(ctx)
				_ = temporalsdk_workflow.ExecuteActivity(
					activityOpts,
					activities.DisposeOriginalActivityName,
					req.WatcherName,
					req.CompletedDir,
					req.Key,
				).Get(activityOpts, nil)
			}
		}
	}

	w.logger.Info(
		"Workflow completed successfully!",
		"SIPUUID", state.sip.uuid,
		"watcher", req.WatcherName,
		"key", req.Key,
		"name", state.sip.name,
		"status", state.status,
	)

	return nil
}

// SessionHandler runs activities that belong to the same session.
func (w *ProcessingWorkflow) SessionHandler(
	sessCtx temporalsdk_workflow.Context,
	attempt int,
	state *workflowState,
) error {
	// Cleanup session files on exit.
	defer w.sessionCleanup(sessCtx, state)

	sipStartedAt := temporalsdk_workflow.Now(sessCtx).UTC()

	// Set in-progress status.
	{
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(
			ctx,
			setStatusInProgressLocalActivity,
			w.ingestsvc,
			state.sip.uuid,
			sipStartedAt,
		).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Persist the workflow for the ingest workflow.
	{
		// TODO: Create deterministic UUIDs and make activities idempotent.
		state.workflowUUID = uuid.Must(uuid.NewRandomFromReader(w.rng))
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(
			ctx,
			createWorkflowLocalActivity,
			w.ingestsvc,
			&createWorkflowLocalActivityParams{
				UUID:       state.workflowUUID,
				TemporalID: temporalsdk_workflow.GetInfo(ctx).WorkflowExecution.ID,
				Type:       state.req.Type,
				Status:     enums.WorkflowStatusInProgress,
				StartedAt:  sipStartedAt,
				SIPUUID:    state.sip.uuid,
			},
		).Get(ctx, &state.workflowID)
		if err != nil {
			return err
		}
	}

	// Download.
	{
		var destinationPath string
		if w.cfg.Preprocessing.Enabled {
			destinationPath = w.cfg.Preprocessing.SharedPath
		}
		activityOpts := withActivityOptsForLongLivedRequest(sessCtx)

		if state.req.WatcherName != "" {
			// Watcher request, use watcher bucket.
			var downloadResult activities.DownloadActivityResult
			err := temporalsdk_workflow.ExecuteActivity(
				activityOpts,
				activities.DownloadActivityName,
				&activities.DownloadActivityParams{
					Key:             state.req.Key,
					WatcherName:     state.req.WatcherName,
					DestinationPath: destinationPath,
				},
			).Get(activityOpts, &downloadResult)
			if err != nil {
				return err
			}
			state.sip.path = downloadResult.Path
		} else {
			// API upload request, use internal bucket. If the destination path
			// is set, the bucketdownload activity will download directly there.
			// We will create an extra directory with the SIP UUID to avoid
			// collisions and normalize the deletion of the original SIP from
			// the parent directory.
			if destinationPath != "" {
				destinationPath = filepath.Join(destinationPath, state.sip.uuid.String())
			}
			var re bucketdownload.Result
			err := temporalsdk_workflow.ExecuteActivity(
				activityOpts,
				bucketdownload.Name,
				&bucketdownload.Params{
					DirPath: destinationPath,
					Key:     state.req.Key,
				},
			).Get(activityOpts, &re)
			if err != nil {
				return err
			}
			state.sip.path = re.FilePath
		}

		state.initialPath = state.sip.path

		// Delete download tmp dir when session ends.
		state.addTempPath(filepath.Dir(state.sip.path))
	}

	// Get SIP file extension.
	if state.sip.extension == "" && !state.sip.isDir {
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		var result activities.GetSIPExtensionActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.GetSIPExtensionActivityName,
			&activities.GetSIPExtensionActivityParams{Path: state.sip.path},
		).Get(activityOpts, &result)
		if err != nil {
			switch err {
			case activities.ErrInvalidArchive:
				// Not an archive file, bundle the source file as-is.
			default:
				return temporal_tools.NewNonRetryableError(err)
			}
		} else {
			state.sip.extension = result.Extension
		}
	}

	// Unarchive the transfer if it's not a directory and it's not part of the preprocessing child workflow.
	if !state.sip.isDir && (!w.cfg.Preprocessing.Enabled || !w.cfg.Preprocessing.Extract) {
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		var result archiveextract.Result
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			archiveextract.Name,
			&archiveextract.Params{SourcePath: state.sip.path},
		).Get(activityOpts, &result)
		if err != nil {
			switch err {
			case archiveextract.ErrInvalidArchive:
				// Not an archive file, bundle the source file as-is.
			default:
				return temporal_tools.NewNonRetryableError(err)
			}
		} else {
			state.sip.path = result.ExtractPath
			state.sip.isDir = true
		}
	}

	// Preprocessing child workflow.
	if err := w.preprocessing(sessCtx, state); err != nil {
		return err
	}

	// Classify the SIP.
	{
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		var result activities.ClassifySIPActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.ClassifySIPActivityName,
			activities.ClassifySIPActivityParams{Path: state.sip.path},
		).Get(activityOpts, &result)
		if err != nil {
			return fmt.Errorf("classify SIP: %v", err)
		}

		state.sip.sipType = result.Type
	}

	// Stop the workflow if preprocessing returned a SIP path that is not a valid bag.
	if state.sip.sipType != enums.SIPTypeBagIt && w.cfg.Preprocessing.Enabled {
		return errors.New("preprocessing returned a path that is not a valid bag")
	}

	// If the SIP is a BagIt Bag, validate it.
	if state.sip.sipType == enums.SIPTypeBagIt {
		id, err := w.createTask(
			sessCtx,
			&datatypes.Task{
				Name:         "Validate Bag",
				Status:       enums.TaskStatusInProgress,
				WorkflowUUID: state.workflowUUID,
			},
		)
		if err != nil {
			return fmt.Errorf("create validate bag task: %v", err)
		}

		// Set the default (successful) validate bag task completion values.
		task := datatypes.Task{
			ID:     id,
			Status: enums.TaskStatusDone,
			Note:   "Bag successfully validated",
		}

		// Validate the bag.
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		var result bagvalidate.Result
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			bagvalidate.Name,
			&bagvalidate.Params{Path: state.sip.path},
		).Get(activityOpts, &result)
		if err != nil {
			task.SystemError(
				"SIP bag validation has failed.",
				"An error has occurred while attempting to validate the SIP bag. Please try again, or ask a system administrator to investigate.",
			)
			state.status = enums.WorkflowStatusError
		}
		if !result.Valid {
			task.Failed(
				"SIP bag validation has failed.",
				result.Error,
				"Please ensure the bag is well-formed before reattempting ingest.",
			)

			// Fail the workflow with an error for now. In the future we may
			// want to stop the workflow without returning an error, but this
			// will require some changes to clean up appropriately (e.g. move
			// the failed SIP/PIP to the internal bucket).
			state.status = enums.WorkflowStatusFailed
			err = errors.New(result.Error)
		}

		// Update the validate bag task.
		if e := w.completeTask(sessCtx, task); e != nil {
			return errors.Join(
				err,
				fmt.Errorf("complete validate bag task: %v", e),
			)
		}

		if err != nil {
			return fmt.Errorf("validate bag: %v", err)
		}
	}

	// Do preservation.
	{
		var err error
		if w.cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
			err = w.transferAM(sessCtx, state)
		} else {
			err = w.transferA3m(sessCtx, state)
		}
		if err != nil {
			return err
		}
	}

	// Persist the SIP adding the AIP UUID.
	{
		activityOpts := withLocalActivityOpts(sessCtx)
		_ = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, updateSIPLocalActivity, w.ingestsvc, &updateSIPLocalActivityParams{
			UUID:    state.sip.uuid,
			Name:    state.sip.name,
			AIPUUID: state.aip.id,
			Status:  enums.SIPStatusProcessing,
		}).
			Get(activityOpts, nil)
	}

	// Stop here for the Archivematica workflow.
	if w.cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
		// Set status to done so it's considered in the session cleanup.
		state.status = enums.WorkflowStatusDone

		return nil
	}

	// Identifier of the task for upload to AIPs bucket.
	var uploadTaskID int

	// Add task for upload to review bucket.
	if state.req.Type == enums.WorkflowTypeCreateAndReviewAip {
		id, err := w.createTask(
			sessCtx,
			&datatypes.Task{
				Name:         "Move AIP",
				Status:       enums.TaskStatusInProgress,
				Note:         "Moving to review bucket",
				WorkflowUUID: state.workflowUUID,
			},
		)
		if err != nil {
			return err
		}
		uploadTaskID = id
	}

	// Upload AIP to MinIO.
	{
		activityOpts := temporalsdk_workflow.WithActivityOptions(sessCtx, temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour * 24,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2,
				MaximumAttempts:    3,
			},
		})
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.UploadActivityName, &activities.UploadActivityParams{
			AIPPath: state.aip.path,
			AIPID:   state.aip.id,
			Name:    state.sip.name,
		}).
			Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Complete task for upload to review bucket.
	if state.req.Type == enums.WorkflowTypeCreateAndReviewAip {
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completeTaskLocalActivity, w.ingestsvc, &completeTaskLocalActivityParams{
			ID:          uploadTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: temporalsdk_workflow.Now(sessCtx).UTC(),
			Note:        ref.New("Moved to review bucket"),
		}).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	var reviewResult *ingest.ReviewPerformedSignal

	// Identifier of the task for SIP/AIP review
	var reviewTaskID int

	if state.req.Type == enums.WorkflowTypeCreateAip {
		reviewResult = &ingest.ReviewPerformedSignal{
			Accepted:   true,
			LocationID: &w.cfg.Storage.DefaultPermanentLocationID,
		}
	} else {
		// Set SIP to pending status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.ingestsvc, state.sip.uuid, enums.SIPStatusPending).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Set workflow to pending status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setWorkflowStatusLocalActivity, w.ingestsvc, state.workflowID, enums.WorkflowStatusPending).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Add task for SIP/AIP review
		{
			id, err := w.createTask(
				sessCtx,
				&datatypes.Task{
					Name:         "Review AIP",
					Status:       enums.TaskStatusPending,
					Note:         "Awaiting user decision",
					WorkflowUUID: state.workflowUUID,
				},
			)
			if err != nil {
				return err
			}
			reviewTaskID = id
		}

		reviewResult = w.waitForReview(sessCtx)

		// Set SIP to in progress status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.ingestsvc, state.sip.uuid, enums.SIPStatusProcessing).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Set workflow to in progress status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setWorkflowStatusLocalActivity, w.ingestsvc, state.workflowID, enums.WorkflowStatusInProgress).Get(ctx, nil)
			if err != nil {
				return err
			}
		}
	}

	reviewCompletedAt := temporalsdk_workflow.Now(sessCtx).UTC()

	if reviewResult.Accepted {
		// Record SIP/AIP confirmation in review task.
		if state.req.Type == enums.WorkflowTypeCreateAndReviewAip {
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completeTaskLocalActivity, w.ingestsvc, &completeTaskLocalActivityParams{
				ID:          reviewTaskID,
				Status:      enums.TaskStatusDone,
				CompletedAt: reviewCompletedAt,
				Note:        ref.New("Reviewed and accepted"),
			}).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Identifier of the task for permanent storage move.
		var moveTaskID int

		// Add task for permanent storage move.
		{
			id, err := w.createTask(
				sessCtx,
				&datatypes.Task{
					Name:         "Move AIP",
					Status:       enums.TaskStatusInProgress,
					Note:         "Moving to permanent storage",
					WorkflowUUID: state.workflowUUID,
				},
			)
			if err != nil {
				return err
			}
			moveTaskID = id
		}

		// Move AIP to permanent storage
		{
			activityOpts := withActivityOptsForRequest(sessCtx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.MoveToPermanentStorageActivityName, &activities.MoveToPermanentStorageActivityParams{
				AIPID:      state.aip.id,
				LocationID: *reviewResult.LocationID,
			}).
				Get(activityOpts, nil)
			if err != nil {
				return err
			}
		}

		// Poll AIP move to permanent storage
		{
			activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.PollMoveToPermanentStorageActivityName, &activities.PollMoveToPermanentStorageActivityParams{
				AIPID: state.aip.id,
			}).
				Get(activityOpts, nil)
			if err != nil {
				return err
			}
		}

		// Complete task for permanent storage move.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completeTaskLocalActivity, w.ingestsvc, &completeTaskLocalActivityParams{
				ID:          moveTaskID,
				Status:      enums.TaskStatusDone,
				CompletedAt: temporalsdk_workflow.Now(sessCtx).UTC(),
				Note:        ref.New(fmt.Sprintf("Moved to location %s", *reviewResult.LocationID)),
			}).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		if err := w.poststorage(sessCtx, state.aip.id); err != nil {
			return err
		}
	} else if state.req.Type == enums.WorkflowTypeCreateAndReviewAip {
		// Record SIP/AIP rejection in review task
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completeTaskLocalActivity, w.ingestsvc, &completeTaskLocalActivityParams{
				ID:          reviewTaskID,
				Status:      enums.TaskStatusDone,
				CompletedAt: reviewCompletedAt,
				Note:        ref.New("Reviewed and rejected"),
			}).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Reject SIP
		{
			activityOpts := withActivityOptsForRequest(sessCtx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.RejectSIPActivityName, &activities.RejectSIPActivityParams{
				AIPID: state.aip.id,
			}).Get(activityOpts, nil)
			if err != nil {
				return err
			}
		}
	}

	// Set status to done so it's considered in the session cleanup.
	state.status = enums.WorkflowStatusDone

	return nil
}

func (w *ProcessingWorkflow) waitForReview(ctx temporalsdk_workflow.Context) *ingest.ReviewPerformedSignal {
	var review ingest.ReviewPerformedSignal
	signalChan := temporalsdk_workflow.GetSignalChannel(ctx, ingest.ReviewPerformedSignalName)
	selector := temporalsdk_workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel temporalsdk_workflow.ReceiveChannel, more bool) {
		_ = channel.Receive(ctx, &review)
	})
	selector.Select(ctx)
	return &review
}

func (w *ProcessingWorkflow) transferA3m(
	sessCtx temporalsdk_workflow.Context,
	state *workflowState,
) error {
	// Bundle SIP as an Archivematica standard transfer.
	{
		activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
		var bundleResult activities.BundleActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.BundleActivityName,
			&activities.BundleActivityParams{
				SourcePath:  state.sip.path,
				TransferDir: w.cfg.A3m.ShareDir,
				IsDir:       state.sip.isDir,
			},
		).Get(activityOpts, &bundleResult)
		if err != nil {
			return err
		}

		state.sip.path = bundleResult.FullPath
		state.sip.isDir = true
		state.sip.sipType = enums.SIPTypeArchivematicaStandardTransfer
		state.sip.transformed = true

		// Delete bundled transfer when session ends.
		state.addTempPath(bundleResult.FullPath)
	}

	err := w.validatePREMIS(
		sessCtx,
		filepath.Join(state.sip.path, "metadata", "premis.xml"),
		state.workflowUUID,
	)
	if err != nil {
		return err
	}

	// Send PIP to a3m for preservation.
	{
		activityOpts := temporalsdk_workflow.WithActivityOptions(sessCtx, temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour * 24,
			HeartbeatTimeout:    time.Second * 5,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				MaximumAttempts: 1,
			},
		})

		params := &a3m.CreateAIPActivityParams{
			Name:         state.sip.name,
			Path:         state.sip.path,
			WorkflowUUID: state.workflowUUID,
		}

		result := a3m.CreateAIPActivityResult{}
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, a3m.CreateAIPActivityName, params).
			Get(sessCtx, &result)
		if err != nil {
			return err
		}

		state.aip = &aipInfo{
			id:   result.UUID,
			path: result.Path,
		}
	}

	return nil
}

func (w *ProcessingWorkflow) transferAM(
	ctx temporalsdk_workflow.Context,
	state *workflowState,
) error {
	var err error

	// Bag the SIP if it's not already a bag.
	if state.sip.sipType != enums.SIPTypeBagIt {
		lctx := withActivityOptsForLocalAction(ctx)
		var result bagcreate.Result
		err = temporalsdk_workflow.ExecuteActivity(
			lctx,
			bagcreate.Name,
			&bagcreate.Params{SourcePath: state.sip.path},
		).Get(lctx, &result)
		if err != nil {
			return err
		}
		state.sip.isDir = true
		state.sip.sipType = enums.SIPTypeBagIt
		state.sip.transformed = true
	}

	err = w.validatePREMIS(
		ctx,
		filepath.Join(state.sip.path, "data", "metadata", "premis.xml"),
		state.workflowUUID,
	)
	if err != nil {
		return err
	}

	// Zip PIP, if necessary.
	if w.cfg.AM.ZipPIP {
		activityOpts := withActivityOptsForLocalAction(ctx)
		var zipResult archivezip.Result
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			archivezip.Name,
			&archivezip.Params{SourceDir: state.sip.path},
		).Get(activityOpts, &zipResult)
		if err != nil {
			return err
		}

		state.sip.path = zipResult.Path
		state.sip.isDir = false

		// Delete the zipped PIP when the workflow completes.
		state.addTempPath(zipResult.Path)
	}

	// Upload the PIP to AMSS.
	activityOpts := temporalsdk_workflow.WithActivityOptions(ctx,
		temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour * 2,
			HeartbeatTimeout:    2 * w.cfg.AM.PollInterval,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				InitialInterval:    time.Second * 5,
				BackoffCoefficient: 2,
				MaximumAttempts:    3,
				NonRetryableErrorTypes: []string{
					"TemporalTimeout:StartToClose",
				},
			},
		},
	)
	uploadResult := am.UploadTransferActivityResult{}
	err = temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		am.UploadTransferActivityName,
		&am.UploadTransferActivityParams{SourcePath: state.sip.path},
	).Get(activityOpts, &uploadResult)
	if err != nil {
		return err
	}

	// Start AM transfer.
	activityOpts = withActivityOptsForRequest(ctx)
	transferResult := am.StartTransferActivityResult{}
	err = temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		am.StartTransferActivityName,
		&am.StartTransferActivityParams{
			Name:         state.sip.name,
			RelativePath: uploadResult.RemoteRelativePath,
			ZipPIP:       w.cfg.AM.ZipPIP,
		},
	).Get(activityOpts, &transferResult)
	if err != nil {
		return err
	}

	pollOpts := temporalsdk_workflow.WithActivityOptions(
		ctx,
		temporalsdk_workflow.ActivityOptions{
			HeartbeatTimeout:    2 * w.cfg.AM.PollInterval,
			StartToCloseTimeout: w.cfg.AM.TransferDeadline,
			RetryPolicy: &temporalsdk_temporal.RetryPolicy{
				InitialInterval:    5 * time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    5,
			},
		},
	)

	// Poll transfer status.
	var pollTransferResult am.PollTransferActivityResult
	err = temporalsdk_workflow.ExecuteActivity(
		pollOpts,
		am.PollTransferActivityName,
		am.PollTransferActivityParams{
			WorkflowUUID: state.workflowUUID,
			TransferID:   transferResult.TransferID,
		},
	).Get(pollOpts, &pollTransferResult)
	if err != nil {
		return err
	}

	// Set AIP id to Archivematica SIP ID.
	state.aip.id = pollTransferResult.SIPID

	// Poll ingest status.
	var pollIngestResult am.PollIngestActivityResult
	err = temporalsdk_workflow.ExecuteActivity(
		pollOpts,
		am.PollIngestActivityName,
		am.PollIngestActivityParams{
			WorkflowUUID: state.workflowUUID,
			SIPID:        state.aip.id,
		},
	).Get(pollOpts, &pollIngestResult)
	if err != nil {
		return err
	}

	// Create storage AIP record and set location to AMSS location.
	{
		activityOpts := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.CreateStorageAIPActivityName,
			&activities.CreateStorageAIPActivityParams{
				Name:       state.sip.name,
				AIPID:      state.aip.id,
				ObjectKey:  state.aip.id,
				LocationID: &w.cfg.Storage.DefaultPermanentLocationID,
				Status:     "stored",
			}).
			Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	if err := w.poststorage(ctx, state.aip.id); err != nil {
		return err
	}

	// Delete the PIP from the Archivematica transfer source directory.
	activityOpts = withActivityOptsForRequest(ctx)
	err = temporalsdk_workflow.ExecuteActivity(activityOpts, am.DeleteTransferActivityName, am.DeleteTransferActivityParams{
		Destination: uploadResult.RemoteRelativePath,
	}).
		Get(activityOpts, nil)
	if err != nil {
		return err
	}

	return nil
}

func (w *ProcessingWorkflow) preprocessing(ctx temporalsdk_workflow.Context, state *workflowState) error {
	if !w.cfg.Preprocessing.Enabled {
		return nil
	}

	// TODO: move SIP if sip.Path is not inside w.cfg.Preprocessing.SharedPath.
	relPath, err := filepath.Rel(w.cfg.Preprocessing.SharedPath, state.sip.path)
	if err != nil {
		return err
	}

	preCtx := temporalsdk_workflow.WithChildOptions(ctx, temporalsdk_workflow.ChildWorkflowOptions{
		Namespace:         w.cfg.Preprocessing.Temporal.Namespace,
		TaskQueue:         w.cfg.Preprocessing.Temporal.TaskQueue,
		WorkflowID:        fmt.Sprintf("%s-%s", w.cfg.Preprocessing.Temporal.WorkflowName, state.sip.uuid.String()),
		ParentClosePolicy: temporalapi_enums.PARENT_CLOSE_POLICY_TERMINATE,
	})
	var ppResult preprocessing.WorkflowResult
	err = temporalsdk_workflow.ExecuteChildWorkflow(
		preCtx,
		w.cfg.Preprocessing.Temporal.WorkflowName,
		preprocessing.WorkflowParams{RelativePath: relPath},
	).Get(preCtx, &ppResult)
	if err != nil {
		return err
	}

	// Set SIP info from preprocessing result on success.
	if ppResult.Outcome == preprocessing.OutcomeSuccess {
		state.sip.path = filepath.Join(w.cfg.Preprocessing.SharedPath, filepath.Clean(ppResult.RelativePath))
		state.sip.isDir = true
		state.sip.transformed = true
	}

	// Save preprocessing task data.
	if len(ppResult.PreservationTasks) > 0 {
		opts := withLocalActivityOpts(ctx)
		var savePPTasksResult localact.SavePreprocessingTasksActivityResult
		err = temporalsdk_workflow.ExecuteLocalActivity(
			opts,
			localact.SavePreprocessingTasksActivity,
			localact.SavePreprocessingTasksActivityParams{
				Ingestsvc:    w.ingestsvc,
				RNG:          w.rng,
				WorkflowUUID: state.workflowUUID,
				Tasks:        ppResult.PreservationTasks,
			},
		).Get(opts, &savePPTasksResult)
		if err != nil {
			return err
		}
	}

	switch ppResult.Outcome {
	case preprocessing.OutcomeSuccess:
		return nil
	case preprocessing.OutcomeSystemError:
		state.status = enums.WorkflowStatusError
		return errors.New("preprocessing workflow: system error")
	case preprocessing.OutcomeContentError:
		state.status = enums.WorkflowStatusFailed
		return errors.New("preprocessing workflow: validation failed")
	default:
		state.status = enums.WorkflowStatusError
		return fmt.Errorf("preprocessing workflow: unknown outcome %d", ppResult.Outcome)
	}
}

// poststorage executes the configured poststorage child workflows. It uses
// a disconnected context, abandon as parent close policy and only waits
// until the workflows are started, ignoring their results.
func (w *ProcessingWorkflow) poststorage(ctx temporalsdk_workflow.Context, aipUUID string) error {
	var err error
	disconnectedCtx, _ := temporalsdk_workflow.NewDisconnectedContext(ctx)

	for _, cfg := range w.cfg.Poststorage {
		psCtx := temporalsdk_workflow.WithChildOptions(
			disconnectedCtx,
			temporalsdk_workflow.ChildWorkflowOptions{
				Namespace:         cfg.Namespace,
				TaskQueue:         cfg.TaskQueue,
				WorkflowID:        fmt.Sprintf("%s-%s", cfg.WorkflowName, aipUUID),
				ParentClosePolicy: temporalapi_enums.PARENT_CLOSE_POLICY_ABANDON,
			},
		)
		err = errors.Join(
			err,
			temporalsdk_workflow.ExecuteChildWorkflow(
				psCtx,
				cfg.WorkflowName,
				poststorage.WorkflowParams{AIPUUID: aipUUID},
			).GetChildWorkflowExecution().Get(psCtx, nil),
		)
	}

	return err
}

func (w *ProcessingWorkflow) createTask(
	ctx temporalsdk_workflow.Context,
	task *datatypes.Task,
) (int, error) {
	// TODO: Create deterministic UUIDs and make activities idempotent.
	task.UUID = uuid.Must(uuid.NewRandomFromReader(w.rng))
	task.StartedAt = sql.NullTime{
		Time:  temporalsdk_workflow.Now(ctx).UTC(),
		Valid: true,
	}

	var id int
	ctx = withLocalActivityOpts(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		ctx,
		createTaskLocalActivity,
		&createTaskLocalActivityParams{
			Ingestsvc: w.ingestsvc,
			Task:      task,
		},
	).Get(ctx, &id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (w *ProcessingWorkflow) completeTask(
	ctx temporalsdk_workflow.Context,
	task datatypes.Task,
) error {
	ctx = withLocalActivityOpts(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		ctx,
		completeTaskLocalActivity,
		w.ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          task.ID,
			Status:      task.Status,
			CompletedAt: temporalsdk_workflow.Now(ctx).UTC(),
			Note:        ref.New(task.Note),
		},
	).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (w *ProcessingWorkflow) sendFailedToInternalBucket(
	sessCtx temporalsdk_workflow.Context,
	state *workflowState,
) error {
	// If the SIP has not been transformed, use the initial details;
	// otherwise, use the current values and treat it as a PIP.
	path := state.initialPath
	isDir := state.req.IsDir
	ext := state.sip.extension
	prefix := ingest.FailedSIPPrefix
	if state.sip.transformed {
		path = state.sip.path
		isDir = state.sip.isDir
		ext = ".zip"
		prefix = ingest.FailedPIPPrefix
	}

	state.sip.failed_key = fmt.Sprintf(
		"%s%s-%s%s",
		prefix,
		strings.TrimSuffix(state.sip.name, ext),
		state.sip.uuid.String(),
		ext,
	)

	// The SIP is already in the internal bucket.
	if state.req.WatcherName == "" && !state.sip.transformed {
		// Copy the SIP.
		activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			bucketcopy.Name,
			&bucketcopy.Params{
				SourceKey: state.req.Key,
				DestKey:   state.sip.failed_key,
			},
		).Get(activityOpts, nil)
		if err != nil {
			return err
		}

		// Delete the original SIP.
		activityOpts = withActivityOptsForRequest(sessCtx)
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			bucketdelete.Name,
			&bucketdelete.Params{Key: state.req.Key},
		).Get(activityOpts, nil)
		if err != nil {
			return err
		}

		return nil
	}

	if isDir {
		var zipResult archivezip.Result
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			archivezip.Name,
			&archivezip.Params{SourceDir: path},
		).Get(activityOpts, &zipResult)
		if err != nil {
			return err
		}
		path = zipResult.Path
	}

	activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
	err := temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		bucketupload.Name,
		&bucketupload.Params{
			Path:       path,
			Key:        state.sip.failed_key,
			BufferSize: 100_000_000,
		},
	).Get(activityOpts, nil)
	if err != nil {
		return err
	}

	return nil
}

func (w *ProcessingWorkflow) validatePREMIS(
	ctx temporalsdk_workflow.Context,
	xmlPath string,
	wUUID uuid.UUID,
) error {
	if !w.cfg.ValidatePREMIS.Enabled {
		return nil
	}

	// Create task for PREMIS validation.
	id, err := w.createTask(
		ctx,
		&datatypes.Task{
			Name:         "Validate PREMIS",
			Status:       enums.TaskStatusInProgress,
			WorkflowUUID: wUUID,
		},
	)
	if err != nil {
		return fmt.Errorf("create validate PREMIS task: %v", err)
	}

	// Set task default status and note.
	task := datatypes.Task{
		ID:     id,
		Status: enums.TaskStatusDone,
		Note:   "PREMIS is valid",
	}

	// Attempt to validate PREMIS.
	var xmlvalidateResult xmlvalidate.Result
	activityOpts := withActivityOptsForLocalAction(ctx)
	err = temporalsdk_workflow.ExecuteActivity(activityOpts, xmlvalidate.Name, xmlvalidate.Params{
		XMLPath: xmlPath,
		XSDPath: w.cfg.ValidatePREMIS.XSDPath,
	}).Get(activityOpts, &xmlvalidateResult)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("%s: no such file or directory", xmlPath)) {
			// Allow PREMIS XML to not exist without failing workflow.
			err = nil
		} else {
			task.Status = enums.TaskStatusError
			task.Note = "System error"
		}
	}

	// Set task status to error and add PREMIS validation failures to note.
	if len(xmlvalidateResult.Failures) > 0 {
		task.Status = enums.TaskStatusError
		task.Note = "PREMIS is invalid"

		for _, failure := range xmlvalidateResult.Failures {
			task.Note += fmt.Sprintf("\n%s", failure)
		}

		err = errors.New(task.Note)
	}

	// Mark task as complete.
	if e := w.completeTask(ctx, task); e != nil {
		return errors.Join(
			err,
			fmt.Errorf("complete validate PREMIS task: %v", e),
		)
	}

	if err != nil {
		return fmt.Errorf("validate PREMIS: %v", err)
	}

	return nil
}

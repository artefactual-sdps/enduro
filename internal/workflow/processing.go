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

	// Mark SIP as failed unless it completed successfully or it was
	// abandoned.
	if state.sip.status != enums.SIPStatusDone && state.sip.status != enums.SIPStatusAbandoned {
		state.sip.status = enums.SIPStatusError
	}

	// Use a disconnected context so it also runs after cancellation.
	dctx, _ := temporalsdk_workflow.NewDisconnectedContext(ctx)
	activityOpts := withLocalActivityOpts(dctx)

	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		updateSIPLocalActivity,
		w.ingestsvc,
		&updateSIPLocalActivityParams{
			SIPID:       state.sip.dbID,
			Name:        state.sip.name,
			AIPUUID:     state.aip.id,
			CompletedAt: state.aip.storedAt,
			Status:      state.sip.status,
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
		if err := w.sendToFailedBucket(ctx, state.sendToFailed, state.sip.name); err != nil {
			w.logger.Error(
				"session cleanup: error sending package to failed bucket",
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

	// Persist the SIP as early as possible.
	{
		activityOpts := withLocalActivityOpts(ctx)
		var err error

		if state.sip.dbID == 0 {
			err = temporalsdk_workflow.ExecuteLocalActivity(
				activityOpts,
				createSIPLocalActivity,
				w.ingestsvc,
				&createSIPLocalActivityParams{
					Name:   state.sip.name,
					Status: state.sip.status,
				},
			).Get(activityOpts, &state.sip.dbID)
		} else {
			// TODO: investigate better way to reset the ingest.
			err = temporalsdk_workflow.ExecuteLocalActivity(
				activityOpts,
				updateSIPLocalActivity,
				w.ingestsvc,
				&updateSIPLocalActivityParams{
					SIPID:       state.sip.dbID,
					Name:        state.sip.name,
					AIPUUID:     "",
					CompletedAt: temporalsdk_workflow.Now(ctx).UTC(),
					Status:      state.sip.status,
				},
			).Get(activityOpts, nil)
		}

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

		state.sip.status = enums.SIPStatusDone
		state.status = enums.WorkflowStatusDone
	}

	// Schedule deletion of the original in the watched data source.
	{
		if state.sip.status == enums.SIPStatusDone {
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
						state.sip.name,
					).Get(activityOpts, nil)
				}
			} else if req.CompletedDir != "" {
				activityOpts := withActivityOptsForLocalAction(ctx)
				_ = temporalsdk_workflow.ExecuteActivity(
					activityOpts,
					activities.DisposeOriginalActivityName,
					req.WatcherName,
					req.CompletedDir,
					state.sip.name,
				).Get(activityOpts, nil)
			}
		}
	}

	w.logger.Info(
		"Workflow completed successfully!",
		"SIPID", state.sip.dbID,
		"watcher", req.WatcherName,
		"name", state.sip.name,
		"status", state.sip.status.String(),
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
			state.sip.dbID,
			sipStartedAt,
		).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Persist the workflow for the ingest workflow.
	{
		{
			var workflowType enums.WorkflowType
			if state.req.AutoApproveAIP {
				workflowType = enums.WorkflowTypeCreateAip
			} else {
				workflowType = enums.WorkflowTypeCreateAndReviewAip
			}

			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, createWorkflowLocalActivity, w.ingestsvc, &createWorkflowLocalActivityParams{
				TemporalID: temporalsdk_workflow.GetInfo(ctx).WorkflowExecution.ID,
				Type:       workflowType,
				Status:     enums.WorkflowStatusInProgress,
				StartedAt:  sipStartedAt,
				SIPID:      state.sip.dbID,
			}).
				Get(ctx, &state.workflowID)
			if err != nil {
				return err
			}
		}
	}

	// Download.
	{
		var downloadResult activities.DownloadActivityResult
		activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
		params := &activities.DownloadActivityParams{
			Key:         state.sip.name,
			WatcherName: state.req.WatcherName,
		}
		if w.cfg.Preprocessing.Enabled {
			params.DestinationPath = w.cfg.Preprocessing.SharedPath
		}
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.DownloadActivityName, params).
			Get(activityOpts, &downloadResult)
		if err != nil {
			return err
		}
		state.sip.path = downloadResult.Path

		// Delete download tmp dir when session ends.
		state.addTempPath(filepath.Dir(downloadResult.Path))

		state.sendToFailed.path = downloadResult.Path
		state.sendToFailed.activityName = activities.SendToFailedSIPsName
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

	// After this point we treat the package as a PIP, as preprocessing may have
	// modified it.

	// Classify the PIP.
	{
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		var result activities.ClassifySIPActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.ClassifySIPActivityName,
			activities.ClassifySIPActivityParams{Path: state.pip.path},
		).Get(activityOpts, &result)
		if err != nil {
			return fmt.Errorf("classify PIP: %v", err)
		}

		state.pip.pipType = result.Type
	}

	// Stop the workflow if preprocessing returned a PIP path that is not a
	// valid bag.
	if state.pip.pipType != enums.SIPTypeBagIt && w.cfg.Preprocessing.Enabled {
		return errors.New("preprocessing returned a path that is not a valid bag")
	}

	// If the PIP is a BagIt Bag, validate it.
	if state.pip.isDir && state.pip.pipType == enums.SIPTypeBagIt {
		id, err := w.createTask(
			sessCtx,
			datatypes.Task{
				Name:       "Validate Bag",
				Status:     enums.TaskStatusInProgress,
				WorkflowID: state.workflowID,
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
			&bagvalidate.Params{Path: state.pip.path},
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
			// the failed SIP to "failed" bucket).
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

	// Persist AIP UUID and storedAt time.
	{
		activityOpts := withLocalActivityOpts(sessCtx)
		_ = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, updateSIPLocalActivity, w.ingestsvc, &updateSIPLocalActivityParams{
			SIPID:       state.sip.dbID,
			Name:        state.sip.name,
			AIPUUID:     state.aip.id,
			CompletedAt: state.aip.storedAt,
			Status:      enums.SIPStatusInProgress,
		}).
			Get(activityOpts, nil)
	}

	// Stop here for the Archivematica workflow.
	if w.cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
		return nil
	}

	// Identifier of the task for upload to AIPs bucket.
	var uploadTaskID int

	// Add task for upload to review bucket.
	if !state.req.AutoApproveAIP {
		id, err := w.createTask(
			sessCtx,
			datatypes.Task{
				Name:       "Move AIP",
				Status:     enums.TaskStatusInProgress,
				Note:       "Moving to review bucket",
				WorkflowID: state.workflowID,
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
	if !state.req.AutoApproveAIP {
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

	if state.req.AutoApproveAIP {
		reviewResult = &ingest.ReviewPerformedSignal{
			Accepted:   true,
			LocationID: state.req.DefaultPermanentLocationID,
		}
	} else {
		// Set SIP to pending status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.ingestsvc, state.sip.dbID, enums.SIPStatusPending).Get(ctx, nil)
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
				datatypes.Task{
					TaskID:     uuid.NewString(),
					Name:       "Review AIP",
					Status:     enums.TaskStatusPending,
					Note:       "Awaiting user decision",
					WorkflowID: state.workflowID,
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
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.ingestsvc, state.sip.dbID, enums.SIPStatusInProgress).Get(ctx, nil)
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
		if !state.req.AutoApproveAIP {
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
				datatypes.Task{
					Name:       "Move AIP",
					Status:     enums.TaskStatusInProgress,
					Note:       "Moving to permanent storage",
					WorkflowID: state.workflowID,
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
	} else if !state.req.AutoApproveAIP {
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
	// Bundle PIP as an Archivematica standard transfer.
	{
		activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
		var bundleResult activities.BundleActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.BundleActivityName,
			&activities.BundleActivityParams{
				SourcePath:  state.pip.path,
				TransferDir: w.cfg.A3m.ShareDir,
				IsDir:       state.pip.isDir,
			},
		).Get(activityOpts, &bundleResult)
		if err != nil {
			return err
		}

		state.pip.path = bundleResult.FullPath
		state.pip.pipType = enums.SIPTypeArchivematicaStandardTransfer

		state.sendToFailed.path = state.pip.path
		state.sendToFailed.activityName = activities.SendToFailedPIPsName
		state.sendToFailed.needsZipping = true

		// Delete bundled transfer when session ends.
		state.addTempPath(bundleResult.FullPath)
	}

	err := w.validatePREMIS(
		sessCtx,
		filepath.Join(state.pip.path, "metadata", "premis.xml"),
		state.workflowID,
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
			Name:       state.sip.name,
			Path:       state.pip.path,
			WorkflowID: state.workflowID,
		}

		result := a3m.CreateAIPActivityResult{}
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, a3m.CreateAIPActivityName, params).
			Get(sessCtx, &result)
		if err != nil {
			return err
		}

		state.aip = &aipInfo{
			id:       result.UUID,
			path:     result.Path,
			storedAt: temporalsdk_workflow.Now(sessCtx).UTC(),
		}
	}

	return nil
}

func (w *ProcessingWorkflow) transferAM(
	ctx temporalsdk_workflow.Context,
	state *workflowState,
) error {
	var err error

	// Bag the PIP if it's not already a bag.
	if state.pip.pipType != enums.SIPTypeBagIt {
		lctx := withActivityOptsForLocalAction(ctx)
		var result bagcreate.Result
		err = temporalsdk_workflow.ExecuteActivity(
			lctx,
			bagcreate.Name,
			&bagcreate.Params{SourcePath: state.pip.path},
		).Get(lctx, &result)
		if err != nil {
			return err
		}
		state.pip.pipType = enums.SIPTypeBagIt
	}

	err = w.validatePREMIS(
		ctx,
		filepath.Join(state.pip.path, "data", "metadata", "premis.xml"),
		state.workflowID,
	)
	if err != nil {
		return err
	}

	// Send the PIP to the "failed PIP" bucket if preservation fails.
	state.sendToFailed.path = state.pip.path
	state.sendToFailed.activityName = activities.SendToFailedPIPsName

	// Zip PIP, if necessary.
	if w.cfg.AM.ZipPIP {
		activityOpts := withActivityOptsForLocalAction(ctx)
		var zipResult archivezip.Result
		err = temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			archivezip.Name,
			&archivezip.Params{SourceDir: state.pip.path},
		).Get(activityOpts, &zipResult)
		if err != nil {
			return err
		}

		state.pip.path = zipResult.Path
		state.pip.isDir = false

		state.sendToFailed.path = zipResult.Path
		state.sendToFailed.needsZipping = false

		// Delete the zipped PIP when the workflow completes.
		state.addTempPath(zipResult.Path)
	}

	// Upload the PIP to AMSS.
	activityOpts := temporalsdk_workflow.WithActivityOptions(ctx,
		temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour * 2,
			HeartbeatTimeout:    2 * state.req.PollInterval,
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
		&am.UploadTransferActivityParams{SourcePath: state.pip.path},
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
			HeartbeatTimeout:    2 * state.req.PollInterval,
			StartToCloseTimeout: state.req.TransferDeadline,
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
			WorkflowID: state.workflowID,
			TransferID: transferResult.TransferID,
		},
	).Get(pollOpts, &pollTransferResult)
	if err != nil {
		return err
	}

	// Set PIP id to Archivematica SIP ID.
	state.pip.id = pollTransferResult.SIPID

	// Poll ingest status.
	var pollIngestResult am.PollIngestActivityResult
	err = temporalsdk_workflow.ExecuteActivity(
		pollOpts,
		am.PollIngestActivityName,
		am.PollIngestActivityParams{
			WorkflowID: state.workflowID,
			SIPID:      state.pip.id,
		},
	).Get(pollOpts, &pollIngestResult)
	if err != nil {
		return err
	}

	// Set AIP data.
	state.aip = &aipInfo{
		id:       state.pip.id,
		storedAt: temporalsdk_workflow.Now(ctx).UTC(),
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
				LocationID: state.req.DefaultPermanentLocationID,
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
		// Alias the PIP info to the SIP to have consistent calls going forward.
		state.pip.path = state.sip.path
		state.pip.isDir = state.sip.isDir

		return nil
	}

	// TODO: move SIP if sip.Path is not inside w.cfg.Preprocessing.SharedPath.
	relPath, err := filepath.Rel(w.cfg.Preprocessing.SharedPath, state.sip.path)
	if err != nil {
		return err
	}

	// TODO: Use SIP UUID instead SIPID when that field is added to the SIP.
	preCtx := temporalsdk_workflow.WithChildOptions(ctx, temporalsdk_workflow.ChildWorkflowOptions{
		Namespace:         w.cfg.Preprocessing.Temporal.Namespace,
		TaskQueue:         w.cfg.Preprocessing.Temporal.TaskQueue,
		WorkflowID:        fmt.Sprintf("%s-%d", w.cfg.Preprocessing.Temporal.WorkflowName, state.sip.dbID),
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

	// Set PIP info from preprocessing result.
	state.pip.path = filepath.Join(w.cfg.Preprocessing.SharedPath, filepath.Clean(ppResult.RelativePath))
	state.pip.isDir = true

	// Save preprocessing task data.
	if len(ppResult.PreservationTasks) > 0 {
		opts := withLocalActivityOpts(ctx)
		var savePPTasksResult localact.SavePreprocessingTasksActivityResult
		err = temporalsdk_workflow.ExecuteLocalActivity(
			opts,
			localact.SavePreprocessingTasksActivity,
			localact.SavePreprocessingTasksActivityParams{
				Ingestsvc:  w.ingestsvc,
				RNG:        w.rng,
				WorkflowID: state.workflowID,
				Tasks:      ppResult.PreservationTasks,
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
	task datatypes.Task,
) (int, error) {
	var id int
	ctx = withLocalActivityOpts(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		ctx,
		createTaskLocalActivity,
		&createTaskLocalActivityParams{
			Ingestsvc: w.ingestsvc,
			RNG:       w.rng,
			Task: datatypes.Task{
				Name:   task.Name,
				Status: task.Status,
				StartedAt: sql.NullTime{
					Time:  temporalsdk_workflow.Now(ctx).UTC(),
					Valid: true,
				},
				Note:       task.Note,
				WorkflowID: task.WorkflowID,
			},
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

func (w *ProcessingWorkflow) sendToFailedBucket(
	sessCtx temporalsdk_workflow.Context,
	stf sendToFailed,
	key string,
) error {
	if stf.path == "" || stf.activityName == "" {
		return nil
	}

	if stf.needsZipping {
		var zipResult archivezip.Result
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			archivezip.Name,
			&archivezip.Params{SourceDir: stf.path},
		).Get(activityOpts, &zipResult)
		if err != nil {
			return err
		}
		stf.path = zipResult.Path
	}

	var sendToFailedResult bucketupload.Result
	activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
	err := temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		stf.activityName,
		&bucketupload.Params{
			Path:       stf.path,
			Key:        fmt.Sprintf("Failed_%s", key),
			BufferSize: 100_000_000,
		},
	).Get(activityOpts, &sendToFailedResult)
	if err != nil {
		return err
	}

	return nil
}

func (w *ProcessingWorkflow) validatePREMIS(
	ctx temporalsdk_workflow.Context,
	xmlPath string,
	wID int,
) error {
	if !w.cfg.ValidatePREMIS.Enabled {
		return nil
	}

	// Create task for PREMIS validation.
	id, err := w.createTask(
		ctx,
		datatypes.Task{
			Name:       "Validate PREMIS",
			Status:     enums.TaskStatusInProgress,
			WorkflowID: wID,
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

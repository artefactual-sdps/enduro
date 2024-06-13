// Package workflow contains an experimental workflow for Archivemica transfers.
//
// It's not generalized since it contains client-specific activities. However,
// the long-term goal is to build a system where workflows and activities are
// dynamically set up based on user input.
package workflow

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archive"
	"github.com/artefactual-sdps/temporal-activities/filesys"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/fsutil"
	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/watcher"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
)

type ProcessingWorkflow struct {
	logger logr.Logger
	cfg    config.Configuration
	rng    io.Reader
	pkgsvc package_.Service
	wsvc   watcher.Service
}

func NewProcessingWorkflow(
	logger logr.Logger,
	cfg config.Configuration,
	rng io.Reader,
	pkgsvc package_.Service,
	wsvc watcher.Service,
) *ProcessingWorkflow {
	return &ProcessingWorkflow{
		logger: logger,
		cfg:    cfg,
		rng:    rng,
		pkgsvc: pkgsvc,
		wsvc:   wsvc,
	}
}

// TransferInfo is shared state that is passed down to activities. It can be
// useful for hooks that may require quick access to processing state.
type TransferInfo struct {
	// It is populated by the workflow request.
	req package_.ProcessingWorkflowRequest

	// IsDir indicates whether the current working copy of the transfer is a
	// filesystem directory.
	IsDir bool

	// TempPath is the temporary location of a working copy of the transfer.
	TempPath string

	// SIPID given by a3m.
	//
	// It is populated by CreateAIPActivity.
	SIPID string

	// Path to the compressed AIP generated by the preservation system.
	//
	// It is populated once the preservation system creates the AIP.
	AIPPath string

	// StoredAt is the time when the AIP is stored.
	//
	// It is populated by PollIngestActivity as long as Ingest completes.
	StoredAt time.Time

	// Information about the bundle (transfer) that we submit to Archivematica.
	// Full path, relative path...
	//
	// It is populated by BundleActivity.
	Bundle activities.BundleActivityResult

	// Identifier of the preservation action that creates the AIP
	//
	// It is populated by createPreservationActionLocalActivity .
	PreservationActionID uint

	// Identifier of the preservation system task queue name
	//
	// It is populated by the workflow request.
	GlobalTaskQueue       string
	PreservationTaskQueue string
}

func (t *TransferInfo) Name() string {
	return fsutil.BaseNoExt(t.req.Key)
}

// cleanupRegistry contains items that should be cleaned up when a workflow
// session completes.
type cleanupRegistry struct {
	// tempDirs are working directories registered for deletion during cleanup.
	tempDirs []string
}

// registerPath registers a filepath for deletion when a workflow session
// completes.
func (c *cleanupRegistry) registerPath(path string) {
	if path == "" || slices.Contains(c.tempDirs, path) {
		return
	}
	c.tempDirs = append(c.tempDirs, path)
}

func (w *ProcessingWorkflow) sessionCleanup(ctx temporalsdk_workflow.Context, cleanup *cleanupRegistry) {
	ctx = temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
		StartToCloseTimeout: time.Second,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	err := temporalsdk_workflow.ExecuteActivity(
		ctx,
		filesys.RemoveActivityName,
		filesys.RemoveActivityParams{Paths: cleanup.tempDirs},
	).Get(ctx, nil)
	if err != nil {
		w.logger.V(1).Info("session cleanup: error(s) removing temporary directories",
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
func (w *ProcessingWorkflow) Execute(ctx temporalsdk_workflow.Context, req *package_.ProcessingWorkflowRequest) error {
	var (
		logger = temporalsdk_workflow.GetLogger(ctx)

		tinfo = &TransferInfo{
			req:                   *req,
			IsDir:                 req.IsDir,
			GlobalTaskQueue:       req.GlobalTaskQueue,
			PreservationTaskQueue: req.PreservationTaskQueue,
		}

		// Package status. All packages start in queued status.
		status = enums.PackageStatusQueued

		// Create AIP preservation action status.
		paStatus = enums.PreservationActionStatusUnspecified
	)

	// Persist package as early as possible.
	{
		activityOpts := withLocalActivityOpts(ctx)
		var err error

		if req.PackageID == 0 {
			err = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, createPackageLocalActivity, w.logger, w.pkgsvc, &createPackageLocalActivityParams{
				Key:    req.Key,
				Status: status,
			}).
				Get(activityOpts, &tinfo.req.PackageID)
		} else {
			// TODO: investigate better way to reset the package_.
			err = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, updatePackageLocalActivity, w.logger, w.pkgsvc, &updatePackageLocalActivityParams{
				PackageID: req.PackageID,
				Key:       req.Key,
				SIPID:     "",
				StoredAt:  temporalsdk_workflow.Now(ctx).UTC(),
				Status:    status,
			}).Get(activityOpts, nil)
		}

		if err != nil {
			return fmt.Errorf("error persisting package: %v", err)
		}
	}

	// Ensure that the status of the package and the preservation action is always updated when this
	// workflow function returns.
	defer func() {
		// Mark as failed unless it completed successfully or it was abandoned.
		if status != enums.PackageStatusDone && status != enums.PackageStatusAbandoned {
			status = enums.PackageStatusError
		}

		// Use disconnected context so it also runs after cancellation.
		dctx, _ := temporalsdk_workflow.NewDisconnectedContext(ctx)
		activityOpts := withLocalActivityOpts(dctx)
		_ = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, updatePackageLocalActivity, w.logger, w.pkgsvc, &updatePackageLocalActivityParams{
			PackageID: tinfo.req.PackageID,
			Key:       tinfo.req.Key,
			SIPID:     tinfo.SIPID,
			StoredAt:  tinfo.StoredAt,
			Status:    status,
		}).
			Get(activityOpts, nil)

		if paStatus != enums.PreservationActionStatusDone {
			paStatus = enums.PreservationActionStatusError
		}

		_ = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, completePreservationActionLocalActivity, w.pkgsvc, &completePreservationActionLocalActivityParams{
			PreservationActionID: tinfo.PreservationActionID,
			Status:               paStatus,
			CompletedAt:          temporalsdk_workflow.Now(dctx).UTC(),
		}).
			Get(activityOpts, nil)
	}()

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

			sessErr = w.SessionHandler(sessCtx, attempt, tinfo)

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

				logger.Error(
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

		status = enums.PackageStatusDone

		paStatus = enums.PreservationActionStatusDone
	}

	// Schedule deletion of the original in the watched data source.
	{
		if status == enums.PackageStatusDone {
			if tinfo.req.RetentionPeriod != nil {
				err := temporalsdk_workflow.NewTimer(ctx, *tinfo.req.RetentionPeriod).Get(ctx, nil)
				if err != nil {
					logger.Warn("Retention policy timer failed", "err", err.Error())
				} else {
					activityOpts := withActivityOptsForRequest(ctx)
					_ = temporalsdk_workflow.ExecuteActivity(activityOpts, activities.DeleteOriginalActivityName, tinfo.req.WatcherName, tinfo.req.Key).Get(activityOpts, nil)
				}
			} else if tinfo.req.CompletedDir != "" {
				activityOpts := withActivityOptsForLocalAction(ctx)
				_ = temporalsdk_workflow.ExecuteActivity(activityOpts, activities.DisposeOriginalActivityName, tinfo.req.WatcherName, tinfo.req.CompletedDir, tinfo.req.Key).Get(activityOpts, nil)
			}
		}
	}

	logger.Info(
		"Workflow completed successfully!",
		"packageID", tinfo.req.PackageID,
		"watcher", tinfo.req.WatcherName,
		"key", tinfo.req.Key,
		"status", status.String(),
	)

	return nil
}

// SessionHandler runs activities that belong to the same session.
func (w *ProcessingWorkflow) SessionHandler(
	sessCtx temporalsdk_workflow.Context,
	attempt int,
	tinfo *TransferInfo,
) error {
	var cleanup cleanupRegistry
	defer w.sessionCleanup(sessCtx, &cleanup)

	packageStartedAt := temporalsdk_workflow.Now(sessCtx).UTC()

	// Set in-progress status.
	{
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusInProgressLocalActivity, w.pkgsvc, tinfo.req.PackageID, packageStartedAt).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Persist the preservation action for creating the AIP.
	{
		{
			var preservationActionType enums.PreservationActionType
			if tinfo.req.AutoApproveAIP {
				preservationActionType = enums.PreservationActionTypeCreateAIP
			} else {
				preservationActionType = enums.PreservationActionTypeCreateAndReviewAIP
			}

			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, createPreservationActionLocalActivity, w.pkgsvc, &createPreservationActionLocalActivityParams{
				WorkflowID: temporalsdk_workflow.GetInfo(ctx).WorkflowExecution.ID,
				Type:       preservationActionType,
				Status:     enums.PreservationActionStatusInProgress,
				StartedAt:  packageStartedAt,
				PackageID:  tinfo.req.PackageID,
			}).
				Get(ctx, &tinfo.PreservationActionID)
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
			Key:         tinfo.req.Key,
			WatcherName: tinfo.req.WatcherName,
		}
		if w.cfg.Preprocessing.Enabled {
			params.DestinationPath = w.cfg.Preprocessing.SharedPath
		}
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.DownloadActivityName, params).
			Get(activityOpts, &downloadResult)
		if err != nil {
			return err
		}
		tinfo.TempPath = downloadResult.Path

		// Delete download tmp dir when session ends.
		cleanup.registerPath(filepath.Dir(downloadResult.Path))
	}

	// Unarchive the transfer if it's not a directory and it's not part of the preprocessing child workflow.
	if !tinfo.IsDir && (!w.cfg.Preprocessing.Enabled || !w.cfg.Preprocessing.Extract) {
		activityOpts := withActivityOptsForLocalAction(sessCtx)
		var result archive.ExtractActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			archive.ExtractActivityName,
			&archive.ExtractActivityParams{SourcePath: tinfo.TempPath},
		).Get(activityOpts, &result)
		if err != nil {
			switch err {
			case archive.ErrInvalidArchive:
				// Not an archive file, bundle the source file as-is.
			default:
				return temporal_tools.NewNonRetryableError(err)
			}
		} else {
			tinfo.TempPath = result.ExtractPath
			tinfo.IsDir = true
		}
	}

	// Preprocessing child workflow.
	if err := w.preprocessing(sessCtx, tinfo); err != nil {
		return err
	}

	// Bundle transfer as an Archivematica standard transfer.
	{
		// For the a3m workflow bundle the transfer to a directory shared with
		// the a3m container.
		var transferDir string
		if w.cfg.Preservation.TaskQueue == temporal.A3mWorkerTaskQueue {
			transferDir = w.cfg.A3m.ShareDir
		}

		activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
		var bundleResult activities.BundleActivityResult
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.BundleActivityName,
			&activities.BundleActivityParams{
				SourcePath:  tinfo.TempPath,
				TransferDir: transferDir,
				IsDir:       tinfo.IsDir,
			},
		).Get(activityOpts, &bundleResult)
		if err != nil {
			return err
		}

		tinfo.Bundle = bundleResult

		// Delete bundled transfer when session ends.
		cleanup.registerPath(bundleResult.FullPath)
	}

	// Do preservation activities.
	{
		var err error
		if w.cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
			err = w.transferAM(sessCtx, tinfo)
		} else {
			err = w.transferA3m(sessCtx, tinfo)
		}
		if err != nil {
			return err
		}
	}

	// Persist SIPID.
	{
		activityOpts := withLocalActivityOpts(sessCtx)
		_ = temporalsdk_workflow.ExecuteLocalActivity(activityOpts, updatePackageLocalActivity, w.logger, w.pkgsvc, &updatePackageLocalActivityParams{
			PackageID: tinfo.req.PackageID,
			Key:       tinfo.req.Key,
			SIPID:     tinfo.SIPID,
			StoredAt:  tinfo.StoredAt,
			Status:    enums.PackageStatusInProgress,
		}).
			Get(activityOpts, nil)
	}

	// Stop here for the Archivematica workflow.
	if w.cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
		return nil
	}

	// Identifier of the preservation task for upload to sips bucket.
	var uploadPreservationTaskID uint

	// Add preservation task for upload to review bucket.
	if !tinfo.req.AutoApproveAIP {
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, createPreservationTaskLocalActivity, w.pkgsvc, &createPreservationTaskLocalActivityParams{
			TaskID:               uuid.NewString(),
			Name:                 "Move AIP",
			Status:               enums.PreservationTaskStatusInProgress,
			StartedAt:            temporalsdk_workflow.Now(sessCtx).UTC(),
			Note:                 "Moving to review bucket",
			PreservationActionID: tinfo.PreservationActionID,
		}).
			Get(ctx, &uploadPreservationTaskID)
		if err != nil {
			return err
		}
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
			AIPPath: tinfo.AIPPath,
			AIPID:   tinfo.SIPID,
			Name:    tinfo.req.Key,
		}).
			Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Complete preservation task for upload to review bucket.
	if !tinfo.req.AutoApproveAIP {
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completePreservationTaskLocalActivity, w.pkgsvc, &completePreservationTaskLocalActivityParams{
			ID:          uploadPreservationTaskID,
			Status:      enums.PreservationTaskStatusDone,
			CompletedAt: temporalsdk_workflow.Now(sessCtx).UTC(),
			Note:        ref.New("Moved to review bucket"),
		}).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	var reviewResult *package_.ReviewPerformedSignal

	// Identifier of the preservation task for package review
	var reviewPreservationTaskID uint

	if tinfo.req.AutoApproveAIP {
		reviewResult = &package_.ReviewPerformedSignal{
			Accepted:   true,
			LocationID: tinfo.req.DefaultPermanentLocationID,
		}
	} else {
		// Set package to pending status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.pkgsvc, tinfo.req.PackageID, enums.PackageStatusPending).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Set preservation action to pending status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setPreservationActionStatusLocalActivity, w.pkgsvc, tinfo.PreservationActionID, enums.PreservationActionStatusPending).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Add preservation task for package review
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, createPreservationTaskLocalActivity, w.pkgsvc, &createPreservationTaskLocalActivityParams{
				TaskID:               uuid.NewString(),
				Name:                 "Review AIP",
				Status:               enums.PreservationTaskStatusPending,
				StartedAt:            temporalsdk_workflow.Now(sessCtx).UTC(),
				Note:                 "Awaiting user decision",
				PreservationActionID: tinfo.PreservationActionID,
			}).Get(ctx, &reviewPreservationTaskID)
			if err != nil {
				return err
			}
		}

		reviewResult = w.waitForReview(sessCtx)

		// Set package to in progress status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.pkgsvc, tinfo.req.PackageID, enums.PackageStatusInProgress).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Set preservation action to in progress status.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setPreservationActionStatusLocalActivity, w.pkgsvc, tinfo.PreservationActionID, enums.PreservationActionStatusInProgress).Get(ctx, nil)
			if err != nil {
				return err
			}
		}
	}

	reviewCompletedAt := temporalsdk_workflow.Now(sessCtx).UTC()

	if reviewResult.Accepted {
		// Record package confirmation in review preservation task
		if !tinfo.req.AutoApproveAIP {
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completePreservationTaskLocalActivity, w.pkgsvc, &completePreservationTaskLocalActivityParams{
				ID:          reviewPreservationTaskID,
				Status:      enums.PreservationTaskStatusDone,
				CompletedAt: reviewCompletedAt,
				Note:        ref.New("Reviewed and accepted"),
			}).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Identifier of the preservation task for permanent storage move.
		var movePreservationTaskID uint

		// Add preservation task for permanent storage move.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, createPreservationTaskLocalActivity, w.pkgsvc, &createPreservationTaskLocalActivityParams{
				TaskID:               uuid.NewString(),
				Name:                 "Move AIP",
				Status:               enums.PreservationTaskStatusInProgress,
				StartedAt:            temporalsdk_workflow.Now(sessCtx).UTC(),
				Note:                 "Moving to permanent storage",
				PreservationActionID: tinfo.PreservationActionID,
			}).
				Get(ctx, &movePreservationTaskID)
			if err != nil {
				return err
			}
		}

		// Move package to permanent storage
		{
			activityOpts := withActivityOptsForRequest(sessCtx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.MoveToPermanentStorageActivityName, &activities.MoveToPermanentStorageActivityParams{
				AIPID:      tinfo.SIPID,
				LocationID: *reviewResult.LocationID,
			}).
				Get(activityOpts, nil)
			if err != nil {
				return err
			}
		}

		// Poll package move to permanent storage
		{
			activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.PollMoveToPermanentStorageActivityName, &activities.PollMoveToPermanentStorageActivityParams{
				AIPID: tinfo.SIPID,
			}).
				Get(activityOpts, nil)
			if err != nil {
				return err
			}
		}

		// Complete preservation task for permanent storage move.
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completePreservationTaskLocalActivity, w.pkgsvc, &completePreservationTaskLocalActivityParams{
				ID:          movePreservationTaskID,
				Status:      enums.PreservationTaskStatusDone,
				CompletedAt: temporalsdk_workflow.Now(sessCtx).UTC(),
				Note:        ref.New(fmt.Sprintf("Moved to location %s", *reviewResult.LocationID)),
			}).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Set package location
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setLocationIDLocalActivity, w.pkgsvc, tinfo.req.PackageID, *reviewResult.LocationID).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}
	} else if !tinfo.req.AutoApproveAIP {
		// Record package rejection in review preservation task
		{
			ctx := withLocalActivityOpts(sessCtx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, completePreservationTaskLocalActivity, w.pkgsvc, &completePreservationTaskLocalActivityParams{
				ID:          reviewPreservationTaskID,
				Status:      enums.PreservationTaskStatusDone,
				CompletedAt: reviewCompletedAt,
				Note:        ref.New("Reviewed and rejected"),
			}).Get(ctx, nil)
			if err != nil {
				return err
			}
		}

		// Reject package
		{
			activityOpts := withActivityOptsForRequest(sessCtx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.RejectPackageActivityName, &activities.RejectPackageActivityParams{
				AIPID: tinfo.SIPID,
			}).Get(activityOpts, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *ProcessingWorkflow) waitForReview(ctx temporalsdk_workflow.Context) *package_.ReviewPerformedSignal {
	var review package_.ReviewPerformedSignal
	signalChan := temporalsdk_workflow.GetSignalChannel(ctx, package_.ReviewPerformedSignalName)
	selector := temporalsdk_workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel temporalsdk_workflow.ReceiveChannel, more bool) {
		_ = channel.Receive(ctx, &review)
	})
	selector.Select(ctx)
	return &review
}

func (w *ProcessingWorkflow) transferA3m(sessCtx temporalsdk_workflow.Context, tinfo *TransferInfo) error {
	activityOpts := temporalsdk_workflow.WithActivityOptions(sessCtx, temporalsdk_workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 24,
		HeartbeatTimeout:    time.Second * 5,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	params := &a3m.CreateAIPActivityParams{
		Name:                 tinfo.Name(),
		Path:                 tinfo.Bundle.FullPath,
		PreservationActionID: tinfo.PreservationActionID,
	}

	result := a3m.CreateAIPActivityResult{}
	err := temporalsdk_workflow.ExecuteActivity(activityOpts, a3m.CreateAIPActivityName, params).Get(sessCtx, &result)
	if err != nil {
		return err
	}

	tinfo.SIPID = result.UUID
	tinfo.AIPPath = result.Path
	tinfo.StoredAt = temporalsdk_workflow.Now(sessCtx).UTC()

	return nil
}

func (w *ProcessingWorkflow) transferAM(sessCtx temporalsdk_workflow.Context, tinfo *TransferInfo) error {
	var err error

	// Zip transfer.
	activityOpts := withActivityOptsForLongLivedRequest(sessCtx)
	var zipResult activities.ZipActivityResult
	err = temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		activities.ZipActivityName,
		&activities.ZipActivityParams{SourceDir: tinfo.Bundle.FullPath},
	).Get(activityOpts, &zipResult)
	if err != nil {
		return err
	}

	// Upload transfer to AMSS.
	activityOpts = temporalsdk_workflow.WithActivityOptions(sessCtx,
		temporalsdk_workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour * 2,
			HeartbeatTimeout:    2 * tinfo.req.PollInterval,
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
		&am.UploadTransferActivityParams{SourcePath: zipResult.Path},
	).Get(activityOpts, &uploadResult)
	if err != nil {
		return err
	}

	// Start AM transfer.
	activityOpts = withActivityOptsForRequest(sessCtx)
	transferResult := am.StartTransferActivityResult{}
	err = temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		am.StartTransferActivityName,
		&am.StartTransferActivityParams{
			Name: tinfo.req.Key,
			Path: uploadResult.RemoteFullPath,
		},
	).Get(activityOpts, &transferResult)
	if err != nil {
		return err
	}

	pollOpts := temporalsdk_workflow.WithActivityOptions(
		sessCtx,
		temporalsdk_workflow.ActivityOptions{
			HeartbeatTimeout:    2 * tinfo.req.PollInterval,
			StartToCloseTimeout: tinfo.req.TransferDeadline,
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
			PresActionID: tinfo.PreservationActionID,
			TransferID:   transferResult.TransferID,
		},
	).Get(pollOpts, &pollTransferResult)
	if err != nil {
		return err
	}

	// Set SIP ID.
	tinfo.SIPID = pollTransferResult.SIPID

	// Poll ingest status.
	var pollIngestResult am.PollIngestActivityResult
	err = temporalsdk_workflow.ExecuteActivity(
		pollOpts,
		am.PollIngestActivityName,
		am.PollIngestActivityParams{
			PresActionID: tinfo.PreservationActionID,
			SIPID:        tinfo.SIPID,
		},
	).Get(pollOpts, &pollIngestResult)
	if err != nil {
		return err
	}

	// Set AIP "stored at" time.
	tinfo.StoredAt = temporalsdk_workflow.Now(sessCtx).UTC()

	// Set package location
	{
		ctx := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteLocalActivity(
			ctx,
			setLocationIDLocalActivity,
			w.pkgsvc,
			tinfo.req.PackageID,
			ref.DerefZero(tinfo.req.DefaultPermanentLocationID),
		).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Create storage package record and set location to AMSS location.
	{
		activityOpts := withLocalActivityOpts(sessCtx)
		err := temporalsdk_workflow.ExecuteActivity(
			activityOpts,
			activities.CreateStoragePackageActivityName,
			&activities.CreateStoragePackageActivityParams{
				Name:       tinfo.req.Key,
				AIPID:      tinfo.SIPID,
				ObjectKey:  tinfo.SIPID,
				LocationID: tinfo.req.DefaultPermanentLocationID,
				Status:     "stored",
			}).
			Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Delete transfer.
	activityOpts = withActivityOptsForRequest(sessCtx)
	err = temporalsdk_workflow.ExecuteActivity(activityOpts, am.DeleteTransferActivityName, am.DeleteTransferActivityParams{
		Destination: uploadResult.RemoteRelativePath,
	}).
		Get(activityOpts, nil)
	if err != nil {
		return err
	}

	return nil
}

func (w *ProcessingWorkflow) preprocessing(ctx temporalsdk_workflow.Context, tinfo *TransferInfo) error {
	if !w.cfg.Preprocessing.Enabled {
		return nil
	}

	// TODO: move package if tinfo.TempPath is not inside w.cfg.Preprocessing.SharedPath.
	relPath, err := filepath.Rel(w.cfg.Preprocessing.SharedPath, tinfo.TempPath)
	if err != nil {
		return err
	}

	preCtx := temporalsdk_workflow.WithChildOptions(ctx, temporalsdk_workflow.ChildWorkflowOptions{
		Namespace:         w.cfg.Preprocessing.Temporal.Namespace,
		TaskQueue:         w.cfg.Preprocessing.Temporal.TaskQueue,
		WorkflowID:        fmt.Sprintf("%s-%s", w.cfg.Preprocessing.Temporal.WorkflowName, uuid.New().String()),
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

	tinfo.TempPath = filepath.Join(w.cfg.Preprocessing.SharedPath, filepath.Clean(ppResult.RelativePath))
	tinfo.IsDir = true

	// Save preprocessing preservation task data.
	if len(ppResult.PreservationTasks) > 0 {
		opts := withLocalActivityOpts(ctx)
		var savePPTasksResult localact.SavePreprocessingTasksActivityResult
		err = temporalsdk_workflow.ExecuteLocalActivity(
			opts,
			localact.SavePreprocessingTasksActivity,
			localact.SavePreprocessingTasksActivityParams{
				PkgSvc:               w.pkgsvc,
				RNG:                  w.rng,
				PreservationActionID: tinfo.PreservationActionID,
				Tasks:                ppResult.PreservationTasks,
			},
		).Get(opts, &savePPTasksResult)
		if err != nil {
			return err
		}

		switch ppResult.Outcome {
		case preprocessing.OutcomeSuccess:
			return nil
		case preprocessing.OutcomeSystemError:
			return errors.New("preprocessing workflow: system error")
		case preprocessing.OutcomeContentError:
			return errors.New("preprocessing workflow: validation failed")
		default:
			return fmt.Errorf("preprocessing workflow: unknown outcome %d", ppResult.Outcome)
		}
	}

	return nil
}

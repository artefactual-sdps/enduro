package workflows

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	temporalsdk_log "go.temporal.io/sdk/log"
	temporalsdk_temporal "go.temporal.io/sdk/temporal"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type StorageDeleteWorkflow struct {
	storagesvc storage.Service
}

func NewStorageDeleteWorkflow(storagesvc storage.Service) *StorageDeleteWorkflow {
	return &StorageDeleteWorkflow{storagesvc: storagesvc}
}

func (w *StorageDeleteWorkflow) Execute(
	ctx temporalsdk_workflow.Context,
	req storage.StorageDeleteWorkflowRequest,
) (e error) {
	logger := temporalsdk_workflow.GetLogger(ctx)
	logger.Info("Started AIP deletion workflow", "request", req)

	// Fail workflow if the AIP is no longer stored.
	var aip goastorage.AIP
	activityOpts := localActivityOptions(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.ReadAIPLocalActivity,
		w.storagesvc,
		req.AIPID,
	).Get(activityOpts, &aip)
	if err != nil {
		return err
	}
	if aip.Status != enums.AIPStatusStored.String() {
		return fmt.Errorf("AIP is no longer stored")
	}
	if aip.LocationID == nil || *aip.LocationID == uuid.Nil {
		return fmt.Errorf("AIP location is missing")
	}

	// Set AIP status to processing.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusProcessing); err != nil {
		return err
	}

	// Create persistence workflow.
	workflowDBID, err := createWorkflow(ctx, w.storagesvc, req.AIPID, enums.WorkflowTypeDeleteAip)
	if err != nil {
		return err
	}

	aipStatus := enums.AIPStatusStored
	defer func() {
		workflowStatus := enums.WorkflowStatusDone
		// Only looking at internal cancellation.
		if errors.Is(e, temporalsdk_workflow.ErrCanceled) {
			workflowStatus = enums.WorkflowStatusCanceled
		} else if e != nil {
			workflowStatus = enums.WorkflowStatusError
		}

		// Complete persistence workflow.
		if err := completeWorkflow(ctx, w.storagesvc, workflowDBID, workflowStatus); err != nil {
			e = errors.Join(e, err)
		}

		// Set AIP status to stored/deleted.
		if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, aipStatus); err != nil {
			e = errors.Join(e, err)
		}
	}()

	// Create review task.
	taskNote := fmt.Sprintf("An AIP deletion has been requested by %s. Reason:\n\n%s", req.UserEmail, req.Reason)
	reviewTaskDBID, err := createTask(
		ctx,
		w.storagesvc,
		workflowDBID,
		enums.TaskStatusPending,
		"Review AIP deletion request",
		fmt.Sprintf("%s\n\nAwaiting user review.", taskNote),
	)
	if err != nil {
		return err
	}

	// Wrap review into a function to be able to complete review Task on error.
	reviewSignal, err := w.review(ctx, logger, req, workflowDBID)

	// Complete review task.
	taskStatus := enums.TaskStatusDone
	if err != nil {
		taskStatus = enums.TaskStatusError
		taskNote = fmt.Sprintf("%s\n\nFailed to review AIP deletion request:\n%v", taskNote, err)
	} else {
		switch reviewSignal.Status {
		case enums.DeletionRequestStatusApproved:
			taskNote = fmt.Sprintf("%s\n\nAIP deletion request approved by %s.", taskNote, reviewSignal.UserEmail)
		case enums.DeletionRequestStatusRejected:
			taskNote = fmt.Sprintf("%s\n\nAIP deletion request rejected by %s.", taskNote, reviewSignal.UserEmail)
		case enums.DeletionRequestStatusCanceled:
			taskNote = fmt.Sprintf("%s\n\nAIP deletion request canceled by %s.", taskNote, reviewSignal.UserEmail)
		}
	}
	taskErr := completeTask(ctx, w.storagesvc, reviewTaskDBID, taskStatus, taskNote)
	if err = errors.Join(err, taskErr); err != nil {
		return err
	}

	// Cancel workflow if the request is not approved.
	if reviewSignal.Status != enums.DeletionRequestStatusApproved {
		return temporalsdk_workflow.ErrCanceled
	}

	// Create delete AIP task.
	deleteTaskID, err := createTask(
		ctx,
		w.storagesvc,
		workflowDBID,
		enums.TaskStatusInProgress,
		"Delete AIP",
		"Deleting AIP",
	)
	if err != nil {
		return err
	}

	// Get location info.
	var locationInfo storage.ReadLocationInfoLocalActivityResult
	activityOpts = localActivityOptions(ctx)
	err = temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.ReadLocationInfoLocalActivity,
		w.storagesvc,
		*aip.LocationID,
	).Get(activityOpts, &locationInfo)
	if err != nil {
		return errors.Join(err, completeTask(
			ctx,
			w.storagesvc,
			deleteTaskID,
			enums.TaskStatusError,
			fmt.Sprintf("Failed to get location information:\n%v", err),
		))
	}

	// Delete AIP based on location source.
	deleted := true
	switch locationInfo.Source {
	case enums.LocationSourceAmss:
		deleted, err = w.deleteAIPFromAMSSLocation(ctx, aip.UUID, locationInfo.Config)
	case enums.LocationSourceMinio:
		activityOpts := localActivityOptions(ctx)
		err = temporalsdk_workflow.ExecuteLocalActivity(
			activityOpts,
			storage.DeleteFromMinIOLocationLocalActivity,
			w.storagesvc,
			&storage.DeleteFromMinIOLocationLocalActivityParams{AIPID: req.AIPID},
		).Get(activityOpts, nil)
	default:
		err = fmt.Errorf("unsupported location source: %s", locationInfo.Source)
	}

	// Complete delete AIP task.
	taskStatus = enums.TaskStatusDone
	source := strings.ToUpper(locationInfo.Source.String())
	taskNote = fmt.Sprintf("AIP deleted from %s source location", source)
	if err != nil {
		taskStatus = enums.TaskStatusError
		taskNote = fmt.Sprintf("Failed to delete AIP:\n%v", err)
	} else if !deleted {
		taskNote = fmt.Sprintf("AIP request rejected in %s source location", source)
	}
	taskErr = completeTask(ctx, w.storagesvc, deleteTaskID, taskStatus, taskNote)
	if err = errors.Join(err, taskErr); err != nil {
		return err
	}

	// Cancel workflow if the request was rejected in AMSS.
	if !deleted {
		return temporalsdk_workflow.ErrCanceled
	}

	// If all goes well update AIP status to deleted, used in the defer function.
	aipStatus = enums.AIPStatusDeleted

	return nil
}

func (w *StorageDeleteWorkflow) review(
	ctx temporalsdk_workflow.Context,
	logger temporalsdk_log.Logger,
	req storage.StorageDeleteWorkflowRequest,
	workflowDBID int,
) (*storage.DeletionDecisionSignal, error) {
	// Create DeletionRequest.
	var drDBID int
	activityOpts := localActivityOptions(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.CreateDeletionRequestLocalActivity,
		w.storagesvc,
		&storage.CreateDeletionRequestLocalActivityParams{
			Requester:    req.UserEmail,
			RequesterIss: req.UserIss,
			RequesterSub: req.UserSub,
			Reason:       req.Reason,
			WorkflowDBID: workflowDBID,
			AIPUUID:      req.AIPID,
		},
	).Get(activityOpts, &drDBID)
	if err != nil {
		return nil, err
	}

	// Set Workflow status to pending.
	if err := updateWorkflowStatus(ctx, w.storagesvc, workflowDBID, enums.WorkflowStatusPending); err != nil {
		return nil, err
	}

	// Set AIP status to pending.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusPending); err != nil {
		return nil, err
	}

	// Wait for a delete request decision signal.
	var signal storage.DeletionDecisionSignal
	open := temporalsdk_workflow.GetSignalChannel(ctx, storage.DeletionDecisionSignalName).Receive(ctx, &signal)
	if !open {
		return nil, fmt.Errorf("deletion decision signal channel closed")
	}

	logger.Info("Received AIP deletion workflow decision", "signal", signal)

	// Set AIP status to processing.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusProcessing); err != nil {
		return nil, err
	}

	// Set Workflow status to in progress.
	if err := updateWorkflowStatus(ctx, w.storagesvc, workflowDBID, enums.WorkflowStatusInProgress); err != nil {
		return nil, err
	}

	// Update DeletionRequest.
	err = temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		storage.UpdateDeletionRequestLocalActivity,
		w.storagesvc,
		drDBID,
		signal,
	).Get(activityOpts, nil)
	if err != nil {
		return nil, err
	}

	return &signal, nil
}

func (w *StorageDeleteWorkflow) deleteAIPFromAMSSLocation(
	ctx temporalsdk_workflow.Context,
	aipID uuid.UUID,
	config types.LocationConfig,
) (bool, error) {
	var configValue *types.AMSSConfig
	switch c := config.Value.(type) {
	case *types.AMSSConfig:
		configValue = config.Value.(*types.AMSSConfig)
	default:
		return false, fmt.Errorf("unsupported config type: %T", c)
	}
	if configValue == nil {
		return false, fmt.Errorf("missing AMSS config")
	}

	var re activities.DeleteFromAMSSLocationActivityResult
	activityOpts := temporalsdk_workflow.WithActivityOptions(ctx, temporalsdk_workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 2,
		HeartbeatTimeout:    time.Second * 20,
		RetryPolicy: &temporalsdk_temporal.RetryPolicy{
			InitialInterval:    time.Second * 30,
			BackoffCoefficient: 1.5,
			MaximumInterval:    time.Minute * 10,
			MaximumAttempts:    5,
			NonRetryableErrorTypes: []string{
				"TemporalTimeout:StartToClose",
			},
		},
	})
	err := temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		storage.DeleteFromAMSSLocationActivityName,
		&activities.DeleteFromAMSSLocationActivityParams{
			Config:  *configValue,
			AIPUUID: aipID,
		},
	).Get(activityOpts, &re)
	if err != nil {
		return false, err
	}

	return re.Deleted, nil
}

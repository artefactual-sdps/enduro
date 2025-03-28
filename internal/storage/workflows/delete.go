package workflows

import (
	"errors"
	"fmt"

	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
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
	// TODO: Check AIP existence and status, fail workflow.

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
		// Only looking at internal cancelation.
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
	taskNote := fmt.Sprintf("An AIP deletion has been requested. Reason:\n\n%s", req.Reason)
	reviewTaskID, err := createTask(
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

	// TODO: Create DeletionRequest.

	// Set Workflow status to pending.
	if err := updateWorkflowStatus(ctx, w.storagesvc, workflowDBID, enums.WorkflowStatusPending); err != nil {
		return err
	}

	// Set AIP status to pending.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusPending); err != nil {
		return err
	}

	// Wait for user review signal.
	reviewResult := w.waitForReview(ctx)

	// Set AIP status to processing.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusProcessing); err != nil {
		return err
	}

	// Set Workflow status to pending.
	if err := updateWorkflowStatus(ctx, w.storagesvc, workflowDBID, enums.WorkflowStatusInProgress); err != nil {
		return err
	}

	// TODO: Update DeletionRequest.

	// Complete review task.
	if reviewResult.Approved {
		taskNote = fmt.Sprintf("%s\n\nAIP deletion request approved.", taskNote)
	} else {
		taskNote = fmt.Sprintf("%s\n\nAIP deletion request rejected.", taskNote)
	}
	if err = completeTask(ctx, w.storagesvc, reviewTaskID, enums.TaskStatusDone, taskNote); err != nil {
		return err
	}

	// Cancel workflow if the request is not approved.
	if !reviewResult.Approved {
		return temporalsdk_workflow.ErrCanceled
	}

	// TODO: Delete the AIP from AMSS or MinIO location.
	aipStatus = enums.AIPStatusDeleted

	return nil
}

func (w *StorageDeleteWorkflow) waitForReview(ctx temporalsdk_workflow.Context) *storage.DeletionReviewedSignal {
	var review storage.DeletionReviewedSignal
	signalChan := temporalsdk_workflow.GetSignalChannel(ctx, storage.DeletionReviewedSignalName)
	selector := temporalsdk_workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel temporalsdk_workflow.ReceiveChannel, more bool) {
		_ = channel.Receive(ctx, &review)
	})
	selector.Select(ctx)
	return &review
}

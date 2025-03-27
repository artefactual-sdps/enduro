package workflows

import (
	"errors"

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
	// TODO: Update AIP status enum and use proper status.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusInReview); err != nil {
		return err
	}

	// Create persistence workflow.
	workflowDBID, err := createWorkflow(ctx, w.storagesvc, req.AIPID, enums.WorkflowTypeDeleteAip)
	if err != nil {
		return err
	}

	// Complete persistence workflow.
	defer func() {
		// TODO: Consider rejected/cancelled case.
		status := enums.WorkflowStatusDone
		if e != nil {
			status = enums.WorkflowStatusError
		}

		err := completeWorkflow(ctx, w.storagesvc, workflowDBID, status)
		if err != nil {
			e = errors.Join(e, err)
		}
	}()

	// Create review task.
	reviewTaskID, err := createTask(
		ctx,
		w.storagesvc,
		workflowDBID,
		"Review AIP deletion request",
		"Awaiting user decision",
	)
	if err != nil {
		return err
	}

	// TODO:
	// - Create DeletionRequest.
	// - Add signal channel, etc.
	// - Update DeletionRequest.

	// Complete review task.
	taskNote := "AIP deletion request approved"
	if false {
		taskNote = "AIP deletion request rejected"
	}
	taskErr := completeTask(ctx, w.storagesvc, reviewTaskID, enums.TaskStatusDone, taskNote)
	if err = errors.Join(err, taskErr); err != nil {
		return err
	}

	// TODO: Delete the AIP from AMSS or MinIO location source.

	// Set AIP status to deleted.
	// TODO: Update AIP status enum and use proper status.
	if err := updateAIPStatus(ctx, w.storagesvc, req.AIPID, enums.AIPStatusRejected); err != nil {
		return err
	}

	return nil
}

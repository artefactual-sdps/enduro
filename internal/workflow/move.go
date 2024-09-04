package workflow

import (
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

type MoveWorkflow struct {
	pkgsvc package_.Service
}

func NewMoveWorkflow(pkgsvc package_.Service) *MoveWorkflow {
	return &MoveWorkflow{
		pkgsvc: pkgsvc,
	}
}

func (w *MoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req *package_.MoveWorkflowRequest) error {
	// Save starting time for preservation action.
	startedAt := temporalsdk_workflow.Now(ctx).UTC()

	// Assume the preservation action will be successful.
	status := enums.PreservationActionStatusDone

	// Set package to in progress status.
	{
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.pkgsvc, req.ID, enums.PackageStatusInProgress).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Move package to permanent storage
	{
		activityOpts := withActivityOptsForRequest(ctx)
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.MoveToPermanentStorageActivityName, &activities.MoveToPermanentStorageActivityParams{
			AIPID:      req.AIPID,
			LocationID: req.LocationID,
		}).
			Get(activityOpts, nil)
		if err != nil {
			status = enums.PreservationActionStatusError
		}
	}

	// Poll package move to permanent storage
	{
		if status != enums.PreservationActionStatusError {
			activityOpts := withActivityOptsForLongLivedRequest(ctx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.PollMoveToPermanentStorageActivityName, &activities.PollMoveToPermanentStorageActivityParams{
				AIPID: req.AIPID,
			}).
				Get(activityOpts, nil)
			if err != nil {
				status = enums.PreservationActionStatusError
			}
		}
	}

	// TODO: add compensation to these next activities?

	// Set package to done status.
	{
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.pkgsvc, req.ID, enums.PackageStatusDone).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Set package location.
	{
		if status != enums.PreservationActionStatusError {
			ctx := withLocalActivityOpts(ctx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setLocationIDLocalActivity, w.pkgsvc, req.ID, req.LocationID).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}
	}

	// Create preservation action.
	{
		ctx := withLocalActivityOpts(ctx)
		completedAt := temporalsdk_workflow.Now(ctx).UTC()
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, saveLocationMovePreservationActionLocalActivity, w.pkgsvc, &saveLocationMovePreservationActionLocalActivityParams{
			PackageID:   req.ID,
			LocationID:  req.LocationID,
			WorkflowID:  temporalsdk_workflow.GetInfo(ctx).WorkflowExecution.ID,
			Type:        enums.PreservationActionTypeMovePackage,
			Status:      status,
			StartedAt:   startedAt,
			CompletedAt: completedAt,
		}).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

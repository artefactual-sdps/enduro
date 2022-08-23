package workflow

import (
	"github.com/go-logr/logr"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

type MoveWorkflow struct {
	logger logr.Logger
	pkgsvc package_.Service
}

func NewMoveWorkflow(logger logr.Logger, pkgsvc package_.Service) *MoveWorkflow {
	return &MoveWorkflow{
		logger: logger,
		pkgsvc: pkgsvc,
	}
}

func (w *MoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req *package_.MoveWorkflowRequest) error {
	// Save starting time for preservation action.
	startedAt := temporalsdk_workflow.Now(ctx).UTC()

	// Assume the preservation action will be successful.
	status := package_.ActionStatusDone

	// Set package to in progress status.
	{
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.pkgsvc, req.ID, package_.StatusInProgress).Get(ctx, nil)
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
		}).Get(activityOpts, nil)
		if err != nil {
			status = package_.ActionStatusError
		}
	}

	// Poll package move to permanent storage
	{
		if status != package_.ActionStatusError {
			activityOpts := withActivityOptsForLongLivedRequest(ctx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.PollMoveToPermanentStorageActivityName, &activities.PollMoveToPermanentStorageActivityParams{
				AIPID: req.AIPID,
			}).Get(activityOpts, nil)
			if err != nil {
				status = package_.ActionStatusError
			}
		}
	}

	// TODO: add compensation to these next activities?

	// Set package to done status.
	{
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.pkgsvc, req.ID, package_.StatusDone).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Set package location.
	{
		if status != package_.ActionStatusError {
			ctx := withLocalActivityOpts(ctx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setLocationIDLocalActivity, w.pkgsvc, req.ID, req.LocationID).Get(ctx, nil)
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
			Type:        package_.ActionTypeMovePackage,
			Status:      status,
			StartedAt:   startedAt,
			CompletedAt: completedAt,
		}).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

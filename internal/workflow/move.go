package workflow

import (
	"github.com/go-logr/logr"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-labs/enduro/internal/package_"
	"github.com/artefactual-labs/enduro/internal/workflow/activities"
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
			AIPID:    req.AIPID,
			Location: req.Location,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
		}
	}

	// Poll package move to permanent storage
	{
		activityOpts := withActivityOptsForLongLivedRequest(ctx)
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.PollMoveToPermanentStorageActivityName, &activities.PollMoveToPermanentStorageActivityParams{
			AIPID: req.AIPID,
		}).Get(activityOpts, nil)
		if err != nil {
			return err
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
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setLocationLocalActivity, w.pkgsvc, req.ID, req.Location).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
